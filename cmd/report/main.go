package main

import (
	"flag"
	"log"
	"time"

	"main/configs"
	"main/internal/repositories"
	"main/internal/services"
	"main/pkg/postgresql"
)

func main() {
	dateStr := flag.String("date", "", "Date for the report in YYYY-MM-DD format (e.g. 2026-05-04)")
	flag.Parse()

	if *dateStr == "" {
		log.Fatal("Please provide a date using --date=YYYY-MM-DD")
	}

	date, err := time.Parse("2006-01-02", *dateStr)
	if err != nil {
		log.Fatalf("Invalid date format: %v. Please use YYYY-MM-DD.", err)
	}

	err = configs.LoadConfig()
	if err != nil {
		log.Fatalf("Cannot load config: %v", err)
	}

	db, err := postgresql.ConnectDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	reportRepo := repositories.NewReportRepository(db)
	reportService := services.NewReportService(reportRepo)

	log.Printf("Starting daily report generation for date: %s", date.Format("2006-01-02"))
	
	report, err := reportService.CreateDailyReport(date)
	if err != nil {
		log.Fatalf("Failed to create daily report: %v", err)
	}

	log.Printf("Successfully generated daily report:\nID=%d\nOrders=%d\nNew=%d\nDelivered=%d\nCancelled=%d\nRefunded=%d\nIncome=%d\nAvg Deliver Time=%.2f hours",
		report.ID, report.TotalOrders, report.TotalNew, report.TotalDelivered, report.TotalCancelled, report.TotalRefunded, report.TotalIncome, report.AvgDeliverTime)
}
