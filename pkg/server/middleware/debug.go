package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"go.uber.org/zap"
)

// DebugRequestBody logs request bodies for POST/PUT/PATCH requests in debug mode
func DebugRequestBody() fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := logger.Get()
		method := c.Method()

		// Only log bodies for methods that typically have request bodies
		if method == "POST" || method == "PUT" || method == "PATCH" {
			body := c.Body()
			if len(body) > 0 {
				// Try to pretty-print JSON
				var prettyJSON bytes.Buffer
				if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
					log.Debug("Request body",
						zap.String("method", method),
						zap.String("path", c.Path()),
						zap.String("body", prettyJSON.String()),
					)
				} else {
					// Not JSON or invalid JSON, log as-is (truncated if too long)
					bodyStr := string(body)
					if len(bodyStr) > 2000 {
						bodyStr = bodyStr[:2000] + "... (truncated)"
					}
					log.Debug("Request body",
						zap.String("method", method),
						zap.String("path", c.Path()),
						zap.String("body", bodyStr),
					)
				}
			}
		}

		// Log query parameters for all requests
		if c.Request().URI().QueryString() != nil && len(c.Request().URI().QueryString()) > 0 {
			log.Debug("Request query params",
				zap.String("method", method),
				zap.String("path", c.Path()),
				zap.String("query", string(c.Request().URI().QueryString())),
			)
		}

		return c.Next()
	}
}

// DebugErrorHandler wraps responses to log detailed error information
func DebugErrorHandler(c *fiber.Ctx, err error) error {
	log := logger.Get()

	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Log detailed error information
	log.Error("Request error",
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.Int("status", code),
		zap.Error(err),
		zap.String("ip", c.IP()),
		zap.String("user_agent", c.Get("User-Agent")),
	)

	// Log request body for failed POST/PUT/PATCH requests
	if c.Method() == "POST" || c.Method() == "PUT" || c.Method() == "PATCH" {
		// Re-read the body since it might have been consumed
		body := c.Body()
		if len(body) > 0 {
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
				log.Error("Failed request body",
					zap.String("body", prettyJSON.String()),
				)
			} else {
				bodyStr := string(body)
				if len(bodyStr) > 2000 {
					bodyStr = bodyStr[:2000] + "... (truncated)"
				}
				log.Error("Failed request body",
					zap.String("body", bodyStr),
				)
			}
		}
	}

	// Custom handling for 403 Forbidden
	if code == fiber.StatusForbidden {
		path := c.Path()
		// Redirect to root for UI routes that return 403
		if path == "/login" ||
			strings.HasPrefix(path, "/events") ||
			strings.HasPrefix(path, "/inventory") ||
			strings.HasPrefix(path, "/llm") ||
			strings.HasPrefix(path, "/registry") ||
			strings.HasPrefix(path, "/scans") ||
			strings.HasPrefix(path, "/schedules") ||
			strings.HasPrefix(path, "/utilities") ||
			strings.HasPrefix(path, "/vuln") ||
			strings.HasPrefix(path, "/workflow") {
			return c.Redirect("/", fiber.StatusFound)
		}
		return c.Status(code).JSON(fiber.Map{
			"error":   true,
			"message": "Oh dear! It seems you've wandered off the path. If you'd like to see the UI page, please pop back root route at /",
			"code":    code,
			"path":    path,
			"method":  c.Method(),
		})
	}

	// Return detailed error response in debug mode
	return c.Status(code).JSON(fiber.Map{
		"error":   true,
		"message": err.Error(),
		"code":    code,
		"path":    c.Path(),
		"method":  c.Method(),
	})
}

// DebugResponseLogger logs response status for debugging
func DebugResponseLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		log := logger.Get()
		status := c.Response().StatusCode()

		// Log non-2xx responses with more detail
		if status >= 400 {
			log.Debug("Response",
				zap.String("method", c.Method()),
				zap.String("path", c.Path()),
				zap.Int("status", status),
				zap.Int("body_size", len(c.Response().Body())),
			)
		}

		return err
	}
}

// Ensure body can be re-read for logging (use before DebugRequestBody)
func BodyReusable() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Store original body so it can be read multiple times
		body := c.Body()
		c.Request().SetBody(body)

		// Also set body reader for handlers that use io.Reader
		c.Request().SetBodyStream(io.NopCloser(bytes.NewReader(body)), len(body))

		return c.Next()
	}
}
