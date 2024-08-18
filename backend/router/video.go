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

	privVideo.Get("/list", ListVideos)
	privVideo.Get("/:id", GetVideo)
	privVideo.Post("/create", CreateSchedule)
	privVideo.Post("/recreate/:id", RecreateVideo)
}

func ListVideos(c *fiber.Ctx) error {
	userId := c.Locals("id").(string)

	newestFirstQuery := c.Query("newestFirst")
	newestFirst := true

	if newestFirstQuery == "false" {
		newestFirst = false
	}

	videos, err := util.GetVideosByOwner(userId, newestFirst)
	if err != nil {
		log.Printf("[ERROR] Error getting videos: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"message": "Error getting videos",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
		"videos": videos,
	})
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

	if video.Error != "" {
		video.Error = "An error happened in a step. Try creating the video again"
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
		"video": video,
	})
}


func RecreateVideo(c *fiber.Ctx) error {
	// if video exists but had an error, we start the background job again
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

	// paying customers have full authority
	// if !(len(video.Error) > 0) {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"error": true,
	// 		"message": "Video is not in error state",
	// 	})
	// }

	go util.CreateVideo(video, true)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
		"message": "Recreating video",
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
		VideoTheme string `json:"videoTheme"`
		BackgroundMusic string `json:"backgroundMusic"`
	}

	var req CreateScheduleRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] Error parsing request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"message": "Invalid request",
		})
	}

	// const [musicTracks] = useState([
    //     { name: "Another love", value: "_another-love" },
    //     { name: "Bladerunner 2049", value: "_bladerunner-2049" },
    //     { name: "Constellations", value: "_constellations" },
    //     { name: "Fallen", value: "_fallen" },
    //     { name: "Hotline", value: "_hotline" },
    //     { name: "Izzamuzzic", value: "_izzamuzzic" },
    //     { name: "Nas", value: "_nas" },
    //     { name: "Paris else", value: "_paris-else" },
    //     { name: "Snowfall", value: "_snowfall" },
    // ])

	validTracks := []string{"_another-love", "_bladerunner-2049", "_constellations", "_fallen", "_hotline", "_izzamuzzic", "_nas", "_paris-else", "_snowfall"}

	if !util.Contains(validTracks, req.BackgroundMusic) {
		log.Printf("[ERROR] Invalid background music: %v", req.BackgroundMusic)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"message": "Invalid background music",
		})
	}

	// verify if narrator is valid
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

	if req.BackgroundMusic == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"message": "Background music is required",
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
		VideoTheme: req.VideoTheme,
		BackgroundMusic: req.BackgroundMusic,
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
	go util.CreateVideo(video, false)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
		"video": video,
	})


}




