package models

import (
	"database/sql"
	"strings"
	"time"
)

type LinkInput struct {
	Keyword     string `json:"keyword" validate:"required,min=1,max=100,ne=api,ne=static,ne=directory,ne=healthz,ne=favicon.ico,keywordfmt"`
	Description string `json:"description"`
	Destination string `json:"destination" validate:"required,http_url"`
	Views       int    `json:"views"`
	Tags        []Tag  `json:"tags"`
}

type Link struct {
	ID          uint         `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   sql.NullTime `json:"-" gorm:"index"`
	Keyword     string       `json:"keyword" gorm:"unique;index"`
	Description string       `json:"description"`
	Destination string       `json:"destination"`
	Views       int          `json:"views"`
	Tags        []Tag        `json:"tags" gorm:"many2many:link_tags;"`
}

// TODO Put in validation for programtic links gh/%s
func (l LinkInput) Validate() error {
	return validate.Struct(l)
}

func (li *LinkInput) ToNative() Link {

	return Link{
		Keyword:     strings.ToLower(li.Keyword),
		Destination: li.Destination,
		Description: li.Description,
	}
}

type QueryString struct {
	Sort  string `validate:"omitempty,lettersfmt"`
	Order string `validate:"omitempty,lettersfmt"`
}

func (qs QueryString) Validate() error {
	return validate.Struct(qs)
}

type Links []Link
