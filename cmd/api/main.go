package main

import (
	"fmt"
	"main/configs"
	"main/internal/routers"

	"github.com/gofiber/fiber/v3"
)

func main() {
	err := configs.LoadConfig()
	if err != nil {
		fmt.Println("Cannot load config:", err)
		return
	}

	app := fiber.New()

	routers.SetupRouter(app)

	app.Listen(":3000")
}
