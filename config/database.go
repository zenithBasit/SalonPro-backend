package config

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := os.Getenv("DB_URL")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect database")
	}

	// Optimize connection pool settings
	// sqlDB.SetMaxIdleConns(25)                 // Increase idle connections
	// sqlDB.SetMaxOpenConns(100)                // Increase max connections
	// sqlDB.SetConnMaxLifetime(5 * time.Minute) // Shorter lifetime
	// sqlDB.SetConnMaxIdleTime(time.Minute)     // Close idle connections faster

	DB = db
}
