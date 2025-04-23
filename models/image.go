package models

import (
	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

type Image struct {
	gorm.Model
	ID       uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	FileName string    `gorm:"varchar(255);not null" json:"file_name"`
}
