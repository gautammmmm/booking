package models

import "time"

// TimeSlot represents an available time slot for booking
type TimeSlot struct {
	ID          int       `json:"id"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	IsAvailable bool      `json:"is_available"`
	ServiceID   int       `json:"service_id"`
	ServiceName string    `json:"service_name,omitempty"` // for responses
	BusinessID  int       `json:"business_id"`
}

// GenerateSlotsRequest represents the data needed to generate time slots
type GenerateSlotsRequest struct {
	ServiceID int       `json:"service_id" binding:"required"`
	StartDate time.Time `json:"start_date" binding:"required"`
	EndDate   time.Time `json:"end_date" binding:"required"`
	StartTime string    `json:"start_time" binding:"required"` // e.g., "09:00"
	EndTime   string    `json:"end_time" binding:"required"`   // e.g., "17:00"
	Interval  int       `json:"interval" binding:"required"`   // minutes between slots (e.g., 30)
}

// PublicTimeSlot represents slot data for public API (customers)
type PublicTimeSlot struct {
	ID          int       `json:"id"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	ServiceID   int       `json:"service_id"`
	ServiceName string    `json:"service_name"`
	Duration    int       `json:"duration"`
}
