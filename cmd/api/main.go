package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"

	// Shared packages
	"mini-payment-switch/internal/shared/config"
	"mini-payment-switch/internal/shared/logger"
	"mini-payment-switch/internal/shared/observability"
	"mini-payment-switch/internal/shared/server"

	// Payment module — Adapters (outermost ring)
	paymentCache "mini-payment-switch/internal/payment/adapter/cache"
	paymentHandler "mini-payment-switch/internal/payment/adapter/handler"
	paymentPublisher "mini-payment-switch/internal/payment/adapter/publisher"
	paymentRepo "mini-payment-switch/internal/payment/adapter/repository"

	// Payment module — Use Case (application ring)
	paymentUseCase "mini-payment-switch/internal/payment/usecase"

	// Notification module
	notifConsumer "mini-payment-switch/internal/notification/adapter/consumer"
	"mini-payment-switch/internal/notification/adapter/telegram"
	notifPort "mini-payment-switch/internal/notification/port"
	notifUseCase "mini-payment-switch/internal/notification/usecase"
)

// @title Mini Payment Switch API
// @version 1.0
// @description This is a modular payment switch API with inquiry, execute, and status check.
// @host localhost:8182
// @BasePath /
func main() {
	// ========================================================================
	// 1. LOAD CONFIGURATION
	// ========================================================================
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		slog.Error("failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// ========================================================================
	// 2. SETUP STRUCTURED LOGGER
	// ========================================================================
	log := logger.Setup(cfg.App.Env)
	log.Info("🚀 Starting application",
		slog.String("app", cfg.App.Name),
		slog.String("env", cfg.App.Env),
		slog.Int("port", cfg.App.Port),
	)

	// ========================================================================
	// 2.5 SETUP OPENTELEMETRY TRACER
	// ========================================================================
	shutdownTracer, err := observability.InitTracer(cfg.App.Name, cfg.App.OTel.CollectorURL)
	if err != nil {
		log.Error("Failed to initialize OpenTelemetry tracer", slog.String("error", err.Error()))
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = shutdownTracer(ctx)
		}()
	}

	// ========================================================================
	// 3. INITIALIZE INFRASTRUCTURE (External Dependencies)
	// ========================================================================
	ctx := context.Background()

	// --- PostgreSQL Connection Pool ---
	dbPool, err := initPostgres(ctx, cfg.Database)
	if err != nil {
		log.Error("failed to connect to PostgreSQL", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbPool.Close()
	log.Info("✅ PostgreSQL connected",
		slog.String("host", cfg.Database.Host),
		slog.Int("port", cfg.Database.Port),
		slog.String("database", cfg.Database.Name),
	)

	// --- Redis Client ---
	rdb := initRedis(cfg.Redis)
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Error("failed to connect to Redis", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("✅ Redis connected", slog.String("addr", cfg.Redis.Addr))

	// --- Kafka Producer (Writer) ---
	kafkaWriter := initKafkaWriter(cfg.Kafka)
	defer kafkaWriter.Close()
	log.Info("✅ Kafka producer initialized", slog.Any("brokers", cfg.Kafka.Brokers))

	// ========================================================================
	// 4. DEPENDENCY INJECTION (Constructor Injection)
	//    Wiring order: Adapters → Use Cases → Handlers
	//    Each layer only knows about the layer directly below it (via interfaces).
	// ========================================================================

	// --- Payment Module ---
	// Layer 4 (Adapters): Concrete implementations of port interfaces
	txRepo := paymentRepo.NewPgTransactionRepository(dbPool)
	distLock := paymentCache.NewRedisDistributedLock(rdb)
	cacheStore := paymentCache.NewRedisCache(rdb)
	eventPub := paymentPublisher.NewKafkaEventPublisher(kafkaWriter, cfg.Kafka.Topic.PaymentSuccess)

	// Layer 2 (Use Cases): Business logic orchestrators
	inquiryUC := paymentUseCase.NewInquiryUseCase(cacheStore, log)
	executePaymentUC := paymentUseCase.NewExecutePaymentUseCase(txRepo, cacheStore, distLock, eventPub, log)
	checkStatusUC := paymentUseCase.NewCheckStatusUseCase(txRepo, log)

	// Layer 1 (Handlers): Outer layer exposing APIs
	payHandler := paymentHandler.NewPaymentHandler(inquiryUC, executePaymentUC, checkStatusUC)

	// --- Notification Module ---
	// Initialize Telegram notifier (optional — gracefully skips if not configured)
	var notifier notifPort.Notifier
	if cfg.Telegram.BotToken != "" && cfg.Telegram.ChatID != "" {
		notifier = telegram.NewTelegramNotifier(cfg.Telegram.BotToken, cfg.Telegram.ChatID)
		log.Info("✅ Telegram notifier configured", slog.String("chat_id", cfg.Telegram.ChatID))
	} else {
		log.Info("⚠️  Telegram not configured — notifications will be logged only")
	}
	notifUC := notifUseCase.NewSendNotificationUseCase(notifier, log)
	consumer := notifConsumer.NewKafkaNotificationConsumer(cfg.Kafka, notifUC, log)

	// ========================================================================
	// 5. SETUP HTTP SERVER & ROUTES
	// ========================================================================
	e := server.NewEchoServer()

	// Register module routes
	payHandler.RegisterRoutes(e)

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "healthy",
			"service": cfg.App.Name,
		})
	})

	// ========================================================================
	// 6. START KAFKA CONSUMER (Background Goroutine)
	// ========================================================================
	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	defer consumerCancel()
	go consumer.Start(consumerCtx)

	// ========================================================================
	// 7. START HTTP SERVER (Non-blocking)
	// ========================================================================
	go func() {
		addr := fmt.Sprintf(":%d", cfg.App.Port)
		log.Info("🌐 HTTP server starting", slog.String("addr", addr))
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server fatal error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// ========================================================================
	// 8. GRACEFUL SHUTDOWN
	//    Listens for OS signals (SIGINT from Ctrl+C, SIGTERM from Docker/K8s).
	//    Shutdown order:
	//      1. Stop accepting new HTTP requests
	//      2. Cancel Kafka consumer context (stop reading new messages)
	//      3. Wait for in-flight requests and Kafka messages to complete
	//      4. Close infrastructure connections (DB pool, Redis, Kafka writer)
	// ========================================================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Info("⚠️  Shutdown signal received, starting graceful shutdown...",
		slog.String("signal", sig.String()),
	)

	// Give in-flight operations up to 30 seconds to complete
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Step 1: Shutdown HTTP server (stops accepting new requests, waits for in-flight)
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server shutdown error", slog.String("error", err.Error()))
	} else {
		log.Info("✅ HTTP server stopped")
	}

	// Step 2: Stop Kafka consumer (cancel context + close reader)
	consumerCancel()
	consumer.Stop()
	log.Info("✅ Kafka consumer stopped")

	// Step 3: Infrastructure cleanup happens via deferred Close() calls above
	// (dbPool.Close(), rdb.Close(), kafkaWriter.Close())

	log.Info("👋 Application stopped gracefully")
}

// =============================================================================
// Infrastructure Initialization Functions
// =============================================================================

// initPostgres creates a PostgreSQL connection pool using pgxpool (pgx v5).
// pgxpool manages connection lifecycle, pooling, and health checks automatically.
func initPostgres(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure pool settings
	poolCfg.MaxConns = int32(cfg.MaxOpenConns)
	poolCfg.MinConns = int32(cfg.MaxIdleConns)
	poolCfg.MaxConnLifetime = cfg.ConnMaxLifetimeDuration()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connectivity
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// initRedis creates a Redis client configured from the application config.
func initRedis(cfg config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}

// initKafkaWriter creates a Kafka producer (Writer) with LeastBytes balancing.
func initKafkaWriter(cfg config.KafkaConfig) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    cfg.Topic.PaymentSuccess,
		Balancer: &kafka.LeastBytes{},
	}
}
