package api

import (
	"github.com/unburdy/invoice-number-module/services"
)

// InvoiceNumberService exports the invoice number service for use by external modules
type InvoiceNumberService = services.InvoiceNumberService

// NewInvoiceNumberService creates a new invoice number service
// This is exposed for external modules that need invoice number generation
var NewInvoiceNumberService = services.NewInvoiceNumberService
