package services

import (
	"encoding/xml"
	"fmt"

	baseAPI "github.com/ae-base-server/api"
	"github.com/unburdy/unburdy-server-api/internal/models"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	clientServices "github.com/unburdy/unburdy-server-api/modules/client_management/services"
	"gorm.io/gorm"
)

// XRechnungService handles generation of XRechnung XML for German government invoicing
type XRechnungService struct {
	db *gorm.DB
}

// NewXRechnungService creates a new XRechnung service instance
func NewXRechnungService(db *gorm.DB) *XRechnungService {
	return &XRechnungService{db: db}
}

// XRechnung UBL XML Structures
// Based on XRechnung 3.0.1 (UBL 2.1)

type UBLInvoice struct {
	XMLName xml.Name `xml:"ubl:Invoice"`
	XMLNS   string   `xml:"xmlns:ubl,attr"`
	XSICBC  string   `xml:"xmlns:cbc,attr"`
	XSICAC  string   `xml:"xmlns:cac,attr"`

	CustomizationID      string           `xml:"cbc:CustomizationID"`
	ProfileID            string           `xml:"cbc:ProfileID"`
	ID                   string           `xml:"cbc:ID"`
	IssueDate            string           `xml:"cbc:IssueDate"`
	DueDate              string           `xml:"cbc:DueDate"`
	InvoiceTypeCode      string           `xml:"cbc:InvoiceTypeCode"`
	DocumentCurrencyCode string           `xml:"cbc:DocumentCurrencyCode"`
	BuyerReference       string           `xml:"cbc:BuyerReference"`
	AccountingSupplier   UBLParty         `xml:"cac:AccountingSupplierParty"`
	AccountingCustomer   UBLParty         `xml:"cac:AccountingCustomerParty"`
	PaymentMeans         *UBLPaymentMeans `xml:"cac:PaymentMeans,omitempty"`
	TaxTotal             UBLTaxTotal      `xml:"cac:TaxTotal"`
	LegalMonetaryTotal   UBLMonetaryTotal `xml:"cac:LegalMonetaryTotal"`
	InvoiceLines         []UBLInvoiceLine `xml:"cac:InvoiceLine"`
}

type UBLParty struct {
	Party UBLPartyDetail `xml:"cac:Party"`
}

type UBLPartyDetail struct {
	EndpointID       *UBLEndpoint         `xml:"cbc:EndpointID,omitempty"`
	PartyName        *UBLPartyName        `xml:"cac:PartyName,omitempty"`
	PostalAddress    UBLAddress           `xml:"cac:PostalAddress"`
	PartyTaxScheme   *UBLPartyTaxScheme   `xml:"cac:PartyTaxScheme,omitempty"`
	PartyLegalEntity *UBLPartyLegalEntity `xml:"cac:PartyLegalEntity,omitempty"`
	Contact          *UBLContact          `xml:"cac:Contact,omitempty"`
}

type UBLEndpoint struct {
	SchemeID string `xml:"schemeID,attr"`
	Value    string `xml:",chardata"`
}

type UBLPartyName struct {
	Name string `xml:"cbc:Name"`
}

type UBLAddress struct {
	StreetName           string      `xml:"cbc:StreetName,omitempty"`
	AdditionalStreetName string      `xml:"cbc:AdditionalStreetName,omitempty"`
	CityName             string      `xml:"cbc:CityName,omitempty"`
	PostalZone           string      `xml:"cbc:PostalZone,omitempty"`
	Country              *UBLCountry `xml:"cac:Country,omitempty"`
}

type UBLCountry struct {
	IdentificationCode string `xml:"cbc:IdentificationCode"`
}

type UBLPartyTaxScheme struct {
	CompanyID string       `xml:"cbc:CompanyID"`
	TaxScheme UBLTaxScheme `xml:"cac:TaxScheme"`
}

type UBLTaxScheme struct {
	ID string `xml:"cbc:ID"`
}

type UBLPartyLegalEntity struct {
	RegistrationName string `xml:"cbc:RegistrationName"`
}

type UBLContact struct {
	Name           string `xml:"cbc:Name,omitempty"`
	Telephone      string `xml:"cbc:Telephone,omitempty"`
	ElectronicMail string `xml:"cbc:ElectronicMail,omitempty"`
}

type UBLPaymentMeans struct {
	PaymentMeansCode      string               `xml:"cbc:PaymentMeansCode"`
	PaymentID             *string              `xml:"cbc:PaymentID,omitempty"`
	PayeeFinancialAccount *UBLFinancialAccount `xml:"cac:PayeeFinancialAccount,omitempty"`
}

type UBLFinancialAccount struct {
	ID                         string     `xml:"cbc:ID"`
	Name                       *string    `xml:"cbc:Name,omitempty"`
	FinancialInstitutionBranch *UBLBranch `xml:"cac:FinancialInstitutionBranch,omitempty"`
}

type UBLBranch struct {
	ID string `xml:"cbc:ID"`
}

type UBLTaxTotal struct {
	TaxAmount    UBLAmount        `xml:"cbc:TaxAmount"`
	TaxSubtotals []UBLTaxSubtotal `xml:"cac:TaxSubtotal"`
}

type UBLTaxSubtotal struct {
	TaxableAmount UBLAmount      `xml:"cbc:TaxableAmount"`
	TaxAmount     UBLAmount      `xml:"cbc:TaxAmount"`
	TaxCategory   UBLTaxCategory `xml:"cac:TaxCategory"`
}

type UBLTaxCategory struct {
	ID                 string       `xml:"cbc:ID"`
	Percent            *float64     `xml:"cbc:Percent,omitempty"`
	TaxExemptionReason *string      `xml:"cbc:TaxExemptionReason,omitempty"`
	TaxScheme          UBLTaxScheme `xml:"cac:TaxScheme"`
}

type UBLMonetaryTotal struct {
	LineExtensionAmount UBLAmount `xml:"cbc:LineExtensionAmount"`
	TaxExclusiveAmount  UBLAmount `xml:"cbc:TaxExclusiveAmount"`
	TaxInclusiveAmount  UBLAmount `xml:"cbc:TaxInclusiveAmount"`
	PayableAmount       UBLAmount `xml:"cbc:PayableAmount"`
}

type UBLInvoiceLine struct {
	ID                  string      `xml:"cbc:ID"`
	InvoicedQuantity    UBLQuantity `xml:"cbc:InvoicedQuantity"`
	LineExtensionAmount UBLAmount   `xml:"cbc:LineExtensionAmount"`
	Item                UBLItem     `xml:"cac:Item"`
	Price               UBLPrice    `xml:"cac:Price"`
}

type UBLQuantity struct {
	UnitCode string  `xml:"unitCode,attr"`
	Value    float64 `xml:",chardata"`
}

type UBLAmount struct {
	CurrencyID string  `xml:"currencyID,attr"`
	Value      float64 `xml:",chardata"`
}

type UBLItem struct {
	Description           string         `xml:"cbc:Description"`
	Name                  string         `xml:"cbc:Name"`
	ClassifiedTaxCategory UBLTaxCategory `xml:"cac:ClassifiedTaxCategory"`
}

type UBLPrice struct {
	PriceAmount UBLAmount `xml:"cbc:PriceAmount"`
}

// GenerateXRechnungXML generates XRechnung-compliant UBL XML for a finalized invoice
func (s *XRechnungService) GenerateXRechnungXML(
	invoice *entities.Invoice,
	organization *baseAPI.Organization,
	costProvider *models.CostProvider,
) ([]byte, error) {
	// Validation
	if invoice.Status != entities.InvoiceStatusSent &&
		invoice.Status != entities.InvoiceStatusPaid &&
		invoice.Status != entities.InvoiceStatusOverdue {
		return nil, fmt.Errorf("invoice must be finalized (status: sent, paid, or overdue)")
	}

	if !costProvider.IsGovernmentCustomer {
		return nil, fmt.Errorf("XRechnung is only for government customers")
	}

	if costProvider.LeitwegID == "" {
		return nil, fmt.Errorf("Leitweg-ID is required for XRechnung")
	}

	// Build UBL Invoice structure
	ublInvoice := UBLInvoice{
		XMLNS:  "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2",
		XSICBC: "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2",
		XSICAC: "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2",

		// XRechnung profile identifiers
		CustomizationID: "urn:cen.eu:en16931:2017#compliant#urn:xeinkauf.de:kosit:xrechnung_3.0",
		ProfileID:       "urn:fdc:peppol.eu:2017:poacc:billing:01:1.0",

		ID:                   invoice.InvoiceNumber,
		IssueDate:            invoice.InvoiceDate.Format("2006-01-02"),
		InvoiceTypeCode:      s.getInvoiceTypeCode(invoice),
		DocumentCurrencyCode: "EUR",
		BuyerReference:       costProvider.LeitwegID,
	}

	// Due date (if payment not yet received)
	if invoice.PayedDate == nil {
		// Get payment terms from settings
		settingsHelper := clientServices.NewSettingsHelper(s.db)
		paymentTerms, err := settingsHelper.GetPaymentTerms(organization.TenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to get payment terms: %w", err)
		}
		dueDate := invoice.InvoiceDate.AddDate(0, 0, paymentTerms.PaymentDueDays)
		ublInvoice.DueDate = dueDate.Format("2006-01-02")
	}

	// Supplier (Organization)
	ublInvoice.AccountingSupplier = s.buildSupplierParty(organization)

	// Customer (Cost Provider / Government Authority)
	ublInvoice.AccountingCustomer = s.buildCustomerParty(costProvider)

	// Payment means (if bank details available)
	if organization.BankAccountIBAN != "" {
		ublInvoice.PaymentMeans = s.buildPaymentMeans(organization, invoice)
	}

	// Tax totals
	ublInvoice.TaxTotal = s.buildTaxTotal(invoice)

	// Monetary totals
	ublInvoice.LegalMonetaryTotal = s.buildMonetaryTotal(invoice)

	// Invoice lines
	ublInvoice.InvoiceLines = s.buildInvoiceLines(invoice)

	// Marshal to XML with indentation
	xmlData, err := xml.MarshalIndent(ublInvoice, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to generate XML: %w", err)
	}

	// Add XML declaration
	xmlWithDeclaration := []byte(xml.Header + string(xmlData))

	return xmlWithDeclaration, nil
}

func (s *XRechnungService) getInvoiceTypeCode(invoice *entities.Invoice) string {
	if invoice.IsCreditNote {
		return "384" // Credit note
	}
	return "380" // Commercial invoice
}

func (s *XRechnungService) buildSupplierParty(org *baseAPI.Organization) UBLParty {
	party := UBLParty{
		Party: UBLPartyDetail{
			PartyName: &UBLPartyName{
				Name: org.Name,
			},
			PostalAddress: UBLAddress{
				StreetName: org.StreetAddress,
				CityName:   org.City,
				PostalZone: org.Zip,
				Country: &UBLCountry{
					IdentificationCode: "DE",
				},
			},
		},
	}

	// Tax ID (required for supplier)
	if org.TaxID != "" {
		party.Party.PartyTaxScheme = &UBLPartyTaxScheme{
			CompanyID: org.TaxID,
			TaxScheme: UBLTaxScheme{
				ID: "VAT",
			},
		}
	}

	// Legal entity name
	party.Party.PartyLegalEntity = &UBLPartyLegalEntity{
		RegistrationName: org.Name,
	}

	// Contact information
	if org.Email != "" || org.Phone != "" {
		contact := &UBLContact{}
		if org.Email != "" {
			contact.ElectronicMail = org.Email
		}
		if org.Phone != "" {
			contact.Telephone = org.Phone
		}
		party.Party.Contact = contact
	}

	return party
}

func (s *XRechnungService) buildCustomerParty(cp *models.CostProvider) UBLParty {
	party := UBLParty{
		Party: UBLPartyDetail{
			// Leitweg-ID as endpoint for routing
			EndpointID: &UBLEndpoint{
				SchemeID: "0204", // Leitweg-ID scheme
				Value:    cp.LeitwegID,
			},
			PartyName: &UBLPartyName{
				Name: cp.AuthorityName,
			},
			PostalAddress: UBLAddress{
				StreetName: cp.StreetAddress,
				CityName:   cp.City,
				PostalZone: cp.Zip,
				Country: &UBLCountry{
					IdentificationCode: "DE", // Germany
				},
			},
			PartyLegalEntity: &UBLPartyLegalEntity{
				RegistrationName: cp.AuthorityName,
			},
		},
	}

	return party
}

func (s *XRechnungService) buildPaymentMeans(org *baseAPI.Organization, invoice *entities.Invoice) *UBLPaymentMeans {
	pm := &UBLPaymentMeans{
		PaymentMeansCode: "58", // SEPA credit transfer
	}

	// Payment reference
	paymentRef := invoice.InvoiceNumber
	pm.PaymentID = &paymentRef

	// Bank account details
	if org.BankAccountIBAN != "" {
		account := &UBLFinancialAccount{
			ID: org.BankAccountIBAN,
		}

		if org.BankAccountOwner != "" {
			account.Name = &org.BankAccountOwner
		}

		if org.BankAccountBIC != "" {
			account.FinancialInstitutionBranch = &UBLBranch{
				ID: org.BankAccountBIC,
			}
		}

		pm.PayeeFinancialAccount = account
	}

	return pm
}

func (s *XRechnungService) buildTaxTotal(invoice *entities.Invoice) UBLTaxTotal {
	// Group tax amounts by rate
	taxByRate := make(map[float64]float64)
	subtotalByRate := make(map[float64]float64)
	exemptSubtotal := 0.0
	exemptionText := ""

	for _, item := range invoice.InvoiceItems {
		if item.VATExempt {
			exemptSubtotal += item.TotalAmount
			if exemptionText == "" && item.VATExemptionText != "" {
				exemptionText = item.VATExemptionText
			}
		} else {
			rate := item.VATRate
			taxByRate[rate] += item.TotalAmount * (rate / 100.0)
			subtotalByRate[rate] += item.TotalAmount
		}
	}

	// Build tax subtotals
	var subtotals []UBLTaxSubtotal

	// Tax-exempt category (if any)
	if exemptSubtotal > 0 {
		exemptReason := exemptionText
		if exemptReason == "" {
			exemptReason = "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG"
		}

		subtotals = append(subtotals, UBLTaxSubtotal{
			TaxableAmount: UBLAmount{CurrencyID: "EUR", Value: exemptSubtotal},
			TaxAmount:     UBLAmount{CurrencyID: "EUR", Value: 0.0},
			TaxCategory: UBLTaxCategory{
				ID:                 "E", // Exempt
				Percent:            nil,
				TaxExemptionReason: &exemptReason,
				TaxScheme: UBLTaxScheme{
					ID: "VAT",
				},
			},
		})
	}

	// Taxable categories
	for rate, taxAmount := range taxByRate {
		rateValue := rate
		subtotals = append(subtotals, UBLTaxSubtotal{
			TaxableAmount: UBLAmount{CurrencyID: "EUR", Value: subtotalByRate[rate]},
			TaxAmount:     UBLAmount{CurrencyID: "EUR", Value: taxAmount},
			TaxCategory: UBLTaxCategory{
				ID:      "S", // Standard rated
				Percent: &rateValue,
				TaxScheme: UBLTaxScheme{
					ID: "VAT",
				},
			},
		})
	}

	return UBLTaxTotal{
		TaxAmount:    UBLAmount{CurrencyID: "EUR", Value: invoice.TaxAmount},
		TaxSubtotals: subtotals,
	}
}

func (s *XRechnungService) buildMonetaryTotal(invoice *entities.Invoice) UBLMonetaryTotal {
	return UBLMonetaryTotal{
		LineExtensionAmount: UBLAmount{CurrencyID: "EUR", Value: invoice.SumAmount},
		TaxExclusiveAmount:  UBLAmount{CurrencyID: "EUR", Value: invoice.SumAmount},
		TaxInclusiveAmount:  UBLAmount{CurrencyID: "EUR", Value: invoice.TotalAmount},
		PayableAmount:       UBLAmount{CurrencyID: "EUR", Value: invoice.TotalAmount},
	}
}

func (s *XRechnungService) buildInvoiceLines(invoice *entities.Invoice) []UBLInvoiceLine {
	lines := make([]UBLInvoiceLine, 0, len(invoice.InvoiceItems))

	for i, item := range invoice.InvoiceItems {
		line := UBLInvoiceLine{
			ID: fmt.Sprintf("%d", i+1),
			InvoicedQuantity: UBLQuantity{
				UnitCode: s.getUnitCode(&item),
				Value:    item.NumberUnits,
			},
			LineExtensionAmount: UBLAmount{
				CurrencyID: "EUR",
				Value:      item.TotalAmount,
			},
			Item: UBLItem{
				Description:           item.Description,
				Name:                  s.getItemName(&item),
				ClassifiedTaxCategory: s.getItemTaxCategory(&item),
			},
			Price: UBLPrice{
				PriceAmount: UBLAmount{
					CurrencyID: "EUR",
					Value:      item.UnitPrice,
				},
			},
		}

		lines = append(lines, line)
	}

	return lines
}

func (s *XRechnungService) getUnitCode(item *entities.InvoiceItem) string {
	// UN/CECE Recommendation 20 unit codes
	switch item.ItemType {
	case "session":
		return "HUR" // Hour
	case "extra_effort":
		return "C62" // Unit (piece)
	case "custom":
		return "C62" // Unit (piece)
	default:
		return "C62" // Default: unit
	}
}

func (s *XRechnungService) getItemName(item *entities.InvoiceItem) string {
	switch item.ItemType {
	case "session":
		return "Therapiesitzung"
	case "extra_effort":
		return "Zusatzaufwand"
	case "custom":
		return "Position"
	default:
		return "Leistung"
	}
}

func (s *XRechnungService) getItemTaxCategory(item *entities.InvoiceItem) UBLTaxCategory {
	if item.VATExempt {
		exemptionReason := item.VATExemptionText
		if exemptionReason == "" {
			exemptionReason = "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG"
		}

		return UBLTaxCategory{
			ID:                 "E", // Exempt
			Percent:            nil,
			TaxExemptionReason: &exemptionReason,
			TaxScheme: UBLTaxScheme{
				ID: "VAT",
			},
		}
	}

	rate := item.VATRate
	return UBLTaxCategory{
		ID:      "S", // Standard rated
		Percent: &rate,
		TaxScheme: UBLTaxScheme{
			ID: "VAT",
		},
	}
}
