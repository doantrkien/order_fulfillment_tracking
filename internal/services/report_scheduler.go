package services

import (
	"fmt"
	"time"
)

func StartDailyReportScheduler(reportService ReportService) {
	go func() {
		for {
			now := time.Now()
			nextRun := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
			if !nextRun.After(now) {
				nextRun = nextRun.Add(24 * time.Hour)
			}

			time.Sleep(time.Until(nextRun))

			report, err := reportService.CreateDailyReport(nextRun.AddDate(0, 0, -1))
			if err != nil {
				fmt.Printf("failed to create daily report at %s: %v\n", nextRun.Format("2006-01-02 15:04"), err)
			} else {
				fmt.Printf("daily report generated for period ending %s: report id=%d\n", nextRun.Format("2006-01-02 15:04"), report.ID)
			}
		}
	}()
}
