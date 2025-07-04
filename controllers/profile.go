package controllers

import (
	"net/http"
	"salonpro-backend/config"
	"salonpro-backend/models"
	"salonpro-backend/utils"

	"github.com/gin-gonic/gin"
)

type UpdateProfileInput struct {
	SalonName    string       `json:"salonName"`
	SalonAddress string       `json:"salonAddress"`
	Phone        string       `json:"phone"`
	Email        string       `json:"email"`
	// WorkingHours models.JSONB `json:"workingHours"` // or your working hours struct
}

func GetProfile(c *gin.Context) {
    userID, exists := c.Get("userId")
    if !exists {
        utils.RespondWithError(c, http.StatusUnauthorized, "User not found")
        return
    }

    var user models.User
    if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
        utils.RespondWithError(c, http.StatusNotFound, "User not found")
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "salonName":    user.SalonName,
        "salonAddress": user.SalonAddress,
        "phone":        user.Phone,
        "email":        user.Email,
        "workingHours": user.WorkingHours,
        "birthdayReminders":    user.BirthdayReminders,
        "anniversaryReminders": user.AnniversaryReminders,
        "whatsAppNotifications": user.WhatsAppNotifications,
        "smsNotifications":      user.SMSNotifications,
    })
}

func UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "User not found")
		return
	}

	var input UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input")
		return
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "User not found")
		return
	}

	// Update fields
	user.SalonName = input.SalonName
	user.SalonAddress = input.SalonAddress
	user.Phone = input.Phone
	user.Email = input.Email
	// user.WorkingHours = input.WorkingHours

	if err := config.DB.Save(&user).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated"})
}

func UpdateWorkingHours(c *gin.Context) {
    userID, exists := c.Get("userId")
    if !exists {
        utils.RespondWithError(c, http.StatusUnauthorized, "User not found")
        return
    }

    var input struct {
        WorkingHours models.JSONB `json:"workingHours"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        utils.RespondWithError(c, http.StatusBadRequest, "Invalid input")
        return
    }

    if err := config.DB.Model(&models.User{}).Where("id = ?", userID).
        Update("working_hours", input.WorkingHours).Error; err != nil {
        utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update working hours")
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Working hours updated"})
}

func UpdateNotificationSettings(c *gin.Context) {
    userID, exists := c.Get("userId")
    if !exists {
        utils.RespondWithError(c, http.StatusUnauthorized, "User not found")
        return
    }

    var input struct {
        BirthdayReminders   bool `json:"birthdayReminders"`
        AnniversaryReminders bool `json:"anniversaryReminders"`
        WhatsAppNotifications bool `json:"whatsAppNotifications"`
        SMSNotifications      bool `json:"smsNotifications"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        utils.RespondWithError(c, http.StatusBadRequest, "Invalid input")
        return
    }

    // Save these fields in your user or a separate settings table as needed
    if err := config.DB.Model(&models.User{}).Where("id = ?", userID).
        Updates(map[string]interface{}{
            "birthday_reminders":    input.BirthdayReminders,
            "anniversary_reminders": input.AnniversaryReminders,
            "whatsapp_notifications": input.WhatsAppNotifications,
            "sms_notifications":      input.SMSNotifications,
        }).Error; err != nil {
        utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update notification settings")
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Notification settings updated"})
}


type UpdateReminderSettingInput struct {
    Type     string `json:"type" binding:"required,oneof=birthday anniversary"`
    IsActive *bool  `json:"isActive"`
    Message  *string `json:"message"`
}

func UpdateReminderSetting(c *gin.Context) {
    salonID, exists := c.Get("salonId")
    if !exists {
        utils.RespondWithError(c, http.StatusUnauthorized, "Salon ID not found")
        return
    }

    var input UpdateReminderSettingInput
    if err := c.ShouldBindJSON(&input); err != nil {
        utils.RespondWithError(c, http.StatusBadRequest, "Invalid input")
        return
    }

    var template models.ReminderTemplate
    if err := config.DB.Where("salon_id = ? AND type = ?", salonID, input.Type).First(&template).Error; err != nil {
        utils.RespondWithError(c, http.StatusNotFound, "Reminder template not found")
        return
    }

    if input.IsActive != nil {
        template.IsActive = *input.IsActive
    }
    if input.Message != nil {
        template.Message = *input.Message
    }

    if err := config.DB.Save(&template).Error; err != nil {
        utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update reminder setting")
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Reminder setting updated"})
}

func GetReminderSettings(c *gin.Context) {
    salonID, exists := c.Get("salonId")
    if !exists {
        utils.RespondWithError(c, http.StatusUnauthorized, "Salon ID not found")
        return
    }

    var templates []models.ReminderTemplate
    if err := config.DB.Where("salon_id = ?", salonID).Find(&templates).Error; err != nil {
        utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch reminder settings")
        return
    }

    settings := gin.H{}
    for _, t := range templates {
        settings[t.Type+"_reminder"] = t.IsActive
        settings[t.Type+"_message"] = t.Message
    }

    c.JSON(http.StatusOK, settings)
}