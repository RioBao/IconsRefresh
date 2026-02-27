//go:build windows

package main

import (
	"image"
	"image/color"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func eventLoop(w *app.Window) error {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	s := newUIState()
	var ops op.Ops

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			select {
			case res := <-s.resultC:
				s.logLines = res.logLines
				s.footer = res.footer
				s.running = false
			default:
			}

			if !s.running {
				for i := range modes {
					if s.btns[i].Clicked(gtx) {
						preset := modes[i].preset
						s.running = true
						s.logLines = nil
						s.footer = "Running..."
						go func() {
							s.resultC <- execPreset(preset)
							w.Invalidate()
						}()
						break
					}
				}
			}

			if s.running {
				gtx = gtx.Disabled()
			}

			drawUI(gtx, th, s)
			e.Frame(gtx.Ops)
		}
	}
}

func drawUI(gtx layout.Context, th *material.Theme, s *uiState) layout.Dimensions {
	paint.Fill(gtx.Ops, colBg)
	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(drawHeader(th)),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			layout.Rigid(drawButtonRows(th, s)),
			layout.Rigid(drawModeTooltip(th, s)),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Flexed(1, drawLogCard(th, s)),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			layout.Rigid(drawFooter(th, s)),
		)
	})
}

func drawModeTooltip(th *material.Theme, s *uiState) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		desc := hoveredModeDescription(s)
		if desc == "" {
			return layout.Dimensions{}
		}
		return layout.Inset{Top: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return drawCard(gtx, func(gtx layout.Context) layout.Dimensions {
				l := material.Label(th, unit.Sp(11), desc)
				l.Color = colText2
				return l.Layout(gtx)
			})
		})
	}
}

func hoveredModeDescription(s *uiState) string {
	for i := range modes {
		if s.btns[i].Hovered() {
			return modes[i].description
		}
	}
	return ""
}

func drawHeader(th *material.Theme) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				sz := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Dp(unit.Dp(52))}
				r := gtx.Dp(unit.Dp(8))
				fillRRect(gtx.Ops, image.Rectangle{Max: sz}, r, colHeader)
				return layout.Dimensions{Size: sz}
			}),
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(52))
				gtx.Constraints.Max.Y = gtx.Dp(unit.Dp(52))
				return layout.Inset{
					Top:    unit.Dp(8),
					Bottom: unit.Dp(8),
					Left:   unit.Dp(12),
					Right:  unit.Dp(12),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							// Slight left nudge so the icon doesn't crowd the title.
							defer op.Offset(image.Pt(-gtx.Dp(unit.Dp(3)), 0)).Push(gtx.Ops).Pop()
							return drawHeaderIcon(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									l := material.Label(th, unit.Sp(18), appTitle)
									l.Font = font.Font{Weight: font.SemiBold}
									l.Color = colText1
									return l.Layout(gtx)
								}),
								layout.Rigid(layout.Spacer{Height: unit.Dp(2)}.Layout),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									l := material.Label(th, unit.Sp(12), appSubtitle)
									l.Color = colText2
									return l.Layout(gtx)
								}),
							)
						}),
					)
				})
			}),
		)
	}
}

func drawHeaderIcon(gtx layout.Context) layout.Dimensions {
	iconPx := gtx.Dp(unit.Dp(20))
	gtx.Constraints = layout.Exact(image.Pt(iconPx, iconPx))
	if !hasUIIcon {
		return layout.Dimensions{Size: image.Pt(iconPx, iconPx)}
	}
	return widget.Image{
		Src:      uiIconOp,
		Fit:      widget.Contain,
		Position: layout.Center,
	}.Layout(gtx)
}

func drawButtonRows(th *material.Theme, s *uiState) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(styledBtn(th, &s.btns[0], modes[0].label, s.running)),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(styledBtn(th, &s.btns[1], modes[1].label, s.running)),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(styledBtn(th, &s.btns[2], modes[2].label, s.running)),
		)
	}
}

func styledBtn(th *material.Theme, btn *widget.Clickable, label string, disabled bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		h := gtx.Dp(unit.Dp(34))
		gtx.Constraints.Min.Y = h
		gtx.Constraints.Max.Y = h

		return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			sz := image.Point{X: gtx.Constraints.Max.X, Y: h}
			r := gtx.Dp(unit.Dp(8))
			rect := image.Rectangle{Max: sz}

			var bg color.NRGBA
			switch {
			case disabled:
				bg = colDisabledBg
			case btn.Pressed():
				bg = colBtnPress
			case btn.Hovered():
				bg = colBtnHover
			default:
				bg = colBtnBg
			}

			fillRRect(gtx.Ops, rect, r, colBtnBorder)
			ir := r - 1
			if ir < 0 {
				ir = 0
			}
			fillRRect(gtx.Ops, image.Rectangle{Min: image.Pt(1, 1), Max: image.Pt(sz.X-1, sz.Y-1)}, ir, bg)

			textCol := colBtnFg
			if disabled {
				textCol = colDisabledText
			}
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				l := material.Label(th, unit.Sp(13), label)
				l.Font = font.Font{Weight: font.Medium}
				l.Color = textCol
				return l.Layout(gtx)
			})
		})
	}
}

func drawLogCard(th *material.Theme, s *uiState) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return drawCard(gtx, func(gtx layout.Context) layout.Dimensions {
			if len(s.logLines) == 0 {
				l := material.Label(th, unit.Sp(11), "No output yet.")
				l.Color = colText3
				l.Font = font.Font{Typeface: "Go Mono"}
				return l.Layout(gtx)
			}
			return s.logList.Layout(gtx, len(s.logLines), func(gtx layout.Context, i int) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					l := material.Label(th, unit.Sp(11), s.logLines[i])
					l.Color = logLineColor(s.logLines[i])
					l.Font = font.Font{Typeface: "Go Mono"}
					return l.Layout(gtx)
				})
			})
		})
	}
}

func drawCard(gtx layout.Context, content layout.Widget) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			sz := gtx.Constraints.Min
			r := gtx.Dp(unit.Dp(8))
			fillRRect(gtx.Ops, image.Rectangle{Max: sz}, r, colBorder)
			ir := r - 1
			if ir < 0 {
				ir = 0
			}
			inner := image.Rectangle{Min: image.Pt(1, 1), Max: image.Pt(sz.X-1, sz.Y-1)}
			fillRRect(gtx.Ops, inner, ir, colSurface)
			return layout.Dimensions{Size: sz}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(10)).Layout(gtx, content)
		}),
	)
}

func drawFooter(th *material.Theme, s *uiState) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		txt := s.footer
		if txt == "" {
			txt = "Ready"
		}
		l := material.Label(th, unit.Sp(12), txt)
		l.Color = colText2
		return l.Layout(gtx)
	}
}
