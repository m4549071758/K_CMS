package models

import (
	"gorm.io/gorm"
)

type SiteConfig struct {
	gorm.Model
	// 基本設定
	SiteTitle       string `gorm:"size:255;not null" json:"site_title"`
	SiteDescription string `gorm:"type:text" json:"site_description"`

	// 高度なSEO設定
	GoogleAnalyticsId string `gorm:"size:50" json:"google_analytics_id"`
	OgpImageUrl       string `gorm:"size:255" json:"ogp_image_url"`
	TwitterCardType   string `gorm:"size:50;default:'summary_large_image'" json:"twitter_card_type"`
	TwitterSite       string `gorm:"size:50" json:"twitter_site"`
	RobotIndex        bool   `gorm:"default:true" json:"robot_index"` // true: index, false: noindex

	// リッチリザルト（構造化データ）設定
	PublisherType        string `gorm:"size:50;default:'Person'" json:"publisher_type"` // Person or Organization
	PublisherLogoUrl     string `gorm:"size:255" json:"publisher_logo_url"`
	PublisherDescription string `gorm:"type:text" json:"publisher_description"`
	SocialLinks          string `gorm:"type:text" json:"social_links"` // JSON string of []string
}
