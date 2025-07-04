package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"salonpro-backend/utils"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                    uuid.UUID `gorm:"type:uuid;primary_key"`
	Email                 string    `gorm:"uniqueIndex;not null"`
	Password              string    `gorm:"not null"`
	Name                  string    `gorm:"not null"`
	Phone                 string
	SalonName             string `gorm:"not null"`
	SalonAddress          string
	WorkingHours          JSONB `gorm:"type:jsonb;default:'{}'"`
	IsActive              bool  `gorm:"default:true"`
	BirthdayReminders     bool  `gorm:"default:true"`
	AnniversaryReminders  bool  `gorm:"default:true"`
	WhatsAppNotifications bool  `gorm:"default:false"`
	SMSNotifications      bool  `gorm:"default:false"`
	LastLogin             *time.Time
	Customers             []Customer         `gorm:"foreignKey:SalonID;references:ID"`
	Services              []Service          `gorm:"foreignKey:SalonID;references:ID"`
	Invoices              []Invoice          `gorm:"foreignKey:SalonID;references:ID"`
	ReminderTemplates     []ReminderTemplate `gorm:"foreignKey:SalonID;references:ID"`
	gorm.Model
}

// Initialize UUID before creating
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	hashed, err := utils.HashPassword(u.Password)
	if err != nil {
		return err
	}
	u.Password = hashed
	return
}

// Custom JSONB type for working hours
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &j)
}
