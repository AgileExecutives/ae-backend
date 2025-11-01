package seeding

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/unburdy/calendar-module/entities"
	"gorm.io/gorm"
)

// CalendarSeeder handles seeding of calendar data
type CalendarSeeder struct {
	db     *gorm.DB
	config *SeedConfig
}

// SeedConfig represents the seeding configuration
type SeedConfig struct {
	SeedConfig         SeedConfigDetails        `json:"seed_config"`
	CalendarTemplates  []CalendarTemplate       `json:"calendar_templates"`
	RecurringTemplates []RecurringSeriesTemplate `json:"recurring_series_templates"`
	EventTemplates     EventTemplateGroups      `json:"event_templates"`
	ExternalTemplates  []ExternalCalendarTemplate `json:"external_calendar_templates"`
	GenerationRules    GenerationRules          `json:"generation_rules"`
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
			Weeks             int `json:"weeks"`
			EventsPerWeek     struct {
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
	Title               string                   `json:"title"`
	Description         string                   `json:"description"`
	Weekday             int                      `json:"weekday"`
	Interval            int                      `json:"interval"`
	TimeStart           string                   `json:"time_start"`
	TimeEnd             string                   `json:"time_end"`
	Location            string                   `json:"location"`
	Type                string                   `json:"type"`
	Participants        []map[string]interface{} `json:"participants"`
	CalendarType        string                   `json:"calendar_type"`
	ProbabilityCurrent  float64                  `json:"probability_current"`
	ProbabilityPast     float64                  `json:"probability_past"`
	ProbabilityFuture   float64                  `json:"probability_future"`
}

type EventTemplateGroups struct {
	Work     []EventTemplate `json:"work"`
	Personal []EventTemplate `json:"personal"`
}

type EventTemplate struct {
	Title             string            `json:"title"`
	Type              string            `json:"type"`
	DurationMinutes   int               `json:"duration_minutes"`
	Locations         []string          `json:"locations"`
	ParticipantsCount EventRange        `json:"participants_count"`
	Descriptions      []string          `json:"descriptions"`
	TimeSlots         []string          `json:"time_slots"`
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
	AvoidConflicts           bool    `json:"avoid_conflicts"`
	MinGapBetweenEvents      int     `json:"min_gap_between_events"`
	WeekendWorkProbability   float64 `json:"weekend_work_probability"`
	AllDayEventProbability   float64 `json:"all_day_event_probability"`
	ExceptionProbability     float64 `json:"exception_probability"`
	NoShowProbability        float64 `json:"no_show_probability"`
}

// NewCalendarSeeder creates a new calendar seeder
func NewCalendarSeeder(db *gorm.DB, configPath string) (*CalendarSeeder, error) {
	config, err := loadSeedConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load seed config: %w", err)
	}

	return &CalendarSeeder{
		db:     db,
		config: config,
	}, nil
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

	// Create external calendars
	for _, calendar := range calendars {
		if rand.Float64() < 0.3 { // 30% chance of having external calendars
			if err := cs.createExternalCalendars(calendar); err != nil {
				return fmt.Errorf("failed to create external calendars: %w", err)
			}
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

	// Determine how many series to create based on time period density
	seriesCount := cs.getRecurringSeriesCount(now, startDate, endDate)
	
	for i := 0; i < seriesCount && i < len(relevantTemplates); i++ {
		template := relevantTemplates[i]
		
		// Check probability based on time period
		probability := cs.getRecurringProbability(template, now)
		if rand.Float64() > probability {
			continue
		}

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

		// Random chance of exceptions or no-shows
		isException := rand.Float64() < cs.config.GenerationRules.ExceptionProbability

		entry := &entities.CalendarEntry{
			TenantID:     series.TenantID,
			UserID:       series.UserID,
			CalendarID:   series.CalendarID,
			SeriesID:     &series.ID,
			Title:        series.Title,
			IsException:  isException,
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

// createIndividualEvents creates standalone calendar events
func (cs *CalendarSeeder) createIndividualEvents(calendar *entities.Calendar, startDate, endDate, now time.Time) error {
	calendarType := cs.getCalendarTypeFromTitle(calendar.Title)
	
	// Get event templates for this calendar type
	var eventTemplates []EventTemplate
	switch calendarType {
	case "work":
		eventTemplates = cs.config.EventTemplates.Work
	case "personal":
		eventTemplates = cs.config.EventTemplates.Personal
	default:
		eventTemplates = append(cs.config.EventTemplates.Work, cs.config.EventTemplates.Personal...)
	}

	// Process time periods with different densities
	periods := []struct {
		start, end   time.Time
		eventsPerWeek EventRange
		label        string
	}{
		{
			start:         startDate,
			end:           now.AddDate(0, 0, -int(cs.config.SeedConfig.DensityProfile.CurrentPeriod.Weeks)*7),
			eventsPerWeek: cs.config.SeedConfig.DensityProfile.PastPeriod.EventsPerWeek,
			label:         "past",
		},
		{
			start: now.AddDate(0, 0, -int(cs.config.SeedConfig.DensityProfile.CurrentPeriod.Weeks)*7),
			end:   now.AddDate(0, 0, int(cs.config.SeedConfig.DensityProfile.CurrentPeriod.Weeks)*7),
			eventsPerWeek: EventRange{
				Min: cs.config.SeedConfig.DensityProfile.CurrentPeriod.EventsPerWeek.Min,
				Max: cs.config.SeedConfig.DensityProfile.CurrentPeriod.EventsPerWeek.Max,
			},
			label: "current",
		},
		{
			start:         now.AddDate(0, 0, int(cs.config.SeedConfig.DensityProfile.CurrentPeriod.Weeks)*7),
			end:           endDate,
			eventsPerWeek: cs.config.SeedConfig.DensityProfile.FuturePeriod.EventsPerWeek,
			label:         "future",
		},
	}

	for _, period := range periods {
		if period.start.Before(period.end) {
			weeks := int(period.end.Sub(period.start).Hours() / (24 * 7))
			if weeks > 0 {
				if err := cs.createEventsForPeriod(calendar, eventTemplates, period.start, period.end, period.eventsPerWeek, weeks, calendarType); err != nil {
					return fmt.Errorf("failed to create events for %s period: %w", period.label, err)
				}
			}
		}
	}

	return nil
}

// createEventsForPeriod creates events for a specific time period
func (cs *CalendarSeeder) createEventsForPeriod(calendar *entities.Calendar, templates []EventTemplate, startDate, endDate time.Time, eventsPerWeek EventRange, weeks int, calendarType string) error {
	totalEvents := (eventsPerWeek.Min + rand.Intn(eventsPerWeek.Max-eventsPerWeek.Min+1)) * weeks / 2 // Divide by 2 since we have 2 calendars typically

	for i := 0; i < totalEvents; i++ {
		template := templates[rand.Intn(len(templates))]
		
		// Generate random date within period
		dayRange := int(endDate.Sub(startDate).Hours() / 24)
		randomDay := rand.Intn(dayRange)
		eventDate := startDate.AddDate(0, 0, randomDay)
		
		// Skip weekends for work events (with some probability)
		if calendarType == "work" && (eventDate.Weekday() == time.Saturday || eventDate.Weekday() == time.Sunday) {
			if rand.Float64() > cs.config.GenerationRules.WeekendWorkProbability {
				continue
			}
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

		// Generate participants
		participantCount := template.ParticipantsCount.Min + rand.Intn(template.ParticipantsCount.Max-template.ParticipantsCount.Min+1)
		participants := cs.generateParticipants(participantCount)
		participantsJSON, _ := json.Marshal(participants)

		// Random location and description
		location := template.Locations[rand.Intn(len(template.Locations))]
		description := template.Descriptions[rand.Intn(len(template.Descriptions))]

		// Check for all-day events
		isAllDay := rand.Float64() < cs.config.GenerationRules.AllDayEventProbability

		entry := &entities.CalendarEntry{
			TenantID:     calendar.TenantID,
			UserID:       calendar.UserID,
			CalendarID:   calendar.ID,
			Title:        template.Title,
			IsException:  false,
			Participants: participantsJSON,
			DateFrom:     &eventDate,
			DateTo:       &eventDate,
			TimeFrom:     &eventStart,
			TimeTo:       &eventEnd,
			Timezone:     "Europe/Berlin",
			Type:         template.Type,
			Description:  description,
			Location:     location,
			IsAllDay:     isAllDay,
		}

		if err := cs.db.Create(entry).Error; err != nil {
			return fmt.Errorf("failed to create individual event: %w", err)
		}
	}

	return nil
}

// createExternalCalendars creates external calendar integrations
func (cs *CalendarSeeder) createExternalCalendars(calendar *entities.Calendar) error {
	for _, template := range cs.config.ExternalTemplates {
		if rand.Float64() < 0.5 { // 50% chance per external calendar
			settingsJSON, err := json.Marshal(template.Settings)
			if err != nil {
				return fmt.Errorf("failed to marshal settings: %w", err)
			}

			external := &entities.ExternalCalendar{
				TenantID:     calendar.TenantID,
				UserID:       calendar.UserID,
				CalendarID:   calendar.ID,
				Title:        template.Title,
				URL:          template.URL,
				Settings:     settingsJSON,
				Color:        template.Color,
				CalendarUUID: uuid.New().String(),
			}

			if err := cs.db.Create(external).Error; err != nil {
				return fmt.Errorf("failed to create external calendar: %w", err)
			}
		}
	}

	return nil
}

// Helper functions

func (cs *CalendarSeeder) getCalendarTypeFromTitle(title string) string {
	title = strings.ToLower(title)
	if strings.Contains(title, "work") || strings.Contains(title, "business") || strings.Contains(title, "office") {
		return "work"
	}
	if strings.Contains(title, "personal") || strings.Contains(title, "private") {
		return "personal"
	}
	return "mixed"
}

func (cs *CalendarSeeder) getRecurringSeriesCount(now, startDate, endDate time.Time) int {
	current := cs.config.SeedConfig.DensityProfile.CurrentPeriod
	past := cs.config.SeedConfig.DensityProfile.PastPeriod
	future := cs.config.SeedConfig.DensityProfile.FuturePeriod

	// Calculate weighted average based on time periods
	currentWeight := float64(current.Weeks * 7) / float64(endDate.Sub(startDate).Hours()/24)
	
	avgCount := (float64(current.RecurringSeriesCount.Min+current.RecurringSeriesCount.Max)/2)*currentWeight +
		(float64(past.RecurringSeriesCount.Min+past.RecurringSeriesCount.Max)/2)*(1-currentWeight)*0.3 +
		(float64(future.RecurringSeriesCount.Min+future.RecurringSeriesCount.Max)/2)*(1-currentWeight)*0.7

	return int(avgCount)
}

func (cs *CalendarSeeder) getRecurringProbability(template RecurringSeriesTemplate, now time.Time) float64 {
	// For simplicity, use current probability as baseline
	// In a more sophisticated version, this could vary based on the time period
	return template.ProbabilityCurrent
}

func (cs *CalendarSeeder) generateParticipants(count int) []map[string]interface{} {
	participants := make([]map[string]interface{}, count)
	names := []string{"Alice Smith", "Bob Johnson", "Carol Williams", "David Brown", "Eva Davis", "Frank Miller", "Grace Wilson", "Henry Moore", "Ivy Taylor", "Jack Anderson"}
	
	for i := 0; i < count; i++ {
		name := names[rand.Intn(len(names))]
		email := strings.ToLower(strings.ReplaceAll(name, " ", ".")) + "@company.com"
		
		participants[i] = map[string]interface{}{
			"name":  name,
			"email": email,
		}
	}
	
	return participants
}

// loadSeedConfig loads the seed configuration from a JSON file
func loadSeedConfig(configPath string) (*SeedConfig, error) {
	// This would typically read from a file, but for now we'll create a default config
	// In a real implementation, you'd use ioutil.ReadFile and json.Unmarshal
	return &SeedConfig{}, fmt.Errorf("config loading not implemented - use embedded config for now")
}