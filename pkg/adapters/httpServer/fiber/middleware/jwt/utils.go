package jwt

import (
	"crypto/rsa"
	"fmt"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

func decodeToken(c *fiber.Ctx, publicKey *rsa.PublicKey) (*jwt.MapClaims, error) {
	var token *jwt.Token
	var err error

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

	if publicKey != nil {
		token, err = jwt.Parse(tokenParts[1], func(token *jwt.Token) (interface{}, error) {
			// Check that the signing method is RSA
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return publicKey, nil
		})
		if err != nil {
			return nil, err
		}
		if !token.Valid {
			return nil, fmt.Errorf("invalid token")
		}

	} else {
		token, _, err = new(jwt.Parser).ParseUnverified(tokenParts[1], jwt.MapClaims{})
		if err != nil {
			return nil, err
		}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token")
	}

	if expiresAt, ok := claims["exp"]; ok && int64(expiresAt.(float64)) < time.Now().UTC().Unix() {
		return nil, fmt.Errorf("token is expired")
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
