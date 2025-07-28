# Hedgehog Image Upload Feature

This document describes the implementation of the image upload feature for hedgehogs using Cloudinary.

## Overview

The feature allows users to:
- Upload images for hedgehogs
- View hedgehog images in the list and detail views
- Update images for existing hedgehogs

## Implementation Details

### 1. Database Changes

Added a `Picture` field to the `Hedgehog` entity in `models.go`:

```go
Picture string `json:"picture" gorm:"default:'https://res.cloudinary.com/demo/image/upload/v1312461204/sample.jpg'" example:"https://res.cloudinary.com/demo/image/upload/v1312461204/sample.jpg" description:"URL to the hedgehog's picture"`
```

The field has a default value pointing to a placeholder cartoon hedgehog image.

### 2. Cloudinary Integration

Created a new file `cloudinary.go` that implements:
- `CloudinaryService` struct for managing Cloudinary operations
- `NewCloudinaryService` function for initializing the service with credentials
- `UploadImage` method for uploading images to Cloudinary
- `UploadHedgehogImageHandler` for handling HTTP requests to upload images

### 3. API Endpoint

Added a new API endpoint in `main.go`:

```go
// Hedgehog image upload (only if Cloudinary is configured)
if cloudinaryService != nil {
    protected.POST("/hedgehogs/:id/image", UploadHedgehogImageHandler(db, cloudinaryService))
}
```

This endpoint accepts POST requests with multipart form data containing an image file.

### 4. Frontend Changes

#### Hedgehog Form

Updated `hedgehog-form.html` to:
- Add an image upload field
- Show a preview of the selected image
- Upload the image after creating the hedgehog

#### Hedgehog List

Updated `hedgehogs.html` to:
- Display hedgehog images in the card view
- Show the image in the details modal
- Add an upload button in the details modal for updating images

## Setup Instructions

### 1. Cloudinary Account

1. Sign up for a free Cloudinary account at https://cloudinary.com/
2. Get your Cloud Name, API Key, and API Secret from the dashboard

### 2. Environment Variables

Add the following environment variables to your `.env` file:

```
CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret
```

## Testing

### 1. Creating a New Hedgehog with Image

1. Click "Nuovo Riccio" to open the creation form
2. Fill in the required fields
3. Click "Choose File" in the Image section and select an image
4. Click "Salva" to create the hedgehog with the image

### 2. Updating an Existing Hedgehog's Image

1. Click on a hedgehog card to open the details modal
2. Hover over the image to see the camera icon
3. Click the camera icon and select a new image
4. The image will be uploaded and the UI will update automatically

## Troubleshooting

### Image Upload Fails

1. Check that your Cloudinary credentials are correct
2. Verify that the image file is in a supported format (JPEG, PNG, GIF)
3. Check the browser console for any JavaScript errors
4. Check the server logs for any backend errors

### Default Image Not Showing

If the default placeholder image is not showing, make sure the URL in the `Picture` field default value is accessible and valid.