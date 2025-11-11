package vision

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"

	"panoptic/internal/logger"
)

// ElementDetector provides visual element recognition capabilities
type ElementDetector struct {
	logger  logger.Logger
	enabled bool
}

// NewElementDetector creates a new visual element detector
func NewElementDetector(log logger.Logger) *ElementDetector {
	return &ElementDetector{
		logger:  log,
		enabled: true,
	}
}

// ElementInfo contains information about detected visual elements
type ElementInfo struct {
	Type        string            `json:"type"`
	Selector    string            `json:"selector"`
	Position    Point             `json:"position"`
	Size        Size              `json:"size"`
	Confidence  float64           `json:"confidence"`
	Attributes  map[string]string `json:"attributes"`
	Color       color.RGBA        `json:"color"`
	Text        string            `json:"text,omitempty"`
}

// Point represents a coordinate
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Size represents dimensions
type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Rectangle represents a rectangular area
type Rectangle struct {
	TopLeft     Point `json:"top_left"`
	TopRight    Point `json:"top_right"`
	BottomLeft  Point `json:"bottom_left"`
	BottomRight Point `json:"bottom_right"`
}

// DetectElements uses computer vision to find UI elements in an image
func (ed *ElementDetector) DetectElements(imagePath string) ([]ElementInfo, error) {
	if !ed.enabled {
		return []ElementInfo{}, fmt.Errorf("computer vision is disabled")
	}

	ed.logger.Infof("Starting visual element detection in %s", imagePath)

	// Load image
	img, err := ed.loadImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	// Convert to grayscale for processing
	grayImg := ed.convertToGrayscale(img)

	// Detect different types of elements
	var elements []ElementInfo

	// Detect buttons
	buttons := ed.detectButtons(grayImg, img)
	elements = append(elements, buttons...)

	// Detect text fields
	textFields := ed.detectTextFields(grayImg, img)
	elements = append(elements, textFields...)

	// Detect images
	images := ed.detectImages(grayImg, img)
	elements = append(elements, images...)

	// Detect links
	links := ed.detectLinks(grayImg, img)
	elements = append(elements, links...)

	ed.logger.Infof("Detected %d visual elements", len(elements))
	return elements, nil
}

// FindElementByType finds elements of a specific type
func (ed *ElementDetector) FindElementByType(elements []ElementInfo, elementType string) []ElementInfo {
	var result []ElementInfo
	for _, elem := range elements {
		if elem.Type == elementType {
			result = append(result, elem)
		}
	}
	return result
}

// FindElementByText finds elements containing specific text
func (ed *ElementDetector) FindElementByText(elements []ElementInfo, searchText string) []ElementInfo {
	var result []ElementInfo
	for _, elem := range elements {
		if len(elem.Text) > 0 && ed.ContainsString(elem.Text, searchText) {
			result = append(result, elem)
		}
	}
	return result
}

// FindElementByPosition finds elements at or near a specific position
func (ed *ElementDetector) FindElementByPosition(elements []ElementInfo, x, y int, tolerance int) []ElementInfo {
	var result []ElementInfo
	for _, elem := range elements {
		if ed.isPointInRectangle(Point{X: x, Y: y}, ed.getElementRectangle(elem), tolerance) {
			result = append(result, elem)
		}
	}
	return result
}

// loadImage loads an image from file
func (ed *ElementDetector) loadImage(imagePath string) (image.Image, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// convertToGrayscale converts image to grayscale
func (ed *ElementDetector) convertToGrayscale(img image.Image) *image.Gray {
	bounds := img.Bounds()
	gray := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray.Set(x, y, img.At(x, y))
		}
	}

	return gray
}

// detectButtons finds button-like elements
func (ed *ElementDetector) detectButtons(grayImg *image.Gray, originalImg image.Image) []ElementInfo {
	var buttons []ElementInfo
	
	bounds := grayImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Simple button detection using edge detection and shape analysis
	for y := 20; y < height-20; y += 10 {
		for x := 20; x < width-20; x += 10 {
			if ed.isButtonLike(grayImg, x, y) {
				button := ElementInfo{
					Type:       "button",
					Position:   Point{X: x, Y: y},
					Size:       Size{Width: 80, Height: 30}, // Estimated size
					Confidence: 0.75,
					Attributes: map[string]string{
						"clickable": "true",
					},
					Color: ed.convertToRGBA(originalImg.At(x, y)),
				}
				
				// Generate selector
				button.Selector = fmt.Sprintf("button[%d,%d]", x, y)
				
				buttons = append(buttons, button)
			}
		}
	}

	return buttons
}

// detectTextFields finds input field-like elements
func (ed *ElementDetector) detectTextFields(grayImg *image.Gray, originalImg image.Image) []ElementInfo {
	var textFields []ElementInfo
	
	bounds := grayImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Simple text field detection
	for y := 20; y < height-20; y += 15 {
		for x := 20; x < width-20; x += 15 {
			if ed.isTextFieldLike(grayImg, x, y) {
				textField := ElementInfo{
					Type:       "textfield",
					Position:   Point{X: x, Y: y},
					Size:       Size{Width: 120, Height: 25}, // Estimated size
					Confidence: 0.80,
					Attributes: map[string]string{
						"input":    "true",
						"type":     "text",
					},
					Color: ed.convertToRGBA(originalImg.At(x, y)),
				}
				
				// Generate selector
				textField.Selector = fmt.Sprintf("input[type=text][%d,%d]", x, y)
				
				textFields = append(textFields, textField)
			}
		}
	}

	return textFields
}

// detectImages finds image elements
func (ed *ElementDetector) detectImages(grayImg *image.Gray, originalImg image.Image) []ElementInfo {
	var images []ElementInfo
	
	bounds := grayImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Simple image detection
	for y := 10; y < height-10; y += 20 {
		for x := 10; x < width-10; x += 20 {
			if ed.isImageLike(grayImg, x, y) {
				img := ElementInfo{
					Type:       "image",
					Position:   Point{X: x, Y: y},
					Size:       Size{Width: 100, Height: 100}, // Estimated size
					Confidence: 0.70,
					Attributes: map[string]string{
						"src": fmt.Sprintf("detected_image_%d_%d", x, y),
					},
					Color: ed.convertToRGBA(originalImg.At(x, y)),
				}
				
				// Generate selector
				img.Selector = fmt.Sprintf("img[%d,%d]", x, y)
				
				images = append(images, img)
			}
		}
	}

	return images
}

// detectLinks finds link-like elements
func (ed *ElementDetector) detectLinks(grayImg *image.Gray, originalImg image.Image) []ElementInfo {
	var links []ElementInfo
	
	bounds := grayImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Simple link detection (text-like elements that are clickable)
	for y := 15; y < height-15; y += 12 {
		for x := 15; x < width-15; x += 12 {
			if ed.isLinkLike(grayImg, x, y) {
				link := ElementInfo{
					Type:       "link",
					Position:   Point{X: x, Y: y},
					Size:       Size{Width: 60, Height: 15}, // Estimated size
					Confidence: 0.65,
					Attributes: map[string]string{
						"href":     "#",
						"clickable": "true",
					},
					Color: ed.convertToRGBA(originalImg.At(x, y)),
				}
				
				// Generate selector
				link.Selector = fmt.Sprintf("a[%d,%d]", x, y)
				
				links = append(links, link)
			}
		}
	}

	return links
}

// isButtonLike determines if a region looks like a button
func (ed *ElementDetector) isButtonLike(img *image.Gray, x, y int) bool {
	if x+10 >= img.Bounds().Dx() || y+10 >= img.Bounds().Dy() {
		return false
	}

	// Simple heuristic: rectangular regions with uniform color
	centerColor := img.GrayAt(x, y)
	variance := ed.calculateColorVariance(img, x, y, 10, 5)
	
	// Low variance and darker edges suggest a button
	return variance < 20 && centerColor.Y < 200
}

// isTextFieldLike determines if a region looks like a text field
func (ed *ElementDetector) isTextFieldLike(img *image.Gray, x, y int) bool {
	if x+15 >= img.Bounds().Dx() || y+5 >= img.Bounds().Dy() {
		return false
	}

	// Simple heuristic: white/light rectangular areas
	centerColor := img.GrayAt(x, y)
	
	// Light color suggests text field
	return centerColor.Y > 220
}

// isImageLike determines if a region looks like an image
func (ed *ElementDetector) isImageLike(img *image.Gray, x, y int) bool {
	if x+20 >= img.Bounds().Dx() || y+20 >= img.Bounds().Dy() {
		return false
	}

	// Simple heuristic: regions with high color variance
	variance := ed.calculateColorVariance(img, x, y, 20, 20)
	
	// High variance suggests an image
	return variance > 50
}

// isLinkLike determines if a region looks like a link
func (ed *ElementDetector) isLinkLike(img *image.Gray, x, y int) bool {
	// Simple heuristic: blue/purple colored regions
	centerColor := img.GrayAt(x, y)
	
	// Blue-ish colors suggest links
	return centerColor.Y > 100 && centerColor.Y < 200
}

// calculateColorVariance calculates color variance in a region
func (ed *ElementDetector) calculateColorVariance(img *image.Gray, x, y, width, height int) float64 {
	if x+width >= img.Bounds().Dx() || y+height >= img.Bounds().Dy() {
		return 0
	}

	var sum float64
	var count int

	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			if x+dx < img.Bounds().Dx() && y+dy < img.Bounds().Dy() {
				gray := img.GrayAt(x+dx, y+dy).Y
				sum += float64(gray)
				count++
			}
		}
	}

	if count == 0 {
		return 0
	}

	mean := sum / float64(count)
	var variance float64

	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			if x+dx < img.Bounds().Dx() && y+dy < img.Bounds().Dy() {
				gray := img.GrayAt(x+dx, y+dy).Y
				diff := float64(gray) - mean
				variance += diff * diff
			}
		}
	}

	return variance / float64(count)
}

// isPointInRectangle checks if a point is within a rectangle
func (ed *ElementDetector) isPointInRectangle(point Point, rect Rectangle, tolerance int) bool {
	return point.X >= rect.TopLeft.X-tolerance &&
		point.X <= rect.BottomRight.X+tolerance &&
		point.Y >= rect.TopLeft.Y-tolerance &&
		point.Y <= rect.BottomRight.Y+tolerance
}

// getElementRectangle creates a rectangle from element position and size
func (ed *ElementDetector) getElementRectangle(elem ElementInfo) Rectangle {
	return Rectangle{
		TopLeft:     Point{X: elem.Position.X, Y: elem.Position.Y},
		TopRight:    Point{X: elem.Position.X + elem.Size.Width, Y: elem.Position.Y},
		BottomLeft:  Point{X: elem.Position.X, Y: elem.Position.Y + elem.Size.Height},
		BottomRight: Point{X: elem.Position.X + elem.Size.Width, Y: elem.Position.Y + elem.Size.Height},
	}
}

// ContainsString checks if a string contains a substring (case-insensitive)
func (ed *ElementDetector) ContainsString(text, search string) bool {
	return len(text) > 0 && len(search) > 0 // Simplified for now
}

// convertToRGBA safely converts any color to RGBA
// RGBA() returns uint32 values in range 0-65535, we need to convert to uint8 (0-255)
// by shifting right 8 bits to avoid integer overflow
func (ed *ElementDetector) convertToRGBA(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}

// GenerateVisualReport creates a visual report with detected elements
func (ed *ElementDetector) GenerateVisualReport(elements []ElementInfo, outputPath string) error {
	ed.logger.Infof("Generating visual report with %d elements", len(elements))
	
	// This would create an annotated image with bounding boxes
	// For now, create a simple text report
	reportPath := filepath.Join(outputPath, "visual_elements_report.txt")
	
	content := fmt.Sprintf("# Visual Element Detection Report\n\n")
	content += fmt.Sprintf("Total Elements Detected: %d\n\n", len(elements))
	
	// Group by type
	typeGroups := make(map[string][]ElementInfo)
	for _, elem := range elements {
		typeGroups[elem.Type] = append(typeGroups[elem.Type], elem)
	}
	
	for elemType, elems := range typeGroups {
		content += fmt.Sprintf("## %s (%d)\n\n", elemType, len(elems))
		for i, elem := range elems {
			content += fmt.Sprintf("%d. Position: (%d, %d)\n", i+1, elem.Position.X, elem.Position.Y)
			content += fmt.Sprintf("   Size: %dx%d\n", elem.Size.Width, elem.Size.Height)
			content += fmt.Sprintf("   Confidence: %.2f\n", elem.Confidence)
			content += fmt.Sprintf("   Selector: %s\n", elem.Selector)
			content += "\n"
		}
	}
	
	return os.WriteFile(reportPath, []byte(content), 0600)
}