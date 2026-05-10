package main

import (
	"fmt"
	"main/configs"
	"main/internal/handlers"
	"main/internal/repositories"
	"main/internal/routers"
	"main/internal/services"
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

	routers.SetupRouter(app)
	routers.SetupOrderRouter(app, orderHandler)

	app.Listen(":3000")
}
