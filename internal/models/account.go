package models

import "time"

type Account struct {
	ID        int64     `gorm:"primaryKey;column:id" json:"id"`
	Username  string    `gorm:"column:username;uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"column:email;uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"column:password;not null" json:"-"` // bcrypt hash, hidden from JSON
	Role      string    `gorm:"column:role;not null" json:"role"`  // "admin" | "driver"
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
