package cli

import (
	"image"
	"image/color"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/npillmayer/pmmp/pmmp/ui/gui"
)

func NewLineView(scaled float32) (gui.WidgetView, []app.Option, error) {
	th := material.NewTheme(gofont.Collection())
	sizingOption := app.Size(unit.Dp(250), unit.Dp(250))
	grey := color.NRGBA{R: 90, G: 90, B: 90, A: 255}
	title1 := "Font: q"
	return gui.WidgetView(func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(material.Body1(th, title1).Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				n := float32(gtx.Px(unit.Dp(200)))
				drawLine(gtx, f32.Pt(0, 0), f32.Pt(n, n), grey)
				a := gtx.Px(unit.Sp(12))
				tracer().Debugf("a = %d, n = %f", a, n)
				// _d/_em = gtx.Px(DPI) * (72.27 / PT)
				tracer().Infof("ppem(12pt) = %f", float32(gtx.Px(unit.Dp(72)))*(12.0/72.27))
				return layout.Dimensions{Size: image.Pt(250, 250)}
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				//op.Offset(f32.Pt(200, 150)).Add(gtx.Ops)
				op.Offset(f32.Pt(20, scaled+45)).Add(gtx.Ops) // TODO layout.Rigid(...)
				//tracer().Infof("op = %#v", opText)
				d := image.Point{Y: 400}
				return layout.Dimensions{Size: d}
			}),
		)
	}), []app.Option{sizingOption}, nil
}

func drawLine(gtx layout.Context, p1, p2 f32.Point, c color.NRGBA) {
	var linePath clip.Path
	//op.Offset(f32.Pt(0, -d)).Add(gtx.Ops)
	linePath.Begin(gtx.Ops)
	linePath.MoveTo(p1)
	linePath.LineTo(p2)
	pspec := linePath.End()
	stroke := clip.Stroke{
		Path:  pspec,
		Width: 2,
	}
	line := clip.Stroke(stroke).Op()
	paint.FillShape(gtx.Ops, c, line)
}

func drawBox(gtx layout.Context, box image.Rectangle, c color.NRGBA) {
	p1 := f32.Pt(float32(box.Min.X), -float32(box.Min.Y))
	p2 := f32.Pt(float32(box.Max.X), -float32(box.Max.Y))
	drawRect(gtx, p1, p2, c)
}

func drawRect(gtx layout.Context, p1, p2 f32.Point, c color.NRGBA) {
	var linePath clip.Path
	//op.Offset(f32.Pt(0, -d)).Add(gtx.Ops)
	linePath.Begin(gtx.Ops)
	linePath.MoveTo(p1)
	linePath.LineTo(f32.Pt(p1.X, p2.Y))
	linePath.LineTo(p2)
	linePath.LineTo(f32.Pt(p2.X, p1.Y))
	linePath.Close()
	pspec := linePath.End()
	stroke := clip.Stroke{
		Path:  pspec,
		Width: 2,
	}
	line := clip.Stroke(stroke).Op()
	paint.FillShape(gtx.Ops, c, line)
}
