package db

import (
	"log"
	"minitwit/src/model"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(dsn string) (*gorm.DB, error) {
	log.Println("Connecting to database...")
	retries := 3
	for i := range retries {
		var err error
		DB, err = gorm.Open(postgres.Open(dsn))
		if err == nil {
			log.Println("Database connnection complete")
			break
		}
		if i < 2 {
			log.Printf("Failed to connect to database [%d/%d]! Retrying in 5 seconds...\n", i+1, retries)
		} else {
			return nil, err
		}
		time.Sleep(time.Second * 5)
	}

	// Auto-migrate models
	log.Println("Running migrations...")
	err := DB.AutoMigrate(&model.User{}, &model.Message{}, &model.Follower{})
	if err != nil {
		log.Printf("Failed to run migrations: %s", err.Error())
		return nil, err
	}
	log.Println("Migrations complete")
	return DB, nil
}
