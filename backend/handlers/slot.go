package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"booking-backend/models"

	"github.com/gin-gonic/gin"
)

// GenerateSlots handles creating time slots for a service
func GenerateSlots(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get business ID from authenticated user
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

		// 2. Bind and validate request
		var genReq models.GenerateSlotsRequest
		if err := c.ShouldBindJSON(&genReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
			return
		}

		// 3. Verify the service belongs to this business
		var service models.Service
		err := db.QueryRow(
			"SELECT id, name, duration FROM services WHERE id = $1 AND business_id = $2",
			genReq.ServiceID, businessID,
		).Scan(&service.ID, &service.Name, &service.Duration)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Service not found or access denied"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			}
			return
		}

		// 4. Generate time slots
		generatedSlots, err := generateTimeSlots(genReq, service.Duration, businessID, db)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Could not generate slots: " + err.Error()})
			return
		}

		// 5. Save slots to database
		createdSlots, err := saveSlotsToDB(db, generatedSlots)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save slots: " + err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": fmt.Sprintf("Generated %d time slots", len(createdSlots)),
			"slots":   createdSlots,
		})
	}
}

// generateTimeSlots creates time slot objects based on the request
func generateTimeSlots(req models.GenerateSlotsRequest, serviceDuration int, businessID int, db *sql.DB) ([]models.TimeSlot, error) {
	var slots []models.TimeSlot

	// 1. Get business timezone from database
	var timezone string
	err := db.QueryRow("SELECT timezone FROM businesses WHERE id = $1", businessID).Scan(&timezone)
	if err != nil {
		// Fallback to UTC if timezone not found
		timezone = "UTC"
	}

	// 2. Load the timezone location
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		// Fallback to UTC if timezone is invalid
		loc = time.UTC
	}

	fmt.Printf("DEBUG: Using timezone %s for business %d\n", timezone, businessID)

	// 3. Convert input dates to the business's timezone
	startDate := req.StartDate.In(loc).Truncate(24 * time.Hour)
	endDate := req.EndDate.In(loc).Truncate(24 * time.Hour).Add(24 * time.Hour)

	// Parse time strings (handle both "09:00" and "09:00:00")
	startTimeStr := req.StartTime
	if len(strings.Split(startTimeStr, ":")) == 2 {
		startTimeStr += ":00"
	}

	endTimeStr := req.EndTime
	if len(strings.Split(endTimeStr, ":")) == 2 {
		endTimeStr += ":00"
	}

	currentDate := startDate
	for currentDate.Before(endDate) {
		// Skip Sundays
		if currentDate.Weekday() == time.Sunday {
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		// Create full datetime objects for the current day in business timezone
		startDateTimeStr := currentDate.Format("2006-01-02") + " " + startTimeStr
		endDateTimeStr := currentDate.Format("2006-01-02") + " " + endTimeStr

		startDateTime, err := time.ParseInLocation("2006-01-02 15:04:05", startDateTimeStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid start time format: %v", err)
		}

		endDateTime, err := time.ParseInLocation("2006-01-02 15:04:05", endDateTimeStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid end time format: %v", err)
		}

		currentSlotTime := startDateTime
		for currentSlotTime.Before(endDateTime) {
			slotEnd := currentSlotTime.Add(time.Minute * time.Duration(serviceDuration))

			// Don't create slots that would extend beyond working hours
			if slotEnd.After(endDateTime) {
				break
			}

			// Convert to UTC for storage
			slots = append(slots, models.TimeSlot{
				StartTime:   currentSlotTime.UTC(), // Store in UTC
				EndTime:     slotEnd.UTC(),         // Store in UTC
				IsAvailable: true,
				ServiceID:   req.ServiceID,
				BusinessID:  businessID,
			})

			// Move to next potential slot time
			currentSlotTime = currentSlotTime.Add(time.Minute * time.Duration(req.Interval+serviceDuration))
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	fmt.Printf("DEBUG: Generated %d slots in timezone %s\n", len(slots), timezone)
	return slots, nil
}

// saveSlotsToDB inserts generated slots into the database
func saveSlotsToDB(db *sql.DB, slots []models.TimeSlot) ([]models.TimeSlot, error) {
	var createdSlots []models.TimeSlot

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for _, slot := range slots {
		var slotID int
		err := tx.QueryRow(
			`INSERT INTO appointment_slots (start_time, end_time, is_available, service_id, business_id) 
             VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			slot.StartTime, slot.EndTime, slot.IsAvailable, slot.ServiceID, slot.BusinessID,
		).Scan(&slotID)

		if err != nil {
			return nil, err
		}

		slot.ID = slotID
		createdSlots = append(createdSlots, slot)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return createdSlots, nil
}

// GetBusinessSlots gets all slots for a business (admin view)
func GetBusinessSlots(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		rows, err := db.Query(`
            SELECT s.id, s.start_time, s.end_time, s.is_available, s.service_id, sv.name as service_name
            FROM appointment_slots s
            JOIN services sv ON s.service_id = sv.id
            WHERE s.business_id = $1
            ORDER BY s.start_time
        `, businessID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch slots"})
			return
		}
		defer rows.Close()

		var slots []map[string]interface{}
		for rows.Next() {
			var slot models.TimeSlot
			var serviceName string
			if err := rows.Scan(&slot.ID, &slot.StartTime, &slot.EndTime, &slot.IsAvailable, &slot.ServiceID, &serviceName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading slots"})
				return
			}

			slots = append(slots, map[string]interface{}{
				"id":           slot.ID,
				"start_time":   slot.StartTime,
				"end_time":     slot.EndTime,
				"is_available": slot.IsAvailable,
				"service_id":   slot.ServiceID,
				"service_name": serviceName,
			})
		}

		c.JSON(http.StatusOK, slots)
	}
}

// GetPublicSlots gets available slots for public booking
func GetPublicSlots(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		businessIDStr := c.Query("business_id")
		serviceIDStr := c.Query("service_id")
		dateStr := c.Query("date")

		if businessIDStr == "" || serviceIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "business_id and service_id are required"})
			return
		}

		businessID, err := strconv.Atoi(businessIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid business_id"})
			return
		}

		serviceID, err := strconv.Atoi(serviceIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service_id"})
			return
		}

		var query string
		var args []interface{}

		if dateStr != "" {
			query = `
                SELECT s.id, s.start_time, s.end_time, sv.name as service_name, sv.duration
                FROM appointment_slots s
                JOIN services sv ON s.service_id = sv.id
                WHERE s.business_id = $1 AND s.service_id = $2 
                AND s.is_available = true
                AND DATE(s.start_time) = $3
                ORDER BY s.start_time
            `
			args = []interface{}{businessID, serviceID, dateStr}
		} else {
			query = `
				SELECT s.id, s.start_time, s.end_time, sv.name as service_name, sv.duration
				FROM appointment_slots s
				JOIN services sv ON s.service_id = sv.id
				WHERE s.business_id = $1 AND s.service_id = $2 
				AND s.is_available = true
				AND s.start_time AT TIME ZONE 'UTC' > NOW() AT TIME ZONE 'UTC'
				ORDER BY s.start_time
			`
			args = []interface{}{businessID, serviceID}
		}
		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch available slots"})
			return
		}
		defer rows.Close()

		var slots []models.PublicTimeSlot
		for rows.Next() {
			var slot models.PublicTimeSlot
			if err := rows.Scan(&slot.ID, &slot.StartTime, &slot.EndTime, &slot.ServiceName, &slot.Duration); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading slots"})
				return
			}
			slot.ServiceID = serviceID
			slots = append(slots, slot)
		}

		c.JSON(http.StatusOK, slots)
	}
}
