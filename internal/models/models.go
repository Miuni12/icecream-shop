package models

import "time"

type Role string

const (
	RoleOwner   Role = "owner"
	RoleManager Role = "manager"
	RoleCashier Role = "cashier"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `json:"name"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"column:password;not null" json:"-"`
	Role         Role      `gorm:"not null" json:"role"`
	UpdatedAt    time.Time `gorm:"column:update_date" json:"updated_at"`
}

type Category struct {
	ID   uint   `gorm:"column:category_id;primaryKey" json:"id"`
	Name string `gorm:"uniqueIndex;not null" json:"name"`
}

type Item struct {
	ID                uint      `gorm:"column:item_id;primaryKey" json:"id"`
	Name              string    `gorm:"index;not null" json:"name"`
	CategoryID        uint      `gorm:"column:category_id;not null" json:"category_id"`
	Category          Category  `gorm:"foreignKey:CategoryID" json:"category"`
	Price             float64   `gorm:"not null" json:"price"`
	Quantity          int       `gorm:"not null" json:"quantity"`
	MinimumStockLevel int       `gorm:"column:minimum_stock_level;not null" json:"minimum_stock_level"`
	IsActive          bool      `gorm:"column:is_active;not null" json:"is_active"`
	UpdatedAt         time.Time `gorm:"column:update_date" json:"updated_at"`
}

type Customer struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	UpdatedAt time.Time `gorm:"column:update_date" json:"updated_at"`
}