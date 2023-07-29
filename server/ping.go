package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/libs"
)

// Ping is a function to get all books data from database
// @Summary Get all books
// @Description Get all books
// @Tags books
// @Accept json
// @Produce json
// @Success 200 {object} ResponseHTTP{}
// @Failure 503 {object} ResponseHTTP{}
// @Router /v1/books [get]
func Ping(c *fiber.Ctx) error {
	return c.JSON(ResponseHTTP{
		Status:  200,
		Message: "pong",
	})
}

// Health is a function to get all books data from database
// @Summary Get all books
// @Description Get all books
// @Tags books
// @Accept json
// @Produce json
// @Success 200 {object} ResponseHTTP{}
// @Failure 503 {object} ResponseHTTP{}
// @Router /v1/books [get]
func Health(c *fiber.Ctx) error {
	return c.JSON(ResponseHTTP{
		Status: 200,
		Data: fiber.Map{
			"version": libs.VERSION,
		},
		Message: "server is up",
	})
}
