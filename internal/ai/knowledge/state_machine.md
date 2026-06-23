# Order State Machine

Valid order statuses: created, paid, packed, shipped, delivered, cancelled, refunded

Valid state transitions:
  created  → paid, cancelled
  paid     → packed, refunded
  packed   → shipped
  shipped  → delivered

Terminal states (no further transitions): delivered, cancelled, refunded

Any transition not listed above is an INVALID_TRANSITION (always CRITICAL).
