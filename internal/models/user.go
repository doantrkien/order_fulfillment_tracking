package models

import "time"

type UserRole string

const (
	ROLE_ADMIN    = "admin"
	ROLE_CUSTOMER = "customer"
	ROLE_SHIPPER  = "shipper"
)

type User struct {
	ID        int64     `gorm:"primaryKey;column:id" json:"id"`
	Username  string    `gorm:"column:username;type:varchar(255);not null;unique" json:"username"`
	Password  string    `gorm:"column:password;type:varchar(255);not null" json:"-"`
	Address   string    `gorm:"column:address;type:varchar(255);not null" json:"address"`
	Phone     string    `gorm:"column:phone;type:varchar(20);not null" json:"phone"`
	Role      UserRole  `gorm:"column:role;type:varchar(50);not null" json:"role"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
