package util

import (
	db "go-authentication-boilerplate/database"
	models "go-authentication-boilerplate/models"
	"log"
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

func SetVideo(video *models.Video) (*models.Video, error) {
	// check if video with ID exists
	if video.ID == "" {
		video.CreatedAt = db.DB.NowFunc().String()
		video.UpdatedAt = db.DB.NowFunc().String()
		txn := db.DB.Create(video)
		if txn.Error != nil {
			log.Printf("[ERROR] Error creating video: %v", txn.Error)
			return video, txn.Error
		}
	} else {
		video.UpdatedAt = db.DB.NowFunc().String()
		txn := db.DB.Save(video)
		if txn.Error != nil {
			log.Printf("[ERROR] Error saving video: %v", txn.Error)
			return video, txn.Error
		}
	}

	return video, nil
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