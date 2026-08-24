package ui

import (
	"image/color"

	"gioui.org/unit"
	"gioui.org/widget/material"
)

type M3Theme struct {
	MaterialTheme           *material.Theme
	Primary                 color.NRGBA
	OnPrimary               color.NRGBA
	PrimaryContainer        color.NRGBA
	OnPrimaryContainer      color.NRGBA
	Secondary               color.NRGBA
	OnSecondary             color.NRGBA
	SecondaryContainer      color.NRGBA
	OnSecondaryContainer    color.NRGBA
	Tertiary                color.NRGBA
	OnTertiary              color.NRGBA
	TertiaryContainer       color.NRGBA
	OnTertiaryContainer     color.NRGBA
	Background              color.NRGBA
	OnBackground            color.NRGBA
	Surface                 color.NRGBA
	OnSurface               color.NRGBA
	SurfaceVariant          color.NRGBA
	OnSurfaceVariant        color.NRGBA
	SurfaceContainerLowest  color.NRGBA
	SurfaceContainerLow     color.NRGBA
	SurfaceContainer        color.NRGBA
	SurfaceContainerHigh    color.NRGBA
	SurfaceContainerHighest color.NRGBA
	Outline                 color.NRGBA
	OutlineVariant          color.NRGBA
	Error                   color.NRGBA
	OnError                 color.NRGBA
	ErrorContainer          color.NRGBA
	OnErrorContainer        color.NRGBA
	Success                 color.NRGBA
	OnSuccess               color.NRGBA
	Warning                 color.NRGBA
	BadgeWeakBg             color.NRGBA
	BadgeWeakFg             color.NRGBA
	BadgeMediumBg           color.NRGBA
	BadgeMediumFg           color.NRGBA
	BadgeStrongBg           color.NRGBA
	BadgeStrongFg           color.NRGBA
}

func HexColor(hex uint32) color.NRGBA {
	return color.NRGBA{
		R: byte(hex >> 24),
		G: byte(hex >> 16),
		B: byte(hex >> 8),
		A: byte(hex),
	}
}

func NewM3Theme() *M3Theme {
	th := material.NewTheme()

	m3 := &M3Theme{
		MaterialTheme:           th,
		Primary:                 HexColor(0x8AB4F8FF),
		OnPrimary:               HexColor(0x002B75FF),
		PrimaryContainer:        HexColor(0x2B4A8EFF),
		OnPrimaryContainer:      HexColor(0xD3E3FDFF),
		Secondary:               HexColor(0x81C995FF),
		OnSecondary:             HexColor(0x00391AFF),
		SecondaryContainer:      HexColor(0x1E5232FF),
		OnSecondaryContainer:    HexColor(0xCEEAD6FF),
		Tertiary:                HexColor(0xFDD663FF),
		OnTertiary:              HexColor(0x402C05FF),
		TertiaryContainer:       HexColor(0x5E4200FF),
		OnTertiaryContainer:     HexColor(0xFFE082FF),
		Background:              HexColor(0x101418FF),
		OnBackground:            HexColor(0xE3E3E3FF),
		Surface:                 HexColor(0x101418FF),
		OnSurface:               HexColor(0xE3E3E3FF),
		SurfaceVariant:          HexColor(0x181C20FF),
		OnSurfaceVariant:        HexColor(0xC4C7C5FF),
		SurfaceContainerLowest:  HexColor(0x0B0E11FF),
		SurfaceContainerLow:     HexColor(0x181C20FF),
		SurfaceContainer:        HexColor(0x1E2226FF),
		SurfaceContainerHigh:    HexColor(0x282C31FF),
		SurfaceContainerHighest: HexColor(0x33373CFF),
		Outline:                 HexColor(0x8E918FFF),
		OutlineVariant:          HexColor(0x444746FF),
		Error:                   HexColor(0xF28B82FF),
		OnError:                 HexColor(0x601410FF),
		ErrorContainer:          HexColor(0x8C1D18FF),
		OnErrorContainer:        HexColor(0xFAD2CFFF),
		Success:                 HexColor(0x81C995FF),
		OnSuccess:               HexColor(0x00391AFF),
		Warning:                 HexColor(0xFDD663FF),
		BadgeWeakBg:             HexColor(0x4C1A17FF),
		BadgeWeakFg:             HexColor(0xF6AEA9FF),
		BadgeMediumBg:           HexColor(0x4F3000FF),
		BadgeMediumFg:           HexColor(0xFDD663FF),
		BadgeStrongBg:           HexColor(0x0E381CFF),
		BadgeStrongFg:           HexColor(0x81C995FF),
	}

	th.Palette.Bg = m3.Background
	th.Palette.Fg = m3.OnBackground
	th.Palette.ContrastBg = m3.Primary
	th.Palette.ContrastFg = m3.OnPrimary
	th.TextSize = unit.Sp(14)
	th.FingerSize = unit.Dp(48)

	return m3
}
