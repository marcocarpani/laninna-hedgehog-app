// export.go - Handlers per l'export di dati
package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type ExportRequest struct {
	Type      string     `json:"type" binding:"required"`   // hedgehogs, rooms, therapies, weights
	Format    string     `json:"format" binding:"required"` // pdf, excel, csv
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	Status    string     `json:"status"`
	RoomID    *uint      `json:"room_id"`
}

// Handler principale per l'export
func exportDataHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ExportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		switch req.Format {
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

// Export PDF
func handlePDFExport(c *gin.Context, db *gorm.DB, req ExportRequest) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8Font("DejaVu", "", "./fonts/DejaVuSans.ttf")
	pdf.AddUTF8Font("DejaVu", "B", "./fonts/DejaVuSans-Bold.ttf")

	switch req.Type {
	case "hedgehogs":
		generateHedgehogsPDF(pdf, db, req)
	case "rooms":
		generateRoomsPDF(pdf, db, req)
	case "therapies":
		generateTherapiesPDF(pdf, db, req)
	case "weights":
		generateWeightsPDF(pdf, db, req)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo di report non supportato"})
		return
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Errore nella generazione PDF"})
		return
	}

	filename := fmt.Sprintf("la-ninna-%s-%s.pdf", req.Type, time.Now().Format("2006-01-02"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/pdf")
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

func generateHedgehogsPDF(pdf *gofpdf.Fpdf, db *gorm.DB, req ExportRequest) {
	// Header del documento
	pdf.AddPage()
	addPDFHeader(pdf, "Report Ricci in Cura")

	// Query ricci con filtri
	var hedgehogs []Hedgehog
	query := db.Preload("Area").Preload("Area.Room").Preload("Therapies").Preload("WeightRecords")

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.StartDate != nil {
		query = query.Where("arrival_date >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("arrival_date <= ?", req.EndDate)
	}

	query.Find(&hedgehogs)

	// Statistiche generali
	pdf.SetFont("DejaVu", "", 10)
	pdf.Ln(5)
	pdf.Cell(40, 6, fmt.Sprintf("Totale ricci: %d", len(hedgehogs)))
	pdf.Ln(5)

	inCare := 0
	recovered := 0
	deceased := 0
	for _, h := range hedgehogs {
		switch h.Status {
		case "in_care":
			inCare++
		case "recovered":
			recovered++
		case "deceased":
			deceased++
		}
	}

	pdf.Cell(40, 6, fmt.Sprintf("In cura: %d", inCare))
	pdf.Ln(5)
	pdf.Cell(40, 6, fmt.Sprintf("Recuperati: %d", recovered))
	pdf.Ln(5)
	pdf.Cell(40, 6, fmt.Sprintf("Deceduti: %d", deceased))
	pdf.Ln(10)

	// Tabella ricci
	addPDFTableHeader(pdf, []string{"Nome", "Stato", "Arrivo", "Stanza", "Peso Attuale"})

	for _, hedgehog := range hedgehogs {
		status := map[string]string{
			"in_care":   "In cura",
			"recovered": "Recuperato",
			"deceased":  "Deceduto",
		}[hedgehog.Status]

		room := "Non assegnato"
		if hedgehog.Area != nil && hedgehog.Area.Room.Name != "" {
			room = fmt.Sprintf("%s - %s", hedgehog.Area.Room.Name, hedgehog.Area.Name)
		}

		currentWeight := "N/D"
		if len(hedgehog.WeightRecords) > 0 {
			// Ordina per data più recente
			latest := hedgehog.WeightRecords[0]
			for _, w := range hedgehog.WeightRecords {
				if w.Date.After(latest.Date) {
					latest = w
				}
			}
			currentWeight = fmt.Sprintf("%.1fg", latest.Weight)
		}

		addPDFTableRow(pdf, []string{
			hedgehog.Name,
			status,
			hedgehog.ArrivalDate.Format("02/01/2006"),
			room,
			currentWeight,
		})
	}
}

func generateRoomsPDF(pdf *gofpdf.Fpdf, db *gorm.DB, req ExportRequest) {
	pdf.AddPage()
	addPDFHeader(pdf, "Report Stanze e Occupazione")

	var rooms []Room
	query := db.Preload("Areas").Preload("Areas.Hedgehogs")
	if req.RoomID != nil {
		query = query.Where("id = ?", *req.RoomID)
	}
	query.Find(&rooms)

	// Statistiche generali
	pdf.SetFont("DejaVu", "", 10)
	pdf.Ln(5)

	totalRooms := len(rooms)
	totalAreas := 0
	totalCapacity := 0
	totalOccupied := 0

	for _, room := range rooms {
		totalAreas += len(room.Areas)
		for _, area := range room.Areas {
			totalCapacity += area.MaxCapacity
			totalOccupied += len(area.Hedgehogs)
		}
	}

	pdf.Cell(40, 6, fmt.Sprintf("Totale stanze: %d", totalRooms))
	pdf.Ln(5)
	pdf.Cell(40, 6, fmt.Sprintf("Totale aree: %d", totalAreas))
	pdf.Ln(5)
	pdf.Cell(40, 6, fmt.Sprintf("Capacità totale: %d", totalCapacity))
	pdf.Ln(5)
	pdf.Cell(40, 6, fmt.Sprintf("Posti occupati: %d", totalOccupied))
	pdf.Ln(5)
	occupancyRate := float64(totalOccupied) / float64(totalCapacity) * 100
	pdf.Cell(40, 6, fmt.Sprintf("Tasso di occupazione: %.1f%%", occupancyRate))
	pdf.Ln(10)

	// Tabella stanze
	addPDFTableHeader(pdf, []string{"Stanza", "Dimensioni", "Aree", "Occupazione", "Capacità"})

	for _, room := range rooms {
		areaCount := len(room.Areas)
		occupiedCount := 0
		capacityCount := 0

		for _, area := range room.Areas {
			occupiedCount += len(area.Hedgehogs)
			capacityCount += area.MaxCapacity
		}

		addPDFTableRow(pdf, []string{
			room.Name,
			fmt.Sprintf("%.1fm x %.1fm", room.Width, room.Height),
			fmt.Sprintf("%d", areaCount),
			fmt.Sprintf("%d/%d", occupiedCount, capacityCount),
			fmt.Sprintf("%.1f%%", float64(occupiedCount)/float64(capacityCount)*100),
		})
	}
}

func generateTherapiesPDF(pdf *gofpdf.Fpdf, db *gorm.DB, req ExportRequest) {
	pdf.AddPage()
	addPDFHeader(pdf, "Report Terapie")

	var therapies []Therapy
	query := db

	if req.StartDate != nil {
		query = query.Where("start_date >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("start_date <= ?", req.EndDate)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	query.Find(&therapies)

	// Statistiche
	pdf.SetFont("DejaVu", "", 10)
	pdf.Ln(5)

	activeCount := 0
	completedCount := 0
	suspendedCount := 0

	for _, therapy := range therapies {
		switch therapy.Status {
		case "active":
			activeCount++
		case "completed":
			completedCount++
		case "suspended":
			suspendedCount++
		}
	}

	pdf.Cell(40, 6, fmt.Sprintf("Totale terapie: %d", len(therapies)))
	pdf.Ln(5)
	pdf.Cell(40, 6, fmt.Sprintf("Attive: %d", activeCount))
	pdf.Ln(5)
	pdf.Cell(40, 6, fmt.Sprintf("Completate: %d", completedCount))
	pdf.Ln(5)
	pdf.Cell(40, 6, fmt.Sprintf("Sospese: %d", suspendedCount))
	pdf.Ln(10)

	// Tabella terapie
	addPDFTableHeader(pdf, []string{"Riccio", "Terapia", "Inizio", "Stato", "Durata"})

	for _, therapy := range therapies {
		status := map[string]string{
			"active":    "Attiva",
			"completed": "Completata",
			"suspended": "Sospesa",
		}[therapy.Status]

		duration := "In corso"
		if therapy.EndDate != nil {
			days := int(therapy.EndDate.Sub(therapy.StartDate).Hours() / 24)
			duration = fmt.Sprintf("%d giorni", days)
		} else if therapy.Status == "active" {
			days := int(time.Since(therapy.StartDate).Hours() / 24)
			duration = fmt.Sprintf("%d giorni", days)
		}

		// Get hedgehog name by ID
		var hedgehog Hedgehog
		hedgehogName := "N/A"
		if err := db.First(&hedgehog, therapy.HedgehogID).Error; err == nil {
			hedgehogName = hedgehog.Name
		}

		addPDFTableRow(pdf, []string{
			hedgehogName,
			therapy.Name,
			therapy.StartDate.Format("02/01/2006"),
			status,
			duration,
		})
	}
}

func generateWeightsPDF(pdf *gofpdf.Fpdf, db *gorm.DB, req ExportRequest) {
	pdf.AddPage()
	addPDFHeader(pdf, "Report Pesature")

	var records []WeightRecord
	query := db.Order("date DESC")

	if req.StartDate != nil {
		query = query.Where("date >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("date <= ?", req.EndDate)
	}

	query.Limit(1000).Find(&records) // Limita a 1000 record per performance

	// Statistiche
	pdf.SetFont("DejaVu", "", 10)
	pdf.Ln(5)

	if len(records) > 0 {
		totalWeight := 0.0
		for _, record := range records {
			totalWeight += record.Weight
		}
		avgWeight := totalWeight / float64(len(records))

		pdf.Cell(40, 6, fmt.Sprintf("Totale pesature: %d", len(records)))
		pdf.Ln(5)
		pdf.Cell(40, 6, fmt.Sprintf("Peso medio: %.1fg", avgWeight))
		pdf.Ln(10)
	}

	// Tabella pesature
	addPDFTableHeader(pdf, []string{"Riccio", "Data", "Peso", "Variazione", "Note"})

	hedgehogWeights := make(map[uint][]WeightRecord)
	for _, record := range records {
		hedgehogWeights[record.HedgehogID] = append(hedgehogWeights[record.HedgehogID], record)
	}

	for _, record := range records {
		variation := ""
		weights := hedgehogWeights[record.HedgehogID]
		if len(weights) > 1 {
			// Trova la pesatura precedente
			for _, w := range weights {
				if w.Date.Before(record.Date) && w.ID != record.ID {
					diff := record.Weight - w.Weight
					if diff > 0 {
						variation = fmt.Sprintf("+%.1fg", diff)
					} else {
						variation = fmt.Sprintf("%.1fg", diff)
					}
					break
				}
			}
		}

		notes := record.Notes
		if len(notes) > 30 {
			notes = notes[:30] + "..."
		}

		// Get hedgehog name by ID
		var hedgehog Hedgehog
		hedgehogName := "N/A"
		if err := db.First(&hedgehog, record.HedgehogID).Error; err == nil {
			hedgehogName = hedgehog.Name
		}

		addPDFTableRow(pdf, []string{
			hedgehogName,
			record.Date.Format("02/01/2006"),
			fmt.Sprintf("%.1fg", record.Weight),
			variation,
			notes,
		})
	}
}

// Funzioni helper per PDF
func addPDFHeader(pdf *gofpdf.Fpdf, title string) {
	// Logo e intestazione
	pdf.SetFont("DejaVu", "B", 16)
	pdf.Cell(0, 10, "🦔 Centro Recupero Ricci \"La Ninna\"")
	pdf.Ln(8)

	pdf.SetFont("DejaVu", "", 12)
	pdf.Cell(0, 8, "Novello (CN) - Sistema di Gestione")
	pdf.Ln(10)

	pdf.SetFont("DejaVu", "B", 14)
	pdf.Cell(0, 10, title)
	pdf.Ln(8)

	pdf.SetFont("DejaVu", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Generato il: %s", time.Now().Format("02/01/2006 15:04")))
	pdf.Ln(8)

	// Linea separatrice
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(5)
}

func addPDFTableHeader(pdf *gofpdf.Fpdf, headers []string) {
	pdf.SetFont("DejaVu", "B", 9)
	pdf.SetFillColor(240, 240, 240)

	colWidth := 190.0 / float64(len(headers))
	for _, header := range headers {
		pdf.CellFormat(colWidth, 8, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(8)
}

func addPDFTableRow(pdf *gofpdf.Fpdf, data []string) {
	pdf.SetFont("DejaVu", "", 8)
	pdf.SetFillColor(255, 255, 255)

	colWidth := 190.0 / float64(len(data))
	for _, cell := range data {
		if len(cell) > 25 {
			cell = cell[:25] + "..."
		}
		pdf.CellFormat(colWidth, 6, cell, "1", 0, "L", false, 0, "")
	}
	pdf.Ln(6)
}

// Export Excel
func handleExcelExport(c *gin.Context, db *gorm.DB, req ExportRequest) {
	f := excelize.NewFile()

	switch req.Type {
	case "hedgehogs":
		generateHedgehogsExcel(f, db, req)
	case "rooms":
		generateRoomsExcel(f, db, req)
	case "therapies":
		generateTherapiesExcel(f, db, req)
	case "weights":
		generateWeightsExcel(f, db, req)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo di report non supportato"})
		return
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Errore nella generazione Excel"})
		return
	}

	filename := fmt.Sprintf("la-ninna-%s-%s.xlsx", req.Type, time.Now().Format("2006-01-02"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func generateHedgehogsExcel(f *excelize.File, db *gorm.DB, req ExportRequest) {
	sheetName := "Ricci"
	f.NewSheet(sheetName)
	f.DeleteSheet("Sheet1")

	// Header
	headers := []string{"ID", "Nome", "Stato", "Data Arrivo", "Descrizione", "Stanza", "Area", "Terapie Attive", "Ultimo Peso", "Data Ultima Pesata"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// Stile header
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#f0f0f0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", string(rune('A'+len(headers)-1))+"1", style)

	// Query dati
	var hedgehogs []Hedgehog
	query := db.Preload("Area").Preload("Area.Room").Preload("Therapies").Preload("WeightRecords")

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.StartDate != nil {
		query = query.Where("arrival_date >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("arrival_date <= ?", req.EndDate)
	}

	query.Find(&hedgehogs)

	// Popola dati
	for i, hedgehog := range hedgehogs {
		row := i + 2

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), hedgehog.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), hedgehog.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), map[string]string{
			"in_care":   "In cura",
			"recovered": "Recuperato",
			"deceased":  "Deceduto",
		}[hedgehog.Status])
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), hedgehog.ArrivalDate.Format("02/01/2006"))
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), hedgehog.Description)

		room := ""
		area := ""
		if hedgehog.Area != nil {
			if hedgehog.Area.Room.Name != "" {
				room = hedgehog.Area.Room.Name
			}
			area = hedgehog.Area.Name
		}
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), room)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), area)

		activeTherapies := 0
		for _, therapy := range hedgehog.Therapies {
			if therapy.Status == "active" {
				activeTherapies++
			}
		}
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), activeTherapies)

		if len(hedgehog.WeightRecords) > 0 {
			latest := hedgehog.WeightRecords[0]
			for _, w := range hedgehog.WeightRecords {
				if w.Date.After(latest.Date) {
					latest = w
				}
			}
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), latest.Weight)
			f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), latest.Date.Format("02/01/2006"))
		}
	}

	// Auto-adjust column width
	for i := 0; i < len(headers); i++ {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 15)
	}
}

func generateRoomsExcel(f *excelize.File, db *gorm.DB, req ExportRequest) {
	// Sheet principale stanze
	roomSheet := "Stanze"
	f.NewSheet(roomSheet)
	f.DeleteSheet("Sheet1")

	roomHeaders := []string{"ID", "Nome", "Descrizione", "Larghezza (m)", "Altezza (m)", "Numero Aree", "Capacità Totale", "Posti Occupati", "Tasso Occupazione"}
	for i, header := range roomHeaders {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(roomSheet, cell, header)
	}

	// Sheet dettaglio aree
	areaSheet := "Aree"
	f.NewSheet(areaSheet)

	areaHeaders := []string{"ID", "Nome", "Stanza", "Posizione X", "Posizione Y", "Larghezza", "Altezza", "Capacità Max", "Ricci Ospitati", "Nomi Ricci"}
	for i, header := range areaHeaders {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(areaSheet, cell, header)
	}

	var rooms []Room
	query := db.Preload("Areas").Preload("Areas.Hedgehogs")
	if req.RoomID != nil {
		query = query.Where("id = ?", *req.RoomID)
	}
	query.Find(&rooms)

	// Popola sheet stanze
	for i, room := range rooms {
		row := i + 2

		totalCapacity := 0
		totalOccupied := 0
		for _, area := range room.Areas {
			totalCapacity += area.MaxCapacity
			totalOccupied += len(area.Hedgehogs)
		}

		occupancyRate := 0.0
		if totalCapacity > 0 {
			occupancyRate = float64(totalOccupied) / float64(totalCapacity) * 100
		}

		f.SetCellValue(roomSheet, fmt.Sprintf("A%d", row), room.ID)
		f.SetCellValue(roomSheet, fmt.Sprintf("B%d", row), room.Name)
		f.SetCellValue(roomSheet, fmt.Sprintf("C%d", row), room.Description)
		f.SetCellValue(roomSheet, fmt.Sprintf("D%d", row), room.Width)
		f.SetCellValue(roomSheet, fmt.Sprintf("E%d", row), room.Height)
		f.SetCellValue(roomSheet, fmt.Sprintf("F%d", row), len(room.Areas))
		f.SetCellValue(roomSheet, fmt.Sprintf("G%d", row), totalCapacity)
		f.SetCellValue(roomSheet, fmt.Sprintf("H%d", row), totalOccupied)
		f.SetCellValue(roomSheet, fmt.Sprintf("I%d", row), fmt.Sprintf("%.1f%%", occupancyRate))
	}

	// Popola sheet aree
	areaRow := 2
	for _, room := range rooms {
		for _, area := range room.Areas {
			hedgehogNames := make([]string, len(area.Hedgehogs))
			for i, hedgehog := range area.Hedgehogs {
				hedgehogNames[i] = hedgehog.Name
			}

			f.SetCellValue(areaSheet, fmt.Sprintf("A%d", areaRow), area.ID)
			f.SetCellValue(areaSheet, fmt.Sprintf("B%d", areaRow), area.Name)
			f.SetCellValue(areaSheet, fmt.Sprintf("C%d", areaRow), room.Name)
			f.SetCellValue(areaSheet, fmt.Sprintf("D%d", areaRow), area.X)
			f.SetCellValue(areaSheet, fmt.Sprintf("E%d", areaRow), area.Y)
			f.SetCellValue(areaSheet, fmt.Sprintf("F%d", areaRow), area.Width)
			f.SetCellValue(areaSheet, fmt.Sprintf("G%d", areaRow), area.Height)
			f.SetCellValue(areaSheet, fmt.Sprintf("H%d", areaRow), area.MaxCapacity)
			f.SetCellValue(areaSheet, fmt.Sprintf("I%d", areaRow), len(area.Hedgehogs))
			f.SetCellValue(areaSheet, fmt.Sprintf("J%d", areaRow), fmt.Sprintf("%v", hedgehogNames))

			areaRow++
		}
	}
}

func generateTherapiesExcel(f *excelize.File, db *gorm.DB, req ExportRequest) {
	sheetName := "Terapie"
	f.NewSheet(sheetName)
	f.DeleteSheet("Sheet1")

	headers := []string{"ID", "Riccio", "Nome Terapia", "Descrizione", "Data Inizio", "Data Fine", "Stato", "Durata (giorni)"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	var therapies []Therapy
	query := db

	if req.StartDate != nil {
		query = query.Where("start_date >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("start_date <= ?", req.EndDate)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	query.Find(&therapies)

	for i, therapy := range therapies {
		row := i + 2

		duration := ""
		if therapy.EndDate != nil {
			days := int(therapy.EndDate.Sub(therapy.StartDate).Hours() / 24)
			duration = fmt.Sprintf("%d", days)
		} else if therapy.Status == "active" {
			days := int(time.Since(therapy.StartDate).Hours() / 24)
			duration = fmt.Sprintf("%d (in corso)", days)
		}

		endDate := ""
		if therapy.EndDate != nil {
			endDate = therapy.EndDate.Format("02/01/2006")
		}

		// Get hedgehog name by ID
		var hedgehog Hedgehog
		hedgehogName := "N/A"
		if err := db.First(&hedgehog, therapy.HedgehogID).Error; err == nil {
			hedgehogName = hedgehog.Name
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), therapy.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), hedgehogName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), therapy.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), therapy.Description)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), therapy.StartDate.Format("02/01/2006"))
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), endDate)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), map[string]string{
			"active":    "Attiva",
			"completed": "Completata",
			"suspended": "Sospesa",
		}[therapy.Status])
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), duration)
	}
}

func generateWeightsExcel(f *excelize.File, db *gorm.DB, req ExportRequest) {
	sheetName := "Pesature"
	f.NewSheet(sheetName)
	f.DeleteSheet("Sheet1")

	headers := []string{"ID", "Riccio", "Data", "Peso (g)", "Variazione (g)", "Note", "Trend"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	var records []WeightRecord
	query := db.Order("hedgehog_id, date")

	if req.StartDate != nil {
		query = query.Where("date >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("date <= ?", req.EndDate)
	}

	query.Find(&records)

	// Raggruppa per riccio per calcolare variazioni
	hedgehogRecords := make(map[uint][]WeightRecord)
	for _, record := range records {
		hedgehogRecords[record.HedgehogID] = append(hedgehogRecords[record.HedgehogID], record)
	}

	row := 2
	for _, record := range records {
		variation := ""
		trend := ""

		// Trova record precedente dello stesso riccio
		hedgehogWeights := hedgehogRecords[record.HedgehogID]
		for i, w := range hedgehogWeights {
			if w.ID == record.ID && i > 0 {
				prevWeight := hedgehogWeights[i-1].Weight
				diff := record.Weight - prevWeight
				variation = fmt.Sprintf("%.1f", diff)

				if diff > 0 {
					trend = "↗"
				} else if diff < 0 {
					trend = "↘"
				} else {
					trend = "→"
				}
				break
			}
		}

		// Get hedgehog name by ID
		var hedgehog Hedgehog
		hedgehogName := "N/A"
		if err := db.First(&hedgehog, record.HedgehogID).Error; err == nil {
			hedgehogName = hedgehog.Name
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), record.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), hedgehogName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), record.Date.Format("02/01/2006"))
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), record.Weight)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), variation)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), record.Notes)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), trend)

		row++
	}

	// Crea un secondo sheet con statistiche per riccio
	statsSheet := "Statistiche Peso"
	f.NewSheet(statsSheet)

	statsHeaders := []string{"Riccio", "Primo Peso (g)", "Ultimo Peso (g)", "Variazione Totale (g)", "Peso Medio (g)", "Numero Pesature", "Periodo (giorni)"}
	for i, header := range statsHeaders {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(statsSheet, cell, header)
	}

	statsRow := 2
	for _, weights := range hedgehogRecords {
		if len(weights) == 0 {
			continue
		}

		firstWeight := weights[0]
		lastWeight := weights[len(weights)-1]

		totalWeight := 0.0
		for _, w := range weights {
			totalWeight += w.Weight
		}
		avgWeight := totalWeight / float64(len(weights))

		totalVariation := lastWeight.Weight - firstWeight.Weight
		days := int(lastWeight.Date.Sub(firstWeight.Date).Hours() / 24)

		// Get hedgehog name by ID
		var hedgehog Hedgehog
		hedgehogName := "N/A"
		if err := db.First(&hedgehog, firstWeight.HedgehogID).Error; err == nil {
			hedgehogName = hedgehog.Name
		}

		f.SetCellValue(statsSheet, fmt.Sprintf("A%d", statsRow), hedgehogName)
		f.SetCellValue(statsSheet, fmt.Sprintf("B%d", statsRow), firstWeight.Weight)
		f.SetCellValue(statsSheet, fmt.Sprintf("C%d", statsRow), lastWeight.Weight)
		f.SetCellValue(statsSheet, fmt.Sprintf("D%d", statsRow), fmt.Sprintf("%.1f", totalVariation))
		f.SetCellValue(statsSheet, fmt.Sprintf("E%d", statsRow), fmt.Sprintf("%.1f", avgWeight))
		f.SetCellValue(statsSheet, fmt.Sprintf("F%d", statsRow), len(weights))
		f.SetCellValue(statsSheet, fmt.Sprintf("G%d", statsRow), days)

		statsRow++
	}
}

// Export CSV
func handleCSVExport(c *gin.Context, db *gorm.DB, req ExportRequest) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Comma = ';' // Usa punto e virgola per compatibilità italiana

	switch req.Type {
	case "hedgehogs":
		generateHedgehogsCSV(writer, db, req)
	case "rooms":
		generateRoomsCSV(writer, db, req)
	case "therapies":
		generateTherapiesCSV(writer, db, req)
	case "weights":
		generateWeightsCSV(writer, db, req)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo di report non supportato"})
		return
	}

	writer.Flush()

	filename := fmt.Sprintf("la-ninna-%s-%s.csv", req.Type, time.Now().Format("2006-01-02"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "text/csv; charset=utf-8")

	// Aggiungi BOM per Excel
	data := append([]byte{0xEF, 0xBB, 0xBF}, buf.Bytes()...)
	c.Data(http.StatusOK, "text/csv", data)
}

func generateHedgehogsCSV(writer *csv.Writer, db *gorm.DB, req ExportRequest) {
	// Header CSV
	writer.Write([]string{"ID", "Nome", "Stato", "Data Arrivo", "Descrizione", "Stanza", "Area", "Terapie Attive", "Ultimo Peso", "Data Ultima Pesata"})

	var hedgehogs []Hedgehog
	query := db.Preload("Area").Preload("Area.Room").Preload("Therapies").Preload("WeightRecords")

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.StartDate != nil {
		query = query.Where("arrival_date >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("arrival_date <= ?", req.EndDate)
	}

	query.Find(&hedgehogs)

	for _, hedgehog := range hedgehogs {
		status := map[string]string{
			"in_care":   "In cura",
			"recovered": "Recuperato",
			"deceased":  "Deceduto",
		}[hedgehog.Status]

		room := ""
		area := ""
		if hedgehog.Area != nil {
			if hedgehog.Area.Room.Name != "" {
				room = hedgehog.Area.Room.Name
			}
			area = hedgehog.Area.Name
		}

		activeTherapies := 0
		for _, therapy := range hedgehog.Therapies {
			if therapy.Status == "active" {
				activeTherapies++
			}
		}

		lastWeight := ""
		lastWeightDate := ""
		if len(hedgehog.WeightRecords) > 0 {
			latest := hedgehog.WeightRecords[0]
			for _, w := range hedgehog.WeightRecords {
				if w.Date.After(latest.Date) {
					latest = w
				}
			}
			lastWeight = fmt.Sprintf("%.1f", latest.Weight)
			lastWeightDate = latest.Date.Format("02/01/2006")
		}

		writer.Write([]string{
			fmt.Sprintf("%d", hedgehog.ID),
			hedgehog.Name,
			status,
			hedgehog.ArrivalDate.Format("02/01/2006"),
			hedgehog.Description,
			room,
			area,
			fmt.Sprintf("%d", activeTherapies),
			lastWeight,
			lastWeightDate,
		})
	}
}

func generateRoomsCSV(writer *csv.Writer, db *gorm.DB, req ExportRequest) {
	writer.Write([]string{"ID", "Nome", "Descrizione", "Larghezza", "Altezza", "Numero Aree", "Capacità Totale", "Posti Occupati", "Tasso Occupazione"})

	var rooms []Room
	query := db.Preload("Areas").Preload("Areas.Hedgehogs")
	if req.RoomID != nil {
		query = query.Where("id = ?", *req.RoomID)
	}
	query.Find(&rooms)

	for _, room := range rooms {
		totalCapacity := 0
		totalOccupied := 0
		for _, area := range room.Areas {
			totalCapacity += area.MaxCapacity
			totalOccupied += len(area.Hedgehogs)
		}

		occupancyRate := 0.0
		if totalCapacity > 0 {
			occupancyRate = float64(totalOccupied) / float64(totalCapacity) * 100
		}

		writer.Write([]string{
			fmt.Sprintf("%d", room.ID),
			room.Name,
			room.Description,
			fmt.Sprintf("%.1f", room.Width),
			fmt.Sprintf("%.1f", room.Height),
			fmt.Sprintf("%d", len(room.Areas)),
			fmt.Sprintf("%d", totalCapacity),
			fmt.Sprintf("%d", totalOccupied),
			fmt.Sprintf("%.1f%%", occupancyRate),
		})
	}
}

func generateTherapiesCSV(writer *csv.Writer, db *gorm.DB, req ExportRequest) {
	writer.Write([]string{"ID", "Riccio", "Nome Terapia", "Descrizione", "Data Inizio", "Data Fine", "Stato", "Durata Giorni"})

	var therapies []Therapy
	query := db

	if req.StartDate != nil {
		query = query.Where("start_date >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("start_date <= ?", req.EndDate)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	query.Find(&therapies)

	for _, therapy := range therapies {
		status := map[string]string{
			"active":    "Attiva",
			"completed": "Completata",
			"suspended": "Sospesa",
		}[therapy.Status]

		duration := ""
		endDate := ""
		if therapy.EndDate != nil {
			days := int(therapy.EndDate.Sub(therapy.StartDate).Hours() / 24)
			duration = fmt.Sprintf("%d", days)
			endDate = therapy.EndDate.Format("02/01/2006")
		} else if therapy.Status == "active" {
			days := int(time.Since(therapy.StartDate).Hours() / 24)
			duration = fmt.Sprintf("%d", days)
		}

		// Get hedgehog name by ID
		var hedgehog Hedgehog
		hedgehogName := "N/A"
		if err := db.First(&hedgehog, therapy.HedgehogID).Error; err == nil {
			hedgehogName = hedgehog.Name
		}

		writer.Write([]string{
			fmt.Sprintf("%d", therapy.ID),
			hedgehogName,
			therapy.Name,
			therapy.Description,
			therapy.StartDate.Format("02/01/2006"),
			endDate,
			status,
			duration,
		})
	}
}

func generateWeightsCSV(writer *csv.Writer, db *gorm.DB, req ExportRequest) {
	writer.Write([]string{"ID", "Riccio", "Data", "Peso", "Variazione", "Note"})

	var records []WeightRecord
	query := db.Order("hedgehog_id, date")

	if req.StartDate != nil {
		query = query.Where("date >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("date <= ?", req.EndDate)
	}

	query.Find(&records)

	// Raggruppa per riccio per calcolare variazioni
	hedgehogRecords := make(map[uint][]WeightRecord)
	for _, record := range records {
		hedgehogRecords[record.HedgehogID] = append(hedgehogRecords[record.HedgehogID], record)
	}

	for _, record := range records {
		variation := ""

		// Trova record precedente dello stesso riccio
		hedgehogWeights := hedgehogRecords[record.HedgehogID]
		for i, w := range hedgehogWeights {
			if w.ID == record.ID && i > 0 {
				prevWeight := hedgehogWeights[i-1].Weight
				diff := record.Weight - prevWeight
				variation = fmt.Sprintf("%.1f", diff)
				break
			}
		}

		// Get hedgehog name by ID
		var hedgehog Hedgehog
		hedgehogName := "N/A"
		if err := db.First(&hedgehog, record.HedgehogID).Error; err == nil {
			hedgehogName = hedgehog.Name
		}

		writer.Write([]string{
			fmt.Sprintf("%d", record.ID),
			hedgehogName,
			record.Date.Format("02/01/2006"),
			fmt.Sprintf("%.1f", record.Weight),
			variation,
			record.Notes,
		})
	}
}
