package main

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/laninna/hedgehog-app/logger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"time"
)

var jwtSecret = []byte("your-secret-key") // In produzione usa una variabile d'ambiente

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"admin123"`
}

type TokenResponse struct {
	Token        string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt    int64  `json:"expires_at" example:"1640995200"`
}

// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /login [post]
func loginHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request-scoped logger from context
		log := logger.GetLoggerFromContext(c)
		
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Warn().
				Err(err).
				Str("client_ip", c.ClientIP()).
				Msg("Invalid login request format")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Log login attempt (without password)
		log.Debug().
			Str("username", req.Username).
			Str("client_ip", c.ClientIP()).
			Msg("Login attempt")

		var user User
		if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
			log.Warn().
				Err(err).
				Str("username", req.Username).
				Str("client_ip", c.ClientIP()).
				Msg("Login failed: user not found")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			log.Warn().
				Err(err).
				Str("username", req.Username).
				Str("client_ip", c.ClientIP()).
				Msg("Login failed: invalid password")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		token, refreshToken, expiresAt, err := generateTokens(user.ID, user.Username)
		if err != nil {
			log.Error().
				Err(err).
				Str("username", req.Username).
				Uint("user_id", user.ID).
				Msg("Failed to generate token")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
			return
		}

		log.Info().
			Str("username", user.Username).
			Uint("user_id", user.ID).
			Str("client_ip", c.ClientIP()).
			Time("expires_at", time.Unix(expiresAt, 0)).
			Msg("User logged in successfully")

		c.JSON(http.StatusOK, TokenResponse{
			Token:        token,
			RefreshToken: refreshToken,
			ExpiresAt:    expiresAt,
		})
	}
}

func refreshTokenHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verifica il refresh token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}

		// Genera nuovi token
		newToken, newRefreshToken, expiresAt, err := generateTokens(claims.UserID, claims.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
			return
		}

		c.JSON(http.StatusOK, TokenResponse{
			Token:        newToken,
			RefreshToken: newRefreshToken,
			ExpiresAt:    expiresAt,
		})
	}
}

func generateTokens(userID uint, username string) (string, string, int64, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	refreshExpirationTime := time.Now().Add(7 * 24 * time.Hour)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", "", 0, err
	}

	refreshClaims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return "", "", 0, err
	}

	return tokenString, refreshTokenString, expirationTime.Unix(), nil
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request-scoped logger from context
		log := logger.GetLoggerFromContext(c)
		
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn().
				Str("client_ip", c.ClientIP()).
				Str("path", c.Request.URL.Path).
				Msg("Authentication failed: missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			log.Warn().
				Str("client_ip", c.ClientIP()).
				Str("path", c.Request.URL.Path).
				Str("auth_header", authHeader[:min(len(authHeader), 10)] + "..."). // Log only the beginning of the header
				Msg("Authentication failed: invalid authorization format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := authHeader[7:] // Remove "Bearer " prefix
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			log.Warn().
				Err(err).
				Str("client_ip", c.ClientIP()).
				Str("path", c.Request.URL.Path).
				Msg("Authentication failed: invalid or expired token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Update logger with user information
		newLogger := log.With().
			Uint("user_id", claims.UserID).
			Str("username", claims.Username).
			Logger()
		c.Set(logger.ContextKeyLogger, &newLogger)
		
		log.Debug().
			Uint("user_id", claims.UserID).
			Str("username", claims.Username).
			Str("path", c.Request.URL.Path).
			Msg("User authenticated successfully")

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// Helper function for string length comparison
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func createDefaultUser(db *gorm.DB) {
	var count int64
	db.Model(&User{}).Count(&count)

	if count == 0 {
		logger.Info("No users found, creating default admin user")
		
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			logger.Error("Failed to hash default admin password", err)
			return
		}
		
		user := User{
			Username: "admin",
			Password: string(hashedPassword),
		}
		
		if err := db.Create(&user).Error; err != nil {
			logger.Error("Failed to create default admin user", err)
			return
		}
		
		logger.Info("Default admin user created successfully", 
			logger.Uint("user_id", user.ID),
			logger.Str("username", user.Username))
	}
}

// Health check handler
func healthCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		health := map[string]interface{}{
			"status":            "healthy",
			"timestamp":         time.Now().Format(time.RFC3339),
			"external_services": checkExternalServicesHealth(),
			"notification_system": map[string]interface{}{
				"scheduler_running": true,
				"last_check":        "< 30m ago", // Placeholder
			},
		}

		c.JSON(http.StatusOK, health)
	}
}
