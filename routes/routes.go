package routes

import (
	"salonpro-backend/config"
	"salonpro-backend/controllers"
	"salonpro-backend/services"
	"salonpro-backend/utils"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/test-reminder", func(c *gin.Context) {
		reminderService := services.NewReminderService(config.DB)
		reminderService.SendDailyReminders()
		c.JSON(200, gin.H{"message": "Reminders triggered"})
	})

	auth := r.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
	}

	api := r.Group("/api")
	api.Use(utils.AuthMiddleware())
	{
		// Customer routes
		api.POST("/customers", controllers.CreateCustomer)
		api.GET("/customers", controllers.GetCustomers)
		api.GET("/customers/:id", controllers.GetCustomer)
		api.PUT("/customers/:id", controllers.UpdateCustomer)
		api.DELETE("/customers/:id", controllers.DeleteCustomer)

		// Services routes
		api.POST("/services", controllers.CreateService)
		api.GET("/services", controllers.GetServices)
		api.GET("/services/:id", controllers.GetService)
		api.PUT("/services/:id", controllers.UpdateService)
		api.DELETE("/services/:id", controllers.DeleteService)

		// Invoices routes
		api.POST("/invoices", controllers.CreateInvoice)
		api.GET("/invoices", controllers.GetInvoices)
		api.GET("/invoices/:id", controllers.GetInvoice)
		api.PUT("/invoices/:id", controllers.UpdateInvoice)
		api.DELETE("/invoices/:id", controllers.DeleteInvoice)

		// Reminders routes
		api.POST("/reminder-templates", controllers.CreateReminderTemplate)
		api.GET("/reminder-templates", controllers.GetReminderTemplates)
		api.GET("/reminder-templates/:id", controllers.GetReminderTemplate)
		api.PUT("/reminder-templates/:id", controllers.UpdateReminderTemplate)
		api.DELETE("/reminder-templates/:id", controllers.DeleteReminderTemplate)

		//Reports routes
		reportController := controllers.ReportController{}
        api.GET("/reports", reportController.GetReportAnalytics)
	}

	return r
}
