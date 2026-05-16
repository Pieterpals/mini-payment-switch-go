package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Response is the standard API response envelope.
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo provides structured error details in the API response.
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Success returns a 200 OK response with the given data.
func Success(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// Created returns a 201 Created response with the given data.
func Created(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// Error returns an error response with the given HTTP status, error code, and message.
func Error(c echo.Context, status int, code, message string) error {
	return c.JSON(status, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	})
}
