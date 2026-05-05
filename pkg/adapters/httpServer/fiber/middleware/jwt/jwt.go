package jwt

import (
	"crypto/rsa"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/armineyvazi/common.git/pkg/adapters/errorUtil/appErr"
	"github.com/armineyvazi/common.git/pkg/ports"
)

func New(publicKeyString string) fiber.Handler {
	var publicKey *rsa.PublicKey
	var err error

	if publicKeyString != "" {
		publicKey, err = getPublicKey(publicKeyString)
		if err != nil {
			return func(c *fiber.Ctx) error {
				return appErr.HandleInternalError(err)
			}
		}
	}

	return func(c *fiber.Ctx) error {
		claims, err := decodeToken(c, publicKey)
		if err != nil {
			return appErr.NewUnauthenticatedErr(err)
		}

		return extractUserId(c, claims)
	}
}

// ----------------- individual middleware -----------------
type jwtToken struct {
	publicKey *rsa.PublicKey
}

func NewMiddleware(publicKeyString string) ports.JwtMiddleware {
	var publicKey *rsa.PublicKey
	var err error

	if publicKeyString != "" {
		publicKey, err = getPublicKey(publicKeyString)
		if err != nil {
			panic(err)
		}
	}
	return &jwtToken{
		publicKey: publicKey,
	}
}

func (j *jwtToken) TokenParser(c *ports.HttpContext) error {
	claims, err := decodeToken(c, nil)
	if err != nil {
		return appErr.NewUnauthenticatedErr(err)
	}

	return extractUserId(c, claims)
}

func (j *jwtToken) TokenValidation(c *ports.HttpContext) error {
	if j.publicKey == nil {
		return appErr.HandleInternalError(fmt.Errorf("could not load public key"))
	}

	claims, err := decodeToken(c, j.publicKey)
	if err != nil {
		return appErr.NewUnauthenticatedErr(err)
	}

	return extractUserId(c, claims)
}
