package repository

import (
	"minitwit/src/model"
	"time"

	"gorm.io/gorm"
)

type MessageRepository struct{
    db *gorm.DB
}

func NewMessageRepository(database *gorm.DB) *MessageRepository {
    return &MessageRepository{db: database}
}

func (r *MessageRepository) GetPublicTimeline(limit int) ([]model.Message, error) {
    var messages []model.Message
    err := r.db.Preload("Author").
        Where("flagged = 0").
        Order("pub_date DESC").
        Limit(limit).
        Find(&messages).Error
    return messages, err
}

func (r *MessageRepository) GetUserTimeline(userID uint, limit int) ([]model.Message, error) {
    var messages []model.Message
    err := r.db.Preload("Author").
        Where("author_id = ? AND flagged = 0", userID).
        Order("pub_date DESC").
        Limit(limit).
        Find(&messages).Error
    return messages, err
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