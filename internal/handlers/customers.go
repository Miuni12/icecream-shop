package handlers

import (
	"net/mail"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/miuni12/icecream-shop/internal/models"
)

type CustomerHandler struct{ DB *gorm.DB }

type customerInput struct {
	Name    *string `json:"name"`
	Email   *string `json:"email"`
	Phone   *string `json:"phone"`
	Address *string `json:"address"`
}

func (h *CustomerHandler) List(c *fiber.Ctx) error {
	customers := []models.Customer{}
	q := h.DB
	if s := strings.TrimSpace(c.Query("search")); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		q = q.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", like, like)
	}
	if err := q.Order("id").Find(&customers).Error; err != nil {
		return err
	}
	return c.JSON(customers)
}

func (h *CustomerHandler) Get(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var customer models.Customer
	if err := h.DB.First(&customer, id).Error; err != nil {
		return dbError(err, "customer")
	}
	return c.JSON(customer)
}

func (h *CustomerHandler) Create(c *fiber.Ctx) error {
	var in customerInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if in.Name == nil || strings.TrimSpace(*in.Name) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	if in.Email == nil || strings.TrimSpace(*in.Email) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email is required")
	}

	customer := models.Customer{
		Name:  strings.TrimSpace(*in.Name),
		Email: strings.TrimSpace(*in.Email),
	}
	if in.Phone != nil {
		customer.Phone = strings.TrimSpace(*in.Phone)
	}
	if in.Address != nil {
		customer.Address = strings.TrimSpace(*in.Address)
	}

	if err := h.validate(&customer, 0); err != nil {
		return err
	}
	if err := h.DB.Create(&customer).Error; err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(customer)
}

// Update handles both PUT and PATCH: only the fields sent are changed.
func (h *CustomerHandler) Update(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var customer models.Customer
	if err := h.DB.First(&customer, id).Error; err != nil {
		return dbError(err, "customer")
	}

	var in customerInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if in.Name != nil {
		customer.Name = strings.TrimSpace(*in.Name)
	}
	if in.Email != nil {
		customer.Email = strings.TrimSpace(*in.Email)
	}
	if in.Phone != nil {
		customer.Phone = strings.TrimSpace(*in.Phone)
	}
	if in.Address != nil {
		customer.Address = strings.TrimSpace(*in.Address)
	}

	if err := h.validate(&customer, customer.ID); err != nil {
		return err
	}
	if err := h.DB.Save(&customer).Error; err != nil {
		return err
	}
	return c.JSON(customer)
}

func (h *CustomerHandler) Delete(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	res := h.DB.Delete(&models.Customer{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "customer not found")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CustomerHandler) validate(cu *models.Customer, exceptID uint) error {
	if cu.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name cannot be empty")
	}
	addr, err := mail.ParseAddress(cu.Email)
	if err != nil || addr.Address != cu.Email {
		return fiber.NewError(fiber.StatusBadRequest, "email is not valid")
	}
	var n int64
	if err := h.DB.Model(&models.Customer{}).Where("email = ? AND id <> ?", cu.Email, exceptID).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return fiber.NewError(fiber.StatusConflict, "a customer with this email already exists")
	}
	return nil
}