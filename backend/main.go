package main

import (
	"log"
	util "go-authentication-boilerplate/util" 
	"context"

// 	"go-authentication-boilerplate/database"
// 	"go-authentication-boilerplate/router"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/gofiber/fiber/v2/middleware/cors"
)

// // CreateServer creates a new Fiber instance
// func CreateServer() *fiber.App {
// 	app := fiber.New()
// 	return app
// }

// func main() {
// 	// Connect to Postgres
// 	database.ConnectToDB()
// 	app := CreateServer()

// 	app.Use(cors.New())

// 	router.SetupRoutes(app)

// 	// 404 Handler
// 	app.Use(func(c *fiber.Ctx) error {
// 		return c.SendStatus(404) // => 404 "Not Found"
// 	})

// 	log.Println("[INFO] Server started on :5002")
// 	log.Fatal(app.Listen(":5002"))
// }

func main() {
	storageclient, _ := util.GetGCPClient()	
	log.Println(storageclient)

	ctx := context.Background()

	log.Printf("Stitching video..")

	_, err := util.StitchVideo(
		ctx,
		storageclient,
		"zappush_public",
		"22791df7-13d1-474b-9fca-62b908c1579e",
	)
	
	log.Println(err)
}

