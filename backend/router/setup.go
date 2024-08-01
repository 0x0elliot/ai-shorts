package router

import (
	"github.com/gofiber/fiber/v2"
	cors "github.com/gofiber/fiber/v2/middleware/cors"
	logger "github.com/gofiber/fiber/v2/middleware/logger"
)

var USER fiber.Router

func SetupRoutes(app *fiber.App) {
	app.Use(logger.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*", // Change this to the allowed origins, e.g., "http://example.com"
		AllowMethods:     "GET,POST,PUT,DELETE",
		AllowHeaders:     "Content-Type, Authorization",
		AllowCredentials: true,
	}))

	// this is just for testing with shopify
	// remember that in production we will be using app proxies
	app.Static("/public", "./public", fiber.Static{
		Compress:  true,
		ByteRange: true,
	})

	api := app.Group("/api")

	USER = api.Group("/user")
	SetupUserRoutes()
}
