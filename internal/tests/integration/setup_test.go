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
	"path/filepath"
	"runtime"
	"testing"

	"main/configs"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"gorm.io/gorm"
)

var (
	app            *fiber.App
	db             *gorm.DB
	customerAPIKey string
	adminAPIKey    string
	driverAPIKey   string
)

func TestMain(m *testing.M) {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	os.Chdir(filepath.Join(basepath, "../../.."))

	if err := configs.LoadConfig(); err != nil {
		log.Println("LoadConfig error:", err)
	}

	customerAPIKey = os.Getenv("CUSTOMER_API_KEY")
	adminAPIKey = os.Getenv("ADMIN_API_KEY")
	driverAPIKey = os.Getenv("DRIVER_API_KEY")

	var err error
	db, err = postgresql.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	db.Exec("DROP TABLE IF EXISTS reports CASCADE")
	db.Exec("DROP TABLE IF EXISTS order_events CASCADE")

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
