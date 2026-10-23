package database

import (
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/miuni12/icecream-shop/internal/models"
)

func Connect(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&models.User{}, &models.Category{}, &models.Item{}, &models.Customer{}); err != nil {
		return nil, err
	}
	return db, nil
}

// Seed is idempotent: running it repeatedly does not create duplicates.
func Seed(db *gorm.DB) error {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	users := []models.User{
		{Name: "Olivia Owner", Email: "owner@shop.com", Role: models.RoleOwner},
		{Name: "Marcus Manager", Email: "manager@shop.com", Role: models.RoleManager},
		{Name: "Casey Cashier", Email: "cashier@shop.com", Role: models.RoleCashier},
	}
	for _, u := range users {
		u.PasswordHash = string(hash)
		if err := db.Where("email = ?", u.Email).FirstOrCreate(&u).Error; err != nil {
			return err
		}
	}

	categoryIDs := map[string]uint{}
	for _, name := range []string{"Chocolate", "Vanilla", "Strawberry"} {
		cat := models.Category{Name: name}
		if err := db.Where("name = ?", name).FirstOrCreate(&cat).Error; err != nil {
			return err
		}
		categoryIDs[name] = cat.ID
	}

	items := []models.Item{
		{Name: "Chocolate Scoop", CategoryID: categoryIDs["Chocolate"], Price: 2.50, Quantity: 100, MinimumStockLevel: 10, IsActive: true},
		{Name: "Vanilla Scoop", CategoryID: categoryIDs["Vanilla"], Price: 2.00, Quantity: 120, MinimumStockLevel: 10, IsActive: true},
		{Name: "Strawberry Scoop", CategoryID: categoryIDs["Strawberry"], Price: 2.75, Quantity: 80, MinimumStockLevel: 10, IsActive: true},
	}
	for _, it := range items {
		if err := db.Where("name = ?", it.Name).FirstOrCreate(&it).Error; err != nil {
			return err
		}
	}

	customer := models.Customer{Name: "Sample Customer", Email: "customer@example.com", Phone: "0771234567", Address: "Colombo"}
	return db.Where("email = ?", customer.Email).FirstOrCreate(&customer).Error
}