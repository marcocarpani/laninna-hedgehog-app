package logger

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogLevels(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	
	// Initialize logger with test configuration
	config := DefaultConfig()
	config.Output = &buf
	config.Level = zerolog.DebugLevel
	config.Pretty = false // Use JSON format for testing
	Init(config)
	
	// Test different log levels
	Debug("This is a debug message", Str("test", "debug"))
	Info("This is an info message", Str("test", "info"))
	Warn("This is a warning message", Str("test", "warn"))
	Error("This is an error message", nil, Str("test", "error"))
	
	// Parse log output
	logs := parseLogOutput(buf.String())
	
	// Verify log levels
	if len(logs) != 4 {
		t.Errorf("Expected 4 log entries, got %d", len(logs))
	}
	
	// Check log levels
	checkLogLevel(t, logs[0], "debug")
	checkLogLevel(t, logs[1], "info")
	checkLogLevel(t, logs[2], "warn")
	checkLogLevel(t, logs[3], "error")
	
	// Check fields
	for i, level := range []string{"debug", "info", "warn", "error"} {
		if logs[i]["test"] != level {
			t.Errorf("Expected test field to be %s, got %s", level, logs[i]["test"])
		}
	}
}

func TestRequestLogger(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	
	// Initialize logger with test configuration
	config := DefaultConfig()
	config.Output = &buf
	config.Level = zerolog.InfoLevel
	config.Pretty = false // Use JSON format for testing
	Init(config)
	
	// Setup Gin router with logger middleware
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestLogger())
	
	// Add a test handler
	r.GET("/test", func(c *gin.Context) {
		log := GetLoggerFromContext(c)
		log.Info().Str("custom_field", "test_value").Msg("Test handler executed")
		c.String(http.StatusOK, "OK")
	})
	
	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "test-request-id")
	w := httptest.NewRecorder()
	
	// Serve the request
	r.ServeHTTP(w, req)
	
	// Parse log output
	logs := parseLogOutput(buf.String())
	
	// We should have 2 log entries: one from the middleware and one from the handler
	if len(logs) != 2 {
		t.Errorf("Expected 2 log entries, got %d", len(logs))
	}
	
	// Check request ID in logs
	for i, log := range logs {
		if log["request_id"] != "test-request-id" {
			t.Errorf("Log entry %d missing request_id", i)
		}
	}
	
	// Check custom field in handler log
	if logs[0]["custom_field"] != "test_value" {
		t.Errorf("Handler log missing custom_field")
	}
	
	// Check request details in middleware log
	if logs[1]["method"] != "GET" {
		t.Errorf("Middleware log missing or incorrect method")
	}
	
	if logs[1]["path"] != "/test" {
		t.Errorf("Middleware log missing or incorrect path")
	}
	
	if logs[1]["status"] != float64(200) { // JSON numbers are parsed as float64
		t.Errorf("Middleware log missing or incorrect status")
	}
}

func TestUserContextMiddleware(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	
	// Initialize logger with test configuration
	config := DefaultConfig()
	config.Output = &buf
	config.Level = zerolog.DebugLevel
	config.Pretty = false // Use JSON format for testing
	Init(config)
	
	// Setup Gin router with logger and user context middleware
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestLogger())
	r.Use(UserContextMiddleware())
	
	// Add a test handler
	r.GET("/protected", func(c *gin.Context) {
		log := GetLoggerFromContext(c)
		log.Info().Msg("Protected endpoint accessed")
		c.String(http.StatusOK, "OK")
	})
	
	// Create a test request with a mock JWT token
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("X-Request-ID", "test-request-id")
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImFkbWluIn0.8tat9AtRQT4ZF_LFgSDc-YFRxG2WkIvpKmP-AnXJOfA")
	w := httptest.NewRecorder()
	
	// Serve the request
	r.ServeHTTP(w, req)
	
	// Parse log output
	logs := parseLogOutput(buf.String())
	
	// Check if user information is included in logs
	foundUserInfo := false
	for _, log := range logs {
		if log["username"] == "admin" {
			foundUserInfo = true
			break
		}
	}
	
	if !foundUserInfo {
		t.Errorf("User information not found in logs")
	}
}

// Helper functions

func parseLogOutput(output string) []map[string]interface{} {
	var logs []map[string]interface{}
	
	// Split output by newlines
	lines := bytes.Split([]byte(output), []byte("\n"))
	
	// Parse each line as JSON
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		
		var log map[string]interface{}
		if err := json.Unmarshal(line, &log); err != nil {
			continue
		}
		
		logs = append(logs, log)
	}
	
	return logs
}

func checkLogLevel(t *testing.T, log map[string]interface{}, expectedLevel string) {
	if log["level"] != expectedLevel {
		t.Errorf("Expected log level %s, got %s", expectedLevel, log["level"])
	}
}