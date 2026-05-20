package main

import (
	"log"
	"main/configs"
	"main/internal/handlers"
	"main/internal/repositories"
	"main/internal/routers/v1"
	"main/internal/services"
	"main/internal/swagger"
	"main/pkg/postgresql"

	"github.com/gofiber/fiber/v3"
)

func main() {
	// 1. Load config
	err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Cannot load config: %v", err)
	}

	// 2. Initialize DB (exit on failure)
	db, err := postgresql.ConnectDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	app := fiber.New()

	// 3. Dependency Injection (Wiring)
	orderRepo := repositories.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepo)
	orderHandler := handlers.NewOrderHandler(orderService)

	orderEventRepo := repositories.NewOrderEventRepository(db)
	orderEventService := services.NewOrderEventService(orderEventRepo, 7)
	orderEventHandler := handlers.NewOrderEventHandler(orderEventService)

	reportRepo := repositories.NewReportRepository(db)
	reportService := services.NewReportService(reportRepo)
	reportHandler := handlers.NewReportHandler(reportService)

	// 4. Background tasks
	services.StartDailyReportScheduler(reportService)

	// 5. Setup Routers
	routers.SetupOrderRouter(app, orderHandler)
	routers.SetupOrderEventRouter(app, orderEventHandler)
	routers.SetupReportRouter(app, reportHandler)
	swagger.SetupSwaggerRoutes(app)

	// 6. Start Server
	log.Fatal(app.Listen(":3000"))
}
