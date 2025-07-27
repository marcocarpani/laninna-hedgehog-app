# 🦔 La Ninna - Setup Instructions

## 📋 Prerequisites

### System Requirements
- **Go**: Version 1.19 or higher
- **SQLite3**: For database operations
- **Git**: For version control
- **Web Browser**: Modern browser (Chrome, Firefox, Safari, Edge)

### Development Tools (Optional)
- **VS Code** or **GoLand** for development
- **Postman** for API testing
- **SQLite Browser** for database inspection
- **Swagger CLI** for documentation generation

## 🚀 Installation Steps

### 1. Clone Repository
```bash
git clone <repository-url>
cd laninna-hedgehog-app
```

### 2. Install Go Dependencies
```bash
go mod tidy
```

### 3. Verify Installation
```bash
go version  # Should show Go 1.19+
```

### 4. Run Application
```bash
go run .
```

### 5. Access Application
- **Web Application**: http://localhost:8080
- **API Documentation**: http://localhost:8080/swagger/index.html
- Default credentials:
  - **Username**: `admin`
  - **Password**: `admin123`

## 🔧 Configuration

### Environment Variables
Create `.env` file (optional):
```bash
PORT=8080
JWT_SECRET=your-secret-key-here
DB_PATH=./laninna.db
```

### Database Setup
- SQLite database is **auto-created** on first run
- Location: `./laninna.db`
- No manual setup required

## 🏗️ Development Setup

### 1. Project Structure
```
laninna-hedgehog-app/
├── main.go              # Entry point
├── models.go            # Database models
├── handlers.go          # API handlers
├── auth.go              # Authentication
├── export.go            # Export features
├── notifications.go     # Notification system
├── templates/           # HTML templates
├── static/             # CSS, JS, images
└── laninna.db          # SQLite database (auto-created)
```

### 2. Hot Reload (Development)
Install air for hot reload:
```bash
go install github.com/cosmtrek/air@latest
air
```

### 3. Database Inspection
View database contents:
```bash
sqlite3 laninna.db
.tables
.schema hedgehogs
```

## 🧪 Testing Setup

### Run Tests
```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Specific test
go test -run TestHedgehogHandler
```

### API Testing

#### Swagger UI (Recommended)
- Open: http://localhost:8080/swagger/index.html
- Interactive API testing with authentication
- Complete endpoint documentation

#### Manual Testing with curl
```bash
# Login
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# Get hedgehogs (with token)
curl -X GET http://localhost:8080/api/hedgehogs \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## 🚀 Production Deployment

### 1. Build Binary
```bash
go build -o laninna-app
```

### 2. Run Production
```bash
export PORT=8080
export JWT_SECRET=your-production-secret
./laninna-app
```

### 3. Docker Deployment
```dockerfile
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o laninna-app

FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/
COPY --from=builder /app/laninna-app .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
EXPOSE 8080
CMD ["./laninna-app"]
```

Build and run:
```bash
docker build -t laninna-app .
docker run -p 8080:8080 -v $(pwd)/data:/root laninna-app
```

## 🔒 Security Setup

### 1. Change Default Credentials
After first login, create new admin user and disable default account.

### 2. JWT Secret
Set strong JWT secret in production:
```bash
export JWT_SECRET=$(openssl rand -base64 32)
```

### 3. Database Security
- Restrict file permissions: `chmod 600 laninna.db`
- Regular backups
- Use environment variables for sensitive data

## 🔧 Troubleshooting

### Common Issues

#### Port Already in Use
```bash
# Find process using port 8080
lsof -i :8080
# Kill process
kill -9 <PID>
```

#### Database Locked
```bash
# Stop application
# Remove lock file
rm laninna.db-wal laninna.db-shm
```

#### Permission Denied
```bash
# Make binary executable
chmod +x laninna-app
```

#### Go Module Issues
```bash
# Clean module cache
go clean -modcache
go mod download
```

### Logs and Debugging
- Application logs to stdout
- Enable debug mode: `export GIN_MODE=debug`
- Check browser console for frontend errors

## 📊 Initial Data Setup

### 1. Create First Room
1. Go to **Rooms** page
2. Click **"Nuova Stanza"**
3. Fill in room details
4. Save room

### 2. Add Areas to Room
1. Open **Room Builder**
2. Select your room
3. Switch to **"Aggiungi Area"** mode
4. Draw areas on canvas
5. Configure area properties

### 3. Add First Hedgehog
1. Go to **Hedgehogs** page
2. Click **"Nuovo Riccio"**
3. Fill in hedgehog details
4. Assign to area
5. Save hedgehog

### 4. Add Weight Records
1. Click weight icon on hedgehog card
2. Enter weight and date
3. Add notes if needed
4. Save record

### 5. Add Therapies
1. Click therapy icon on hedgehog card
2. Enter therapy details
3. Set start/end dates
4. Save therapy

## 🔔 Notification Setup

### Configure Alerts
1. Go to **Notifications** page
2. Access notification settings
3. Configure thresholds:
   - Weight drop threshold (default: 50g)
   - Therapy expiring days (default: 3)
   - No weighing days (default: 7)

### Email Notifications (Optional)
- Configure SMTP settings in code
- Enable email notifications in settings
- Test with sample notification

## 📈 Monitoring Setup

### Health Checks
- Endpoint: `GET /health`
- Database connectivity check
- Application status

### Performance Monitoring
- Request logging enabled by default
- Monitor database file size
- Check notification processing

## 🆘 Getting Help

### Documentation
- [README.md](README.md) - Project overview
- [PROJECT_GUIDELINES.md](PROJECT_GUIDELINES.md) - Development guidelines
- **Swagger UI**: http://localhost:8080/swagger/index.html - Interactive API docs

### Generate API Documentation
```bash
# Install Swagger CLI
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
swag init

# Documentation files created in docs/
```

### Support Channels
- Create GitHub issue for bugs
- Check existing documentation
- Review code comments

### Useful Commands
```bash
# Check Go version
go version

# List dependencies
go list -m all

# Format code
go fmt ./...

# Vet code
go vet ./...

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o laninna-linux
GOOS=windows GOARCH=amd64 go build -o laninna.exe
```

---

**Setup complete! Your hedgehog rescue center management system is ready to use** 🦔✨