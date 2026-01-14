package services

import (
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ae-base-server/pkg/utils"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// PDFGenerator represents a PDF generator service
type PDFGenerator struct{}

// NewPDFGenerator creates a new PDF generator instance
func NewPDFGenerator() *PDFGenerator {
	return &PDFGenerator{}
}

// getFullPath returns absolute path for chromedp file:// URL
func (pg *PDFGenerator) getFullPath(filename string) string {
	absPath, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting working directory: %v", err)
		return filename // fallback to relative path
	}
	return absPath + "/" + filename
}

// ConvertHtmlToPdf converts HTML file to PDF using Chrome/Chromium
func (pg *PDFGenerator) ConvertHtmlToPdf(htmlPath string, pdfPath string) error {
	// Create Chrome context with custom allocator for better environment compatibility
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-web-security", true),
	)

	// Check for custom Chrome path from environment
	if chromePath := os.Getenv("CHROME_BIN"); chromePath != "" {
		opts = append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.ExecPath(chromePath),
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.Flag("disable-extensions", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-web-security", true),
		)
		log.Printf("Using custom Chrome path: %s", chromePath)
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Allocate PDF buffer
	var pdfBuf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate("file://"+pg.getFullPath(htmlPath)),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			// page.PrintToPDF().Do returns ([]byte, *page.PrintToPDFReply, error)
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithMarginTop(0).
				WithMarginBottom(0).
				WithMarginLeft(0).
				WithMarginRight(0).
				WithPaperWidth(210 / 25.4).  // 210mm in inches
				WithPaperHeight(297 / 25.4). // 297mm in inches
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		return err
	}

	// Save PDF file
	if err := ioutil.WriteFile(pdfPath, pdfBuf, 0644); err != nil {
		return fmt.Errorf("error writing PDF file: %v", err)
	}

	log.Println("PDF successfully generated:", pdfPath)
	return nil
}

// CreateHtmlFromTemplate generates HTML from a template file and data
func (pg *PDFGenerator) CreateHtmlFromTemplate(data interface{}, templatePath, outputPath string) error {
	tmpl, err := template.New(filepath.Base(templatePath)).Option("missingkey=zero").ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	if err := tmpl.Execute(outFile, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	log.Println("HTML generated successfully:", outputPath)
	return nil
}

// GeneratePDFFromTemplate is a convenience method that generates HTML from template and then converts to PDF
func (pg *PDFGenerator) GeneratePDFFromTemplate(data interface{}, templateFile, outputFileName string) (string, error) {

	templatesDir := utils.GetEnv("TEMPLATES_DIR", "./statics/templates")
	templatePath := filepath.Join(templatesDir, templateFile)
	// Check if file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return "", fmt.Errorf("template file not found: %v", err)
	}

	outDir := utils.GetEnv("TEMP_PATH", "./tmp")
	// Check if directory exists
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		return "", fmt.Errorf("output directory not found: %v", err)
	}

	outFile := strings.Replace(outputFileName, ".html", "", 1)
	if outFile == "" {
		timestamp := time.Now().Format("20060102_150405")
		outFile = fmt.Sprintf("pdf_%s", timestamp)
	}
	pdfFile := outFile + ".pdf"
	outFile = outFile + ".html"

	// Generate HTML first
	if err := pg.CreateHtmlFromTemplate(data, templatePath, filepath.Join(outDir, outFile)); err != nil {
		return "", fmt.Errorf("HTML generation failed: %v", err)
	}

	// Convert HTML to PDF
	if err := pg.ConvertHtmlToPdf(filepath.Join(outDir, outFile), filepath.Join(outDir, pdfFile)); err != nil {
		return "", fmt.Errorf("PDF conversion failed: %v", err)
	}

	log.Println("PDF generated successfully:", pdfFile)
	return pdfFile, nil
}

// GeneratePDF generates a PDF with the given data, template name, and output filename
func (pg *PDFGenerator) GeneratePDF(data interface{}, templateName, outputFileName string) (string, error) {
	return pg.GeneratePDFFromTemplate(data, templateName, outputFileName)
}
