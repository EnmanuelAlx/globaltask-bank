# Add Workflow Event Skill

## Purpose
Automate the creation of new event types and their handlers in the event-outbox workflow system. This skill scaffolds all necessary code to add a new event to the worker processing pipeline.

## When to Use
- User wants to add a new event type (e.g., `SEND_NOTIFICATION`, `CALCULATE_SCORE`, `SYNC_EXTERNAL`)
- User wants to extend the workflow with new processing steps
- User describes a new background task that should be handled by the worker

## Prerequisites
- Event-outbox pattern knowledge
- Go programming
- Project uses: Go + PostgreSQL + Event Outbox + Worker Pool

---

## Workflow to Execute

### Step 1: Validate Event Name
The skill will ask or parse the desired event name.

**Naming Convention:**
- Format: `ACTION_TARGET` (uppercase with underscores)
- Examples: `SEND_APPROVAL_EMAIL`, `CALCULATE_CREDIT_SCORE`, `SYNC_EXTERNAL_DATA`

If the name doesn't follow convention, suggest a better one.

### Step 2: Identify Existing Files
The skill must modify these files (verify they exist):

| File | Purpose |
|------|---------|
| `backend/internal/domain/entity/entities.go` | Event constant definition |
| `backend/internal/domain/workflow/workflow.go` | Handler implementation |
| `backend/internal/infrastructure/worker/worker.go` | Event routing |
| `backend/config/workflows.yaml` | Workflow configuration |

### Step 3: Add Event Constant
Read `entities.go` and find where event constants are defined (search for `EventType` or `Event*` constants).

**Insert pattern:**
```go
const (
    // ... existing events
    EventEvaluateApplicationRisk EventType = "EVALUATE_APPLICATION_RISK"
    
    // NEW EVENT:
    Event{EventName} EventType = "{EVENT_NAME}"
)
```

### Step 4: Create Handler Function
Read `workflow.go` and understand existing handler patterns:

**Template:**
```go
func (e *WorkflowEngine) Handle{EventName}(
    ctx context.Context,
    uow repository.UnitOfWork,
    event *entity.EventOutbox,
) error {
    // 1. Extract application_id from payload
    payload := event.Payload
    appIDStr, ok := payload["application_id"].(string)
    if !ok {
        return fmt.Errorf("invalid payload: missing application_id")
    }

    appID, err := uuid.Parse(appIDStr)
    if err != nil {
        return fmt.Errorf("invalid application_id: %w", err)
    }

    // 2. Load application (or skip if not needed)
    app, err := uow.LoanApplications().GetByID(ctx, appID)
    if err != nil {
        return fmt.Errorf("failed to get application: %w", err)
    }
    if app == nil {
        return fmt.Errorf("application not found: %s", appID)
    }

    // 3. TODO: Implement business logic
    // - Call external APIs
    // - Update records
    // - Send notifications, etc.

    // 4. Get next step from workflow config (or end)
    profile, err := uow.Profiles().GetByID(ctx, app.UserID)
    if err != nil {
        return fmt.Errorf("failed to get profile: %w", err)
    }

    country, err := e.countryRepo.GetByID(ctx, profile.CountryID)
    if err != nil {
        return fmt.Errorf("failed to get country: %w", err)
    }

    nextStep := e.GetNextStep(country.ISOCode, string(event.EventType))
    
    if nextStep != nil {
        // 5. Create next event to continue workflow
        newEvent := &entity.EventOutbox{
            ID:        uuid.New(),
            EventType: entity.EventType(*nextStep),
            Payload:   entity.JSONB{"application_id": app.ID.String()},
            Status:    entity.EventStatusPending,
            CreatedAt: time.Now(),
        }
        return uow.EventOutbox().Create(ctx, newEvent)
    }

    // No next step = workflow ends for this country
    return nil
}
```

### Step 5: Register Handler in Worker
Read `worker.go` and find the switch statement in `processOne()`.

**Insert pattern:**
```go
switch event.EventType {
case entity.EventLoanApplicationCreated:
    err = wp.workflowEngine.HandleLoanApplicationCreated(ctx, uow, event)
case entity.EventFetchBankData:
    err = wp.workflowEngine.HandleFetchBankData(ctx, uow, event)
// ... existing cases

// NEW CASE:
case entity.Event{EventName}:
    err = wp.workflowEngine.Handle{EventName}(ctx, uow, event)

default:
    log.Printf("Unknown event type: %s", event.EventType)
    return fmt.Errorf("unknown event type: %s", event.EventType)
}
```

### Step 6: Update Workflow Config
Read `config/workflows.yaml` and add the new event to country workflows.

**Insert pattern:**
```yaml
workflows:
  PT:
    - event: EVALUATE_APPLICATION_RISK
      next: {EVENT_NAME}
    - event: {EVENT_NAME}
      next: null  # End of workflow

  CO:
    # Add to Colombia workflow as well
```

---

## Output Format

The skill returns a structured summary:

```markdown
## Event Created: {EVENT_NAME}

### Files Modified:
1. `backend/internal/domain/entity/entities.go` - Added constant
2. `backend/internal/domain/workflow/workflow.go` - Added handler
3. `backend/internal/infrastructure/worker/worker.go` - Registered handler
4. `backend/config/workflows.yaml` - Updated workflow config

### Handler Location:
`workflow.go:Handle{EventName}` (lines {XX}-{XX})

### Next Steps:
- Implement the business logic in the handler
- Add unit tests
- Consider adding database migrations if new tables/columns needed
```

---

## Important Patterns

### Payload Convention
Always use `application_id` as the primary key in payload:
```go
entity.JSONB{
    "application_id": app.ID.String(),
}
```

### Transaction Safety
- All DB operations MUST go through `uow.Do()` or pass the `uow` to handlers
- Event creation happens inside the transaction
- Return error to trigger rollback

### Error Handling
- Use descriptive errors with context
- Application not found = stop workflow (return error)
- External API failure = consider retry or fallback

### Skipping Workflow Step
If an event should be skipped for a specific country, just don't add it to that country's workflow in `workflows.yaml`. The `GetNextStep()` will return `nil` and workflow ends.

---

## Example Conversation

**User:** "Quiero agregar un evento para enviar email de aprobación"

**Skill Response:**
1. Validates name: `SEND_APPROVAL_EMAIL` ✓
2. Adds constant to `entities.go`
3. Creates handler skeleton in `workflow.go`
4. Registers in `worker.go`
5. Adds to `workflows.yaml` for PT and CO
6. Returns summary with `// TODO` markers for business logic

---

## File Paths Reference

```
backend/
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   └── entities.go          # Event constants
│   │   └── workflow/
│   │       └── workflow.go         # Handler implementations
│   └── infrastructure/
│       ├── worker/
│       │   └── worker.go           # Event routing
│       └── repository/
│           └── event_outbox.go     # Event repository
└── config/
    └── workflows.yaml              # Workflow configuration
```
