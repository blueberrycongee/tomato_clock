package ui

import (
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// chineseTheme 尝试使用系统可用的中文字体来渲染文字，避免乱码。
// 若运行平台未找到中文字体，会自动回退到默认主题。
// 这里只重写 Font，其他部分沿用默认主题。

type chineseTheme struct{}

var (
	fontData  []byte
	loadTried bool
)

func tryLoadFont() {
	if loadTried {
		return
	}
	loadTried = true

	// Windows 首选体积较小的 simhei.ttf，减少首次读取等待
	candidates := []string{
		`C:\\Windows\\Fonts\\simhei.ttf`,
		`C:\\Windows\\Fonts\\msyh.ttf`,
		`C:\\Windows\\Fonts\\msyhbd.ttf`,
		`C:\\Windows\\Fonts\\simfang.ttf`,
		`C:\\Windows\\Fonts\\simkai.ttf`,
		`C:\\Windows\\Fonts\\simsun.ttc`,
		`/System/Library/Fonts/PingFang.ttc`,
		`/System/Library/Fonts/Hiragino Sans GB.ttc`,
		`/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc`,
		`/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc`,
		`/usr/share/fonts/truetype/wqy/wqy-zenhei.ttc`,
	}
	for _, p := range candidates {
		if b, err := os.ReadFile(p); err == nil {
			fontData = b
			break
		}
	}
}

// Font 返回指定文本样式的字体资源。
func (chineseTheme) Font(style fyne.TextStyle) fyne.Resource {
	if fontData == nil {
		tryLoadFont()
	}

	if fontData == nil || style.Monospace {
		return theme.DefaultTheme().Font(style)
	}
	return fyne.NewStaticResource("custom_chinese.ttf", fontData)
}

// 其余主题要素直接代理到默认主题
func (chineseTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(n, v)
}

func (chineseTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (chineseTheme) Size(n fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(n)
}

// NewChineseTheme 供外部调用
func NewChineseTheme() fyne.Theme {
	return chineseTheme{}
}
