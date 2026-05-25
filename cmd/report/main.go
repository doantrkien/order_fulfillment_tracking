package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"main/configs"
	"main/pkg/postgresql"
)

func main() {
	if err := configs.LoadConfig(); err != nil {
		log.Printf("warning: load config: %v", err)
	}

	dateFlag := flag.String("date", "", "Report date in YYYY-MM-DD format")
	flag.Parse()

	if *dateFlag == "" {
		log.Fatal("missing required --date value")
	}

	reportDate, err := time.Parse("2006-01-02", *dateFlag)
	if err != nil {
		log.Fatalf("invalid date format: %v", err)
	}

	db, err := postgresql.ConnectDB()
	if err != nil {
		log.Fatalf("cannot connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("cannot get sql.DB from gorm: %v", err)
	}
	defer sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}

	fmt.Printf("Daily report generated for %s\n", reportDate.Format("2006-01-02"))

	var totalOrders, totalNew, totalDelivered, totalCancelled, totalRefunded int64
	var totalIncome, avgDeliverTime float64
	row := sqlDB.QueryRowContext(ctx, `
		SELECT total_orders, total_new, total_delivered, total_cancelled,
			total_refunded, total_income, avg_deliver_time
		FROM reports
		WHERE date = $1
	`, reportDate.Format("2006-01-02"))
	if err := row.Scan(&totalOrders, &totalNew, &totalDelivered, &totalCancelled, &totalRefunded, &totalIncome, &avgDeliverTime); err != nil {
		log.Fatalf("unable to fetch created report: %v", err)
	}

	fmt.Printf("report summary: orders=%d, new=%d, delivered=%d, cancelled=%d, refunded=%d, income=%.2f, avg_hours=%.2f\n",
		totalOrders, totalNew, totalDelivered, totalCancelled, totalRefunded, totalIncome, avgDeliverTime)
}
