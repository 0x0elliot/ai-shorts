package database

import (
	"fmt"
	"go-authentication-boilerplate/models"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB represents a Database instance
var DB *gorm.DB

var PRIVKEY string

// ConnectToDB connects the server with database
func ConnectToDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	user := os.Getenv("PSQL_USER")
	if len(user) == 0 {
		user = "postgres"
	}

	password := os.Getenv("PSQL_PASS")
	if len(password) == 0 {
		password = "postgres"
	}

	dbname := os.Getenv("PSQL_DBNAME")
	if len(dbname) == 0 {
		dbname = "postgres"
	}

	port := os.Getenv("PSQL_PORT")
	if len(port) == 0 {
		port = "5432"
	}

	PRIVKEY := os.Getenv("JWT_PRIVKEY")
	if len(PRIVKEY) == 0 {
		log.Fatal("JWT_PRIVKEY is not set")
	}

	dsn := fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Kolkata",
		user, password, dbname, port)

	log.Print("Connecting to Postgres DB...")
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database. \n", err)
		os.Exit(2)

	}
	log.Println("connected")

	// turned on the loger on info mode
	DB.Logger = logger.Default.LogMode(logger.Info)

	log.Print("Running the migrations...")
	DB.AutoMigrate(
		&models.User{}, 
		&models.Claims{},
		&models.Video{},

		// billing
		&models.Subscription{},
		&models.CheckoutSession{},
		&models.Invoice{},
	)
}
