package db

import (
	"minitwit/src/model"

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
	var err error
	DB, err = gorm.Open(postgres.Open(dsn))
	if err != nil {
		return nil, err
	}
	// Auto-migrate models
	if shouldMigrate(DB) {
    err = DB.AutoMigrate(&model.User{}, &model.Message{}, &model.Follower{})
    if err != nil {
        return nil, err
    }
}

	return DB, nil
}
