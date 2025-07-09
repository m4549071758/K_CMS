package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

// json <-> stringの相互変換
type StringArray []string

func (sa StringArray) Value() (driver.Value, error) {
	if len(sa) == 0 {
		return "[]", nil
	}
	return json.Marshal(sa)
}
func (sa *StringArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal StringArray value")
	}

	return json.Unmarshal(bytes, &sa)
}

type Article struct {
	gorm.Model
	ID            uuid.UUID   `gorm:"type:char(36);primaryKey" json:"id"`
	Title         string      `gorm:"varchar(255);not null" json:"title"`
	Excerpt       string      `gorm:"varchar(255);not null" json:"excerpt"`
	CoverImageURL string      `gorm:"varchar(255);not null" json:"cover_image"`
	OgImageURL    string      `gorm:"varchar(255);not null" json:"og_image"`
	Tags          StringArray `gorm:"type:text;not null" json:"tags"`
	Datetime      string      `gorm:"type:date;not null" json:"datetime"`
	Content       string      `gorm:"type:longtext;not null" json:"content"`
	LikeCount     int         `gorm:"default:0;not null" json:"like_count"`

	UserID uuid.UUID `gorm:"type:char(36);not null" json:"user_id"`
	User   User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Likes  []Like    `gorm:"foreignKey:ArticleID" json:"likes,omitempty"`
}

func (article *Article) BeforeCreate(tx *gorm.DB) (err error) {
	if article.ID == uuid.Nil {
		article.ID = NewUUIDv7()
	}
	return nil
}
