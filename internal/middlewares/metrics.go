package middlewares

import (
	"main/pkg/metrics"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
)

func MetricsMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		// Xử lý request
		err := c.Next()

		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Response().StatusCode())
		method := c.Method()
		path := c.Route().Path // VD: /api/v1/orders/:id

		metrics.OrderHTTPRequestTotal.
			WithLabelValues(method, path, statusCode).Inc()

		metrics.OrderHTTPDuration.
			WithLabelValues(method, path).Observe(duration)

		return err
	}
}
