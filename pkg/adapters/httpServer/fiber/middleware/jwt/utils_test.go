package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/valyala/fasthttp"
)

func generateRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	return key
}

func makeToken(t *testing.T, key *rsa.PrivateKey, claims gojwt.MapClaims) string {
	t.Helper()
	tok := gojwt.NewWithClaims(gojwt.SigningMethodRS256, claims)
	signed, err := tok.SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func newCtxWithAuth(app *fiber.App, authHeader string) *fiber.Ctx {
	fctx := &fasthttp.RequestCtx{}
	fctx.Request.Header.SetMethod("GET")
	fctx.Request.SetRequestURI("/")
	if authHeader != "" {
		fctx.Request.Header.Set(fiber.HeaderAuthorization, authHeader)
	}
	return app.AcquireCtx(fctx)
}

func TestDecodeToken_MissingHeader(t *testing.T) {
	app := fiber.New()
	ctx := newCtxWithAuth(app, "")
	_, err := decodeToken(ctx, nil)
	if err == nil {
		t.Error("expected error for missing Authorization header")
	}
}

func TestDecodeToken_InvalidFormat(t *testing.T) {
	app := fiber.New()
	ctx := newCtxWithAuth(app, "invalidtoken")
	_, err := decodeToken(ctx, nil)
	if err == nil {
		t.Error("expected error for invalid token format (no space)")
	}
}

func TestDecodeToken_NotBearer(t *testing.T) {
	app := fiber.New()
	ctx := newCtxWithAuth(app, "Basic abc123")
	_, err := decodeToken(ctx, nil)
	if err == nil {
		t.Error("expected error for non-Bearer scheme")
	}
}

func TestDecodeToken_UnverifiedValid(t *testing.T) {
	key := generateRSAKey(t)
	tokenStr := makeToken(t, key, gojwt.MapClaims{
		"sub": "user-123",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
	})

	app := fiber.New()
	ctx := newCtxWithAuth(app, "Bearer "+tokenStr)

	result, err := decodeToken(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if (*result)["sub"] != "user-123" {
		t.Errorf("sub: got %v, want user-123", (*result)["sub"])
	}
}

func TestDecodeToken_UnverifiedExpired(t *testing.T) {
	key := generateRSAKey(t)
	tokenStr := makeToken(t, key, gojwt.MapClaims{
		"sub": "user-expired",
		"exp": float64(time.Now().Add(-time.Hour).Unix()),
	})

	app := fiber.New()
	ctx := newCtxWithAuth(app, "Bearer "+tokenStr)

	_, err := decodeToken(ctx, nil)
	if err == nil {
		t.Error("expected error for expired token in unverified mode")
	}
}

func TestDecodeToken_WithPublicKey_Valid(t *testing.T) {
	key := generateRSAKey(t)
	tokenStr := makeToken(t, key, gojwt.MapClaims{
		"sub": "user-456",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
	})

	app := fiber.New()
	ctx := newCtxWithAuth(app, "Bearer "+tokenStr)

	result, err := decodeToken(ctx, &key.PublicKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if (*result)["sub"] != "user-456" {
		t.Errorf("sub: got %v, want user-456", (*result)["sub"])
	}
}

func TestDecodeToken_WithPublicKey_WrongKey(t *testing.T) {
	signingKey := generateRSAKey(t)
	wrongKey := generateRSAKey(t)
	tokenStr := makeToken(t, signingKey, gojwt.MapClaims{
		"sub": "user-789",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
	})

	app := fiber.New()
	ctx := newCtxWithAuth(app, "Bearer "+tokenStr)

	_, err := decodeToken(ctx, &wrongKey.PublicKey)
	if err == nil {
		t.Error("expected error when verifying with wrong public key")
	}
}

func TestGetPublicKey_InvalidPEM(t *testing.T) {
	_, err := getPublicKey("not a pem block")
	if err == nil {
		t.Error("expected error for invalid PEM")
	}
}

func TestGetPublicKey_WrongBlockType(t *testing.T) {
	pem := "-----BEGIN CERTIFICATE-----\nYWJj\n-----END CERTIFICATE-----\n"
	_, err := getPublicKey(pem)
	if err == nil {
		t.Error("expected error for wrong PEM block type")
	}
}
