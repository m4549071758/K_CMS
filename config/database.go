package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// getEnvWithDefault returns the environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func ConnectDB() *gorm.DB {
	Initialize()

	// デフォルト値を設定（Docker環境での動作を考慮）
	user := getEnvWithDefault("DB_USER", "kcms_user")
	password := getEnvWithDefault("DB_PASSWORD", "kcms_password")
	dbName := getEnvWithDefault("DB_NAME", "kcms_db")
	dbHost := getEnvWithDefault("DB_HOST", "10.0.1.101")
	dbPort := getEnvWithDefault("DB_PORT", "3306")

	log.Printf("DB接続情報: Host=%s, Port=%s, User=%s, DB=%s", dbHost, dbPort, user, dbName)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, dbHost, dbPort, dbName)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("DB接続エラー: %v, DSN: %s", err, dsn)
		panic("Failed to connect to database.")
	}

	log.Println("Database connected successfully.")
	return DB
}

func Initialize() {
	// Docker環境では.envファイルではなく環境変数を優先
	if os.Getenv("DOCKER_ENV") == "true" {
		log.Println("Docker環境: 環境変数を使用")
		return
	}

	// .envファイルの読み込みを試行
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
		// .envが見つからない場合も続行
	} else {
		log.Println(".env file loaded successfully")
	}
}
