package ui

import (
	"fmt"

	"tomato_clock/internal/logic"
	"tomato_clock/internal/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// runOnMain 保证在 UI 主线程执行给定函数
func runOnMain(f func()) {
	if f == nil {
		return
	}
	// 使用 fyne.Do 以兼容 2.6+ 新线程模型，避免日志警告
	fyne.Do(f)
}

// ShowFloating 创建始终置顶的悬浮计时小窗
// onEnd 在计时结束（正常完成或中断）后回调，用于刷新统计等。
func ShowFloating(app fyne.App, timer *logic.Timer, sessionID int64, onEnd func()) fyne.Window {
	label := widget.NewLabel("00:00")

	win := app.NewWindow("")

	endBtn := widget.NewButtonWithIcon("结束", theme.MediaStopIcon(), func() {
		// 判断本次计时是否被中断：
		// 仅在倒计时模式且仍有剩余时间时，才标记为中断。
		interrupted := false
		if timer.Mode == logic.ModeCountDown {
			remain := timer.TargetSeconds - timer.ElapsedSeconds()
			if remain > 0 {
				interrupted = true
			}
		}
		_ = model.EndSession(sessionID, interrupted)
		runOnMain(onEnd)

		timer.Stop()
		win.Close()
	})

	cont := container.NewVBox(label, endBtn)

	win.SetContent(container.NewCenter(cont))
	win.SetFixedSize(true)
	win.Resize(fyne.NewSize(140, 90))

	// 监听 tick 通道刷新
	go func() {
		for tick := range timer.Chan() {
			var text string
			if timer.Mode == logic.ModeCountDown {
				text = fmt.Sprintf("%02d:%02d", tick.RemainSeconds/60, tick.RemainSeconds%60)
			} else {
				text = fmt.Sprintf("%02d:%02d", tick.ElapsedSeconds/60, tick.ElapsedSeconds%60)
			}
			label.SetText(text)
			if tick.Done {
				_ = model.EndSession(sessionID, false)
				runOnMain(onEnd)
			}
		}
	}()

	win.Show()
	return win
}
