package audio

import (
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

// AlertPlayer 表示一个提示音播放器
type AlertPlayer struct {
	isInitialized bool
	soundFile     string
	buffer        *beep.Buffer
	isPlaying     bool
	stopChan      chan struct{}
}

// NewAlertPlayer 创建一个新的提示音播放器
func NewAlertPlayer(soundFile string) *AlertPlayer {
	return &AlertPlayer{
		soundFile: soundFile,
		stopChan:  make(chan struct{}),
	}
}

// Init 初始化音频播放器
func (a *AlertPlayer) Init() error {
	if a.isInitialized {
		return nil
	}

	// 初始化音频播放器，采样率44100，缓冲区大小512
	err := speaker.Init(44100, 512)
	if err != nil {
		return err
	}

	a.isInitialized = true
	return nil
}

// LoadSound 加载提示音文件
func (a *AlertPlayer) LoadSound() error {
	if a.buffer != nil {
		return nil // 已经加载过了
	}

	f, err := os.Open(a.soundFile)
	if err != nil {
		return err
	}
	defer f.Close()

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return err
	}
	defer streamer.Close()

	// 创建一个缓冲区以存储所有音频
	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)
	a.buffer = buffer

	return nil
}

// PlayLoop 循环播放提示音，直到调用Stop
func (a *AlertPlayer) PlayLoop() error {
	// 确保已初始化
	if err := a.Init(); err != nil {
		return err
	}

	// 确保声音文件已加载
	if err := a.LoadSound(); err != nil {
		return err
	}

	if a.isPlaying {
		return nil // 已经在播放中
	}

	a.isPlaying = true
	a.stopChan = make(chan struct{})

	// 在新协程中处理循环播放
	go func() {
		for {
			select {
			case <-a.stopChan:
				a.isPlaying = false
				return
			default:
				streamer := a.buffer.Streamer(0, a.buffer.Len())
				speaker.Play(beep.Seq(streamer, beep.Callback(func() {
					// 播放完成后延迟一小段时间
					time.Sleep(500 * time.Millisecond)
				})))

				// 等待当前播放完成后再继续循环
				time.Sleep(time.Duration(a.buffer.Len()) * time.Second / 44100)
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	return nil
}

// Stop 停止循环播放
func (a *AlertPlayer) Stop() {
	if a.isPlaying {
		// 通知播放协程退出
		close(a.stopChan)
		a.isPlaying = false

		// 使用speaker.Clear()清除所有正在播放的音频
		speaker.Clear()
	}
}
