package integration

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"main/internal/handlers"
	"main/internal/models"
	"main/internal/repositories"
	routers "main/internal/routers/v1"
	"main/internal/services"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
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

	// Hardcode API keys for testing to avoid depending on .env
	customerAPIKey = "test_customer_key"
	adminAPIKey = "test_admin_key"
	driverAPIKey = "test_driver_key"

	os.Setenv("CUSTOMER_API_KEY", customerAPIKey)
	os.Setenv("ADMIN_API_KEY", adminAPIKey)
	os.Setenv("DRIVER_API_KEY", driverAPIKey)

	ctx := context.Background()

	// Spin up postgres container
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_pass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(20*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("failed to start postgres container: %v", err)
	}

	// Clean up the container after tests
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %v", err)
		}
	}()

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	// Connect GORM to the test database
	db, err = gorm.Open(gormpostgres.Open(connStr), &gorm.Config{})
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
