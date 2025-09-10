package main

import (
	"booking-backend/database"
	"booking-backend/handlers"
	"booking-backend/middleware"
	"fmt"
	"log"
	"net/http"
	"time" // Add this import

	"github.com/gin-contrib/cors" // Add this import
	"github.com/gin-gonic/gin"
)

func printRoutes(router *gin.Engine) {
	fmt.Println("\nRegistered Routes:")
	for _, route := range router.Routes() {
		fmt.Printf("%-6s %s\n", route.Method, route.Path)
	}
	fmt.Println()
}

func main() {
	// Connect to the database
	database.ConnectDB()
	defer database.DB.Close()

	router := gin.Default()

	// === ADD CORS MIDDLEWARE RIGHT HERE ===
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Public routes (no authentication needed)
	router.GET("/api/health", func(c *gin.Context) {
		err := database.DB.Ping()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database is down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Server and database are running!"})
	})

	router.POST("/api/login", handlers.Login(database.DB))
	router.POST("/api/register", handlers.Register(database.DB))

	// Protected routes (require authentication)
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware()) // Apply auth middleware to all routes in this group
	{
		protected.GET("/profile", handlers.ProtectedProfile) // protected endpoint
		protected.POST("/services", handlers.CreateService(database.DB))
		protected.GET("/services", handlers.GetServices(database.DB))
		protected.DELETE("/services/:id", handlers.DeleteService(database.DB))
		protected.POST("/slots/generate", handlers.GenerateSlots(database.DB))
		protected.GET("/slots", handlers.GetBusinessSlots(database.DB))
	}

	// Add public route for customers to see available slots:
	router.GET("/api/public/slots", handlers.GetPublicSlots(database.DB))
	// Add this with your other public routes
	router.GET("/api/public/services", handlers.GetPublicServices(database.DB))

	// Print all routes for debugging
	printRoutes(router)

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("CORS enabled for: http://localhost:5173, http://localhost:3000")
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /api/health (public)")
	fmt.Println("  POST /api/login (public)")
	fmt.Println("  POST /api/register (public)")
	fmt.Println("  GET  /api/public/slots (public)")
	fmt.Println("  GET  /api/profile (protected - requires auth token)")
	fmt.Println("  POST /api/services (protected)")
	fmt.Println("  GET  /api/services (protected)")
	fmt.Println("  DELETE /api/services/:id (protected)")
	fmt.Println("  POST /api/slots/generate (protected)")
	fmt.Println("  GET  /api/slots (protected)")

	err := router.Run(":8080")
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}
