package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/miuni12/icecream-shop/internal/models"
)

type ItemHandler struct{ DB *gorm.DB }

type itemInput struct {
	Name              *string  `json:"name"`
	CategoryID        *uint    `json:"category_id"`
	Price             *float64 `json:"price"`
	Quantity          *int     `json:"quantity"`
	MinimumStockLevel *int     `json:"minimum_stock_level"`
	IsActive          *bool    `json:"is_active"`
}

func (h *ItemHandler) List(c *fiber.Ctx) error {
	items := []models.Item{}
	q := h.DB.Preload("Category")
	if s := strings.TrimSpace(c.Query("search")); s != "" {
		q = q.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(s)+"%")
	}
	if err := q.Order("item_id").Find(&items).Error; err != nil {
		return err
	}
	return c.JSON(items)
}

func (h *ItemHandler) Get(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var item models.Item
	if err := h.DB.Preload("Category").First(&item, id).Error; err != nil {
		return dbError(err, "item")
	}
	return c.JSON(item)
}

func (h *ItemHandler) Create(c *fiber.Ctx) error {
	var in itemInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if in.Name == nil || strings.TrimSpace(*in.Name) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	if in.CategoryID == nil {
		return fiber.NewError(fiber.StatusBadRequest, "category_id is required")
	}
	if in.Price == nil {
		return fiber.NewError(fiber.StatusBadRequest, "price is required")
	}

	item := models.Item{
		Name:       strings.TrimSpace(*in.Name),
		CategoryID: *in.CategoryID,
		Price:      *in.Price,
		IsActive:   true,
	}
	if in.Quantity != nil {
		item.Quantity = *in.Quantity
	}
	if in.MinimumStockLevel != nil {
		item.MinimumStockLevel = *in.MinimumStockLevel
	}
	if in.IsActive != nil {
		item.IsActive = *in.IsActive
	}

	if err := validateItem(&item); err != nil {
		return err
	}
	if err := h.checkCategory(item.CategoryID); err != nil {
		return err
	}
	if err := h.DB.Omit("Category").Create(&item).Error; err != nil {
		return err
	}
	return h.respond(c, fiber.StatusCreated, item.ID)
}

// Update handles both PUT and PATCH: only the fields sent are changed.
func (h *ItemHandler) Update(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var item models.Item
	if err := h.DB.First(&item, id).Error; err != nil {
		return dbError(err, "item")
	}

	var in itemInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if in.Name != nil {
		item.Name = strings.TrimSpace(*in.Name)
	}
	if in.CategoryID != nil {
		item.CategoryID = *in.CategoryID
	}
	if in.Price != nil {
		item.Price = *in.Price
	}
	if in.Quantity != nil {
		item.Quantity = *in.Quantity
	}
	if in.MinimumStockLevel != nil {
		item.MinimumStockLevel = *in.MinimumStockLevel
	}
	if in.IsActive != nil {
		item.IsActive = *in.IsActive
	}

	if err := validateItem(&item); err != nil {
		return err
	}
	if err := h.checkCategory(item.CategoryID); err != nil {
		return err
	}
	if err := h.DB.Omit("Category").Save(&item).Error; err != nil {
		return err
	}
	return h.respond(c, fiber.StatusOK, item.ID)
}

func (h *ItemHandler) Delete(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	res := h.DB.Delete(&models.Item{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "item not found")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ItemHandler) respond(c *fiber.Ctx, status int, id uint) error {
	var item models.Item
	if err := h.DB.Preload("Category").First(&item, id).Error; err != nil {
		return err
	}
	return c.Status(status).JSON(item)
}

func (h *ItemHandler) checkCategory(id uint) error {
	var n int64
	if err := h.DB.Model(&models.Category{}).Where("category_id = ?", id).Count(&n).Error; err != nil {
		return err
	}
	if n == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "category does not exist")
	}
	return nil
}

func validateItem(item *models.Item) error {
	switch {
	case item.Name == "":
		return fiber.NewError(fiber.StatusBadRequest, "name cannot be empty")
	case item.Price < 0:
		return fiber.NewError(fiber.StatusBadRequest, "price cannot be negative")
	case item.Quantity < 0:
		return fiber.NewError(fiber.StatusBadRequest, "quantity cannot be negative")
	case item.MinimumStockLevel < 0:
		return fiber.NewError(fiber.StatusBadRequest, "minimum_stock_level cannot be negative")
	}
	return nil
}