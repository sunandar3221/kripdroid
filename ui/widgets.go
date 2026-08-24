package ui

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func DrawCard(gtx layout.Context, bg color.NRGBA, radius unit.Dp, w layout.Widget) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			rr := gtx.Dp(radius)
			rect := image.Rectangle{Max: gtx.Constraints.Min}
			if bg.A > 0 && rect.Dx() > 0 && rect.Dy() > 0 {
				paint.FillShape(gtx.Ops, bg, clip.UniformRRect(rect, rr).Op(gtx.Ops))
			}
			return layout.Dimensions{Size: gtx.Constraints.Min}
		}),
		layout.Stacked(w),
	)
}

func DrawOutlinedCard(gtx layout.Context, bg, borderColor color.NRGBA, radius, borderWidth unit.Dp, w layout.Widget) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			rr := gtx.Dp(radius)
			bw := float32(gtx.Dp(borderWidth))
			if bw <= 0 {
				bw = 1
			}
			rect := image.Rectangle{Max: gtx.Constraints.Min}
			if rect.Dx() > 0 && rect.Dy() > 0 {
				if bg.A > 0 {
					paint.FillShape(gtx.Ops, bg, clip.UniformRRect(rect, rr).Op(gtx.Ops))
				}
				if borderColor.A > 0 {
					paint.FillShape(gtx.Ops, borderColor, clip.Stroke{
						Path:  clip.UniformRRect(rect, rr).Path(gtx.Ops),
						Width: bw,
					}.Op())
				}
			}
			return layout.Dimensions{Size: gtx.Constraints.Min}
		}),
		layout.Stacked(w),
	)
}

func DrawModalDialog(gtx layout.Context, th *M3Theme, content layout.Widget) layout.Dimensions {
	paint.FillShape(gtx.Ops, color.NRGBA{R: 0, G: 0, B: 0, A: 210}, clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Op())

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		maxWidth := gtx.Dp(unit.Dp(320))
		if gtx.Constraints.Max.X < maxWidth {
			maxWidth = gtx.Constraints.Max.X - gtx.Dp(unit.Dp(24))
		}
		if maxWidth < gtx.Dp(unit.Dp(200)) {
			maxWidth = gtx.Constraints.Max.X
		}
		gtx.Constraints.Min.X = maxWidth
		gtx.Constraints.Max.X = maxWidth

		return DrawOutlinedCard(gtx, th.SurfaceContainerHigh, th.OutlineVariant, unit.Dp(16), unit.Dp(1), func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(16),
				Bottom: unit.Dp(16),
				Left:   unit.Dp(16),
				Right:  unit.Dp(16),
			}.Layout(gtx, content)
		})
	})
}

func DrawPill(gtx layout.Context, bg, fg color.NRGBA, text string, textSize unit.Sp, th *M3Theme) layout.Dimensions {
	return DrawCard(gtx, bg, unit.Dp(12), func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(4),
			Bottom: unit.Dp(4),
			Left:   unit.Dp(10),
			Right:  unit.Dp(10),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th.MaterialTheme, textSize, text)
			lbl.Color = fg
			lbl.Font.Weight = font.Bold
			return lbl.Layout(gtx)
		})
	})
}

func DrawM3Chip(gtx layout.Context, th *M3Theme, btn *widget.Clickable, label string, isSelected bool) layout.Dimensions {
	var bg, fg, borderColor color.NRGBA
	if isSelected {
		bg = th.PrimaryContainer
		fg = th.OnPrimaryContainer
		borderColor = th.Primary
	} else {
		bg = th.SurfaceContainer
		fg = th.OnSurfaceVariant
		borderColor = th.OutlineVariant
	}

	return DrawOutlinedCard(gtx, bg, borderColor, unit.Dp(16), unit.Dp(1), func(gtx layout.Context) layout.Dimensions {
		return material.Clickable(gtx, btn, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(8),
				Bottom: unit.Dp(8),
				Left:   unit.Dp(12),
				Right:  unit.Dp(12),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.MaterialTheme, unit.Sp(12), label)
					lbl.Color = fg
					if isSelected {
						lbl.Font.Weight = font.Bold
					} else {
						lbl.Font.Weight = font.Medium
					}
					return lbl.Layout(gtx)
				})
			})
		})
	})
}

func DrawSecurityBadge(gtx layout.Context, th *M3Theme, category string, rating int) layout.Dimensions {
	var bg, fg color.NRGBA
	var label string

	switch {
	case rating <= 2:
		bg = th.BadgeWeakBg
		fg = th.BadgeWeakFg
		label = "DASAR / EDU"
	case rating <= 4:
		bg = th.BadgeMediumBg
		fg = th.BadgeMediumFg
		label = "MENENGAH"
	default:
		bg = th.BadgeStrongBg
		fg = th.BadgeStrongFg
		label = "STANDAR MILITER"
	}

	return DrawPill(gtx, bg, fg, label, unit.Sp(10), th)
}

func DrawM3Button(gtx layout.Context, th *M3Theme, btn *widget.Clickable, text string, bg, fg color.NRGBA, radius unit.Dp) layout.Dimensions {
	b := material.Button(th.MaterialTheme, btn, text)
	b.Background = bg
	b.Color = fg
	b.CornerRadius = radius
	b.TextSize = unit.Sp(14)
	b.Font.Weight = font.Bold
	b.Inset = layout.Inset{
		Top:    unit.Dp(14),
		Bottom: unit.Dp(14),
		Left:   unit.Dp(16),
		Right:  unit.Dp(16),
	}
	return b.Layout(gtx)
}

func DrawM3OutlinedButton(gtx layout.Context, th *M3Theme, btn *widget.Clickable, text string, borderColor, fg color.NRGBA, radius unit.Dp) layout.Dimensions {
	return DrawOutlinedCard(gtx, color.NRGBA{}, borderColor, radius, unit.Dp(1), func(gtx layout.Context) layout.Dimensions {
		b := material.Button(th.MaterialTheme, btn, text)
		b.Background = color.NRGBA{}
		b.Color = fg
		b.CornerRadius = radius
		b.TextSize = unit.Sp(13)
		b.Font.Weight = font.SemiBold
		b.Inset = layout.Inset{
			Top:    unit.Dp(10),
			Bottom: unit.Dp(10),
			Left:   unit.Dp(16),
			Right:  unit.Dp(16),
		}
		return b.Layout(gtx)
	})
}

func DrawTabItem(gtx layout.Context, th *M3Theme, btn *widget.Clickable, label string, isSelected bool) layout.Dimensions {
	var bg, fg color.NRGBA
	if isSelected {
		bg = th.PrimaryContainer
		fg = th.OnPrimaryContainer
	} else {
		bg = th.SurfaceContainerHigh
		fg = th.OnSurfaceVariant
	}

	return DrawCard(gtx, bg, unit.Dp(12), func(gtx layout.Context) layout.Dimensions {
		return material.Clickable(gtx, btn, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(10),
				Bottom: unit.Dp(10),
				Left:   unit.Dp(12),
				Right:  unit.Dp(12),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.W.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.MaterialTheme, unit.Sp(13), label)
					lbl.Color = fg
					if isSelected {
						lbl.Font.Weight = font.Bold
					} else {
						lbl.Font.Weight = font.Medium
					}
					return lbl.Layout(gtx)
				})
			})
		})
	})
}

func DrawHeader(gtx layout.Context, th *M3Theme, title, subtitle string) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th.MaterialTheme, unit.Sp(22), title)
			lbl.Color = th.Primary
			lbl.Font.Weight = font.Bold
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th.MaterialTheme, unit.Sp(13), subtitle)
			lbl.Color = th.OnSurfaceVariant
			return lbl.Layout(gtx)
		}),
	)
}

func DrawSectionTitle(gtx layout.Context, th *M3Theme, title string) layout.Dimensions {
	return layout.Inset{
		Left:   unit.Dp(2),
		Bottom: unit.Dp(4),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		lbl := material.Label(th.MaterialTheme, unit.Sp(14), title)
		lbl.Color = th.Primary
		lbl.Font.Weight = font.Bold
		return lbl.Layout(gtx)
	})
}

func DrawHamburgerIcon(gtx layout.Context, th *M3Theme, fg color.NRGBA) layout.Dimensions {
	sz := gtx.Dp(24)
	gtx.Constraints.Min = image.Point{X: sz, Y: sz}
	gtx.Constraints.Max = image.Point{X: sz, Y: sz}

	lineH := float32(gtx.Dp(2))
	if lineH < 1 {
		lineH = 1
	}
	spacing := float32(gtx.Dp(6))

	startY := float32(sz)/2 - spacing - lineH/2
	for i := 0; i < 3; i++ {
		rect := clip.Rect(image.Rect(gtx.Dp(2), int(startY), sz-gtx.Dp(2), int(startY+lineH)))
		paint.FillShape(gtx.Ops, fg, rect.Op())
		startY += spacing
	}

	return layout.Dimensions{Size: gtx.Constraints.Min}
}

func DrawAvatar(gtx layout.Context, th *M3Theme, initial string, size unit.Dp) layout.Dimensions {
	sz := gtx.Dp(size)
	gtx.Constraints.Min = image.Point{X: sz, Y: sz}
	gtx.Constraints.Max = image.Point{X: sz, Y: sz}
	
	// Draw circle
	rr := sz / 2
	rect := image.Rect(0, 0, sz, sz)
	paint.FillShape(gtx.Ops, th.Primary, clip.UniformRRect(rect, rr).Op(gtx.Ops))
	
	// Draw text
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		lbl := material.Label(th.MaterialTheme, unit.Sp(float32(size)/2), initial)
		lbl.Color = th.OnPrimary
		lbl.Font.Weight = font.Bold
		return lbl.Layout(gtx)
	})
}
