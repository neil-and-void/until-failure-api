package middleware

import (
	"strings"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/gofiber/fiber/v2"
)

const SESSION_CLAIMS = "SESSION_CLAIMS"

type Middleware struct {
	Clerk clerk.Client
}

func (m Middleware) JWTAuthMiddleware(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	sessClaims, err := m.Clerk.VerifyToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	c.Locals("SESSION_CLAIMS", &sessClaims)

	return c.Next()
}

func (m Middleware) MockAuthMiddleware(c *fiber.Ctx) error {
	claims := clerk.SessionClaims{Claims: jwt.Claims{Subject: "user_2bhnSTV705sfQLZrf6Id3HXUR40"}}

	c.Locals("SESSION_CLAIMS", &claims)

	return c.Next()
}

func GetSession(c *fiber.Ctx) {

}
