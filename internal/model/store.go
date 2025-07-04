package model

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// 保存到用户目录下的文件名
const dataFileName = ".tomato_clock.json"

// DefaultLabel 表示未分类计时记录的标签
const DefaultLabel = "未分类"

// nowFunc 用于获取当前时间，测试时可覆盖
var nowFunc = time.Now

// in-memory 数据结构
var (
	mu   sync.Mutex
	data struct {
		NextTaskID    int64          `json:"next_task_id"`
		NextSessionID int64          `json:"next_session_id"`
		Tasks         []Task         `json:"tasks"`
		Sessions      []TimerSession `json:"sessions"`
	}
	filePath string
)

// Init 在程序启动时调用，负责加载数据文件（若存在）。
func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	filePath = filepath.Join(home, dataFileName)

	b, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 第一次运行：初始化默认值即可
			data.NextTaskID = 1
			data.NextSessionID = 1
			return Save()
		}
		return err
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	return nil
}

// Save 将内存数据写回文件。
func Save() error {
	mu.Lock()
	defer mu.Unlock()

	log.Printf("[DEBUG] 准备保存数据: TaskCount=%d, SessionCount=%d", len(data.Tasks), len(data.Sessions))

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("[DEBUG] JSON序列化失败: %v", err)
		return err
	}
	err = os.WriteFile(filePath, b, 0644)
	if err != nil {
		log.Printf("[DEBUG] 写入文件失败: %v", err)
		return err
	}

	log.Printf("[DEBUG] 成功保存数据到: %s", filePath)
	return nil
}

// internal helpers ------------------------------------------------------

func nextTaskID() int64 {
	id := data.NextTaskID
	data.NextTaskID++
	return id
}

func nextSessionID() int64 {
	id := data.NextSessionID
	data.NextSessionID++
	return id
}

// Task helpers ----------------------------------------------------------

// AddTask 插入新任务并返回完整对象
func AddTask(t *Task) error {
	mu.Lock()
	now := time.Now()
	// 如果没有设置标签，默认赋值为"学习"
	if t.Label == "" {
		t.Label = "学习"
	}
	t.ID = nextTaskID()
	t.CreatedAt = now
	t.UpdatedAt = now

	data.Tasks = append(data.Tasks, *t)
	mu.Unlock()
	if err := Save(); err != nil {
		return err
	}
	log.Printf("[AddTask] id=%d title=%s", t.ID, t.Title)
	return nil
}

// AllTasks 获取全部任务切片（拷贝）
func AllTasks() []Task {
	mu.Lock()
	defer mu.Unlock()
	res := make([]Task, len(data.Tasks))
	copy(res, data.Tasks)
	return res
}

// UpdateTask 更新已有任务信息（根据ID）
func UpdateTask(t Task) error {
	mu.Lock()
	var found bool
	for i, task := range data.Tasks {
		if task.ID == t.ID {
			// 保留创建时间
			t.CreatedAt = task.CreatedAt
			// 更新时间戳
			t.UpdatedAt = time.Now()
			data.Tasks[i] = t
			found = true
			break
		}
	}
	mu.Unlock()
	if !found {
		return fmt.Errorf("task id %d not found", t.ID)
	}
	if err := Save(); err != nil {
		return err
	}
	log.Printf("[UpdateTask] id=%d title=%s label=%s", t.ID, t.Title, t.Label)
	return nil
}

// TimerSession helpers --------------------------------------------------

func StartTimerSession(taskID *int64, mode string, targetSeconds int) (int64, error) {
	mu.Lock()
	s := TimerSession{
		ID:            nextSessionID(),
		TaskID:        taskID,
		Mode:          mode,
		TargetSeconds: targetSeconds,
		StartedAt:     time.Now(),
	}
	data.Sessions = append(data.Sessions, s)
	mu.Unlock()
	if err := Save(); err != nil {
		return 0, err
	}
	log.Printf("[StartTimerSession] id=%d mode=%s target=%d", s.ID, mode, targetSeconds)
	return s.ID, nil
}

func EndTimerSession(id int64, interrupted bool) error {
	mu.Lock()
	var modified bool
	for i, s := range data.Sessions {
		if s.ID == id {
			if s.EndedAt.IsZero() {
				now := time.Now()
				data.Sessions[i].EndedAt = now
				data.Sessions[i].Interrupted = interrupted
				data.Sessions[i].DurationSec = int(now.Sub(s.StartedAt).Seconds())
				modified = true
			}
			break
		}
	}
	mu.Unlock()
	if modified {
		if err := Save(); err != nil {
			return err
		}
		log.Printf("[EndTimerSession] id=%d interrupted=%v duration=%d", id, interrupted, data.Sessions[len(data.Sessions)-1].DurationSec)
	}
	return nil
}

// UpdateSessionTask 修改计时记录的 TaskID (nil 表示自由计时)
func updateSessionTask(id int64, taskID *int64) error {
	mu.Lock()
	var found bool
	for i, s := range data.Sessions {
		if s.ID == id {
			data.Sessions[i].TaskID = taskID
			found = true
			break
		}
	}
	mu.Unlock()
	if !found {
		return nil // 未找到记录
	}
	if err := Save(); err != nil {
		return err
	}
	if taskID == nil {
		log.Printf("[UpdateSessionTask] id=%d task_id=nil", id)
	} else {
		log.Printf("[UpdateSessionTask] id=%d task_id=%d", id, *taskID)
	}
	return nil
}

// UpdateTimerSession 更新计时记录的所有字段
func UpdateTimerSession(session TimerSession) error {
	log.Printf("[DEBUG] 开始更新计时记录: ID=%d, TaskID=%v, Mode=%s, 开始=%s, 结束=%s",
		session.ID, session.TaskID, session.Mode,
		session.StartedAt.Format("2006-01-02 15:04:05"),
		session.EndedAt.Format("2006-01-02 15:04:05"))

	mu.Lock()
	var found bool
	for i, s := range data.Sessions {
		if s.ID == session.ID {
			log.Printf("[DEBUG] 找到要更新的记录: 索引=%d, 原Mode=%s, 原Duration=%d",
				i, data.Sessions[i].Mode, data.Sessions[i].DurationSec)
			// 保留原始ID
			session.ID = s.ID
			data.Sessions[i] = session
			found = true
			break
		}
	}
	mu.Unlock()

	if !found {
		log.Printf("[DEBUG] 未找到ID=%d的计时记录", session.ID)
		return nil // 未找到记录
	}

	if err := Save(); err != nil {
		log.Printf("[DEBUG] 保存数据失败: %v", err)
		return err
	}

	log.Printf("[DEBUG] 成功更新计时记录: id=%d, mode=%s, target=%d, duration=%d",
		session.ID, session.Mode, session.TargetSeconds, session.DurationSec)
	return nil
}

// AggregatedDurations 按任务标题汇总专注秒数（过滤中断、未结束）
func AggregatedDurations() map[string]float64 {
	mu.Lock()
	defer mu.Unlock()

	res := map[string]float64{}
	for _, s := range data.Sessions {
		if s.Interrupted || s.EndedAt.IsZero() {
			continue
		}
		title := "自由计时"
		if s.TaskID != nil {
			for _, t := range data.Tasks {
				if t.ID == *s.TaskID {
					title = t.Title
					break
				}
			}
		}
		res[title] += float64(s.DurationSec)
	}
	defer func() {
		log.Printf("[AggregatedDurations] %d entries", len(res))
	}()
	return res
}

// CompletedSessions 返回已结束且未中断的计时记录副本，以结束时间倒序排序。
func CompletedSessions() []TimerSession {
	mu.Lock()
	defer mu.Unlock()
	log.Printf("[DEBUG] CompletedSessions - 总记录数: %d", len(data.Sessions))

	var list []TimerSession
	for i, s := range data.Sessions {
		if s.EndedAt.IsZero() {
			log.Printf("[DEBUG] 跳过未结束记录 #%d: ID=%d", i, s.ID)
			continue
		}
		if s.Interrupted {
			log.Printf("[DEBUG] 跳过被中断记录 #%d: ID=%d", i, s.ID)
			continue
		}
		log.Printf("[DEBUG] 添加已完成记录 #%d: ID=%d, Mode=%s, Duration=%d",
			i, s.ID, s.Mode, s.DurationSec)
		list = append(list, s)
	}

	log.Printf("[DEBUG] CompletedSessions - 过滤后记录数: %d", len(list))
	sort.Slice(list, func(i, j int) bool { return list[i].EndedAt.After(list[j].EndedAt) })
	return list
}

// PrintSessionsSummary 将完成的计时记录打印到 stdout。
func PrintSessionsSummary() {
	sessions := CompletedSessions()
	log.Println("=== Sessions Summary ===")
	for _, s := range sessions {
		mode := "正计时"
		if s.Mode == "countdown" {
			mode = "倒计时"
		}
		start := s.StartedAt.Format("2006-01-02 15:04:05")
		end := s.EndedAt.Format("2006-01-02 15:04:05")
		log.Printf("[%s] -> [%s] %ds (%s)\n", start, end, s.DurationSec, mode)
	}
	log.Println("========================")
}

// ClearSessions 清空所有计时记录
func ClearSessions() error {
	mu.Lock()
	data.Sessions = nil
	mu.Unlock()
	return Save()
}

// DeleteTask 根据 ID 删除任务及其关联计时记录
func DeleteTask(id int64) error {
	mu.Lock()
	// 删除任务
	var newTasks []Task
	for _, t := range data.Tasks {
		if t.ID != id {
			newTasks = append(newTasks, t)
		}
	}
	data.Tasks = newTasks
	// 删除引用该任务的 session
	var newSess []TimerSession
	for _, s := range data.Sessions {
		if s.TaskID == nil || *s.TaskID != id {
			newSess = append(newSess, s)
		}
	}
	data.Sessions = newSess
	mu.Unlock()
	return Save()
}

// sessionOverlapSeconds 计算计时记录与指定区间 [from,to] 的重叠秒数。
func sessionOverlapSeconds(s TimerSession, from, to time.Time) int {
	// 仅处理已结束的记录
	if s.EndedAt.IsZero() {
		return 0
	}
	// 快速排除无交集
	if s.EndedAt.Before(from) || s.StartedAt.After(to) {
		return 0
	}

	start := s.StartedAt
	end := s.EndedAt
	if start.Before(from) {
		start = from
	}
	if end.After(to) {
		end = to
	}
	dur := int(end.Sub(start).Seconds())
	if dur < 0 {
		return 0
	}
	return dur
}

// Last24HoursFocusTime 精确计算过去 24 小时内与窗口重叠的专注总时长(秒)。
func Last24HoursFocusTime() int {
	mu.Lock()
	defer mu.Unlock()

	now := nowFunc()
	startTime := now.Add(-24 * time.Hour)

	var totalSeconds int
	for _, s := range data.Sessions {
		if s.Interrupted || s.EndedAt.IsZero() {
			continue
		}
		totalSeconds += sessionOverlapSeconds(s, startTime, now)
	}
	return totalSeconds
}

// FormatDuration 将秒数格式化为易读的时间格式 (X小时Y分钟)
func FormatDuration(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60

	if hours > 0 {
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	} else {
		return fmt.Sprintf("%d分钟", minutes)
	}
}

// DeleteTimerSession 根据 ID 删除单个专注记录
func DeleteTimerSession(id int64) error {
	mu.Lock()
	var found bool
	var newSessions []TimerSession

	for _, s := range data.Sessions {
		if s.ID != id {
			newSessions = append(newSessions, s)
		} else {
			found = true
		}
	}

	if !found {
		mu.Unlock()
		log.Printf("[DeleteSession] 未找到ID=%d的计时记录", id)
		return nil // 未找到记录，不视为错误
	}

	// 替换会话列表
	data.Sessions = newSessions
	mu.Unlock()

	err := Save()
	if err != nil {
		log.Printf("[DeleteSession] 保存数据失败: %v", err)
		return err
	}

	log.Printf("[DeleteSession] 成功删除计时记录: id=%d", id)
	return nil
}

// Last24HoursFocusTimeByLabel 返回过去 24 小时各标签的专注时长(秒)。
func Last24HoursFocusTimeByLabel() map[string]int {
	mu.Lock()
	defer mu.Unlock()

	now := nowFunc()
	startTime := now.Add(-24 * time.Hour)

	result := map[string]int{}

	for _, s := range data.Sessions {
		if s.Interrupted || s.EndedAt.IsZero() {
			continue
		}

		overlap := sessionOverlapSeconds(s, startTime, now)
		if overlap == 0 {
			continue
		}

		label := DefaultLabel
		if s.TaskID != nil {
			for _, t := range data.Tasks {
				if t.ID == *s.TaskID {
					if t.Label != "" {
						label = t.Label
					}
					break
				}
			}
		}

		result[label] += overlap
	}

	return result
}
