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
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Email    string    `gorm:"uniqueIndex;not null"`
	Password string    `gorm:"not null"`
	Name     string    `gorm:"not null"`
	Phone    string

	Role    string    `gorm:"type:varchar(20);not null"` // 'owner' or 'employee'
	SalonID uuid.UUID `gorm:"type:uuid;index;not null"`

	Salon Salon `gorm:"foreignKey:SalonID"`

	LastLogin *time.Time
	IsActive  bool `gorm:"default:true"`

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
