package controllers

import (
	"errors"
	"net/http"
	"salonpro-backend/config"
	"salonpro-backend/models"
	"salonpro-backend/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RegisterInput struct {
	Email        string       `json:"email" binding:"required,email"`
	Phone        string       `json:"phone" binding:"required"`
	Name         string       `json:"name" binding:"required"` // Default name is email, can be changed later
	Password     string       `json:"password" binding:"required,min=8"`
	SalonName    string       `json:"salonName" binding:"required"`
	SalonAddress string       `json:"salonAddress"`
	WorkingHours models.JSONB `json:"workingHours"`
}

type LoginInput struct {
	Identifier string `json:"identifier" binding:"required"` // Can be email or phone
	Password   string `json:"password" binding:"required"`
}

// controllers/auth.go
func Register(c *gin.Context) {
	var input RegisterInput

	// Bind and validate input
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	// Check if email or phone already exists
	var existingUser models.User
	result := config.DB.Where("email = ? OR phone = ?", input.Email, input.Phone).First(&existingUser)

	if result.Error == nil {
		utils.RespondWithError(c, http.StatusConflict, "Email or phone already registered")
		return
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		utils.RespondWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Create new user
	newUser := models.User{
		Email:        input.Email,
		Phone:        input.Phone,
		Name:         input.Name,     // Default name is email, can be changed later
		Password:     input.Password, // Will be hashed in BeforeCreate hook
		SalonName:    input.SalonName,
		SalonAddress: input.SalonAddress,
		WorkingHours: input.WorkingHours,
	}

	// Set default working hours if not provided
	if newUser.WorkingHours == nil {
		newUser.WorkingHours = models.JSONB{
			"monday":    map[string]interface{}{"open": "09:00", "close": "20:00", "closed": false},
			"tuesday":   map[string]interface{}{"open": "09:00", "close": "20:00", "closed": false},
			"wednesday": map[string]interface{}{"open": "09:00", "close": "20:00", "closed": false},
			"thursday":  map[string]interface{}{"open": "09:00", "close": "20:00", "closed": false},
			"friday":    map[string]interface{}{"open": "09:00", "close": "20:00", "closed": false},
			"saturday":  map[string]interface{}{"open": "09:00", "close": "21:00", "closed": false},
			"sunday":    map[string]interface{}{"open": "10:00", "close": "19:00", "closed": true},
		}
	}

	// Create user in database
	if err := config.DB.Create(&newUser).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate token
	token, err := utils.GenerateToken(newUser.ID.String(), newUser.ID.String())
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}
	expiryHours := 24
	maxAge := expiryHours * 3600

	c.SetCookie(
		"token",
		token,
		maxAge,
		"/",
		"",
		true,
		true,
	)

	// Return response without password
	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful",
		"token":   token,
		"user": gin.H{
			"id":        newUser.ID,
			"email":     newUser.Email,
			"phone":     newUser.Phone,
			"salonName": newUser.SalonName,
		},
	})
}

func Login(c *gin.Context) {
	var input LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Clean identifier
	identifier := strings.TrimSpace(input.Identifier)

	// Determine if identifier is email or phone
	var user models.User
	query := config.DB.Where("email = ? OR phone = ?", identifier, identifier)
	result := query.First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			utils.RespondWithError(c, http.StatusUnauthorized, "Invalid credentials")
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "Database error")
		}
		return
	}

	// Check password
	if !utils.CheckPasswordHash(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID.String(), user.ID.String()) // Using user ID as salon ID
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Update last login
	now := time.Now()
	config.DB.Model(&user).Update("last_login", &now)

	expiryHours := 24
	maxAge := expiryHours * 3600

	c.SetCookie(
		"token",
		token,
		maxAge,
		"/",
		"",
		true,
		true,
	)

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"phone":     user.Phone,
			"salonName": user.SalonName,
		},
	})
}

func createDefaultReminderTemplates(salonID uuid.UUID) error {
	defaultTemplates := []models.ReminderTemplate{
		{
			SalonID: salonID,
			Type:    "birthday",
			Message: "Hi [CustomerName], SalonPro wishes you a very happy birthday! 🎉 Enjoy 20% off on your next visit this month!",
		},
		{
			SalonID: salonID,
			Type:    "anniversary",
			Message: "Hi [CustomerName], happy salon anniversary! 🎊 Thank you for being our valued customer. Here's 15% off your next service!",
		},
	}

	for _, template := range defaultTemplates {
		template.ID = uuid.New()
		if err := config.DB.Create(&template).Error; err != nil {
			return err
		}
	}
	return nil
}

func Me(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Return user info
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"name":      user.Name,
			"salonName": user.SalonName,
		},
	})
}
