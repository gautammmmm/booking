package models

// User represents a user in the system
type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"` // The - means this field won't be shown in JSON
	FullName     string `json:"full_name"`
	Role         string `json:"role"`
	BusinessID   *int   `json:"business_id"` // Use pointer to allow NULL values
}

// LoginRequest represents the data sent for login
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the data returned after successful login
type LoginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
	User    User   `json:"user"`
}
type RegistrationRequest struct {
	BusinessName string `json:"business_name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	FullName     string `json:"full_name" binding:"required"`
	Password     string `json:"password" binding:"required,min=6"`
}

// RegistrationResponse represents the data returned after successful registration
type RegistrationResponse struct {
	Message  string   `json:"message"`
	Token    string   `json:"token"`
	User     User     `json:"user"`
	Business Business `json:"business"`
}

// Business represents a business entity
type Business struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Service represents a service offered by a business
type Service struct {
	ID          int    `json:"id"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`             // optional
	Duration    int    `json:"duration" binding:"required,min=1"` // in minutes
	BusinessID  int    `json:"business_id"`                       // will be set from context, not from request
}

// CreateServiceRequest represents the data needed to create a service
type CreateServiceRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
	Duration    int    `json:"duration" binding:"required,min=1"`
}

// ServiceResponse represents the service data returned in responses
type ServiceResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Duration    int    `json:"duration"`
}
