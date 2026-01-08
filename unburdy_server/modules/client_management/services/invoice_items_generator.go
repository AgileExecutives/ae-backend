package services

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"gorm.io/datatypes"
)

// OrganizationBillingConfig contains the billing configuration from an organization
type OrganizationBillingConfig struct {
	BillingMode    string
	Config         datatypes.JSON
	UnitPrice      float64
	SingleUnitText string
	DoubleUnitText string
}

// InvoiceLineItem represents a calculated line item for an invoice
type InvoiceLineItem struct {
	ItemType        string  // "session", "extra_effort", "preparation"
	Description     string  // German text describing the item
	NumberUnits     int     // Number of units (1 or 2 for sessions)
	UnitPrice       float64 // Price per unit
	TotalAmount     float64 // Total for this line item
	UnitDurationMin int     // Duration in minutes (for reference)
	IsEditable      bool    // Whether this can be edited in draft mode

	// Source references
	SessionIDs     []uint // Sessions included in this item
	ExtraEffortIDs []uint // Extra efforts included in this item
}

// InvoiceItemsResult contains the calculated items and totals
type InvoiceItemsResult struct {
	LineItems  []InvoiceLineItem
	TotalUnits int
	SubTotal   float64
	TaxAmount  float64
	GrandTotal float64
}

// InvoiceItemsGenerator calculates invoice line items based on billing mode
type InvoiceItemsGenerator struct {
	billingMode string
	config      map[string]interface{}
	unitPrice   float64
	taxRate     float64

	singleUnitText string
	doubleUnitText string
}

// NewInvoiceItemsGenerator creates a new generator
func NewInvoiceItemsGenerator(config OrganizationBillingConfig, taxRate float64) (*InvoiceItemsGenerator, error) {
	if config.UnitPrice <= 0 {
		return nil, fmt.Errorf("organization must have a valid unit price")
	}

	// Parse config JSONB
	configMap := make(map[string]interface{})
	if len(config.Config) > 0 {
		if err := json.Unmarshal(config.Config, &configMap); err != nil {
			return nil, fmt.Errorf("failed to parse extra efforts config: %w", err)
		}
	}

	return &InvoiceItemsGenerator{
		billingMode:    config.BillingMode,
		config:         configMap,
		unitPrice:      config.UnitPrice,
		taxRate:        taxRate,
		singleUnitText: config.SingleUnitText,
		doubleUnitText: config.DoubleUnitText,
	}, nil
}

// GenerateItems calculates all invoice line items
func (g *InvoiceItemsGenerator) GenerateItems(sessions []entities.Session, extraEfforts []entities.ExtraEffort) (*InvoiceItemsResult, error) {
	switch g.billingMode {
	case "ignore":
		return g.generateIgnoreMode(sessions)
	case "bundle_double_units":
		return g.generateBundleDoubleUnits(sessions, extraEfforts)
	case "separate_items":
		return g.generateSeparateItems(sessions, extraEfforts)
	case "preparation_allowance":
		return g.generatePreparationAllowance(sessions, extraEfforts)
	default:
		return g.generateIgnoreMode(sessions) // Default to ignore mode
	}
}

// Mode A: Ignore - Only bill sessions, track efforts for reference
func (g *InvoiceItemsGenerator) generateIgnoreMode(sessions []entities.Session) (*InvoiceItemsResult, error) {
	items := make([]InvoiceLineItem, 0)
	totalUnits := 0

	for _, session := range sessions {
		units := session.NumberUnits
		totalUnits += units

		description := g.singleUnitText
		if units == 2 {
			description = g.doubleUnitText
		}

		items = append(items, InvoiceLineItem{
			ItemType:        "session",
			Description:     description,
			NumberUnits:     units,
			UnitPrice:       g.unitPrice,
			TotalAmount:     float64(units) * g.unitPrice,
			UnitDurationMin: session.DurationMin,
			IsEditable:      true,
			SessionIDs:      []uint{session.ID},
			ExtraEffortIDs:  []uint{},
		})
	}

	return g.calculateTotals(items, totalUnits), nil
}

// Mode B: Bundle Double Units - Combine session + efforts, threshold determines 1 or 2 units
func (g *InvoiceItemsGenerator) generateBundleDoubleUnits(sessions []entities.Session, extraEfforts []entities.ExtraEffort) (*InvoiceItemsResult, error) {
	// Get threshold from config (default 90 minutes)
	thresholdMin := 90
	if val, ok := g.config["threshold_minutes"].(float64); ok {
		thresholdMin = int(val)
	}

	// Group efforts by session
	effortsBySession := make(map[uint][]entities.ExtraEffort)
	for _, effort := range extraEfforts {
		if effort.SessionID != nil {
			effortsBySession[*effort.SessionID] = append(effortsBySession[*effort.SessionID], effort)
		}
	}

	items := make([]InvoiceLineItem, 0)
	totalUnits := 0

	for _, session := range sessions {
		sessionDuration := session.DurationMin
		effortIDs := make([]uint, 0)

		// Add efforts linked to this session
		if sessionEfforts, exists := effortsBySession[session.ID]; exists {
			for _, effort := range sessionEfforts {
				sessionDuration += effort.DurationMin
				effortIDs = append(effortIDs, effort.ID)
			}
		}

		// Determine units based on total duration
		units := 1
		description := g.singleUnitText
		if sessionDuration >= thresholdMin {
			units = 2
			description = g.doubleUnitText
		}

		totalUnits += units

		items = append(items, InvoiceLineItem{
			ItemType:        "session",
			Description:     description,
			NumberUnits:     units,
			UnitPrice:       g.unitPrice,
			TotalAmount:     float64(units) * g.unitPrice,
			UnitDurationMin: sessionDuration,
			IsEditable:      true,
			SessionIDs:      []uint{session.ID},
			ExtraEffortIDs:  effortIDs,
		})
	}

	return g.calculateTotals(items, totalUnits), nil
}

// Mode C: Separate Items - Bill efforts as individual line items
func (g *InvoiceItemsGenerator) generateSeparateItems(sessions []entities.Session, extraEfforts []entities.ExtraEffort) (*InvoiceItemsResult, error) {
	// Get rounding mode from config
	roundingMode := "none"
	if val, ok := g.config["rounding_mode"].(string); ok {
		roundingMode = val
	}

	items := make([]InvoiceLineItem, 0)
	totalUnits := 0

	// Add session items
	for _, session := range sessions {
		units := session.NumberUnits
		totalUnits += units

		description := g.singleUnitText
		if units == 2 {
			description = g.doubleUnitText
		}

		items = append(items, InvoiceLineItem{
			ItemType:        "session",
			Description:     description,
			NumberUnits:     units,
			UnitPrice:       g.unitPrice,
			TotalAmount:     float64(units) * g.unitPrice,
			UnitDurationMin: session.DurationMin,
			IsEditable:      true,
			SessionIDs:      []uint{session.ID},
			ExtraEffortIDs:  []uint{},
		})
	}

	// Add separate items for each effort
	for _, effort := range extraEfforts {
		duration := effort.DurationMin

		// Apply rounding
		if roundingMode == "nearest_quarter_hour" {
			duration = roundToQuarterHour(duration)
		} else if roundingMode == "nearest_half_hour" {
			duration = roundToHalfHour(duration)
		}

		// Calculate units (duration / 60, minimum 0.25)
		units := math.Max(float64(duration)/60.0, 0.25)

		description := getEffortTypeGerman(effort.EffortType)
		if effort.Description != "" {
			description = effort.Description
		}

		totalAmount := units * g.unitPrice

		items = append(items, InvoiceLineItem{
			ItemType:        "extra_effort",
			Description:     description,
			NumberUnits:     0, // Fractional, shown in description
			UnitPrice:       g.unitPrice,
			TotalAmount:     totalAmount,
			UnitDurationMin: duration,
			IsEditable:      true,
			SessionIDs:      []uint{},
			ExtraEffortIDs:  []uint{effort.ID},
		})
	}

	return g.calculateTotals(items, totalUnits), nil
}

// Mode D: Preparation Allowance - Auto-add prep time to each session
func (g *InvoiceItemsGenerator) generatePreparationAllowance(sessions []entities.Session, extraEfforts []entities.ExtraEffort) (*InvoiceItemsResult, error) {
	// Get auto-add minutes from config (default 15)
	autoAddMin := 15
	if val, ok := g.config["auto_add_minutes"].(float64); ok {
		autoAddMin = int(val)
	}

	// Get description from config
	prepDescription := "Vor- und Nachbereitung"
	if val, ok := g.config["description"].(string); ok && val != "" {
		prepDescription = val
	}

	items := make([]InvoiceLineItem, 0)
	totalUnits := 0

	for _, session := range sessions {
		// Session item
		sessionUnits := session.NumberUnits
		totalUnits += sessionUnits

		sessionDesc := g.singleUnitText
		if sessionUnits == 2 {
			sessionDesc = g.doubleUnitText
		}

		items = append(items, InvoiceLineItem{
			ItemType:        "session",
			Description:     sessionDesc,
			NumberUnits:     sessionUnits,
			UnitPrice:       g.unitPrice,
			TotalAmount:     float64(sessionUnits) * g.unitPrice,
			UnitDurationMin: session.DurationMin,
			IsEditable:      true,
			SessionIDs:      []uint{session.ID},
			ExtraEffortIDs:  []uint{},
		})

		// Automatic preparation item
		prepUnits := float64(autoAddMin) / 60.0
		prepAmount := prepUnits * g.unitPrice

		items = append(items, InvoiceLineItem{
			ItemType:        "preparation",
			Description:     prepDescription,
			NumberUnits:     0, // Fractional
			UnitPrice:       g.unitPrice,
			TotalAmount:     prepAmount,
			UnitDurationMin: autoAddMin,
			IsEditable:      false, // Auto-generated, not editable
			SessionIDs:      []uint{session.ID},
			ExtraEffortIDs:  []uint{},
		})
	}

	return g.calculateTotals(items, totalUnits), nil
}

// Helper: Calculate totals
func (g *InvoiceItemsGenerator) calculateTotals(items []InvoiceLineItem, totalUnits int) *InvoiceItemsResult {
	subTotal := 0.0
	for _, item := range items {
		subTotal += item.TotalAmount
	}

	taxAmount := subTotal * (g.taxRate / 100.0)
	grandTotal := subTotal + taxAmount

	return &InvoiceItemsResult{
		LineItems:  items,
		TotalUnits: totalUnits,
		SubTotal:   subTotal,
		TaxAmount:  taxAmount,
		GrandTotal: grandTotal,
	}
}

// Helper functions
func roundToQuarterHour(minutes int) int {
	return int(math.Round(float64(minutes)/15.0) * 15)
}

func roundToHalfHour(minutes int) int {
	return int(math.Round(float64(minutes)/30.0) * 30)
}
