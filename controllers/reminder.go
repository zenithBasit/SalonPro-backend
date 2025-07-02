// controllers/reminder.go
package controllers

import (
	"errors"
	"net/http"

	"salonpro-backend/config"
	"salonpro-backend/models"
	"salonpro-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateReminderTemplateInput defines the expected JSON structure
type CreateReminderTemplateInput struct {
	Type    string `json:"type" binding:"required,oneof=birthday anniversary"`
	Message string `json:"message" binding:"required"`
}

// UpdateReminderTemplateInput defines the expected JSON structure
type UpdateReminderTemplateInput struct {
	Type     *string `json:"type" binding:"omitempty,oneof=birthday anniversary"`
	Message  *string `json:"message"`
	IsActive *bool   `json:"isActive"`
}

// CreateReminderTemplate creates a new reminder template
func CreateReminderTemplate(c *gin.Context) {
	salonID, exists := c.Get("salonId")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Salon ID not found in context")
		return
	}

	salonUUID, err := uuid.Parse(salonID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid salon ID format")
		return
	}

	var input CreateReminderTemplateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	// Check if template type already exists for this salon
	var existingTemplate models.ReminderTemplate
	if err := config.DB.Where("salon_id = ? AND type = ?", salonUUID, input.Type).
		First(&existingTemplate).Error; err == nil {
		utils.RespondWithError(c, http.StatusConflict, "Template for this type already exists")
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.RespondWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Create new template
	template := models.ReminderTemplate{
		ID:       uuid.New(),
		SalonID:  salonUUID,
		Type:     input.Type,
		Message:  input.Message,
		IsActive: true,
	}

	if err := config.DB.Create(&template).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to create template")
		return
	}

	c.JSON(http.StatusCreated, template)
}

// GetReminderTemplates retrieves all reminder templates for the salon
func GetReminderTemplates(c *gin.Context) {
	salonID, exists := c.Get("salonId")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Salon ID not found in context")
		return
	}

	salonUUID, err := uuid.Parse(salonID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid salon ID format")
		return
	}

	var templates []models.ReminderTemplate
	if err := config.DB.Where("salon_id = ?", salonUUID).Find(&templates).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to retrieve templates")
		return
	}

	c.JSON(http.StatusOK, templates)
}

// GetReminderTemplate retrieves a specific template by ID
func GetReminderTemplate(c *gin.Context) {
	salonID, exists := c.Get("salonId")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Salon ID not found in context")
		return
	}

	salonUUID, err := uuid.Parse(salonID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid salon ID format")
		return
	}

	templateID := c.Param("id")
	templateUUID, err := uuid.Parse(templateID)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid template ID format")
		return
	}

	var template models.ReminderTemplate
	if err := config.DB.Where("salon_id = ? AND id = ?", salonUUID, templateUUID).
		First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.RespondWithError(c, http.StatusNotFound, "Template not found")
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "Database error")
		}
		return
	}

	c.JSON(http.StatusOK, template)
}

// UpdateReminderTemplate updates an existing template
func UpdateReminderTemplate(c *gin.Context) {
	salonID, exists := c.Get("salonId")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Salon ID not found in context")
		return
	}

	salonUUID, err := uuid.Parse(salonID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid salon ID format")
		return
	}

	templateID := c.Param("id")
	templateUUID, err := uuid.Parse(templateID)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid template ID format")
		return
	}

	var input UpdateReminderTemplateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	// Retrieve existing template
	var template models.ReminderTemplate
	if err := config.DB.Where("salon_id = ? AND id = ?", salonUUID, templateUUID).
		First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.RespondWithError(c, http.StatusNotFound, "Template not found")
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "Database error")
		}
		return
	}

	// If changing type, check for conflict
	if input.Type != nil && *input.Type != template.Type {
		var existingTemplate models.ReminderTemplate
		if err := config.DB.Where("salon_id = ? AND type = ?", salonUUID, *input.Type).
			First(&existingTemplate).Error; err == nil {
			utils.RespondWithError(c, http.StatusConflict, "Template for this type already exists")
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			utils.RespondWithError(c, http.StatusInternalServerError, "Database error")
			return
		}
		template.Type = *input.Type
	}

	// Update other fields
	if input.Message != nil {
		template.Message = *input.Message
	}
	if input.IsActive != nil {
		template.IsActive = *input.IsActive
	}

	if err := config.DB.Save(&template).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update template")
		return
	}

	c.JSON(http.StatusOK, template)
}

// DeleteReminderTemplate deletes a template
func DeleteReminderTemplate(c *gin.Context) {
	salonID, exists := c.Get("salonId")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Salon ID not found in context")
		return
	}

	salonUUID, err := uuid.Parse(salonID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid salon ID format")
		return
	}

	templateID := c.Param("id")
	templateUUID, err := uuid.Parse(templateID)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid template ID format")
		return
	}

	result := config.DB.Where("salon_id = ? AND id = ?", salonUUID, templateUUID).
		Delete(&models.ReminderTemplate{})

	if result.Error != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to delete template")
		return
	}

	if result.RowsAffected == 0 {
		utils.RespondWithError(c, http.StatusNotFound, "Template not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully"})
}
