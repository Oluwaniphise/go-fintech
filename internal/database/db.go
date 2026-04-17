package database

import (
	"fmt"
	"log"
	"os"

	"fintech/internal/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {
	// load .env file

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// build the connection string

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	log.Println("Running Migrations...")
	migrator := db.Migrator()
	if migrator.HasTable(&models.Transaction{}) &&
		migrator.HasColumn(&models.Transaction{}, "provider_ref") &&
		!migrator.HasColumn(&models.Transaction{}, "outbound_provider_ref") {
		if err := migrator.RenameColumn(&models.Transaction{}, "provider_ref", "outbound_provider_ref"); err != nil {
			log.Fatal("Failed to rename provider_ref column: ", err)
		}
	}

	db.AutoMigrate(&models.User{}, &models.Wallet{}, &models.Transaction{}, &models.EmailVerificationToken{}, &models.LoginOTPToken{})

	return db
}
