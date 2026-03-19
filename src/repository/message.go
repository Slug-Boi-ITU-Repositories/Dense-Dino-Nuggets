package repository

import (
	"minitwit/src/model"
	"time"

	"gorm.io/gorm"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(database *gorm.DB) *MessageRepository {
	return &MessageRepository{db: database}
}

func (r *MessageRepository) GetPublicTimeline(limit int, offset int) ([]model.Message, error) {
	var messages []model.Message
	err := r.db.Preload("Author").
		Where("flagged = 0").
		Order("pub_date DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}

func (r *MessageRepository) CountPublicTimeline() (int64, error) {
	var count int64
	err := r.db.Model(&model.Message{}).
		Where("flagged = 0").
		Count(&count).Error
	return count, err
}

func (r *MessageRepository) GetUserTimeline(userID uint, limit int, offset int) ([]model.Message, error) {
	var messages []model.Message
	err := r.db.Preload("Author").
		Where("author_id = ? AND flagged = 0", userID).
		Order("pub_date DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}

func (r *MessageRepository) CountUserTimeline(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Message{}).
		Where("author_id = ? AND flagged = 0", userID).
		Count(&count).Error
	return count, err
}

// GetPersonalTimeline returns messages from the user and users they follow
func (r *MessageRepository) GetPersonalTimeline(userID uint, limit int, offset int) ([]model.Message, error) {
	var messages []model.Message
	err := r.db.Preload("Author").
		Where("(author_id = ? OR author_id IN (SELECT whom_id FROM follower WHERE who_id = ?)) AND flagged = 0", userID, userID).
		Order("pub_date DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}

func (r *MessageRepository) CountPersonalTimeline(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Message{}).
		Where("(author_id = ? OR author_id IN (SELECT whom_id FROM follower WHERE who_id = ?)) AND flagged = 0", userID, userID).
		Count(&count).Error
	return count, err
}

// Function for adding a new message to the database
func (r *MessageRepository) AddMessage(authorID uint, text string) error {
	message := model.Message{
		AuthorID: authorID,
		Text:     text,
		PubDate:  time.Now().Unix(),
		Flagged:  0,
	}
	return r.db.Create(&message).Error
}

func (r *MessageRepository) Create(message *model.Message) error {
	return r.db.Create(message).Error
}
