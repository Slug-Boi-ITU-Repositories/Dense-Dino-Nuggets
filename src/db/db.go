package db

import (
	"log"
	"minitwit/src/model"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func shouldMigrate(db *gorm.DB) bool {
	m := db.Migrator()

	if !m.HasTable(&model.User{}) || !m.HasTable(&model.Message{}) || !m.HasTable(&model.Follower{}) {
		return true
	}

	if !m.HasColumn(&model.User{}, "pw_hash") ||
		!m.HasColumn(&model.Message{}, "pub_date") ||
		!m.HasColumn(&model.Message{}, "flagged") {
		return true
	}

	return false
}

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

	log.Println("Running migrations...")
	if shouldMigrate(DB) {
		err := DB.AutoMigrate(&model.User{}, &model.Message{}, &model.Follower{}, &model.Latest{})
		if err != nil {
			log.Printf("Failed to run migrations: %s", err.Error())
			return nil, err
		}
	}

	if !DB.Migrator().HasIndex(&model.User{}, "idx_user_username") {
		err := DB.Migrator().CreateIndex(&model.User{}, "Username")
		if err != nil {
			log.Printf("Failed to create username unique index: %s", err.Error())
			return nil, err
		}
	}

	log.Println("Migrations complete")

	return DB, nil
}
