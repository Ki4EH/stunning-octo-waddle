package middleware

import (
	"github.com/Ki4EH/stunning-octo-waddle/internal/logger"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"time"
)

func LoggingMiddleware(log logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			eclipse := time.Since(start)

			if eclipse > 50*time.Millisecond {
				log.Info("request",
					zap.String("method", c.Request().Method),
					zap.String("path", c.Request().URL.Path),
					zap.Int("status", c.Response().Status),
					zap.Duration("duration", eclipse),
				)
			}

			return err
		}
	}
}
