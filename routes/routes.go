package routes

import (
	"salonpro-backend/controllers"
	"salonpro-backend/utils"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

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
		serviceGroup := r.Group("/services")

		serviceGroup.POST("", controllers.CreateService)
		serviceGroup.GET("", controllers.GetServices)
		serviceGroup.GET("/:id", controllers.GetService)
		serviceGroup.PUT("/:id", controllers.UpdateService)
		serviceGroup.DELETE("/:id", controllers.DeleteService)
	}

	return r
}
