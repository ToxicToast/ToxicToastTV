package ocr

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ReceiptItem represents a parsed item from the receipt
type ReceiptItem struct {
	Name         string
	Quantity     int
	UnitPrice    float64
	TotalPrice   float64
	ArticleNumber string
}

// ReceiptData represents parsed receipt data
type ReceiptData struct {
	Items      []ReceiptItem
	TotalPrice float64
	RawText    string
}

// TesseractOCR performs OCR on an image file
func TesseractOCR(imagePath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return "", fmt.Errorf("image file not found: %s", imagePath)
	}

	// Run tesseract
	cmd := exec.Command("tesseract", imagePath, "stdout", "-l", "deu")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tesseract error: %v, stderr: %s", err, stderr.String())
	}

	return out.String(), nil
}

// SaveImageFile saves image data to a file and converts PDFs to images
func SaveImageFile(imageData []byte, receiptID string) (string, error) {
	// Create receipts directory if it doesn't exist
	receiptsDir := "/app/receipts"
	if _, err := os.Stat(receiptsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(receiptsDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create receipts directory: %v", err)
		}
	}

	// Detect file type by checking magic bytes
	isPDF := len(imageData) > 4 && string(imageData[:4]) == "%PDF"

	if isPDF {
		// Save PDF temporarily
		pdfPath := filepath.Join(receiptsDir, fmt.Sprintf("%s.pdf", receiptID))
		if err := os.WriteFile(pdfPath, imageData, 0644); err != nil {
			return "", fmt.Errorf("failed to write PDF file: %v", err)
		}

		// Convert PDF to PNG using pdftoppm
		outputPath := filepath.Join(receiptsDir, receiptID)
		cmd := exec.Command("pdftoppm", "-png", "-singlefile", "-r", "300", pdfPath, outputPath)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("PDF conversion error: %v, stderr: %s", err, stderr.String())
		}

		// Clean up PDF
		os.Remove(pdfPath)

		// Return path to converted image
		imagePath := outputPath + ".png"
		return imagePath, nil
	}

	// Handle regular image files (JPG, PNG, etc.)
	filename := fmt.Sprintf("%s.jpg", receiptID)
	filePath := filepath.Join(receiptsDir, filename)

	// Write file
	if err := os.WriteFile(filePath, imageData, 0644); err != nil {
		return "", fmt.Errorf("failed to write image file: %v", err)
	}

	return filePath, nil
}

// ParseReceipt performs OCR and parses the receipt data
func ParseReceipt(imagePath string) (*ReceiptData, error) {
	// Perform OCR
	text, err := TesseractOCR(imagePath)
	if err != nil {
		return nil, err
	}

	// Parse the OCR text
	data := &ReceiptData{
		RawText: text,
		Items:   make([]ReceiptItem, 0),
	}

	// Parse items and total
	parseReceiptText(text, data)

	return data, nil
}

// parseReceiptText parses the OCR text to extract items and total
func parseReceiptText(text string, data *ReceiptData) {
	lines := strings.Split(text, "\n")

	// Regex patterns for parsing
	articlePattern := regexp.MustCompile(`Art[\.:]\s*(\d+)`)
	totalPattern := regexp.MustCompile(`(?i)(summe|total|gesamt|betrag).*?(\d+[,\.]\d{2})`)

	// REWE-specific patterns
	// Line 1: "PRODUCT NAME 12,34 B" (product + total price + tax class)
	reweItemPattern := regexp.MustCompile(`^(.+?)\s+(\d+[,\.]\d{2})\s+([A-Z])\s*[\*]?\s*$`)
	// Line 2: "2Stkx 3,29" or "2 ST x 3,29" or "2 x 3,29"
	reweDetailPattern := regexp.MustCompile(`^\s*(\d+)\s*(?:ST|Stk)?\s*x\s+(\d+[,\.]\d{2})`)

	// End-of-items markers
	endMarkers := regexp.MustCompile(`(?i)(geg\.|gegeben|visa|bar|karte|steuer|summe|total|gesamt|betrag|bonus|transaktion)`)

	// PFAND pattern - don't add as item, but track the price
	pfandPattern := regexp.MustCompile(`(?i)^pfand`)

	// Items to ignore (excluding PFAND as we handle it separately)
	ignoreItems := regexp.MustCompile(`(?i)^(mehrwertsteuer|mwst|rabatt|coupon)`)

	// Try REWE format
	i := 0
	itemsEnded := false
	totalPfand := 0.0

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		if line == "" {
			i++
			continue
		}

		// Check if items section has ended
		if endMarkers.MatchString(line) {
			itemsEnded = true
		}

		// Check for total
		if match := totalPattern.FindStringSubmatch(line); len(match) > 2 {
			totalStr := strings.ReplaceAll(match[2], ",", ".")
			if total, err := strconv.ParseFloat(totalStr, 64); err == nil {
				data.TotalPrice = total
			}
			i++
			continue
		}

		// Skip if items section has ended
		if itemsEnded {
			i++
			continue
		}

		// Try to match REWE item line (Line 1)
		if match := reweItemPattern.FindStringSubmatch(line); len(match) > 3 {
			itemName := strings.TrimSpace(match[1])
			totalPriceStr := strings.ReplaceAll(match[2], ",", ".")
			// taxClass := match[3] // A or B or C - we ignore this now

			// Check if this is PFAND - add to total but don't create item
			if pfandPattern.MatchString(itemName) {
				if pfandPrice, err := strconv.ParseFloat(totalPriceStr, 64); err == nil {
					totalPfand += pfandPrice
				}
				i++
				continue
			}

			// Skip other ignored items
			if ignoreItems.MatchString(itemName) {
				i++
				continue
			}

			totalPrice, _ := strconv.ParseFloat(totalPriceStr, 64)

			// Default values (will be overridden if Line 2 exists)
			quantity := 1
			unitPrice := totalPrice

			// Check if next line exists for quantity details (Line 2)
			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])

				if detailMatch := reweDetailPattern.FindStringSubmatch(nextLine); len(detailMatch) >= 3 {
					// We have a Line 2 with quantity details
					quantity, _ = strconv.Atoi(detailMatch[1])
					unitPriceStr := strings.ReplaceAll(detailMatch[2], ",", ".")
					unitPrice, _ = strconv.ParseFloat(unitPriceStr, 64)

					// Skip the detail line
					i++
				}
			}

			// Extract article number if present
			articleNumber := ""
			if artMatch := articlePattern.FindStringSubmatch(itemName); len(artMatch) > 1 {
				articleNumber = artMatch[1]
				itemName = strings.TrimSpace(articlePattern.ReplaceAllString(itemName, ""))
			}

			if itemName != "" && unitPrice > 0 {
				data.Items = append(data.Items, ReceiptItem{
					Name:          itemName,
					Quantity:      quantity,
					UnitPrice:     unitPrice,
					TotalPrice:    totalPrice,
					ArticleNumber: articleNumber,
				})
			}
		}

		i++
	}

	// If no total was found, calculate from items + pfand
	if data.TotalPrice == 0 {
		for _, item := range data.Items {
			data.TotalPrice += item.TotalPrice
		}
		data.TotalPrice += totalPfand
	}
}
