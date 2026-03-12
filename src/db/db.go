package db

import (
	"minitwit/src/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(dbPath string) (*gorm.DB, error) {
    var err error
    DB, err = gorm.Open(sqlite.Open(dbPath))
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