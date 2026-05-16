package utils

import (
	"sync"
	"time"
)

// BuildState 定数
type BuildState string

const (
	BuildStateIdle    BuildState = "idle"
	BuildStateRunning BuildState = "running"
	BuildStateSuccess BuildState = "success"
	BuildStateFailed  BuildState = "failed"
)

// BuildStatus 構造体
type BuildStatus struct {
	State     BuildState `json:"state"`
	Logs      []string   `json:"logs"`
	StartTime time.Time  `json:"start_time"`
	EndTime   time.Time  `json:"end_time"`
	Action    string     `json:"action"` // create, update, delete
	ArticleID string     `json:"article_id"`
}

// BuildStore はビルド状態を管理します
var (
	currentStatus BuildStatus
	statusMutex   sync.RWMutex
)

func init() {
	currentStatus = BuildStatus{
		State: BuildStateIdle,
		Logs:  []string{},
	}
}

// SetBuildStart はビルド開始を設定します
func SetBuildStart(action, articleID string) {
	statusMutex.Lock()
	defer statusMutex.Unlock()

	currentStatus = BuildStatus{
		State:     BuildStateRunning,
		Logs:      []string{},
		StartTime: time.Now(),
		Action:    action,
		ArticleID: articleID,
	}
}

const maxBuildLogLines = 1000

// AppendBuildLog はログを追加します
func AppendBuildLog(logLine string) {
	statusMutex.Lock()
	defer statusMutex.Unlock()

	currentStatus.Logs = append(currentStatus.Logs, logLine)
	if len(currentStatus.Logs) > maxBuildLogLines {
		currentStatus.Logs = currentStatus.Logs[len(currentStatus.Logs)-maxBuildLogLines:]
	}
}

// SetBuildComplete はビルド完了（成功/失敗）を設定します
func SetBuildComplete(success bool) {
	statusMutex.Lock()
	defer statusMutex.Unlock()

	currentStatus.EndTime = time.Now()
	if success {
		currentStatus.State = BuildStateSuccess
	} else {
		currentStatus.State = BuildStateFailed
	}
}

// GetBuildStatus は現在の状態を返します
func GetBuildStatus() BuildStatus {
	statusMutex.RLock()
	defer statusMutex.RUnlock()

	logsCopy := make([]string, len(currentStatus.Logs))
	copy(logsCopy, currentStatus.Logs)

	return BuildStatus{
		State:     currentStatus.State,
		Logs:      logsCopy,
		StartTime: currentStatus.StartTime,
		EndTime:   currentStatus.EndTime,
		Action:    currentStatus.Action,
		ArticleID: currentStatus.ArticleID,
	}
}
