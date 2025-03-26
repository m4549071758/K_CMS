package models

import (
	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID       uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	Username string    `gorm:"size:255;not null;unique" json:"username"`
	Email    string    `gorm:"size:255;not null;unique" json:"email"`
	Password string    `gorm:"size:255;not null" json:"-"`
}

func NewUUIDv7() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}

func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	if user.ID == uuid.Nil {
		user.ID = NewUUIDv7()
	}
	return nil
}
