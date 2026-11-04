package models

import (
	"database/sql"
	"time"
)

type Tag struct {
	ID        uint         `json:"id" gorm:"primarykey"`
	CreatedAt time.Time    `json:"-"`
	UpdatedAt time.Time    `json:"-"`
	DeletedAt sql.NullTime `json:"-" gorm:"index"`
	Name      string       `json:"name" gorm:"unique" validate:"required,min=3,max=20"`
}

func (t Tag) Validate() error {
	return validate.Struct(t)
}
