package logger

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
	"math/rand"
	"fmt"
	"strings"
)

const (
	// ContextKeyRequestID is the key used to store the request ID in the context
	ContextKeyRequestID = "request_id"
	// ContextKeyUserID is the key used to store the user ID in the context
	ContextKeyUserID = "user_id"
	// ContextKeyUsername is the key used to store the username in the context
	ContextKeyUsername = "username"
	// ContextKeyLogger is the key used to store the logger in the context
	ContextKeyLogger = "logger"
)

// RequestLogger is a middleware that logs incoming HTTP requests
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Generate request ID if not already set
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set(ContextKeyRequestID, requestID)
		c.Header("X-Request-ID", requestID)

		// Create a request-scoped logger with request ID
		reqLogger := log.With().
			Str("request_id", requestID).
			Str("method", c.Request.Method).
			Str("path", path).
			Logger()

		// Store the logger in the context
		c.Set(ContextKeyLogger, &reqLogger)

		// Process request
		c.Next()

		// Log request details after processing
		latency := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()

		// Determine log level based on status code
		var event *zerolog.Event
		switch {
		case status >= 500:
			event = reqLogger.Error()
		case status >= 400:
			event = reqLogger.Warn()
		default:
			event = reqLogger.Info()
		}

		// Add query if present
		if raw != "" {
			event = event.Str("query", raw)
		}

		// Log request completion
		event.Int("status", status).
			Int("size", size).
			Dur("latency", latency).
			Msg("Request completed")
	}
}

// UserContextMiddleware extracts user information from JWT token and adds it to the context
func UserContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the logger from context
		logger := GetLoggerFromContext(c)

		// Extract JWT token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		
		// Parse the JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// This is a simplified example - in a real app, you'd use the actual JWT secret
			// For now, we're just trying to extract claims, not fully validate the token
			return []byte("your-jwt-secret"), nil
		})

		if err != nil || !token.Valid {
			// Just log and continue - don't block the request
			logger.Debug().Err(err).Msg("Failed to parse JWT token")
			c.Next()
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Extract user ID
			if userID, exists := claims["user_id"]; exists {
				c.Set(ContextKeyUserID, userID)
				
				// Update logger with user ID
				newLogger := logger.With().
					Interface("user_id", userID).
					Logger()
				
				c.Set(ContextKeyLogger, &newLogger)
			}

			// Extract username
			if username, exists := claims["username"]; exists {
				c.Set(ContextKeyUsername, username)
				
				// Update logger with username
				newLogger := GetLoggerFromContext(c).With().
					Interface("username", username).
					Logger()
				
				c.Set(ContextKeyLogger, &newLogger)
			}
		}

		c.Next()
	}
}

// generateRequestID generates a random request ID
func generateRequestID() string {
	// Simple implementation - in production, consider using UUID
	return fmt.Sprintf("%d", rand.Int63())
}