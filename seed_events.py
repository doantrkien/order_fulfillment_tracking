import json
import urllib.request
import urllib.error
import time
from datetime import datetime, timezone

statuses = ["paid", "packed", "shipped", "delivered"]
url = "http://localhost:3000/api/v1/order-events/import"
headers = {"Content-Type": "application/json"}

# 50 valid order IDs from the database
order_ids = [
    100174, 100173, 100172, 100171, 100170, 100169, 100168, 100167, 100166, 100165, 
    100164, 100163, 100162, 100161, 100160, 100159, 100158, 100157, 100156, 100155, 
    100154, 100153, 100152, 100151, 100150, 100149, 100148, 100147, 100146, 100145, 
    100144, 100143, 100142, 100141, 100140, 100139, 100138, 100137, 100136, 100135, 
    100134, 100133, 100132, 100131, 100130, 100129, 100128, 100127, 100126, 100125
]

for status in statuses:
    events = []
    for oid in order_ids:
        events.append({
            "order_id": oid,
            "status": status,
            "event_at": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
            "updated_by": "system"
        })
    
    req = urllib.request.Request(url, data=json.dumps(events).encode('utf-8'), headers=headers, method='POST')
    try:
        response = urllib.request.urlopen(req)
        print(f"Batch {status} response:", response.read().decode())
    except urllib.error.HTTPError as e:
        print(f"Batch {status} failed:", e.read().decode())
    
    # Wait a bit between batches to allow worker to process
    time.sleep(2)
