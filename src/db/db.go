package db

import (
	"minitwit/src/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(dsn string) (*gorm.DB, error) {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn))
	if err != nil {
		return nil, err
	}
	// Auto-migrate models
	err = DB.AutoMigrate(&model.User{}, &model.Message{}, &model.Follower{})
	if err != nil {
		return nil, err
	}

	return DB, nil
}
