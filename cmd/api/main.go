package main

import (
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	"main/configs"
	"main/internal/handlers"
	"main/internal/repositories"
	routers "main/internal/routers/v1"
	"main/internal/services"
<<<<<<< HEAD
=======
	"main/pkg/metrics"
>>>>>>> dev
	"main/pkg/postgresql"
	"net/http"

	_ "main/docs"

	"github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// @title Order Fulfillment Tracking API
// @version 1.0.0
// @description HTTP API for order fulfillment tracking.
// @host localhost:5000
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
func main() {
	err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Cannot load config: %v", err)
	}

	db, err := postgresql.ConnectDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
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
	maxWorkers, _ := strconv.Atoi(os.Getenv("IMPORT_MAX_WORKERS"))
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}
	orderEventService := services.NewOrderEventService(orderEventRepo, maxWorkers)
	orderEventHandler := handlers.NewOrderEventHandler(orderEventService)

	reportRepo := repositories.NewReportRepository(db)
	reportService := services.NewReportService(reportRepo)
	reportHandler := handlers.NewReportHandler(reportService)

	services.StartDailyReportScheduler(reportService)
	metrics.StartMemoryCollector(5 * time.Second) 

	routers.SetupOrderRouter(app, orderHandler)
	routers.SetupOrderEventRouter(app, orderEventHandler)
	routers.SetupReportRouter(app, reportHandler)
	app.Get("/docs/*", swaggo.HandlerDefault)

	log.Fatal(app.Listen(":5000"))
}
