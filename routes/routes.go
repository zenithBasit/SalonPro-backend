package routes

import (
	"salonpro-backend/config"
	"salonpro-backend/controllers"
	"salonpro-backend/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"https://white-sky-0debbc31e.1.azurestaticapps.net",
			"https://salon.zenithive.digital",
			"http://localhost:3000",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://white-sky-0debbc31e.1.azurestaticapps.net" ||
				origin == "https://salon.zenithive.digital" ||
				origin == "http://localhost:3000"
		},
	}))

	r.Use(config.PerformanceLogger())

	auth := r.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)

		auth.Use(utils.AuthMiddleware())
		auth.GET("/me", controllers.Me)
	}

	api := r.Group("/api")
	api.Use(utils.AuthMiddleware())
	{
		// Customer routes
		customers := api.Group("/customers")
		{
			customers.POST("", controllers.CreateCustomer)
			customers.GET("", controllers.GetCustomers)
			customers.GET("/:id", controllers.GetCustomer)
			customers.PUT("/:id", controllers.UpdateCustomer)
			customers.DELETE("/:id", controllers.DeleteCustomer)
		}

		// Service routes
		services := api.Group("/services")
		{
			services.POST("", controllers.CreateService)
			services.GET("", controllers.GetServices)
			services.GET("/:id", controllers.GetService)
			services.PUT("/:id", controllers.UpdateService)
			services.DELETE("/:id", controllers.DeleteService)
		}

		// Invoice routes
		invoices := api.Group("/invoices")
		{
			invoices.POST("", controllers.CreateInvoice)
			invoices.GET("", controllers.GetInvoices)
			invoices.GET("/:id", controllers.GetInvoice)
			invoices.PUT("/:id", controllers.UpdateInvoice)
			invoices.DELETE("/:id", controllers.DeleteInvoice)
		}

		//Reports routes
		reportController := controllers.ReportController{}
		api.GET("/reports", reportController.GetReportAnalytics)

		// Dashboard routes
		api.GET("/dashboard", controllers.GetDashboardOverview)

		// Settings routes
		profile := auth.Group("/profile", utils.AuthMiddleware())
		{
			profile.GET("", controllers.GetProfile)
			profile.PUT("/update-salon", controllers.UpdateSalonProfile)
			profile.PUT("/update-hours", controllers.UpdateWorkingHours)
			profile.PUT("/update-templates", controllers.UpdateReminderTemplates)
			profile.PUT("/update-notifications", controllers.UpdateNotifications)
		}

		employees := api.Group("/employees")
		{
			employees.GET("", controllers.GetEmployees)          // GET /api/employees
			employees.POST("", controllers.AddEmployee)          // POST /api/employees
			employees.PUT("/:id", controllers.UpdateEmployee)    // PUT /api/employees/:id
			employees.DELETE("/:id", controllers.DeleteEmployee) // DELETE /api/employees/:id
		}

	}

	return r
}
