package route

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"gitlab.smartbet.am/golang/smart-image/internal/service/handler"
	"gitlab.smartbet.am/golang/smart-image/internal/web/middleware"
	"go.uber.org/fx"
)

func NewApp(result *handler.Result, authMiddleware *middleware.AuthMiddleware) *fiber.App {
	app := fiber.New(fiber.Config{
		Prefork:       false,
		CaseSensitive: false,
		StrictRouting: true,
		ServerHeader:  "Fiber",
		AppName:       "Bat Server 1.0",
	})

	app.Static("/", "./static")

	app.Use(cors.New())

	api := app.Group("/api")
	v1 := api.Group("/v1")
	v1.Use(authMiddleware.Handle)
	v1.Post("/upload", result.Upload)

	return app
}

func StartServer(lc fx.Lifecycle, app *fiber.App) {
	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				go app.Listen(":8080")
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return app.Shutdown()
			},
		})
}
