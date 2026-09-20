package jwt

import (
	"crypto/rsa"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func decodeToken(c *fiber.Ctx, publicKey *rsa.PublicKey) (*jwt.MapClaims, error) {
	authHeader := c.Get(fiber.HeaderAuthorization)
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header is required")
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) < 2 {
		return nil, fmt.Errorf("token is invalid")
	}

	if tokenParts[0] != "Bearer" {
		return nil, fmt.Errorf("token is not Bearer type")
	}

	claims := jwt.MapClaims{}

	if publicKey != nil {
		_, err := jwt.ParseWithClaims(tokenParts[1], claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return publicKey, nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		token, _, err := jwt.NewParser().ParseUnverified(tokenParts[1], claims)
		if err != nil {
			return nil, err
		}
		mc, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("invalid token claims")
		}
		claims = mc

		// Manually check expiry since signature is not verified.
		if exp, err := claims.GetExpirationTime(); err == nil && exp != nil {
			if exp.Before(time.Now()) {
				return nil, fmt.Errorf("token is expired")
			}
		}
	}

	return &claims, nil
}

func extractUserId(c *fiber.Ctx, claims *jwt.MapClaims) error {
	sub := (*claims)["sub"]
	c.Locals("user_id", sub)
	if sub == nil {
		id := (*claims)["id"]
		c.Locals("guest_id", id)
	}
	return c.Next()
}
