package main

import (
	"fmt"
	"main/configs"
	"main/internal/handlers"
	"main/internal/repositories"
	routers "main/internal/routers/v1"
	"main/internal/services"
	"main/internal/swagger"
	_ "main/pkg/metrics"
	"main/pkg/postgresql"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	err := configs.LoadConfig()
	if err != nil {
		fmt.Println("Cannot load config:", err)
		return
	}

	db, err := postgresql.ConnectDB()
	if err != nil {
		fmt.Printf("Error initializing database: %v", err)
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":9091", nil)
	}()

	app := fiber.New()

	orderRepo := repositories.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepo)
	orderHandler := handlers.NewOrderHandler(orderService)

	orderEventRepo := repositories.NewOrderEventRepository(db)
	orderEventService := services.NewOrderEventService(orderEventRepo)
	orderEventHandler := handlers.NewOrderEventHandler(orderEventService)

	routers.SetupOrderRouter(app, orderHandler)
	routers.SetupOrderEventRouter(app, orderEventHandler)
	swagger.SetupSwaggerRoutes(app)

	reportRepo := repositories.NewReportRepository(db)
	reportService := services.NewReportService(reportRepo)
	reportHandler := handlers.NewReportHandler(reportService)

	services.StartDailyReportScheduler(reportService)
	routers.SetupReportRouter(app, reportHandler)

	app.Listen(":3000")
}
