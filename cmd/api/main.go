package main

import (
	"log"
	"os"
	"runtime"
	"strconv"
	_ "time/tzdata"

	"main/configs"
	"main/internal/handlers"
	"main/internal/repositories"
	routers "main/internal/routers/v1"
	"main/internal/services"
	"main/pkg/postgresql"

	_ "main/docs"

	"github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
)

// @title Order Fulfillment Tracking API
// @version 1.0.0
// @description HTTP API for order fulfillment tracking.
// @host localhost:5000
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.
func main() {
	err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Cannot load config: %v", err)
	}

	db, err := postgresql.ConnectDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024,
	})

	accountRepo := repositories.NewAccountRepository(db)
	authService := services.NewAuthService(accountRepo)
	authHandler := handlers.NewAuthHandler(authService)

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

	routers.SetupAuthRouter(app, authHandler)
	routers.SetupOrderRouter(app, orderHandler)
	routers.SetupOrderEventRouter(app, orderEventHandler)
	routers.SetupReportRouter(app, reportHandler)
	app.Get("/docs/swagger/*", swaggo.HandlerDefault)

	log.Fatal(app.Listen(":5000"))
}
