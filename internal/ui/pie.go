package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/fogleman/gg"
)

const pieMinSize = 100 // 最小饼图尺寸，可按需调整

// PieChartSegment 定义了饼图的单个部分
type PieChartSegment struct {
	Label string
	Value float64
}

// PieChart 是一个自定义的饼图小部件
type PieChart struct {
	widget.BaseWidget
	segments    []PieChartSegment
	colors      []color.Color
	title       string
	titleLabel  *widget.Label
	legend      *fyne.Container
	chartRaster *canvas.Raster
	onHovered   func(segment *PieChartSegment)
	onUnhovered func()
}

// NewPieChart 创建一个新的饼图实例
func NewPieChart(title string, segments []PieChartSegment) *PieChart {
	pc := &PieChart{
		segments: segments,
		colors:   MorandiColors, // 默认使用莫兰迪色系
		title:    title,
	}
	pc.ExtendBaseWidget(pc) // 非常重要
	pc.titleLabel = widget.NewLabel(pc.title)
	pc.titleLabel.Alignment = fyne.TextAlignCenter
	pc.titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	pc.chartRaster = canvas.NewRaster(pc.drawPie)
	pc.legend = container.NewVBox()
	pc.UpdateData(segments)
	return pc
}

// UpdateData 更新饼图的数据
func (pc *PieChart) UpdateData(segments []PieChartSegment) {
	pc.segments = segments
	// 按值降序排序
	sort.Slice(pc.segments, func(i, j int) bool {
		return pc.segments[i].Value > pc.segments[j].Value
	})
	pc.chartRaster.Refresh()
	pc.updateLegend()
	pc.Refresh()
}

// drawPie 是光栅化绘图函数
func (pc *PieChart) drawPie(w, h int) image.Image {
	dc := gg.NewContext(w, h)
	dc.SetRGB(1, 1, 1) // 设置背景色为白色
	dc.Clear()

	totalValue := 0.0
	for _, s := range pc.segments {
		totalValue += s.Value
	}

	if totalValue == 0 {
		// 画一个灰色的圆表示没有数据
		dc.SetColor(color.NRGBA{R: 230, G: 230, B: 230, A: 255})
		dc.DrawCircle(float64(w)/2, float64(h)/2, float64(w)/2.2)
		dc.Fill()
		return dc.Image()
	}

	x, y := float64(w)/2, float64(h)/2
	radius := math.Min(x, y) / 1.1

	currentAngle := -math.Pi / 2 // 从12点钟方向开始

	for i, s := range pc.segments {
		sliceAngle := (s.Value / totalValue) * 2 * math.Pi
		// 设置切片颜色
		dc.SetColor(GetMorandiColor(i))

		// 绘制扇形
		dc.MoveTo(x, y)
		dc.LineTo(x+radius*math.Cos(currentAngle), y+radius*math.Sin(currentAngle))
		dc.DrawArc(x, y, radius, currentAngle, currentAngle+sliceAngle)
		dc.LineTo(x, y)
		dc.Fill()

		currentAngle += sliceAngle
	}

	return dc.Image()
}

// updateLegend 更新图例
func (pc *PieChart) updateLegend() {
	pc.legend.Objects = nil
	totalValue := 0.0
	for _, s := range pc.segments {
		totalValue += s.Value
	}

	if totalValue == 0 {
		noDataLabel := widget.NewLabel("暂无数据")
		noDataLabel.Alignment = fyne.TextAlignCenter
		pc.legend.Add(noDataLabel)
		pc.legend.Refresh()
		return
	}

	for i, s := range pc.segments {
		percentage := (s.Value / totalValue) * 100
		colorBox := canvas.NewRectangle(GetMorandiColor(i))
		colorBox.SetMinSize(fyne.NewSize(14, 14))

		// 限制标签长度
		labelStr := s.Label
		if len(labelStr) > 10 {
			labelStr = labelStr[:10] + "..."
		}

		legendLabel := widget.NewLabel(fmt.Sprintf("%s: %.1f%%", labelStr, percentage))
		legendRow := container.NewHBox(colorBox, legendLabel)
		pc.legend.Add(legendRow)
	}
	pc.legend.Refresh()
}

// CreateRenderer 创建此小部件的渲染器
func (pc *PieChart) CreateRenderer() fyne.WidgetRenderer {
	pc.chartRaster.SetMinSize(fyne.NewSize(pieMinSize, pieMinSize))
	layout := container.NewBorder(pc.titleLabel, pc.legend, nil, nil, pc.chartRaster)
	return &pieChartRenderer{
		pieChart: pc,
		layout:   layout,
		objects:  []fyne.CanvasObject{layout},
	}
}

// pieChartRenderer 是PieChart的渲染器实现
type pieChartRenderer struct {
	pieChart *PieChart
	layout   *fyne.Container
	objects  []fyne.CanvasObject
}

func (r *pieChartRenderer) Layout(size fyne.Size) {
	r.layout.Resize(size)
}

func (r *pieChartRenderer) MinSize() fyne.Size {
	return r.layout.MinSize()
}

func (r *pieChartRenderer) Refresh() {
	r.pieChart.titleLabel.SetText(r.pieChart.title)
	r.pieChart.updateLegend()
	r.pieChart.chartRaster.Refresh()
	canvas.Refresh(r.pieChart)
}

func (r *pieChartRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *pieChartRenderer) Destroy() {}
