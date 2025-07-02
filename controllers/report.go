// controllers/report.go
package controllers

import (
	"net/http"
	"salonpro-backend/config"
	"salonpro-backend/models"
	"salonpro-backend/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ReportController handles all reporting functions
type ReportController struct{}

// AnalyticsSummary represents the Analytics data
type AnalyticsSummary struct {
	CurrentMonthRevenue   float64           `json:"currentMonthRevenue"`
	MonthGrowth           float64           `json:"monthGrowth"`
	CurrentQuarterRevenue float64           `json:"currentQuarterRevenue"`
	QuarterGrowth         float64           `json:"quarterGrowth"`
	CurrentYearRevenue    float64           `json:"currentYearRevenue"`
	YearGrowth            float64           `json:"yearGrowth"`
	TopServices           []ServiceSummary  `json:"topServices"`
	TopCustomers          []CustomerSummary `json:"topCustomers"`
	QuickStats            QuickStatistics   `json:"quickStats"`
}

type ServiceSummary struct {
	Name    string  `json:"name"`
	Count   int     `json:"count"`
	Revenue float64 `json:"revenue"`
}

type CustomerSummary struct {
	Name   string  `json:"name"`
	Visits int     `json:"visits"`
	Spent  float64 `json:"spent"`
}

type QuickStatistics struct {
	TotalCustomers   int     `json:"totalCustomers"`
	TotalInvoices    int     `json:"totalInvoices"`
	AvgMonthlyVisits float64 `json:"avgMonthlyVisits"`
	AvgOrderValue    float64 `json:"avgOrderValue"`
}

// GetDashboard returns the complete dashboard summary
func (rc *ReportController) GetReportAnalytics(c *gin.Context) {
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

	// Get current time
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	// Calculate date ranges
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	// Get revenue reports
	currentMonthRevenue, err := rc.getRevenue(salonUUID, firstOfMonth, lastOfMonth)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get monthly revenue")
		return
	}

	lastMonthRevenue, err := rc.getRevenue(salonUUID,
		firstOfMonth.AddDate(0, -1, 0),
		lastOfMonth.AddDate(0, -1, 0))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get last month revenue")
		return
	}

	currentQuarterRevenue, err := rc.getRevenue(salonUUID,
		rc.getQuarterStart(now),
		rc.getQuarterEnd(now))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get quarterly revenue")
		return
	}

	lastQuarterRevenue, err := rc.getRevenue(salonUUID,
		rc.getQuarterStart(now).AddDate(0, -3, 0),
		rc.getQuarterEnd(now).AddDate(0, -3, 0))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get last quarter revenue")
		return
	}

	currentYearRevenue, err := rc.getRevenue(salonUUID,
		time.Date(currentYear, 1, 1, 0, 0, 0, 0, currentLocation),
		time.Date(currentYear, 12, 31, 23, 59, 59, 0, currentLocation))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get yearly revenue")
		return
	}

	lastYearRevenue, err := rc.getRevenue(salonUUID,
		time.Date(currentYear-1, 1, 1, 0, 0, 0, 0, currentLocation),
		time.Date(currentYear-1, 12, 31, 23, 59, 59, 0, currentLocation))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get last year revenue")
		return
	}

	// Calculate growth percentages
	monthGrowth := rc.calculateGrowthPercentage(currentMonthRevenue, lastMonthRevenue)
	quarterGrowth := rc.calculateGrowthPercentage(currentQuarterRevenue, lastQuarterRevenue)
	yearGrowth := rc.calculateGrowthPercentage(currentYearRevenue, lastYearRevenue)

	// Get top services
	topServices, err := rc.getTopServices(salonUUID, firstOfMonth, lastOfMonth, 4)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get top services")
		return
	}

	// Get top customers
	topCustomers, err := rc.getTopCustomers(salonUUID, firstOfMonth, lastOfMonth, 4)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get top customers")
		return
	}

	// Get quick statistics
	quickStats, err := rc.getQuickStatistics(salonUUID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get quick statistics")
		return
	}

	// Create response
	summary := AnalyticsSummary{
    CurrentMonthRevenue:   currentMonthRevenue,
    MonthGrowth:           monthGrowth,
    CurrentQuarterRevenue: currentQuarterRevenue,
    QuarterGrowth:         quarterGrowth,
    CurrentYearRevenue:    currentYearRevenue,
    YearGrowth:            yearGrowth,
    TopServices:           topServices,
    TopCustomers:          topCustomers,
    QuickStats:            quickStats,
}

	c.JSON(http.StatusOK, summary)
}

// Helper functions for reports

func (rc *ReportController) getRevenue(salonID uuid.UUID, start, end time.Time) (float64, error) {
	var total float64
	err := config.DB.Model(&models.Invoice{}).
		Where("salon_id = ? AND invoice_date BETWEEN ? AND ?", salonID, start, end).
		Select("COALESCE(SUM(total), 0)").
		Scan(&total).Error
	return total, err
}

func (rc *ReportController) getQuarterStart(date time.Time) time.Time {
	quarter := (int(date.Month())-1)/3 + 1
	startMonth := time.Month((quarter-1)*3 + 1)
	return time.Date(date.Year(), startMonth, 1, 0, 0, 0, 0, date.Location())
}

func (rc *ReportController) getQuarterEnd(date time.Time) time.Time {
	return rc.getQuarterStart(date).AddDate(0, 3, -1)
}

func (rc *ReportController) calculateGrowthPercentage(current, previous float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return ((current - previous) / previous) * 100
}

func (rc *ReportController) getTopServices(salonID uuid.UUID, start, end time.Time, limit int) ([]ServiceSummary, error) {
	var services []ServiceSummary

	err := config.DB.Table("invoice_items").
		Select("services.name, SUM(invoice_items.quantity) as count, SUM(invoice_items.total_price) as revenue").
		Joins("JOIN invoices ON invoices.id = invoice_items.invoice_id").
		Joins("JOIN services ON services.id = invoice_items.service_id").
		Where("invoices.salon_id = ? AND invoices.invoice_date BETWEEN ? AND ? AND invoices.deleted_at IS NULL AND services.deleted_at IS NULL", salonID, start, end).
		Group("services.name").
		Order("revenue DESC").
		Limit(limit).
		Scan(&services).Error

	return services, err
}

func (rc *ReportController) getTopCustomers(salonID uuid.UUID, start, end time.Time, limit int) ([]CustomerSummary, error) {
	var customers []CustomerSummary

	err := config.DB.Table("invoices").
		Select("customers.name, COUNT(invoices.id) as visits, SUM(invoices.total) as spent").
		Joins("JOIN customers ON customers.id = invoices.customer_id").
		Where("invoices.salon_id = ? AND invoices.invoice_date BETWEEN ? AND ? AND invoices.deleted_at IS NULL AND customers.deleted_at IS NULL", salonID, start, end).
		Group("customers.name").
		Order("spent DESC").
		Limit(limit).
		Scan(&customers).Error

	return customers, err
}

func (rc *ReportController) getQuickStatistics(salonID uuid.UUID) (QuickStatistics, error) {
	var stats QuickStatistics

	// Total Customers
	var totalCustomers int64
	if err := config.DB.Model(&models.Customer{}).
		Where("salon_id = ? AND deleted_at IS NULL", salonID).
		Count(&totalCustomers).Error; err != nil {
		return stats, err
	}
	stats.TotalCustomers = int(totalCustomers)

	// Total Invoices
	var totalInvoices int64
	if err := config.DB.Model(&models.Invoice{}).
		Where("salon_id = ? AND deleted_at IS NULL", salonID).
		Count(&totalInvoices).Error; err != nil {
		return stats, err
	}
	stats.TotalInvoices = int(totalInvoices)

	// Average Monthly Visits
	var avgVisits float64
	err := config.DB.Raw(`
		SELECT AVG(visits) FROM (
			SELECT COUNT(*) as visits
			FROM invoices
			WHERE salon_id = ? AND deleted_at IS NULL
			GROUP BY DATE_TRUNC('month', invoice_date)
		) monthly_visits
	`, salonID).Scan(&avgVisits).Error
	if err != nil {
		return stats, err
	}
	stats.AvgMonthlyVisits = avgVisits

	// Average Order Value
	var totalRevenue float64
	if err := config.DB.Model(&models.Invoice{}).
		Where("salon_id = ? AND deleted_at IS NULL", salonID).
		Select("COALESCE(SUM(total), 0)").
		Scan(&totalRevenue).Error; err != nil {
		return stats, err
	}

	if stats.TotalInvoices > 0 {
		stats.AvgOrderValue = totalRevenue / float64(stats.TotalInvoices)
	}

	return stats, nil
}
