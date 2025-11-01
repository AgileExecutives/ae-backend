package seeding

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/unburdy/calendar-module/entities"
	"gorm.io/gorm"
)

// CalendarSeeder handles seeding of calendar data
type CalendarSeeder struct {
	db      *gorm.DB
	config  *SeedConfig
	clients []UnburdyClient
}

// UnburdyClient represents a client from unburdy_server seed data
type UnburdyClient struct {
	FirstName            string    `json:"first_name"`
	LastName             string    `json:"last_name"`
	Email                string    `json:"email"`
	Phone                string    `json:"phone"`
	ContactFirstName     string    `json:"contact_first_name"`
	ContactLastName      string    `json:"contact_last_name"`
	ContactEmail         string    `json:"contact_email"`
	ContactPhone         string    `json:"contact_phone"`
	AlternativeFirstName string    `json:"alternative_first_name"`
	AlternativeLastName  string    `json:"alternative_last_name"`
	AlternativeEmail     string    `json:"alternative_email"`
	AlternativePhone     string    `json:"alternative_phone"`
	TherapyTitle         string    `json:"therapy_title"`
	Status               string    `json:"status"`
	City                 string    `json:"city"`
	StreetAddress        string    `json:"street_address"`
	Zip                  string    `json:"zip"`
	DateOfBirth          time.Time `json:"date_of_birth"`
	Gender               string    `json:"gender"`
	PrimaryLanguage      string    `json:"primary_language"`
	Notes                string    `json:"notes"`
	ProviderApprovalCode string    `json:"provider_approval_code"`
	ProviderApprovalDate string    `json:"provider_approval_date"`
}

// UnburdySeedData represents the structure of unburdy_server/seed_app_data.json
type UnburdySeedData struct {
	Clients []UnburdyClient `json:"clients"`
}

// HolidayData represents the structure of holidays.json
type HolidayData struct {
	State           string                            `json:"state"`
	SchoolHolidays  map[string]map[string][2]string   `json:"school_holidays"`
	PublicHolidays  map[string]map[string]string      `json:"public_holidays"`
}

// SeedConfig represents the seeding configuration
type SeedConfig struct {
	SeedConfig         SeedConfigDetails          `json:"seed_config"`
	CalendarTemplates  []CalendarTemplate         `json:"calendar_templates"`
	RecurringTemplates []RecurringSeriesTemplate  `json:"recurring_series_templates"`
	EventTemplates     EventTemplateGroups        `json:"event_templates"`
	ExternalTemplates  []ExternalCalendarTemplate `json:"external_calendar_templates"`
	GenerationRules    GenerationRules            `json:"generation_rules"`
}

type SeedConfigDetails struct {
	Description string `json:"description"`
	TimeRange   struct {
		MonthsBack    int `json:"months_back"`
		MonthsForward int `json:"months_forward"`
	} `json:"time_range"`
	DensityProfile struct {
		Description   string `json:"description"`
		CurrentPeriod struct {
			Weeks         int `json:"weeks"`
			EventsPerWeek struct {
				Min int `json:"min"`
				Max int `json:"max"`
			} `json:"events_per_week"`
			RecurringSeriesCount struct {
				Min int `json:"min"`
				Max int `json:"max"`
			} `json:"recurring_series_count"`
		} `json:"current_period"`
		PastPeriod struct {
			EventsPerWeek        EventRange `json:"events_per_week"`
			RecurringSeriesCount EventRange `json:"recurring_series_count"`
		} `json:"past_period"`
		FuturePeriod struct {
			EventsPerWeek        EventRange `json:"events_per_week"`
			RecurringSeriesCount EventRange `json:"recurring_series_count"`
		} `json:"future_period"`
	} `json:"density_profile"`
}

type EventRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type CalendarTemplate struct {
	Title              string                 `json:"title"`
	Color              string                 `json:"color"`
	Timezone           string                 `json:"timezone"`
	WeeklyAvailability map[string]interface{} `json:"weekly_availability"`
}

type RecurringSeriesTemplate struct {
	Title              string                   `json:"title"`
	Description        string                   `json:"description"`
	Weekday            int                      `json:"weekday"`
	Interval           int                      `json:"interval"`
	TimeStart          string                   `json:"time_start"`
	TimeEnd            string                   `json:"time_end"`
	Location           string                   `json:"location"`
	Type               string                   `json:"type"`
	Participants       []map[string]interface{} `json:"participants"`
	CalendarType       string                   `json:"calendar_type"`
	ProbabilityCurrent float64                  `json:"probability_current"`
	ProbabilityPast    float64                  `json:"probability_past"`
	ProbabilityFuture  float64                  `json:"probability_future"`
}

type EventTemplateGroups struct {
	Therapy     []EventTemplate `json:"therapy"`
	Parent      []EventTemplate `json:"parent"`
	Information []EventTemplate `json:"information"`
	Generic     []EventTemplate `json:"generic"`
}

type EventTemplate struct {
	Title             string     `json:"title"`
	Type              string     `json:"type"`
	DurationMinutes   int        `json:"duration_minutes"`
	Locations         []string   `json:"locations"`
	ParticipantsCount EventRange `json:"participants_count"`
	Descriptions      []string   `json:"descriptions"`
	TimeSlots         []string   `json:"time_slots"`
}

type ExternalCalendarTemplate struct {
	Title    string                 `json:"title"`
	URL      string                 `json:"url"`
	Color    string                 `json:"color"`
	Settings map[string]interface{} `json:"settings"`
}

type GenerationRules struct {
	WorkingHours struct {
		Start      string `json:"start"`
		End        string `json:"end"`
		LunchBreak struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"lunch_break"`
	} `json:"working_hours"`
	PersonalHours struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"personal_hours"`
	AvoidConflicts         bool    `json:"avoid_conflicts"`
	MinGapBetweenEvents    int     `json:"min_gap_between_events"`
	WeekendWorkProbability float64 `json:"weekend_work_probability"`
	AllDayEventProbability float64 `json:"all_day_event_probability"`
	ExceptionProbability   float64 `json:"exception_probability"`
	NoShowProbability      float64 `json:"no_show_probability"`
}

// NewCalendarSeeder creates a new calendar seeder with embedded config and loads unburdy clients
func NewCalendarSeeder(db *gorm.DB) *CalendarSeeder {
	// Load unburdy clients from seed data file
	clients := loadUnburdyClients()

	// Use embedded default configuration for therapy appointments
	config := &SeedConfig{
		SeedConfig: SeedConfigDetails{
			Description: "Therapy appointment calendar seeding - generates data 2 months back and 6 months forward",
			TimeRange: struct {
				MonthsBack    int `json:"months_back"`
				MonthsForward int `json:"months_forward"`
			}{
				MonthsBack:    2,
				MonthsForward: 6,
			},
		},
		CalendarTemplates: []CalendarTemplate{
			{
				Title:    "Therapy Appointments",
				Color:    "#2196F3",
				Timezone: "Europe/Berlin",
				WeeklyAvailability: map[string]interface{}{
					"monday":    map[string]interface{}{"start": "08:00", "end": "18:00"},
					"tuesday":   map[string]interface{}{"start": "08:00", "end": "18:00"},
					"wednesday": map[string]interface{}{"start": "08:00", "end": "18:00"},
					"thursday":  map[string]interface{}{"start": "08:00", "end": "18:00"},
					"friday":    map[string]interface{}{"start": "08:00", "end": "17:00"},
					"saturday":  map[string]interface{}{"start": "09:00", "end": "14:00"},
					"sunday":    map[string]interface{}{"start": nil, "end": nil},
				},
			},
		},
		RecurringTemplates: []RecurringSeriesTemplate{
			{
				Title:       "Weekly Supervision",
				Description: "Weekly team supervision meeting",
				Weekday:     1,
				Interval:    1,
				TimeStart:   "09:00",
				TimeEnd:     "10:00",
				Location:    "Supervision Room",
				Type:        "supervision",
				Participants: []map[string]interface{}{
					{"name": "Supervisor", "email": "supervisor@therapy.com"},
					{"name": "Therapist", "email": "therapist@therapy.com"},
				},
				CalendarType:       "therapy",
				ProbabilityCurrent: 0.95,
				ProbabilityPast:    0.8,
				ProbabilityFuture:  0.9,
			},
		},
		EventTemplates: EventTemplateGroups{
			Therapy: []EventTemplate{
				{
					Title:             "Individual Therapy Session",
					Type:              "therapy",
					DurationMinutes:   45,
					Locations:         []string{"Therapy Room 1", "Therapy Room 2", "Online Session"},
					ParticipantsCount: EventRange{Min: 1, Max: 2},
					Descriptions:      []string{"Individual therapy session", "Therapeutic intervention"},
					TimeSlots:         []string{"08:00", "09:00", "10:00", "11:00", "13:00", "14:00", "15:00", "16:00"},
				},
			},
			Parent: []EventTemplate{
				{
					Title:             "Parent Consultation",
					Type:              "parent",
					DurationMinutes:   45,
					Locations:         []string{"Consultation Room", "Online Session"},
					ParticipantsCount: EventRange{Min: 1, Max: 2},
					Descriptions:      []string{"Parent consultation session", "Family guidance meeting"},
					TimeSlots:         []string{"08:00", "09:00", "14:00", "15:00", "16:00", "17:00"},
				},
			},
			Information: []EventTemplate{
				{
					Title:             "Information Session",
					Type:              "information",
					DurationMinutes:   45,
					Locations:         []string{"Information Room", "Online Session"},
					ParticipantsCount: EventRange{Min: 1, Max: 3},
					Descriptions:      []string{"Information and guidance session", "Initial consultation"},
					TimeSlots:         []string{"09:00", "10:00", "11:00", "13:00", "14:00", "15:00"},
				},
			},
			Generic: []EventTemplate{
				{
					Title:             "Appointment",
					Type:              "generic",
					DurationMinutes:   30,
					Locations:         []string{"Office", "Online Session"},
					ParticipantsCount: EventRange{Min: 1, Max: 2},
					Descriptions:      []string{"General appointment", "Follow-up meeting"},
					TimeSlots:         []string{"08:00", "09:00", "10:00", "11:00", "13:00", "14:00", "15:00", "16:00"},
				},
			},
		},
	}

	return &CalendarSeeder{
		db:      db,
		config:  config,
		clients: clients,
	}
}

// SeedCalendarData generates and seeds calendar data for a user
func (cs *CalendarSeeder) SeedCalendarData(tenantID, userID uint) error {
	now := time.Now()

	// Calculate time boundaries
	startDate := now.AddDate(0, -cs.config.SeedConfig.TimeRange.MonthsBack, 0)
	endDate := now.AddDate(0, cs.config.SeedConfig.TimeRange.MonthsForward, 0)

	fmt.Printf("Seeding calendar data for user %d from %s to %s\n",
		userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// Create calendars
	calendars, err := cs.createCalendars(tenantID, userID)
	if err != nil {
		return fmt.Errorf("failed to create calendars: %w", err)
	}

	// Create recurring series
	for _, calendar := range calendars {
		if err := cs.createRecurringSeries(calendar, startDate, endDate, now); err != nil {
			return fmt.Errorf("failed to create recurring series: %w", err)
		}
	}

	// Create individual events with proper density distribution
	for _, calendar := range calendars {
		if err := cs.createIndividualEvents(calendar, startDate, endDate, now); err != nil {
			return fmt.Errorf("failed to create individual events: %w", err)
		}
	}

	// Create holiday entries for each calendar
	for _, calendar := range calendars {
		fmt.Printf("Creating holidays for calendar %d (%s)\n", calendar.ID, calendar.Title)
		if err := cs.createHolidayEntries(calendar, startDate, endDate); err != nil {
			fmt.Printf("Warning: Failed to create holidays for calendar %d (%s): %v\n", calendar.ID, calendar.Title, err)
		} else {
			fmt.Printf("Successfully created holidays for calendar %d (%s)\n", calendar.ID, calendar.Title)
		}
	}

	fmt.Printf("Successfully seeded calendar data for user %d\n", userID)
	return nil
}

// createCalendars creates the base calendars from templates
func (cs *CalendarSeeder) createCalendars(tenantID, userID uint) ([]*entities.Calendar, error) {
	var calendars []*entities.Calendar

	for _, template := range cs.config.CalendarTemplates {
		availabilityJSON, err := json.Marshal(template.WeeklyAvailability)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal availability: %w", err)
		}

		calendar := &entities.Calendar{
			TenantID:           tenantID,
			UserID:             userID,
			Title:              template.Title,
			Color:              template.Color,
			WeeklyAvailability: availabilityJSON,
			CalendarUUID:       uuid.New().String(),
			Timezone:           template.Timezone,
		}

		if err := cs.db.Create(calendar).Error; err != nil {
			return nil, fmt.Errorf("failed to create calendar: %w", err)
		}

		calendars = append(calendars, calendar)
	}

	return calendars, nil
}

// createRecurringSeries creates recurring event series
func (cs *CalendarSeeder) createRecurringSeries(calendar *entities.Calendar, startDate, endDate, now time.Time) error {
	calendarType := cs.getCalendarTypeFromTitle(calendar.Title)

	// Get appropriate templates for this calendar type
	var relevantTemplates []RecurringSeriesTemplate
	for _, template := range cs.config.RecurringTemplates {
		if template.CalendarType == calendarType {
			relevantTemplates = append(relevantTemplates, template)
		}
	}

	// Create series for relevant templates
	for _, template := range relevantTemplates {
		participantsJSON, err := json.Marshal(template.Participants)
		if err != nil {
			return fmt.Errorf("failed to marshal participants: %w", err)
		}

		timeStart, err := time.Parse("15:04", template.TimeStart)
		if err != nil {
			return fmt.Errorf("failed to parse start time: %w", err)
		}

		timeEnd, err := time.Parse("15:04", template.TimeEnd)
		if err != nil {
			return fmt.Errorf("failed to parse end time: %w", err)
		}

		series := &entities.CalendarSeries{
			TenantID:     calendar.TenantID,
			UserID:       calendar.UserID,
			CalendarID:   calendar.ID,
			Title:        template.Title,
			Participants: participantsJSON,
			Weekday:      template.Weekday,
			Interval:     template.Interval,
			TimeStart:    &timeStart,
			TimeEnd:      &timeEnd,
			Description:  template.Description,
			Location:     template.Location,
			EntryUUID:    uuid.New().String(),
		}

		if err := cs.db.Create(series).Error; err != nil {
			return fmt.Errorf("failed to create series: %w", err)
		}

		// Create individual entries for this series
		if err := cs.createSeriesEntries(series, startDate, endDate, template.Type); err != nil {
			return fmt.Errorf("failed to create series entries: %w", err)
		}
	}

	return nil
}

// createSeriesEntries creates individual calendar entries for a recurring series
func (cs *CalendarSeeder) createSeriesEntries(series *entities.CalendarSeries, startDate, endDate time.Time, eventType string) error {
	current := startDate

	// Find the first occurrence of the specified weekday
	for current.Weekday() != time.Weekday(series.Weekday) {
		current = current.AddDate(0, 0, 1)
	}

	for current.Before(endDate) {
		// Create date and time
		eventDate := current

		// Combine date with time from series
		startTime := time.Date(
			eventDate.Year(), eventDate.Month(), eventDate.Day(),
			series.TimeStart.Hour(), series.TimeStart.Minute(), 0, 0,
			eventDate.Location(),
		)

		endTime := time.Date(
			eventDate.Year(), eventDate.Month(), eventDate.Day(),
			series.TimeEnd.Hour(), series.TimeEnd.Minute(), 0, 0,
			eventDate.Location(),
		)

		entry := &entities.CalendarEntry{
			TenantID:     series.TenantID,
			UserID:       series.UserID,
			CalendarID:   series.CalendarID,
			SeriesID:     &series.ID,
			Title:        series.Title,
			IsException:  false,
			Participants: series.Participants,
			DateFrom:     &eventDate,
			DateTo:       &eventDate,
			TimeFrom:     &startTime,
			TimeTo:       &endTime,
			Timezone:     "Europe/Berlin",
			Type:         eventType,
			Description:  series.Description,
			Location:     series.Location,
			IsAllDay:     false,
		}

		if err := cs.db.Create(entry).Error; err != nil {
			return fmt.Errorf("failed to create series entry: %w", err)
		}

		// Move to next occurrence (every N weeks based on interval)
		current = current.AddDate(0, 0, 7*series.Interval)
	}

	return nil
}

// createIndividualEvents creates standalone calendar events with density-based distribution
func (cs *CalendarSeeder) createIndividualEvents(calendar *entities.Calendar, startDate, endDate, now time.Time) error {
	// Get all therapy-related event templates
	var eventTemplates []EventTemplate
	eventTemplates = append(eventTemplates, cs.config.EventTemplates.Therapy...)
	eventTemplates = append(eventTemplates, cs.config.EventTemplates.Parent...)
	eventTemplates = append(eventTemplates, cs.config.EventTemplates.Information...)
	eventTemplates = append(eventTemplates, cs.config.EventTemplates.Generic...)

	if len(eventTemplates) == 0 {
		return nil // No templates to work with
	}

	// Create events with different densities for different periods

	// Past period (less crowded)
	pastEnd := now.AddDate(0, 0, -28) // 4 weeks ago
	if startDate.Before(pastEnd) {
		eventsCount := 15 // Lower density for past
		if err := cs.createTherapyAppointments(calendar, eventTemplates, startDate, pastEnd, eventsCount); err != nil {
			return fmt.Errorf("failed to create past appointments: %w", err)
		}
	}

	// Current period (busy - next 4 weeks)
	currentStart := now.AddDate(0, 0, -28)
	currentEnd := now.AddDate(0, 0, 28)
	eventsCount := 60 // High density for current period
	if err := cs.createTherapyAppointments(calendar, eventTemplates, currentStart, currentEnd, eventsCount); err != nil {
		return fmt.Errorf("failed to create current appointments: %w", err)
	}

	// Future period (moderate density)
	futureStart := now.AddDate(0, 0, 28)
	if futureStart.Before(endDate) {
		eventsCount := 35 // Moderate density for future
		if err := cs.createTherapyAppointments(calendar, eventTemplates, futureStart, endDate, eventsCount); err != nil {
			return fmt.Errorf("failed to create future appointments: %w", err)
		}
	}

	return nil
}

// createTherapyAppointments creates therapy appointments using client data and therapy titles
func (cs *CalendarSeeder) createTherapyAppointments(calendar *entities.Calendar, templates []EventTemplate, startDate, endDate time.Time, count int) error {
	for i := 0; i < count; i++ {
		template := templates[rand.Intn(len(templates))]

		// Generate random date within period (Monday-Friday for most appointments)
		dayRange := int(endDate.Sub(startDate).Hours() / 24)
		if dayRange <= 0 {
			continue
		}

		var eventDate time.Time
		attempts := 0
		for attempts < 10 {
			randomDay := rand.Intn(dayRange)
			eventDate = startDate.AddDate(0, 0, randomDay)

			// Prefer weekdays, but allow some weekend appointments
			if eventDate.Weekday() == time.Saturday {
				if rand.Float64() < 0.3 { // 30% chance for Saturday
					break
				}
			} else if eventDate.Weekday() != time.Sunday {
				break // Weekdays are fine
			}
			attempts++
		}

		if attempts >= 10 {
			continue // Skip if couldn't find suitable date
		}

		// Choose random time slot
		timeSlot := template.TimeSlots[rand.Intn(len(template.TimeSlots))]
		startTime, err := time.Parse("15:04", timeSlot)
		if err != nil {
			continue // Skip invalid time slots
		}

		// Calculate end time
		endTime := startTime.Add(time.Duration(template.DurationMinutes) * time.Minute)

		// Combine with event date
		eventStart := time.Date(
			eventDate.Year(), eventDate.Month(), eventDate.Day(),
			startTime.Hour(), startTime.Minute(), 0, 0,
			eventDate.Location(),
		)

		eventEnd := time.Date(
			eventDate.Year(), eventDate.Month(), eventDate.Day(),
			endTime.Hour(), endTime.Minute(), 0, 0,
			eventDate.Location(),
		)

		// Select a random client and generate participants with real data
		client := cs.clients[rand.Intn(len(cs.clients))]
		participants := cs.generateTherapyParticipants(template, client)
		participantsJSON, _ := json.Marshal(participants)

		// Use therapy title in appointment title and description
		appointmentTitle := cs.generateAppointmentTitle(template, client)
		appointmentDescription := cs.generateAppointmentDescription(template, client)

		// Random location
		location := template.Locations[rand.Intn(len(template.Locations))]

		entry := &entities.CalendarEntry{
			TenantID:     calendar.TenantID,
			UserID:       calendar.UserID,
			CalendarID:   calendar.ID,
			Title:        appointmentTitle,
			IsException:  false,
			Participants: participantsJSON,
			DateFrom:     &eventDate,
			DateTo:       &eventDate,
			TimeFrom:     &eventStart,
			TimeTo:       &eventEnd,
			Timezone:     "Europe/Berlin",
			Type:         template.Type,
			Description:  appointmentDescription,
			Location:     location,
			IsAllDay:     false,
		}

		if err := cs.db.Create(entry).Error; err != nil {
			return fmt.Errorf("failed to create therapy appointment: %w", err)
		}
	}

	return nil
}

// generateTherapyParticipants creates participants based on appointment type and client data
func (cs *CalendarSeeder) generateTherapyParticipants(template EventTemplate, client UnburdyClient) []map[string]interface{} {
	participants := []map[string]interface{}{}

	// Always include the primary client
	participants = append(participants, map[string]interface{}{
		"name":  fmt.Sprintf("%s %s", client.FirstName, client.LastName),
		"email": client.Email,
		"phone": client.Phone,
		"role":  "client",
	})

	// Add additional participants based on appointment type
	switch template.Type {
	case "parent":
		// Add parent/guardian
		if client.ContactFirstName != "" {
			participants = append(participants, map[string]interface{}{
				"name":  fmt.Sprintf("%s %s", client.ContactFirstName, client.ContactLastName),
				"email": client.ContactEmail,
				"phone": client.ContactPhone,
				"role":  "parent/guardian",
			})
		}
		// Sometimes add alternative contact
		if client.AlternativeFirstName != "" && rand.Float64() < 0.4 {
			participants = append(participants, map[string]interface{}{
				"name":  fmt.Sprintf("%s %s", client.AlternativeFirstName, client.AlternativeLastName),
				"email": client.AlternativeEmail,
				"phone": client.AlternativePhone,
				"role":  "alternative_contact",
			})
		}
	case "information":
		// Sometimes include parent for information sessions
		if client.ContactFirstName != "" && rand.Float64() < 0.6 {
			participants = append(participants, map[string]interface{}{
				"name":  fmt.Sprintf("%s %s", client.ContactFirstName, client.ContactLastName),
				"email": client.ContactEmail,
				"phone": client.ContactPhone,
				"role":  "parent/guardian",
			})
		}
	}

	return participants
}

// generateAppointmentTitle creates appointment title using therapy title
func (cs *CalendarSeeder) generateAppointmentTitle(template EventTemplate, client UnburdyClient) string {
	baseTitles := map[string][]string{
		"therapy": {
			"Therapie: %s",
			"Einzeltherapie: %s",
			"Behandlung: %s",
			"Therapiesitzung: %s",
		},
		"parent": {
			"Elterngespräch: %s",
			"Beratung Eltern: %s",
			"Familiengespräch: %s",
			"Elternberatung: %s",
		},
		"information": {
			"Informationsgespräch: %s",
			"Beratung: %s",
			"Information: %s",
			"Erstberatung: %s",
		},
		"generic": {
			"Termin: %s",
			"Gespräch: %s",
			"Beratung: %s",
			"Sitzung: %s",
		},
	}

	titles, exists := baseTitles[template.Type]
	if !exists {
		titles = baseTitles["generic"]
	}

	titleTemplate := titles[rand.Intn(len(titles))]
	therapyTitle := client.TherapyTitle
	if therapyTitle == "" {
		therapyTitle = "Allgemeine Beratung"
	}

	return fmt.Sprintf(titleTemplate, therapyTitle)
}

// generateAppointmentDescription creates detailed description using client and therapy data
func (cs *CalendarSeeder) generateAppointmentDescription(template EventTemplate, client UnburdyClient) string {
	baseDescriptions := []string{
		"Fortsetzung der Behandlung",
		"Therapiesitzung gemäß Behandlungsplan",
		"Individuelle Betreuung und Förderung",
		"Therapeutische Intervention",
		"Weitere Schritte besprechen",
	}

	baseDescription := baseDescriptions[rand.Intn(len(baseDescriptions))]

	// Add client-specific details
	details := []string{}
	if client.PrimaryLanguage != "English" && client.PrimaryLanguage != "" {
		details = append(details, fmt.Sprintf("Sprache: %s", client.PrimaryLanguage))
	}
	if client.Status == "active" {
		details = append(details, "Status: Aktiv in Behandlung")
	}

	description := baseDescription
	if len(details) > 0 {
		description += fmt.Sprintf(" (%s)", strings.Join(details, ", "))
	}

	return description
}

// createHolidayEntries creates holiday calendar entries from holidays.json
func (cs *CalendarSeeder) createHolidayEntries(calendar *entities.Calendar, startDate, endDate time.Time) error {
	// Load holidays data
	holidays, err := cs.loadHolidaysData()
	if err != nil {
		return fmt.Errorf("failed to load holidays data: %w", err)
	}

	if len(holidays) == 0 {
		return fmt.Errorf("no holiday data found")
	}

	// Use the first (BW) holidays data
	holidayData := holidays[0]

	// Create public holiday entries
	if err := cs.createPublicHolidayEntries(calendar, holidayData, startDate, endDate); err != nil {
		return fmt.Errorf("failed to create public holidays: %w", err)
	}

	// Create school holiday entries
	if err := cs.createSchoolHolidayEntries(calendar, holidayData, startDate, endDate); err != nil {
		return fmt.Errorf("failed to create school holidays: %w", err)
	}

	return nil
}

// loadHolidaysData loads holiday data from the holidays.json file
func (cs *CalendarSeeder) loadHolidaysData() ([]HolidayData, error) {
	// Try different possible locations for the holidays.json file
	possiblePaths := []string{
		"../statics/json/holidays.json",                     // From seed directory
		"./statics/json/holidays.json",                      // From project root
		"../../base-server/statics/json/holidays.json",     // From seed to base-server
		"../../../base-server/statics/json/holidays.json",  // From deeper nested
	}

	var data []byte
	var err error
	var usedPath string

	for _, path := range possiblePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			usedPath = path
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("holidays.json not found in any expected location: %w", err)
	}

	fmt.Printf("Loading holidays data from: %s\n", usedPath)

	var holidays []HolidayData
	if err := json.Unmarshal(data, &holidays); err != nil {
		return nil, fmt.Errorf("failed to parse holidays.json: %w", err)
	}

	return holidays, nil
}

// createPublicHolidayEntries creates calendar entries for public holidays
func (cs *CalendarSeeder) createPublicHolidayEntries(calendar *entities.Calendar, holidayData HolidayData, startDate, endDate time.Time) error {
	publicHolidayCount := 0
	for year, holidays := range holidayData.PublicHolidays {
		for name, dateStr := range holidays {
			// Parse holiday date
			holidayDate, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				fmt.Printf("Warning: Failed to parse holiday date %s: %v\n", dateStr, err)
				continue
			}

			// Check if holiday is within our seeding range
			if holidayDate.Before(startDate) || holidayDate.After(endDate) {
				continue
			}

			// Create all-day holiday entry
			entry := &entities.CalendarEntry{
				TenantID:     calendar.TenantID,
				UserID:       calendar.UserID,
				CalendarID:   calendar.ID,
				Title:        fmt.Sprintf("🎉 %s", name),
				IsException:  false,
				Participants: json.RawMessage("[]"),
				DateFrom:     &holidayDate,
				DateTo:       &holidayDate,
				TimeFrom:     nil, // All-day events have no specific time
				TimeTo:       nil,
				Timezone:     "Europe/Berlin",
				Type:         "public_holiday",
				Description:  fmt.Sprintf("Gesetzlicher Feiertag in Baden-Württemberg (%s %s)", name, year),
				Location:     "",
				IsAllDay:     true,
			}

			if err := cs.db.Create(entry).Error; err != nil {
				fmt.Printf("Warning: Failed to create public holiday entry for %s: %v\n", name, err)
			} else {
				publicHolidayCount++
			}
		}
	}

	fmt.Printf("Created %d public holidays for calendar %d (%s)\n", publicHolidayCount, calendar.ID, calendar.Title)
	return nil
}

// createSchoolHolidayEntries creates calendar entries for school holidays
func (cs *CalendarSeeder) createSchoolHolidayEntries(calendar *entities.Calendar, holidayData HolidayData, startDate, endDate time.Time) error {
	schoolHolidayCount := 0
	for year, holidays := range holidayData.SchoolHolidays {
		for name, dateRange := range holidays {
			if len(dateRange) != 2 {
				fmt.Printf("Warning: Invalid date range for school holiday %s\n", name)
				continue
			}

			// Parse start and end dates
			startHoliday, err := time.Parse("2006-01-02", dateRange[0])
			if err != nil {
				fmt.Printf("Warning: Failed to parse holiday start date %s: %v\n", dateRange[0], err)
				continue
			}

			endHoliday, err := time.Parse("2006-01-02", dateRange[1])
			if err != nil {
				fmt.Printf("Warning: Failed to parse holiday end date %s: %v\n", dateRange[1], err)
				continue
			}

			// Check if holiday period overlaps with our seeding range
			if endHoliday.Before(startDate) || startHoliday.After(endDate) {
				continue
			}

			// Adjust dates to fit within seeding range
			if startHoliday.Before(startDate) {
				startHoliday = startDate
			}
			if endHoliday.After(endDate) {
				endHoliday = endDate
			}

			// Create multi-day holiday entry
			entry := &entities.CalendarEntry{
				TenantID:     calendar.TenantID,
				UserID:       calendar.UserID,
				CalendarID:   calendar.ID,
				Title:        fmt.Sprintf("🏫 %s", name),
				IsException:  false,
				Participants: json.RawMessage("[]"),
				DateFrom:     &startHoliday,
				DateTo:       &endHoliday,
				TimeFrom:     nil, // All-day events have no specific time
				TimeTo:       nil,
				Timezone:     "Europe/Berlin",
				Type:         "school_holiday",
				Description:  fmt.Sprintf("Schulferien in Baden-Württemberg (%s %s) - %s bis %s", name, year, startHoliday.Format("02.01.2006"), endHoliday.Format("02.01.2006")),
				Location:     "",
				IsAllDay:     true,
			}

			if err := cs.db.Create(entry).Error; err != nil {
				fmt.Printf("Warning: Failed to create school holiday entry for %s: %v\n", name, err)
			} else {
				schoolHolidayCount++
			}
		}
	}

	fmt.Printf("Created %d school holidays for calendar %d (%s)\n", schoolHolidayCount, calendar.ID, calendar.Title)
	return nil
}

// Helper functions

func (cs *CalendarSeeder) getCalendarTypeFromTitle(title string) string {
	title = strings.ToLower(title)
	if strings.Contains(title, "therapy") || strings.Contains(title, "appointment") || strings.Contains(title, "treatment") {
		return "therapy"
	}
	if strings.Contains(title, "work") || strings.Contains(title, "business") || strings.Contains(title, "office") {
		return "work"
	}
	if strings.Contains(title, "personal") || strings.Contains(title, "private") {
		return "personal"
	}
	return "mixed"
}

func (cs *CalendarSeeder) generateParticipants(count int) []map[string]interface{} {
	if count == 0 {
		return []map[string]interface{}{}
	}

	participants := make([]map[string]interface{}, count)
	names := []string{"Alice Smith", "Bob Johnson", "Carol Williams", "David Brown", "Eva Davis"}

	for i := 0; i < count && i < len(names); i++ {
		name := names[i]
		email := strings.ToLower(strings.ReplaceAll(name, " ", ".")) + "@company.com"

		participants[i] = map[string]interface{}{
			"name":  name,
			"email": email,
		}
	}

	return participants
}

// loadUnburdyClients loads client data from unburdy_server seed file
func loadUnburdyClients() []UnburdyClient {
	// For now, return a subset of the clients with therapy_title data
	// In a real implementation, this would read from the JSON file
	return []UnburdyClient{
		{
			FirstName:            "Emma",
			LastName:             "Johnson",
			Email:                "Emma.Johnson@academy.edu",
			Phone:                "+1-555-3584",
			ContactFirstName:     "Christina",
			ContactLastName:      "Johnson",
			ContactEmail:         "Christina.Johnson@email.com",
			ContactPhone:         "+1-555-2936",
			AlternativeFirstName: "Gregory",
			AlternativeLastName:  "Johnson",
			AlternativeEmail:     "Gregory.Johnson@work.com",
			AlternativePhone:     "+1-555-2375",
			TherapyTitle:         "Leichte Intelligenzminderung",
			Status:               "active",
			City:                 "Springfield",
			PrimaryLanguage:      "English",
		},
		{
			FirstName:            "Michael",
			LastName:             "Chen",
			Email:                "Michael.Chen@academy.edu",
			Phone:                "+1-555-7301",
			ContactFirstName:     "Patricia",
			ContactLastName:      "Chen",
			ContactEmail:         "Patricia.Chen@email.com",
			ContactPhone:         "+1-555-2537",
			AlternativeFirstName: "Eric",
			AlternativeLastName:  "Chen",
			AlternativeEmail:     "Eric.Chen@work.com",
			AlternativePhone:     "+1-555-8441",
			TherapyTitle:         "Isolierte Rechtschreibstörung",
			Status:               "waiting",
			City:                 "Chicago",
			PrimaryLanguage:      "English",
		},
		{
			FirstName:            "Sarah",
			LastName:             "Williams",
			Email:                "Sarah.Williams@school.org",
			Phone:                "+1-555-2400",
			ContactFirstName:     "Michelle",
			ContactLastName:      "Williams",
			ContactEmail:         "Michelle.Williams@email.com",
			ContactPhone:         "+1-555-0736",
			AlternativeFirstName: "Michael",
			AlternativeLastName:  "Williams",
			AlternativeEmail:     "Michael.Williams@work.com",
			AlternativePhone:     "+1-555-1472",
			TherapyTitle:         "Anpassungsstörung",
			Status:               "archived",
			City:                 "Madison",
			PrimaryLanguage:      "English",
		},
		{
			FirstName:            "David",
			LastName:             "Rodriguez",
			Email:                "David.Rodriguez@student.edu",
			Phone:                "+1-555-5074",
			ContactFirstName:     "Patricia",
			ContactLastName:      "Rodriguez",
			ContactEmail:         "Patricia.Rodriguez@email.com",
			ContactPhone:         "+1-555-9270",
			AlternativeFirstName: "Brandon",
			AlternativeLastName:  "Rodriguez",
			AlternativeEmail:     "Brandon.Rodriguez@work.com",
			AlternativePhone:     "+1-555-9845",
			TherapyTitle:         "Leichte Intelligenzminderung",
			Status:               "active",
			City:                 "Milwaukee",
			PrimaryLanguage:      "Spanish",
		},
		{
			FirstName:            "Maria",
			LastName:             "Garcia",
			Email:                "Maria.Garcia@student.edu",
			Phone:                "+1-555-7322",
			ContactFirstName:     "Christina",
			ContactLastName:      "Garcia",
			ContactEmail:         "Christina.Garcia@email.com",
			ContactPhone:         "+1-555-6605",
			AlternativeFirstName: "Michael",
			AlternativeLastName:  "Garcia",
			AlternativeEmail:     "Michael.Garcia@work.com",
			AlternativePhone:     "+1-555-8381",
			TherapyTitle:         "Lese-Rechtschreib-Störung",
			Status:               "active",
			City:                 "Peoria",
			PrimaryLanguage:      "Spanish",
		},
		{
			FirstName:            "Jennifer",
			LastName:             "Davis",
			Email:                "Jennifer.Davis@highschool.edu",
			Phone:                "+1-555-4610",
			ContactFirstName:     "Patricia",
			ContactLastName:      "Davis",
			ContactEmail:         "Patricia.Davis@email.com",
			ContactPhone:         "+1-555-1967",
			AlternativeFirstName: "Justin",
			AlternativeLastName:  "Davis",
			AlternativeEmail:     "Justin.Davis@work.com",
			AlternativePhone:     "+1-555-7996",
			TherapyTitle:         "Generalisierte Angststörung",
			Status:               "active",
			City:                 "Naperville",
			PrimaryLanguage:      "English",
		},
	}
}
