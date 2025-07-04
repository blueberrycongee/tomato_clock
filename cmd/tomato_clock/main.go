package main

import (
	"log"

	"tomato_clock/internal/model"
	"tomato_clock/internal/ui"

	"fyne.io/fyne/v2/app"
)

func main() {
	log.Println("开始启动番茄钟应用...")

	if err := model.Init(); err != nil {
		log.Fatalf("数据初始化失败: %v", err)
	}
	log.Println("数据初始化成功")

	a := app.New()
	log.Println("创建应用实例")

	a.Settings().SetTheme(ui.NewMorandiTheme())
	log.Println("设置应用主题")

	log.Println("创建主窗口...")
	win := ui.NewMainWindow(a)
	log.Println("主窗口创建成功，准备显示...")

	win.ShowAndRun()
}
