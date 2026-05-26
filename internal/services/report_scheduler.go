package services

import (
	"fmt"
	"time"
)

func StartDailyReportScheduler(reportService ReportService) {
	go func() {
		loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
		if err != nil {
			loc = time.UTC
		}

		for {
			now := time.Now().In(loc)
			nextRun := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, loc)
			if !nextRun.After(now) {
				nextRun = nextRun.Add(24 * time.Hour)
			}

			fmt.Printf("next daily report scheduled at: %s\n", nextRun.Format("2006-01-02 15:04:05"))
			time.Sleep(time.Until(nextRun))

			yesterday := nextRun.AddDate(0, 0, -1)
			report, err := reportService.CreateDailyReport(yesterday)
			if err != nil {
				fmt.Printf("failed to create daily report for %s: %v\n", yesterday.Format("2006-01-02"), err)
			} else {
				fmt.Printf("daily report created for %s: id=%d\n", yesterday.Format("2006-01-02"), report.ID)
			}
		}
	}()
}
