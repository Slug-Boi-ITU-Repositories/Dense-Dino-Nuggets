package repository

import (
    "minitwit/src/model"
    "gorm.io/gorm"
)

type LatestRepository struct{
    db *gorm.DB
}

func NewLatestRepository(database *gorm.DB) *LatestRepository {
    return &LatestRepository{db: database}
}

func (r *LatestRepository) GetLatest() (uint, error) {
    var latest model.Latest
    result := r.db.First(&latest)
    if result.Error != nil {
        return 0, result.Error
    }
    return latest.Latest, nil
}

func (r *LatestRepository) IncrementLatest() (uint, error) {
    var latest model.Latest
    
    result := r.db.First(&latest)
    if result.Error != nil {
        return 0, result.Error
    }
    latest.Latest++
    
    result = r.db.Save(&latest)
    if result.Error != nil {
        return 0, result.Error
    }
    
    return latest.Latest, nil
}