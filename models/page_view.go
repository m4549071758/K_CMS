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
	ArticleID   uuid.UUID `gorm:"type:char(36);not null;uniqueIndex:idx_pv_article_fp_date" json:"article_id"`
	Fingerprint string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_pv_article_fp_date" json:"fingerprint"`
	IPAddress   string    `gorm:"type:varchar(45);not null" json:"ip_address"`
	VisitedDate time.Time `gorm:"type:date;not null;uniqueIndex:idx_pv_article_fp_date" json:"visited_date"`
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

// MigratePageView はテーブル作成を行う。
func MigratePageView(db *gorm.DB) error {
	return db.AutoMigrate(&PageView{})
}
