package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

// RequestLogger логирует входящие HTTP-запросы через slog.
func RequestLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			res := c.Response()
			duration := time.Since(start)

			logger.Info("http request",
				slog.String("method", req.Method),
				slog.String("uri", req.RequestURI),
				slog.Int("status", res.Status),
				slog.Duration("latency", duration),
				slog.String("remote_ip", c.RealIP()),
			)

			return err
		}
	}
}

