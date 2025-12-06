package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

const (
	// ビルドのタイムアウト時間（5分）
	buildTimeout = 5 * time.Minute
)

// TriggerBuild はフロントエンドビルドを非同期で実行します
// action: "create", "update", "delete"のいずれか
// articleID: 対象の記事ID
func TriggerBuild(action, articleID string) {
	scriptPath := os.Getenv("BUILD_SCRIPT_PATH")
	if scriptPath == "" {
		log.Println("[Build] BUILD_SCRIPT_PATH環境変数が設定されていません。ビルドをスキップします。")
		return
	}

	log.Printf("[Build] ビルドリクエストを受信: action=%s, articleID=%s", action, articleID)

	// 非同期でビルドを実行（APIレスポンスをブロックしない）
	go executeBuild(scriptPath, action, articleID)
}

// executeBuild はビルドスクリプトを実際に実行します
func executeBuild(scriptPath, action, articleID string) {
	startTime := time.Now()
	log.Printf("[Build] ビルドプロセスを開始: action=%s, articleID=%s", action, articleID)

	// タイムアウト付きコンテキストを作成
	ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
	defer cancel()

	// ビルドスクリプトを実行
	cmd := exec.CommandContext(ctx, "/bin/bash", scriptPath, action, articleID)

	// 標準出力と標準エラー出力を取得
	output, err := cmd.CombinedOutput()

	duration := time.Since(startTime)

	// エラーハンドリング
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			// タイムアウト
			logBuildFailure(action, articleID, duration, "タイムアウト", string(output))
			sendBuildFailureNotification(action, articleID, "タイムアウト", string(output))
		} else {
			// その他のエラー
			logBuildFailure(action, articleID, duration, err.Error(), string(output))
			sendBuildFailureNotification(action, articleID, err.Error(), string(output))
		}
		return
	}

	// 成功
	log.Printf("[Build] ✅ ビルド成功: action=%s, articleID=%s, 所要時間=%v", action, articleID, duration)
	log.Printf("[Build] 出力:\n%s", string(output))
}

// logBuildFailure はビルド失敗時の詳細ログを出力します
func logBuildFailure(action, articleID string, duration time.Duration, errorMsg string, output string) {
	log.Printf("[Build] ❌ ビルド失敗: action=%s, articleID=%s, 所要時間=%v", action, articleID, duration)
	log.Printf("[Build] エラー: %s", errorMsg)
	log.Printf("[Build] 出力:\n%s", output)
}

// sendBuildFailureNotification はビルド失敗時の通知を送信します
// 現在はログ出力のみですが、将来的にメールやSlack通知を追加できます
func sendBuildFailureNotification(action, articleID, errorMsg string, output string) {
	// 通知メッセージを構築
	notificationMsg := fmt.Sprintf(
		"🚨 フロントエンドビルド失敗通知\n"+
			"アクション: %s\n"+
			"記事ID: %s\n"+
			"エラー: %s\n"+
			"出力:\n%s",
		action, articleID, errorMsg, output,
	)

	// 現在はログに出力（将来的に拡張可能）
	log.Printf("[Build Notification] %s", notificationMsg)

	// TODO: 外部通知サービスとの連携
	// 例：
	// - メール送信
	// - Slack Webhook
	// - Discord Webhook
	// - データベースに失敗履歴を保存

	// Slack Webhook の実装例（コメントアウト）
	/*
		slackWebhookURL := os.Getenv("SLACK_WEBHOOK_URL")
		if slackWebhookURL != "" {
			payload := map[string]interface{}{
				"text": notificationMsg,
			}
			jsonData, _ := json.Marshal(payload)
			http.Post(slackWebhookURL, "application/json", bytes.NewBuffer(jsonData))
		}
	*/
}
