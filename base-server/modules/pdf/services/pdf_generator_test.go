package services_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ae-base-server/modules/pdf/services"
	"github.com/ae-base-server/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPDFGenerator_HTMLStringConversion tests converting HTML string to PDF
func TestPDFGenerator_HTMLStringConversion(t *testing.T) {
	t.Skip("Skipping PDF tests - requires Chrome/Chromium installation")

	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	ctx := context.Background()

	t.Run("converts simple HTML string to PDF", func(t *testing.T) {
		generator := services.NewPDFGenerator()

		htmlContent := `<!DOCTYPE html>
<html>
<head><title>Test Invoice</title></head>
<body>
<h1>Invoice INV-2026-0001</h1>
<p>Customer: Test Customer GmbH</p>
<p>Total: €1,000.00</p>
</body>
</html>`

		pdfBytes, err := generator.ConvertHtmlStringToPDF(ctx, htmlContent)

		require.NoError(t, err)
		require.NotNil(t, pdfBytes)

		// Verify PDF format (should start with %PDF header)
		assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")), "PDF should start with %PDF header")
		assert.Greater(t, len(pdfBytes), 0, "PDF should have content")
	})

	t.Run("handles empty HTML", func(t *testing.T) {
		generator := services.NewPDFGenerator()

		pdfBytes, err := generator.ConvertHtmlStringToPDF(ctx, "")

		// Should still generate a PDF even with empty content
		require.NoError(t, err)
		assert.NotNil(t, pdfBytes)
	})

	t.Run("handles complex HTML with styles", func(t *testing.T) {
		generator := services.NewPDFGenerator()

		htmlContent := `<!DOCTYPE html>
<html>
<head>
<style>
body { font-family: Arial, sans-serif; margin: 20px; }
.header { background-color: #4CAF50; color: white; padding: 10px; }
.invoice-item { border-bottom: 1px solid #ddd; padding: 5px; }
</style>
</head>
<body>
<div class="header"><h1>Invoice</h1></div>
<div class="invoice-item">Item 1: €100.00</div>
<div class="invoice-item">Item 2: €200.00</div>
</body>
</html>`

		pdfBytes, err := generator.ConvertHtmlStringToPDF(ctx, htmlContent)

		require.NoError(t, err)
		assert.NotNil(t, pdfBytes)
	})
}

// TestPDFGenerator_TemplateGeneration tests generating PDF from template files
func TestPDFGenerator_TemplateGeneration(t *testing.T) {
	t.Skip("Skipping PDF template tests - requires Chrome/Chromium and template files")

	t.Run("generates PDF from template file", func(t *testing.T) {
		generator := services.NewPDFGenerator()

		// Set environment variables for test
		os.Setenv("TEMPLATES_DIR", "./statics/templates")
		os.Setenv("TEMP_PATH", "./tmp")
		defer os.Unsetenv("TEMPLATES_DIR")
		defer os.Unsetenv("TEMP_PATH")

		// Create test data
		data := map[string]interface{}{
			"invoice_number": "INV-2026-0001",
			"customer_name":  "Test Customer GmbH",
			"total":          "1000.00",
		}

		// Generate PDF (requires template file to exist)
		pdfPath, err := generator.GeneratePDFFromTemplate(data, "invoice_template.html", "test_invoice")

		if err != nil {
			t.Logf("Expected error if template doesn't exist: %v", err)
			return
		}

		require.NotEmpty(t, pdfPath)

		// Clean up generated file
		defer os.Remove(filepath.Join("./tmp", pdfPath))
	})

	t.Run("handles missing template file", func(t *testing.T) {
		generator := services.NewPDFGenerator()

		os.Setenv("TEMPLATES_DIR", "./statics/templates")
		os.Setenv("TEMP_PATH", "./tmp")
		defer os.Unsetenv("TEMPLATES_DIR")
		defer os.Unsetenv("TEMP_PATH")

		data := map[string]interface{}{}

		_, err := generator.GeneratePDFFromTemplate(data, "nonexistent_template.html", "test_output")
		assert.Error(t, err, "Should error when template file doesn't exist")
		assert.Contains(t, err.Error(), "template file not found")
	})

	t.Run("handles missing output directory", func(t *testing.T) {
		generator := services.NewPDFGenerator()

		os.Setenv("TEMPLATES_DIR", "./statics/templates")
		os.Setenv("TEMP_PATH", "/nonexistent/directory")
		defer os.Unsetenv("TEMPLATES_DIR")
		defer os.Unsetenv("TEMP_PATH")

		data := map[string]interface{}{}

		_, err := generator.GeneratePDFFromTemplate(data, "invoice_template.html", "test_output")
		assert.Error(t, err, "Should error when output directory doesn't exist")
		assert.Contains(t, err.Error(), "output directory not found")
	})
}

// TestPDFGenerator_FileConversion tests converting HTML file to PDF file
func TestPDFGenerator_FileConversion(t *testing.T) {
	t.Skip("Skipping PDF file conversion tests - requires Chrome/Chromium installation")

	t.Run("converts HTML file to PDF file", func(t *testing.T) {
		generator := services.NewPDFGenerator()

		// Create temporary HTML file
		tmpDir := t.TempDir()
		htmlFile := filepath.Join(tmpDir, "test.html")
		pdfFile := filepath.Join(tmpDir, "test.pdf")

		htmlContent := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body><h1>Test PDF Generation</h1></body>
</html>`

		err := os.WriteFile(htmlFile, []byte(htmlContent), 0644)
		require.NoError(t, err)

		// Convert HTML to PDF
		err = generator.ConvertHtmlToPdf(htmlFile, pdfFile)
		require.NoError(t, err)

		// Verify PDF file was created
		pdfBytes, err := os.ReadFile(pdfFile)
		require.NoError(t, err)
		assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")), "PDF should start with %PDF header")
	})

	t.Run("handles missing HTML file", func(t *testing.T) {
		generator := services.NewPDFGenerator()

		tmpDir := t.TempDir()
		pdfFile := filepath.Join(tmpDir, "test.pdf")

		err := generator.ConvertHtmlToPdf("/nonexistent/file.html", pdfFile)
		assert.Error(t, err, "Should error when HTML file doesn't exist")
	})
}

// TestPDFGenerator_CreateHTML tests HTML generation from templates
func TestPDFGenerator_CreateHTML(t *testing.T) {
	t.Run("creates HTML from template", func(t *testing.T) {
		generator := services.NewPDFGenerator()

		// Create temporary template file
		tmpDir := t.TempDir()
		templateFile := filepath.Join(tmpDir, "template.html")
		outputFile := filepath.Join(tmpDir, "output.html")

		templateContent := `<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body><h1>{{.Heading}}</h1></body>
</html>`

		err := os.WriteFile(templateFile, []byte(templateContent), 0644)
		require.NoError(t, err)

		// Create HTML from template
		data := map[string]interface{}{
			"Title":   "Test Invoice",
			"Heading": "Invoice Details",
		}

		err = generator.CreateHtmlFromTemplate(data, templateFile, outputFile)
		require.NoError(t, err)

		// Verify output file was created
		htmlBytes, err := os.ReadFile(outputFile)
		require.NoError(t, err)

		htmlContent := string(htmlBytes)
		assert.Contains(t, htmlContent, "Test Invoice")
		assert.Contains(t, htmlContent, "Invoice Details")
	})

	t.Run("handles template parsing errors", func(t *testing.T) {
		generator := services.NewPDFGenerator()

		tmpDir := t.TempDir()
		templateFile := filepath.Join(tmpDir, "template.html")
		outputFile := filepath.Join(tmpDir, "output.html")

		// Invalid template syntax
		invalidTemplate := `<html>{{.MissingClosing</html>`

		err := os.WriteFile(templateFile, []byte(invalidTemplate), 0644)
		require.NoError(t, err)

		data := map[string]interface{}{}

		err = generator.CreateHtmlFromTemplate(data, templateFile, outputFile)
		assert.Error(t, err, "Should error with invalid template")
		assert.Contains(t, err.Error(), "failed to parse template")
	})

	t.Run("handles missing template file", func(t *testing.T) {
		generator := services.NewPDFGenerator()

		tmpDir := t.TempDir()
		outputFile := filepath.Join(tmpDir, "output.html")

		data := map[string]interface{}{}

		err := generator.CreateHtmlFromTemplate(data, "/nonexistent/template.html", outputFile)
		assert.Error(t, err, "Should error when template file doesn't exist")
	})
}
