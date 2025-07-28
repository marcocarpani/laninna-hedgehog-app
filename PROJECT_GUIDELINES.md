# La Ninna - Hedgehog Management System Guidelines

## 🦔 Project Overview
La Ninna is a hedgehog rescue center management system built with Go (Gin), SQLite, and vanilla JavaScript. The system manages hedgehogs, rooms, areas, weight records, therapies, and notifications.

## 📁 Project Structure
```
laninna-hedgehog-app/
├── main.go              # Application entry point & routing
├── models.go            # Database models (GORM)
├── handlers.go          # API handlers
├── auth.go              # Authentication middleware
├── export.go            # PDF/Excel/CSV export functionality
├── notifications.go     # Notification system
├── templates/           # HTML templates
├── static/             # Static assets (CSS, JS, images)
└── laninna.db          # SQLite database
```

## 🏗️ Architecture Principles

### Model Relationships
- **NO circular references** in models
- **Foreign keys only** (no embedded entities in WeightRecord/Therapy)
- **Service layer** handles data relationships via preloading
- **ID-based lookups** for cross-entity references

### Data Access Pattern
```go
// ✅ CORRECT - Use preloading in main entity
var hedgehogs []Hedgehog
db.Preload("WeightRecords").Preload("Therapies").Find(&hedgehogs)

// ❌ WRONG - Don't preload in child entities
var weights []WeightRecord
db.Preload("Hedgehog").Find(&weights) // AVOID THIS

// ✅ CORRECT - Lookup by ID when needed
var hedgehog Hedgehog
hedgehogName := "N/A"
if err := db.First(&hedgehog, record.HedgehogID).Error; err == nil {
    hedgehogName = hedgehog.Name
}
```

## 🔧 Development Standards

### Backend (Go)
- **Gin framework** for HTTP routing
- **GORM** for database operations
- **JWT authentication** with middleware
- **Swagger UI** for API documentation
- **Structured error handling** with proper HTTP status codes
- **Validation** for required fields (hedgehog_id, etc.)

### Frontend (JavaScript)
- **Vanilla JS** (no frameworks)
- **Fetch API** for HTTP requests
- **Authorization headers** for all API calls
- **Form validation** before API submission
- **Toast notifications** for user feedback

### Structured Logging Guide

This paragraph describes the structured logging system implemented in the application.

#### Overview

The application uses [zerolog](https://github.com/rs/zerolog) for structured logging, which provides:

- JSON-formatted logs for easy parsing and analysis
- Multiple log levels (debug, info, warn, error, fatal)
- Contextual information in logs (request ID, user ID, etc.)
- High performance with minimal allocations

#### Log Levels

The following log levels are available, in order of increasing severity:

1. **Debug**: Detailed information useful for debugging during development
2. **Info**: General operational information about system behavior
3. **Warn**: Potentially harmful situations that might indicate problems
4. **Error**: Error events that might still allow the application to continue running
5. **Fatal**: Severe error events that cause the application to terminate

The default log level is **Info**. You can change the log level by setting the `LOG_LEVEL` environment variable to one of: `debug`, `info`, `warn`, or `error`.

#### Using the Logger

##### Basic Logging

```go
// Import the logger package
import "github.com/laninna/hedgehog-app/logger"

// Log at different levels
logger.Debug("This is a debug message")
logger.Info("This is an info message")
logger.Warn("This is a warning message")
logger.Error("Failed to process request", err)
logger.Fatal("Application cannot continue", err)
```

##### Adding Context to Logs

You can add contextual information to logs using field helpers:

```go
logger.Info("Hedgehog created", 
    logger.Uint("id", hedgehog.ID),
    logger.Str("name", hedgehog.Name),
    logger.Str("status", hedgehog.Status))
```

Available field helpers:
- `logger.Str(key, value)`: Add a string field
- `logger.Int(key, value)`: Add an integer field
- `logger.Uint(key, value)`: Add an unsigned integer field
- `logger.Bool(key, value)`: Add a boolean field
- `logger.Err(err)`: Add an error field
- `logger.F(key, value)`: Add a field with any value

##### Request-Scoped Logging

In HTTP handlers, you should use the request-scoped logger from the context, which includes request ID and user information:

```go
func myHandler(c *gin.Context) {
    // Get the logger from context
    log := logger.GetLoggerFromContext(c)
    
    // Use the logger
    log.Info().
        Str("some_field", "some_value").
        Msg("Processing request")
}
```

#### Contextual Information

The logging system automatically includes the following contextual information:

1. **Timestamp**: When the log entry was created
2. **Level**: The log level (debug, info, warn, error, fatal)
3. **Message**: The log message
4. **Request ID**: A unique identifier for each HTTP request (added by middleware)
5. **User ID**: The ID of the authenticated user, if available (added by middleware)
6. **Username**: The username of the authenticated user, if available (added by middleware)
7. **Method**: The HTTP method (GET, POST, etc.) for request logs
8. **Path**: The request path for request logs
9. **Status**: The HTTP status code for response logs
10. **Latency**: The request processing time for response logs

#### Log Output Example

```json
{
  "level": "info",
  "time": "2025-07-28T21:50:00Z",
  "request_id": "1234567890",
  "user_id": 1,
  "username": "admin",
  "method": "POST",
  "path": "/api/hedgehogs",
  "status": 201,
  "latency": 0.015,
  "id": 42,
  "name": "Sonic",
  "status": "in_care",
  "message": "Hedgehog created successfully"
}
```

#### Best Practices

1. **Choose the appropriate log level**: Use debug for development information, info for normal operations, warn for potential issues, and error for actual errors.
2. **Include relevant context**: Add fields that will help troubleshoot issues, but avoid including sensitive information.
3. **Use structured fields**: Instead of embedding values in the message string, add them as structured fields.
4. **Be consistent with field names**: Use consistent naming conventions for fields across the application.
5. **Keep messages clear and concise**: The message should describe what happened, with details in the structured fields.

#### Configuration

The logger is configured in `main.go` with the following options:

```go
logConfig := logger.DefaultConfig()
logConfig.Pretty = true // Use pretty console output for development
logger.Init(logConfig)
```

In production, you might want to set `Pretty` to `false` to output JSON logs that can be easily parsed by log management systems.

### Database Design
```go
// Core entities
type Hedgehog struct {
    ID            uint
    Name          string
    Status        string // in_care, recovered, deceased
    AreaID        *uint
    WeightRecords []WeightRecord // Preloaded relationship
    Therapies     []Therapy      // Preloaded relationship
}

type WeightRecord struct {
    ID         uint
    HedgehogID uint    // Foreign key only
    Weight     float64
    Date       time.Time
    Notes      string
}

type Therapy struct {
    ID          uint
    HedgehogID  uint      // Foreign key only
    Name        string
    Description string    // Used for dosage
    StartDate   time.Time
    EndDate     *time.Time
    Status      string    // active, completed, suspended
}
```

## 🎨 UI/UX Guidelines

### Design System
- **Hedgehog theme** with brown/tan color palette
- **Tailwind CSS** for styling
- **Font Awesome** icons
- **Responsive design** (mobile-first)
- **Hover effects** with scale transforms

### Modal Patterns
```javascript
// Standard modal opening
function openModal(title, content) {
    document.getElementById('modal-title').textContent = title;
    document.getElementById('modal-content').innerHTML = content;
    document.getElementById('main-modal').classList.remove('hidden');
}

// Form submission with validation
async function handleSubmit(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    
    // Validate required fields
    const hedgehogId = parseInt(formData.get('hedgehog_id'));
    if (!hedgehogId || isNaN(hedgehogId)) {
        showToast('Seleziona un riccio', 'error');
        return;
    }
    
    // API call with proper headers
    const response = await fetch('/api/endpoint', {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    });
    
    if (response.ok) {
        document.getElementById('main-modal').classList.add('hidden');
        showToast('Operazione completata', 'success');
        loadData(); // Refresh data
    }
}
```

## 📊 Export System

### Implementation Pattern
```go
// Remove preloads, use ID lookups
func generateExport(db *gorm.DB) {
    var records []WeightRecord
    db.Find(&records) // No preload
    
    for _, record := range records {
        // Lookup hedgehog name by ID
        var hedgehog Hedgehog
        hedgehogName := "N/A"
        if err := db.First(&hedgehog, record.HedgehogID).Error; err == nil {
            hedgehogName = hedgehog.Name
        }
        // Use hedgehogName in export
    }
}
```

## 🔔 Notification System

### Notification Types
- `therapy_expired` - Therapy past end date
- `therapy_expiring` - Therapy ending soon
- `weight_drop` - Significant weight loss
- `weight_stagnation` - No weight change
- `no_weighing` - Missing weight records

### Implementation
- **Background scheduler** runs every 30 minutes
- **Duplicate prevention** with time-based checks
- **Priority levels**: low, medium, high, critical
- **Auto-expiry** after 30 days

## 🚀 Deployment

### Environment Setup
```bash
# Install dependencies
go mod tidy

# Run development server
go run .

# Build for production
go build -o laninna-app
```

### Configuration
- **Port**: 8080 (default)
- **Database**: SQLite (laninna.db)
- **Authentication**: JWT tokens in localStorage
- **File uploads**: Static directory serving

## 🧪 Testing Guidelines

### API Testing
- Test all CRUD operations
- Verify authentication middleware
- Check proper error responses
- Validate data relationships

### Frontend Testing
- Form validation
- Modal interactions
- Data refresh after operations
- Error handling and user feedback

## 📝 Code Style

### Go Conventions
- Use `gofmt` for formatting
- Follow standard Go naming conventions
- Handle errors explicitly
- Use structured logging

### JavaScript Conventions
- Use `const`/`let` (no `var`)
- Async/await for promises
- Event delegation for dynamic content
- Consistent error handling

## 📚 API Documentation

### Swagger Annotations
```go
// @Summary Get all hedgehogs
// @Description Get list of all hedgehogs with their areas
// @Tags Hedgehogs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Hedgehog
// @Router /hedgehogs [get]
func getHedgehogs(db *gorm.DB) gin.HandlerFunc {
```

### Model Documentation
```go
type Hedgehog struct {
    ID     uint   `json:"id" example:"1"`
    Name   string `json:"name" example:"Spillo"`
    Status string `json:"status" example:"in_care" enums:"in_care,recovered,deceased"`
}
```

### Generate Documentation
```bash
swag init  # Regenerate docs after changes
```

## 🔒 Security

### Authentication
- JWT tokens with expiration
- Protected routes with middleware
- Token validation on all API calls
- Logout clears localStorage

### Data Validation
- Server-side validation for all inputs
- Required field checks
- Type validation (integers, dates)
- SQL injection prevention via GORM

## 📈 Performance

### Database Optimization
- Use preloading for related data
- Limit query results (pagination)
- Index foreign keys
- Clean up old notifications

### Frontend Optimization
- Minimize DOM manipulations
- Use event delegation
- Lazy load data when needed
- Cache frequently used elements

## 🐛 Common Issues & Solutions

### Model Relationship Errors
**Problem**: Circular reference errors
**Solution**: Remove embedded entities, use ID lookups

### Authentication Failures
**Problem**: 401 Unauthorized errors
**Solution**: Check token in localStorage, verify Authorization header

### Form Submission Issues
**Problem**: Null hedgehog_id in API calls
**Solution**: Validate form data before submission, use hidden inputs

### Export Failures
**Problem**: Missing hedgehog names in exports
**Solution**: Use ID-based lookups instead of preloaded relationships

## 📚 Resources

- [Gin Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
- [Tailwind CSS](https://tailwindcss.com/docs)
- [Font Awesome Icons](https://fontawesome.com/icons)

---

**Remember**: Keep it simple, maintain consistency, and always validate user input! 🦔