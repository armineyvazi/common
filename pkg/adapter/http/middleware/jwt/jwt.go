package jwt

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-errors/errors"
	"github.com/gofiber/fiber/v2"
)

func decodeToken(c *fiber.Ctx) (*jwt.MapClaims, error) {
	authHeader := c.Get(fiber.HeaderAuthorization)
	if authHeader == "" {
		return nil, errors.New("Authorization header is required")
	}

	token, _, err := new(jwt.Parser).ParseUnverified(authHeader[7:], jwt.MapClaims{})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("Invalid token")
	}

	if expiresAt, ok := claims["exp"]; ok && int64(expiresAt.(float64)) < time.Now().UTC().Unix() {
		return nil, errors.New("jwt is expired")
	}

	return &claims, nil
}

func New() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := decodeToken(c)
		if err == nil {
			sub := (*claims)["sub"]

			c.Locals("user_id", sub)
		}
		return c.Next()
	}
}
