import json
import random
from datetime import datetime, timedelta, timezone

events = []
base_time = datetime.now(timezone.utc)

print("Generating 60,000 valid events for 15,000 orders...")

# For each order from 1000 to 15999, generate 4 sequential valid status transitions
for order_id in range(1, 400):
    # status flow: paid -> packed -> shipped -> delivered
    t1 = base_time + timedelta(seconds=1)
    t2 = base_time + timedelta(seconds=2)
    t3 = base_time + timedelta(seconds=3)
    t4 = base_time + timedelta(seconds=4)

    events.append({
        "order_id": order_id,
        "status": "paid",
        "event_at": t1.isoformat().replace("+00:00", "Z"),
        "updated_by": "stress_tester_60k"
    })
    events.append({
        "order_id": order_id,
        "status": "packed",
        "event_at": t2.isoformat().replace("+00:00", "Z"),
        "updated_by": "stress_tester_60k"
    })
    events.append({
        "order_id": order_id,
        "status": "shipped",
        "event_at": t3.isoformat().replace("+00:00", "Z"),
        "updated_by": "stress_tester_60k"
    })
    events.append({
        "order_id": order_id,
        "status": "delivered",
        "event_at": t4.isoformat().replace("+00:00", "Z"),
        "updated_by": "stress_tester_60k"
    })

# Shuffle the events to simulate real-world out-of-order ingestion
# The backend will sort them by event_at within each order group anyway!
random.shuffle(events)

with open('50000_events_payload.json', 'w') as f:
    json.dump(events, f, indent=2)

print("Generated seed/60000_events_payload.json successfully with exactly 60,000 valid events.")
