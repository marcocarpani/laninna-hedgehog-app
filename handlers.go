package main

import (
	"github.com/gin-gonic/gin"
	"github.com/laninna/hedgehog-app/logger"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

// @Summary Get all hedgehogs
// @Description Get list of all hedgehogs with their areas and relationships
// @Tags Hedgehogs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Hedgehog
// @Failure 401 {object} map[string]string
// @Router /hedgehogs [get]
func getHedgehogs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var hedgehogs []Hedgehog
		db.Preload("Area").Preload("Area.Room").Preload("Therapies").Preload("WeightRecords").Find(&hedgehogs)
		c.JSON(http.StatusOK, hedgehogs)
	}
}

// @Summary Create new hedgehog
// @Description Create a new hedgehog record
// @Tags Hedgehogs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param hedgehog body Hedgehog true "Hedgehog data"
// @Success 201 {object} Hedgehog
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /hedgehogs [post]
func createHedgehog(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request-scoped logger from context
		log := logger.GetLoggerFromContext(c)
		
		var hedgehog Hedgehog
		if err := c.ShouldBindJSON(&hedgehog); err != nil {
			log.Warn().Err(err).Msg("Invalid hedgehog data received")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if hedgehog.ArrivalDate.IsZero() {
			hedgehog.ArrivalDate = time.Now()
		}
		
		// Validate that if status is 'released', ReleaseDate is set
		if hedgehog.Status == "released" && hedgehog.ReleaseDate == nil {
			log.Warn().
				Str("status", hedgehog.Status).
				Str("name", hedgehog.Name).
				Msg("Validation error: Release date is required when status is 'released'")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Release date is required when status is 'released'"})
			return
		}

		if err := db.Create(&hedgehog).Error; err != nil {
			log.Error().Err(err).
				Str("name", hedgehog.Name).
				Str("status", hedgehog.Status).
				Msg("Failed to create hedgehog")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		db.Preload("Area").Preload("Area.Room").First(&hedgehog, hedgehog.ID)
		
		log.Info().
			Uint("id", hedgehog.ID).
			Str("name", hedgehog.Name).
			Str("status", hedgehog.Status).
			Time("arrival_date", hedgehog.ArrivalDate).
			Msg("Hedgehog created successfully")
			
		c.JSON(http.StatusCreated, hedgehog)
	}
}

// @Summary Get hedgehog by ID
// @Description Get a single hedgehog by its ID with related data
// @Tags Hedgehogs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Hedgehog ID"
// @Success 200 {object} Hedgehog
// @Failure 404 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /hedgehogs/{id} [get]
func getHedgehog(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var hedgehog Hedgehog

		if err := db.Preload("Area").Preload("Area.Room").Preload("Therapies").Preload("WeightRecords").First(&hedgehog, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Hedgehog not found"})
			return
		}

		c.JSON(http.StatusOK, hedgehog)
	}
}

// @Summary Update hedgehog
// @Description Update an existing hedgehog's information
// @Tags Hedgehogs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Hedgehog ID"
// @Param hedgehog body Hedgehog true "Updated hedgehog data"
// @Success 200 {object} Hedgehog
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /hedgehogs/{id} [put]
func updateHedgehog(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var hedgehog Hedgehog

		if err := db.First(&hedgehog, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Hedgehog not found"})
			return
		}

		if err := c.ShouldBindJSON(&hedgehog); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Validate that if status is 'released', ReleaseDate is set
		if hedgehog.Status == "released" && hedgehog.ReleaseDate == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Release date is required when status is 'released'"})
			return
		}

		if err := db.Save(&hedgehog).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		db.Preload("Area").Preload("Area.Room").First(&hedgehog, hedgehog.ID)
		c.JSON(http.StatusOK, hedgehog)
	}
}

// @Summary Delete hedgehog
// @Description Delete a hedgehog by its ID
// @Tags Hedgehogs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Hedgehog ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /hedgehogs/{id} [delete]
func deleteHedgehog(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := db.Delete(&Hedgehog{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Hedgehog deleted"})
	}
}

// Room handlers
// @Summary Get all rooms
// @Description Get list of all rooms with their areas and hedgehogs
// @Tags Rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Room
// @Failure 401 {object} map[string]string
// @Router /rooms [get]
func getRooms(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rooms []Room
		db.Preload("Areas").Preload("Areas.Hedgehogs").Find(&rooms)
		c.JSON(http.StatusOK, rooms)
	}
}

// @Summary Create new room
// @Description Create a new room record
// @Tags Rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param room body Room true "Room data"
// @Success 201 {object} Room
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /rooms [post]
func createRoom(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var room Room
		if err := c.ShouldBindJSON(&room); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Create(&room).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, room)
	}
}

// @Summary Get room by ID
// @Description Get a single room by its ID with related areas and hedgehogs
// @Tags Rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Room ID"
// @Success 200 {object} Room
// @Failure 404 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /rooms/{id} [get]
func getRoom(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var room Room

		if err := db.Preload("Areas").Preload("Areas.Hedgehogs").First(&room, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
			return
		}

		c.JSON(http.StatusOK, room)
	}
}

// @Summary Update room
// @Description Update an existing room's information
// @Tags Rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Room ID"
// @Param room body Room true "Updated room data"
// @Success 200 {object} Room
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /rooms/{id} [put]
func updateRoom(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var room Room

		if err := db.First(&room, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
			return
		}

		if err := c.ShouldBindJSON(&room); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Save(&room).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, room)
	}
}

// @Summary Delete room
// @Description Delete a room by its ID
// @Tags Rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Room ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /rooms/{id} [delete]
func deleteRoom(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := db.Delete(&Room{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Room deleted"})
	}
}

// Area handlers
// @Summary Get all areas
// @Description Get list of all areas with their rooms and hedgehogs
// @Tags Areas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Area
// @Failure 401 {object} map[string]string
// @Router /areas [get]
func getAreas(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var areas []Area
		db.Preload("Room").Preload("Hedgehogs").Find(&areas)
		c.JSON(http.StatusOK, areas)
	}
}

// @Summary Create new area
// @Description Create a new area record
// @Tags Areas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param area body Area true "Area data"
// @Success 201 {object} Area
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /areas [post]
func createArea(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var area Area
		if err := c.ShouldBindJSON(&area); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Create(&area).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, area)
	}
}

// @Summary Update area
// @Description Update an existing area's information
// @Tags Areas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Area ID"
// @Param area body Area true "Updated area data"
// @Success 200 {object} Area
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /areas/{id} [put]
func updateArea(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var area Area

		if err := db.First(&area, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Area not found"})
			return
		}

		if err := c.ShouldBindJSON(&area); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Save(&area).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, area)
	}
}

// @Summary Delete area
// @Description Delete an area by its ID
// @Tags Areas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Area ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /areas/{id} [delete]
func deleteArea(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := db.Delete(&Area{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Area deleted"})
	}
}

// Therapy handlers
// @Summary Get all therapies
// @Description Get list of all therapies with optional hedgehog_id filter
// @Tags Therapies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param hedgehog_id query int false "Filter by hedgehog ID"
// @Success 200 {array} Therapy
// @Failure 401 {object} map[string]string
// @Router /therapies [get]
func getTherapies(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var therapies []Therapy
		hedgehogID := c.Query("hedgehog_id")

		query := db
		if hedgehogID != "" {
			query = query.Where("hedgehog_id = ?", hedgehogID)
		}

		query.Find(&therapies)
		c.JSON(http.StatusOK, therapies)
	}
}

// @Summary Create new therapy
// @Description Create a new therapy record for a hedgehog
// @Tags Therapies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param therapy body Therapy true "Therapy data"
// @Success 201 {object} Therapy
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /therapies [post]
func createTherapy(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var therapy Therapy
		if err := c.ShouldBindJSON(&therapy); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if therapy.HedgehogID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "hedgehog_id is required"})
			return
		}

		if therapy.StartDate.IsZero() {
			therapy.StartDate = time.Now()
		}

		if err := db.Create(&therapy).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, therapy)
	}
}

// @Summary Update therapy
// @Description Update an existing therapy's information
// @Tags Therapies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Therapy ID"
// @Param therapy body Therapy true "Updated therapy data"
// @Success 200 {object} Therapy
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /therapies/{id} [put]
func updateTherapy(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var therapy Therapy

		if err := db.First(&therapy, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Therapy not found"})
			return
		}

		if err := c.ShouldBindJSON(&therapy); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Save(&therapy).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, therapy)
	}
}

// @Summary Delete therapy
// @Description Delete a therapy by its ID
// @Tags Therapies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Therapy ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /therapies/{id} [delete]
func deleteTherapy(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := db.Delete(&Therapy{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Therapy deleted"})
	}
}

// @Summary Get weight records
// @Description Get weight records with optional hedgehog_id filter
// @Tags Weight Records
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param hedgehog_id query int false "Filter by hedgehog ID"
// @Param limit query int false "Limit results" default(100)
// @Success 200 {array} WeightRecord
// @Failure 401 {object} map[string]string
// @Router /weight-records [get]
func getWeightRecords(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var records []WeightRecord
		hedgehogID := c.Query("hedgehog_id")

		query := db.Order("date DESC")
		if hedgehogID != "" {
			query = query.Where("hedgehog_id = ?", hedgehogID)
		}

		// Limit per performance
		limit := 100
		if l := c.Query("limit"); l != "" {
			if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}
		query = query.Limit(limit)

		query.Find(&records)
		c.JSON(http.StatusOK, records)
	}
}

// @Summary Create weight record
// @Description Add new weight record for a hedgehog
// @Tags Weight Records
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param weight_record body WeightRecord true "Weight record data"
// @Success 201 {object} WeightRecord
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /weight-records [post]
func createWeightRecord(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var record WeightRecord
		if err := c.ShouldBindJSON(&record); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if record.HedgehogID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "hedgehog_id is required"})
			return
		}

		if record.Date.IsZero() {
			record.Date = time.Now()
		}

		if err := db.Create(&record).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, record)
	}
}

// @Summary Update weight record
// @Description Update an existing weight record's information
// @Tags Weight Records
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Weight Record ID"
// @Param weight_record body WeightRecord true "Updated weight record data"
// @Success 200 {object} WeightRecord
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /weight-records/{id} [put]
func updateWeightRecord(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var record WeightRecord

		if err := db.First(&record, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Weight record not found"})
			return
		}

		if err := c.ShouldBindJSON(&record); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Save(&record).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, record)
	}
}

// @Summary Delete weight record
// @Description Delete a weight record by its ID
// @Tags Weight Records
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Weight Record ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /weight-records/{id} [delete]
func deleteWeightRecord(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := db.Delete(&WeightRecord{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Weight record deleted"})
	}
}

// Handler per pagina notifiche
func notificationsPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "notifications.html", gin.H{
		"title": "Centro Notifiche - La Ninna",
	})
}

// Handler per pagina export
// @Summary Quick export data
// @Description Export data in various formats (PDF, Excel, CSV)
// @Tags Export
// @Accept json
// @Produce application/pdf,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,text/csv
// @Security BearerAuth
// @Param start_date query string false "Start date filter (YYYY-MM-DD)"
// @Param end_date query string false "End date filter (YYYY-MM-DD)"
// @Param status query string false "Status filter"
// @Param room_id query int false "Room ID filter"
// @Success 200 {file} file "Exported file"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /export/{dataType}/{format} [get]
func quickExportHandler(db *gorm.DB, dataType, format string) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := ExportRequest{
			Type:   dataType,
			Format: format,
		}

		// Parse query parameters per filtri
		if startDate := c.Query("start_date"); startDate != "" {
			if date, err := time.Parse("2006-01-02", startDate); err == nil {
				req.StartDate = &date
			}
		}

		if endDate := c.Query("end_date"); endDate != "" {
			if date, err := time.Parse("2006-01-02", endDate); err == nil {
				req.EndDate = &date
			}
		}

		if status := c.Query("status"); status != "" {
			req.Status = status
		}

		if roomID := c.Query("room_id"); roomID != "" {
			if id, err := strconv.ParseUint(roomID, 10, 32); err == nil {
				roomIDUint := uint(id)
				req.RoomID = &roomIDUint
			}
		}

		switch format {
		case "pdf":
			handlePDFExport(c, db, req)
		case "excel":
			handleExcelExport(c, db, req)
		case "csv":
			handleCSVExport(c, db, req)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato non supportato"})
		}
	}
}
