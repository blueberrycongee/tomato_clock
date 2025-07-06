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

// MorandiColors 定义了一组莫兰迪色系的颜色
var MorandiColors = []color.Color{
	// 灰粉色系
	color.NRGBA{R: 218, G: 195, B: 194, A: 255}, // #DAC3C2
	color.NRGBA{R: 195, G: 177, B: 171, A: 255}, // #C3B1AB
	// 灰绿色系
	color.NRGBA{R: 176, G: 190, B: 181, A: 255}, // #B0BEB5
	color.NRGBA{R: 147, G: 161, B: 152, A: 255}, // #93A198
	// 灰蓝色系
	color.NRGBA{R: 163, G: 182, B: 191, A: 255}, // #A3B6BF
	color.NRGBA{R: 132, G: 153, B: 164, A: 255}, // #8499A4
	// 大地色系
	color.NRGBA{R: 202, G: 189, B: 162, A: 255}, // #CABDA2
	color.NRGBA{R: 180, G: 166, B: 143, A: 255}, // #B4A68F
}

// GetMorandiColor 根据索引获取一个莫兰迪颜色，循环使用
func GetMorandiColor(index int) color.Color {
	if len(MorandiColors) == 0 {
		return theme.PrimaryColor()
	}
	return MorandiColors[index%len(MorandiColors)]
}
