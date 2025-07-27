// notifications.go - Sistema Notifiche Completo
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Strutture per analisi dati
type WeightAnalysis struct {
	HedgehogID     uint      `json:"hedgehog_id"`
	HedgehogName   string    `json:"hedgehog_name"`
	CurrentWeight  float64   `json:"current_weight"`
	PreviousWeight float64   `json:"previous_weight"`
	WeightChange   float64   `json:"weight_change"`
	DaysSinceWeigh int       `json:"days_since_weigh"`
	LastWeighDate  time.Time `json:"last_weigh_date"`
	Trend          string    `json:"trend"` // "improving", "stable", "declining", "critical"
	Alert          bool      `json:"alert"`
	AlertReason    string    `json:"alert_reason"`
}

type TherapyAnalysis struct {
	TherapyID    uint       `json:"therapy_id"`
	HedgehogID   uint       `json:"hedgehog_id"`
	HedgehogName string     `json:"hedgehog_name"`
	TherapyName  string     `json:"therapy_name"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	DaysActive   int        `json:"days_active"`
	DaysUntilEnd int        `json:"days_until_end"`
	Status       string     `json:"status"`
	Alert        bool       `json:"alert"`
	AlertReason  string     `json:"alert_reason"`
}

// NotificationService - Servizio principale
type NotificationService struct {
	db       *gorm.DB
	settings *NotificationSettings
}

func NewNotificationService(db *gorm.DB) *NotificationService {
	service := &NotificationService{db: db}
	service.loadSettings()
	return service
}

func (ns *NotificationService) loadSettings() {
	var settings NotificationSettings
	if err := ns.db.First(&settings).Error; err != nil {
		// Crea impostazioni di default
		settings = NotificationSettings{
			TherapyExpiredEnabled:     true,
			TherapyExpiringDays:       3,
			WeightDropThreshold:       50,
			WeightDropDays:            7,
			WeightStagnationDays:      14,
			NoWeighingDays:            7,
			EmailNotificationsEnabled: false,
		}
		ns.db.Create(&settings)
	}
	ns.settings = &settings
}

// Checker principale per tutte le notifiche
func (ns *NotificationService) CheckAllNotifications() error {
	log.Println("🔍 Running notification checks...")

	// Pulisci notifiche vecchie
	ns.cleanOldNotifications()

	// Controlla terapie
	if err := ns.checkTherapyNotifications(); err != nil {
		log.Printf("Errore controllo terapie: %v", err)
	}

	// Controlla peso
	if err := ns.checkWeightNotifications(); err != nil {
		log.Printf("Errore controllo peso: %v", err)
	}

	// Controlla pesature mancanti
	if err := ns.checkMissingWeighings(); err != nil {
		log.Printf("Errore controllo pesature: %v", err)
	}

	log.Println("✅ Notification checks completed")
	return nil
}

func (ns *NotificationService) checkTherapyNotifications() error {
	if !ns.settings.TherapyExpiredEnabled {
		return nil
	}

	var therapies []Therapy
	ns.db.Where("status = 'active'").Find(&therapies)

	now := time.Now()
	expiringThreshold := now.AddDate(0, 0, ns.settings.TherapyExpiringDays)

	for _, therapy := range therapies {
		if therapy.EndDate == nil {
			continue
		}

		// Get hedgehog name by ID
		var hedgehog Hedgehog
		hedgehogName := "N/A"
		if err := ns.db.First(&hedgehog, therapy.HedgehogID).Error; err == nil {
			hedgehogName = hedgehog.Name
		}

		endDate := *therapy.EndDate

		// Terapia scaduta
		if endDate.Before(now) {
			ns.createNotification(Notification{
				Type:        NotificationTherapyExpired,
				Priority:    PriorityHigh,
				Title:       fmt.Sprintf("Terapia Scaduta: %s", therapy.Name),
				Message:     fmt.Sprintf("La terapia '%s' per %s è scaduta il %s", therapy.Name, hedgehogName, endDate.Format("02/01/2006")),
				HedgehogID:  &therapy.HedgehogID,
				TherapyID:   &therapy.ID,
				ActionURL:   fmt.Sprintf("/hedgehogs/%d", therapy.HedgehogID),
				ActionLabel: "Gestisci Terapia",
				Data:        fmt.Sprintf(`{"days_overdue": %d}`, int(now.Sub(endDate).Hours()/24)),
			})
		} else if endDate.Before(expiringThreshold) {
			// Terapia in scadenza
			daysLeft := int(endDate.Sub(now).Hours() / 24)
			ns.createNotification(Notification{
				Type:        NotificationTherapyExpiring,
				Priority:    PriorityMedium,
				Title:       fmt.Sprintf("Terapia in Scadenza: %s", therapy.Name),
				Message:     fmt.Sprintf("La terapia '%s' per %s scadrà tra %d giorni", therapy.Name, hedgehogName, daysLeft),
				HedgehogID:  &therapy.HedgehogID,
				TherapyID:   &therapy.ID,
				ActionURL:   fmt.Sprintf("/hedgehogs/%d", therapy.HedgehogID),
				ActionLabel: "Rinnova Terapia",
				Data:        fmt.Sprintf(`{"days_left": %d}`, daysLeft),
			})
		}
	}

	return nil
}

func (ns *NotificationService) checkWeightNotifications() error {
	analyses := ns.analyzeWeightTrends()

	for _, analysis := range analyses {
		if !analysis.Alert {
			continue
		}

		var priority NotificationPriority
		var notifType NotificationType

		if analysis.WeightChange <= -ns.settings.WeightDropThreshold {
			priority = PriorityCritical
			notifType = NotificationWeightDrop
		} else {
			priority = PriorityMedium
			notifType = NotificationWeightStagnation
		}

		// Evita notifiche duplicate recenti
		if ns.hasRecentNotification(analysis.HedgehogID, notifType, 24*time.Hour) {
			continue
		}

		data, _ := json.Marshal(analysis)
		ns.createNotification(Notification{
			Type:        notifType,
			Priority:    priority,
			Title:       fmt.Sprintf("Allarme Peso: %s", analysis.HedgehogName),
			Message:     analysis.AlertReason,
			HedgehogID:  &analysis.HedgehogID,
			ActionURL:   fmt.Sprintf("/hedgehogs/%d", analysis.HedgehogID),
			ActionLabel: "Controlla Peso",
			Data:        string(data),
		})
	}

	return nil
}

func (ns *NotificationService) checkMissingWeighings() error {
	var hedgehogs []Hedgehog
	ns.db.Where("status = 'in_care'").Find(&hedgehogs)

	threshold := time.Now().AddDate(0, 0, -ns.settings.NoWeighingDays)

	for _, hedgehog := range hedgehogs {
		var lastWeight WeightRecord
		err := ns.db.Where("hedgehog_id = ?", hedgehog.ID).
			Order("date DESC").
			First(&lastWeight).Error

		if err != nil || lastWeight.Date.Before(threshold) {
			daysSince := int(time.Since(lastWeight.Date).Hours() / 24)
			if err != nil {
				daysSince = int(time.Since(hedgehog.ArrivalDate).Hours() / 24)
			}

			// Evita spam di notifiche
			if ns.hasRecentNotification(hedgehog.ID, NotificationNoWeighing, 48*time.Hour) {
				continue
			}

			ns.createNotification(Notification{
				Type:        NotificationNoWeighing,
				Priority:    PriorityMedium,
				Title:       fmt.Sprintf("Pesatura Mancante: %s", hedgehog.Name),
				Message:     fmt.Sprintf("%s non viene pesato da %d giorni", hedgehog.Name, daysSince),
				HedgehogID:  &hedgehog.ID,
				ActionURL:   fmt.Sprintf("/hedgehogs/%d", hedgehog.ID),
				ActionLabel: "Aggiungi Pesatura",
				Data:        fmt.Sprintf(`{"days_since": %d}`, daysSince),
			})
		}
	}

	return nil
}

func (ns *NotificationService) analyzeWeightTrends() []WeightAnalysis {
	var hedgehogs []Hedgehog
	ns.db.Where("status = 'in_care'").Find(&hedgehogs)

	var analyses []WeightAnalysis

	for _, hedgehog := range hedgehogs {
		var weights []WeightRecord
		ns.db.Where("hedgehog_id = ?", hedgehog.ID).
			Order("date DESC").
			Limit(10).
			Find(&weights)

		if len(weights) < 2 {
			continue
		}

		analysis := WeightAnalysis{
			HedgehogID:     hedgehog.ID,
			HedgehogName:   hedgehog.Name,
			CurrentWeight:  weights[0].Weight,
			PreviousWeight: weights[1].Weight,
			WeightChange:   weights[0].Weight - weights[1].Weight,
			DaysSinceWeigh: int(time.Since(weights[0].Date).Hours() / 24),
			LastWeighDate:  weights[0].Date,
		}

		// Analisi trend
		analysis.Trend = ns.calculateWeightTrend(weights)

		// Controllo allarmi
		if analysis.WeightChange <= -ns.settings.WeightDropThreshold {
			analysis.Alert = true
			analysis.AlertReason = fmt.Sprintf("Perdita di peso significativa: %.1fg in %d giorni",
				math.Abs(analysis.WeightChange),
				int(weights[0].Date.Sub(weights[1].Date).Hours()/24))
		} else if analysis.Trend == "declining" && len(weights) >= 3 {
			// Controllo trend negativo su più pesature
			recentWeights := weights[:3]
			declining := true
			for i := 0; i < len(recentWeights)-1; i++ {
				if recentWeights[i].Weight >= recentWeights[i+1].Weight {
					declining = false
					break
				}
			}
			if declining {
				analysis.Alert = true
				analysis.AlertReason = "Trend di peso in calo nelle ultime pesature"
			}
		}

		// Controllo stagnazione
		if analysis.Trend == "stable" && len(weights) >= 4 {
			weekSpan := weights[0].Date.Sub(weights[3].Date)
			if weekSpan.Hours() > float64(ns.settings.WeightStagnationDays*24) {
				variation := ns.calculateWeightVariation(weights[:4])
				if variation < 10 { // Meno di 10g di variazione
					analysis.Alert = true
					analysis.AlertReason = fmt.Sprintf("Peso stagnante da %d giorni (variazione: %.1fg)",
						int(weekSpan.Hours()/24), variation)
				}
			}
		}

		analyses = append(analyses, analysis)
	}

	return analyses
}

func (ns *NotificationService) calculateWeightTrend(weights []WeightRecord) string {
	if len(weights) < 3 {
		return "insufficient_data"
	}

	// Calcola trend su ultimi 5 punti
	recentWeights := weights
	if len(weights) > 5 {
		recentWeights = weights[:5]
	}

	// Calcola pendenza media
	var slopes []float64
	for i := 0; i < len(recentWeights)-1; i++ {
		daysDiff := recentWeights[i].Date.Sub(recentWeights[i+1].Date).Hours() / 24
		if daysDiff > 0 {
			slope := (recentWeights[i].Weight - recentWeights[i+1].Weight) / daysDiff
			slopes = append(slopes, slope)
		}
	}

	if len(slopes) == 0 {
		return "stable"
	}

	// Media delle pendenze
	avgSlope := 0.0
	for _, slope := range slopes {
		avgSlope += slope
	}
	avgSlope /= float64(len(slopes))

	if avgSlope > 2 {
		return "improving"
	} else if avgSlope < -2 {
		return "declining"
	} else if avgSlope < -5 {
		return "critical"
	}
	return "stable"
}

func (ns *NotificationService) calculateWeightVariation(weights []WeightRecord) float64 {
	if len(weights) < 2 {
		return 0
	}

	min := weights[0].Weight
	max := weights[0].Weight

	for _, w := range weights {
		if w.Weight < min {
			min = w.Weight
		}
		if w.Weight > max {
			max = w.Weight
		}
	}

	return max - min
}

func (ns *NotificationService) createNotification(notification Notification) {
	// Controlla duplicati recenti
	if ns.hasRecentNotification(*notification.HedgehogID, notification.Type, 24*time.Hour) {
		return
	}

	// Imposta scadenza automatica
	if notification.ExpiresAt == nil {
		expiry := time.Now().AddDate(0, 0, 30) // 30 giorni
		notification.ExpiresAt = &expiry
	}

	if err := ns.db.Create(&notification).Error; err != nil {
		log.Printf("Errore creazione notifica: %v", err)
		return
	}

	log.Printf("📢 Notifica creata: %s - %s", notification.Type, notification.Title)

	// Invia notifiche esterne se abilitate
	go ns.sendExternalNotifications(notification)
}

func (ns *NotificationService) hasRecentNotification(hedgehogID uint, notifType NotificationType, duration time.Duration) bool {
	var count int64
	since := time.Now().Add(-duration)

	ns.db.Model(&Notification{}).
		Where("hedgehog_id = ? AND type = ? AND created_at > ?", hedgehogID, notifType, since).
		Count(&count)

	return count > 0
}

func (ns *NotificationService) cleanOldNotifications() {
	// Elimina notifiche scadute
	ns.db.Where("expires_at < ?", time.Now()).Delete(&Notification{})

	// Elimina notifiche lette vecchie di 30 giorni
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	ns.db.Where("read = true AND created_at < ?", thirtyDaysAgo).Delete(&Notification{})
}

func (ns *NotificationService) sendEmailNotification(notification Notification) {
	// Placeholder per invio email
	log.Printf("📧 Email notification sent: %s", notification.Title)
}

func (ns *NotificationService) sendWebhookNotification(notification Notification) {
	// Placeholder per webhook
	log.Printf("🔗 Webhook notification sent: %s", notification.Title)
}

// API Handlers
// @Summary Get notifications
// @Description Get list of notifications with optional filters
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param unread query boolean false "Filter by unread status"
// @Param priority query string false "Filter by priority"
// @Param type query string false "Filter by notification type"
// @Param limit query int false "Limit results" default(50)
// @Success 200 {array} Notification
// @Failure 401 {object} map[string]string
// @Router /notifications [get]
func getNotificationsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var notifications []Notification
		query := db

		// Filtri query
		if unreadOnly := c.Query("unread"); unreadOnly == "true" {
			query = query.Where("read = false")
		}

		if priority := c.Query("priority"); priority != "" {
			query = query.Where("priority = ?", priority)
		}

		if notifType := c.Query("type"); notifType != "" {
			query = query.Where("type = ?", notifType)
		}

		limit := 50 // Default limit
		if l := c.Query("limit"); l != "" {
			fmt.Sscanf(l, "%d", &limit)
		}

		query.Where("dismissed = false").
			Order("priority DESC, created_at DESC").
			Limit(limit).
			Find(&notifications)

		c.JSON(http.StatusOK, notifications)
	}
}

// @Summary Mark notification as read
// @Description Mark a notification as read by its ID
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Success 200 {object} Notification
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /notifications/{id}/read [put]
func markNotificationReadHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var notification Notification
		if err := db.First(&notification, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}

		notification.Read = true
		if err := db.Save(&notification).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, notification)
	}
}

// @Summary Dismiss notification
// @Description Dismiss a notification by its ID
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /notifications/{id} [delete]
func dismissNotificationHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var notification Notification
		if err := db.First(&notification, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}

		notification.Dismissed = true
		if err := db.Save(&notification).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Notification dismissed"})
	}
}

// @Summary Get notification statistics
// @Description Get statistics about notifications (counts by status, priority, and type)
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object
// @Failure 401 {object} map[string]string
// @Router /notifications/stats [get]
func getNotificationStatsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var stats struct {
			Total    int64            `json:"total"`
			Unread   int64            `json:"unread"`
			Critical int64            `json:"critical"`
			High     int64            `json:"high"`
			ByType   map[string]int64 `json:"by_type"`
		}

		db.Model(&Notification{}).Where("dismissed = false").Count(&stats.Total)
		db.Model(&Notification{}).Where("dismissed = false AND read = false").Count(&stats.Unread)
		db.Model(&Notification{}).Where("dismissed = false AND priority = 'critical'").Count(&stats.Critical)
		db.Model(&Notification{}).Where("dismissed = false AND priority = 'high'").Count(&stats.High)

		// Conta per tipo
		var typeCounts []struct {
			Type  string `json:"type"`
			Count int64  `json:"count"`
		}
		db.Model(&Notification{}).
			Select("type, count(*) as count").
			Where("dismissed = false").
			Group("type").
			Find(&typeCounts)

		stats.ByType = make(map[string]int64)
		for _, tc := range typeCounts {
			stats.ByType[tc.Type] = tc.Count
		}

		c.JSON(http.StatusOK, stats)
	}
}

// @Summary Get weight analysis
// @Description Get analysis of hedgehog weight trends and alerts
// @Tags Analysis
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} WeightAnalysis
// @Failure 401 {object} map[string]string
// @Router /analysis/weight [get]
func getWeightAnalysisHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ns := NewNotificationService(db)
		analyses := ns.analyzeWeightTrends()

		// Ordina per priorità (alert prima)
		sort.Slice(analyses, func(i, j int) bool {
			if analyses[i].Alert != analyses[j].Alert {
				return analyses[i].Alert
			}
			return analyses[i].WeightChange < analyses[j].WeightChange
		})

		c.JSON(http.StatusOK, analyses)
	}
}

// @Summary Get therapy analysis
// @Description Get analysis of active therapies with alerts for expiring or overdue therapies
// @Tags Analysis
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} TherapyAnalysis
// @Failure 401 {object} map[string]string
// @Router /analysis/therapy [get]
func getTherapyAnalysisHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var therapies []Therapy
		db.Where("status = 'active'").Find(&therapies)

		var analyses []TherapyAnalysis
		now := time.Now()

		for _, therapy := range therapies {
			// Get hedgehog name by ID
			var hedgehog Hedgehog
			hedgehogName := "N/A"
			if err := db.First(&hedgehog, therapy.HedgehogID).Error; err == nil {
				hedgehogName = hedgehog.Name
			}

			analysis := TherapyAnalysis{
				TherapyID:    therapy.ID,
				HedgehogID:   therapy.HedgehogID,
				HedgehogName: hedgehogName,
				TherapyName:  therapy.Name,
				StartDate:    therapy.StartDate,
				EndDate:      therapy.EndDate,
				DaysActive:   int(now.Sub(therapy.StartDate).Hours() / 24),
				Status:       therapy.Status,
			}

			if therapy.EndDate != nil {
				analysis.DaysUntilEnd = int(therapy.EndDate.Sub(now).Hours() / 24)

				if therapy.EndDate.Before(now) {
					analysis.Alert = true
					analysis.AlertReason = fmt.Sprintf("Terapia scaduta da %d giorni", -analysis.DaysUntilEnd)
				} else if analysis.DaysUntilEnd <= 3 {
					analysis.Alert = true
					analysis.AlertReason = fmt.Sprintf("Terapia in scadenza tra %d giorni", analysis.DaysUntilEnd)
				}
			} else if analysis.DaysActive > 30 {
				analysis.Alert = true
				analysis.AlertReason = fmt.Sprintf("Terapia attiva da %d giorni senza data fine", analysis.DaysActive)
			}

			analyses = append(analyses, analysis)
		}

		// Ordina per priorità
		sort.Slice(analyses, func(i, j int) bool {
			if analyses[i].Alert != analyses[j].Alert {
				return analyses[i].Alert
			}
			return analyses[i].DaysUntilEnd < analyses[j].DaysUntilEnd
		})

		c.JSON(http.StatusOK, analyses)
	}
}

// @Summary Update notification settings
// @Description Update the notification settings for the system
// @Tags Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param settings body NotificationSettings true "Notification settings"
// @Success 200 {object} NotificationSettings
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /notification-settings [put]
func updateNotificationSettingsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var settings NotificationSettings
		if err := c.ShouldBindJSON(&settings); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Trova o crea impostazioni
		var existing NotificationSettings
		if err := db.First(&existing).Error; err != nil {
			settings.ID = 0 // Nuovo record
		} else {
			settings.ID = existing.ID // Aggiorna esistente
		}

		if err := db.Save(&settings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, settings)
	}
}

// @Summary Get notification settings
// @Description Get the current notification settings for the system
// @Tags Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} NotificationSettings
// @Failure 404 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /notification-settings [get]
func getNotificationSettingsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var settings NotificationSettings
		if err := db.First(&settings).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Settings not found"})
			return
		}

		c.JSON(http.StatusOK, settings)
	}
}

// Background job per controlli periodici
func StartNotificationScheduler(db *gorm.DB) {
	ns := NewNotificationService(db)

	// Controlli ogni 30 minuti
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				ns.CheckAllNotifications()
			}
		}
	}()

	// Controllo iniziale
	go func() {
		time.Sleep(10 * time.Second) // Aspetta che il server sia pronto
		ns.CheckAllNotifications()
	}()

	log.Println("📅 Notification scheduler started")
}

// Aggiungi route al router principale
func addNotificationRoutes(r *gin.Engine, db *gorm.DB) {
	api := r.Group("/api")

	protected := api.Group("/")
	protected.Use(authMiddleware())
	{
		// Notifiche base
		protected.GET("/notifications", getNotificationsHandler(db))
		protected.PUT("/notifications/:id/read", markNotificationReadHandler(db))
		protected.DELETE("/notifications/:id", dismissNotificationHandler(db))
		protected.GET("/notifications/stats", getNotificationStatsHandler(db))

		// Analisi avanzate
		protected.GET("/analysis/weight", getWeightAnalysisHandler(db))
		protected.GET("/analysis/therapy", getTherapyAnalysisHandler(db))

		// Impostazioni
		protected.GET("/notification-settings", getNotificationSettingsHandler(db))
		protected.PUT("/notification-settings", updateNotificationSettingsHandler(db))

		// Trigger manuale controlli
		// @Summary Trigger notification check
		// @Description Manually trigger a check of all notifications
		// @Tags Notifications
		// @Accept json
		// @Produce json
		// @Security BearerAuth
		// @Success 200 {object} map[string]string
		// @Failure 401 {object} map[string]string
		// @Router /notifications/check [post]
		protected.POST("/notifications/check", func(c *gin.Context) {
			ns := NewNotificationService(db)
			go ns.CheckAllNotifications()
			c.JSON(http.StatusOK, gin.H{"message": "Notification check triggered"})
		})
	}
}
