package repository

import (
    "minitwit/src/model"
    "gorm.io/gorm"
)

type FollowerRepository struct{
    db *gorm.DB
}

func NewFollowerRepository(database *gorm.DB) *FollowerRepository {
    return &FollowerRepository{db: database}
}

func (r *FollowerRepository) Follow(whoID, whomID uint) error {
    follower := model.Follower{WhoID: whoID, WhomID: whomID}
    return r.db.Create(&follower).Error
}

func (r *FollowerRepository) Unfollow(whoID, whomID uint) error {
    return r.db.Where("who_id = ? AND whom_id = ?", whoID, whomID).
        Delete(&model.Follower{}).Error
}