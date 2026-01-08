package services

import (
	"testing"

	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"gorm.io/datatypes"
)

func TestInvoiceItemsGenerator_ModeA_Ignore(t *testing.T) {
	config := OrganizationBillingConfig{
		BillingMode:    "ignore",
		Config:         datatypes.JSON(`{}`),
		UnitPrice:      100.0,
		SingleUnitText: "Einzelstunde",
		DoubleUnitText: "Doppelstunde",
	}

	generator, err := NewInvoiceItemsGenerator(config, 19.0)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	sessions := []entities.Session{
		{ID: 1, NumberUnits: 1, DurationMin: 60},
		{ID: 2, NumberUnits: 2, DurationMin: 90},
		{ID: 3, NumberUnits: 1, DurationMin: 50},
	}

	extraEfforts := []entities.ExtraEffort{
		{ID: 1, SessionID: uintPtr(1), DurationMin: 15, EffortType: "phone_call"},
		{ID: 2, SessionID: uintPtr(2), DurationMin: 30, EffortType: "documentation"},
	}

	result, err := generator.GenerateItems(sessions, extraEfforts)
	if err != nil {
		t.Fatalf("Failed to generate items: %v", err)
	}

	if len(result.LineItems) != 3 {
		t.Errorf("Expected 3 line items, got %d", len(result.LineItems))
	}

	if result.TotalUnits != 4 {
		t.Errorf("Expected 4 total units, got %d", result.TotalUnits)
	}

	expectedSubTotal := 400.0
	if result.SubTotal != expectedSubTotal {
		t.Errorf("Expected subtotal %.2f, got %.2f", expectedSubTotal, result.SubTotal)
	}

	expectedTax := 76.0
	if result.TaxAmount != expectedTax {
		t.Errorf("Expected tax %.2f, got %.2f", expectedTax, result.TaxAmount)
	}

	for _, item := range result.LineItems {
		if len(item.ExtraEffortIDs) > 0 {
			t.Error("Extra efforts should not be included in ignore mode")
		}
	}
}

func TestInvoiceItemsGenerator_ModeB_BundleDoubleUnits(t *testing.T) {
	config := OrganizationBillingConfig{
		BillingMode:    "bundle_double_units",
		Config:         datatypes.JSON(`{"threshold_minutes": 75}`),
		UnitPrice:      100.0,
		SingleUnitText: "Einzelstunde",
		DoubleUnitText: "Doppelstunde",
	}

	generator, err := NewInvoiceItemsGenerator(config, 19.0)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	sessions := []entities.Session{
		{ID: 1, NumberUnits: 1, DurationMin: 50},
		{ID: 2, NumberUnits: 1, DurationMin: 60},
	}

	extraEfforts := []entities.ExtraEffort{
		{ID: 1, SessionID: uintPtr(1), DurationMin: 20, EffortType: "phone_call"},
		{ID: 2, SessionID: uintPtr(2), DurationMin: 20, EffortType: "email_correspondence"},
	}

	result, err := generator.GenerateItems(sessions, extraEfforts)
	if err != nil {
		t.Fatalf("Failed to generate items: %v", err)
	}

	if len(result.LineItems) != 2 {
		t.Errorf("Expected 2 line items, got %d", len(result.LineItems))
	}

	if result.LineItems[0].NumberUnits != 1 {
		t.Errorf("Session 1 should be 1 unit, got %d", result.LineItems[0].NumberUnits)
	}
	if result.LineItems[0].Description != "Einzelstunde" {
		t.Errorf("Session 1 should be Einzelstunde, got %s", result.LineItems[0].Description)
	}
	if result.LineItems[0].UnitDurationMin != 70 {
		t.Errorf("Session 1 total duration should be 70min, got %d", result.LineItems[0].UnitDurationMin)
	}

	if result.LineItems[1].NumberUnits != 2 {
		t.Errorf("Session 2 should be 2 units, got %d", result.LineItems[1].NumberUnits)
	}
	if result.LineItems[1].Description != "Doppelstunde" {
		t.Errorf("Session 2 should be Doppelstunde, got %s", result.LineItems[1].Description)
	}

	expectedTotal := 3
	if result.TotalUnits != expectedTotal {
		t.Errorf("Expected %d total units, got %d", expectedTotal, result.TotalUnits)
	}
}

func TestInvoiceItemsGenerator_ModeC_SeparateItems(t *testing.T) {
	config := OrganizationBillingConfig{
		BillingMode:    "separate_items",
		Config:         datatypes.JSON(`{"rounding_mode": "nearest_quarter_hour"}`),
		UnitPrice:      100.0,
		SingleUnitText: "Einzelstunde",
		DoubleUnitText: "Doppelstunde",
	}

	generator, err := NewInvoiceItemsGenerator(config, 19.0)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	sessions := []entities.Session{
		{ID: 1, NumberUnits: 1, DurationMin: 60},
	}

	extraEfforts := []entities.ExtraEffort{
		{ID: 1, DurationMin: 17, EffortType: "phone_call", Description: "Telefonat mit Kostenträger"},
		{ID: 2, DurationMin: 28, EffortType: "email_correspondence", Description: ""},
	}

	result, err := generator.GenerateItems(sessions, extraEfforts)
	if err != nil {
		t.Fatalf("Failed to generate items: %v", err)
	}

	if len(result.LineItems) != 3 {
		t.Errorf("Expected 3 line items, got %d", len(result.LineItems))
	}

	if result.LineItems[0].ItemType != "session" {
		t.Errorf("First item should be session, got %s", result.LineItems[0].ItemType)
	}
	if result.LineItems[0].TotalAmount != 100.0 {
		t.Errorf("Session should cost 100, got %.2f", result.LineItems[0].TotalAmount)
	}

	if result.LineItems[1].ItemType != "extra_effort" {
		t.Errorf("Second item should be extra_effort, got %s", result.LineItems[1].ItemType)
	}
	if result.LineItems[1].UnitDurationMin != 15 {
		t.Errorf("First effort should be rounded to 15min, got %d", result.LineItems[1].UnitDurationMin)
	}
	expectedAmount1 := 25.0
	if result.LineItems[1].TotalAmount != expectedAmount1 {
		t.Errorf("First effort should cost %.2f, got %.2f", expectedAmount1, result.LineItems[1].TotalAmount)
	}

	if result.LineItems[2].UnitDurationMin != 30 {
		t.Errorf("Second effort should be rounded to 30min, got %d", result.LineItems[2].UnitDurationMin)
	}
	expectedAmount2 := 50.0
	if result.LineItems[2].TotalAmount != expectedAmount2 {
		t.Errorf("Second effort should cost %.2f, got %.2f", expectedAmount2, result.LineItems[2].TotalAmount)
	}

	expectedSubTotal := 175.0
	if result.SubTotal != expectedSubTotal {
		t.Errorf("Expected subtotal %.2f, got %.2f", expectedSubTotal, result.SubTotal)
	}
}

func TestInvoiceItemsGenerator_ModeD_PreparationAllowance(t *testing.T) {
	config := OrganizationBillingConfig{
		BillingMode:    "preparation_allowance",
		Config:         datatypes.JSON(`{"auto_add_minutes": 15, "description": "Vor- und Nachbereitung"}`),
		UnitPrice:      100.0,
		SingleUnitText: "Einzelstunde",
		DoubleUnitText: "Doppelstunde",
	}

	generator, err := NewInvoiceItemsGenerator(config, 19.0)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	sessions := []entities.Session{
		{ID: 1, NumberUnits: 1, DurationMin: 60},
		{ID: 2, NumberUnits: 2, DurationMin: 90},
	}

	extraEfforts := []entities.ExtraEffort{}

	result, err := generator.GenerateItems(sessions, extraEfforts)
	if err != nil {
		t.Fatalf("Failed to generate items: %v", err)
	}

	if len(result.LineItems) != 4 {
		t.Errorf("Expected 4 line items, got %d", len(result.LineItems))
	}

	if result.LineItems[0].ItemType != "session" {
		t.Error("First item should be session")
	}
	if result.LineItems[1].ItemType != "preparation" {
		t.Error("Second item should be preparation")
	}
	if result.LineItems[2].ItemType != "session" {
		t.Error("Third item should be session")
	}
	if result.LineItems[3].ItemType != "preparation" {
		t.Error("Fourth item should be preparation")
	}

	prepItem := result.LineItems[1]
	if prepItem.Description != "Vor- und Nachbereitung" {
		t.Errorf("Prep description should be 'Vor- und Nachbereitung', got %s", prepItem.Description)
	}
	if prepItem.UnitDurationMin != 15 {
		t.Errorf("Prep duration should be 15min, got %d", prepItem.UnitDurationMin)
	}
	expectedPrepAmount := 25.0
	if prepItem.TotalAmount != expectedPrepAmount {
		t.Errorf("Prep amount should be %.2f, got %.2f", expectedPrepAmount, prepItem.TotalAmount)
	}
	if prepItem.IsEditable {
		t.Error("Prep items should not be editable")
	}

	expectedSubTotal := 350.0
	if result.SubTotal != expectedSubTotal {
		t.Errorf("Expected subtotal %.2f, got %.2f", expectedSubTotal, result.SubTotal)
	}
}

func TestInvoiceItemsGenerator_NoUnitPrice(t *testing.T) {
	config := OrganizationBillingConfig{
		BillingMode:    "ignore",
		Config:         datatypes.JSON(`{}`),
		UnitPrice:      0,
		SingleUnitText: "Einzelstunde",
		DoubleUnitText: "Doppelstunde",
	}

	_, err := NewInvoiceItemsGenerator(config, 19.0)
	if err == nil {
		t.Error("Expected error when unit price is 0")
	}
}

func TestInvoiceItemsGenerator_InvalidConfig(t *testing.T) {
	config := OrganizationBillingConfig{
		BillingMode:    "ignore",
		Config:         datatypes.JSON(`{invalid json`),
		UnitPrice:      100.0,
		SingleUnitText: "Einzelstunde",
		DoubleUnitText: "Doppelstunde",
	}

	_, err := NewInvoiceItemsGenerator(config, 19.0)
	if err == nil {
		t.Error("Expected error when config JSON is invalid")
	}
}

func TestRoundingFunctions(t *testing.T) {
	tests := []struct {
		input       int
		quarterHour int
		halfHour    int
	}{
		{0, 0, 0},
		{7, 0, 0},
		{8, 15, 0},
		{12, 15, 0},
		{17, 15, 30}, // rounds to 30 for half hour
		{22, 15, 30}, // rounds to 15 for quarter hour
		{28, 30, 30},
		{37, 30, 30}, // rounds to 30 for quarter hour
		{43, 45, 30}, // rounds to 30 for half hour
		{52, 45, 60}, // rounds to 45 for quarter hour
		{67, 60, 60},
	}

	for _, tt := range tests {
		qh := roundToQuarterHour(tt.input)
		if qh != tt.quarterHour {
			t.Errorf("roundToQuarterHour(%d) = %d, want %d", tt.input, qh, tt.quarterHour)
		}

		hh := roundToHalfHour(tt.input)
		if hh != tt.halfHour {
			t.Errorf("roundToHalfHour(%d) = %d, want %d", tt.input, hh, tt.halfHour)
		}
	}
}

func TestGetEffortTypeGerman(t *testing.T) {
	tests := []struct {
		effortType string
		expected   string
	}{
		{"phone_call", "phone_call"},
		{"email_correspondence", "email_correspondence"},
		{"report_writing", "report_writing"},
		{"consultation", "Beratung"},
		{"meeting", "meeting"},
		{"documentation", "Dokumentation"},
		{"preparation", "Vorbereitung"},
		{"parent_meeting", "Elterngespräch"},
		{"other", "Sonstiges"},
		{"unknown_type", "unknown_type"},
	}

	for _, tt := range tests {
		result := getEffortTypeGerman(tt.effortType)
		if result != tt.expected {
			t.Errorf("getEffortTypeGerman(%s) = %s, want %s", tt.effortType, result, tt.expected)
		}
	}
}

func uintPtr(u uint) *uint {
	return &u
}
