package ui

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"tomato_clock/internal/agent"
	"tomato_clock/internal/audio"
	"tomato_clock/internal/config"
	"tomato_clock/internal/logic"
	"tomato_clock/internal/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// 全局饼图变量
var pieChart24h *PieChart
var pieChartToday *PieChart
var studyGoalLabel = "学习"       // 可以配置的目标标签
var studyGoalSeconds = 8 * 3600 // 8小时目标

// 全局音频播放器
var alertPlayer *audio.AlertPlayer // 倒计时结束提示音（alert.mp3）
var hintPlayer *audio.AlertPlayer  // 随机提示音（alarm.mp3）

// 全局设置
var (
	muteAlerts = false // 是否静音提示音
	// 随机提示音功能开关
	randomHintEnabled = false
)

// 初始化音频播放器
func initAudioPlayer(app fyne.App) {
	log.Printf("[DEBUG] 开始初始化音频播放器")

	// 获取应用程序目录
	dir, err := filepath.Abs(filepath.Dir("./"))
	if err != nil {
		log.Printf("[ERROR] 无法获取应用程序目录: %v", err)
		return
	}

	// 设置音频文件路径（倒计时结束）
	alertSoundPath := filepath.Join(dir, "resources", "sounds", "alert.mp3")
	log.Printf("[DEBUG] 音频文件路径: %s", alertSoundPath)

	// 检查文件是否存在
	if _, err := os.Stat(alertSoundPath); os.IsNotExist(err) {
		log.Printf("[ERROR] 提示音文件不存在: %s", alertSoundPath)
		// 创建resources/sounds目录
		soundDir := filepath.Join(dir, "resources", "sounds")
		if err := os.MkdirAll(soundDir, 0755); err != nil {
			log.Printf("[ERROR] 无法创建声音目录: %v", err)
		}
		// 显示错误消息
		log.Printf("[WARNING] 请在 %s 放置一个名为alert.mp3的声音文件", soundDir)
		return
	}

	// 创建音频播放器
	alertPlayer = audio.NewAlertPlayer(alertSoundPath)
	log.Printf("[DEBUG] 音频播放器已创建")

	// 尝试初始化音频系统
	err = alertPlayer.Init()
	if err != nil {
		log.Printf("[ERROR] 无法初始化音频系统: %v", err)
		alertPlayer = nil // 设置为nil以防止后续使用
		return
	}

	log.Printf("[DEBUG] 音频系统初始化成功")

	// 预加载提示音文件
	err = alertPlayer.LoadSound()
	if err != nil {
		log.Printf("[ERROR] 加载提示音文件失败: %v", err)
		alertPlayer = nil // 设置为nil以防止后续使用
	} else {
		log.Printf("[DEBUG] 提示音文件已成功加载")
	}

	// -------------- 初始化随机提示音播放器 --------------
	alarmSoundPath := filepath.Join(dir, "resources", "sounds", "alarm.mp3")
	log.Printf("[DEBUG] 随机提示音文件路径: %s", alarmSoundPath)

	if _, err := os.Stat(alarmSoundPath); os.IsNotExist(err) {
		log.Printf("[WARNING] 随机提示音文件不存在: %s", alarmSoundPath)
	} else {
		hintPlayer = audio.NewAlertPlayer(alarmSoundPath)
		if err := hintPlayer.Init(); err != nil {
			log.Printf("[ERROR] 初始化随机提示音播放器失败: %v", err)
			hintPlayer = nil
		} else if err := hintPlayer.LoadSound(); err != nil {
			log.Printf("[ERROR] 加载随机提示音文件失败: %v", err)
			hintPlayer = nil
		}
	}
}

// 显示编辑专注记录的弹出式表单
func showSessionEditDialog(session model.TimerSession, w fyne.Window, updateCallback func()) {
	log.Printf("[DEBUG] 显示编辑对话框: session.ID=%d, Mode=%s, TaskID=%v",
		session.ID, session.Mode, session.TaskID)

	// 创建表单项
	// 1. 关联任务下拉框
	taskOptions := []string{"自由计时"}
	allTasks, _ := model.ListTasks()
	log.Printf("[DEBUG] 加载任务列表: 共%d个任务", len(allTasks))

	for _, t := range allTasks {
		taskOptions = append(taskOptions, t.Title)
	}

	taskSelect := widget.NewSelect(taskOptions, nil)
	if session.TaskID == nil {
		taskSelect.SetSelected("自由计时")
		log.Printf("[DEBUG] 设置默认任务: 自由计时")
	} else {
		// 查找并设置当前关联的任务
		for _, t := range allTasks {
			if t.ID == *session.TaskID {
				taskSelect.SetSelected(t.Title)
				log.Printf("[DEBUG] 设置当前任务: ID=%d, Title=%s", t.ID, t.Title)
				break
			}
		}
	}

	// 2. 计时模式单选按钮
	modeRadio := widget.NewRadioGroup([]string{"正计时", "倒计时"}, nil)
	if session.Mode == "countdown" {
		modeRadio.SetSelected("倒计时")
		log.Printf("[DEBUG] 设置计时模式: 倒计时")
	} else {
		modeRadio.SetSelected("正计时")
		log.Printf("[DEBUG] 设置计时模式: 正计时")
	}

	// 3. 目标时长（分钟）
	targetMinuteEntry := widget.NewEntry()
	targetMinuteEntry.SetText(fmt.Sprintf("%d", session.TargetSeconds/60))
	log.Printf("[DEBUG] 设置目标时长: %d分钟", session.TargetSeconds/60)

	// 4. 开始时间（日期+时间）
	startDate := widget.NewEntry()
	startDate.SetText(session.StartedAt.Format("2006-01-02"))
	startTime := widget.NewEntry()
	startTime.SetText(session.StartedAt.Format("15:04:05"))
	log.Printf("[DEBUG] 设置开始时间: %s", session.StartedAt.Format("2006-01-02 15:04:05"))

	// 5. 结束时间（日期+时间）
	endDate := widget.NewEntry()
	endDate.SetText(session.EndedAt.Format("2006-01-02"))
	endTime := widget.NewEntry()
	endTime.SetText(session.EndedAt.Format("15:04:05"))
	log.Printf("[DEBUG] 设置结束时间: %s", session.EndedAt.Format("2006-01-02 15:04:05"))

	// 6. 实际持续时间（小时和分钟）
	hours := session.DurationSec / 3600
	minutes := (session.DurationSec % 3600) / 60

	hoursEntry := widget.NewEntry()
	hoursEntry.SetText(fmt.Sprintf("%d", hours))
	minutesEntry := widget.NewEntry()
	minutesEntry.SetText(fmt.Sprintf("%d", minutes))
	log.Printf("[DEBUG] 设置持续时间: %d小时%d分钟", hours, minutes)

	// 自动计算持续时间的函数
	calculateDuration := func() {
		// 尝试解析开始和结束时间
		startDateStr := startDate.Text
		startTimeStr := startTime.Text
		endDateStr := endDate.Text
		endTimeStr := endTime.Text

		log.Printf("[DEBUG] 计算持续时间 - 输入: 开始=%s %s, 结束=%s %s",
			startDateStr, startTimeStr, endDateStr, endTimeStr)

		startStr := startDateStr + " " + startTimeStr
		endStr := endDateStr + " " + endTimeStr

		// 使用 ParseInLocation 确保解析结果带有本地时区，避免时区丢失
		startT, err1 := time.ParseInLocation("2006-01-02 15:04:05", startStr, time.Local)
		endT, err2 := time.ParseInLocation("2006-01-02 15:04:05", endStr, time.Local)

		if err1 != nil {
			log.Printf("[DEBUG] 开始时间解析错误: %v", err1)
		}
		if err2 != nil {
			log.Printf("[DEBUG] 结束时间解析错误: %v", err2)
		}

		if err1 == nil && err2 == nil {
			// 计算持续时间（秒）
			durationSec := int(endT.Sub(startT).Seconds())
			log.Printf("[DEBUG] 计算持续时间: %d秒", durationSec)
			if durationSec >= 0 {
				hours := durationSec / 3600
				minutes := (durationSec % 3600) / 60
				hoursEntry.SetText(fmt.Sprintf("%d", hours))
				minutesEntry.SetText(fmt.Sprintf("%d", minutes))
				log.Printf("[DEBUG] 更新持续时间字段: %d小时%d分钟", hours, minutes)
			} else {
				log.Printf("[DEBUG] 持续时间为负值，不更新")
			}
		}
	}

	// 为时间字段添加变更回调来自动计算持续时间
	startDate.OnChanged = func(s string) {
		log.Printf("[DEBUG] 开始日期变更: %s", s)
		calculateDuration()
	}
	startTime.OnChanged = func(s string) {
		log.Printf("[DEBUG] 开始时间变更: %s", s)
		calculateDuration()
	}
	endDate.OnChanged = func(s string) {
		log.Printf("[DEBUG] 结束日期变更: %s", s)
		calculateDuration()
	}
	endTime.OnChanged = func(s string) {
		log.Printf("[DEBUG] 结束时间变更: %s", s)
		calculateDuration()
	}

	// 创建对话框
	log.Printf("[DEBUG] 创建表单对话框")
	form := dialog.NewForm(
		"编辑专注记录",
		"保存",
		"取消",
		[]*widget.FormItem{
			widget.NewFormItem("关联任务", taskSelect),
			widget.NewFormItem("计时模式", modeRadio),
			widget.NewFormItem("目标时长(分钟)", targetMinuteEntry),
			widget.NewFormItem("开始日期", startDate),
			widget.NewFormItem("开始时间", startTime),
			widget.NewFormItem("结束日期", endDate),
			widget.NewFormItem("结束时间", endTime),
			widget.NewFormItem("持续时间(小时)", hoursEntry),
			widget.NewFormItem("持续时间(分钟)", minutesEntry),
		},
		func(saved bool) {
			if !saved {
				log.Printf("[DEBUG] 用户取消表单")
				return
			}

			log.Printf("[DEBUG] 表单提交: 尝试验证和保存")

			// 验证和保存修改
			// 1. 解析目标时长（分钟）
			targetMinutes, err := strconv.Atoi(targetMinuteEntry.Text)
			if err != nil {
				log.Printf("[DEBUG] 目标时长验证失败: %v", err)
				dialog.ShowError(errors.New("目标时长必须是有效的整数"), w)
				return
			}
			// 转换为秒
			targetSec := targetMinutes * 60
			log.Printf("[DEBUG] 解析目标时长: %d分钟 (%d秒)", targetMinutes, targetSec)

			// 2. 解析开始时间和结束时间
			startStr := startDate.Text + " " + startTime.Text
			endStr := endDate.Text + " " + endTime.Text

			log.Printf("[DEBUG] 尝试解析开始时间: %s", startStr)
			startedAt, err := time.ParseInLocation("2006-01-02 15:04:05", startStr, time.Local)
			if err != nil {
				log.Printf("[DEBUG] 开始时间解析失败: %v", err)
				dialog.ShowError(errors.New("开始时间格式无效，请使用YYYY-MM-DD格式的日期和HH:MM:SS格式的时间"), w)
				return
			}

			log.Printf("[DEBUG] 尝试解析结束时间: %s", endStr)
			endedAt, err := time.ParseInLocation("2006-01-02 15:04:05", endStr, time.Local)
			if err != nil {
				log.Printf("[DEBUG] 结束时间解析失败: %v", err)
				dialog.ShowError(errors.New("结束时间格式无效，请使用YYYY-MM-DD格式的日期和HH:MM:SS格式的时间"), w)
				return
			}

			// 3. 解析持续时间（小时和分钟）
			log.Printf("[DEBUG] 尝试解析持续时间")
			hours, err := strconv.Atoi(hoursEntry.Text)
			if err != nil {
				log.Printf("[DEBUG] 持续时间(小时)解析失败: %v", err)
				dialog.ShowError(errors.New("持续时间(小时)必须是有效的整数"), w)
				return
			}

			minutes, err := strconv.Atoi(minutesEntry.Text)
			if err != nil {
				log.Printf("[DEBUG] 持续时间(分钟)解析失败: %v", err)
				dialog.ShowError(errors.New("持续时间(分钟)必须是有效的整数"), w)
				return
			}

			// 转换为秒
			durationSec := (hours * 3600) + (minutes * 60)
			log.Printf("[DEBUG] 解析持续时间: %d小时%d分钟 (%d秒)", hours, minutes, durationSec)

			// 4. 解析任务ID
			var taskIDPtr *int64
			log.Printf("[DEBUG] 选择的任务: %s", taskSelect.Selected)
			if taskSelect.Selected != "自由计时" {
				for _, t := range allTasks {
					if t.Title == taskSelect.Selected {
						idCopy := t.ID
						taskIDPtr = &idCopy
						log.Printf("[DEBUG] 找到任务ID: %d", *taskIDPtr)
						break
					}
				}
			} else {
				log.Printf("[DEBUG] 任务为自由计时，taskID=nil")
			}

			// 5. 解析模式
			mode := "countup"
			if modeRadio.Selected == "倒计时" {
				mode = "countdown"
			}
			log.Printf("[DEBUG] 计时模式: %s", mode)

			// 创建更新后的记录
			updatedSession := model.TimerSession{
				ID:            session.ID,
				TaskID:        taskIDPtr,
				Mode:          mode,
				TargetSeconds: targetSec,
				StartedAt:     startedAt,
				EndedAt:       endedAt,
				Interrupted:   session.Interrupted,
				DurationSec:   durationSec,
			}

			log.Printf("[DEBUG] 准备更新记录: ID=%d, TaskID=%v, Mode=%s, Duration=%d, Start=%v, End=%v",
				updatedSession.ID, updatedSession.TaskID, updatedSession.Mode,
				updatedSession.DurationSec, updatedSession.StartedAt, updatedSession.EndedAt)

			// 保存更新
			if err := model.UpdateSession(updatedSession); err != nil {
				log.Printf("[DEBUG] 更新失败: %v", err)
				dialog.ShowError(err, w)
				return
			}

			log.Printf("[DEBUG] 更新成功，刷新界面")
			// 刷新界面
			updateCallback()
		},
		w,
	)

	form.Resize(fyne.NewSize(400, 500))
	log.Printf("[DEBUG] 显示表单对话框")
	form.Show()
}

// 函数已经定义，此处删除重复定义

// 函数已删除

// 显示计时结束通知对话框
func showTimerCompletedDialog(w fyne.Window, callback func()) {
	log.Printf("[DEBUG] 显示计时完成对话框")

	confirmDialog := dialog.NewCustomConfirm(
		"番茄钟完成",
		"确认",
		"停止提示音",
		widget.NewLabel("倒计时已结束！"),
		func(confirmed bool) {
			if confirmed {
				// "确认"按钮被点击
				log.Printf("[DEBUG] 用户点击了确认按钮")
			} else {
				// "停止提示音"按钮被点击
				log.Printf("[DEBUG] 用户点击了停止提示音按钮，正在停止提示音...")
				if alertPlayer != nil {
					alertPlayer.Stop()
					log.Printf("[DEBUG] 提示音已停止")
				} else {
					log.Printf("[DEBUG] alertPlayer为空，无法停止提示音")
				}
			}

			if callback != nil {
				callback()
			}
		},
		w,
	)
	confirmDialog.Show()
}

// updatePieCharts 更新两个饼图的数据
func updatePieCharts() {
	// 1. 更新24小时专注占比图
	durations := model.Last24HoursFocusTimeByLabel()
	var segments24h []PieChartSegment
	for label, sec := range durations {
		if sec > 0 {
			segments24h = append(segments24h, PieChartSegment{Label: label, Value: float64(sec)})
		}
	}
	if pieChart24h != nil {
		pieChart24h.UpdateData(segments24h)
	}

	// 2. 更新今日学习目标图
	todayStudySec := 0
	now := time.Now()
	year, month, day := now.Date()
	startOfToday := time.Date(year, month, day, 0, 0, 0, 0, now.Location())

	// 这个计算需要访问 taskLabelMap，确保它已经被填充
	// 我们将在 updateHistory 中调用此函数，届时 taskLabelMap 是最新的
	allSessions := model.CompletedSessions()
	allTasks, _ := model.ListTasks()
	taskLabelMap := make(map[int64]string)
	for _, t := range allTasks {
		taskLabelMap[t.ID] = t.Label
	}

	for _, s := range allSessions {
		if s.EndedAt.After(startOfToday) {
			if s.TaskID != nil {
				if label, ok := taskLabelMap[*s.TaskID]; ok && label == studyGoalLabel {
					todayStudySec += s.DurationSec
				}
			}
		}
	}

	completed := float64(todayStudySec)
	remaining := math.Max(0, float64(studyGoalSeconds)-completed)

	segmentsToday := []PieChartSegment{
		{Label: "已完成", Value: completed},
		{Label: "未完成", Value: remaining},
	}
	if pieChartToday != nil {
		pieChartToday.UpdateData(segmentsToday)
		// 更新标题以显示具体时间
		pieChartToday.titleLabel.SetText(
			fmt.Sprintf("今日%s目标 (%s/%s)",
				studyGoalLabel,
				model.FormatDuration(int(completed)),
				model.FormatDuration(studyGoalSeconds),
			),
		)
	}
}

// createPieChartsPanel 创建饼图面板
func createPieChartsPanel() fyne.CanvasObject {
	pieChart24h = NewPieChart("过去24小时专注占比", []PieChartSegment{})
	pieChartToday = NewPieChart(fmt.Sprintf("今日%s目标", studyGoalLabel), []PieChartSegment{})

	// 立即进行一次初始更新
	updatePieCharts()

	grid := container.NewGridWithColumns(2, pieChart24h, pieChartToday)
	return grid
}

func NewMainWindow(app fyne.App) fyne.Window {
	w := app.NewWindow("Tomato Clock")

	// 初始化音频播放器
	initAudioPlayer(app)

	tasks, err := model.ListTasks()
	if err != nil {
		log.Printf("加载任务失败: %v", err)
	}

	// 未来可保存所选任务以记录计时
	var selectedTask *model.Task

	var updateHistory func()
	var updateStats func()

	var list *widget.List
	list = widget.NewList(
		func() int { return len(tasks) },
		func() fyne.CanvasObject {
			chk := widget.NewCheck("", nil)
			lbl := widget.NewLabel("title")
			edit := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
			edit.Importance = widget.LowImportance
			edit.Resize(fyne.NewSize(24, 24))
			del := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			del.Importance = widget.LowImportance
			del.Resize(fyne.NewSize(24, 24))
			colorRect := canvas.NewRectangle(ColorForLabel(""))
			colorRect.SetMinSize(fyne.NewSize(10, 10))
			return container.NewHBox(chk, lbl, layout.NewSpacer(), edit, del, colorRect)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			h := o.(*fyne.Container)
			chk := h.Objects[0].(*widget.Check)
			lbl := h.Objects[1].(*widget.Label)
			editBtn := h.Objects[3].(*widget.Button)
			delBtn := h.Objects[4].(*widget.Button)
			colorRect := h.Objects[5].(*canvas.Rectangle)
			if i < 0 || i >= len(tasks) {
				return
			}
			t := tasks[i]
			chk.SetChecked(t.IsDone)
			lbl.SetText(t.Title)
			colorRect.FillColor = ColorForLabel(t.Label)
			colorRect.Refresh()

			// 编辑任务按钮
			editBtn.OnTapped = func() {
				titleEntry := widget.NewEntry()
				titleEntry.SetText(t.Title)
				labelEntry := widget.NewEntry()
				labelEntry.SetText(t.Label)
				dialog.ShowForm("编辑任务", "保存", "取消",
					[]*widget.FormItem{
						widget.NewFormItem("标题", titleEntry),
						widget.NewFormItem("标签", labelEntry),
					},
					func(confirm bool) {
						if !confirm || titleEntry.Text == "" {
							return
						}
						updated := t
						updated.Title = titleEntry.Text
						updated.Label = labelEntry.Text
						if err := model.UpdateTask(updated); err != nil {
							dialog.ShowError(err, w)
							return
						}
						tasks, _ = model.ListTasks()
						list.Refresh()
						if updateHistory != nil {
							updateHistory()
						}
					}, w)
			}

			delBtn.OnTapped = func() {
				dialog.ShowConfirm("确认删除", fmt.Sprintf("删除任务 '%s'?", t.Title), func(ok bool) {
					if !ok {
						return
					}
					if err := model.DeleteTask(t.ID); err != nil {
						dialog.ShowError(err, w)
						return
					}
					tasks, _ = model.ListTasks()
					list.Refresh()
					updateHistory()
				}, w)
			}
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(tasks) {
			selectedTask = &tasks[id]
		}
	}

	addBtn := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		titleEntry := widget.NewEntry()
		titleEntry.SetPlaceHolder("任务标题")
		labelEntry := widget.NewEntry()
		labelEntry.SetText("学习")
		labelEntry.SetPlaceHolder("任务标签，如：学习/娱乐...")
		dialog.ShowForm("新建任务", "创建", "取消",
			[]*widget.FormItem{
				widget.NewFormItem("标题", titleEntry),
				widget.NewFormItem("标签", labelEntry),
			},
			func(confirm bool) {
				if !confirm || titleEntry.Text == "" {
					return
				}
				task := &model.Task{
					Title:      titleEntry.Text,
					IsDone:     false,
					RepeatRule: model.RepeatNone,
					Label:      labelEntry.Text,
				}
				if err := model.CreateTask(task); err != nil {
					dialog.ShowError(err, w)
					return
				}
				tasks, _ = model.ListTasks()
				list.Refresh()
				if updateHistory != nil {
					updateHistory()
				}
			}, w)
	})
	addBtn.Importance = widget.LowImportance
	// 使用 24x24 尺寸以保持与输入框同高
	addBtn.Resize(fyne.NewSize(24, 24))

	// 新建倒数日按钮
	countdownBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {
		titleEntry := widget.NewEntry()
		titleEntry.SetPlaceHolder("事件标题")
		dateEntry := widget.NewEntry()
		dateEntry.SetPlaceHolder("YYYY-MM-DD")
		labelEntry := widget.NewEntry()
		labelEntry.SetText("重要")
		labelEntry.SetPlaceHolder("标签，可选")
		dialog.ShowForm("新建倒数日", "创建", "取消",
			[]*widget.FormItem{
				widget.NewFormItem("标题", titleEntry),
				widget.NewFormItem("日期", dateEntry),
				widget.NewFormItem("标签", labelEntry),
			},
			func(confirm bool) {
				if !confirm || titleEntry.Text == "" || dateEntry.Text == "" {
					return
				}
				due, err := time.Parse("2006-01-02", dateEntry.Text)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				task := &model.Task{
					Title:      titleEntry.Text,
					IsDone:     false,
					RepeatRule: model.RepeatNone,
					Label:      labelEntry.Text,
					DueDate:    &due,
				}
				if err := model.CreateTask(task); err != nil {
					dialog.ShowError(err, w)
					return
				}
				tasks, _ = model.ListTasks()
				list.Refresh()
				if updateStats != nil {
					updateStats()
				}
			}, w)
	})
	countdownBtn.Importance = widget.LowImportance
	countdownBtn.Resize(fyne.NewSize(24, 24))

	// 清空历史按钮
	clearBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		dialog.ShowConfirm("确认", "确定要清空所有历史专注记录吗？", func(ok bool) {
			if !ok {
				return
			}
			if err := model.ClearSessions(); err != nil {
				dialog.ShowError(err, w)
				return
			}
			if updateHistory != nil {
				updateHistory()
			}
		}, w)
	})
	clearBtn.Importance = widget.LowImportance
	clearBtn.Resize(fyne.NewSize(24, 24))

	// 将三个小按钮包裹为固定24×24大小，保证与输入框在同一水平线
	gridAdd := container.NewGridWrap(fyne.NewSize(24, 24), addBtn)
	gridCountdown := container.NewGridWrap(fyne.NewSize(24, 24), countdownBtn)
	gridClear := container.NewGridWrap(fyne.NewSize(24, 24), clearBtn)
	smallBtns := container.NewHBox(gridAdd, gridCountdown, gridClear)

	// 实时系统时间标签
	clockLabel := widget.NewLabel("")
	go func() {
		for now := range time.Tick(time.Second) {
			tStr := now.Format("2006-01-02 15:04:05")
			runOnMain(func() { clockLabel.SetText(tStr) })
		}
	}()

	// --- 计时记录列表 ---
	var sessions []model.TimerSession
	var taskTitleMap map[int64]string
	var taskLabelMap map[int64]string

	// 双击相关状态
	var lastClickID widget.ListItemID = -1
	var lastClickTime time.Time

	sessionList := widget.NewList(
		func() int { return len(sessions) },
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("record")
			lbl.Wrapping = fyne.TextTruncate

			// 创建按钮容器
			buttonBox := container.NewHBox()

			// 创建一个编辑按钮
			editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
			editBtn.Importance = widget.LowImportance
			editBtn.Resize(fyne.NewSize(24, 24))

			// 创建一个删除按钮
			delBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			delBtn.Importance = widget.LowImportance
			delBtn.Resize(fyne.NewSize(24, 24))

			// 添加按钮
			buttonBox.Add(editBtn)
			buttonBox.Add(delBtn)

			// 颜色指示矩形
			colorRect := canvas.NewRectangle(ColorForLabel(""))
			colorRect.SetMinSize(fyne.NewSize(10, 10))

			// 右侧容器包含按钮和颜色块
			rightBox := container.NewHBox(buttonBox, colorRect)

			return container.NewBorder(nil, nil, nil, rightBox, lbl)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(sessions) {
				log.Printf("[DEBUG] 列表渲染 - 索引越界: %d (总数: %d)", i, len(sessions))
				return
			}
			s := sessions[i]
			border := o.(*fyne.Container)
			lbl := border.Objects[0].(*widget.Label)
			rightBox := border.Objects[1].(*fyne.Container)
			buttonBox := rightBox.Objects[0].(*fyne.Container)
			colorRect := rightBox.Objects[1].(*canvas.Rectangle)
			editBtn := buttonBox.Objects[0].(*widget.Button)
			delBtn := buttonBox.Objects[1].(*widget.Button)

			// 构造显示文本
			modeStr := "正计时"
			if s.Mode == "countdown" {
				modeStr = "倒计时"
			}
			taskTitle := "自由计时"
			if s.TaskID != nil {
				if title, ok := taskTitleMap[*s.TaskID]; ok {
					taskTitle = title
				}
			}
			// 将秒数转换为更易读的格式
			var durationStr string
			hours := s.DurationSec / 3600
			minutes := (s.DurationSec % 3600) / 60

			if hours > 0 {
				durationStr = fmt.Sprintf("%d小时%d分钟", hours, minutes)
			} else {
				durationStr = fmt.Sprintf("%d分钟", minutes)
			}

			displayText := fmt.Sprintf("%s | %s | %s | %s",
				s.EndedAt.Format("01-02 15:04"), taskTitle, durationStr, modeStr)

			if i < 5 { // 只记录前几项，避免日志太多
				log.Printf("[DEBUG] 渲染列表项 #%d: ID=%d, Text=%s", i, s.ID, displayText)
			}

			lbl.SetText(displayText)

			// 获取标签并更新颜色块
			labelValue := "学习"
			if s.TaskID != nil {
				if l, ok := taskLabelMap[*s.TaskID]; ok {
					labelValue = l
				}
			}
			colorRect.FillColor = ColorForLabel(labelValue)
			colorRect.Refresh()

			// 设置编辑按钮的点击事件
			editBtn.OnTapped = func() {
				log.Printf("[DEBUG] 点击编辑按钮: id=%d, session.ID=%d", i, s.ID)
				showSessionEditDialog(s, w, updateHistory)
			}

			// 设置删除按钮的点击事件
			delBtn.OnTapped = func() {
				log.Printf("[DEBUG] 点击删除按钮: id=%d, session.ID=%d", i, s.ID)
				// 显示确认对话框
				dialog.ShowConfirm(
					"删除专注记录",
					fmt.Sprintf("确定要删除这条专注记录吗？\n%s", displayText),
					func(confirmed bool) {
						if !confirmed {
							return
						}

						// 用户确认删除
						if err := model.DeleteSession(s.ID); err != nil {
							dialog.ShowError(err, w)
							return
						}

						// 更新列表
						log.Printf("[DEBUG] 专注记录已删除，正在刷新列表")
						updateHistory()
					},
					w,
				)
			}
		},
	)

	// 双击检测
	sessionList.OnSelected = func(id widget.ListItemID) {
		now := time.Now()
		log.Printf("[DEBUG] 列表项被选中: id=%d, lastClickID=%d, 时间差=%v毫秒",
			id, lastClickID, now.Sub(lastClickTime).Milliseconds())

		// 增大双击时间阈值到1000毫秒(1秒)，使双击检测更容易被触发
		if id == lastClickID && now.Sub(lastClickTime) < 1000*time.Millisecond {
			if id >= 0 && id < len(sessions) {
				log.Printf("[DEBUG] 检测到双击: id=%d, session.ID=%d", id, sessions[id].ID)
				// 双击打开编辑表单，而不是直接进入行内编辑模式
				showSessionEditDialog(sessions[id], w, updateHistory)
				// 重置双击状态，避免连续多次触发
				lastClickID = -1
				lastClickTime = time.Time{}
				return
			}
		}
		lastClickID = id
		lastClickTime = now
	}

	updateHistory = func() {
		log.Printf("[DEBUG] 开始更新历史记录列表")
		sessions = model.CompletedSessions()
		log.Printf("[DEBUG] 加载了%d条完成的专注记录", len(sessions))

		// build task title map
		taskTitleMap = map[int64]string{}
		taskLabelMap = map[int64]string{}
		allTasks, _ := model.ListTasks()
		for _, t := range allTasks {
			taskTitleMap[t.ID] = t.Title
			taskLabelMap[t.ID] = t.Label
		}
		model.PrintSessionsSummary() // 同步输出到终端

		log.Printf("[DEBUG] 刷新列表显示")
		sessionList.Refresh()
		// 更新顶部统计信息
		if updateStats != nil {
			updateStats()
		}
	}

	// 初始化一次
	updateHistory()

	// 模式选择、时长输入、开始按钮
	modeRadio := widget.NewRadioGroup([]string{"正计时", "倒计时"}, func(string) {})
	modeRadio.Horizontal = true
	modeRadio.SetSelected("倒计时")

	minuteEntry := widget.NewEntry()
	minuteEntry.SetText("25")
	minuteEntry.Validator = func(s string) error { _, err := strconv.Atoi(s); return err }

	var runningTimer *logic.Timer
	var currentSessionID int64
	var timerLabel = widget.NewLabel("")
	timerLabel.Hide()
	var stopBtn *widget.Button

	// 随机提示音取消通道
	var randomHintCancel chan struct{}

	startBtn := widget.NewButtonWithIcon("开始", theme.MediaPlayIcon(), func() {
		if runningTimer != nil {
			return // already running
		}
		var mode string
		if modeRadio.Selected == "倒计时" {
			mode = logic.ModeCountDown
		} else {
			mode = logic.ModeCountUp
		}

		mins, _ := strconv.Atoi(minuteEntry.Text)
		secs := mins * 60
		if mode == logic.ModeCountDown && secs <= 0 {
			dialog.ShowError(errors.New("请输入大于0的分钟数"), w)
			return
		}

		var taskIDPtr *int64
		if selectedTask != nil {
			taskIDPtr = &selectedTask.ID
		}

		sessionID, err := model.StartSession(taskIDPtr, mode, secs)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		runningTimer = logic.NewTimer(mode, secs)
		runningTimer.Start()

		// show timer label & stop button
		timerLabel.Show()
		stopBtn.Enable()

		currentSessionID = sessionID

		// 如果启用随机提示音功能，为本次计时创建调度
		if randomHintEnabled {
			randomHintCancel = make(chan struct{})
			log.Printf("[RANDOM] 随机提示音调度已启动，模式=%s，目标时长=%d秒", mode, secs)
			go func(t *logic.Timer, cancelCh <-chan struct{}) {
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				for {
					// 生成5-10分钟的随机间隔
					delayMin := r.Intn(6) + 5 // [5,10]
					delay := time.Duration(delayMin) * time.Minute
					log.Printf("[RANDOM] 下一次随机提示音将在 %d 分钟后触发", delayMin)
					select {
					case <-time.After(delay):
						// 检查是否被取消
						select {
						case <-cancelCh:
							log.Printf("[RANDOM] 收到取消信号，随机提示音调度结束")
							return
						default:
						}

						// 倒计时模式下，若已不足10分钟则不再播放，直接退出循环
						if t.Mode == logic.ModeCountDown {
							remain := t.TargetSeconds - t.ElapsedSeconds()
							if remain <= 600 {
								log.Printf("[RANDOM] 剩余时间 %d 秒 ≤ 600，跳过并结束随机提示音调度", remain)
								return
							}
						}

						if !muteAlerts && hintPlayer != nil {
							if err := hintPlayer.PlayFor(10 * time.Second); err != nil {
								log.Printf("[ERROR] 随机提示音播放失败: %v", err)
							}
						}

						log.Printf("[RANDOM] 已播放随机提示音 10 秒片段")

					case <-cancelCh:
						log.Printf("[RANDOM] 收到取消信号，随机提示音调度结束")
						return
					}
				}
			}(runningTimer, randomHintCancel)
		}

		go func(sessID int64, mode string) {
			for tick := range runningTimer.Chan() {
				// update label
				runOnMain(func() {
					if mode == logic.ModeCountDown {
						timerLabel.SetText(fmt.Sprintf("%02d:%02d", tick.RemainSeconds/60, tick.RemainSeconds%60))
					} else {
						timerLabel.SetText(fmt.Sprintf("%02d:%02d", tick.ElapsedSeconds/60, tick.ElapsedSeconds%60))
					}
				})
				if tick.Done {
					// complete session
					_ = model.EndSession(sessID, false)

					// 如果是倒计时模式，播放提示音并显示通知
					if mode == logic.ModeCountDown {
						// 在主线程上处理提示音和对话框
						runOnMain(func() {
							// 先停止任何可能正在播放的提示音
							if hintPlayer != nil {
								hintPlayer.Stop()
							}

							// 如果没有静音，播放提示音
							if !muteAlerts && hintPlayer != nil {
								log.Printf("[DEBUG] 倒计时结束，开始播放提示音...")
								if err := hintPlayer.PlayLoop(); err != nil {
									log.Printf("[ERROR] 播放提示音失败: %v", err)
								} else {
									log.Printf("[DEBUG] 提示音开始播放")
								}
							} else if muteAlerts {
								log.Printf("[DEBUG] 倒计时结束，但静音已启用，不播放提示音")
							}

							// 显示通知对话框
							showTimerCompletedDialog(w, nil)
						})
					}

					runOnMain(func() {
						runningTimer = nil
						timerLabel.Hide()
						stopBtn.Disable()
						// 停止随机提示音调度
						if randomHintCancel != nil {
							close(randomHintCancel)
							randomHintCancel = nil
						}
						updateHistory()
					})
				}
			}
		}(sessionID, mode)
	})

	// 停止按钮
	stopBtn = widget.NewButtonWithIcon("结束", theme.MediaStopIcon(), func() {
		if runningTimer == nil {
			return
		}
		interrupted := false
		if runningTimer.Mode == logic.ModeCountDown {
			remain := runningTimer.TargetSeconds - runningTimer.ElapsedSeconds()
			if remain > 0 {
				interrupted = true
			}
		}
		_ = model.EndSession(currentSessionID, interrupted)
		runningTimer.Stop()
		runningTimer = nil
		timerLabel.Hide()
		stopBtn.Disable()

		// 停止任何正在播放的提示音
		if hintPlayer != nil {
			log.Printf("[DEBUG] 停止按钮被点击，正在停止所有提示音")
			hintPlayer.Stop()
			log.Printf("[DEBUG] 提示音已停止")
		}

		// 停止随机提示音调度
		if randomHintCancel != nil {
			close(randomHintCancel)
			randomHintCancel = nil
		}

		updateHistory()
	})
	stopBtn.Disable()

	// 创建随机提示音按钮
	var randomBtn *widget.Button
	randomBtn = widget.NewButton("随机提示音: 关", func() {
		randomHintEnabled = !randomHintEnabled
		if randomHintEnabled {
			randomBtn.SetText("随机提示音: 开")
		} else {
			randomBtn.SetText("随机提示音: 关")
		}
	})
	randomBtn.Importance = widget.LowImportance

	// 创建静音按钮
	var muteBtn *widget.Button
	muteBtn = widget.NewButtonWithIcon("静音", theme.VolumeUpIcon(), func() {
		muteAlerts = !muteAlerts
		if muteAlerts {
			log.Printf("[DEBUG] 用户启用静音，设置图标为静音")
			muteBtn.SetIcon(theme.VolumeMuteIcon())

			// 立即停止正在播放的提示音
			if hintPlayer != nil {
				log.Printf("[DEBUG] 尝试停止正在播放的提示音")
				hintPlayer.Stop()
				log.Printf("[DEBUG] 提示音已停止")
			}
		} else {
			log.Printf("[DEBUG] 用户关闭静音，设置图标为有声")
			muteBtn.SetIcon(theme.VolumeUpIcon())
		}
	})
	muteBtn.Importance = widget.LowImportance

	timeRow := container.NewHBox(layout.NewSpacer(), timerLabel, stopBtn, muteBtn, randomBtn)

	controlBar := container.NewVBox(container.NewHBox(modeRadio, widget.NewLabel("时长(分钟):"), minuteEntry, startBtn), timeRow)

	// 创建统计信息标签，并封装更新函数
	statsLabel := widget.NewLabel("")
	statsLabel.TextStyle = fyne.TextStyle{Bold: true}

	updateStats = func() {
		var parts []string
		// --- 专注统计 ---
		durations := model.Last24HoursFocusTimeByLabel()
		if len(durations) == 0 {
			parts = append(parts, "过去24小时专注: 暂无数据")
		} else {
			for label, sec := range durations {
				if sec == 0 {
					continue
				}
				parts = append(parts, fmt.Sprintf("过去24小时%s: %s", label, model.FormatDuration(sec)))
			}
		}

		// --- 倒数日统计 ---
		now := time.Now()
		allTasks, _ := model.ListTasks()
		for _, t := range allTasks {
			if t.DueDate == nil {
				continue
			}
			days := int(t.DueDate.Sub(now).Hours() / 24)
			if days < 0 {
				continue // 已过期
			}
			weeks := int(math.Ceil(float64(days) / 7.0))
			months := int(math.Ceil(float64(days) / 30.0))
			parts = append(parts, fmt.Sprintf("距离%s: %d天 / %d周 / %d月", t.Title, days, weeks, months))
		}

		sort.Strings(parts)
		statsLabel.SetText(strings.Join(parts, "\n"))

		// 更新饼图
		updatePieCharts()
	}
	// 初次计算
	updateStats()

	// 创建刷新按钮
	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		if updateStats != nil {
			updateStats()
		}
	})
	refreshBtn.Importance = widget.LowImportance

	// 初始化配置并启动 AI 代理
	var initAPIKey string
	if cfg, err := config.Load(); err == nil {
		initAPIKey = cfg.APIKey
	} else {
		log.Printf("[INFO] 未找到配置文件或读取失败: %v", err)
	}

	aiMgr := agent.Get()
	if initAPIKey != "" {
		if err := aiMgr.Start(initAPIKey); err != nil {
			log.Printf("[ERROR] 启动 AI 代理失败: %v", err)
		}
	}

	// 创建 API Key 输入框（密码模式）
	entryApiKey := widget.NewPasswordEntry()
	entryApiKey.SetPlaceHolder("DeepSeek API Key")
	entryApiKey.SetText(initAPIKey)

	btnSaveKey := widget.NewButton("保存Key", func() {
		key := strings.TrimSpace(entryApiKey.Text)
		if key == "" {
			dialog.ShowError(errors.New("API Key 不能为空"), w)
			return
		}
		if err := config.Save(key); err != nil {
			dialog.ShowError(err, w)
			return
		}
		if err := aiMgr.Start(key); err != nil {
			dialog.ShowError(err, w)
			return
		}
		dialog.ShowInformation("保存成功", "API Key 已保存并生效", w)
	})

	// 创建聊天输入框和发送按钮
	entryChat := widget.NewEntry()
	entryChat.SetPlaceHolder("向 DeepSeek 提问…")

	sendChat := func() {
		msg := strings.TrimSpace(entryChat.Text)
		if msg == "" {
			return
		}
		entryChat.Disable()
		go func(q string) {
			reply, err := aiMgr.SendMessage(q)
			runOnMain(func() {
				entryChat.Enable()
			})
			if err != nil {
				runOnMain(func() {
					dialog.ShowError(err, w)
				})
				return
			}
			// reload data and refresh UI
			_ = model.Reload()
			runOnMain(func() {
				if updateHistory != nil {
					updateHistory()
				}
				dialog.ShowInformation("AI 回复", reply, w)
			})
		}(msg)
	}

	entryChat.OnSubmitted = func(_ string) { sendChat() }
	btnSendChat := widget.NewButton("发送", sendChat)

	// 顶栏：新建任务 + 倒数日 + 清空 + API Key/Chat + 统计 + 时钟
	// 移除 aiBtn
	// 使用 GridWrap 对输入框进行固定宽度包装，使其在 HBox 中占据更大空间
	apiKeyWrapped := container.NewGridWrap(fyne.NewSize(220, entryApiKey.MinSize().Height), entryApiKey)
	chatWrapped := container.NewGridWrap(fyne.NewSize(340, entryChat.MinSize().Height), entryChat)

	// 使保存按钮与发送按钮高度与输入框一致
	saveWrapped := container.NewGridWrap(fyne.NewSize(btnSaveKey.MinSize().Width, entryApiKey.MinSize().Height), btnSaveKey)
	sendWrapped := container.NewGridWrap(fyne.NewSize(btnSendChat.MinSize().Width, entryChat.MinSize().Height), btnSendChat)

	topBar := container.NewHBox(smallBtns, layout.NewSpacer(), apiKeyWrapped, saveWrapped, widget.NewSeparator(), chatWrapped, sendWrapped, widget.NewSeparator(), statsLabel, refreshBtn, widget.NewSeparator(), clockLabel)

	// 主区域改为左右分栏：任务列表 + 饼图 + 历史记录
	chartsPanel := createPieChartsPanel()
	leftPanel := container.NewVSplit(list, chartsPanel)
	leftPanel.Offset = 0.7 // 70%给任务列表，30%给饼图

	split := container.NewHSplit(leftPanel, sessionList)
	split.Offset = 0.4 // 40%给左侧面板，60%给历史记录

	content := container.NewBorder(topBar, controlBar, nil, nil, split)

	w.SetContent(content)
	w.Resize(fyne.NewSize(900, 600)) // 增大窗口尺寸以容纳新组件

	return w
}
