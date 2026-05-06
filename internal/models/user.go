package models

import basemodel "main/internal/models/base_model"

type User struct {
	basemodel.BaseModel
	Name           string `gorm:"column:name;not null" json:"name"`
	HashedPassword string `gorm:"column:hashed_password;not null" json:"-"`
}
