package ui

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	_ "image/png"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/explorer"
)

//go:embed appicon.png
var appIconBytes []byte

type App struct {
	Theme      *M3Theme
	CurrentTab int
	TabButtons [3]widget.Clickable
	TextView   *TextView
	FileView   *FileView
	InfoView   *InfoView
	Window     *app.Window
	Explorer   *explorer.Explorer
	
	MenuBtn      widget.Clickable
	CloseMenuBtn widget.Clickable
	DrawerOpen   bool
	AppIcon      widget.Image
}

func NewApp(window *app.Window, expl *explorer.Explorer) *App {
	theme := NewM3Theme()
	var invalidate func()
	if window != nil {
		invalidate = window.Invalidate
	}

	var appIconWidget widget.Image
	if img, _, err := image.Decode(bytes.NewReader(appIconBytes)); err == nil {
		appIconWidget = widget.Image{Src: paint.NewImageOp(img), Fit: widget.Contain}
	}

	return &App{
		Theme:      theme,
		CurrentTab: 0,
		InfoView:   NewInfoView(),
		TextView:   NewTextView(invalidate),
		FileView:   NewFileView(expl, invalidate),
		Window:     window,
		Explorer:   expl,
		AppIcon:    appIconWidget,
	}
}

func (a *App) Layout(gtx layout.Context) layout.Dimensions {
	if a.MenuBtn.Clicked(gtx) {
		a.DrawerOpen = !a.DrawerOpen
	}
	if a.CloseMenuBtn.Clicked(gtx) {
		a.DrawerOpen = false
	}
	for i := range a.TabButtons {
		if a.TabButtons[i].Clicked(gtx) {
			a.CurrentTab = i
			a.DrawerOpen = false
		}
	}

	paint.FillShape(gtx.Ops, a.Theme.Background, clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Op())

	content := func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutTopAppBar(gtx)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				switch a.CurrentTab {
				case 0:
					return a.InfoView.Layout(gtx, a.Theme)
				case 1:
					return a.TextView.Layout(gtx, a.Theme)
				case 2:
					return a.FileView.Layout(gtx, a.Theme)
				default:
					return a.InfoView.Layout(gtx, a.Theme)
				}
			}),
		)
	}

	if !a.DrawerOpen {
		return content(gtx)
	}

	return layout.Stack{}.Layout(gtx,
		layout.Stacked(content),
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			// Dim background
			paint.FillShape(gtx.Ops, color.NRGBA{0, 0, 0, 150}, clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Op())
			// Close click area
			material.Clickable(gtx, &a.CloseMenuBtn, func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: gtx.Constraints.Max}
			})
			return layout.W.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return a.layoutDrawer(gtx)
			})
		}),
	)
}

func (a *App) layoutTopAppBar(gtx layout.Context) layout.Dimensions {
	return DrawCard(gtx, a.Theme.SurfaceContainerLow, unit.Dp(0), func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(12),
			Bottom: unit.Dp(12),
			Left:   unit.Dp(16),
			Right:  unit.Dp(16),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
				Spacing:   layout.SpaceBetween,
			}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:      layout.Horizontal,
						Alignment: layout.Middle,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return material.Clickable(gtx, &a.MenuBtn, func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{Right: unit.Dp(16), Top: unit.Dp(4), Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return DrawHamburgerIcon(gtx, a.Theme, a.Theme.OnBackground)
								})
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Right: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								sz := gtx.Dp(unit.Dp(32))
								gtx.Constraints.Min = image.Point{X: sz, Y: sz}
								gtx.Constraints.Max = image.Point{X: sz, Y: sz}
								defer clip.UniformRRect(image.Rect(0, 0, sz, sz), sz/4).Push(gtx.Ops).Pop()
								return a.AppIcon.Layout(gtx)
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(a.Theme.MaterialTheme, unit.Sp(20), "KripDroid")
							lbl.Color = a.Theme.Primary
							lbl.Font.Weight = font.Bold
							return lbl.Layout(gtx)
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return DrawAvatar(gtx, a.Theme, "U", unit.Dp(36))
				}),
			)
		})
	})
}

func (a *App) layoutDrawer(gtx layout.Context) layout.Dimensions {
	w := gtx.Dp(unit.Dp(280))
	if w > gtx.Constraints.Max.X-gtx.Dp(unit.Dp(56)) {
		w = gtx.Constraints.Max.X - gtx.Dp(unit.Dp(56))
	}
	
	h := gtx.Constraints.Max.Y

	// Draw the drawer background explicitly
	rect := image.Rect(0, 0, w, h)
	paint.FillShape(gtx.Ops, a.Theme.Surface, clip.Rect(rect).Op())

	// Constrain the content width
	gtx.Constraints.Min.X = w
	gtx.Constraints.Max.X = w
	gtx.Constraints.Min.Y = h
	gtx.Constraints.Max.Y = h

	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Point{X: w, Y: h}}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Top: unit.Dp(32), Bottom: unit.Dp(24), Left: unit.Dp(24), Right: unit.Dp(24)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return DrawAvatar(gtx, a.Theme, "U", unit.Dp(64))
							}),
							layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.Theme.MaterialTheme, unit.Sp(16), "Pengguna KripDroid")
								lbl.Color = a.Theme.OnSurface
								lbl.Font.Weight = font.Bold
								return lbl.Layout(gtx)
							}),
						)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					tabTitles := []string{"Menu Utama", "Kunci Teks", "Kunci File"}
					var children []layout.FlexChild
					for i, title := range tabTitles {
						idx := i
						tabTitle := title
						children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Top: unit.Dp(4), Bottom: unit.Dp(4), Left: unit.Dp(12), Right: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return DrawTabItem(gtx, a.Theme, &a.TabButtons[idx], tabTitle, a.CurrentTab == idx)
							})
						}))
					}
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
				}),
			)
		}),
	)
}


func Run(w *app.Window) error {
	expl := explorer.NewExplorer(w)
	uiApp := NewApp(w, expl)
	var ops op.Ops

	for {
		e := w.Event()
		expl.ListenEvents(e)

		switch evt := e.(type) {
		case app.DestroyEvent:
			return evt.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, evt)
			uiApp.Layout(gtx)
			evt.Frame(gtx.Ops)
		}
	}
}
