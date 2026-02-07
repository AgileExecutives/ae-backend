
# Client Remaining Units feature
please investigate in the code and create and implementation plan

a client - if his therapy is payed by cost-provider - has a number of approved units. a unit usually has 60min depending on the cost provider (add a field minutes_per_unit default 60 to cost_provider with default 60).

the client has a number of already conducted / scheduled minutes of sessions plus a number of extra efforts that are billable.

keep the maximum number of units the client can schedule in the client object.

the remaining units are calculated from the remaining minutes of the sessions and billable extra efforts.

the client module can subscribe to the session created and updated and deleted events as well to the extra effort created updated and deleted events and then call a service that counts the remaining minutes.

remaining units - those are the number of sessions we can schedule.

## Implementation plan (unburdy_server/modules/client_management)

### 1) Data model changes (GORM AutoMigrate)
- Add `minutes_per_unit` to cost providers
	- File: `unburdy_server/modules/client_management/entities/cost_provider.go`
	- Field: `MinutesPerUnit int \`gorm:"not null;default:60" json:"minutes_per_unit"\``
	- Add to `CreateCostProviderRequest` (optional, default 60) and `UpdateCostProviderRequest` (optional pointer)
	- Add to `CostProviderResponse`
- Add approved/max units + cached rollups to clients
	- File: `unburdy_server/modules/client_management/entities/client.go`
	- Fields (suggested):
		- `MaxSchedulableUnits int \`gorm:"not null;default:0" json:"max_schedulable_units"\``
		- `UsedBillableMinutes int \`gorm:"not null;default:0" json:"used_billable_minutes"\``
		- `RemainingBillableMinutes int \`gorm:"not null;default:0" json:"remaining_billable_minutes"\``
	- Add `max_schedulable_units` to Create/Update request types and to `ClientResponse`
	- Add derived response fields (computed, not persisted) for convenience:
		- `RemainingUnits float64` (or string/decimal), computed as `remaining_minutes / minutes_per_unit`

### 2) Calculation rules (service)
- Add a service that recalculates usage for one client
	- New file: `unburdy_server/modules/client_management/services/client_remaining_units_service.go`
	- Method: `RecalculateForClient(ctx context.Context, tenantID, clientID uint) error`
- Query inputs:
	- Sessions: sum `duration_min` for the client where `status != 'canceled'` (matches “conducted / scheduled”; includes invoice-flow statuses)
	- Extra efforts: sum `duration_min` where `billable = true` (and not soft-deleted)
- Compute:
	- `approved_minutes = client.MaxSchedulableUnits * costProvider.MinutesPerUnit`
	- `used_minutes = session_minutes + extra_effort_minutes`
	- `remaining_minutes = max(0, approved_minutes - used_minutes)`
	- `remaining_units = remaining_minutes / minutes_per_unit` (guard `minutes_per_unit <= 0`)
- Behavior when client has no cost provider or is self payer:
	- Option A (simplest): treat `minutes_per_unit=60`, `MaxSchedulableUnits=0` → remaining 0
	- Option B (more explicit): return `remaining_* = null` in API and do not enforce limits
	- Pick one and apply consistently in handlers + frontend contract.

### 3) Wire recalculation to write-paths
There are two viable ways; the TODO suggests an event-driven one.

#### 3A) Event-driven (matches TODO)
- Emit events from services after successful DB mutations
	- SessionService: publish `client_management.session.created|updated|deleted` with `{tenant_id, client_id, session_id}`
		- File: `unburdy_server/modules/client_management/services/session_service.go`
	- ExtraEffortService: publish `client_management.extra_effort.created|updated|deleted` with `{tenant_id, client_id, extra_effort_id}`
		- File: `unburdy_server/modules/client_management/services/extra_effort_service.go`
	- Requires passing `ctx.EventBus` into these services from module init
		- File: `unburdy_server/modules/client_management/module.go`
- Subscribe and recalc
	- New handler type in `unburdy_server/modules/client_management/events/remaining_units_events.go`
	- Implement `core.EventHandler` subscribing to the 6 event types and calling the new recalculation service.
	- Register in `CoreModule.EventHandlers()`.

#### 3B) Direct-call (simpler, fewer moving parts)
- Call `RecalculateForClient()` directly at the end of:
	- `SessionService.CreateSession/UpdateSession/DeleteSession`
	- `ExtraEffortService.CreateExtraEffort/UpdateExtraEffort/DeleteExtraEffort`
- This avoids event bus wiring but couples services.

### 4) Keep results fresh on related updates
- When `client.max_schedulable_units` changes (ClientService.UpdateClient), trigger a recalculation for that client.
- When `cost_provider.minutes_per_unit` changes (CostProviderService.UpdateCostProvider), trigger a recalculation for all clients referencing that cost provider.

### 5) API contract updates
- Ensure `GET /clients` and `GET /clients/:id` return the new fields.
- Consider adding a dedicated endpoint if you want to avoid enlarging default payloads:
	- `GET /clients/:id/remaining-units`

### 6) Tests
- Add unit/integration tests under `unburdy_server/modules/client_management/tests/`
	- New test file: `remaining_units_service_test.go`
	- Setup: in-memory sqlite with AutoMigrate for Client/CostProvider/Session/ExtraEffort
	- Cases:
		- default minutes_per_unit=60, max_units=10, sessions=120 min, billable efforts=60 min → remaining_minutes=420, remaining_units=7
		- canceled sessions don’t count
		- non-billable efforts don’t count
		- updating/deleting a session or effort triggers recalculation (event-driven: publish/subscribe via `core.NewEventBus()`)

### 7) Acceptance criteria
- Cost provider can store `minutes_per_unit` (default 60) and it flows through API.
- Client stores approved/max units and API exposes remaining minutes/units.
- Creating/updating/deleting sessions or billable extra efforts updates remaining units within the same request (direct-call) or shortly after via event handler (event-driven).

### Notes / open questions
- Should “already conducted / scheduled minutes” include sessions with status `invoice-draft` and `billed`? Proposed: yes (they are not canceled).
- Should billed extra efforts still count against remaining units? Proposed: yes (they consume approved minutes regardless of billing status).


# Invoice Number Configured on and Unique in Organization 

# Client can delete bookings 

