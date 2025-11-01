package entities

import (
	"github.com/ae-base-server/pkg/core"
)

// CalendarEntity implements core.Entity for Calendar model
type CalendarEntity struct{}

func NewCalendarEntity() core.Entity {
	return &CalendarEntity{}
}

func (e *CalendarEntity) TableName() string {
	return "calendars"
}

func (e *CalendarEntity) GetModel() interface{} {
	return &Calendar{}
}

func (e *CalendarEntity) GetMigrations() []core.Migration {
	return []core.Migration{} // No custom migrations needed, GORM handles basic schema
}

// CalendarEntryEntity implements core.Entity for CalendarEntry model
type CalendarEntryEntity struct{}

func NewCalendarEntryEntity() core.Entity {
	return &CalendarEntryEntity{}
}

func (e *CalendarEntryEntity) TableName() string {
	return "calendar_entries"
}

func (e *CalendarEntryEntity) GetModel() interface{} {
	return &CalendarEntry{}
}

func (e *CalendarEntryEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}

// CalendarSeriesEntity implements core.Entity for CalendarSeries model
type CalendarSeriesEntity struct{}

func NewCalendarSeriesEntity() core.Entity {
	return &CalendarSeriesEntity{}
}

func (e *CalendarSeriesEntity) TableName() string {
	return "calendar_series"
}

func (e *CalendarSeriesEntity) GetModel() interface{} {
	return &CalendarSeries{}
}

func (e *CalendarSeriesEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}

// ExternalCalendarEntity implements core.Entity for ExternalCalendar model
type ExternalCalendarEntity struct{}

func NewExternalCalendarEntity() core.Entity {
	return &ExternalCalendarEntity{}
}

func (e *ExternalCalendarEntity) TableName() string {
	return "external_calendars"
}

func (e *ExternalCalendarEntity) GetModel() interface{} {
	return &ExternalCalendar{}
}

func (e *ExternalCalendarEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}
