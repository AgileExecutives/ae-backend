package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	baseAPI "github.com/ae-base-server/api"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"github.com/unburdy/unburdy-server-api/modules/client_management/repositories"
	"gorm.io/gorm"
)

// ExtraEffortService handles business logic for extra efforts
type ExtraEffortService struct {
	repo *repositories.ExtraEffortRepository
	db   *gorm.DB
}

// NewExtraEffortService creates a new ExtraEffortService
func NewExtraEffortService(db *gorm.DB) *ExtraEffortService {
	return &ExtraEffortService{
		repo: repositories.NewExtraEffortRepository(db),
		db:   db,
	}
}

// CreateExtraEffort creates a new extra effort
func (s *ExtraEffortService) CreateExtraEffort(req *entities.CreateExtraEffortRequest, tenantID, userID uint) (*entities.ExtraEffort, error) {
	effortDate, err := time.Parse("2006-01-02", req.EffortDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effort_date format: %w", err)
	}

	billable := true
	if req.Billable != nil {
		billable = *req.Billable
	}

	effort := &entities.ExtraEffort{
		TenantID:      tenantID,
		ClientID:      req.ClientID,
		SessionID:     req.SessionID,
		EffortType:    req.EffortType,
		EffortDate:    effortDate,
		DurationMin:   req.DurationMin,
		Description:   req.Description,
		Billable:      billable,
		BillingStatus: "unbilled",
		CreatedBy:     userID,
	}

	if err := s.repo.Create(effort); err != nil {
		return nil, fmt.Errorf("failed to create extra effort: %w", err)
	}

	return effort, nil
}

// GetExtraEffort retrieves an extra effort by ID
func (s *ExtraEffortService) GetExtraEffort(id, tenantID uint) (*entities.ExtraEffort, error) {
	return s.repo.GetByID(id, tenantID)
}

// ListExtraEfforts retrieves extra efforts with filters
func (s *ExtraEffortService) ListExtraEfforts(tenantID uint, filters map[string]interface{}) ([]entities.ExtraEffort, int64, error) {
	efforts, err := s.repo.List(tenantID, filters)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.repo.Count(tenantID, filters)
	if err != nil {
		return efforts, 0, err
	}

	return efforts, count, nil
}

// UpdateExtraEffort updates an existing extra effort
func (s *ExtraEffortService) UpdateExtraEffort(id, tenantID uint, req *entities.UpdateExtraEffortRequest) error {
	effort, err := s.repo.GetByID(id, tenantID)
	if err != nil {
		return fmt.Errorf("extra effort not found: %w", err)
	}

	// Only allow updates if unbilled
	if effort.BillingStatus != "unbilled" {
		return errors.New("cannot update billed extra efforts")
	}

	// Apply updates
	if req.EffortType != nil {
		effort.EffortType = *req.EffortType
	}
	if req.EffortDate != nil {
		effortDate, err := time.Parse("2006-01-02", *req.EffortDate)
		if err != nil {
			return fmt.Errorf("invalid effort_date format: %w", err)
		}
		effort.EffortDate = effortDate
	}
	if req.DurationMin != nil {
		effort.DurationMin = *req.DurationMin
	}
	if req.Description != nil {
		effort.Description = *req.Description
	}
	if req.Billable != nil {
		effort.Billable = *req.Billable
	}

	return s.repo.Update(effort)
}

// DeleteExtraEffort deletes an extra effort
func (s *ExtraEffortService) DeleteExtraEffort(id, tenantID uint) error {
	effort, err := s.repo.GetByID(id, tenantID)
	if err != nil {
		return fmt.Errorf("extra effort not found: %w", err)
	}

	// Only allow deletion if unbilled
	if effort.BillingStatus != "unbilled" {
		return errors.New("cannot delete billed extra efforts")
	}

	return s.repo.Delete(id, tenantID)
}

// GetUnbilledEffortsByClient retrieves unbilled extra efforts for a client
func (s *ExtraEffortService) GetUnbilledEffortsByClient(clientID, tenantID uint) ([]entities.ExtraEffort, error) {
	return s.repo.GetUnbilledByClient(clientID, tenantID)
}

// BillingCalculator calculates invoice units based on organization billing mode
type BillingCalculator struct {
	mode   string
	config map[string]interface{}
	org    *baseAPI.Organization
}

// BillingResult contains the result of billing calculation
type BillingResult struct {
	Units            float64
	Description      string
	BundledEffortIDs []uint // IDs of extra efforts bundled into this line item
}

// NewBillingCalculator creates a new billing calculator
func NewBillingCalculator(org *baseAPI.Organization) (*BillingCalculator, error) {
	var config map[string]interface{}
	if org.ExtraEffortsConfig != nil {
		if err := json.Unmarshal(org.ExtraEffortsConfig, &config); err != nil {
			return nil, fmt.Errorf("failed to parse extra efforts config: %w", err)
		}
	}

	return &BillingCalculator{
		mode:   org.ExtraEffortsBillingMode,
		config: config,
		org:    org,
	}, nil
}

// CalculateSessionUnits calculates billing units for a session including its extra efforts
func (c *BillingCalculator) CalculateSessionUnits(session *entities.Session, sessionEfforts []entities.ExtraEffort) *BillingResult {
	switch c.mode {
	case "ignore":
		// Track but don't bill extra efforts
		return &BillingResult{
			Units:       float64(session.NumberUnits),
			Description: c.getDescriptionForUnits(session.NumberUnits),
		}

	case "bundle_double_units":
		return c.calculateBundleDoubleUnits(session, sessionEfforts)

	case "separate_items":
		// Extra efforts will be billed separately, just return session units
		return &BillingResult{
			Units:       float64(session.NumberUnits),
			Description: c.getDescriptionForUnits(session.NumberUnits),
		}

	case "preparation_allowance":
		// Preparation will be added as separate item, just return session units
		return &BillingResult{
			Units:       float64(session.NumberUnits),
			Description: c.getDescriptionForUnits(session.NumberUnits),
		}

	default:
		// Default to session units only
		return &BillingResult{
			Units:       float64(session.NumberUnits),
			Description: c.getDescriptionForUnits(session.NumberUnits),
		}
	}
}

// calculateBundleDoubleUnits implements mode B: bundle into double units
func (c *BillingCalculator) calculateBundleDoubleUnits(session *entities.Session, sessionEfforts []entities.ExtraEffort) *BillingResult {
	unitDurationMin := 45 // default
	if val, ok := c.config["unit_duration_min"].(float64); ok {
		unitDurationMin = int(val)
	}

	thresholdPercentage := 90.0 // default
	if val, ok := c.config["threshold_percentage"].(float64); ok {
		thresholdPercentage = val
	}

	// Calculate total duration
	totalMin := session.DurationMin
	var bundledEffortIDs []uint
	for _, effort := range sessionEfforts {
		totalMin += effort.DurationMin
		bundledEffortIDs = append(bundledEffortIDs, effort.ID)
	}

	// Check if total meets threshold for 2 units
	twoUnitThreshold := float64(unitDurationMin*2) * thresholdPercentage / 100.0
	units := 1
	if float64(totalMin) >= twoUnitThreshold {
		units = 2
	}

	return &BillingResult{
		Units:            float64(units),
		Description:      c.getDescriptionForUnits(units),
		BundledEffortIDs: bundledEffortIDs,
	}
}

// CalculateSeparateEffortUnits calculates units for standalone extra effort (mode C)
func (c *BillingCalculator) CalculateSeparateEffortUnits(effort *entities.ExtraEffort) *BillingResult {
	if c.mode != "separate_items" {
		return nil
	}

	roundToMin := 15 // default
	if val, ok := c.config["round_to_min"].(float64); ok {
		roundToMin = int(val)
	}

	minimumDurationMin := 10 // default
	if val, ok := c.config["minimum_duration_min"].(float64); ok {
		minimumDurationMin = int(val)
	}

	// Round duration
	roundedMin := roundDuration(effort.DurationMin, roundToMin)

	// Check minimum
	if roundedMin < minimumDurationMin {
		return nil // Below minimum, don't bill
	}

	// Calculate units (assuming 45 min = 1 unit)
	units := float64(roundedMin) / 45.0

	description := fmt.Sprintf("%s - %d min", getEffortTypeGerman(effort.EffortType), roundedMin)
	if effort.Description != "" {
		description = effort.Description
	}

	return &BillingResult{
		Units:       units,
		Description: description,
	}
}

// CalculatePreparationAllowance calculates automatic preparation time (mode D)
func (c *BillingCalculator) CalculatePreparationAllowance(sessionUnits int) *BillingResult {
	if c.mode != "preparation_allowance" {
		return nil
	}

	minutesPerUnit := 15 // default
	if val, ok := c.config["minutes_per_unit"].(float64); ok {
		minutesPerUnit = int(val)
	}

	billingMode := "automatic" // default
	if val, ok := c.config["billing_mode"].(string); ok {
		billingMode = val
	}

	if billingMode != "automatic" {
		return nil // track_actual mode requires actual effort data
	}

	// Calculate preparation time
	prepMin := sessionUnits * minutesPerUnit
	prepUnits := float64(prepMin) / 45.0

	return &BillingResult{
		Units:       prepUnits,
		Description: fmt.Sprintf("Vorbereitung - %d min", prepMin),
	}
}

// getDescriptionForUnits returns the configured description for number of units
func (c *BillingCalculator) getDescriptionForUnits(units int) string {
	if units == 1 {
		if c.org.LineItemSingleUnitText != "" {
			return c.org.LineItemSingleUnitText
		}
		return "Einzelstunde"
	} else if units == 2 {
		if c.org.LineItemDoubleUnitText != "" {
			return c.org.LineItemDoubleUnitText
		}
		return "Doppelstunde"
	}
	return fmt.Sprintf("%d Einheiten", units)
}

// Helper functions

func roundDuration(duration, roundTo int) int {
	return int(math.Round(float64(duration)/float64(roundTo))) * roundTo
}

func getEffortTypeGerman(effortType string) string {
	switch effortType {
	case "preparation":
		return "Vorbereitung"
	case "consultation":
		return "Beratung"
	case "parent_meeting":
		return "Elterngespräch"
	case "documentation":
		return "Dokumentation"
	case "other":
		return "Sonstiges"
	default:
		return effortType
	}
}
