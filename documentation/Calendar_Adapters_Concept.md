# Calendar Provider Abstraction Layer

## Requirements & Implementation Guide

### Goal

Allow users to connect external calendar systems:

* Microsoft Outlook / Microsoft 365
* Google Calendar
* Apple iCloud Calendar
* Generic CalDAV Servers
* Radicale
* Nextcloud Calendar
* Fastmail Calendar
* Internal Unburdy Calendar

without changing booking functionality.

The booking engine must operate against a unified calendar interface.

---

# Architecture

## Current State

Booking Templates
↓
Calendar
↓
Calendar Entries
↓
Calendar Series

The internal database is the source of truth.

---

## Target State

Booking Templates
↓
Calendar Provider Interface
↓
┌─────────────────────┐
│ Internal Calendar   │
├─────────────────────┤
│ Google Provider     │
├─────────────────────┤
│ Outlook Provider    │
├─────────────────────┤
│ iCloud Provider     │
├─────────────────────┤
│ CalDAV Provider     │
├─────────────────────┤
│ Radicale Provider   │
└─────────────────────┘

The booking engine never talks directly to a calendar implementation.

---

# Provider Types

enum CalendarProviderType

* internal
* google
* microsoft
* icloud
* caldav
* radicale

---

# Database Changes

## calendars

Add:

provider_type
provider_account_id
external_calendar_id
sync_token
read_only
last_sync_at
sync_status

Example

{
"id": 12,
"provider_type": "google",
"external_calendar_id": "primary",
"read_only": false
}

---

## calendar_connections

New table

calendar_connections

id
tenant_id
user_id

provider

access_token
refresh_token

token_expires_at

provider_user_id

provider_email

sync_token

created_at
updated_at

---

# Unified Calendar Interface

type CalendarProvider interface {

```
ListCalendars()

GetCalendar()

CreateEvent()

UpdateEvent()

DeleteEvent()

GetEvents()

GetBusyTimes()

WatchChanges()

RefreshToken()
```

}

Every provider must implement this interface.

---

# Provider Implementations

## Internal Provider

Wrap existing database implementation.

No functional changes.

Acts as reference implementation.

---

## Google Provider

API

Google Calendar API

OAuth2

Scopes

calendar
calendar.events

Capabilities

✓ read calendars
✓ read events
✓ create events
✓ update events
✓ delete events
✓ webhooks
✓ incremental sync

---

## Microsoft Provider

API

Microsoft Graph

OAuth2

Scopes

Calendars.ReadWrite

Capabilities

✓ read calendars
✓ create events
✓ update events
✓ delete events
✓ webhooks
✓ delta sync

---

## iCloud Provider

Apple provides no modern public Calendar API.

Recommended implementation:

Use CalDAV

Provider Type:
icloud

Underlying Protocol:
CalDAV

Requirements:

Apple ID
App-specific password

Connection:

https://caldav.icloud.com

Capabilities

✓ read calendars
✓ create events
✓ update events
✓ delete events

No native webhook support.

Polling required.

---

## Generic CalDAV Provider

Used for:

* Radicale
* Nextcloud
* Fastmail
* OwnCloud
* Baikal

Configuration:

server_url
username
password

Capabilities

✓ read calendars
✓ create events
✓ update events
✓ delete events

No webhooks.

Periodic synchronization.

---

# Event Model

Internal Unified Event

{
id
provider

external_id

title

description

location

start_time
end_time

timezone

organizer

attendees

recurrence_rule

status

last_modified
}

All providers map into this model.

---

# Booking Engine Changes

Current

Booking Engine
↓
Calendar Entries

Replace with

Booking Engine
↓
Calendar Availability Service
↓
Provider Interface

---

# Availability Service

New service:

AvailabilityService

Responsibilities:

* get busy slots
* merge multiple calendars
* timezone conversion
* free slot calculation
* booking conflict detection

Methods

GetBusyTimes()

GetFreeSlots()

CreateBooking()

CancelBooking()

RescheduleBooking()

---

# Multi Calendar Support

User may connect multiple calendars.

Example

Google Work Calendar
Google Personal Calendar
iCloud Calendar

Availability must be calculated across all calendars.

Algorithm

busy_slots =
union(
work_calendar,
personal_calendar,
icloud_calendar
)

free_slots =
working_hours - busy_slots

---

# Sync Strategy

Google

Webhook + Incremental Sync

Microsoft

Webhook + Delta Sync

CalDAV

Polling every 5-15 minutes

iCloud

Polling every 5-15 minutes

Radicale

Polling every 5-15 minutes

---

# Webhook Service

New component

CalendarSyncService

Responsibilities

Receive provider notifications

Trigger synchronization

Update internal cache

---

# Caching Layer

Recommended

calendar_event_cache

Fields

provider
calendar_id
external_event_id
etag
last_sync

Purpose

Avoid excessive API calls.

---

# Booking Creation Flow

User selects slot
↓
Availability Service validates slot
↓
Provider.CreateEvent()
↓
External Calendar
↓
Persist Mapping
↓
Return Success

---

# Event Mapping Table

calendar_event_mapping

id

calendar_id

provider

external_event_id

internal_event_id

etag

last_sync

---

# Migration Strategy

Phase 1

Introduce Provider Interface

Keep Internal Calendar

No behavior changes

---

Phase 2

Implement Google Provider

---

Phase 3

Implement Microsoft Provider

---

Phase 4

Implement CalDAV Provider

Supports:

* iCloud
* Radicale
* Nextcloud

with one implementation

---

Phase 5

Replace all booking availability logic with AvailabilityService

---

# Recommended Go Packages

OAuth

golang.org/x/oauth2

Google

google.golang.org/api/calendar/v3

Microsoft

msgraph-sdk-go

CalDAV

github.com/emersion/go-webdav/caldav

Recurrence

github.com/teambition/rrule-go

Timezone

time package + IANA TZ database

---

# Key Design Principle

The booking engine must never know whether an event is stored:

* internally
* in Google Calendar
* in Outlook
* in iCloud
* in Radicale

Everything must pass through the CalendarProvider interface.
