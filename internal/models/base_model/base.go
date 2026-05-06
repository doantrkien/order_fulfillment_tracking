package basemodel

import "time"

type BaseModel struct {
	ID        int64       `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"updated_at" json:"updated_at"`
}
