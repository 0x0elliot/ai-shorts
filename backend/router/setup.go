package router

import (
	"github.com/gofiber/fiber/v2"
	cors "github.com/gofiber/fiber/v2/middleware/cors"
	logger "github.com/gofiber/fiber/v2/middleware/logger"
)

var USER fiber.Router
var VIDEO fiber.Router
var BILLING fiber.Router

func SetupRoutes(app *fiber.App) {
	app.Use(logger.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*", // Change this to the allowed origins, e.g., "http://example.com"
		AllowMethods:     "GET,POST,PUT,DELETE",
		AllowHeaders:     "Content-Type, Authorization",
		AllowCredentials: true,
	}))

	api := app.Group("/api")

	USER = api.Group("/user")
	SetupUserRoutes()

	VIDEO = api.Group("/video")
	SetupVideoRoutes()

	BILLING = api.Group("/billing")
	SetupBillingRoutes()
}
