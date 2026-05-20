# Order Fulfillment Tracking System

## Reporting (Week 8)

This project supports daily report generation and retrieval.

### Run report ETL from CLI

Generate or refresh the daily report row for a specific date:

```bash
go run ./cmd/report --date=2026-05-04
```

The command reads raw `orders` and `order_events` data, calculates:

- total orders
- total new orders
- total delivered orders
- total cancelled orders
- total refunded orders
- total income
- average delivery time (hours)

Then it writes or updates the corresponding row in the `reports` table.

### Fetch report via API

Start the API server and request:

```bash
go run ./cmd/api/main.go
```

Then fetch report data:

```bash
curl "http://localhost:3000/api/v1/reports/daily?date=2026-05-04"
```

The report endpoint returns stored report data for the requested date.
