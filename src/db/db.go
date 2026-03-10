package db

import (
	"minitwit/src/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(dbPath string) (*gorm.DB, error) {
    var err error
    DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
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