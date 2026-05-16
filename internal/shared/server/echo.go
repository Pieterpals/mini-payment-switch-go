package server

import (
	"log/slog"
	"time"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	echoSwagger "github.com/swaggo/echo-swagger"
	_ "mini-payment-switch/docs"

	appValidator "mini-payment-switch/internal/shared/validator"
)

// NewEchoServer creates and configures a new Echo HTTP server with production-ready middleware.
func NewEchoServer() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Register custom validator for request body validation
	e.Validator = appValidator.New()

	// Recovery middleware — catches panics and returns 500 instead of crashing
	e.Use(middleware.Recover())

	// OpenTelemetry middleware — automatically traces all requests
	e.Use(otelecho.Middleware("mini-payment-switch"))

	// Request ID middleware — injects unique X-Request-Id header for tracing
	e.Use(middleware.RequestID())

	// CORS middleware — allows cross-origin requests
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.PATCH},
	}))

	// Prometheus metrics middleware — auto-collects request count, latency, response size
	e.Use(echoprometheus.NewMiddleware("payment_switch"))

	// Expose /metrics endpoint for Prometheus to scrape
	e.GET("/metrics", echoprometheus.NewHandler())

	// Structured logging middleware using slog
	e.Use(slogRequestLogger())

	// Swagger documentation endpoint
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	return e
}

// slogRequestLogger returns an Echo middleware that logs each HTTP request using slog.
func slogRequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Process the request
			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()
			latency := time.Since(start)

			// Skip logging /metrics to avoid noise
			if req.RequestURI == "/metrics" {
				return nil
			}

			attrs := []any{
				slog.String("method", req.Method),
				slog.String("uri", req.RequestURI),
				slog.Int("status", res.Status),
				slog.Duration("latency", latency),
				slog.String("remote_ip", c.RealIP()),
				slog.String("request_id", res.Header().Get(echo.HeaderXRequestID)),
			}

			if err != nil {
				attrs = append(attrs, slog.String("error", err.Error()))
				slog.Error("HTTP request failed", attrs...)
			} else {
				slog.Info("HTTP request", attrs...)
			}

			return nil
		}
	}
}
