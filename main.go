package main

import (
	"log"
	"os"
	"salonpro-backend/config"
	"salonpro-backend/routes"

	"github.com/joho/godotenv"
)

func init() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	config.ConnectDB()

	// config.DB.AutoMigrate(
	// 	&models.Salon{},
	// 	&models.User{},
	// 	&models.Customer{},
	// 	&models.Service{},
	// 	&models.Invoice{},
	// 	&models.InvoiceItem{},
	// 	&models.ReminderTemplate{},
	// 	// &models.ReminderLog{},
	// )
}

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r := routes.SetupRouter()
	// printRoutes(r)
	r.Run(":" + port)
}

// func printRoutes(r *gin.Engine) {
// 	routes := r.Routes()
// 	for _, route := range routes {
// 		fmt.Printf("%-6s %s\n", route.Method, route.Path)
// 	}
// }
