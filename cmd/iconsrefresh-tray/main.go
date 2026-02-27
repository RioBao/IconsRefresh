//go:build windows

package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/crazy-max/IconsRefresh/internal/engine"
)

//go:embed icon.png
var uiIconPNG []byte

var (
	uiIconOp paint.ImageOp
	hasUIcon bool
)

// ── Palette ───────────────────────────────────────────────────────────────
//
// One muted accent (Nord slate blue #5E81AC) across an otherwise neutral
// chrome. Windows 11 uses #F3F3F3 for its app background — we match that
// so the window feels native without being a copy.
var (
	// Surfaces
	colBg      = color.NRGBA{R: 0xF3, G: 0xF3, B: 0xF3, A: 0xFF}
	colSurface = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	colBorder  = color.NRGBA{R: 0xE5, G: 0xE5, B: 0xE5, A: 0xFF}
	colHeader  = color.NRGBA{R: 0xF5, G: 0xF7, B: 0xFA, A: 0xFF}

	// Buttons — uniform neutral style
	colBtnBg     = color.NRGBA{R: 233, G: 238, B: 244, A: 255}
	colBtnFg     = color.NRGBA{R: 32, G: 34, B: 38, A: 255}
	colBtnHover  = color.NRGBA{R: 220, G: 227, B: 235, A: 255}
	colBtnPress  = color.NRGBA{R: 210, G: 218, B: 228, A: 255}
	colBtnBorder = color.NRGBA{R: 200, G: 205, B: 212, A: 255}

	// Disabled
	colDisabledBg   = color.NRGBA{R: 0xE8, G: 0xE8, B: 0xE8, A: 0xFF}
	colDisabledText = color.NRGBA{R: 0xAA, G: 0xAA, B: 0xAA, A: 0xFF}

	// Text hierarchy — three steps is enough
	colText1 = color.NRGBA{R: 0x1C, G: 0x1C, B: 0x1C, A: 0xFF}
	colText2 = color.NRGBA{R: 0x5F, G: 0x5F, B: 0x5F, A: 0xFF}
	colText3 = color.NRGBA{R: 0x9E, G: 0x9E, B: 0x9E, A: 0xFF}

	// Log line semantics — muted so colour codes, not distracts
	colSuccess = color.NRGBA{R: 0x2E, G: 0x7D, B: 0x47, A: 0xFF}
	colFail    = color.NRGBA{R: 0xB0, G: 0x30, B: 0x30, A: 0xFF}
	colWarn    = color.NRGBA{R: 0xB0, G: 0x6A, B: 0x00, A: 0xFF}
	colStep    = color.NRGBA{R: 0x60, G: 0x74, B: 0x88, A: 0xFF}
)

// ── Modes ─────────────────────────────────────────────────────────────────

var modes = [4]struct {
	label  string
	preset engine.Preset
}{
	{"Quick", engine.PresetTrayQuick},
	{"Standard", engine.PresetTrayStandard},
	{"Soft", engine.PresetTraySoft},
	{"Deep", engine.PresetTrayDeep},
}

// ── State ─────────────────────────────────────────────────────────────────

type runResult struct {
	logLines []string
	footer   string
}

type uiState struct {
	btns     [4]widget.Clickable
	running  bool
	logLines []string
	footer   string
	logList  widget.List
	resultC  chan runResult
}

func init() {
	if len(uiIconPNG) == 0 {
		return
	}
	img, err := png.Decode(bytes.NewReader(uiIconPNG))
	if err != nil {
		return
	}
	uiIconOp = paint.NewImageOp(img)
	hasUIcon = true
}

// ── Entry point ───────────────────────────────────────────────────────────

func main() {
	go func() {
		w := new(app.Window)
		w.Option(app.Title("IconsRefresh"), app.Size(unit.Dp(320), unit.Dp(400)))
		if err := eventLoop(w); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func eventLoop(w *app.Window) error {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	s := &uiState{
		logList: widget.List{List: layout.List{Axis: layout.Vertical}},
		resultC: make(chan runResult, 1),
	}
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
						s.footer = "Running\u2026"
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

// ── Layout ────────────────────────────────────────────────────────────────

func drawUI(gtx layout.Context, th *material.Theme, s *uiState) layout.Dimensions {
	paint.Fill(gtx.Ops, colBg)
	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(drawHeader(th)),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			layout.Rigid(drawButtonGrid(th, s)),
			layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout),
			layout.Flexed(1, drawLogCard(th, s)),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			layout.Rigid(drawFooter(th, s)),
		)
	})
}

// drawHeader — SemiBold (not Bold) at 18sp: present without looming.
// Subtitle at 12sp colText2 establishes the hierarchy step below the title.
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
							// Nudge icon slightly left so it doesn't crowd the title.
							defer op.Offset(image.Pt(-gtx.Dp(unit.Dp(3)), 0)).Push(gtx.Ops).Pop()
							return drawHeaderIcon(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									l := material.Label(th, unit.Sp(18), "IconsRefresh")
									l.Font = font.Font{Weight: font.SemiBold}
									l.Color = colText1
									return l.Layout(gtx)
								}),
								layout.Rigid(layout.Spacer{Height: unit.Dp(2)}.Layout),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									l := material.Label(th, unit.Sp(12), "Refresh Windows icon caches")
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
	if !hasUIcon {
		return layout.Dimensions{Size: image.Pt(iconPx, iconPx)}
	}
	return widget.Image{
		Src:      uiIconOp,
		Fit:      widget.Contain,
		Position: layout.Center,
	}.Layout(gtx)
}

// drawButtonGrid — 2×2 with 8 dp gaps, matching outer padding rhythm.
func drawButtonGrid(th *material.Theme, s *uiState) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return btnRow(gtx, th, s, 0, 1)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return btnRow(gtx, th, s, 2, 3)
			}),
		)
	}
}

func btnRow(gtx layout.Context, th *material.Theme, s *uiState, a, b int) layout.Dimensions {
	return layout.Flex{}.Layout(gtx,
		layout.Flexed(1, styledBtn(th, &s.btns[a], modes[a].label, s.running)),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Flexed(1, styledBtn(th, &s.btns[b], modes[b].label, s.running)),
	)
}

// styledBtn draws a neutral button with a 1 px border and hover/press states.
// Height is fixed at 34 dp; text stays vertically centred via layout.Center.
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

			// Border: outer rounded rect filled with border colour,
			// then 1 px-inset interior filled with the button background.
			fillRRect(gtx.Ops, rect, r, colBtnBorder)
			ir := r - 1
			if ir < 0 {
				ir = 0
			}
			fillRRect(gtx.Ops, image.Rectangle{
				Min: image.Pt(1, 1),
				Max: image.Pt(sz.X-1, sz.Y-1),
			}, ir, bg)

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

// drawLogCard wraps the activity log in a white rounded-corner surface.
// Monospace 11 sp keeps lines dense and readable; semantic colours make
// success/failure scannable without reading every word.
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

// logLineColor maps the leading symbol to a semantic colour.
func logLineColor(line string) color.NRGBA {
	switch {
	case strings.HasPrefix(line, "✓"):
		return colSuccess
	case strings.HasPrefix(line, "✗"):
		return colFail
	case strings.HasPrefix(line, "⚠"):
		return colWarn
	case strings.HasPrefix(line, "—"):
		return colStep
	default:
		return colText2
	}
}

// drawCard renders a white rounded surface with a 1 px border.
//
// layout.Stack is used so the Expanded background fills exactly the area
// the Stacked content needs — no hardcoded heights required.
func drawCard(gtx layout.Context, content layout.Widget) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			sz := gtx.Constraints.Min
			r := gtx.Dp(unit.Dp(8))
			// Border: fill the full rounded rect with border colour,
			// then overwrite the 1 px-inset interior with the surface colour.
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

// drawFooter shows the post-run summary ("15 deleted • 1 skipped") or "Ready".
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

// ── Draw helpers ──────────────────────────────────────────────────────────

// fillRRect fills a rounded rectangle with a solid colour.
// The defer Pop() runs when fillRRect returns, after paint.Fill has been
// recorded — giving correct push→fill→pop op ordering.
func fillRRect(ops *op.Ops, rect image.Rectangle, r int, col color.NRGBA) {
	defer clip.RRect{Rect: rect, SE: r, SW: r, NW: r, NE: r}.Push(ops).Pop()
	paint.Fill(ops, col)
}

// ── Engine ────────────────────────────────────────────────────────────────

func execPreset(preset engine.Preset) runResult {
	eng := engine.New(engine.Hooks{})
	req, err := engine.RequestForPreset(preset, false)
	if err != nil {
		return runResult{logLines: []string{"error: " + err.Error()}, footer: "Error"}
	}
	r, err := eng.Run(context.Background(), req)
	return buildResult(r, err)
}

func buildResult(r engine.RunResult, err error) runResult {
	if err != nil {
		return runResult{logLines: []string{"error: " + err.Error()}, footer: "Error"}
	}

	var lines []string
	var deleted, skipped, failed int

	for _, p := range r.Result.Paths {
		switch {
		case p.Deleted:
			deleted++
			lines = append(lines, "✓  "+filepath.Base(p.Path))
		case p.Skipped:
			skipped++
		case p.Error != "":
			failed++
			lines = append(lines, "✗  "+filepath.Base(p.Path)+": "+p.Error)
		}
	}

	if r.Result.ExplorerRestart != nil && r.Result.ExplorerRestart.Restarted {
		lines = append(lines, "—  Explorer restarted")
	}
	if r.Result.IE4UInit != nil && r.Result.IE4UInit.Ran {
		lines = append(lines, "—  ie4uinit ran")
	}
	if r.Result.ShellNotify != nil && r.Result.ShellNotify.AssocChangedSent {
		lines = append(lines, "—  Shell notified")
	}
	for _, w := range collectWarnings(r) {
		lines = append(lines, "⚠  "+w)
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("%d deleted", deleted))
	if skipped > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped", skipped))
	}
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", failed))
	}
	return runResult{logLines: lines, footer: strings.Join(parts, " • ")}
}

func collectWarnings(r engine.RunResult) []string {
	var ws []string
	if r.Result.ExplorerRestart != nil && r.Result.ExplorerRestart.Warning != "" {
		ws = append(ws, r.Result.ExplorerRestart.Warning)
	}
	if r.Result.ShellNotify != nil && r.Result.ShellNotify.Warning != "" {
		ws = append(ws, r.Result.ShellNotify.Warning)
	}
	if r.Result.IE4UInit != nil && r.Result.IE4UInit.Warning != "" {
		ws = append(ws, r.Result.IE4UInit.Warning)
	}
	return ws
}
