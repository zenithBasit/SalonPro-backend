package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReminderTemplate struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key"`
	SalonID  uuid.UUID `gorm:"type:uuid;index;not null"`
	Type     string    `gorm:"type:enum('birthday','anniversary');not null"`
	Message  string    `gorm:"type:text;not null"`
	IsActive bool      `gorm:"default:true"`
	gorm.Model
}
