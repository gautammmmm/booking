package middleware

import (
	"net/http"
	"strings"

	"booking-backend/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// Use the same secret key as in your handlers/auth.go
var jwtSecret = []byte("gtm")

// AuthMiddleware verifies the JWT token and attaches user info to the request
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get the token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// 2. Check if the header has the "Bearer " prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization format must be 'Bearer {token}'"})
			c.Abort()
			return
		}

		// 3. Extract the token (remove "Bearer " prefix)
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 4. Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.NewValidationError("Unexpected signing method", jwt.ValidationErrorSignatureInvalid)
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// 5. Extract claims from the token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// 6. Create a User object from the token claims and attach it to the context
		user := models.User{
			ID:    int(claims["user_id"].(float64)), // JSON numbers are float64
			Email: claims["email"].(string),
			Role:  claims["role"].(string),
		}
		if businessID, exists := claims["business_id"]; exists {
			var businessIDInt int
			switch v := businessID.(type) {
			case float64:
				businessIDInt = int(v)
			case int:
				businessIDInt = v
			default:
				// Handle other types if needed
				businessIDInt = 0
			}

			if businessIDInt > 0 {
				user.BusinessID = &businessIDInt
			}
		}

		// 7. Store the user information in the context for use in handlers
		c.Set("user", user)

		// 8. Continue to the next handler
		c.Next()
	}
}
