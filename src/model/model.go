package model

type User struct {
    UserID   uint   `gorm:"primaryKey;autoIncrement;column:user_id"`
    Username string `gorm:"not null;column:username"`
    Email    string `gorm:"not null;column:email"`
    PwHash   string `gorm:"not null;column:pw_hash"`
    
    Messages  []Message `gorm:"foreignKey:AuthorID"`
    Followers []Follower `gorm:"foreignKey:WhoID"`
    Following []Follower `gorm:"foreignKey:WhomID"`
}

func (User) TableName() string {
    return "user"
}

type Follower struct {
    WhoID  uint `gorm:"column:who_id;index:idx_follower_who"`
    WhomID uint `gorm:"column:whom_id;index:idx_follower_whom"`
    
    Who  User `gorm:"foreignKey:WhoID;references:UserID"`
    Whom User `gorm:"foreignKey:WhomID;references:UserID"`
}

func (Follower) TableName() string {
    return "follower"
}

type Message struct {
    MessageID uint   `gorm:"primaryKey;autoIncrement;column:message_id"`
    AuthorID  uint   `gorm:"not null;column:author_id;index:idx_message_author"`
    Text      string `gorm:"not null;column:text"`
    PubDate   int64  `gorm:"column:pub_date"`
    Flagged   int    `gorm:"column:flagged;default:0"`
    
    Author User `gorm:"foreignKey:AuthorID;references:UserID"`
}

func (Message) TableName() string {
    return "message"
}