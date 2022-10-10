package server

import (
	jwtware "github.com/gofiber/jwt/v2"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

// Login get user and password
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

	if user != dbUser {
		// query user from here
		dbPass = "query-in-DB-here"
		// dbUser = ""
	}

	if pass != dbPass {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["identity"] = user
	claims["admin"] = true
	claims["exp"] = time.Now().Add(time.Hour * 24 * 30).Unix()

	t, err := token.SignedString([]byte(Opt.Server.JWTSecret))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Success login", "token": t})
}

// Protected protect routes
func Protected() func(*fiber.Ctx) error {
	if Opt.Server.NoAuthen {
		return nil
	}

	return jwtware.New(jwtware.Config{
		Filter:         nil,
		SuccessHandler: nil,
		ErrorHandler:   jwtError,
		SigningKey:     []byte(Opt.Server.JWTSecret),
		SigningKeys:    nil,
		SigningMethod:  "",
		ContextKey:     "",
		Claims:         nil,
		TokenLookup:    "",
		AuthScheme:     "Osmedeus",
	})
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
