package models

import (
	"database/sql"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
)

type Tag struct {
	ID        uint         `json:"id" gorm:"primarykey"`
	CreatedAt time.Time    `json:"-"`
	UpdatedAt time.Time    `json:"-"`
	DeletedAt sql.NullTime `json:"-" gorm:"index"`
	Name      string       `json:"name" gorm:"unique"`
}

func (t Tag) Validate() error {
	return validation.ValidateStruct(&t,
		// Name cannot be empty, and the length must be between 3 and 20.
		validation.Field(&t.Name, validation.Required, validation.Length(3, 20)),
	)
}
