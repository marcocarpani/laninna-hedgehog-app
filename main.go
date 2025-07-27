// @title La Ninna - Hedgehog Management API
// @version 1.0
// @description API for managing hedgehog rescue center operations
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"

	_ "github.com/laninna/hedgehog-app/docs"
)

func main() {
	// Carica variabili d'ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Inizializza database
	db, err := initDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Inizializza router
	r := setupRouter(db)

	// Avvia sistema notifiche
	StartNotificationScheduler(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🦔 La Ninna server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func initDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("laninna.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate models (incluse notifiche)
	err = db.AutoMigrate(
		&User{},
		&Hedgehog{},
		&Room{},
		&Area{},
		&Therapy{},
		&WeightRecord{},
		&Notification{},         // ← Nuovo
		&NotificationSettings{}, // ← Nuovo
	)
	if err != nil {
		return nil, err
	}

	// Crea utente admin di default
	createDefaultUser(db)

	return db, nil
}

func setupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
		AllowCredentials: true,
	}))

	// Serve static files
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// Routes
	api := r.Group("/api")
	{
		// Auth routes
		api.POST("/login", loginHandler(db))
		api.POST("/refresh", refreshTokenHandler(db))

		// Protected routes
		protected := api.Group("/")
		protected.Use(authMiddleware())
		{
			// Hedgehogs CRUD
			protected.GET("/hedgehogs", getHedgehogs(db))
			protected.POST("/hedgehogs", createHedgehog(db))
			protected.GET("/hedgehogs/:id", getHedgehog(db))
			protected.PUT("/hedgehogs/:id", updateHedgehog(db))
			protected.DELETE("/hedgehogs/:id", deleteHedgehog(db))

			// Rooms CRUD
			protected.GET("/rooms", getRooms(db))
			protected.POST("/rooms", createRoom(db))
			protected.GET("/rooms/:id", getRoom(db))
			protected.PUT("/rooms/:id", updateRoom(db))
			protected.DELETE("/rooms/:id", deleteRoom(db))

			// Areas CRUD
			protected.GET("/areas", getAreas(db))
			protected.POST("/areas", createArea(db))
			protected.PUT("/areas/:id", updateArea(db))
			protected.DELETE("/areas/:id", deleteArea(db))

			// Therapies CRUD
			protected.GET("/therapies", getTherapies(db))
			protected.POST("/therapies", createTherapy(db))
			protected.PUT("/therapies/:id", updateTherapy(db))
			protected.DELETE("/therapies/:id", deleteTherapy(db))

			// Weight Records CRUD
			protected.GET("/weight-records", getWeightRecords(db))
			protected.POST("/weight-records", createWeightRecord(db))
			protected.PUT("/weight-records/:id", updateWeightRecord(db))
			protected.DELETE("/weight-records/:id", deleteWeightRecord(db))

			// Export routes
			protected.POST("/export", exportDataHandler(db))
			protected.GET("/export/hedgehogs/pdf", quickExportHandler(db, "hedgehogs", "pdf"))
			protected.GET("/export/hedgehogs/excel", quickExportHandler(db, "hedgehogs", "excel"))
			protected.GET("/export/hedgehogs/csv", quickExportHandler(db, "hedgehogs", "csv"))
			protected.GET("/export/rooms/pdf", quickExportHandler(db, "rooms", "pdf"))
			protected.GET("/export/rooms/excel", quickExportHandler(db, "rooms", "excel"))
			protected.GET("/export/rooms/csv", quickExportHandler(db, "rooms", "csv"))
			protected.GET("/export/therapies/pdf", quickExportHandler(db, "therapies", "pdf"))
			protected.GET("/export/therapies/excel", quickExportHandler(db, "therapies", "excel"))
			protected.GET("/export/therapies/csv", quickExportHandler(db, "therapies", "csv"))
			protected.GET("/export/weights/pdf", quickExportHandler(db, "weights", "pdf"))
			protected.GET("/export/weights/excel", quickExportHandler(db, "weights", "excel"))
			protected.GET("/export/weights/csv", quickExportHandler(db, "weights", "csv"))

			// Notification routes  ← NUOVO
			protected.GET("/notifications", getNotificationsHandler(db))
			protected.PUT("/notifications/:id/read", markNotificationReadHandler(db))
			protected.DELETE("/notifications/:id", dismissNotificationHandler(db))
			protected.GET("/notifications/stats", getNotificationStatsHandler(db))
			protected.POST("/notifications/check", func(c *gin.Context) {
				ns := NewNotificationService(db)
				go ns.CheckAllNotifications()
				c.JSON(http.StatusOK, gin.H{"message": "Check triggered"})
			})

			// Analysis routes  ← NUOVO
			protected.GET("/analysis/weight", getWeightAnalysisHandler(db))
			protected.GET("/analysis/therapy", getTherapyAnalysisHandler(db))

			// Settings routes  ← NUOVO
			protected.GET("/notification-settings", getNotificationSettingsHandler(db))
			protected.PUT("/notification-settings", updateNotificationSettingsHandler(db))
		}
	}

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Frontend routes
	r.GET("/", indexHandler)
	r.GET("/login", loginPageHandler)
	r.GET("/hedgehogs", hedgehogsPageHandler)
	r.GET("/rooms", roomsPageHandler)
	r.GET("/room-builder", roomBuilderPageHandler)
	r.GET("/notifications", notificationsPageHandler) // ← NUOVO
	r.GET("/tutorial", docsPageHandler)

	return r
}

// Frontend page handlers
func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Centro Recupero Ricci La Ninna",
	})
}

func loginPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Login - La Ninna",
	})
}

func hedgehogsPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "hedgehogs.html", gin.H{
		"title": "Gestione Ricci - La Ninna",
	})
}

func roomsPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "rooms.html", gin.H{
		"title": "Stanze - La Ninna",
	})
}

func roomBuilderPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "room-builder.html", gin.H{
		"title": "Room Builder - La Ninna",
	})
}

func docsPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "tutorial.html", gin.H{
		"title": "Documentazione - La Ninna",
	})
}
