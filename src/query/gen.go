package query

import (
	"minitwit/src/model"

	"gorm.io/gorm"
)

type Querier interface {
    // Get user by username
    GetUserByUsername(username string) (*model.User, error)

    // Get user's followers
    GetUserFollowers(userID uint) ([]model.User, error)
    
    // Get user's messages with author info
    GetUserMessages(userID uint) ([]model.Message, error)
    
    // Get timeline messages (user's messages + followers' messages)
    GetTimelineMessages(userID uint) ([]model.Message, error)
    
    // Check if user is following another user
    IsFollowing(whoID, whomID uint) (bool, error)
    
    // Get flagged messages
    GetFlaggedMessages() ([]model.Message, error)
}

// You can also add a default implementation if needed
type QuerierImpl struct {
    *gorm.DB
}

func (q *QuerierImpl) GetUserByUsername(username string) (*model.User, error) {
    var user model.User
    err := q.Where("username = ?", username).First(&user).Error
    return &user, err
}
