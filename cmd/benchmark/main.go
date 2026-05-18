package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"net/http"
)

type ImportOrderEventRequest struct {
	OrderID   int64     `json:"order_id"`
	Status    string    `json:"status"`
	EventAt   time.Time `json:"event_at"`
	UpdatedBy string    `json:"updated_by"`
}

func main() {
	_ = godotenv.Load()

	totalEvents := 100_000
	if len(os.Args) > 1 {
		fmt.Sscanf(os.Args[1], "%d", &totalEvents)
	}

	fmt.Printf("=== Benchmark: %d order events ===\n\n", totalEvents)

	// ── 1. Connect to DB and seed orders ──
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_DATABASE"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(20)

	// Clean up previous benchmark data
	fmt.Print("Cleaning previous data... ")
	db.Exec("DELETE FROM order_events")
	db.Exec("DELETE FROM orders")
	db.Exec("ALTER SEQUENCE orders_id_seq RESTART WITH 1")
	db.Exec("ALTER SEQUENCE order_events_id_seq RESTART WITH 1")
	fmt.Println("done")

	// Seed orders in batches
	fmt.Printf("Seeding %d orders... ", totalEvents)
	seedStart := time.Now()

	batchSize := 5000
	for i := 0; i < totalEvents; i += batchSize {
		end := i + batchSize
		if end > totalEvents {
			end = totalEvents
		}
		count := end - i

		// Use raw SQL for speed
		query := "INSERT INTO orders (user_info, total_amount, current_status, created_at, updated_at) VALUES "
		for j := 0; j < count; j++ {
			if j > 0 {
				query += ","
			}
			query += fmt.Sprintf(
				`('{"customer_name":"bench_%d","customer_phone":"0900000000","shipping_addr":"bench"}', %d, 'created', NOW(), NOW())`,
				i+j, 100000+int64(i+j),
			)
		}
		if err := db.Exec(query).Error; err != nil {
			log.Fatalf("Seed failed at batch %d: %v", i, err)
		}
	}
	fmt.Printf("done in %v\n", time.Since(seedStart))

	// ── 2. Build 100K events payload ──
	fmt.Printf("Building %d event payloads... ", totalEvents)
	buildStart := time.Now()

	events := make([]ImportOrderEventRequest, totalEvents)
	now := time.Now()
	for i := 0; i < totalEvents; i++ {
		events[i] = ImportOrderEventRequest{
			OrderID:   int64(i + 1),
			Status:    "paid",
			EventAt:   now,
			UpdatedBy: "benchmark",
		}
	}

	payload, err := json.Marshal(events)
	if err != nil {
		log.Fatalf("JSON marshal failed: %v", err)
	}
	fmt.Printf("done in %v (payload size: %.1f MB)\n", time.Since(buildStart), float64(len(payload))/1024/1024)

	// ── 3. Send to API and measure ──
	fmt.Printf("\nSending POST /api/v1/order-events/import ...\n")
	apiStart := time.Now()

	resp, err := http.Post(
		"http://localhost:3000/api/v1/order-events/import",
		"application/json",
		bytes.NewReader(payload),
	)
	apiDuration := time.Since(apiStart)

	if err != nil {
		log.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	// ── 4. Print results ──
	fmt.Printf("\n=== Results ===\n")
	fmt.Printf("HTTP Status:    %d\n", resp.StatusCode)
	fmt.Printf("API Duration:   %v\n", apiDuration)
	fmt.Printf("Throughput:     %.0f events/sec\n", float64(totalEvents)/apiDuration.Seconds())

	if data, ok := result["data"].(map[string]interface{}); ok {
		fmt.Printf("Accepted:       %.0f\n", data["accepted_count"])
		fmt.Printf("Rejected:       %.0f\n", data["rejected_count"])
		fmt.Printf("Duplicate:      %.0f\n", data["duplicate_count"])
	}

	fmt.Printf("Message:        %v\n", result["message"])
}
