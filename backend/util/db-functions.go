package util

import (
	db "go-authentication-boilerplate/database"
	models "go-authentication-boilerplate/models"
	"log"

	"gorm.io/gorm"
)

func GetUserById(id string) (*models.User, error) {
	user := new(models.User)
	txn := db.DB.Where("id = ?", id).First(&user)
	if txn.Error != nil {
		log.Printf("[ERROR] Error getting user: %v", txn.Error)
		return nil, txn.Error
	}
	return user, nil
}

func GetActiveSubscriptionByUserID(userID string) (*models.Subscription, error) {
	subscription := new(models.Subscription)
	txn := db.DB.Preload("Invoices").Where("user_id = ? AND status = ?", userID, "active").First(&subscription)
	if txn.Error != nil {
		if txn.Error.Error() == "record not found" {
			log.Printf("[INFO] No active subscription found for user: %s", userID)
			return nil, nil
		}

		log.Printf("[ERROR] Error getting active subscription: %v", txn.Error)
		return nil, txn.Error
	}
	return subscription, nil
}

func GetVideosByOwner(ownerID string, newestFirst bool) ([]models.Video, error) {
	videos := []models.Video{}

	var txn *gorm.DB

	if newestFirst {
		txn = db.DB.Where("owner_id = ?", ownerID).Preload("Owner").Order("created_at desc").Find(&videos)
	} else {
		txn = db.DB.Where("owner_id = ?", ownerID).Preload("Owner").Order("created_at asc").Find(&videos)
	}

	if txn.Error != nil {
		if txn.Error.Error() == "record not found" {
			return videos, nil
		}

		log.Printf("[ERROR] Error getting videos: %v", txn.Error)
		return nil, txn.Error
	}
	return videos, nil
}

func GetVideoById(id string) (*models.Video, error) {
	video := new(models.Video)
	txn := db.DB.Where("id = ?", id).Preload("Owner").First(&video)
	if txn.Error != nil {
		log.Printf("[ERROR] Error getting video: %v", txn.Error)
		return nil, txn.Error
	}
	return video, nil
}

func SetVideo(video *models.Video) (*models.Video, error) {
	// check if video with ID exists
	if video.ID == "" {
		video.CreatedAt = db.DB.NowFunc().String()
		video.UpdatedAt = db.DB.NowFunc().String()
		txn := db.DB.Omit("Owner").Create(video)
		if txn.Error != nil {
			log.Printf("[ERROR] Error creating video: %v", txn.Error)
			return video, txn.Error
		}
	} else {
		video.UpdatedAt = db.DB.NowFunc().String()
		txn := db.DB.Omit("Owner").Save(video)
		if txn.Error != nil {
			log.Printf("[ERROR] Error saving video: %v", txn.Error)
			return video, txn.Error
		}
	}

	return video, nil
}

func SetCheckoutSession(checkoutSession *models.CheckoutSession) (*models.CheckoutSession, error) {
	// check if checkout session with ID exists
	if checkoutSession.ID == "" {
		checkoutSession.CreatedAt = db.DB.NowFunc().String()
		checkoutSession.UpdatedAt = db.DB.NowFunc().String()
		txn := db.DB.Omit("User").Create(checkoutSession)
		if txn.Error != nil {
			log.Printf("[ERROR] Error creating checkout session: %v", txn.Error)
			return checkoutSession, txn.Error
		}
	} else {
		checkoutSession.UpdatedAt = db.DB.NowFunc().String()
		txn := db.DB.Omit("User").Save(checkoutSession)
		if txn.Error != nil {
			log.Printf("[ERROR] Error saving checkout session: %v", txn.Error)
			return checkoutSession, txn.Error
		}
	}

	return checkoutSession, nil
}

func SetUser(user *models.User) (*models.User, error) {
	// check if user with ID exists
	if user.ID == "" {
		user.CreatedAt = db.DB.NowFunc().String()
		user.UpdatedAt = db.DB.NowFunc().String()
		txn := db.DB.Omit("CurrentShop").Create(user)
		if txn.Error != nil {
			log.Printf("[ERROR] Error creating user: %v", txn.Error)
			return user, txn.Error
		}
	} else {
		user.UpdatedAt = db.DB.NowFunc().String()
		txn := db.DB.Omit("CurrentShop").Save(user)
		if txn.Error != nil {
			log.Printf("[ERROR] Error saving user: %v", txn.Error)
			return user, txn.Error
		}
	}

	return user, nil
}