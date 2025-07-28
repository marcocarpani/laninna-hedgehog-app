# 🦔 La Ninna - Hedgehog Rescue Center Management System

A comprehensive web application for managing hedgehog rescue operations, built with Go and vanilla JavaScript.

## ✨ Features

### 🦔 Hedgehog Management
- Complete hedgehog profiles with status tracking
- Health records and medical history
- Area assignment and location tracking
- Advanced filtering and search capabilities

### ⚖️ Weight Monitoring
- Weight record tracking with trend analysis
- Automatic alerts for weight loss or stagnation
- Visual weight progression charts
- Export capabilities for veterinary reports

### 💊 Therapy Management
- Treatment scheduling and tracking
- Medication dosage recording
- Therapy expiration notifications
- Active/completed therapy status

### 🏠 Facility Management
- Room and area configuration
- Visual room builder with drag-and-drop
- Capacity management and occupancy tracking
- Interactive facility layout

### 🔔 Smart Notifications
- Automated health alerts
- Therapy expiration warnings
- Weight monitoring notifications
- Customizable notification settings

### 📊 Reporting & Export
- PDF, Excel, and CSV export options
- Comprehensive health reports
- Statistical analysis and trends
- Custom date range filtering

## 🚀 Quick Start

### Prerequisites
- Go 1.19 or higher
- SQLite3

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd laninna-hedgehog-app
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Run the application**
   ```bash
   go run .
   ```

4. **Access the application**
   - **Web App**: http://localhost:8080
   - **API Docs**: http://localhost:8080/swagger/index.html
   - Default login: `admin` / `admin123`

## 🏗️ Architecture

### Backend Stack
- **Go** with Gin web framework
- **GORM** for database operations
- **SQLite** for data storage
- **JWT** authentication
- **Swagger UI** for API documentation
- **PDF/Excel** export libraries

### Frontend Stack
- **Vanilla JavaScript** (no frameworks)
- **Tailwind CSS** for styling
- **Font Awesome** icons
- **Responsive design**

### Database Schema
```
Hedgehogs ──┐
           ├── WeightRecords
           ├── Therapies
           └── Areas ──── Rooms
```

## 📁 Project Structure

```
laninna-hedgehog-app/
├── main.go              # Application entry point
├── models.go            # Database models
├── handlers.go          # API handlers
├── auth.go              # Authentication
├── export.go            # Export functionality
├── notifications.go     # Notification system
├── docs/                # Swagger documentation
├── templates/           # HTML templates
├── static/             # Static assets
└── laninna.db          # SQLite database
```

## 🔧 Configuration

### Environment Variables
- `PORT` - Server port (default: 8080)
- `JWT_SECRET` - JWT signing secret
- `DB_PATH` - Database file path

### Default Settings
- **Database**: SQLite (auto-created)
- **Authentication**: JWT tokens
- **Session**: 24 hours
- **Notifications**: 30-minute intervals

## 📖 API Documentation

### Interactive Swagger UI
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **OpenAPI Spec**: http://localhost:8080/swagger/doc.json
- **Complete API documentation** with interactive testing
- **JWT Authentication** support built-in

### Authentication
```http
POST /api/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

### Hedgehogs
```http
GET    /api/hedgehogs           # List all hedgehogs
POST   /api/hedgehogs           # Create hedgehog
GET    /api/hedgehogs/:id       # Get hedgehog details
PUT    /api/hedgehogs/:id       # Update hedgehog
DELETE /api/hedgehogs/:id       # Delete hedgehog
```

### Weight Records
```http
GET    /api/weight-records      # List weight records
POST   /api/weight-records      # Add weight record
PUT    /api/weight-records/:id  # Update weight record
DELETE /api/weight-records/:id  # Delete weight record
```

### Therapies
```http
GET    /api/therapies           # List therapies
POST   /api/therapies           # Create therapy
PUT    /api/therapies/:id       # Update therapy
DELETE /api/therapies/:id       # Delete therapy
```

### Export
```http
GET /api/export/hedgehogs/pdf   # Export hedgehogs as PDF
GET /api/export/hedgehogs/excel # Export hedgehogs as Excel
GET /api/export/rooms/csv       # Export rooms as CSV
```

## 🎨 User Interface

### Dashboard
- Quick statistics overview
- Recent activity feed
- Quick action buttons
- Health alerts summary

### Hedgehog Management
- Grid view with status indicators
- Detailed modal views
- Inline editing capabilities
- Bulk operations

### Room Builder
- Visual room layout editor
- Drag-and-drop area creation
- Real-time capacity tracking
- Auto-save functionality

## 🔔 Notification System

### Alert Types
- **Critical**: Severe weight loss, expired therapies
- **High**: Therapy expiring soon
- **Medium**: Weight stagnation, missing records
- **Low**: General reminders

### Notification Settings
- Configurable thresholds
- Email notifications (optional)
- Webhook integration
- Auto-cleanup policies

## 📊 Export Features

### Supported Formats
- **PDF**: Professional reports with charts
- **Excel**: Detailed spreadsheets with multiple sheets
- **CSV**: Simple data export for analysis

### Export Options
- Date range filtering
- Status-based filtering
- Custom field selection
- Automated scheduling

## 🛡️ Security

### Authentication
- JWT-based session management
- Password hashing (bcrypt)
- Protected API endpoints
- Session timeout handling

### Data Protection
- Input validation and sanitization
- SQL injection prevention
- XSS protection
- CSRF token validation

## 🧪 Testing

### Running Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestHedgehogHandler
```

### Test Coverage
- API endpoint testing
- Database operations
- Authentication flows
- Export functionality

## 🚀 Deployment

### Render.com Deployment
The application includes a `render.yaml` blueprint for easy deployment to Render.com.

For detailed instructions, see the [Render Deployment Guide](render_deployment.md).

### Production Build
```bash
# Build binary
go build -o laninna-app

# Run production server
./laninna-app
```

### Docker Deployment
```dockerfile
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o laninna-app

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/laninna-app .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
CMD ["./laninna-app"]
```

### Environment Setup
```bash
# Production environment
export PORT=8080
export JWT_SECRET=your-secret-key
export DB_PATH=/data/laninna.db

# Start application
./laninna-app
```

## 📈 Performance

### Optimization Features
- Database indexing on foreign keys
- Query result pagination
- Lazy loading for large datasets
- Efficient preloading strategies

### Monitoring
- Request logging
- Error tracking
- Performance metrics
- Health check endpoints

## 🤝 Contributing

### Development Setup
1. Fork the repository
2. Create feature branch
3. Make changes following guidelines
4. Add tests for new features
5. Submit pull request

### Code Style
- Follow Go conventions
- Use `gofmt` for formatting
- Add comments for public functions
- Write meaningful commit messages

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

### Documentation
- [Project Guidelines](PROJECT_GUIDELINES.md)
- [API Documentation](docs/api.md)
- [Deployment Guide](docs/deployment.md)

### Getting Help
- Create an issue for bugs
- Use discussions for questions
- Check existing documentation first

## 🙏 Acknowledgments

- Built for hedgehog rescue centers
- Inspired by real-world rescue operations
- Community-driven development
- Open source contributions welcome

---

**Made with ❤️ for hedgehog rescue centers worldwide** 🦔