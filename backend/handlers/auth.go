package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"booking-backend/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Secret key for JWT (use environment variable in production!)
var jwtSecret = []byte("gtm") // Change this to a random string!

// Login handles user authentication
func Login(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Bind JSON input to LoginRequest struct
		var loginReq models.LoginRequest
		if err := c.ShouldBindJSON(&loginReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// 2. Find user by email
		var user models.User
		query := `SELECT id, email, password_hash, full_name, role, business_id FROM users WHERE email = $1`
		err := db.QueryRow(query, loginReq.Email).Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.Role, &user.BusinessID,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			}
			return
		}

		// 3. Check password
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginReq.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		var businessIDValue int
		if user.BusinessID != nil {
			businessIDValue = *user.BusinessID
		} else {
			businessIDValue = 0 // or -1, or some value that means "no business"
		}
		// 4. Create JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":     user.ID,
			"email":       user.Email,
			"role":        user.Role,
			"business_id": businessIDValue, // ← Use the concrete value
			"exp":         time.Now().Add(time.Hour * 24).Unix(),
		})

		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
			return
		}

		// 5. Return success response
		c.JSON(http.StatusOK, models.LoginResponse{
			Message: "Login successful",
			Token:   tokenString,
			User:    user,
		})
	}
}

func ProtectedProfile(c *gin.Context) {
	// Get the user from the context (set by the middleware)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	// Type assert back to User model
	currentUser := user.(models.User)

	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to protected route!",
		"user":    currentUser,
	})
}

// Register handles business registration
func Register(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Bind JSON input to RegistrationRequest struct
		var regReq models.RegistrationRequest
		if err := c.ShouldBindJSON(&regReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
			return
		}

		// Start a database transaction
		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not start transaction"})
			return
		}
		defer tx.Rollback() // This will be a no-op if tx.Commit() is successful

		// 2. Check if email already exists
		var existingUser models.User
		err = tx.QueryRow("SELECT id FROM users WHERE email = $1", regReq.Email).Scan(&existingUser.ID)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}

		// 3. Create the business
		var businessID int
		err = tx.QueryRow(
			"INSERT INTO businesses (name) VALUES ($1) RETURNING id",
			regReq.BusinessName,
		).Scan(&businessID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create business"})
			return
		}

		// 4. Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(regReq.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
			return
		}

		// 5. Create the admin user
		var userID int
		err = tx.QueryRow(
			`INSERT INTO users (email, password_hash, full_name, role, business_id) 
             VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			regReq.Email, string(hashedPassword), regReq.FullName, "business_admin", businessID,
		).Scan(&userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
			return
		}

		// 6. Generate JWT token (same as login)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":     userID,
			"email":       regReq.Email,
			"role":        "business_admin",
			"business_id": businessID, // ← ADD THIS LINE
			"exp":         time.Now().Add(time.Hour * 24).Unix(),
		})

		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
			return
		}

		// 7. Commit the transaction
		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
			return
		}

		// 8. Return success response
		c.JSON(http.StatusCreated, models.RegistrationResponse{
			Message: "Registration successful",
			Token:   tokenString,
			User: models.User{
				ID:         userID,
				Email:      regReq.Email,
				FullName:   regReq.FullName,
				Role:       "business_admin",
				BusinessID: &businessID,
			},
			Business: models.Business{
				ID:   businessID,
				Name: regReq.BusinessName,
			},
		})
	}
}
