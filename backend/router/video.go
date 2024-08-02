package router

import (
	db "go-authentication-boilerplate/database"
	"go-authentication-boilerplate/models"
	auth "go-authentication-boilerplate/auth"
	util "go-authentication-boilerplate/util"
	"log"

	"github.com/gofiber/fiber/v2"
)

func SetupVideoRoutes(app fiber.Router) {
	VIDEO := app.Group("/video")

	privVideo := VIDEO.Group("/private")
	privVideo.Use(auth.SecureAuth()) // middleware to secure all routes for this group

	SCHEDULES := privVideo.Group("/schedules")
	SCHEDULES.Post("/create", CreateSchedule)
}

func CreateSchedule(c *fiber.Ctx) error {
	// topic,
	// description,
	// narrator: voice,
	// videoStyle,
	// postingSchedule,
	// isOneTime
	type CreateScheduleRequest struct {
		Topic string `json:"topic"`
		Description string `json:"description"`
		Narrator string `json:"narrator"`
		VideoStyle string `json:"videoStyle"`
		PostingSchedule string `json:"postingSchedule"`
		IsOneTime bool `json:"isOneTime"`
	}

	var req CreateScheduleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"message": "Invalid request",
		})
	}

	// vverify if narrator is valid
	narrators := []string{"p230", "p248", "p251", "p254", "p256", "p260", "p263", "p264", "p267", "p273", "p282", "p345"}
	if !util.Contains(narrators, req.Narrator) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"message": "Invalid narrator",
		})
	}

	// verify if videoStyle is valid
	videoStyles := []string{"default", "anime", "watercolor", "cartoon"}
	if !util.Contains(videoStyles, req.VideoStyle) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"message": "Invalid video style",
		})
	}

	// save video
	video := &models.Video{
		Topic: req.Topic,
		Description: req.Description,
		Narrator: req.Narrator,
		VideoStyle: req.VideoStyle,
		PostingSchedule: req.PostingSchedule,
		IsOneTime: req.IsOneTime,
	}

	video, err := util.SetVideo(video)
	if err != nil {
		log.Printf("[ERROR] Error creating schedule: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"message": "Error creating schedule",
		})
	}

	// start background job to create video
	go util.CreateVideo(video)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
		"video": video,
	})


}




