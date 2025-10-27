package services

import (
	"context"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ae-saas-basic/ae-saas-basic/internal/services"
	"github.com/ae-saas-basic/ae-saas-basic/pkg/utils"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// PDFService provides PDF generation functionality
type PDFService struct {
	internalService *services.PDFService
}

// NewPDFService creates a new public PDF service
func NewPDFService(templateDir, outputDir string) *PDFService {
	config := services.DefaultPDFConfig()
	internalService := services.NewPDFService(templateDir, outputDir, &config)

	return &PDFService{
		internalService: internalService,
	}
}

// Invoice represents invoice data for PDF generation
type Invoice struct {
	ID             string  `json:"id"`
	Date           string  `json:"date"`
	Total          float64 `json:"total"`
	SubTotal       float64 `json:"subtotal"`
	ServicePeriod  string  `json:"service_period"`
	TaxRate        float64 `json:"tax_rate"`
	TaxAmount      float64 `json:"tax_amount"`
	NumberStdUnits int     `json:"number_std_units"`
	UnitRate       float64 `json:"unit_rate"`
}

// Organization represents organization data for PDF generation
type Organization struct {
	City          string `json:"city"`
	Description   string `json:"description"`
	ContactPerson string `json:"contact_person"`
	Email         string `json:"email"`
	Email2        string `json:"email2"`
	Extra         string `json:"extra"`
	Mobile        string `json:"mobile"`
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	Settings      struct {
		Currency   string `json:"currency"`
		DateFormat string `json:"date_format"`
	} `json:"settings"`
	State         string `json:"state"`
	Street        string `json:"street"`
	TaxID         string `json:"tax_id"`
	Vat           string `json:"vat"`
	AccountNumber string `json:"account_number"`
	BankName      string `json:"bank_name"`
	IBAN          string `json:"iban"`
	BIC           string `json:"bic"`
	Zip           string `json:"zip"`
}

// Session represents therapy session data
type Session struct {
	Date     string `json:"date"`
	Duration int    `json:"duration"`
	Notes    string `json:"notes"`
}

// Client represents client data
type Client struct {
	AdmissionDate string `json:"admission_date"`
	City          string `json:"city"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	DateOfBirth   string `json:"date_of_birth"`
}

// Therapy represents therapy information
type Therapy struct {
	Description    string `json:"description"`
	EndDate        string `json:"end_date"`
	StartDate      string `json:"start_date"`
	ApprovalNumber string `json:"approval_number"`
}

// Provider represents provider information
type Provider struct {
	ProviderName  string `json:"provider_name"`
	ContactPerson string `json:"contact_person"`
	City          string `json:"city"`
	State         string `json:"state"`
	Street        string `json:"street"`
	Zip           string `json:"zip"`
}

// DataExample represents complete data structure for PDF generation
type DataExample struct {
	Invoice      Invoice      `json:"invoice"`
	Organization Organization `json:"organization"`
	Sessions     []Session    `json:"sessions"`
	Client       Client       `json:"client"`
	Therapy      Therapy      `json:"therapy"`
	Provider     Provider     `json:"provider"`
}

// createHtml generates HTML from template and data
func (p *PDFService) createHtml(dataFile string, templateFile string, outFile string) error {
	jsonBytes, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return err
	}
	var data DataExample
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return err
	}
	tmpl, err := template.New(filepath.Base(templateFile)).Option("missingkey=zero").ParseFiles(templateFile)
	if err != nil {
		return err
	}
	outF, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer outF.Close()
	if err := tmpl.Execute(outF, data); err != nil {
		return err
	}
	log.Println(outFile, "generated successfully")
	return nil
}

// convertHtmlToPdf converts HTML file to PDF using Chrome/Chromium
func (p *PDFService) convertHtmlToPdf(tmpHtml string, outFile string) error {
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

	// Run chromedp tasks
	err := chromedp.Run(ctx,
		chromedp.Navigate("file://"+p.getFullPath(tmpHtml)),
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
	if err := ioutil.WriteFile(outFile, pdfBuf, 0644); err != nil {
		return err
	}

	log.Println("PDF successfully generated:", outFile)
	return nil
}

// getFullPath returns absolute path for chromedp file:// URL
func (p *PDFService) getFullPath(filename string) string {
	absPath, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting working directory: %v", err)
		return filename // fallback to relative path
	}
	return absPath + "/" + filename
}

// GeneratePDFWithCustomData generates PDF using custom data file
func (p *PDFService) GeneratePDFWithCustomData(customDataFile string) (string, error) {
	return p.GeneratePDFWithCustomDataAndFilename(customDataFile, "invoice_units_template")
}

// GeneratePDFWithCustomDataAndFilename generates PDF with custom data and filename
func (p *PDFService) GeneratePDFWithCustomDataAndFilename(customDataFile string, outputBaseName string) (string, error) {
	templatesDir := utils.GetEnv("PDF_TEMPLATES_DIR", "./statics/pdf_templates")
	outDir := utils.GetEnv("TEMP_PATH", "./tmp")

	templateFile := "invoice_units_template.html"
	outFile := filepath.Join(outDir, outputBaseName+".html")
	pdfFile := filepath.Join(outDir, outputBaseName+".pdf")

	// Use custom data file instead of default data_example.json
	err := p.createHtml(customDataFile, filepath.Join(templatesDir, templateFile), outFile)
	if err != nil {
		return "", err
	}

	err = p.convertHtmlToPdf(outFile, pdfFile)
	if err != nil {
		return "", err
	}

	return pdfFile, nil
}

// GeneratePDF generates a PDF using the default template and data
func (p *PDFService) GeneratePDF() (string, error) {
	templatesDir := utils.GetEnv("PDF_TEMPLATES_DIR", "./statics/pdf_templates")
	outDir := utils.GetEnv("TEMP_PATH", "./tmp")

	dataDir := filepath.Join(templatesDir, "data")
	dataFile := filepath.Join(dataDir, "data_example.json")

	templateFile := "invoice_units_template.html"
	outFile := filepath.Join(outDir, templateFile)
	pdfFile := strings.Replace(outFile, ".html", ".pdf", 1)

	err := p.createHtml(dataFile, filepath.Join(templatesDir, templateFile), outFile)
	if err != nil {
		return "", err
	}

	err = p.convertHtmlToPdf(outFile, pdfFile)
	if err != nil {
		return "", err
	}

	return pdfFile, nil
}
