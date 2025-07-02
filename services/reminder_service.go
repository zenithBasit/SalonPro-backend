// services/reminder_service.go
package services

import (
	"fmt"
	"log"
	"os"
	"salonpro-backend/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
	"gorm.io/gorm"
)

type ReminderService struct {
	db     *gorm.DB
	client *twilio.RestClient
}

func NewReminderService(db *gorm.DB) *ReminderService {
	// Initialize Twilio client
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")

	return &ReminderService{
		db: db,
		client: twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: accountSid,
			Password: authToken,
		}),
	}
}

func (s *ReminderService) StartScheduler() {
	c := cron.New()

	// Run every day at 9 AM
	s.SendDailyReminders()

	c.Start()
	log.Println("Reminder scheduler started")
}

func (s *ReminderService) SendDailyReminders() {
	log.Println("Starting daily reminder processing...")

	// Get all active salons
	var salons []models.User
	if err := s.db.Find(&salons, "is_active = ?", true).Error; err != nil {
		log.Printf("Failed to fetch salons: %v", err)
		return
	}

	for _, salon := range salons {
		s.ProcessSalonReminders(salon.ID)
	}

	log.Println("Daily reminder processing completed")
}

func (s *ReminderService) ProcessSalonReminders(salonID uuid.UUID) {
	// Get upcoming birthdays (7 days from now)
	birthdayCustomers, err := s.getUpcomingCustomers(salonID, "birthday")
	if err != nil {
		log.Printf("Salon %s: Failed to get birthday customers: %v", salonID, err)
	} else {
		s.sendReminders(salonID, birthdayCustomers, "birthday")
	}

	// Get upcoming anniversaries (7 days from now)
	anniversaryCustomers, err := s.getUpcomingCustomers(salonID, "anniversary")
	if err != nil {
		log.Printf("Salon %s: Failed to get anniversary customers: %v", salonID, err)
	} else {
		s.sendReminders(salonID, anniversaryCustomers, "anniversary")
	}
}

func (s *ReminderService) getUpcomingCustomers(salonID uuid.UUID, eventType string) ([]models.Customer, error) {
	now := time.Now()
	startDate := now.AddDate(0, 0, 7) // 7 days from now

	var customers []models.Customer
	var field string
	switch eventType {
	case "birthday":
		field = "birthday"
	case "anniversary":
		field = "anniversary"
	default:
		return nil, fmt.Errorf("invalid event type: %s", eventType)
	}

	// Query for customers with event in the next 7 days
	query := fmt.Sprintf(`
		SELECT * FROM customers
		WHERE salon_id = ?
		AND is_active = true
		AND %s IS NOT NULL
		AND EXTRACT(MONTH FROM %s) = ?
		AND EXTRACT(DAY FROM %s) BETWEEN ? AND ?
	`, field, field, field)

	err := s.db.Raw(query, salonID, startDate.Month(), now.Day(), startDate.Day()).Scan(&customers).Error
	return customers, err
}

func (s *ReminderService) sendReminders(salonID uuid.UUID, customers []models.Customer, eventType string) {
	// Get active template for this event type
	var template models.ReminderTemplate
	if err := s.db.Where("salon_id = ? AND type = ? AND is_active = true", salonID, eventType).
		First(&template).Error; err != nil {
		log.Printf("Salon %s: No active template for %s: %v", salonID, eventType, err)
		return
	}

	for _, customer := range customers {
		// Replace placeholders in the template
		message := strings.ReplaceAll(template.Message, "[CustomerName]", customer.Name)

		// Determine channel (WhatsApp if available, else SMS)
		channel := "sms"
		var to string

		// Use WhatsApp if phone is in E.164 format and starts with '+'
		if strings.HasPrefix(customer.Phone, "+") {
			to = "whatsapp:" + customer.Phone
			channel = "whatsapp"
		} else {
			to = customer.Phone
		}

		// Send message via Twilio
		params := &twilioApi.CreateMessageParams{}
		params.SetTo(to)
		params.SetBody(message)

		// Use WhatsApp sender if available
		if channel == "whatsapp" {
			params.SetFrom("whatsapp:" + os.Getenv("TWILIO_WHATSAPP_NUMBER"))
		} else {
			params.SetFrom(os.Getenv("TWILIO_PHONE_NUMBER"))
		}

		resp, err := s.client.Api.CreateMessage(params)
		status := "sent"
		errorMsg := ""

		if err != nil {
			log.Printf("Failed to send message to %s: %v", customer.Phone, err)
			status = "failed"
			errorMsg = err.Error()
		} else if resp.Sid != nil {
			log.Printf("Message sent to %s, SID: %s", customer.Phone, *resp.Sid)
		} else {
			log.Printf("Message sent to %s, but no SID returned", customer.Phone)
		}

		// Log the reminder
		reminderLog := models.ReminderLog{
			SalonID:      salonID,
			CustomerID:   customer.ID,
			TemplateID:   template.ID,
			Type:         eventType,
			Message:      message,
			Status:       status,
			ErrorMessage: errorMsg,
			Channel:      channel,
			SentAt:       time.Now(),
		}

		if err := s.db.Create(&reminderLog).Error; err != nil {
			log.Printf("Failed to log reminder for customer %s: %v", customer.ID, err)
		}
	}
}
