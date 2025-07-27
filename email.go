// email.go - Sistema invio email e webhook
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// Configurazione email
type EmailConfig struct {
	SMTPHost string
	SMTPPort string
	From     string
	Username string
	Password string
	Enabled  bool
}

// Template email per notifiche
const emailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .notification { background: #f8f9fa; border-left: 4px solid #007bff; padding: 15px; margin: 10px 0; }
        .critical { border-left-color: #dc3545; background: #f8d7da; }
        .high { border-left-color: #fd7e14; background: #fff3cd; }
        .footer { background: #f8f9fa; padding: 15px; text-align: center; font-size: 0.9em; color: #666; }
        .button { display: inline-block; background: #f59e0b; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; margin: 10px 0; }
        .hedgehog-icon { font-size: 2em; margin-bottom: 10px; }
    </style>
</head>
<body>
    <div class="header">
        <div class="hedgehog-icon">🦔</div>
        <h1>Centro Recupero Ricci "La Ninna"</h1>
        <p>Novello (CN) - Sistema di Notifiche</p>
    </div>
    
    <div class="content">
        <h2>{{.Subject}}</h2>
        
        <div class="notification {{.PriorityClass}}">
            <h3>{{.Title}}</h3>
            <p>{{.Message}}</p>
            
            {{if .HedgehogName}}
            <p><strong>Riccio:</strong> {{.HedgehogName}}</p>
            {{end}}
            
            {{if .TherapyName}}
            <p><strong>Terapia:</strong> {{.TherapyName}}</p>
            {{end}}
            
            <p><strong>Data:</strong> {{.CreatedAt}}</p>
            
            {{if .ActionURL}}
            <a href="{{.BaseURL}}{{.ActionURL}}" class="button">{{.ActionLabel}}</a>
            {{end}}
        </div>
        
        {{if .AdditionalData}}
        <div style="margin-top: 20px;">
            <h4>Dettagli aggiuntivi:</h4>
            <pre style="background: #f8f9fa; padding: 10px; border-radius: 5px;">{{.AdditionalData}}</pre>
        </div>
        {{end}}
    </div>
    
    <div class="footer">
        <p>Questo è un messaggio automatico dal sistema di gestione del Centro Ricci "La Ninna".</p>
        <p><a href="{{.BaseURL}}/notifications">Gestisci le notifiche</a> | <a href="{{.BaseURL}}/notification-settings">Impostazioni</a></p>
    </div>
</body>
</html>
`

// Struttura dati per template email
type EmailData struct {
	Subject        string
	Title          string
	Message        string
	PriorityClass  string
	HedgehogName   string
	TherapyName    string
	CreatedAt      string
	ActionURL      string
	ActionLabel    string
	BaseURL        string
	AdditionalData string
}

// Struttura webhook payload
type WebhookPayload struct {
	Source       string                 `json:"source"`
	Timestamp    string                 `json:"timestamp"`
	Notification NotificationWebhook    `json:"notification"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type NotificationWebhook struct {
	ID          uint   `json:"id"`
	Type        string `json:"type"`
	Priority    string `json:"priority"`
	Title       string `json:"title"`
	Message     string `json:"message"`
	HedgehogID  *uint  `json:"hedgehog_id,omitempty"`
	TherapyID   *uint  `json:"therapy_id,omitempty"`
	ActionURL   string `json:"action_url,omitempty"`
	ActionLabel string `json:"action_label,omitempty"`
}

// EmailService - Servizio email
type EmailService struct {
	config   EmailConfig
	template *template.Template
}

func NewEmailService() *EmailService {
	config := EmailConfig{
		SMTPHost: getEnv("EMAIL_SMTP_HOST", "smtp.gmail.com"),
		SMTPPort: getEnv("EMAIL_SMTP_PORT", "587"),
		From:     getEnv("EMAIL_FROM", "notifications@laninna.it"),
		Username: getEnv("EMAIL_USERNAME", ""),
		Password: getEnv("EMAIL_PASSWORD", ""),
		Enabled:  getEnv("EMAIL_USERNAME", "") != "",
	}

	tmpl, err := template.New("email").Parse(emailTemplate)
	if err != nil {
		log.Printf("Errore parsing template email: %v", err)
		return nil
	}

	return &EmailService{
		config:   config,
		template: tmpl,
	}
}

func (es *EmailService) SendNotificationEmail(notification Notification, toEmail string) error {
	if !es.config.Enabled || toEmail == "" {
		log.Println("📧 Email non configurata o destinatario mancante")
		return nil
	}

	// Prepara dati per template
	data := EmailData{
		Subject:       fmt.Sprintf("[La Ninna] %s", notification.Title),
		Title:         notification.Title,
		Message:       notification.Message,
		PriorityClass: getPriorityClass(string(notification.Priority)),
		CreatedAt:     notification.CreatedAt.Format("02/01/2006 15:04"),
		ActionURL:     notification.ActionURL,
		ActionLabel:   notification.ActionLabel,
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),
	}

	if notification.Hedgehog != nil {
		data.HedgehogName = notification.Hedgehog.Name
	}

	if notification.Therapy != nil {
		data.TherapyName = notification.Therapy.Name
	}

	if notification.Data != "" {
		data.AdditionalData = notification.Data
	}

	// Genera HTML da template
	var htmlBuffer bytes.Buffer
	if err := es.template.Execute(&htmlBuffer, data); err != nil {
		log.Printf("Errore generazione template email: %v", err)
		return err
	}

	// Invia email
	return es.sendSMTP(toEmail, data.Subject, htmlBuffer.String())
}

func (es *EmailService) sendSMTP(to, subject, htmlBody string) error {
	// Messaggio email con headers
	message := fmt.Sprintf(`To: %s
Subject: %s
MIME-Version: 1.0
Content-Type: text/html; charset=UTF-8

%s`, to, subject, htmlBody)

	// Autenticazione SMTP
	auth := smtp.PlainAuth("", es.config.Username, es.config.Password, es.config.SMTPHost)

	// Invia email
	addr := fmt.Sprintf("%s:%s", es.config.SMTPHost, es.config.SMTPPort)
	err := smtp.SendMail(addr, auth, es.config.From, []string{to}, []byte(message))

	if err != nil {
		log.Printf("❌ Errore invio email: %v", err)
		return err
	}

	log.Printf("✅ Email inviata a %s: %s", to, subject)
	return nil
}

func getPriorityClass(priority string) string {
	switch priority {
	case "critical":
		return "critical"
	case "high":
		return "high"
	default:
		return ""
	}
}

// WebhookService - Servizio webhook
type WebhookService struct {
	client  *http.Client
	timeout time.Duration
	retries int
}

func NewWebhookService() *WebhookService {
	timeout := 10 * time.Second
	if envTimeout := getEnv("WEBHOOK_TIMEOUT", ""); envTimeout != "" {
		if t, err := time.ParseDuration(envTimeout); err == nil {
			timeout = t
		}
	}

	retries := 3
	if envRetries := getEnv("WEBHOOK_RETRY_ATTEMPTS", ""); envRetries != "" {
		fmt.Sscanf(envRetries, "%d", &retries)
	}

	return &WebhookService{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
		retries: retries,
	}
}

func (ws *WebhookService) SendWebhook(notification Notification, webhookURL string) error {
	if webhookURL == "" {
		return nil
	}

	// Prepara payload
	payload := WebhookPayload{
		Source:    "laninna-hedgehog-app",
		Timestamp: time.Now().Format(time.RFC3339),
		Notification: NotificationWebhook{
			ID:          notification.ID,
			Type:        string(notification.Type),
			Priority:    string(notification.Priority),
			Title:       notification.Title,
			Message:     notification.Message,
			HedgehogID:  notification.HedgehogID,
			TherapyID:   notification.TherapyID,
			ActionURL:   notification.ActionURL,
			ActionLabel: notification.ActionLabel,
		},
		Metadata: map[string]interface{}{
			"app_version": "1.0.0",
			"environment": getEnv("GIN_MODE", "development"),
		},
	}

	// Serializza JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Errore serializzazione webhook: %v", err)
		return err
	}

	// Invia con retry
	for attempt := 1; attempt <= ws.retries; attempt++ {
		if err := ws.sendWebhookAttempt(webhookURL, jsonData); err != nil {
			log.Printf("🔗 Webhook attempt %d/%d failed: %v", attempt, ws.retries, err)
			if attempt == ws.retries {
				return err
			}
			time.Sleep(time.Duration(attempt) * time.Second) // Backoff exponenziale
		} else {
			log.Printf("✅ Webhook inviato con successo a %s", webhookURL)
			return nil
		}
	}

	return fmt.Errorf("webhook failed after %d attempts", ws.retries)
}

func (ws *WebhookService) sendWebhookAttempt(url string, data []byte) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "La-Ninna-Hedgehog-App/1.0")
	req.Header.Set("X-La-Ninna-Source", "notification-system")

	resp, err := ws.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// Aggiorna NotificationService per usare email e webhook
func (ns *NotificationService) sendExternalNotifications(notification Notification) {
	// Invia email se abilitato
	if ns.settings.EmailNotificationsEnabled && ns.settings.EmailAddress != "" {
		emailService := NewEmailService()
		if emailService != nil {
			if err := emailService.SendNotificationEmail(notification, ns.settings.EmailAddress); err != nil {
				log.Printf("❌ Errore invio email: %v", err)
			}
		}
	}

	// Invia webhook se configurato
	if ns.settings.WebhookURL != "" {
		webhookService := NewWebhookService()
		if err := webhookService.SendWebhook(notification, ns.settings.WebhookURL); err != nil {
			log.Printf("❌ Errore invio webhook: %v", err)
		}
	}
}

// Utility function per variabili ambiente
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Test endpoints per email e webhook
func testEmailHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Email string `json:"email" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Crea notifica di test
		testNotification := Notification{
			Type:        NotificationSystemAlert,
			Priority:    PriorityMedium,
			Title:       "Test Notifica Email",
			Message:     "Questa è una notifica di test per verificare la configurazione email del sistema La Ninna.",
			ActionURL:   "/notifications",
			ActionLabel: "Visualizza Notifiche",
			CreatedAt:   time.Now(),
		}

		emailService := NewEmailService()
		if emailService == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Servizio email non configurato"})
			return
		}

		if err := emailService.SendNotificationEmail(testNotification, request.Email); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Errore invio email: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Email di test inviata con successo"})
	}
}

func testWebhookHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			WebhookURL string `json:"webhook_url" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Crea notifica di test
		testNotification := Notification{
			Type:        NotificationSystemAlert,
			Priority:    PriorityMedium,
			Title:       "Test Notifica Webhook",
			Message:     "Questa è una notifica di test per verificare la configurazione webhook del sistema La Ninna.",
			ActionURL:   "/notifications",
			ActionLabel: "Visualizza Notifiche",
			CreatedAt:   time.Now(),
		}

		webhookService := NewWebhookService()
		if err := webhookService.SendWebhook(testNotification, request.WebhookURL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Errore invio webhook: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Webhook di test inviato con successo"})
	}
}

// Batch notification sender per riassunti periodici
type BatchNotificationService struct {
	db           *gorm.DB
	emailService *EmailService
}

func NewBatchNotificationService(db *gorm.DB) *BatchNotificationService {
	return &BatchNotificationService{
		db:           db,
		emailService: NewEmailService(),
	}
}

func (bns *BatchNotificationService) SendDailySummary() error {
	log.Println("📊 Sending daily notification summary...")

	// Ottieni impostazioni
	var settings NotificationSettings
	if err := bns.db.First(&settings).Error; err != nil {
		return err
	}

	if !settings.EmailNotificationsEnabled || settings.EmailAddress == "" {
		return nil
	}

	// Raccogli notifiche ultime 24 ore
	yesterday := time.Now().AddDate(0, 0, -1)
	var notifications []Notification
	bns.db.Preload("Hedgehog").Preload("Therapy").
		Where("created_at >= ? AND dismissed = false", yesterday).
		Order("priority DESC, created_at DESC").
		Find(&notifications)

	if len(notifications) == 0 {
		log.Println("📧 No notifications to summarize")
		return nil
	}

	// Genera email riassunto
	return bns.sendSummaryEmail(notifications, settings.EmailAddress)
}

func (bns *BatchNotificationService) sendSummaryEmail(notifications []Notification, toEmail string) error {
	if bns.emailService == nil {
		return fmt.Errorf("email service not configured")
	}

	// Conta per priorità
	counts := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
	}

	for _, n := range notifications {
		counts[string(n.Priority)]++
	}

	// Template HTML per riassunto
	summaryTemplate := `
    <!DOCTYPE html>
    <html>
    <head>
        <meta charset="UTF-8">
        <style>
            body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
            .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 20px; text-align: center; }
            .summary { background: #f8f9fa; padding: 20px; margin: 20px 0; border-radius: 8px; }
            .stat { display: inline-block; margin: 10px; padding: 15px; background: white; border-radius: 8px; text-align: center; min-width: 100px; }
            .critical { border-left: 4px solid #dc3545; }
            .high { border-left: 4px solid #fd7e14; }
            .medium { border-left: 4px solid #0d6efd; }
            .notification-item { background: white; margin: 10px 0; padding: 15px; border-radius: 8px; border-left: 4px solid #dee2e6; }
            .footer { background: #f8f9fa; padding: 15px; text-align: center; font-size: 0.9em; color: #666; }
        </style>
    </head>
    <body>
        <div class="header">
            <div style="font-size: 2em; margin-bottom: 10px;">🦔</div>
            <h1>Riassunto Giornaliero - La Ninna</h1>
            <p>{{.Date}}</p>
        </div>
        
        <div style="padding: 20px;">
            <div class="summary">
                <h2>Statistiche Notifiche</h2>
                <div>
                    <div class="stat critical">
                        <div style="font-size: 1.5em; font-weight: bold;">{{.Critical}}</div>
                        <div>Critiche</div>
                    </div>
                    <div class="stat high">
                        <div style="font-size: 1.5em; font-weight: bold;">{{.High}}</div>
                        <div>Importanti</div>
                    </div>
                    <div class="stat medium">
                        <div style="font-size: 1.5em; font-weight: bold;">{{.Medium}}</div>
                        <div>Normali</div>
                    </div>
                    <div class="stat">
                        <div style="font-size: 1.5em; font-weight: bold;">{{.Total}}</div>
                        <div>Totali</div>
                    </div>
                </div>
            </div>
            
            {{if .HasCritical}}
            <h3 style="color: #dc3545;">🚨 Notifiche Critiche</h3>
            {{range .CriticalNotifications}}
            <div class="notification-item critical">
                <h4>{{.Title}}</h4>
                <p>{{.Message}}</p>
                {{if .Hedgehog}}<p><strong>Riccio:</strong> {{.Hedgehog.Name}}</p>{{end}}
                <p style="font-size: 0.9em; color: #666;">{{.CreatedAt.Format "15:04"}}</p>
            </div>
            {{end}}
            {{end}}
            
            {{if .HasHigh}}
            <h3 style="color: #fd7e14;">⚠️ Notifiche Importanti</h3>
            {{range .HighNotifications}}
            <div class="notification-item high">
                <h4>{{.Title}}</h4>
                <p>{{.Message}}</p>
                {{if .Hedgehog}}<p><strong>Riccio:</strong> {{.Hedgehog.Name}}</p>{{end}}
                <p style="font-size: 0.9em; color: #666;">{{.CreatedAt.Format "15:04"}}</p>
            </div>
            {{end}}
            {{end}}
        </div>
        
        <div class="footer">
            <p>Centro Recupero Ricci "La Ninna" - Novello (CN)</p>
            <p><a href="{{.BaseURL}}/notifications">Visualizza tutte le notifiche</a></p>
        </div>
    </body>
    </html>
    `

	// Prepara dati template
	criticalNotifications := make([]Notification, 0)
	highNotifications := make([]Notification, 0)

	for _, n := range notifications {
		if n.Priority == PriorityCritical {
			criticalNotifications = append(criticalNotifications, n)
		} else if n.Priority == PriorityHigh {
			highNotifications = append(highNotifications, n)
		}
	}

	data := struct {
		Date                  string
		Critical              int
		High                  int
		Medium                int
		Total                 int
		HasCritical           bool
		HasHigh               bool
		CriticalNotifications []Notification
		HighNotifications     []Notification
		BaseURL               string
	}{
		Date:                  time.Now().Format("02/01/2006"),
		Critical:              counts["critical"],
		High:                  counts["high"],
		Medium:                counts["medium"] + counts["low"],
		Total:                 len(notifications),
		HasCritical:           len(criticalNotifications) > 0,
		HasHigh:               len(highNotifications) > 0,
		CriticalNotifications: criticalNotifications,
		HighNotifications:     highNotifications,
		BaseURL:               getEnv("BASE_URL", "http://localhost:8080"),
	}

	// Genera HTML
	tmpl, err := template.New("summary").Parse(summaryTemplate)
	if err != nil {
		return err
	}

	var htmlBuffer bytes.Buffer
	if err := tmpl.Execute(&htmlBuffer, data); err != nil {
		return err
	}

	// Invia email
	subject := fmt.Sprintf("[La Ninna] Riassunto Giornaliero - %d notifiche", len(notifications))
	return bns.emailService.sendSMTP(toEmail, subject, htmlBuffer.String())
}

// Scheduler per riassunti periodici
func StartBatchNotificationScheduler(db *gorm.DB) {
	batchService := NewBatchNotificationService(db)

	// Invia riassunto giornaliero alle 08:00
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
			if next.Before(now) {
				next = next.Add(24 * time.Hour)
			}

			time.Sleep(time.Until(next))

			if err := batchService.SendDailySummary(); err != nil {
				log.Printf("❌ Errore invio riassunto giornaliero: %v", err)
			}
		}
	}()

	log.Println("📅 Batch notification scheduler started (daily summary at 08:00)")
}

// Health check per servizi esterni
type ExternalServicesHealth struct {
	Email   bool   `json:"email"`
	Webhook bool   `json:"webhook"`
	SMTP    string `json:"smtp_status"`
}

func checkExternalServicesHealth() ExternalServicesHealth {
	health := ExternalServicesHealth{
		Email:   false,
		Webhook: true, // Webhook è sempre disponibile se configurato
		SMTP:    "not_configured",
	}

	// Controlla configurazione email
	if getEnv("EMAIL_USERNAME", "") != "" {
		health.Email = true
		health.SMTP = "configured"

		// Test connessione SMTP (opzionale)
		emailService := NewEmailService()
		if emailService != nil {
			// Qui potresti aggiungere un ping SMTP
			health.SMTP = "ready"
		}
	}

	return health
}

// Aggiungi nuove route al router
func addNotificationSystemRoutes(r *gin.Engine, db *gorm.DB) {
	api := r.Group("/api")

	// Health check pubblico
	r.GET("/health", healthCheckHandler())

	protected := api.Group("/")
	protected.Use(authMiddleware())
	{
		// Test servizi esterni
		protected.POST("/test/email", testEmailHandler(db))
		protected.POST("/test/webhook", testWebhookHandler(db))

		// Trigger riassunto manuale
		protected.POST("/notifications/summary", func(c *gin.Context) {
			batchService := NewBatchNotificationService(db)
			go func() {
				if err := batchService.SendDailySummary(); err != nil {
					log.Printf("Errore riassunto manuale: %v", err)
				}
			}()
			c.JSON(http.StatusOK, gin.H{"message": "Daily summary triggered"})
		})

		// Statistiche avanzate notifiche
		protected.GET("/notifications/analytics", func(c *gin.Context) {
			var stats struct {
				Last24Hours int64                    `json:"last_24_hours"`
				LastWeek    int64                    `json:"last_week"`
				ByType      map[string]int64         `json:"by_type"`
				ByPriority  map[string]int64         `json:"by_priority"`
				TrendData   []map[string]interface{} `json:"trend_data"`
			}

			now := time.Now()
			yesterday := now.AddDate(0, 0, -1)
			lastWeek := now.AddDate(0, 0, -7)

			// Conteggi base
			db.Model(&Notification{}).Where("created_at >= ?", yesterday).Count(&stats.Last24Hours)
			db.Model(&Notification{}).Where("created_at >= ?", lastWeek).Count(&stats.LastWeek)

			// Per tipo
			var typeCounts []struct {
				Type  string `json:"type"`
				Count int64  `json:"count"`
			}
			db.Model(&Notification{}).
				Select("type, count(*) as count").
				Where("created_at >= ?", lastWeek).
				Group("type").
				Find(&typeCounts)

			stats.ByType = make(map[string]int64)
			for _, tc := range typeCounts {
				stats.ByType[tc.Type] = tc.Count
			}

			// Per priorità
			var priorityCounts []struct {
				Priority string `json:"priority"`
				Count    int64  `json:"count"`
			}
			db.Model(&Notification{}).
				Select("priority, count(*) as count").
				Where("created_at >= ?", lastWeek).
				Group("priority").
				Find(&priorityCounts)

			stats.ByPriority = make(map[string]int64)
			for _, pc := range priorityCounts {
				stats.ByPriority[pc.Priority] = pc.Count
			}

			// Trend ultimi 7 giorni
			for i := 6; i >= 0; i-- {
				day := now.AddDate(0, 0, -i)
				dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
				dayEnd := dayStart.Add(24 * time.Hour)

				var count int64
				db.Model(&Notification{}).
					Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
					Count(&count)

				stats.TrendData = append(stats.TrendData, map[string]interface{}{
					"date":  day.Format("2006-01-02"),
					"count": count,
				})
			}

			c.JSON(http.StatusOK, stats)
		})
	}
}

// Notification middleware per intercettare eventi CRUD
func NotificationMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Esegui handler
		c.Next()

		// Solo per operazioni POST/PUT/DELETE con successo
		if c.Writer.Status() < 200 || c.Writer.Status() >= 300 {
			return
		}

		method := c.Request.Method
		path := c.Request.URL.Path

		// Trigger controllo notifiche per operazioni critiche
		go func() {
			ns := NewNotificationService(db)

			// Controlli specifici per endpoint
			if strings.Contains(path, "/weight-records") && method == "POST" {
				// Nuova pesatura - controlla trend peso
				ns.checkWeightNotifications()
			} else if strings.Contains(path, "/therapies") && (method == "POST" || method == "PUT") {
				// Terapia modificata - controlla scadenze
				ns.checkTherapyNotifications()
			} else if strings.Contains(path, "/hedgehogs") && method == "PUT" {
				// Riccio modificato - controlli generali
				ns.checkMissingWeighings()
			}
		}()
	}
}

// Funzione di inizializzazione completa sistema notifiche
func InitNotificationSystem(r *gin.Engine, db *gorm.DB) {
	log.Println("🔔 Initializing notification system...")

	// Aggiungi middleware
	r.Use(NotificationMiddleware(db))

	// Avvia schedulers
	StartNotificationScheduler(db)
	StartBatchNotificationScheduler(db)

	// Aggiungi route
	addNotificationSystemRoutes(r, db)

	log.Println("✅ Notification system initialized successfully")
}
