package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// MorandiTheme 使用柔和莫兰迪配色
// 参考色卡：灰绿、灰蓝、灰紫等低饱和度颜色
// 仅重写常用颜色，其余沿用默认主题

type morandiTheme struct{}

func (morandiTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground:
		return morandiBg
	case theme.ColorNameButton:
		return morandiPrimary
	case theme.ColorNamePrimary:
		return morandiPrimary
	case theme.ColorNameHover:
		return morandiPrimaryHi
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 0xcc, G: 0xcc, B: 0xcc, A: 0xff}
	case theme.ColorNameForeground:
		return morandiText
	default:
		return theme.DefaultTheme().Color(n, v)
	}
}

func (morandiTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (morandiTheme) Font(style fyne.TextStyle) fyne.Resource {
	if fontData == nil {
		tryLoadFont()
	}

	if fontData == nil || style.Monospace {
		return theme.DefaultTheme().Font(style)
	}
	return fyne.NewStaticResource("custom_chinese.ttf", fontData)
}

func (morandiTheme) Size(sz fyne.ThemeSizeName) float32 {
	base := theme.DefaultTheme().Size(sz)
	switch sz {
	case theme.SizeNamePadding:
		return base + 2
	case theme.SizeNameText:
		return base + 1
	}
	return base
}

// NewMorandiTheme 返回自定义主题实例
func NewMorandiTheme() fyne.Theme { return morandiTheme{} }

var (
	morandiBg        = color.NRGBA{R: 0xf2, G: 0xf1, B: 0xee, A: 0xff}
	morandiPrimary   = color.NRGBA{R: 0xa8, G: 0xc5, B: 0xc0, A: 0xff}
	morandiPrimaryHi = color.NRGBA{R: 0x8d, G: 0xb1, B: 0xab, A: 0xff}
	morandiText      = color.NRGBA{R: 0x33, G: 0x34, B: 0x35, A: 0xff}
)
