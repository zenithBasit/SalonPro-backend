package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Customer struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	SalonID     uuid.UUID `gorm:"type:uuid;index;not null"`
	Name        string    `gorm:"not null"`
	Phone       string    `gorm:"not null;index:,unique,composite:salon_phone"`
	Email       string
	Birthday    *time.Time
	Anniversary *time.Time
	Notes       string
	TotalVisits int     `gorm:"default:0"`
	TotalSpent  float64 `gorm:"type:decimal(10,2);default:0.0"`
	LastVisit   *time.Time
	IsActive    bool `gorm:"default:true"`
	Invoices    []Invoice
	gorm.Model
}
