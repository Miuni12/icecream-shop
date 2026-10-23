package routes

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/miuni12/icecream-shop/internal/handlers"
	"github.com/miuni12/icecream-shop/internal/middleware"
	"github.com/miuni12/icecream-shop/internal/models"
)

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	msg := "internal server error"
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
		msg = e.Message
	}
	return c.Status(code).JSON(fiber.Map{"error": msg})
}

func NewApp(db *gorm.DB, jwtSecret string) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: errorHandler})

	revoked := middleware.NewRevoked()
	authH := &handlers.AuthHandler{DB: db, Secret: jwtSecret, Revoked: revoked}
	itemH := &handlers.ItemHandler{DB: db}
	customerH := &handlers.CustomerHandler{DB: db}

	protect := middleware.Auth(jwtSecret, revoked)
	owner, manager, cashier := models.RoleOwner, models.RoleManager, models.RoleCashier

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api")

	// Authentication
	api.Post("/auth/login", authH.Login)
	api.Post("/auth/logout", protect, authH.Logout)

	// Inventory: everyone views; owner+manager add; everyone edits; only owner deletes
	api.Get("/items", protect, itemH.List)
	api.Get("/items/:id", protect, itemH.Get)
	api.Post("/items", protect, middleware.RequireRole(owner, manager), itemH.Create)
	api.Put("/items/:id", protect, middleware.RequireRole(owner, manager, cashier), itemH.Update)
	api.Patch("/items/:id", protect, middleware.RequireRole(owner, manager, cashier), itemH.Update)
	api.Delete("/items/:id", protect, middleware.RequireRole(owner), itemH.Delete)

	// Customers: everyone views/adds/edits; only owner deletes
	api.Get("/customers", protect, customerH.List)
	api.Get("/customers/:id", protect, customerH.Get)
	api.Post("/customers", protect, middleware.RequireRole(owner, manager, cashier), customerH.Create)
	api.Put("/customers/:id", protect, middleware.RequireRole(owner, manager, cashier), customerH.Update)
	api.Patch("/customers/:id", protect, middleware.RequireRole(owner, manager, cashier), customerH.Update)
	api.Delete("/customers/:id", protect, middleware.RequireRole(owner), customerH.Delete)

	return app
}