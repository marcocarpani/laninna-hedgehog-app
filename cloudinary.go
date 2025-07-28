package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	"github.com/laninna/hedgehog-app/logger"
	"gorm.io/gorm"
)

// CloudinaryService handles interactions with the Cloudinary API
type CloudinaryService struct {
	cld *cloudinary.Cloudinary
}

// NewCloudinaryService creates a new instance of CloudinaryService
func NewCloudinaryService() (*CloudinaryService, error) {
	// Get Cloudinary credentials from environment variables
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	// Validate that all required credentials are provided
	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, errors.New("missing Cloudinary credentials: CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY, and CLOUDINARY_API_SECRET must be set")
	}

	// Create a new Cloudinary instance
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, err
	}

	return &CloudinaryService{cld: cld}, nil
}

// boolPtr returns a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}

// UploadImage uploads an image to Cloudinary
func (s *CloudinaryService) UploadImage(ctx context.Context, file io.Reader, filename string) (string, error) {
	// Create upload context with timeout
	uploadCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Set upload parameters
	uploadParams := uploader.UploadParams{
		Folder:         "hedgehogs",
		PublicID:       filename,
		UseFilename:    boolPtr(true),
		UniqueFilename: boolPtr(true),
		Overwrite:      boolPtr(true),
	}

	// Upload the image
	result, err := s.cld.Upload.Upload(uploadCtx, file, uploadParams)
	if err != nil {
		return "", err
	}

	// Debug logging
	fmt.Printf("Cloudinary upload result - SecureURL: %s, URL: %s, PublicID: %s\n", 
		result.SecureURL, result.URL, result.PublicID)

	// Check if SecureURL is empty and use URL as fallback
	if result.SecureURL == "" {
		fmt.Printf("Warning: SecureURL is empty, using URL instead: %s\n", result.URL)
		return result.URL, nil
	}

	// Return the secure URL of the uploaded image
	return result.SecureURL, nil
}

// UploadHedgehogImageHandler handles the upload of a hedgehog image
// @Summary Upload hedgehog image
// @Description Upload an image for a hedgehog to Cloudinary
// @Tags Hedgehogs
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "Hedgehog ID"
// @Param file formData file true "Image file to upload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /hedgehogs/{id}/image [post]
func UploadHedgehogImageHandler(db *gorm.DB, cloudinaryService *CloudinaryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request-scoped logger from context
		log := logger.GetLoggerFromContext(c)

		// Get hedgehog ID from URL parameter
		hedgehogID := c.Param("id")

		// Find the hedgehog in the database
		var hedgehog Hedgehog
		if err := db.First(&hedgehog, hedgehogID).Error; err != nil {
			log.Warn().Err(err).Str("hedgehog_id", hedgehogID).Msg("Hedgehog not found")
			c.JSON(404, gin.H{"error": "Hedgehog not found"})
			return
		}

		// Get the file from the form
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			log.Warn().Err(err).Msg("Failed to get file from form")
			c.JSON(400, gin.H{"error": "No file provided or invalid file"})
			return
		}
		defer file.Close()

		// Validate file type
		if !isValidImageType(header) {
			log.Warn().Str("filename", header.Filename).Msg("Invalid file type")
			c.JSON(400, gin.H{"error": "Invalid file type. Only JPEG, PNG, and GIF are allowed"})
			return
		}

		// Generate a unique filename based on hedgehog ID and timestamp
		filename := "hedgehog_" + hedgehogID + "_" + time.Now().Format("20060102150405")

		// Upload the image to Cloudinary
		imageURL, err := cloudinaryService.UploadImage(c, file, filename)
		if err != nil {
			log.Error().Err(err).Str("hedgehog_id", hedgehogID).Msg("Failed to upload image to Cloudinary")
			c.JSON(500, gin.H{"error": "Failed to upload image: " + err.Error()})
			return
		}

		// Update the hedgehog's Picture field with the new image URL
		hedgehog.Picture = imageURL
		if err := db.Save(&hedgehog).Error; err != nil {
			log.Error().Err(err).Str("hedgehog_id", hedgehogID).Msg("Failed to update hedgehog with new image URL")
			c.JSON(500, gin.H{"error": "Failed to update hedgehog: " + err.Error()})
			return
		}

		log.Info().
			Str("hedgehog_id", hedgehogID).
			Str("image_url", imageURL).
			Msg("Hedgehog image uploaded successfully")

		// Print the URL for debugging
		fmt.Printf("Returning image URL to client: %s\n", imageURL)

		c.JSON(200, gin.H{
			"message":   "Image uploaded successfully",
			"image_url": imageURL,
			"hedgehog":  hedgehog,
		})
	}
}

// isValidImageType checks if the file is a valid image type
func isValidImageType(header *multipart.FileHeader) bool {
	// List of allowed MIME types
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
	}

	contentType := header.Header.Get("Content-Type")
	return allowedTypes[contentType]
}