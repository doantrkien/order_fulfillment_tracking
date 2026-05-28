import json
import urllib.request
import urllib.error
import time
import random
from faker import Faker


fake = Faker()


url = "http://localhost:5000/api/v1/orders"
headers = {"Content-Type": "application/json", "X-API-Key": "8b2062e3c8c1292a47cb900ae480c2e642ae03c22157e311fec14fb40ba8d453"}


print(" Bắt đầu gọi API để seed 400 orders...")


success_count = 0
fail_count = 0


for i in range(1, 401):
    order_payload = {
        "username": fake.name(),
        "user_phone": "0123456789",
        "shipping_address": fake.address().replace("\n", ", "),
        "total_amount": random.randint(50000, 3000000),
    }

    # print(f"\n[{i}/400]  Đang gửi yêu cầu tạo đơn hàng: {order_payload['user_info']['username']} - {order_payload['total_amount']} VND")


    data_bytes = json.dumps(order_payload).encode('utf-8')
    req = urllib.request.Request(url, data=data_bytes, headers=headers, method='POST')


    try:
        response = urllib.request.urlopen(req)
        print(f"[{i}/400]  Thêm thành công đơn hàng! API Response: {response.read().decode().strip()}")
        success_count += 1
    except urllib.error.HTTPError as e:
        print(f"[{i}/400]  API báo lỗi (HTTP {e.code}): {e.read().decode().strip()}")
        fail_count += 1
    except Exception as e:
        print(f"[{i}/400]  Không thể kết nối tới API Server: {e}")
        fail_count += 1


    time.sleep(0.05)


print("\n-------------------------------------------")
print(f" KẾT QUẢ SEED DATA: Thành công: {success_count} | Thất bại: {fail_count}")
print("-------------------------------------------")



