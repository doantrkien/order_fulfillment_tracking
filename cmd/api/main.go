package main

import (
	"fmt"
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
	err := configs.LoadConfig()
	if err != nil {
		fmt.Println("Cannot load config:", err)
		return
	}

	db, err := postgresql.ConnectDB()
	if err != nil {
		fmt.Printf("Error initializing database: %v", err)
	}

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

	app.Listen(":3000")
}
