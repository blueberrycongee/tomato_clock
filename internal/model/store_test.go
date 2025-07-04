package model

import (
	"testing"
	"time"
)

func TestLast24HoursFocusTimeAndByLabel(t *testing.T) {
	// 重置数据
	mu.Lock()
	data.Tasks = nil
	data.Sessions = nil
	data.NextTaskID = 1
	data.NextSessionID = 1
	mu.Unlock()

	// 固定当前时间
	now := time.Date(2025, 7, 3, 12, 0, 0, 0, time.UTC)
	nowFunc = func() time.Time { return now }

	// 创建任务
	taskID := int64(1)
	task := Task{ID: taskID, Title: "任务1", Label: "休息"}
	mu.Lock()
	data.Tasks = []Task{task}
	mu.Unlock()

	// Session #1 完全在窗口内，1 小时
	s1 := TimerSession{
		ID:          1,
		TaskID:      &taskID,
		Mode:        "countup",
		StartedAt:   now.Add(-2 * time.Hour),
		EndedAt:     now.Add(-1 * time.Hour),
		DurationSec: 3600,
	}
	// Session #2 跨越窗口，重叠 1 小时
	s2 := TimerSession{
		ID:          2,
		TaskID:      &taskID,
		Mode:        "countup",
		StartedAt:   now.Add(-26 * time.Hour),
		EndedAt:     now.Add(-23 * time.Hour),
		DurationSec: 3 * 3600,
	}
	// Session #3 未分类，30 分钟
	s3 := TimerSession{
		ID:          3,
		Mode:        "countup",
		StartedAt:   now.Add(-30 * time.Minute),
		EndedAt:     now,
		DurationSec: 1800,
	}

	mu.Lock()
	data.Sessions = []TimerSession{s1, s2, s3}
	mu.Unlock()

	total := Last24HoursFocusTime()
	expected := 3600 + 3600 + 1800 // 1h + 1h(overlap) + 0.5h
	if total != expected {
		t.Fatalf("expected total %d, got %d", expected, total)
	}

	byLabel := Last24HoursFocusTimeByLabel()
	if byLabel["休息"] != 7200 {
		t.Fatalf("label 休息 expected 7200, got %d", byLabel["休息"])
	}
	if byLabel[DefaultLabel] != 1800 {
		t.Fatalf("label 未分类 expected 1800, got %d", byLabel[DefaultLabel])
	}
}
