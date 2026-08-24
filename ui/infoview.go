package ui

import (
	"fmt"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"kripdroid/crypto"
)

type InfoView struct {
	List widget.List
}

func NewInfoView() *InfoView {
	iv := &InfoView{}
	iv.List.Axis = layout.Vertical
	return iv
}

func (iv *InfoView) Layout(gtx layout.Context, th *M3Theme) layout.Dimensions {
	return iv.List.Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(8),
			Bottom: unit.Dp(32),
			Left:   unit.Dp(10),
			Right:  unit.Dp(10),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return iv.layoutIntroCard(gtx, th)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return DrawSectionTitle(gtx, th, "1. TINGKAT DASAR (LATIHAN & BELAJAR)")
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return iv.layoutCategoryList(gtx, th, crypto.CategoryClassic)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return DrawSectionTitle(gtx, th, "2. TINGKAT MENENGAH (STANDAR LAMA)")
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return iv.layoutCategoryList(gtx, th, crypto.CategoryLegacy)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return DrawSectionTitle(gtx, th, "3. TINGKAT TERTINGGI (STANDAR MILITER - REKOMENDASI)")
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return iv.layoutCategoryList(gtx, th, crypto.CategoryModern)
				}),
			)
		})
	})
}

func (iv *InfoView) layoutIntroCard(gtx layout.Context, th *M3Theme) layout.Dimensions {
	return DrawCard(gtx, th.SurfaceContainerHigh, unit.Dp(14), func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(10),
			Bottom: unit.Dp(10),
			Left:   unit.Dp(12),
			Right:  unit.Dp(12),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.MaterialTheme, unit.Sp(13), "PANDUAN KEAMANAN KRIPTOGRAFI")
					lbl.Color = th.Primary
					lbl.Font.Weight = font.Bold
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.MaterialTheme, unit.Sp(11), "Untuk dokumen penting, foto, atau video rahasia, gunakan AES-256 atau ChaCha20-Poly1305 yang dilindungi Argon2id. Sandi klasik hanya untuk pembelajaran.")
					lbl.Color = th.OnSurfaceVariant
					return lbl.Layout(gtx)
				}),
			)
		})
	})
}

func (iv *InfoView) layoutCategoryList(gtx layout.Context, th *M3Theme, cat crypto.SecurityCategory) layout.Dimensions {
	var items []layout.FlexChild

	for _, algo := range crypto.AvailableAlgorithms {
		if algo.Category != cat {
			continue
		}
		item := algo
		desc := item.Description
		if item.Category == crypto.CategoryClassic {
			desc = "Sandi sederhana berbasis pergeseran karakter. Cocok untuk latihan."
		} else if item.ID == "aes256gcm" || item.ID == "chacha20poly1305" {
			desc = "Standar keamanan tertinggi dunia saat ini. Tahan dari pembobolan komputer super."
		}

		items = append(items, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Bottom: unit.Dp(6),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return DrawOutlinedCard(gtx, th.SurfaceContainer, th.OutlineVariant, unit.Dp(12), unit.Dp(1), func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Top:    unit.Dp(8),
						Bottom: unit.Dp(8),
						Left:   unit.Dp(12),
						Right:  unit.Dp(12),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							Axis: layout.Vertical,
						}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{
									Axis:      layout.Horizontal,
									Alignment: layout.Middle,
									Spacing:   layout.SpaceBetween,
								}.Layout(gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										lbl := material.Label(th.MaterialTheme, unit.Sp(12), item.Name)
										lbl.Color = th.OnSurface
										lbl.Font.Weight = font.Bold
										return lbl.Layout(gtx)
									}),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										return DrawSecurityBadge(gtx, th, string(item.Category), item.SecurityRating)
									}),
								)
							}),
							layout.Rigid(layout.Spacer{Height: unit.Dp(3)}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(th.MaterialTheme, unit.Sp(11), desc)
								lbl.Color = th.OnSurfaceVariant
								return lbl.Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Height: unit.Dp(3)}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(th.MaterialTheme, unit.Sp(11), fmt.Sprintf("Kebutuhan Kunci: %s", item.KeyDescription))
								lbl.Color = th.Primary
								return lbl.Layout(gtx)
							}),
						)
					})
				})
			})
		}))
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx, items...)
}
