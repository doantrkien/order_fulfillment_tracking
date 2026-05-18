# 📦 POST /api/v1/order-events/import


## 🧩 Framework
- Built with: Fiber (Go)
- Handler: `OrderEventHandler.ImportOrderEvents`


---


## 🎯 Task Objective
Implement the `POST /api/v1/order-events/import` API to process a batch of order status update events.

- Follow Clean Architecture
- Only modify/create files containing `event` (exception: `cmd/api/main.go` for DI wiring)
- Do NOT touch `order.go` or `report.go`


---


## 📥 Request Payload

```json
[
  {
    "order_id": 100156,
    "status": "packed",
    "event_at": "2026-05-17T09:00:00Z",
    "updated_by": "warehouse_staff_01"
  },
  {
    "order_id": 100157,
    "status": "shipped",
    "event_at": "2026-05-17T09:15:00Z",
    "updated_by": "driver_nguyen_van_a"
  }
]
```

> **Note**: The client does NOT send `previous_status`. The server must query the DB to get the current status of the order.


---


## 📤 Response Payload

All responses follow the shared `ResponseStruct` from `pkg/utils/response/api_response.go`:

```go
type ResponseStruct struct {
    Status  int         `json:"status"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}
```

**Success (200):**
```json
{
  "status": 200,
  "message": "batch processed",
  "data": {
    "accepted_count": 8,
    "rejected_count": 1,
    "duplicate_count": 1,
    "errors": [
      {
        "order_id": 100157,
        "status": "delivered",
        "reason": "Invalid transition from 'created' to 'delivered'"
      },
      {
        "order_id": 100158,
        "status": "packed",
        "reason": "Order is already in status 'packed'"
      }
    ]
  }
}
```

> `errors` contains entries for both **rejected** and **duplicate** events. `accepted` events do not appear here.

**Invalid payload (400):**
```json
{
  "status": 400,
  "message": "invalid payload",
  "data": null
}
```

**Internal error (500):**
```json
{
  "status": 500,
  "message": "<error detail>",
  "data": {
    "accepted_count": 2,
    "rejected_count": 1,
    "duplicate_count": 1,
    "errors": [
      {
        "order_id": 100157,
        "status": "delivered",
        "reason": "Invalid transition from 'created' to 'delivered'"
      },
      {
        "order_id": 100158,
        "status": "packed",
        "reason": "Order is already in status 'packed'"
      }
    ]
  }
}
```


---


## 🧠 Business Rules & Concurrency


### 1. Status Flow Validation

```
created -> paid -> packed -> shipped -> delivered
created -> cancelled
paid    -> refunded
```


---


### 2. State-based Evaluation (done inside DB transaction in Repo)

The **previous status is fetched from the DB** (not from the client), using `FOR UPDATE` row-level locking.

| Condition                                    | Result      |
|----------------------------------------------|-------------|
| `order_id` not found in `orders` table       | Rejected    |
| `new_status == orders.current_status`        | Duplicate   |
| Transition not in valid flow                 | Rejected    |
| Transition valid                             | Accepted    |


---


### 3. Responsibility Split

| Layer       | Responsibility                                                                 |
|-------------|--------------------------------------------------------------------------------|
| **Service** | Basic validation: required fields, `IsValidStatus()` checks known status value |
| **Repo**    | Business validation: fetch current status with `FOR UPDATE`, `IsValidTransition()` to validate flow, decide |


---


### 4. Concurrency Control

- Use **Worker Pool** with **7 goroutines** + channels in the service layer
- Each event runs in its **own isolated DB transaction** — partial failures do NOT affect others
- Row-level locking prevents race conditions on the same `order_id`:

```sql
SELECT current_status FROM orders WHERE id = $1 FOR UPDATE;
```


---


## 📁 Implementation Guide


### 1. `internal/dto/event.go`

Define the inbound request DTO and outbound response DTO.

```go
package dto

import "time"

type ImportOrderEventRequest struct {
    OrderID   int64     `json:"order_id"`
    Status    string    `json:"status"`
    EventAt   time.Time `json:"event_at"`
    UpdatedBy string    `json:"updated_by"`
}

// EventError holds the detail of a single rejected or duplicate event.
type EventError struct {
    OrderID int64  `json:"order_id"`
    Status  string `json:"status"`
    Reason  string `json:"reason"`
}

type ImportOrderEventsResponse struct {
    Accepted  int          `json:"accepted_count"`
    Rejected  int          `json:"rejected_count"`
    Duplicate int          `json:"duplicate_count"`
    Errors    []EventError `json:"errors"`
}
```

> The handler binds the body into `[]dto.ImportOrderEventRequest`, NOT `[]models.OrderEvent`.
> Mapping from DTO → `models.OrderEvent` happens in the service layer.


---


### 2. `internal/models/order_event.go`

Add `UpdatedBy` field to track who triggered the event.

```go
type OrderEvent struct {
    ID             int64       `gorm:"primaryKey;column:id" json:"id"`
    OrderID        int64       `gorm:"column:order_id;not null;index" json:"order_id"`
    Order          Order       `gorm:"foreignKey:OrderID" json:"-"`
    PreviousStatus OrderStatus `gorm:"column:previous_status;not null" json:"previous_status"`
    NewStatus      OrderStatus `gorm:"column:new_status;not null" json:"new_status"`
    UpdatedBy      string      `gorm:"column:updated_by;not null" json:"updated_by"`
    EventAt        time.Time   `gorm:"column:event_at;not null;default:CURRENT_TIMESTAMP;index" json:"event_at"`
    CreatedAt      time.Time   `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
    UpdatedAt      time.Time   `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
```


---


### 3. `internal/models/order_event_validator.go`

Add `IsValidStatus` for service-layer validation. Keep existing `IsValidTransition` for repo-layer validation.

```go
// IsValidStatus checks if a status string is a known OrderStatus value.
// Used by the service layer for basic input validation.
func IsValidStatus(s OrderStatus) bool {
    switch s {
    case ORDER_STATUS_CREATED, ORDER_STATUS_PAID, ORDER_STATUS_PACKED,
        ORDER_STATUS_SHIPPED, ORDER_STATUS_DELIVERED, ORDER_STATUS_CANCELLED,
        ORDER_STATUS_REFUNDED:
        return true
    }
    return false
}

// IsValidTransition checks if a state transition is allowed.
// Used by the repo layer after fetching current status from DB.
func IsValidTransition(prev, next OrderStatus) bool
```


---


### 4. `internal/repositories/event.go`

Define the interface and implement `ProcessSingleEventTx`.

```go
type ProcessResult string

const (
    Accepted  ProcessResult = "accepted"
    Rejected  ProcessResult = "rejected"
    Duplicate ProcessResult = "duplicate"
)

// ProcessResultDetail carries the outcome AND a human-readable reason.
// Reason is empty when Result == Accepted.
type ProcessResultDetail struct {
    Result ProcessResult
    Reason string
}

type OrderEventRepository interface {
    ProcessSingleEventTx(event models.OrderEvent) (ProcessResultDetail, error)
}
```

**Reason strings to use:**

| Case           | Reason format |
|----------------|---------------|
| Not Found      | `"Order not found"` |
| Duplicate      | `"Order is already in status '<current_status>'"` |
| Rejected       | `"Invalid transition from '<current>' to '<new>'"` |

**Transaction flow inside `ProcessSingleEventTx`:**

```sql
BEGIN;

-- 1. Lock the order row
SELECT current_status FROM orders WHERE id = $1 FOR UPDATE;
--    order not found               → return Rejected("Order not found"), ROLLBACK

-- 2. Compare using IsValidTransition(prevStatus, nextStatus)
--    current_status == new_status  → return Duplicate, ROLLBACK
--    !IsValidTransition(prev, new) → return Rejected,  ROLLBACK
--    valid transition              → continue

-- 3. Update order status
UPDATE orders SET current_status = $2 WHERE id = $1;

-- 4. Insert the event record (with previous_status = fetched current_status)
INSERT INTO order_events (order_id, previous_status, new_status, updated_by, event_at, ...)
VALUES ($1, $current, $new, $updated_by, $event_at, ...);

COMMIT;
```

> Each call is a **self-contained transaction**. A failure on one event does NOT roll back others.


---


### 5. `internal/services/event.go`

```go
type OrderEventService interface {
    ImportOrderEvents(reqs []dto.ImportOrderEventRequest) (dto.ImportOrderEventsResponse, error)
}
```

**Service responsibilities:**

1. **Basic validation** per event (before sending to worker):
   - `order_id > 0`
   - `status` is a known `OrderStatus` value → use `models.IsValidStatus(OrderStatus(req.Status))`
   - `event_at` is not zero
   - Count invalid ones as `rejected` immediately

2. **Worker Pool (7 goroutines)** — dispatch valid events concurrently:
   ```go
   const maxWorkers = 7

   jobs    chan models.OrderEvent     // input channel
   results chan workerResult          // output channel
   ```
   Where `workerResult` is a local struct pairing the original request with the repo outcome:
   ```go
   type workerResult struct {
       req    dto.ImportOrderEventRequest
       detail repositories.ProcessResultDetail
       err    error
   }
   ```

3. Map `dto.ImportOrderEventRequest` → `models.OrderEvent` before sending to the repo:
   - `OrderID`   ← req.OrderID
   - `NewStatus`  ← OrderStatus(req.Status)
   - `UpdatedBy`  ← req.UpdatedBy
   - `EventAt`   ← req.EventAt
   - `PreviousStatus` is left zero — the repo fetches it from DB

4. **Aggregate** results from all workers into `dto.ImportOrderEventsResponse`:
   - `Accepted++` if `Result == Accepted`
   - `Rejected++` + append to `Errors` if `Result == Rejected`
   - `Duplicate++` + append to `Errors` if `Result == Duplicate`
   - `EventError{OrderID: req.OrderID, Status: req.Status, Reason: detail.Reason}`


---


### 6. `internal/handlers/order_event.go`

```go
package handlers

import (
    "main/internal/dto"
    "main/internal/services"
    "main/pkg/utils/response"

    "github.com/gofiber/fiber/v3"
)

type OrderEventHandler struct {
    orderEventService services.OrderEventService
}

func NewOrderEventHandler(orderEventService services.OrderEventService) *OrderEventHandler {
    return &OrderEventHandler{orderEventService: orderEventService}
}

func (h *OrderEventHandler) ImportOrderEvents(c fiber.Ctx) error {
    var req []dto.ImportOrderEventRequest

    if err := c.Bind().Body(&req); err != nil {
        return response.Reponse(c, 400, "invalid payload", nil)
    }

    result, err := h.orderEventService.ImportOrderEvents(req)
    if err != nil {
        return response.Reponse(c, 500, err.Error(), result)
    }

    return response.Reponse(c, 200, "batch processed", result)
}
```


---


### 7. `internal/routers/v1/event.go`

```go
func SetupOrderEventRouter(app *fiber.App, orderEventHandler *handlers.OrderEventHandler) {
    orderEvent := app.Group("/api/v1/order-events")
    orderEvent.Post("/import", orderEventHandler.ImportOrderEvents)
}
```


---


### 8. `cmd/api/main.go` — Uncomment DI Wiring

Uncomment the existing lines to wire up the order event layer:

```go
orderEventRepo := repositories.NewOrderEventRepository(db)
orderEventService := services.NewOrderEventService(orderEventRepo)
orderEventHandler := handlers.NewOrderEventHandler(orderEventService)

// ... and the router:
routers.SetupOrderEventRouter(app, orderEventHandler)
```

> These lines already exist in `main.go` but are commented out. Just uncomment them.
