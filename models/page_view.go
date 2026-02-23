package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

// PageView は1記事・1ユーザー・1日を1レコードとして記録する。
// (article_id, fingerprint, visited_date) の複合ユニーク制約により
// 同一日の再訪問はDBレベルで弾かれる。
type PageView struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	ArticleID   uuid.UUID `gorm:"type:char(36);not null;index" json:"article_id"`
	Fingerprint string    `gorm:"type:varchar(255);not null;index" json:"fingerprint"`
	IPAddress   string    `gorm:"type:varchar(45);not null" json:"ip_address"`
	VisitedDate time.Time `gorm:"type:date;not null" json:"visited_date"`
	Article     Article   `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
}

func (PageView) TableName() string {
	return "page_views"
}

func (pv *PageView) BeforeCreate(tx *gorm.DB) (err error) {
	if pv.ID == uuid.Nil {
		pv.ID = NewUUIDv7()
	}
	return nil
}

// MigratePageView はテーブル作成と複合ユニーク制約を設定する。
func MigratePageView(db *gorm.DB) error {
	if err := db.AutoMigrate(&PageView{}); err != nil {
		return err
	}

	// 同一記事・同一fingerprint・同一日は1件のみ許可
	// エラーは無視（制約が既に存在する場合があるため）
	db.Exec("ALTER TABLE page_views ADD CONSTRAINT uq_pageview_article_fp_date UNIQUE (article_id, fingerprint, visited_date)")

	return nil
}
