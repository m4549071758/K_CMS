package utils

import (
	"bufio"
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
	// ステータス更新: 開始
	SetBuildStart(action, articleID)

	startTime := time.Now()
	log.Printf("[Build] ビルドプロセスを開始: action=%s, articleID=%s", action, articleID)
	AppendBuildLog(fmt.Sprintf("Build started: action=%s, articleID=%s", action, articleID))

	// タイムアウト付きコンテキストを作成
	ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
	defer cancel()

	// ビルドスクリプトを実行
	cmd := exec.CommandContext(ctx, "/bin/bash", scriptPath, action, articleID)

	// パイプの取得
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		logErrorAndFinish(action, articleID, startTime, "Failed to get stdout pipe: "+err.Error(), "")
		return
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		logErrorAndFinish(action, articleID, startTime, "Failed to get stderr pipe: "+err.Error(), "")
		return
	}

	// コマンド開始
	if err := cmd.Start(); err != nil {
		logErrorAndFinish(action, articleID, startTime, "Failed to start command: "+err.Error(), "")
		return
	}

	// バッファ付きチャネルでブロッキングを防止
	logChan := make(chan string, 512)
	doneChan := make(chan struct{}, 2)

	// Stdout読み取りゴルーチン
	go func() {
		defer func() { doneChan <- struct{}{} }()
		scanner := bufio.NewScanner(stdoutPipe)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			select {
			case logChan <- scanner.Text():
			case <-ctx.Done():
				return
			}
		}
	}()

	// Stderr読み取りゴルーチン
	go func() {
		defer func() { doneChan <- struct{}{} }()
		scanner := bufio.NewScanner(stderrPipe)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			select {
			case logChan <- scanner.Text():
			case <-ctx.Done():
				return
			}
		}
	}()

	// 両リーダー完了後にチャネルをクローズするゴルーチン
	go func() {
		<-doneChan
		<-doneChan
		close(logChan)
	}()

	// ログをドレインしてから Wait（logChan が close されるまでブロック）
	for text := range logChan {
		AppendBuildLog(text)
	}

	// コマンド完了待ち
	waitErr := cmd.Wait()

	duration := time.Since(startTime)

	// エラーハンドリング
	if waitErr != nil {
		errorMsg := waitErr.Error()
		if ctx.Err() == context.DeadlineExceeded {
			errorMsg = "タイムアウト"
		} else if exitError, ok := waitErr.(*exec.ExitError); ok {
			// 終了コードが含まれる場合
			errorMsg = fmt.Sprintf("Exit code: %d, Error: %s", exitError.ExitCode(), exitError.Error())
		}
		
		logBuildFailure(action, articleID, duration, errorMsg, "See build logs for details")
		sendBuildFailureNotification(action, articleID, errorMsg, "See build logs for details")
		
		AppendBuildLog(fmt.Sprintf("Build failed: %s", errorMsg))
		SetBuildComplete(false)
		return
	}

	// 成功
	log.Printf("[Build] ✅ ビルド成功: action=%s, articleID=%s, 所要時間=%v", action, articleID, duration)
	
	AppendBuildLog(fmt.Sprintf("Build success! Duration: %v", duration))
	SetBuildComplete(true)
}

// エラー終了時のヘルパー関数
func logErrorAndFinish(action, articleID string, startTime time.Time, errorMsg string, output string) {
	duration := time.Since(startTime)
	logBuildFailure(action, articleID, duration, errorMsg, output)
	sendBuildFailureNotification(action, articleID, errorMsg, output)
	AppendBuildLog(fmt.Sprintf("Build failed: %s", errorMsg))
	SetBuildComplete(false)
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
