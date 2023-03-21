package server

import (
	"time"

	// jwtware "github.com/gofiber/jwt/v2"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

// authen stuff

func Login(c *fiber.Ctx) error {
	type LoginInput struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	user := input.Username
	pass := input.Password

	dbUser := Opt.Client.Username
	dbPass := Opt.Client.Password

	// Throws Unauthorized error
	if dbUser != user || dbPass != pass {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Create the Claims
	claims := jwt.MapClaims{
		"name":  "Osmdeus Default",
		"admin": true,
		"exp":   time.Now().Add(time.Hour * 24 * 30).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(Opt.Server.JWTSecret))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.JSON(fiber.Map{"status": "success", "message": "Successfully login", "token": t})

}

func jwtError(c *fiber.Ctx, err error) error {
	if err.Error() == "Missing or malformed JWT" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{"status": "error", "message": "Missing or malformed JWT", "data": nil})

	} else {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{"status": "error", "message": "Invalid or expired JWT", "data": nil})
	}
}
