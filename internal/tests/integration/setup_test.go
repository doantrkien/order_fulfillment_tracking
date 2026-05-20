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

	if err := godotenv.Load("../../../.env"); err != nil {
		log.Println("No .env file found, using process environment")
	}

	if os.Getenv("DB_HOST") == "" {
		log.Println("DB_HOST not set, skipping integration tests")
		os.Exit(0)
	}

	os.Setenv("CUSTOMER_API_KEY", "test-customer-key")
	os.Setenv("ADMIN_API_KEY", "test-admin-key")

	var err error
	db, err = postgresql.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.AutoMigrate(&models.Order{}, &models.OrderEvent{}); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	app = fiber.New()
	//log response request
	app.Use(logger.New())
	orderRepo := repositories.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepo)
	orderHandler := handlers.NewOrderHandler(orderService)
	routers.SetupOrderRouter(app, orderHandler)

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
}

func cleanOrderEvents() {
	db.Exec("TRUNCATE TABLE order_events RESTART IDENTITY CASCADE;")
}

func cleanAll() {
	db.Exec("TRUNCATE TABLE order_events RESTART IDENTITY CASCADE;")
	db.Exec("TRUNCATE TABLE orders RESTART IDENTITY CASCADE;")
}
