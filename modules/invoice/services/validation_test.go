package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/invoice-module/entities"
)

func TestValidateVAT(t *testing.T) {
	tests := []struct {
		name        string
		invoice     *entities.Invoice
		wantErr     bool
		errContains string
	}{
		{
			name: "success - valid VAT",
			invoice: &entities.Invoice{
				Items: []entities.InvoiceItem{
					{
						VATRate:    19.0,
						VATExempt:  false,
						Amount:     100.0,
						TaxRate:    19.0,
						TaxAmount:  19.0,
						TotalPrice: 119.0,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success - VAT exempt with explanation",
			invoice: &entities.Invoice{
				Items: []entities.InvoiceItem{
					{
						VATRate:          0.0,
						VATExempt:        true,
						VATExemptionText: "Reverse charge - Article 196 EU VAT Directive",
						Amount:           100.0,
						TaxRate:          0.0,
						TaxAmount:        0.0,
						TotalPrice:       100.0,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "error - VAT exempt without explanation",
			invoice: &entities.Invoice{
				Items: []entities.InvoiceItem{
					{
						VATRate:          0.0,
						VATExempt:        true,
						VATExemptionText: "",
						Amount:           100.0,
						TaxRate:          0.0,
						TaxAmount:        0.0,
						TotalPrice:       100.0,
					},
				},
			},
			wantErr:     true,
			errContains: "VAT exempt items must include exemption text",
		},
		{
			name: "error - negative VAT rate",
			invoice: &entities.Invoice{
				Items: []entities.InvoiceItem{
					{
						VATRate:    -5.0,
						VATExempt:  false,
						Amount:     100.0,
						TaxRate:    -5.0,
						TaxAmount:  -5.0,
						TotalPrice: 95.0,
					},
				},
			},
			wantErr:     true,
			errContains: "VAT rate cannot be negative",
		},
		{
			name: "error - VAT rate too high",
			invoice: &entities.Invoice{
				Items: []entities.InvoiceItem{
					{
						VATRate:    150.0,
						VATExempt:  false,
						Amount:     100.0,
						TaxRate:    150.0,
						TaxAmount:  150.0,
						TotalPrice: 250.0,
					},
				},
			},
			wantErr:     true,
			errContains: "VAT rate cannot exceed 100%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &InvoiceService{}
			err := service.validateVAT(tt.invoice)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateInvoice(t *testing.T) {
	tests := []struct {
		name        string
		invoice     *entities.Invoice
		wantErr     bool
		errContains string
	}{
		{
			name: "success - valid invoice",
			invoice: &entities.Invoice{
				CustomerName:   "Test Customer",
				InvoiceDate:    time.Now(),
				SubtotalAmount: 100.0,
				TotalAmount:    119.0,
				Items: []entities.InvoiceItem{
					{Position: 1, Amount: 100.0, Quantity: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "error - missing customer name",
			invoice: &entities.Invoice{
				CustomerName:   "",
				InvoiceDate:    time.Now(),
				SubtotalAmount: 100.0,
				TotalAmount:    119.0,
				Items: []entities.InvoiceItem{
					{Position: 1, Amount: 100.0},
				},
			},
			wantErr:     true,
			errContains: "customer name is required",
		},
		{
			name: "error - no line items",
			invoice: &entities.Invoice{
				CustomerName:   "Test Customer",
				InvoiceDate:    time.Now(),
				SubtotalAmount: 100.0,
				TotalAmount:    119.0,
				Items:          []entities.InvoiceItem{},
			},
			wantErr:     true,
			errContains: "invoice must have at least one line item",
		},
		{
			name: "error - negative total amount",
			invoice: &entities.Invoice{
				CustomerName:   "Test Customer",
				InvoiceDate:    time.Now(),
				SubtotalAmount: -100.0,
				TotalAmount:    -119.0,
				Items: []entities.InvoiceItem{
					{Position: 1, Amount: -100.0},
				},
			},
			wantErr:     true,
			errContains: "total amount must be positive",
		},
		{
			name: "error - due date before invoice date",
			invoice: &entities.Invoice{
				CustomerName:   "Test Customer",
				InvoiceDate:    time.Now(),
				DueDate:        ptr(time.Now().Add(-24 * time.Hour)),
				SubtotalAmount: 100.0,
				TotalAmount:    119.0,
				Items: []entities.InvoiceItem{
					{Position: 1, Amount: 100.0},
				},
			},
			wantErr:     true,
			errContains: "due date cannot be before invoice date",
		},
		{
			name: "error - invalid discount rate",
			invoice: &entities.Invoice{
				CustomerName:   "Test Customer",
				InvoiceDate:    time.Now(),
				SubtotalAmount: 100.0,
				TotalAmount:    119.0,
				DiscountRate:   150.0,
				Items: []entities.InvoiceItem{
					{Position: 1, Amount: 100.0},
				},
			},
			wantErr:     true,
			errContains: "discount rate must be between 0 and 100",
		},
		{
			name: "error - performance period end before start",
			invoice: &entities.Invoice{
				CustomerName:           "Test Customer",
				InvoiceDate:            time.Now(),
				SubtotalAmount:         100.0,
				TotalAmount:            119.0,
				PerformancePeriodStart: ptr(time.Now()),
				PerformancePeriodEnd:   ptr(time.Now().Add(-24 * time.Hour)),
				Items: []entities.InvoiceItem{
					{Position: 1, Amount: 100.0},
				},
			},
			wantErr:     true,
			errContains: "performance period end cannot be before start",
		},
		{
			name: "success - zero amount with VAT exempt",
			invoice: &entities.Invoice{
				CustomerName:   "Test Customer",
				InvoiceDate:    time.Now(),
				SubtotalAmount: 0.0,
				TotalAmount:    0.0,
				Items: []entities.InvoiceItem{
					{Position: 1, Amount: 0.0, VATExempt: true, VATExemptionText: "Free sample"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &InvoiceService{}
			err := service.validateInvoice(tt.invoice)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateInvoiceItems(t *testing.T) {
	tests := []struct {
		name        string
		items       []entities.InvoiceItem
		wantErr     bool
		errContains string
	}{
		{
			name: "success - valid items",
			items: []entities.InvoiceItem{
				{Position: 1, Description: "Item 1", Quantity: 2, UnitPrice: 50.0, Amount: 100.0},
				{Position: 2, Description: "Item 2", Quantity: 1, UnitPrice: 75.0, Amount: 75.0},
			},
			wantErr: false,
		},
		{
			name: "error - missing description",
			items: []entities.InvoiceItem{
				{Position: 1, Description: "", Quantity: 1, UnitPrice: 100.0, Amount: 100.0},
			},
			wantErr:     true,
			errContains: "description is required",
		},
		{
			name: "error - zero quantity",
			items: []entities.InvoiceItem{
				{Position: 1, Description: "Item 1", Quantity: 0, UnitPrice: 100.0, Amount: 100.0},
			},
			wantErr:     true,
			errContains: "quantity must be positive",
		},
		{
			name: "error - negative unit price",
			items: []entities.InvoiceItem{
				{Position: 1, Description: "Item 1", Quantity: 1, UnitPrice: -100.0, Amount: -100.0},
			},
			wantErr:     true,
			errContains: "unit price cannot be negative",
		},
		{
			name: "error - amount calculation mismatch",
			items: []entities.InvoiceItem{
				{Position: 1, Description: "Item 1", Quantity: 2, UnitPrice: 50.0, Amount: 90.0}, // Should be 100
			},
			wantErr:     true,
			errContains: "amount does not match quantity * unit price",
		},
		{
			name: "success - discount applied correctly",
			items: []entities.InvoiceItem{
				{
					Position:     1,
					Description:  "Item 1",
					Quantity:     2,
					UnitPrice:    50.0,
					DiscountRate: 10.0,
					Amount:       90.0, // 100 - 10% = 90
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &InvoiceService{}
			invoice := &entities.Invoice{
				CustomerName:   "Test Customer",
				InvoiceDate:    time.Now(),
				SubtotalAmount: 0,
				TotalAmount:    0,
				Items:          tt.items,
			}

			// Calculate totals
			for _, item := range tt.items {
				invoice.SubtotalAmount += item.Amount
				invoice.TotalAmount += item.TotalPrice
			}

			err := service.validateInvoice(invoice)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAmountCalculations(t *testing.T) {
	tests := []struct {
		name           string
		quantity       float64
		unitPrice      float64
		discountRate   float64
		vatRate        float64
		expectedAmount float64
		expectedTax    float64
		expectedTotal  float64
	}{
		{
			name:           "simple - no discount, 19% VAT",
			quantity:       1,
			unitPrice:      100.0,
			discountRate:   0,
			vatRate:        19.0,
			expectedAmount: 100.0,
			expectedTax:    19.0,
			expectedTotal:  119.0,
		},
		{
			name:           "with discount - 10% off, 19% VAT",
			quantity:       1,
			unitPrice:      100.0,
			discountRate:   10.0,
			vatRate:        19.0,
			expectedAmount: 90.0,
			expectedTax:    17.1,
			expectedTotal:  107.1,
		},
		{
			name:           "multiple quantity - 5 items, 7% VAT",
			quantity:       5,
			unitPrice:      20.0,
			discountRate:   0,
			vatRate:        7.0,
			expectedAmount: 100.0,
			expectedTax:    7.0,
			expectedTotal:  107.0,
		},
		{
			name:           "VAT exempt",
			quantity:       1,
			unitPrice:      100.0,
			discountRate:   0,
			vatRate:        0.0,
			expectedAmount: 100.0,
			expectedTax:    0.0,
			expectedTotal:  100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate amount after discount
			baseAmount := tt.quantity * tt.unitPrice
			discountAmount := baseAmount * (tt.discountRate / 100.0)
			amount := baseAmount - discountAmount

			// Calculate tax
			tax := amount * (tt.vatRate / 100.0)

			// Calculate total
			total := amount + tax

			assert.InDelta(t, tt.expectedAmount, amount, 0.01, "Amount mismatch")
			assert.InDelta(t, tt.expectedTax, tax, 0.01, "Tax mismatch")
			assert.InDelta(t, tt.expectedTotal, total, 0.01, "Total mismatch")
		})
	}
}
