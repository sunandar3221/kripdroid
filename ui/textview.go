package ui

import (
	"fmt"
	"image/color"
	"io"
	"strings"

	"gioui.org/font"
	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"kripdroid/crypto"
)

type TextView struct {
	InputEditor              widget.Editor
	KeyEditor                widget.Editor
	IvEditor                 widget.Editor
	OutputEditor             widget.Editor
	MasterPassEditor         widget.Editor
	ConfirmMasterPassEditor  widget.Editor
	OldMasterPassEditor      widget.Editor
	NewMasterPassEditor      widget.Editor
	SelectedAlgoID           string
	IsKeyMasked              bool
	ShowAdvancedIV           bool
	UseDeviceKey             bool
	ShowBiometricPrompt      bool
	ShowMasterPassModal      bool
	ShowChangeMasterPass     bool
	PendingAction            string
	Invalidate               func()
	EncryptBtn               widget.Clickable
	DecryptBtn               widget.Clickable
	CopyBtn                  widget.Clickable
	ClearBtn                 widget.Clickable
	SwapBtn                  widget.Clickable
	ToggleMaskBtn            widget.Clickable
	ToggleIvBtn              widget.Clickable
	SampleBtn                widget.Clickable
	KeyStoreBtn              widget.Clickable
	SwitchToCustomPassBtn    widget.Clickable
	ConfirmBiometricBtn      widget.Clickable
	CancelBiometricBtn       widget.Clickable
	OpenMasterPassBtn        widget.Clickable
	SubmitMasterPassBtn      widget.Clickable
	CancelMasterPassBtn      widget.Clickable
	OpenChangeMasterPassBtn  widget.Clickable
	SubmitChangeMasterPassBt widget.Clickable
	CancelChangeMasterPassBt widget.Clickable
	AlgoButtons              []widget.Clickable
	List                     widget.List
	StatusMessage            string
	IsError                  bool
}

func NewTextView(invalidate func()) *TextView {
	tv := &TextView{
		Invalidate:     invalidate,
		SelectedAlgoID: "aes256gcm",
		IsKeyMasked:    false,
		ShowAdvancedIV: false,
		UseDeviceKey:   false,
		AlgoButtons:    make([]widget.Clickable, len(crypto.AvailableAlgorithms)),
	}

	tv.List.Axis = layout.Vertical
	tv.InputEditor.WrapPolicy = text.WrapGraphemes
	tv.OutputEditor.WrapPolicy = text.WrapGraphemes
	tv.OutputEditor.ReadOnly = true
	tv.KeyEditor.SingleLine = true
	tv.IvEditor.SingleLine = true
	tv.MasterPassEditor.SingleLine = true
	tv.MasterPassEditor.Mask = '•'
	tv.ConfirmMasterPassEditor.SingleLine = true
	tv.ConfirmMasterPassEditor.Mask = '•'
	tv.OldMasterPassEditor.SingleLine = true
	tv.OldMasterPassEditor.Mask = '•'
	tv.NewMasterPassEditor.SingleLine = true
	tv.NewMasterPassEditor.Mask = '•'

	return tv
}

func (tv *TextView) Layout(gtx layout.Context, th *M3Theme) layout.Dimensions {
	tv.handleEvents(gtx)

	algo, _ := crypto.GetAlgorithm(tv.SelectedAlgoID)

	var items []layout.Widget

	items = append(items, func(gtx layout.Context) layout.Dimensions {
		return tv.layoutAlgorithmSelector(gtx, th)
	})

	if algo != nil {
		items = append(items, func(gtx layout.Context) layout.Dimensions {
			return tv.layoutAlgorithmInfoCard(gtx, th, algo)
		})
	}

	items = append(items, func(gtx layout.Context) layout.Dimensions {
		return tv.layoutInputSection(gtx, th, algo)
	})

	items = append(items, func(gtx layout.Context) layout.Dimensions {
		return tv.layoutKeySection(gtx, th, algo)
	})

	items = append(items, func(gtx layout.Context) layout.Dimensions {
		return tv.layoutActionButtons(gtx, th)
	})

	items = append(items, func(gtx layout.Context) layout.Dimensions {
		return tv.layoutOutputSection(gtx, th)
	})

	if tv.StatusMessage != "" {
		items = append(items, func(gtx layout.Context) layout.Dimensions {
			return tv.layoutStatusBanner(gtx, th)
		})
	}

	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return tv.List.Layout(gtx, len(items), func(gtx layout.Context, index int) layout.Dimensions {
				return layout.Inset{
					Top:    unit.Dp(4),
					Bottom: unit.Dp(4),
					Left:   unit.Dp(12),
					Right:  unit.Dp(12),
				}.Layout(gtx, items[index])
			})
		}),
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			if tv.ShowMasterPassModal {
				return DrawModalDialog(gtx, th, func(gtx layout.Context) layout.Dimensions {
					return tv.layoutMasterPassModalContent(gtx, th)
				})
			}
			if tv.ShowBiometricPrompt {
				return DrawModalDialog(gtx, th, func(gtx layout.Context) layout.Dimensions {
					return tv.layoutBiometricModalContent(gtx, th)
				})
			}
			return layout.Dimensions{}
		}),
	)
}

func (tv *TextView) handleEvents(gtx layout.Context) {
	for i := range tv.AlgoButtons {
		if tv.AlgoButtons[i].Clicked(gtx) {
			tv.SelectedAlgoID = crypto.AvailableAlgorithms[i].ID
			tv.StatusMessage = fmt.Sprintf("Algoritma: %s", crypto.AvailableAlgorithms[i].Name)
			tv.IsError = false
		}
	}

	if tv.SampleBtn.Clicked(gtx) {
		tv.InputEditor.SetText("Pesan rahasia ini dilindungi oleh sandi militer Argon2id + AES-256-GCM.")
		tv.KeyEditor.SetText("RahasiaSaya123!")
		tv.UseDeviceKey = false
		tv.ShowBiometricPrompt = false
		tv.ShowMasterPassModal = false
		tv.StatusMessage = "Contoh teks dan kata sandi berhasil dimuat."
		tv.IsError = false
	}

	if tv.KeyStoreBtn.Clicked(gtx) {
		tv.UseDeviceKey = true
		tv.ShowBiometricPrompt = false
		tv.ShowMasterPassModal = false
		tv.StatusMessage = "Kunci Perangkat Android KeyStore aktif!"
		tv.IsError = false
	}

	if tv.SwitchToCustomPassBtn.Clicked(gtx) {
		tv.UseDeviceKey = false
		tv.ShowBiometricPrompt = false
		tv.ShowMasterPassModal = false
		tv.StatusMessage = "Beralih ke mode kata sandi manual."
		tv.IsError = false
	}

	if tv.CancelBiometricBtn.Clicked(gtx) {
		tv.ShowBiometricPrompt = false
		tv.PendingAction = ""
		tv.StatusMessage = "Autentikasi dibatalkan."
		tv.IsError = false
	}

	if tv.ConfirmBiometricBtn.Clicked(gtx) {
		tv.ShowBiometricPrompt = false
		action := tv.PendingAction
		tv.PendingAction = ""
		
		tv.StatusMessage = "Memverifikasi biometrik..."
		
		go func() {
			err := NativeBiometricPrompt("KripDroid Keamanan", "Konfirmasi sidik jari atau lockscreen")
			if err != nil {
				tv.StatusMessage = fmt.Sprintf("Autentikasi dibatalkan atau gagal: %v", err)
				tv.IsError = true
				if tv.Invalidate != nil {
					tv.Invalidate()
				}
				return
			}
			tv.executeActionWithDeviceKey(action)
			if tv.Invalidate != nil {
				tv.Invalidate()
			}
		}()
	}

	if tv.OpenMasterPassBtn.Clicked(gtx) {
		tv.ShowBiometricPrompt = false
		tv.ShowMasterPassModal = true
		tv.ShowChangeMasterPass = false
		tv.MasterPassEditor.SetText("")
		tv.ConfirmMasterPassEditor.SetText("")
	}

	if tv.CancelMasterPassBtn.Clicked(gtx) {
		tv.ShowMasterPassModal = false
		tv.PendingAction = ""
		tv.StatusMessage = "Verifikasi Master Password dibatalkan."
		tv.IsError = false
	}

	if tv.OpenChangeMasterPassBtn.Clicked(gtx) {
		tv.ShowChangeMasterPass = true
		tv.OldMasterPassEditor.SetText("")
		tv.NewMasterPassEditor.SetText("")
	}

	if tv.CancelChangeMasterPassBt.Clicked(gtx) {
		tv.ShowChangeMasterPass = false
	}

	if tv.SubmitChangeMasterPassBt.Clicked(gtx) {
		oldPass := tv.OldMasterPassEditor.Text()
		newPass := tv.NewMasterPassEditor.Text()
		err := ChangeMasterPassword(oldPass, newPass)
		if err != nil {
			tv.StatusMessage = fmt.Sprintf("Gagal Ganti Master Password: %v", err)
			tv.IsError = true
			return
		}
		tv.ShowChangeMasterPass = false
		tv.StatusMessage = "Master Password berhasil diubah dengan sukses!"
		tv.IsError = false
	}

	if tv.SubmitMasterPassBtn.Clicked(gtx) {
		if !HasMasterPassword() {
			p1 := tv.MasterPassEditor.Text()
			p2 := tv.ConfirmMasterPassEditor.Text()
			if strings.TrimSpace(p1) != strings.TrimSpace(p2) {
				tv.StatusMessage = "Konfirmasi Master Password tidak cocok!"
				tv.IsError = true
				return
			}
			err := SetMasterPassword(p1)
			if err != nil {
				tv.StatusMessage = fmt.Sprintf("Gagal membuat Master Password: %v", err)
				tv.IsError = true
				return
			}
			tv.ShowMasterPassModal = false
			action := tv.PendingAction
			tv.PendingAction = ""
			tv.executeActionWithMasterPass(action, p1)
		} else {
			p := tv.MasterPassEditor.Text()
			if !VerifyMasterPassword(p) {
				tv.StatusMessage = "Master Password salah! Silakan coba lagi."
				tv.IsError = true
				return
			}
			tv.ShowMasterPassModal = false
			action := tv.PendingAction
			tv.PendingAction = ""
			tv.executeActionWithMasterPass(action, p)
		}
	}

	if tv.ToggleMaskBtn.Clicked(gtx) {
		tv.IsKeyMasked = !tv.IsKeyMasked
		if tv.IsKeyMasked {
			tv.KeyEditor.Mask = '•'
		} else {
			tv.KeyEditor.Mask = 0
		}
	}

	if tv.ToggleIvBtn.Clicked(gtx) {
		tv.ShowAdvancedIV = !tv.ShowAdvancedIV
	}

	if tv.ClearBtn.Clicked(gtx) {
		tv.InputEditor.SetText("")
		tv.OutputEditor.SetText("")
		tv.ShowBiometricPrompt = false
		tv.ShowMasterPassModal = false
		tv.StatusMessage = "Kotak teks telah dibersihkan."
		tv.IsError = false
	}

	if tv.SwapBtn.Clicked(gtx) {
		outText := strings.TrimSpace(tv.OutputEditor.Text())
		if outText == "" {
			tv.StatusMessage = "Peringatan: Tidak ada teks hasil untuk dipindahkan ke masukan."
			tv.IsError = true
			return
		}
		tv.InputEditor.SetText(outText)
		tv.OutputEditor.SetText("")
		tv.StatusMessage = "Teks hasil berhasil dipindahkan ke masukan."
		tv.IsError = false
	}

	if tv.CopyBtn.Clicked(gtx) {
		outText := tv.OutputEditor.Text()
		if strings.TrimSpace(outText) == "" {
			tv.StatusMessage = "Peringatan: Belum ada teks hasil untuk disalin."
			tv.IsError = true
			return
		}
		gtx.Execute(clipboard.WriteCmd{
			Type: "application/text",
			Data: io.NopCloser(strings.NewReader(outText)),
		})
		tv.StatusMessage = "Teks berhasil disalin ke papan klip!"
		tv.IsError = false
	}

	if tv.EncryptBtn.Clicked(gtx) {
		plain := tv.InputEditor.Text()
		if strings.TrimSpace(plain) == "" {
			tv.StatusMessage = "Peringatan: Ketik atau tempel teks yang ingin Anda kunci terlebih dahulu."
			tv.IsError = true
			return
		}

		if tv.UseDeviceKey {
			tv.PendingAction = "encrypt"
			tv.ShowBiometricPrompt = true
			tv.ShowMasterPassModal = false
			LaunchNativeAndroidBiometricPrompt("KripDroid Keamanan", "Konfirmasi sidik jari atau lockscreen untuk mengunci teks")
			tv.StatusMessage = "Konfirmasi autentikasi biometrik atau Master Password."
			tv.IsError = false
			return
		}

		algo, _ := crypto.GetAlgorithm(tv.SelectedAlgoID)
		key := tv.KeyEditor.Text()
		if algo != nil && algo.NeedsKey && strings.TrimSpace(key) == "" {
			tv.StatusMessage = "Peringatan: Kata sandi masih kosong. Masukkan kata sandi pengaman."
			tv.IsError = true
			return
		}

		iv := tv.IvEditor.Text()
		res, _, err := crypto.EncryptText(tv.SelectedAlgoID, plain, key, iv)
		if err != nil {
			tv.StatusMessage = fmt.Sprintf("Gagal Mengunci Teks: %v", err)
			tv.IsError = true
			return
		}
		tv.OutputEditor.SetText(res)
		tv.StatusMessage = "Teks berhasil dikunci dengan proteksi Argon2id! Anda dapat menyalinnya."
		tv.IsError = false
	}

	if tv.DecryptBtn.Clicked(gtx) {
		cipher := tv.InputEditor.Text()
		if strings.TrimSpace(cipher) == "" {
			tv.StatusMessage = "Peringatan: Masukkan teks rahasia yang ingin Anda buka di kotak masukan."
			tv.IsError = true
			return
		}

		if tv.UseDeviceKey {
			tv.PendingAction = "decrypt"
			tv.ShowBiometricPrompt = true
			tv.ShowMasterPassModal = false
			LaunchNativeAndroidBiometricPrompt("KripDroid Keamanan", "Konfirmasi sidik jari atau lockscreen untuk membuka teks")
			tv.StatusMessage = "Konfirmasi autentikasi biometrik atau Master Password."
			tv.IsError = false
			return
		}

		algo, _ := crypto.GetAlgorithm(tv.SelectedAlgoID)
		key := tv.KeyEditor.Text()
		if algo != nil && algo.NeedsKey && strings.TrimSpace(key) == "" {
			tv.StatusMessage = "Peringatan: Masukkan kata sandi yang digunakan saat mengunci teks."
			tv.IsError = true
			return
		}

		iv := tv.IvEditor.Text()
		res, err := crypto.DecryptText(tv.SelectedAlgoID, cipher, key, iv)
		if err != nil {
			tv.StatusMessage = "Gagal Membuka Teks: Kata sandi tidak cocok atau data rusak."
			tv.IsError = true
			return
		}
		tv.OutputEditor.SetText(res)
		tv.StatusMessage = "Teks berhasil dibuka kembali dengan sukses!"
		tv.IsError = false
	}
}

func (tv *TextView) executeActionWithDeviceKey(action string) {
	hwKey, err := GetDeviceHardwareKey()
	if err != nil {
		tv.StatusMessage = fmt.Sprintf("Gagal mengakses Kunci Perangkat: %v", err)
		tv.IsError = true
		return
	}

	iv := tv.IvEditor.Text()

	if action == "encrypt" {
		plain := tv.InputEditor.Text()
		res, _, err := crypto.EncryptText(tv.SelectedAlgoID, plain, hwKey, iv)
		if err != nil {
			tv.StatusMessage = fmt.Sprintf("Gagal Mengunci Teks: %v", err)
			tv.IsError = true
			return
		}
		tv.OutputEditor.SetText(res)
		tv.StatusMessage = "Verifikasi biometrik berhasil! Teks terkunci aman dengan Kunci Perangkat Hardware + Argon2id."
		tv.IsError = false
	} else if action == "decrypt" {
		cipher := tv.InputEditor.Text()
		res, err := crypto.DecryptText(tv.SelectedAlgoID, cipher, hwKey, iv)
		if err != nil {
			tv.StatusMessage = "Gagal Membuka Teks: Teks ini tidak dikunci dengan Kunci Perangkat HP ini atau data rusak."
			tv.IsError = true
			return
		}
		tv.OutputEditor.SetText(res)
		tv.StatusMessage = "Verifikasi biometrik berhasil! Teks sukses dibuka dengan Kunci Perangkat Android."
		tv.IsError = false
	}
}

func (tv *TextView) executeActionWithMasterPass(action, masterPass string) {
	derivedKey, err := GetDeviceMasterKey(masterPass)
	if err != nil {
		tv.StatusMessage = fmt.Sprintf("Gagal memproses Master Password: %v", err)
		tv.IsError = true
		return
	}

	iv := tv.IvEditor.Text()

	if action == "encrypt" {
		plain := tv.InputEditor.Text()
		res, _, err := crypto.EncryptText(tv.SelectedAlgoID, plain, derivedKey, iv)
		if err != nil {
			tv.StatusMessage = fmt.Sprintf("Gagal Mengunci Teks: %v", err)
			tv.IsError = true
			return
		}
		tv.OutputEditor.SetText(res)
		tv.StatusMessage = "Teks berhasil dikunci aman dengan Master Password + Argon2id!"
		tv.IsError = false
	} else if action == "decrypt" {
		cipher := tv.InputEditor.Text()
		res, err := crypto.DecryptText(tv.SelectedAlgoID, cipher, derivedKey, iv)
		if err != nil {
			tv.StatusMessage = "Gagal Membuka Teks: Master Password tidak sesuai atau data teks rusak."
			tv.IsError = true
			return
		}
		tv.OutputEditor.SetText(res)
		tv.StatusMessage = "Teks sukses dibuka dengan Master Password!"
		tv.IsError = false
	}
}

func (tv *TextView) layoutAlgorithmSelector(gtx layout.Context, th *M3Theme) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
				Spacing:   layout.SpaceBetween,
			}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return DrawSectionTitle(gtx, th, "PILIH ALGORITMA KRIPTOGRAFI")
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return DrawM3Chip(gtx, th, &tv.SampleBtn, "Coba Contoh", false)
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			var rows []layout.FlexChild
			var currentRow []layout.FlexChild

			for i, algo := range crypto.AvailableAlgorithms {
				idx := i
				algoItem := algo
				isSelected := tv.SelectedAlgoID == algoItem.ID

				btnChild := layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Right:  unit.Dp(3),
						Bottom: unit.Dp(4),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return DrawM3Chip(gtx, th, &tv.AlgoButtons[idx], algoItem.Name, isSelected)
					})
				})

				currentRow = append(currentRow, btnChild)
				if len(currentRow) >= 2 || i == len(crypto.AvailableAlgorithms)-1 {
					rowChildren := currentRow
					rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							Axis: layout.Horizontal,
						}.Layout(gtx, rowChildren...)
					}))
					currentRow = nil
				}
			}

			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx, rows...)
		}),
	)
}

func (tv *TextView) layoutAlgorithmInfoCard(gtx layout.Context, th *M3Theme, algo *crypto.AlgorithmInfo) layout.Dimensions {
	simpleDesc := "Tingkat Sangat Kuat & Aman (Standar Militer + Argon2id). Sangat direkomendasikan."
	if algo.Category == crypto.CategoryClassic {
		simpleDesc = "Sandi Klasik Sederhana (Untuk sarana edukasi & latihan logika sandi)."
	} else if algo.Category == crypto.CategoryLegacy {
		simpleDesc = "Sandi Standar Menengah (Algoritma legasi + Argon2id)."
	}

	return DrawOutlinedCard(gtx, th.SurfaceContainer, th.OutlineVariant, unit.Dp(12), unit.Dp(1), func(gtx layout.Context) layout.Dimensions {
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
					return layout.Flex{
						Axis:      layout.Horizontal,
						Alignment: layout.Middle,
						Spacing:   layout.SpaceBetween,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(th.MaterialTheme, unit.Sp(13), algo.Name)
							lbl.Color = th.OnSurface
							lbl.Font.Weight = font.Bold
							return lbl.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return DrawSecurityBadge(gtx, th, string(algo.Category), algo.SecurityRating)
						}),
					)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(3)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.MaterialTheme, unit.Sp(11), simpleDesc)
					lbl.Color = th.OnSurfaceVariant
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if algo.NeedsKey {
						return DrawPill(gtx, th.SecondaryContainer, th.OnSecondaryContainer, "KDF: Argon2id (Anti-Brute Force)", unit.Sp(10), th)
					}
					return layout.Dimensions{}
				}),
			)
		})
	})
}

func (tv *TextView) layoutInputSection(gtx layout.Context, th *M3Theme, algo *crypto.AlgorithmInfo) layout.Dimensions {
	inputLen := len(tv.InputEditor.Text())
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
				Spacing:   layout.SpaceBetween,
			}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return DrawSectionTitle(gtx, th, "KOTAK TEKS MASUKAN")
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return DrawPill(gtx, th.SurfaceContainerHigh, th.OnSurfaceVariant, fmt.Sprintf("%d karakter", inputLen), unit.Sp(10), th)
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return DrawOutlinedCard(gtx, th.SurfaceContainer, th.OutlineVariant, unit.Dp(12), unit.Dp(1), func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top:    unit.Dp(10),
					Bottom: unit.Dp(10),
					Left:   unit.Dp(12),
					Right:  unit.Dp(12),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					ed := material.Editor(th.MaterialTheme, &tv.InputEditor, "Ketik atau tempel teks yang ingin dikunci atau dibuka...")
					ed.Color = th.OnSurface
					ed.HintColor = th.Outline
					return ed.Layout(gtx)
				})
			})
		}),
	)
}

func (tv *TextView) layoutKeySection(gtx layout.Context, th *M3Theme, algo *crypto.AlgorithmInfo) layout.Dimensions {
	if algo != nil && !algo.NeedsKey {
		return layout.Dimensions{}
	}

	isAndroid := IsKeyStoreSupported()

	if tv.UseDeviceKey {
		return DrawCard(gtx, th.SecondaryContainer, unit.Dp(12), func(gtx layout.Context) layout.Dimensions {
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
						return DrawPill(gtx, th.Secondary, th.OnSecondary, "KUNCI PERANGKAT ANDROID AKTIF", unit.Sp(11), th)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(th.MaterialTheme, unit.Sp(11), "Teks diamankan hardware KeyStore & Lockscreen HP. Memerlukan Biometrik atau Master Password.")
						lbl.Color = th.OnSecondaryContainer
						return lbl.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return DrawM3Button(gtx, th, &tv.SwitchToCustomPassBtn, "GANTI KE KATA SANDI MANUAL", th.SurfaceContainerHigh, th.OnSurface, unit.Dp(10))
					}),
				)
			})
		})
	}

	maskLabel := "LIHAT"
	if !tv.IsKeyMasked {
		maskLabel = "TUTUP"
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
				Spacing:   layout.SpaceBetween,
			}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return DrawSectionTitle(gtx, th, "KATA SANDI / KUNCI PENGAMAN")
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return DrawM3OutlinedButton(gtx, th, &tv.ToggleMaskBtn, maskLabel, th.OutlineVariant, th.OnSurfaceVariant, unit.Dp(8))
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return DrawOutlinedCard(gtx, th.SurfaceContainer, th.OutlineVariant, unit.Dp(12), unit.Dp(1), func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top:    unit.Dp(10),
					Bottom: unit.Dp(10),
					Left:   unit.Dp(12),
					Right:  unit.Dp(12),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					ed := material.Editor(th.MaterialTheme, &tv.KeyEditor, "Masukkan kata sandi rahasia Anda...")
					ed.Color = th.OnSurface
					ed.HintColor = th.Outline
					return ed.Layout(gtx)
				})
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if isAndroid {
				return layout.Inset{Top: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return DrawM3Button(gtx, th, &tv.KeyStoreBtn, "GUNAKAN KUNCI PERANGKAT (KEYSTORE HP)", th.SecondaryContainer, th.OnSecondaryContainer, unit.Dp(10))
				})
			}
			return layout.Dimensions{}
		}),
	)
}

func (tv *TextView) layoutBiometricModalContent(gtx layout.Context, th *M3Theme) layout.Dimensions {
	actionName := "Mengunci Teks"
	if tv.PendingAction == "decrypt" {
		actionName = "Membuka Teks"
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th.MaterialTheme, unit.Sp(14), "VERIFIKASI KUNCI PERANGKAT")
			lbl.Color = th.Primary
			lbl.Font.Weight = font.Bold
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th.MaterialTheme, unit.Sp(12), fmt.Sprintf("Autentikasi sidik jari atau master password untuk %s.", actionName))
			lbl.Color = th.OnSurfaceVariant
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return DrawM3Button(gtx, th, &tv.ConfirmBiometricBtn, "VERIFIKASI BIOMETRIK / LOCKSCREEN", th.Secondary, th.OnSecondary, unit.Dp(10))
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return DrawM3Button(gtx, th, &tv.OpenMasterPassBtn, "GUNAKAN MASTER PASSWORD", th.SurfaceContainer, th.OnSurface, unit.Dp(10))
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return DrawM3Button(gtx, th, &tv.CancelBiometricBtn, "BATAL", th.SurfaceContainer, th.Error, unit.Dp(10))
		}),
	)
}

func (tv *TextView) layoutMasterPassModalContent(gtx layout.Context, th *M3Theme) layout.Dimensions {
	hasMaster := HasMasterPassword()

	if tv.ShowChangeMasterPass {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(th.MaterialTheme, unit.Sp(14), "GANTI MASTER PASSWORD")
				lbl.Color = th.Primary
				lbl.Font.Weight = font.Bold
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DrawOutlinedCard(gtx, th.SurfaceContainer, th.OutlineVariant, unit.Dp(8), unit.Dp(1), func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8), Left: unit.Dp(10), Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						ed := material.Editor(th.MaterialTheme, &tv.OldMasterPassEditor, "Master Password Lama...")
						ed.Color = th.OnSurface
						ed.HintColor = th.Outline
						return ed.Layout(gtx)
					})
				})
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DrawOutlinedCard(gtx, th.SurfaceContainer, th.OutlineVariant, unit.Dp(8), unit.Dp(1), func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8), Left: unit.Dp(10), Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						ed := material.Editor(th.MaterialTheme, &tv.NewMasterPassEditor, "Master Password Baru...")
						ed.Color = th.OnSurface
						ed.HintColor = th.Outline
						return ed.Layout(gtx)
					})
				})
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DrawM3Button(gtx, th, &tv.SubmitChangeMasterPassBt, "SIMPAN PASSWORD BARU", th.Primary, th.OnPrimary, unit.Dp(10))
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DrawM3Button(gtx, th, &tv.CancelChangeMasterPassBt, "KEMBALI", th.SurfaceContainer, th.OnSurfaceVariant, unit.Dp(10))
			}),
		)
	}

	title := "MASUKKAN MASTER PASSWORD"
	hint := "Masukkan Master Password..."
	if !hasMaster {
		title = "BUAT MASTER PASSWORD BARU"
		hint = "Buat Master Password baru..."
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
				Spacing:   layout.SpaceBetween,
			}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.MaterialTheme, unit.Sp(14), title)
					lbl.Color = th.Primary
					lbl.Font.Weight = font.Bold
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if hasMaster {
						return DrawM3Chip(gtx, th, &tv.OpenChangeMasterPassBtn, "Ganti Sandi", false)
					}
					return layout.Dimensions{}
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return DrawOutlinedCard(gtx, th.SurfaceContainer, th.OutlineVariant, unit.Dp(8), unit.Dp(1), func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8), Left: unit.Dp(10), Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					ed := material.Editor(th.MaterialTheme, &tv.MasterPassEditor, hint)
					ed.Color = th.OnSurface
					ed.HintColor = th.Outline
					return ed.Layout(gtx)
				})
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !hasMaster {
				return layout.Inset{Top: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return DrawOutlinedCard(gtx, th.SurfaceContainer, th.OutlineVariant, unit.Dp(8), unit.Dp(1), func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8), Left: unit.Dp(10), Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							ed := material.Editor(th.MaterialTheme, &tv.ConfirmMasterPassEditor, "Konfirmasi Master Password baru...")
							ed.Color = th.OnSurface
							ed.HintColor = th.Outline
							return ed.Layout(gtx)
						})
					})
				})
			}
			return layout.Dimensions{}
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return DrawM3Button(gtx, th, &tv.SubmitMasterPassBtn, "KONFIRMASI MASTER PASSWORD", th.Secondary, th.OnSecondary, unit.Dp(10))
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return DrawM3Button(gtx, th, &tv.CancelMasterPassBtn, "BATAL", th.SurfaceContainer, th.Error, unit.Dp(10))
		}),
	)
}

func (tv *TextView) layoutActionButtons(gtx layout.Context, th *M3Theme) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Horizontal,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return DrawM3Button(gtx, th, &tv.EncryptBtn, "KUNCI TEKS", th.Primary, th.OnPrimary, unit.Dp(12))
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return DrawM3Button(gtx, th, &tv.DecryptBtn, "BUKA TEKS", th.Secondary, th.OnSecondary, unit.Dp(12))
		}),
	)
}

func (tv *TextView) layoutOutputSection(gtx layout.Context, th *M3Theme) layout.Dimensions {
	outputLen := len(tv.OutputEditor.Text())
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
				Spacing:   layout.SpaceBetween,
			}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return DrawSectionTitle(gtx, th, "HASIL")
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Horizontal,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return DrawM3Chip(gtx, th, &tv.SwapBtn, "Tukar", false)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return DrawM3Chip(gtx, th, &tv.ClearBtn, "Bersihkan", false)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return DrawM3Chip(gtx, th, &tv.CopyBtn, "Salin", true)
						}),
					)
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return DrawOutlinedCard(gtx, th.SurfaceContainer, th.OutlineVariant, unit.Dp(12), unit.Dp(1), func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top:    unit.Dp(10),
					Bottom: unit.Dp(10),
					Left:   unit.Dp(12),
					Right:  unit.Dp(12),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					ed := material.Editor(th.MaterialTheme, &tv.OutputEditor, "Hasil penguncian atau pembukaan teks akan muncul di sini...")
					ed.Color = th.OnSurface
					ed.HintColor = th.Outline
					return ed.Layout(gtx)
				})
			})
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(3)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if outputLen > 0 {
				lbl := material.Label(th.MaterialTheme, unit.Sp(11), fmt.Sprintf("Panjang teks hasil: %d karakter", len([]rune(tv.OutputEditor.Text()))))
				lbl.Color = th.OnSurfaceVariant
				return lbl.Layout(gtx)
			}
			return layout.Dimensions{}
		}),
	)
}

func (tv *TextView) layoutStatusBanner(gtx layout.Context, th *M3Theme) layout.Dimensions {
	if tv.StatusMessage == "" {
		return layout.Dimensions{}
	}

	var bg, fg color.NRGBA
	if tv.IsError {
		bg = th.ErrorContainer
		fg = th.OnErrorContainer
	} else {
		bg = th.SecondaryContainer
		fg = th.OnSecondaryContainer
	}

	return DrawCard(gtx, bg, unit.Dp(10), func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(8),
			Bottom: unit.Dp(8),
			Left:   unit.Dp(12),
			Right:  unit.Dp(12),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th.MaterialTheme, unit.Sp(12), tv.StatusMessage)
			lbl.Color = fg
			lbl.Font.Weight = font.Medium
			return lbl.Layout(gtx)
		})
	})
}
