package model

import "time"

type UserCollection struct {
	Users []User `json:"users,omitempty"`
}

type User struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Username    string    `gorm:"type:varchar(50);unique" json:"username"`
	PhoneNumber string    `gorm:"type:varchar(20);unique;not null" json:"phoneNumber"`
	Email       string    `gorm:"type:varchar(100)" json:"email"`
	FirstName   string    `gorm:"type:varchar(50)" json:"firstName"`
	LastName    string    `gorm:"type:varchar(50)" json:"lastName"`
	Bio         string    `gorm:"type:text" json:"bio"`
	AvatarURL   string    `gorm:"type:varchar(255)" json:"avatarUrl"`
	IsOnline    bool      `gorm:"default:false" json:"isOnline"`
	LastSeen    time.Time `json:"lastSeen"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}
