package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/libs"
)

// Ping is a function to check server connectivity
// @Summary Check server connectivity
// @Description Simple ping endpoint to verify server is responding
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} ResponseHTTP{}
// @Failure 503 {object} ResponseHTTP{}
// @Router /v1/ping [get]
func Ping(c *fiber.Ctx) error {
	return c.JSON(ResponseHTTP{
		Status:  200,
		Message: "pong",
	})
}

// Health is a function to check server health status
// @Summary Check server health status
// @Description Get server health status and version information
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} ResponseHTTP{}
// @Failure 503 {object} ResponseHTTP{}
// @Router /v1/health [get]
func Health(c *fiber.Ctx) error {
	return c.JSON(ResponseHTTP{
		Status: 200,
		Data: fiber.Map{
			"version": libs.VERSION,
		},
		Message: "server is up",
	})
}
