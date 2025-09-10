package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"booking-backend/models"

	"github.com/gin-gonic/gin"
)

// CreateService handles creating a new service for a business
func CreateService(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get the business ID from the authenticated user (set by middleware)
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		currentUser := user.(models.User)
		if currentUser.BusinessID == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "User is not associated with a business"})
			return
		}

		businessID := *currentUser.BusinessID

		// 2. Bind and validate the request data
		var serviceReq models.CreateServiceRequest
		if err := c.ShouldBindJSON(&serviceReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
			return
		}

		// 3. Create the service in the database
		var serviceID int
		err := db.QueryRow(
			`INSERT INTO services (name, description, duration, business_id) 
             VALUES ($1, $2, $3, $4) RETURNING id`,
			serviceReq.Name, serviceReq.Description, serviceReq.Duration, businessID,
		).Scan(&serviceID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create service: " + err.Error()})
			return
		}

		// 4. Return the created service
		c.JSON(http.StatusCreated, models.ServiceResponse{
			ID:          serviceID,
			Name:        serviceReq.Name,
			Description: serviceReq.Description,
			Duration:    serviceReq.Duration,
		})
	}
}

// GetServices handles fetching all services for a business
func GetServices(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get the business ID from the authenticated user
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		currentUser := user.(models.User)
		if currentUser.BusinessID == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "User is not associated with a business"})
			return
		}

		businessID := *currentUser.BusinessID

		// 2. Query services for this business
		rows, err := db.Query(
			"SELECT id, name, description, duration FROM services WHERE business_id = $1 ORDER BY name",
			businessID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch services"})
			return
		}
		defer rows.Close()

		// 3. Build the response
		var services []models.ServiceResponse
		for rows.Next() {
			var service models.ServiceResponse
			if err := rows.Scan(&service.ID, &service.Name, &service.Description, &service.Duration); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading services"})
				return
			}
			services = append(services, service)
		}

		// 4. Return the services
		c.JSON(http.StatusOK, services)
	}
}

// GetPublicServices gets services for public booking page
func GetPublicServices(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		businessID := c.Query("business_id")
		if businessID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "business_id is required"})
			return
		}

		bizID, err := strconv.Atoi(businessID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid business_id"})
			return
		}

		rows, err := db.Query(
			"SELECT id, name, description, duration FROM services WHERE business_id = $1 ORDER BY name",
			bizID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch services"})
			return
		}
		defer rows.Close()

		var services []models.ServiceResponse
		for rows.Next() {
			var service models.ServiceResponse
			if err := rows.Scan(&service.ID, &service.Name, &service.Description, &service.Duration); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading services"})
				return
			}
			services = append(services, service)
		}

		c.JSON(http.StatusOK, services)
	}
}

// DeleteService handles deleting a service
func DeleteService(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get the business ID from the authenticated user
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		currentUser := user.(models.User)
		if currentUser.BusinessID == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "User is not associated with a business"})
			return
		}

		businessID := *currentUser.BusinessID

		// 2. Get the service ID from URL parameter
		serviceID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
			return
		}

		// 3. Delete the service (only if it belongs to the user's business)
		result, err := db.Exec(
			"DELETE FROM services WHERE id = $1 AND business_id = $2",
			serviceID, businessID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete service"})
			return
		}

		// 4. Check if a service was actually deleted
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found or you don't have permission"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Service deleted successfully"})
	}
}
