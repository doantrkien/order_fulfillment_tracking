package swagger

import "github.com/gofiber/fiber/v3"

// SetupSwaggerRoutes registers the Swagger UI and spec endpoints.
// - GET /swagger       → Swagger UI (served from CDN)
// - GET /swagger/spec  → Raw OpenAPI YAML spec
func SetupSwaggerRoutes(app *fiber.App) {
	// Serve the raw YAML spec
	app.Get("/swagger/spec", func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/x-yaml")
		return c.Send(Spec)
	})

	// Serve Swagger UI via CDN
	app.Get("/swagger", func(c fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(swaggerHTML)
	})
}

const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Order Fulfillment API - Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    html { box-sizing: border-box; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: "/swagger/spec",
      dom_id: "#swagger-ui",
      presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIBundle.SwaggerUIStandalonePreset
      ],
      layout: "BaseLayout",
      deepLinking: true,
    });
  </script>
</body>
</html>`
