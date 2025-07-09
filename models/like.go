package models

import (
	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

type Like struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	ArticleID   uuid.UUID `gorm:"type:char(36);not null;index" json:"article_id"`
	Fingerprint string    `gorm:"type:varchar(255);not null;index" json:"fingerprint"`
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
	// テーブルが存在しない場合は作成
	if err := db.AutoMigrate(&Like{}); err != nil {
		return err
	}

	// ユニーク制約を追加（記事ID + フィンガープリントの組み合わせでユニーク）
	// エラーは無視（制約が既に存在する場合があるため）
	db.Exec("ALTER TABLE likes ADD CONSTRAINT unique_article_fingerprint UNIQUE (article_id, fingerprint)")

	return nil
}
