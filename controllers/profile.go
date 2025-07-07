package controllers

import (
	"net/http"
	"salonpro-backend/config"
	"salonpro-backend/models"
	"salonpro-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UpdateProfileInput struct {
	SalonName    string `json:"salonName"`
	SalonAddress string `json:"salonAddress"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	// WorkingHours models.JSONB `json:"workingHours"` // or your working hours struct
}

func GetProfile(c *gin.Context) {
	// Get salon ID from context
	salonID, exists := c.Get("salonId")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Salon ID not found")
		return
	}
	salonUUID, err := uuid.Parse(salonID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid salon ID format")
		return
	}

	// --- Fetch user (salon profile + working hours + notification settings) ---
	var user models.User
    var salon models.Salon
	if err := config.DB.First(&user, "id = ?", salonUUID).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "User not found")
		return
	}

	// --- Fetch reminder templates ---
	var reminderTemplates []models.ReminderTemplate
	if err := config.DB.Where("salon_id = ?", salonUUID).Find(&reminderTemplates).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch reminder templates")
		return
	}

	// Extract messages
	var birthdayMessage, anniversaryMessage string
	for _, tmpl := range reminderTemplates {
		switch tmpl.Type {
		case "birthday":
			birthdayMessage = tmpl.Message
		case "anniversary":
			anniversaryMessage = tmpl.Message
		}
	}

	// --- Return combined response ---
	c.JSON(http.StatusOK, gin.H{
		"salonProfile": gin.H{
			"salonName":    salon.Name,
			"address":      salon.Address,
			"phone":        user.Phone,
			"email":        user.Email,
			"workingHours": salon.WorkingHours,
		},
		"messageTemplates": gin.H{
			"birthday":    birthdayMessage,
			"anniversary": anniversaryMessage,
		},
		"notifications": gin.H{
			"birthdayReminders":     salon.BirthdayReminders,
			"anniversaryReminders":  salon.AnniversaryReminders,
			"whatsAppNotifications": salon.WhatsAppNotifications,
			"smsNotifications":      salon.SMSNotifications,
		},
	})
}

type UpdateSalonProfileInput struct {
	SalonName string `json:"salonName"`
	Address   string `json:"salonAddress"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
}

func UpdateSalonProfile(c *gin.Context) {
	salonID, exists := c.Get("salonId")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Salon ID not found")
		return
	}
	salonUUID, err := uuid.Parse(salonID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid salon ID")
		return
	}

	var input UpdateSalonProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	if err := config.DB.Model(&models.User{}).
		Where("id = ?", salonUUID).
		Updates(map[string]interface{}{
			"salon_name":    input.SalonName,
			"salon_address": input.Address,
			"phone":         input.Phone,
			"email":         input.Email,
		}).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

type UpdateWorkingHoursInput struct {
	WorkingHours models.JSONB `json:"workingHours"`
}

func UpdateWorkingHours(c *gin.Context) {
	salonID, exists := c.Get("salonId")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Salon ID not found")
		return
	}
	salonUUID, err := uuid.Parse(salonID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid salon ID")
		return
	}

	var input UpdateWorkingHoursInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	if err := config.DB.Model(&models.User{}).
		Where("id = ?", salonUUID).
		Update("working_hours", input.WorkingHours).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update working hours")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Working hours updated successfully"})
}

type UpdateTemplatesInput struct {
	BirthdayMessage    string `json:"birthday" form:"birthday" binding:"omitempty"`
	AnniversaryMessage string `json:"anniversary" form:"anniversary" binding:"omitempty"`
}

func UpdateReminderTemplates(c *gin.Context) {
	salonID, exists := c.Get("salonId")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Salon ID not found")
		return
	}
	salonUUID, err := uuid.Parse(salonID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid salon ID")
		return
	}

	var input UpdateTemplatesInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	updates := []struct {
		Type    string
		Message string
	}{
		{"birthday", input.BirthdayMessage},
		{"anniversary", input.AnniversaryMessage},
	}

	for _, u := range updates {
		if err := config.DB.Model(&models.ReminderTemplate{}).
			Where("salon_id = ? AND type = ?", salonUUID, u.Type).
			Update("message", u.Message).Error; err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update "+u.Type+" template")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Templates updated successfully"})
}

type UpdateNotificationsInput struct {
	BirthdayReminders     bool `json:"birthdayReminders"`
	AnniversaryReminders  bool `json:"anniversaryReminders"`
	WhatsAppNotifications bool `json:"whatsAppNotifications"`
	SMSNotifications      bool `json:"smsNotifications"`
}

func UpdateNotifications(c *gin.Context) {
	salonID, exists := c.Get("salonId")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Salon ID not found")
		return
	}
	salonUUID, err := uuid.Parse(salonID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid salon ID")
		return
	}

	var input UpdateNotificationsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	if err := config.DB.Model(&models.User{}).
		Where("id = ?", salonUUID).
		Updates(map[string]interface{}{
			"birthday_reminders":      input.BirthdayReminders,
			"anniversary_reminders":   input.AnniversaryReminders,
			"whats_app_notifications": input.WhatsAppNotifications,
			"sms_notifications":       input.SMSNotifications,
		}).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update notifications")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification settings updated successfully"})
}
