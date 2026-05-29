package models

import "time"

type RefreshToken struct {
	ID        int64     `gorm:"primaryKey;column:id" json:"id"`
	UserID    int64     `gorm:"column:user_id;not null" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"-"`
	Token     string    `gorm:"column:token;type:text;not null;unique" json:"token"`
	ExpiredAt time.Time `gorm:"column:expired_at;not null" json:"expired_at"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}
