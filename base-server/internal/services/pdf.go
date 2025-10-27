package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// PDFConfig holds configuration for PDF generation
type PDFConfig struct {
	PageSize     string            `json:"page_size"`   // A4, A3, Letter, Legal
	Orientation  string            `json:"orientation"` // Portrait, Landscape
	MarginTop    string            `json:"margin_top"`  // 1cm, 0.75in
	MarginRight  string            `json:"margin_right"`
	MarginBottom string            `json:"margin_bottom"`
	MarginLeft   string            `json:"margin_left"`
	Quality      int               `json:"quality"` // 1-100
	Grayscale    bool              `json:"grayscale"`
	LowQuality   bool              `json:"low_quality"`
	EnableJS     bool              `json:"enable_js"`
	LoadTimeout  int               `json:"load_timeout"` // seconds
	Headers      map[string]string `json:"headers"`
}

// DefaultPDFConfig returns sensible defaults
func DefaultPDFConfig() PDFConfig {
	return PDFConfig{
		PageSize:     "A4",
		Orientation:  "Portrait",
		MarginTop:    "1cm",
		MarginRight:  "1cm",
		MarginBottom: "1cm",
		MarginLeft:   "1cm",
		Quality:      80,
		Grayscale:    false,
		LowQuality:   false,
		EnableJS:     true,
		LoadTimeout:  30,
		Headers:      make(map[string]string),
	}
}

// PDFTemplateData represents data to be injected into PDF templates
type PDFTemplateData struct {
	Template   string                 `json:"template"`    // template name
	Data       map[string]interface{} `json:"data"`        // template data
	Config     *PDFConfig             `json:"config"`      // optional config override
	OutputPath string                 `json:"output_path"` // optional file output path
}

// PDFService handles PDF generation from templates
type PDFService struct {
	TemplateDir string
	OutputDir   string
	Config      PDFConfig
	templates   map[string]*template.Template
}

// NewPDFService creates a new PDF service
func NewPDFService(templateDir, outputDir string, config *PDFConfig) *PDFService {
	if config == nil {
		defaultConfig := DefaultPDFConfig()
		config = &defaultConfig
	}

	service := &PDFService{
		TemplateDir: templateDir,
		OutputDir:   outputDir,
		Config:      *config,
		templates:   make(map[string]*template.Template),
	}

	// Load templates
	if err := service.LoadTemplates(); err != nil {
		log.Printf("Warning: Failed to load PDF templates: %v", err)
	}

	return service
}

// LoadTemplates loads all HTML templates from the template directory
func (s *PDFService) LoadTemplates() error {
	if _, err := os.Stat(s.TemplateDir); os.IsNotExist(err) {
		return fmt.Errorf("template directory does not exist: %s", s.TemplateDir)
	}

	return filepath.Walk(s.TemplateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".html") {
			return nil
		}

		// Extract template name from filename
		relPath, _ := filepath.Rel(s.TemplateDir, path)
		templateName := strings.TrimSuffix(relPath, ".html")

		// Load and parse template
		tmpl, err := template.ParseFiles(path)
		if err != nil {
			log.Printf("Warning: Failed to parse template %s: %v", templateName, err)
			return nil
		}

		s.templates[templateName] = tmpl
		log.Printf("Loaded PDF template: %s", templateName)
		return nil
	})
}

// GetAvailableTemplates returns list of available template names
func (s *PDFService) GetAvailableTemplates() []string {
	var templates []string
	for name := range s.templates {
		templates = append(templates, name)
	}
	return templates
}

// RenderTemplate renders a template with data to HTML string
func (s *PDFService) RenderTemplate(templateName string, data map[string]interface{}) (string, error) {
	tmpl, exists := s.templates[templateName]
	if !exists {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	// Add helper functions and metadata to template data
	templateData := s.enrichTemplateData(data)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to render template %s: %v", templateName, err)
	}

	return buf.String(), nil
}

// GeneratePDF generates a PDF from template data
func (s *PDFService) GeneratePDF(ctx context.Context, templateData PDFTemplateData) ([]byte, error) {
	// Render HTML template
	html, err := s.RenderTemplate(templateData.Template, templateData.Data)
	if err != nil {
		return nil, err
	}

	// Use provided config or default
	config := s.Config
	if templateData.Config != nil {
		config = *templateData.Config
	}

	return s.GeneratePDFFromHTML(ctx, html, config)
}

// GeneratePDFFromHTML generates PDF from HTML string
func (s *PDFService) GeneratePDFFromHTML(ctx context.Context, html string, config PDFConfig) ([]byte, error) {
	// Create temporary HTML file
	tmpDir := os.TempDir()
	tmpHtmlFile, err := os.CreateTemp(tmpDir, "pdf_*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp HTML file: %v", err)
	}
	defer os.Remove(tmpHtmlFile.Name())

	// Write HTML to temp file
	if _, err := tmpHtmlFile.WriteString(html); err != nil {
		return nil, fmt.Errorf("failed to write HTML to temp file: %v", err)
	}
	tmpHtmlFile.Close()

	// Create temporary PDF file
	tmpPdfFile, err := os.CreateTemp(tmpDir, "output_*.pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp PDF file: %v", err)
	}
	tmpPdfFile.Close()
	defer os.Remove(tmpPdfFile.Name())

	// Build wkhtmltopdf command
	args := s.buildWkHtmlToPdfArgs(config, tmpHtmlFile.Name(), tmpPdfFile.Name())

	// Execute command with context
	cmd := exec.CommandContext(ctx, "wkhtmltopdf", args...)

	// Capture output for debugging
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("wkhtmltopdf failed: %v, stderr: %s", err, stderr.String())
	}

	// Read generated PDF
	pdfBytes, err := os.ReadFile(tmpPdfFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read generated PDF: %v", err)
	}

	return pdfBytes, nil
}

// SavePDF saves PDF bytes to file
func (s *PDFService) SavePDF(pdfBytes []byte, filename string) (string, error) {
	if err := os.MkdirAll(s.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	outputPath := filepath.Join(s.OutputDir, filename)
	if err := os.WriteFile(outputPath, pdfBytes, 0644); err != nil {
		return "", fmt.Errorf("failed to save PDF: %v", err)
	}

	return outputPath, nil
}

// GenerateAndSavePDF generates and saves PDF in one operation
func (s *PDFService) GenerateAndSavePDF(ctx context.Context, templateData PDFTemplateData) (string, error) {
	pdfBytes, err := s.GeneratePDF(ctx, templateData)
	if err != nil {
		return "", err
	}

	// Generate filename if not provided
	filename := templateData.OutputPath
	if filename == "" {
		timestamp := time.Now().Format("20060102_150405")
		filename = fmt.Sprintf("%s_%s.pdf", templateData.Template, timestamp)
	}

	return s.SavePDF(pdfBytes, filename)
}

// buildWkHtmlToPdfArgs builds command line arguments for wkhtmltopdf
func (s *PDFService) buildWkHtmlToPdfArgs(config PDFConfig, inputFile, outputFile string) []string {
	args := []string{}

	// Page size
	args = append(args, "--page-size", config.PageSize)

	// Orientation
	if strings.ToLower(config.Orientation) == "landscape" {
		args = append(args, "--orientation", "Landscape")
	} else {
		args = append(args, "--orientation", "Portrait")
	}

	// Margins
	args = append(args, "--margin-top", config.MarginTop)
	args = append(args, "--margin-right", config.MarginRight)
	args = append(args, "--margin-bottom", config.MarginBottom)
	args = append(args, "--margin-left", config.MarginLeft)

	// Quality settings
	if config.Quality > 0 && config.Quality <= 100 {
		args = append(args, "--image-quality", fmt.Sprintf("%d", config.Quality))
	}

	if config.Grayscale {
		args = append(args, "--grayscale")
	}

	if config.LowQuality {
		args = append(args, "--lowquality")
	}

	// JavaScript
	if config.EnableJS {
		args = append(args, "--enable-javascript")
		if config.LoadTimeout > 0 {
			args = append(args, "--javascript-delay", fmt.Sprintf("%d", config.LoadTimeout*1000))
		}
	} else {
		args = append(args, "--disable-javascript")
	}

	// Additional options
	args = append(args, "--encoding", "UTF-8")
	args = append(args, "--quiet") // Suppress verbose output

	// Add custom headers
	for key, value := range config.Headers {
		args = append(args, "--custom-header", key, value)
	}

	// Input and output files
	args = append(args, inputFile, outputFile)

	return args
}

// enrichTemplateData adds helper functions and metadata to template data
func (s *PDFService) enrichTemplateData(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		data = make(map[string]interface{})
	}

	// Add timestamp information
	now := time.Now()
	data["GeneratedAt"] = now.Format(time.RFC3339)
	data["GeneratedAtFormatted"] = now.Format("January 2, 2006 at 3:04 PM")
	data["GeneratedDate"] = now.Format("2006-01-02")
	data["GeneratedTime"] = now.Format("15:04:05")

	// Add helper functions (can be used in templates)
	data["FormatDate"] = func(t time.Time, layout string) string {
		return t.Format(layout)
	}
	data["FormatCurrency"] = func(amount float64, currency string) string {
		return fmt.Sprintf("%.2f %s", amount, currency)
	}
	data["ToUpper"] = strings.ToUpper
	data["ToLower"] = strings.ToLower
	data["Join"] = strings.Join

	return data
}

// PreviewTemplate returns rendered HTML for template preview
func (s *PDFService) PreviewTemplate(templateName string, data map[string]interface{}) (string, error) {
	return s.RenderTemplate(templateName, data)
}

// ValidateTemplateData validates template data structure
func (s *PDFService) ValidateTemplateData(templateData PDFTemplateData) error {
	if templateData.Template == "" {
		return fmt.Errorf("template name is required")
	}

	if _, exists := s.templates[templateData.Template]; !exists {
		return fmt.Errorf("template not found: %s", templateData.Template)
	}

	if templateData.Data == nil {
		templateData.Data = make(map[string]interface{})
	}

	return nil
}

// GetTemplateInfo returns information about a specific template
func (s *PDFService) GetTemplateInfo(templateName string) (map[string]interface{}, error) {
	if _, exists := s.templates[templateName]; !exists {
		return nil, fmt.Errorf("template not found: %s", templateName)
	}

	templatePath := filepath.Join(s.TemplateDir, templateName+".html")
	stat, err := os.Stat(templatePath)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":      templateName,
		"path":      templatePath,
		"size":      stat.Size(),
		"modified":  stat.ModTime(),
		"available": true,
	}, nil
}

// StreamPDF streams PDF generation (useful for large files)
func (s *PDFService) StreamPDF(ctx context.Context, templateData PDFTemplateData, writer io.Writer) error {
	pdfBytes, err := s.GeneratePDF(ctx, templateData)
	if err != nil {
		return err
	}

	_, err = writer.Write(pdfBytes)
	return err
}
