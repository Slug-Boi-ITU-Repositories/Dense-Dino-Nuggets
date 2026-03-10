package repository

import (
	"minitwit/src/model"
	"gorm.io/gorm"
)

type UserRepository struct{
	db *gorm.DB
}

func NewUserRepository(database *gorm.DB) *UserRepository {
	return &UserRepository{db: database}
}

func (r *UserRepository) GetUserByUsername(username string) (*model.User, error) {
    var user model.User
    err := r.db.Where("username = ?", username).First(&user).Error
    return &user, err
}

// Get user messages
func (r *UserRepository) GetUserMessages(userID uint) ([]model.Message, error) {
	var messages []model.Message
	err := r.db.Where("author_id = ?", userID).Find(&messages).Error
	return messages, err
}

// Get user followers
func (r *UserRepository) GetUserFollowers(userID uint) ([]model.User, error) {
	var followers []model.User
	err := r.db.Where("who_id = ?", userID).Find(&followers).Error
	return followers, err
}	


func (r *UserRepository) IsFollowing(whoID, whomID uint) (bool, error) {
	var count int64
    err := r.db.Model(&model.Follower{}).
	Where("who_id = ? AND whom_id = ?", whoID, whomID).
	Count(&count).Error
    return count > 0, err
}

// get user id by username
func (r *UserRepository) GetUserIDByUsername(username string) (uint, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return user.UserID, err
}

// Register a new user in the database
func (r *UserRepository) RegisterUser(username, email, pwHash string) error {
	user := model.User{
		Username: username,
		Email:    email,
		PwHash:   pwHash,
	}
	return r.db.Create(&user).Error
}


func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}