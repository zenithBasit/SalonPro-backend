package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	SalonID      uuid.UUID `gorm:"type:uuid;index;not null"`
	Name         string    `gorm:"not null"`
	Description  string
	Price        float64 `gorm:"type:decimal(10,2);not null"`
	Duration     int     // in minutes
	Category     string  `gorm:"default:'General'"`
	IsActive     bool    `gorm:"default:true"`
	InvoiceItems []InvoiceItem
	gorm.Model
}
