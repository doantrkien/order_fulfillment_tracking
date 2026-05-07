package loaddata

import (
	"encoding/json"
	"main/internal/models"
	"os"
)

func LoadOrders() []models.Order {
	order_path := "/home/kiendoan/code/order_fulfillment_tracking/pkg/mock_data/orders.json"
	file, _ := os.ReadFile(order_path)

	var list []models.Order
	json.Unmarshal(file, &list)

	return list
}

func LoadInputs() []models.Order {
	input_path := "/home/kiendoan/code/order_fulfillment_tracking/pkg/mock_data/input.json"
	file, _ := os.ReadFile(input_path)

	var inputs []models.Order
	json.Unmarshal(file, &inputs)

	return inputs
}
