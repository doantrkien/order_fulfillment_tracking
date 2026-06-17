package integration

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"main/internal/ai"
	"main/internal/dto"
	"main/internal/handlers"
	"main/internal/models"
	"main/internal/repositories"
	routers "main/internal/routers/v1"
	"main/internal/services"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var (
	app               *fiber.App
	db                *gorm.DB
	adminToken        string
	driverToken       string
	testFakeAIAdapter *ai.FakeAIAdapter
)

func generateTestToken(userID int64, role string, email string) string {
	secret := []byte("test_jwt_secret_key_1234567890123456")
	claims := jwt.MapClaims{
		"sub":   strconv.FormatInt(userID, 10),
		"email": email,
		"role":  role,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(secret)
	return tokenString
}

func TestMain(m *testing.M) {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	os.Chdir(filepath.Join(basepath, "../../.."))

	os.Setenv("JWT_SECRET", "test_jwt_secret_key_1234567890123456")
	os.Setenv("JWT_EXPIRE_HOURS", "24")

	adminToken = generateTestToken(1, "admin", "admin@test.com")
	driverToken = generateTestToken(2, "driver", "driver@test.com")

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

	// Connect GORM to the test database with silent logger to avoid record not found noise
	db, err = gorm.Open(gormpostgres.Open(connStr), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.AutoMigrate(&models.Account{}, &models.Order{}, &models.OrderEvent{}, &models.Report{}); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	app = fiber.New()
	//log response request
	app.Use(logger.New())

	accountRepo := repositories.NewAccountRepository(db)
	authService := services.NewAuthService(accountRepo)
	authHandler := handlers.NewAuthHandler(authService)

	orderRepo := repositories.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepo)
	orderHandler := handlers.NewOrderHandler(orderService)

	reportRepo := repositories.NewReportRepository(db)
	reportService := services.NewReportService(reportRepo)
	reportHandler := handlers.NewReportHandler(reportService)

	routers.SetupAuthRouter(app, authHandler)
	routers.SetupOrderRouter(app, orderHandler)
	routers.SetupReportRouter(app, reportHandler)

	orderEventRepo := repositories.NewOrderEventRepository(db)
	orderEventService := services.NewOrderEventService(orderEventRepo, 4)
	orderEventHandler := handlers.NewOrderEventHandler(orderEventService)
	routers.SetupOrderEventRouter(app, orderEventHandler)

	// AI registration
	if err := db.AutoMigrate(&models.AIException{}); err != nil {
		log.Fatalf("Failed to auto-migrate AIException: %v", err)
	}

	aiRepo := repositories.NewAIRepository(db)
	expectedOutput := dto.ExceptionOutput{}
	expectedSummary := dto.ReportSummaryOutput{}
	testFakeAIAdapter = ai.NewFakeAIAdapter(expectedOutput, expectedSummary)

	aiAnalyzer := ai.NewExceptionAnalyzer(testFakeAIAdapter, ai.ExceptionAnalyzerConfig{
		AIEnabled: true,
		AITimeout: 5 * time.Second,
	})

	draftGenerator := ai.NewDraftGenerator(testFakeAIAdapter, ai.DraftGeneratorConfig{
		AIEnabled: true,
		AITimeout: 5 * time.Second,
	})

	aiDraftRepo := repositories.NewAIDraftRepository(db)

	aiEvalRepo := repositories.NewAIEvaluationRepository(db)

	aiService := services.NewAIService(aiRepo, aiAnalyzer, draftGenerator, aiDraftRepo, aiEvalRepo)
	aiHandler := handlers.NewAIHandler(aiService)
	routers.SetupAIRouter(app, aiHandler)

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
