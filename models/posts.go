package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type Article struct {
	ID            uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	Title         string    `gorm:"varchar(255);not null" json:"title"`
	Excerpt       string    `gorm:"varchar(255);not null" json:"excerpt"`
	CoverImageURL string    `gorm:"varchar(255);not null" json:"cover_image"`
	OgImageURL    string    `gorm:"varchar(255);not null" json:"og_image"`
	Tags          []string  `gorm:"type:json;not null" json:"tags"`
	Datetime      time.Time `gorm:"type:datetime;not null" json:"datetime"`
	Content       string    `gorm:"type:longtext;not null" json:"content"`
}

func (article *Article) BeforeCreate() (err error) {
	if article.ID == uuid.Nil {
		article.ID = NewUUIDv7()
	}
	return nil
}
