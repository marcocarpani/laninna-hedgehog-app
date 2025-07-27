# 🦔 La Ninna - Fonts Directory

## DejaVu Fonts for PDF Generation

This directory contains DejaVu fonts used for consistent PDF generation across all deployment platforms.

### Included Fonts
- **DejaVuSans.ttf** - Regular font
- **DejaVuSans-Bold.ttf** - Bold font
- **DejaVuSans-Oblique.ttf** - Italic font
- **DejaVuSans-BoldOblique.ttf** - Bold italic font

### Usage in Code
```go
pdf.AddUTF8Font("DejaVu", "", "./fonts/DejaVuSans.ttf")
pdf.AddUTF8Font("DejaVu", "B", "./fonts/DejaVuSans-Bold.ttf")
```

### License
DejaVu fonts are licensed under a Free license. See individual font files for details.

### Deployment
These fonts are automatically included in:
- Docker builds
- Fly.io deployments
- Railway deployments
- Render deployments

### Font Features
- **Unicode support** for international characters
- **Consistent rendering** across platforms
- **Optimized for PDF** generation
- **Small file size** for efficient deployment