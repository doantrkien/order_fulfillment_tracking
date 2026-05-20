package integration

import (
	"log"
	"main/internal/handlers"
	"main/internal/models"
	"main/internal/repositories"
	routers "main/internal/routers/v1"
	"main/internal/services"
	"main/pkg/postgresql"
	"os"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

var (
	app *fiber.App
	db  *gorm.DB
)

func TestMain(m *testing.M) {

	if err := godotenv.Load("../../../.env.local"); err != nil {
		log.Println("No .env file found, using process environment")
	}

	if os.Getenv("DB_HOST") == "" {
		log.Println("DB_HOST not set, skipping integration tests")
		os.Exit(0)
	}

	os.Setenv("CUSTOMER_API_KEY", "54725cc28e71b4d43646e3697affd2e53d01f502b9f04ccb43a665a83ac2d418")
	os.Setenv("ADMIN_API_KEY", "8b2062e3c8c1292a47cb900ae480c2e642ae03c22157e311fec14fb40ba8d453")
	os.Setenv("SHIPPER_API_KEY", "848cf386682fde792d5a5a7c51588b92e8fe1e9e339b130c7b7e13a104839298")

	var err error
	db, err = postgresql.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.AutoMigrate(&models.Order{}, &models.OrderEvent{}, &models.Report{}); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	app = fiber.New()
	//log response request
	app.Use(logger.New())
	orderRepo := repositories.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepo)
	orderHandler := handlers.NewOrderHandler(orderService)

	reportRepo := repositories.NewReportRepository(db)
	reportService := services.NewReportService(reportRepo)
	reportHandler := handlers.NewReportHandler(reportService)

	routers.SetupOrderRouter(app, orderHandler)
	routers.SetupReportRouter(app, reportHandler)

	orderEventRepo := repositories.NewOrderEventRepository(db)
	orderEventService := services.NewOrderEventService(orderEventRepo, 4)
	orderEventHandler := handlers.NewOrderEventHandler(orderEventService)
	routers.SetupOrderEventRouter(app, orderEventHandler)

	code := m.Run()

	sqlDB, _ := db.DB()
	sqlDB.Close()

	os.Exit(code)
}

func cleanOrders() {
	db.Exec("TRUNCATE TABLE orders RESTART IDENTITY CASCADE;")
	db.Exec("TRUNCATE TABLE reports RESTART IDENTITY CASCADE;")
}

func cleanOrderEvents() {
	db.Exec("TRUNCATE TABLE order_events RESTART IDENTITY CASCADE;")
}

func cleanAll() {
	db.Exec("TRUNCATE TABLE order_events RESTART IDENTITY CASCADE;")
	db.Exec("TRUNCATE TABLE orders RESTART IDENTITY CASCADE;")
}
