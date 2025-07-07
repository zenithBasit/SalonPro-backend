package routes

import (
	"salonpro-backend/controllers"
	"salonpro-backend/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://preview--salonpro-master-plan.lovable.app"}, // change to your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// r.GET("/test-reminder", func(c *gin.Context) {
	// 	reminderService := services.NewReminderService(config.DB)
	// 	reminderService.SendDailyReminders()
	// 	c.JSON(200, gin.H{"message": "Reminders triggered"})
	// })

	auth := r.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)

		auth.Use(utils.AuthMiddleware())
		auth.GET("/me", controllers.Me)

		// User profile routes
		auth.GET("/employees", controllers.GetEmployees)
		auth.POST("/employees", controllers.AddEmployee)
		auth.PUT("/employees", controllers.UpdateEmployee)
		auth.DELETE("/employees", controllers.DeleteEmployee)
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

		//Reports routes
		reportController := controllers.ReportController{}
		api.GET("/reports", reportController.GetReportAnalytics)

		// Dashboard routes
		api.GET("/dashboard", controllers.GetDashboardOverview)

		// Settings routes
		auth.GET("/profile", utils.AuthMiddleware(), controllers.GetProfile)
		auth.PUT("/profile/update-salon", utils.AuthMiddleware(), controllers.UpdateSalonProfile)
		auth.PUT("/profile/update-hours", utils.AuthMiddleware(), controllers.UpdateWorkingHours)
		auth.PUT("/profile/update-templates", utils.AuthMiddleware(), controllers.UpdateReminderTemplates)
		auth.PUT("/profile/update-notifications", utils.AuthMiddleware(), controllers.UpdateNotifications)

		// auth.PUT("/update-profile", utils.AuthMiddleware(), controllers.UpdateProfile)
		// auth.PUT("/working-hours", utils.AuthMiddleware(), controllers.UpdateWorkingHours)
		// auth.PUT("/notification-settings", utils.AuthMiddleware(), controllers.UpdateNotificationSettings)

		// api.GET("/reminder-settings", controllers.GetReminderSettings)
		// api.PUT("/reminder-settings", controllers.UpdateReminderSetting)
	}

	return r
}
