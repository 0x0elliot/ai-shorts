package router

import (
	// db "go-authentication-boilerplate/database"
	"go-authentication-boilerplate/models"
	auth "go-authentication-boilerplate/auth"
	util "go-authentication-boilerplate/util"
	"log"

	"github.com/gofiber/fiber/v2"
)

func SetupVideoRoutes() {
	privVideo := VIDEO.Group("/private")
	privVideo.Use(auth.SecureAuth()) // middleware to secure all routes for this group

	privVideo.Get("/:id", GetVideo)
	privVideo.Post("/create", CreateSchedule)
}

func GetVideo(c *fiber.Ctx) error {
	id := c.Params("id")
	video, err := util.GetVideoById(id)
	if err != nil {
		log.Printf("[ERROR] Error getting video: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"message": "Error getting video",
		})
	}

	if video.OwnerID != c.Locals("id") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": true,
			"message": "Unauthorized",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
		"video": video,
	})
}


func CreateSchedule(c *fiber.Ctx) error {
	type CreateScheduleRequest struct {
		Topic string `json:"topic"`
		Description string `json:"description"`
		Narrator string `json:"narrator"`
		VideoStyle string `json:"videoStyle"`
		PostingMethod []string `json:"postingMethod"`
		IsOneTime bool `json:"isOneTime"`
	}

	var req CreateScheduleRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] Error parsing request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"message": "Invalid request",
		})
	}

	// vverify if narrator is valid
	narrators := []string{"alloy", "echo", "fable", "nova", "onyx", "shimmer"}
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

	user, err := util.GetUserById(c.Locals("id").(string))
	if err != nil {
		log.Printf("[ERROR] Error getting user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"message": "Error getting user",
		})
	}

	// save video
	videoData := &models.Video{
		Topic: req.Topic,
		Description: req.Description,
		Narrator: req.Narrator,
		VideoStyle: req.VideoStyle,
		PostingMethod: req.PostingMethod,
		IsOneTime: req.IsOneTime,
		OwnerID: user.ID,
		Owner: *user,
	}

	video, err := util.SetVideo(videoData)
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




