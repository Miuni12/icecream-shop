package middleware

import (
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/miuni12/icecream-shop/internal/models"
)

type Claims struct {
	UserID uint        `json:"uid"`
	Role   models.Role `json:"role"`
	jwt.RegisteredClaims
}

// Revoked is an in-memory list of logged-out tokens.
type Revoked struct {
	mu     sync.RWMutex
	tokens map[string]struct{}
}

func NewRevoked() *Revoked { return &Revoked{tokens: map[string]struct{}{}} }

func (r *Revoked) Add(token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[token] = struct{}{}
}

func (r *Revoked) Has(token string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.tokens[token]
	return ok
}

// Auth verifies the Bearer JWT and stores the claims for later handlers.
func Auth(secret string, revoked *Revoked) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return fiber.NewError(fiber.StatusUnauthorized, "missing or invalid Authorization header")
		}
		tokenStr := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !token.Valid || revoked.Has(tokenStr) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}

		c.Locals("claims", claims)
		c.Locals("token", tokenStr)
		return c.Next()
	}
}

// RequireRole allows the request only if the user's role is in the list.
func RequireRole(roles ...models.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("claims").(*Claims)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "not authenticated")
		}
		for _, r := range roles {
			if claims.Role == r {
				return c.Next()
			}
		}
		return fiber.NewError(fiber.StatusForbidden, "you do not have permission to perform this action")
	}
}