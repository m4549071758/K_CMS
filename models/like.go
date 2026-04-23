package models

import (
	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

type Like struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	ArticleID   uuid.UUID `gorm:"type:char(36);not null;uniqueIndex:unique_article_fingerprint" json:"article_id"`
	Fingerprint string    `gorm:"type:varchar(255);not null;uniqueIndex:unique_article_fingerprint" json:"fingerprint"`
	IPAddress   string    `gorm:"type:varchar(45);not null" json:"ip_address"`
	Article     Article   `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
}

// 複合ユニークキー：同じ記事に同じfingerprintからは1回のみいいねできる
func (Like) TableName() string {
	return "likes"
}

func (like *Like) BeforeCreate(tx *gorm.DB) (err error) {
	if like.ID == uuid.Nil {
		like.ID = NewUUIDv7()
	}
	return nil
}

// テーブルのインデックスとユニーク制約を定義
func (Like) TableOptions() string {
	return "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4"
}

// AutoMigrateで実行されるSQL
func MigrateLike(db *gorm.DB) error {
	return db.AutoMigrate(&Like{})
}
