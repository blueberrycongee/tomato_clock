package model

import "time"

// TimerSession 以 JSON 持久化的计时记录
type TimerSession struct {
	ID            int64     `json:"id"`
	TaskID        *int64    `json:"task_id,omitempty"`
	Mode          string    `json:"mode"`
	TargetSeconds int       `json:"target_seconds"`
	StartedAt     time.Time `json:"started_at"`
	EndedAt       time.Time `json:"ended_at"` // 零值表示未结束
	Interrupted   bool      `json:"interrupted"`
	DurationSec   int       `json:"duration_sec"` // 方便统计直接累加
}

// StartSession 新建计时记录并返回 ID
func StartSession(taskID *int64, mode string, targetSeconds int) (int64, error) {
	return StartTimerSession(taskID, mode, targetSeconds) // 调用 store.go 提供的实现
}

// EndSession 完成计时（正常或中断）
func EndSession(id int64, interrupted bool) error {
	return EndTimerSession(id, interrupted)
}

// UpdateSessionTask 更新计时记录关联的任务（taskID 为 nil 表示自由计时）
func UpdateSessionTask(id int64, taskID *int64) error {
	return updateSessionTask(id, taskID)
}

// UpdateSession 更新专注记录的所有字段
func UpdateSession(session TimerSession) error {
	return UpdateTimerSession(session)
}

// DeleteSession 删除指定ID的计时记录
func DeleteSession(id int64) error {
	return DeleteTimerSession(id) // 调用 store.go 提供的实现
}
