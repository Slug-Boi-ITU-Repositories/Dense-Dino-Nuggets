package repository

import (
	"errors"
	"log"
	"minitwit/src/model"

	"gorm.io/gorm"
)

type LatestRepository struct {
	db *gorm.DB
}

func NewLatestRepository(database *gorm.DB) *LatestRepository {
	return &LatestRepository{db: database}
}

func (r *LatestRepository) GetLatest() (int32, error) {
	var latest model.Latest
	result := r.db.First(&latest)
	if result.Error != nil {
		return -1, result.Error
	}
	return latest.Latest, nil
}

func (r *LatestRepository) UpdateLatest(latest int32) error {
	if latest < 0 {
		log.Println("Latest lower than 0")
		return nil
	}

	// Get current value
	current, err := r.GetLatest()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if current == -1 {
		result := r.db.Create(&model.Latest{Latest: latest})
		if result.Error != nil {
			return result.Error
		}
		return nil
	}

	// Try to update value
	// Unsure if this where check is needed
	result := r.db.Model(&model.Latest{}).
		Where("latest = ?", current).
		Update("latest", latest)

	if result.Error != nil {
		return result.Error
	}
    
	// Updated row = success
	if result.RowsAffected > 0 {
		return nil
	}
	return nil
}
