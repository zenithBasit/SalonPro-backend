// models/reminder_log.go
package models

import (
	"time"
	
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReminderLog struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	SalonID      uuid.UUID `gorm:"type:uuid;index;not null"`
	CustomerID   uuid.UUID `gorm:"type:uuid;index;not null"`
	TemplateID   uuid.UUID `gorm:"type:uuid;index;not null"`
	Type         string    `gorm:"type:varchar(20)"` // birthday, anniversary
	Message      string    `gorm:"type:text"`
	Status       string    `gorm:"type:varchar(20)"` // sent, failed
	ErrorMessage string    `gorm:"type:text"`
	Channel      string    `gorm:"type:varchar(20)"` // whatsapp, sms
	SentAt       time.Time
	gorm.Model
}

func (r *ReminderLog) BeforeCreate(tx *gorm.DB) (err error) {
	r.ID = uuid.New()
	return
}