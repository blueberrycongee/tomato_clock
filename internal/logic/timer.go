package logic

import (
	"sync"
	"time"
)

// Mode 表示计时器模式
const (
	ModeCountUp   = "countup"
	ModeCountDown = "countdown"
)

// Tick 表示一次计时器更新
type Tick struct {
	ElapsedSeconds int  // 已用秒数
	RemainSeconds  int  // 剩余秒数（倒计时模式时）
	Done           bool // 是否计时结束
}

// Timer 实现可暂停/继续的计数器
// Start 后会每秒向 Chan 发送 Tick
// Stop 后 Chan 会关闭

type Timer struct {
	Mode          string
	TargetSeconds int

	tickCh  chan Tick
	stopCh  chan struct{}
	pauseCh chan bool

	mu         sync.Mutex
	startedAt  time.Time
	paused     bool
	elapsedSec int
}

// NewTimer 创建计时器
func NewTimer(mode string, targetSeconds int) *Timer {
	return &Timer{
		Mode:          mode,
		TargetSeconds: targetSeconds,
		tickCh:        make(chan Tick, 1),
		stopCh:        make(chan struct{}),
		pauseCh:       make(chan bool),
	}
}

// Chan 返回 tick 通道（只读）
func (t *Timer) Chan() <-chan Tick { return t.tickCh }

// Start 启动计时协程
func (t *Timer) Start() {
	t.startedAt = time.Now()
	go t.loop()
}

// Pause 暂停或继续
func (t *Timer) Pause() { t.pauseCh <- true }

// Stop 停止计时并关闭通道
func (t *Timer) Stop() { close(t.stopCh) }

func (t *Timer) loop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-t.stopCh:
			close(t.tickCh)
			return
		case <-t.pauseCh:
			t.mu.Lock()
			t.paused = !t.paused
			t.mu.Unlock()
		case now := <-ticker.C:
			t.mu.Lock()
			if t.paused {
				t.mu.Unlock()
				continue
			}
			t.elapsedSec = int(now.Sub(t.startedAt).Seconds())
			elapsed := t.elapsedSec
			remain := t.TargetSeconds - elapsed
			done := false
			if t.Mode == ModeCountDown && remain <= 0 {
				remain = 0
				done = true
			}
			t.mu.Unlock()
			tick := Tick{ElapsedSeconds: elapsed, RemainSeconds: remain, Done: done}
			t.tickCh <- tick
			if done {
				close(t.stopCh)
				close(t.tickCh)
				return
			}
		}
	}
}

// ElapsedSeconds 返回已计时的秒数（线程安全）
func (t *Timer) ElapsedSeconds() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.elapsedSec
}
