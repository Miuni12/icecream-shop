package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/miuni12/icecream-shop/internal/middleware"
	"github.com/miuni12/icecream-shop/internal/models"
)

type AuthHandler struct {
	DB      *gorm.DB
	Secret  string
	Revoked *middleware.Revoked
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	in.Email = strings.TrimSpace(in.Email)
	if in.Email == "" || in.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email and password are required")
	}

	var user models.User
	if err := h.DB.Where("email = ?", in.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid email or password")
		}
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)) != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid email or password")
	}

	now := time.Now()
	claims := middleware.Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Email,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(h.Secret))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"token": token, "user": user})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	if token, ok := c.Locals("token").(string); ok {
		h.Revoked.Add(token)
	}
	return c.JSON(fiber.Map{"message": "logged out"})
}