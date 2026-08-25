package ui

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gioui.org/font"
	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/explorer"

	"kripdroid/crypto"
)

type NameGetter interface {
	Name() string
}

type chunkReader struct {
	r io.Reader
}

func (c *chunkReader) Read(p []byte) (int, error) {
	if len(p) > 64*1024 {
		p = p[:64*1024]
	}
	return c.r.Read(p)
}

type PersistedState struct {
	LoadedFileName    string `json:"loaded_file_name"`
	LoadedFilePath    string `json:"loaded_file_path"`
	LoadedFileSize    int64  `json:"loaded_file_size"`
	ProcessedFileName string `json:"processed_file_name"`
	ProcessedFilePath string `json:"processed_file_path"`
	ProcessedFileSize int64  `json:"processed_file_size"`
	RecoveredExt      string `json:"recovered_ext"`
	SelectedAlgoID    string `json:"selected_algo_id"`
	StatusMessage     string `json:"status_message"`
}

type FileView struct {
	Explorer                 *explorer.Explorer
	SrcPathEditor            widget.Editor
	DstPathEditor            widget.Editor
	KeyEditor                widget.Editor
	MasterPassEditor         widget.Editor
	ConfirmMasterPassEditor  widget.Editor
	OldMasterPassEditor      widget.Editor
	NewMasterPassEditor      widget.Editor
	SelectedAlgoID           string
	UseDeviceKey             bool
	ShowBiometricPrompt      bool
	ShowMasterPassModal      bool
	ShowChangeMasterPass     bool
	PendingAction            string
	PickFileBtn              widget.Clickable
	SaveFileBtn              widget.Clickable
	QuickSaveBtn             widget.Clickable
	CopyBase64Btn            widget.Clickable
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
	EncryptBtn               widget.Clickable
	DecryptBtn               widget.Clickable
	CreateSampleBtn          widget.Clickable
	ClearBtn                 widget.Clickable
	AlgoButtons              []widget.Clickable
	List                     widget.List
	IsProcessing             bool
	Progress                 float32
	StatusMessage            string
	IsError                  bool
	LastResult               *crypto.FileProcessResult
	LoadedFileName           string
	LoadedFilePath           string
	LoadedFileSize           int64
	ProcessedFileName        string
	ProcessedFilePath        string
	ProcessedFileSize        int64
	RecoveredExt             string
	SavedLocation            string
	Invalidate               func()
	mu                       sync.Mutex
}

func NewFileView(expl *explorer.Explorer, invalidate func()) *FileView {
	fv := &FileView{
		Explorer:       expl,
		SelectedAlgoID: "aes256gcm",
		UseDeviceKey:   false,
		AlgoButtons:    make([]widget.Clickable, len(crypto.AvailableAlgorithms)),
		Invalidate:     invalidate,
	}

	fv.List.Axis = layout.Vertical
	fv.SrcPathEditor.SingleLine = true
	fv.DstPathEditor.SingleLine = true
	fv.KeyEditor.SingleLine = true
	fv.MasterPassEditor.SingleLine = true
	fv.MasterPassEditor.Mask = '•'
	fv.ConfirmMasterPassEditor.SingleLine = true
	fv.ConfirmMasterPassEditor.Mask = '•'
	fv.OldMasterPassEditor.SingleLine = true
	fv.OldMasterPassEditor.Mask = '•'
	fv.NewMasterPassEditor.SingleLine = true
	fv.NewMasterPassEditor.Mask = '•'

	fv.LoadState()

	return fv
}

func (fv *FileView) SaveState() {
	state := PersistedState{
		LoadedFileName:    fv.LoadedFileName,
		LoadedFilePath:    fv.LoadedFilePath,
		LoadedFileSize:    fv.LoadedFileSize,
		ProcessedFileName: fv.ProcessedFileName,
		ProcessedFilePath: fv.ProcessedFilePath,
		ProcessedFileSize: fv.ProcessedFileSize,
		RecoveredExt:      fv.RecoveredExt,
		SelectedAlgoID:    fv.SelectedAlgoID,
		StatusMessage:     fv.StatusMessage,
	}
	data, err := json.Marshal(state)
	if err != nil {
		return
	}
	_ = os.WriteFile(GetStateFilePath(), data, 0644)
}

func (fv *FileView) LoadState() {
	data, err := os.ReadFile(GetStateFilePath())
	if err != nil {
		return
	}
	var state PersistedState
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}

	if state.LoadedFilePath != "" {
		if fi, err := os.Stat(state.LoadedFilePath); err == nil && fi.Size() > 0 {
			fv.LoadedFilePath = state.LoadedFilePath
			fv.LoadedFileName = state.LoadedFileName
			fv.LoadedFileSize = fi.Size()
			fv.SrcPathEditor.SetText(fmt.Sprintf("%s (%d byte)", fv.LoadedFileName, fv.LoadedFileSize))
		}
	}

	if state.ProcessedFilePath != "" {
		if fi, err := os.Stat(state.ProcessedFilePath); err == nil && fi.Size() > 0 {
			fv.ProcessedFilePath = state.ProcessedFilePath
			fv.ProcessedFileName = state.ProcessedFileName
			fv.ProcessedFileSize = fi.Size()
			fv.RecoveredExt = state.RecoveredExt
			fv.LastResult = &crypto.FileProcessResult{
				AlgoID:      state.SelectedAlgoID,
				OriginalExt: state.RecoveredExt,
				InputSize:   fv.LoadedFileSize,
				OutputSize:  fi.Size(),
			}
			fv.StatusMessage = "Sesi sebelumnya dipulihkan: File hasil siap disimpan atau dibuka."
		}
	}

	if state.SelectedAlgoID != "" {
		fv.SelectedAlgoID = state.SelectedAlgoID
	}
}

func (fv *FileView) ClearState() {
	_ = os.Remove(GetStateFilePath())
}

func (fv *FileView) Layout(gtx layout.Context, th *M3Theme) layout.Dimensions {
	fv.handleEvents(gtx)

	algo, _ := crypto.GetAlgorithm(fv.SelectedAlgoID)

	var items []layout.Widget

	items = append(items, func(gtx layout.Context) layout.Dimensions {
		return fv.layoutAlgorithmSelector(gtx, th)
	})

	if algo != nil {
		items = append(items, func(gtx layout.Context) layout.Dimensions {
			return fv.layoutAlgorithmInfoCard(gtx, th, algo)
		})
	}

	items = append(items, func(gtx layout.Context) layout.Dimensions {
		return fv.layoutExplorerPickSection(gtx, th)
	})

	items = append(items, func(gtx layout.Context) layout.Dimensions {
		return fv.layoutKeySection(gtx, th, algo)
	})

	items = append(items, func(gtx layout.Context) layout.Dimensions {
		return fv.layoutActionButtons(gtx, th)
	})

	if fv.IsProcessing || fv.Progress > 0 || fv.StatusMessage != "" {
		items = append(items, func(gtx layout.Context) layout.Dimensions {
			return fv.layoutProgressSection(gtx, th)
		})
	}

	if fv.LastResult != nil {
		items = append(items, func(gtx layout.Context) layout.Dimensions {
			return fv.layoutResultSection(gtx, th)
		})
	}

	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return fv.List.Layout(gtx, len(items), func(gtx layout.Context, index int) layout.Dimensions {
				return layout.Inset{
					Top:    unit.Dp(4),
					Bottom: unit.Dp(4),
					Left:   unit.Dp(12),
					Right:  unit.Dp(12),
				}.Layout(gtx, items[index])
			})
		}),
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			if fv.ShowMasterPassModal {
				return DrawModalDialog(gtx, th, func(gtx layout.Context) layout.Dimensions {
					return fv.layoutMasterPassModalContent(gtx, th)
				})
			}
			if fv.ShowBiometricPrompt {
				return DrawModalDialog(gtx, th, func(gtx layout.Context) layout.Dimensions {
					return fv.layoutBiometricModalContent(gtx, th)
				})
			}
			return layout.Dimensions{}
		}),
	)
}

func (fv *FileView) handleEvents(gtx layout.Context) {
	for i := range fv.AlgoButtons {
		if fv.AlgoButtons[i].Clicked(gtx) {
			fv.SelectedAlgoID = crypto.AvailableAlgorithms[i].ID
			fv.StatusMessage = fmt.Sprintf("Algoritma: %s", crypto.AvailableAlgorithms[i].Name)
			fv.IsError = false
			fv.SaveState()
		}
	}

	hasUnsavedFile := fv.ProcessedFilePath != ""

	if fv.KeyStoreBtn.Clicked(gtx) {
		fv.UseDeviceKey = true
		fv.ShowBiometricPrompt = false
		fv.ShowMasterPassModal = false
		fv.StatusMessage = "Kunci Perangkat Android KeyStore aktif!"
		fv.IsError = false
	}

	if fv.SwitchToCustomPassBtn.Clicked(gtx) {
		fv.UseDeviceKey = false
		fv.ShowBiometricPrompt = false
		fv.ShowMasterPassModal = false
		fv.StatusMessage = "Beralih ke mode kata sandi manual."
		fv.IsError = false
	}

	if fv.CancelBiometricBtn.Clicked(gtx) {
		fv.ShowBiometricPrompt = false
		fv.PendingAction = ""
		fv.StatusMessage = "Autentikasi dibatalkan."
		fv.IsError = false
	}

	if fv.ConfirmBiometricBtn.Clicked(gtx) {
		fv.ShowBiometricPrompt = false
		action := fv.PendingAction
		fv.PendingAction = ""
		
		fv.IsProcessing = true
		fv.StatusMessage = "Memverifikasi biometrik..."
		
		go func() {
			err := NativeBiometricPrompt("KripDroid Keamanan", "Konfirmasi sidik jari atau lockscreen")
			if err != nil {
				fv.StatusMessage = fmt.Sprintf("Autentikasi dibatalkan atau gagal: %v", err)
				fv.IsError = true
				fv.IsProcessing = false
				if fv.Invalidate != nil {
					fv.Invalidate()
				}
				return
			}
			fv.executeFileActionWithDeviceKey(action)
			if fv.Invalidate != nil {
				fv.Invalidate()
			}
		}()
	}

	if fv.OpenMasterPassBtn.Clicked(gtx) {
		fv.ShowBiometricPrompt = false
		fv.ShowMasterPassModal = true
		fv.ShowChangeMasterPass = false
		fv.MasterPassEditor.SetText("")
		fv.ConfirmMasterPassEditor.SetText("")
	}

	if fv.CancelMasterPassBtn.Clicked(gtx) {
		fv.ShowMasterPassModal = false
		fv.PendingAction = ""
		fv.StatusMessage = "Verifikasi Master Password dibatalkan."
		fv.IsError = false
	}

	if fv.OpenChangeMasterPassBtn.Clicked(gtx) {
		fv.ShowChangeMasterPass = true
		fv.OldMasterPassEditor.SetText("")
		fv.NewMasterPassEditor.SetText("")
	}

	if fv.CancelChangeMasterPassBt.Clicked(gtx) {
		fv.ShowChangeMasterPass = false
	}

	if fv.SubmitChangeMasterPassBt.Clicked(gtx) {
		oldPass := fv.OldMasterPassEditor.Text()
		newPass := fv.NewMasterPassEditor.Text()
		err := ChangeMasterPassword(oldPass, newPass)
		if err != nil {
			fv.StatusMessage = fmt.Sprintf("Gagal Ganti Master Password: %v", err)
			fv.IsError = true
			return
		}
		fv.ShowChangeMasterPass = false
		fv.StatusMessage = "Master Password berhasil diubah dengan sukses!"
		fv.IsError = false
	}

	if fv.SubmitMasterPassBtn.Clicked(gtx) {
		if !HasMasterPassword() {
			p1 := fv.MasterPassEditor.Text()
			p2 := fv.ConfirmMasterPassEditor.Text()
			if strings.TrimSpace(p1) != strings.TrimSpace(p2) {
				fv.StatusMessage = "Konfirmasi Master Password tidak cocok!"
				fv.IsError = true
				return
			}
			err := SetMasterPassword(p1)
			if err != nil {
				fv.StatusMessage = fmt.Sprintf("Gagal membuat Master Password: %v", err)
				fv.IsError = true
				return
			}
			fv.ShowMasterPassModal = false
			action := fv.PendingAction
			fv.PendingAction = ""
			fv.executeFileActionWithMasterPass(action, p1)
		} else {
			p := fv.MasterPassEditor.Text()
			if !VerifyMasterPassword(p) {
				fv.StatusMessage = "Master Password salah! Silakan coba lagi."
				fv.IsError = true
				return
			}
			fv.ShowMasterPassModal = false
			action := fv.PendingAction
			fv.PendingAction = ""
			fv.executeFileActionWithMasterPass(action, p)
		}
	}

	if fv.PickFileBtn.Clicked(gtx) && !fv.IsProcessing {
		if hasUnsavedFile {
			fv.StatusMessage = "Perhatian: Simpan file hasil proses di bawah terlebih dahulu, atau klik 'BERSIHKAN'."
			fv.IsError = true
			return
		}

		fv.IsProcessing = true
		fv.StatusMessage = "Membuka pemilihan file dari perangkat..."
		fv.IsError = false

		go func() {
			var reader io.ReadCloser
			var err error

			if fv.Explorer != nil {
				reader, err = fv.Explorer.ChooseFile()
			} else {
				err = fmt.Errorf("file explorer tidak tersedia")
			}

			if err != nil {
				fv.mu.Lock()
				fv.IsProcessing = false
				fv.StatusMessage = fmt.Sprintf("Batal memilih file: %v", err)
				fv.IsError = true
				fv.mu.Unlock()
				if fv.Invalidate != nil {
					fv.Invalidate()
				}
				return
			}

			if reader != nil {
				detectedName := ""
				if ng, ok := reader.(NameGetter); ok && ng.Name() != "" {
					detectedName = ng.Name()
				} else if f, ok := reader.(*os.File); ok && f != nil {
					detectedName = filepath.Base(f.Name())
				}
				if detectedName == "" {
					detectedName = "berkas_pilihan.bin"
				}

				tmpFile, createErr := CreateSafeTempFile("krip_in_*")
				if createErr != nil {
					_ = reader.Close()
					fv.mu.Lock()
					fv.IsProcessing = false
					fv.StatusMessage = fmt.Sprintf("Gagal menyiapkan media penyimpanan: %v", createErr)
					fv.IsError = true
					fv.mu.Unlock()
					if fv.Invalidate != nil {
						fv.Invalidate()
					}
					return
				}

				buf := make([]byte, 64*1024)
				copied, copyErr := io.CopyBuffer(tmpFile, &chunkReader{r: reader}, buf)
				_ = tmpFile.Close()
				_ = reader.Close()

				fv.mu.Lock()
				fv.IsProcessing = false
				if copyErr != nil {
					_ = os.Remove(tmpFile.Name())
					fv.StatusMessage = fmt.Sprintf("Gagal membaca file: %v", copyErr)
					fv.IsError = true
				} else if copied == 0 {
					_ = os.Remove(tmpFile.Name())
					fv.StatusMessage = "Peringatan: File yang dipilih kosong (0 byte)."
					fv.IsError = true
				} else {
					if fv.LoadedFilePath != "" {
						_ = os.Remove(fv.LoadedFilePath)
					}
					fv.LoadedFilePath = tmpFile.Name()
					fv.LoadedFileSize = copied
					fv.LoadedFileName = detectedName
					fv.SrcPathEditor.SetText(fmt.Sprintf("%s (%d byte)", detectedName, copied))
					fv.StatusMessage = fmt.Sprintf("File siap diproses: %s (%d byte)", detectedName, copied)
					fv.IsError = false
					fv.SaveState()
				}
				fv.mu.Unlock()

				if fv.Invalidate != nil {
					fv.Invalidate()
				}
			}
		}()
	}

	if fv.CopyBase64Btn.Clicked(gtx) && fv.ProcessedFilePath != "" {
		if fv.ProcessedFileSize > 5*1024*1024 {
			fv.StatusMessage = "Ukuran file terlalu besar untuk papan klip (>5 MB). Silakan gunakan tombol Simpan File."
			fv.IsError = true
		} else {
			data, err := os.ReadFile(fv.ProcessedFilePath)
			if err != nil {
				fv.StatusMessage = fmt.Sprintf("Gagal membaca hasil: %v", err)
				fv.IsError = true
			} else {
				b64 := base64.StdEncoding.EncodeToString(data)
				gtx.Execute(clipboard.WriteCmd{
					Type: "application/text",
					Data: io.NopCloser(strings.NewReader(b64)),
				})
				fv.StatusMessage = fmt.Sprintf("Data file (%d byte) disalin ke papan klip dalam format teks!", len(data))
				fv.IsError = false
			}
		}
	}

	if fv.QuickSaveBtn.Clicked(gtx) && !fv.IsProcessing && fv.ProcessedFilePath != "" {
		outName := fv.ProcessedFileName
		if outName == "" {
			if fv.RecoveredExt != "" {
				outName = "hasil_buka_sandi" + fv.RecoveredExt
			} else {
				outName = "kripdroid_hasil.bin"
			}
		}

		targetPath, err := SaveFileFromPathToWindowsDownloads(outName, fv.ProcessedFilePath)
		if err == nil && targetPath != "" {
			fv.SavedLocation = targetPath
			fv.StatusMessage = fmt.Sprintf("File berhasil disimpan di folder Downloads: %s", targetPath)
			fv.IsError = false
			fv.SaveState()
		} else {
			fv.StatusMessage = fmt.Sprintf("Gagal menyimpan ke folder Downloads: %v (Gunakan opsi Simpan ke Folder Lain)", err)
			fv.IsError = true
		}
	}

	if fv.SaveFileBtn.Clicked(gtx) && !fv.IsProcessing && fv.ProcessedFilePath != "" {
		fv.IsProcessing = true
		fv.StatusMessage = "Membuka folder penyimpanan perangkat..."
		fv.IsError = false

		outName := fv.ProcessedFileName
		if outName == "" {
			if fv.RecoveredExt != "" {
				outName = "hasil_buka_sandi" + fv.RecoveredExt
			} else {
				outName = "kripdroid_hasil.bin"
			}
		}

		safName := outName
		ext := filepath.Ext(safName)
		if ext == ".krip" {
			safName = safName + ".bin"
		}

		srcPath := fv.ProcessedFilePath

		go func() {
			var writer io.WriteCloser
			var err error

			if fv.Explorer != nil {
				writer, err = fv.Explorer.CreateFile(safName)
			} else {
				err = fmt.Errorf("penyimpanan tidak tersedia")
			}

			if err != nil {
				fv.mu.Lock()
				fv.IsProcessing = false
				fv.StatusMessage = fmt.Sprintf("Batal menyimpan file: %v", err)
				fv.IsError = true
				fv.mu.Unlock()
				if fv.Invalidate != nil {
					fv.Invalidate()
				}
				return
			}

			var writeErr error
			if writer != nil {
				srcFile, openErr := os.Open(srcPath)
				if openErr != nil {
					writeErr = openErr
				} else {
					buf := make([]byte, 64*1024)
					_, writeErr = io.CopyBuffer(writer, srcFile, buf)
					_ = srcFile.Close()
				}
				if closeErr := writer.Close(); closeErr != nil && writeErr == nil {
					writeErr = closeErr
				}
			}

			fv.mu.Lock()
			fv.IsProcessing = false
			if writeErr != nil {
				fv.StatusMessage = fmt.Sprintf("Gagal menulis data file: %v", writeErr)
				fv.IsError = true
			} else {
				fv.SavedLocation = safName
				fv.StatusMessage = fmt.Sprintf("File berhasil disimpan sebagai '%s'!", safName)
				fv.IsError = false
				fv.SaveState()
			}
			fv.mu.Unlock()

			if fv.Invalidate != nil {
				fv.Invalidate()
			}
		}()
	}

	if fv.CreateSampleBtn.Clicked(gtx) {
		if hasUnsavedFile {
			fv.StatusMessage = "Perhatian: Simpan file hasil proses di bawah terlebih dahulu, atau klik 'BERSIHKAN'."
			fv.IsError = true
			return
		}
		sampleContent := "KripDroid Simulasi Berkas Media Video MP4\nWaktu: " + time.Now().Format(time.RFC3339) + "\nPengujian pemulihan otomatis format .mp4 dengan Argon2id."
		tmpSample, err := CreateSafeTempFile("sample_*.mp4")
		if err == nil {
			_, _ = tmpSample.WriteString(sampleContent)
			_ = tmpSample.Close()
			if fv.LoadedFilePath != "" {
				_ = os.Remove(fv.LoadedFilePath)
			}
			fv.LoadedFilePath = tmpSample.Name()
			fv.LoadedFileSize = int64(len(sampleContent))
			fv.LoadedFileName = "video_contoh.mp4"
			fv.SrcPathEditor.SetText("video_contoh.mp4 (108 byte)")
			fv.KeyEditor.SetText("KataSandi123!")
			fv.UseDeviceKey = false
			fv.ShowBiometricPrompt = false
			fv.ShowMasterPassModal = false
			fv.StatusMessage = "File contoh (video_contoh.mp4) siap dikunci/dienkripsi."
			fv.IsError = false
			fv.SaveState()
		}
	}

	if fv.ClearBtn.Clicked(gtx) {
		if fv.LoadedFilePath != "" {
			_ = os.Remove(fv.LoadedFilePath)
			fv.LoadedFilePath = ""
		}
		if fv.ProcessedFilePath != "" {
			_ = os.Remove(fv.ProcessedFilePath)
			fv.ProcessedFilePath = ""
		}
		fv.SrcPathEditor.SetText("")
		fv.DstPathEditor.SetText("")
		fv.LoadedFileName = ""
		fv.LoadedFileSize = 0
		fv.ProcessedFileName = ""
		fv.ProcessedFileSize = 0
		fv.RecoveredExt = ""
		fv.SavedLocation = ""
		fv.ShowBiometricPrompt = false
		fv.ShowMasterPassModal = false
		fv.StatusMessage = "Formulir telah direset. Anda dapat memilih file baru."
		fv.IsError = false
		fv.LastResult = nil
		fv.ClearState()
	}

	if fv.EncryptBtn.Clicked(gtx) && !fv.IsProcessing {
		if fv.LoadedFilePath == "" {
			fv.StatusMessage = "Peringatan: Belum ada file yang dipilih. Silakan klik tombol 'PILIH FILE' terlebih dahulu."
			fv.IsError = true
			return
		}

		if fv.UseDeviceKey {
			fv.PendingAction = "encrypt"
			fv.ShowBiometricPrompt = true
			fv.ShowMasterPassModal = false
			LaunchNativeAndroidBiometricPrompt("KripDroid Keamanan", "Konfirmasi sidik jari atau lockscreen untuk mengunci file")
			fv.StatusMessage = "Konfirmasi autentikasi biometrik atau Master Password."
			fv.IsError = false
			return
		}

		algoID := fv.SelectedAlgoID
		algo, _ := crypto.GetAlgorithm(algoID)
		key := fv.KeyEditor.Text()
		if algo != nil && algo.NeedsKey && strings.TrimSpace(key) == "" {
			fv.StatusMessage = "Peringatan: Kata sandi masih kosong. Masukkan kata sandi untuk mengunci file."
			fv.IsError = true
			return
		}

		fv.startFileEncryption(key)
	}

	if fv.DecryptBtn.Clicked(gtx) && !fv.IsProcessing {
		if fv.LoadedFilePath == "" {
			fv.StatusMessage = "Peringatan: Belum ada file yang dipilih. Silakan klik tombol 'PILIH FILE' terlebih dahulu."
			fv.IsError = true
			return
		}

		if fv.UseDeviceKey {
			fv.PendingAction = "decrypt"
			fv.ShowBiometricPrompt = true
			fv.ShowMasterPassModal = false
			LaunchNativeAndroidBiometricPrompt("KripDroid Keamanan", "Konfirmasi sidik jari atau lockscreen untuk membuka file")
			fv.StatusMessage = "Konfirmasi autentikasi biometrik atau Master Password."
			fv.IsError = false
			return
		}

		algoID := fv.SelectedAlgoID
		algo, _ := crypto.GetAlgorithm(algoID)
		key := fv.KeyEditor.Text()
		if algo != nil && algo.NeedsKey && strings.TrimSpace(key) == "" {
			fv.StatusMessage = "Peringatan: Masukkan kata sandi yang digunakan saat mengunci file."
			fv.IsError = true
			return
		}

		fv.startFileDecryption(key)
	}
}

func (fv *FileView) executeFileActionWithDeviceKey(action string) {
	hwKey, err := GetDeviceHardwareKey()
	if err != nil {
		fv.StatusMessage = fmt.Sprintf("Gagal mengakses Kunci Perangkat: %v", err)
		fv.IsError = true
		return
	}

	if action == "encrypt" {
		fv.startFileEncryption(hwKey)
	} else if action == "decrypt" {
		fv.startFileDecryption(hwKey)
	}
}

func (fv *FileView) executeFileActionWithMasterPass(action, masterPass string) {
	derivedKey, err := GetDeviceMasterKey(masterPass)
	if err != nil {
		fv.StatusMessage = fmt.Sprintf("Gagal memproses Master Password: %v", err)
		fv.IsError = true
		return
	}

	if action == "encrypt" {
		fv.startFileEncryption(derivedKey)
	} else if action == "decrypt" {
		fv.startFileDecryption(derivedKey)
	}
}

func (fv *FileView) startFileEncryption(effectiveKey string) {
	algoID := fv.SelectedAlgoID
	fv.IsProcessing = true
	fv.Progress = 0.1
	fv.StatusMessage = "Sedang mengunci file secara aman dengan Argon2id..."
	fv.IsError = false
	fv.LastResult = nil
	fv.RecoveredExt = ""
	fv.SavedLocation = ""

	srcPath := fv.LoadedFilePath
	baseName := fv.LoadedFileName
	if baseName == "" {
		baseName = "file_terkunci"
	}
	ext := filepath.Ext(baseName)

	tmpOut, err := CreateSafeTempFile("krip_enc_*")
	if err != nil {
		fv.IsProcessing = false
		fv.StatusMessage = fmt.Sprintf("Gagal membuat file sementara: %v", err)
		fv.IsError = true
		return
	}
	tmpOutPath := tmpOut.Name()
	_ = tmpOut.Close()

	go func() {
		res, err := crypto.EncryptFileWithExt(srcPath, tmpOutPath, algoID, effectiveKey, nil, ext, func(processed, total int64) {
			fv.mu.Lock()
			if total > 0 {
				fv.Progress = float32(processed) / float32(total)
			}
			fv.mu.Unlock()
			if fv.Invalidate != nil {
				fv.Invalidate()
			}
		})

		fv.mu.Lock()
		fv.IsProcessing = false
		if err != nil {
			_ = os.Remove(tmpOutPath)
			fv.StatusMessage = fmt.Sprintf("Gagal Mengunci File: %v", err)
			fv.IsError = true
			fv.Progress = 0
		} else {
			if fv.ProcessedFilePath != "" {
				_ = os.Remove(fv.ProcessedFilePath)
			}
			fv.Progress = 1.0
			fv.ProcessedFilePath = tmpOutPath
			fv.ProcessedFileSize = res.OutputSize
			fv.ProcessedFileName = baseName + ".krip"
			fv.LastResult = res
			if fv.UseDeviceKey {
				fv.StatusMessage = "File sukses dikunci dengan Kunci Perangkat Hardware + Argon2id!"
			} else if ext != "" {
				fv.StatusMessage = fmt.Sprintf("File sukses dikunci dengan Argon2id! Format '%s' aman tersimpan.", ext)
			} else {
				fv.StatusMessage = "File sukses dikunci dengan Argon2id! Silakan simpan file hasil di bawah."
			}
			fv.IsError = false
			fv.SaveState()
		}
		fv.mu.Unlock()

		if fv.Invalidate != nil {
			fv.Invalidate()
		}
	}()
}

func (fv *FileView) startFileDecryption(effectiveKey string) {
	algoID := fv.SelectedAlgoID
	fv.IsProcessing = true
	fv.Progress = 0.1
	fv.StatusMessage = "Sedang membuka kunci file dengan Argon2id..."
	fv.IsError = false
	fv.LastResult = nil
	fv.RecoveredExt = ""
	fv.SavedLocation = ""

	srcPath := fv.LoadedFilePath
	baseName := fv.LoadedFileName

	tmpOut, err := CreateSafeTempFile("krip_dec_*")
	if err != nil {
		fv.IsProcessing = false
		fv.StatusMessage = fmt.Sprintf("Gagal membuat file sementara: %v", err)
		fv.IsError = true
		return
	}
	tmpOutPath := tmpOut.Name()
	_ = tmpOut.Close()

	go func() {
		res, err := crypto.DecryptFile(srcPath, tmpOutPath, effectiveKey, algoID, func(processed, total int64) {
			fv.mu.Lock()
			if total > 0 {
				fv.Progress = float32(processed) / float32(total)
			}
			fv.mu.Unlock()
			if fv.Invalidate != nil {
				fv.Invalidate()
			}
		})

		fv.mu.Lock()
		fv.IsProcessing = false
		if err != nil {
			_ = os.Remove(tmpOutPath)
			if fv.UseDeviceKey {
				fv.StatusMessage = "Gagal Membuka File: Kunci Perangkat / Master Password tidak cocok atau file rusak."
			} else {
				fv.StatusMessage = "Gagal Membuka File: Kata sandi tidak cocok atau file rusak."
			}
			fv.IsError = true
			fv.Progress = 0
		} else {
			if fv.ProcessedFilePath != "" {
				_ = os.Remove(fv.ProcessedFilePath)
			}
			fv.Progress = 1.0
			fv.ProcessedFilePath = tmpOutPath
			fv.ProcessedFileSize = res.OutputSize
			fv.RecoveredExt = res.OriginalExt

			outName := baseName
			if strings.HasSuffix(outName, ".krip") {
				outName = strings.TrimSuffix(outName, ".krip")
			}
			if strings.HasSuffix(outName, ".bin") && res.OriginalExt != "" {
				outName = strings.TrimSuffix(outName, ".bin")
			}
			if res.OriginalExt != "" && !strings.HasSuffix(outName, res.OriginalExt) {
				outName = outName + res.OriginalExt
			}
			if outName == "" || outName == ".bin" {
				if res.OriginalExt != "" {
					outName = "file_terbuka" + res.OriginalExt
				} else {
					outName = "file_terbuka.bin"
				}
			}

			fv.ProcessedFileName = outName
			fv.LastResult = res
			if fv.UseDeviceKey {
				if res.OriginalExt != "" {
					fv.StatusMessage = fmt.Sprintf("Sukses dibuka! Format '%s' otomatis dipulihkan. Silakan simpan file di bawah.", res.OriginalExt)
				} else {
					fv.StatusMessage = "File sukses dibuka dengan Kunci Perangkat Android."
				}
			} else if res.OriginalExt != "" {
				fv.StatusMessage = fmt.Sprintf("Sukses dibuka! Format asli '%s' otomatis dipulihkan.", res.OriginalExt)
			} else {
				fv.StatusMessage = "File sukses dibuka! Silakan simpan file hasil di bawah."
			}
			fv.IsError = false
			fv.SaveState()
		}
		fv.mu.Unlock()

		if fv.Invalidate != nil {
			fv.Invalidate()
		}
	}()
}

func (fv *FileView) layoutAlgorithmSelector(gtx layout.Context, th *M3Theme) layout.Dimensions {
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
					return DrawM3Chip(gtx, th, &fv.CreateSampleBtn, "Coba Contoh Video", false)
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
				isSelected := fv.SelectedAlgoID == algoItem.ID

				btnChild := layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Right:  unit.Dp(3),
						Bottom: unit.Dp(4),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return DrawM3Chip(gtx, th, &fv.AlgoButtons[idx], algoItem.Name, isSelected)
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

func (fv *FileView) layoutAlgorithmInfoCard(gtx layout.Context, th *M3Theme, algo *crypto.AlgorithmInfo) layout.Dimensions {
	simpleDesc := algo.Description
	if algo.ID == "aes256gcm" || algo.ID == "chacha20poly1305" {
		simpleDesc = "Tingkat Tertinggi (Standar Militer + Argon2id). Sangat aman untuk dokumen rahasia, video, dan foto penting."
	} else if algo.Category == crypto.CategoryClassic {
		simpleDesc = "Sandi Sederhana (Untuk pembelajaran dasar kriptografi)."
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
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Right: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(th.MaterialTheme, unit.Sp(13), algo.Name)
								lbl.Color = th.OnSurface
								lbl.Font.Weight = font.Bold
								return lbl.Layout(gtx)
							})
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

func (fv *FileView) layoutExplorerPickSection(gtx layout.Context, th *M3Theme) layout.Dimensions {
	hasProcessed := fv.ProcessedFilePath != ""
	isAndroid := IsAndroidPlatform()

	sectionHeader := "PILIH FILE DARI KOMPUTER"
	pickBtnText := "PILIH FILE DARI FILE EXPLORER..."
	helperDesc := "Pilih dokumen, video, foto, atau file apapun dari komputer/laptop Anda."
	if isAndroid {
		sectionHeader = "PILIH FILE DARI HP"
		pickBtnText = "PILIH FILE DARI HP ANDROID..."
		helperDesc = "Pilih video, dokumen, foto, atau file apapun dari memori HP Anda."
	}

	return DrawOutlinedCard(gtx, th.SurfaceContainer, th.OutlineVariant, unit.Dp(12), unit.Dp(1), func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(12),
			Bottom: unit.Dp(12),
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
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(th.MaterialTheme, unit.Sp(12), sectionHeader)
							lbl.Color = th.Primary
							lbl.Font.Weight = font.Bold
							return lbl.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if fv.RecoveredExt != "" {
								return DrawPill(gtx, th.Success, th.OnSuccess, "Format: "+fv.RecoveredExt, unit.Sp(10), th)
							}
							return layout.Dimensions{}
						}),
					)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if hasProcessed {
						return DrawCard(gtx, th.SurfaceContainerHigh, unit.Dp(10), func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Top:    unit.Dp(8),
								Bottom: unit.Dp(8),
								Left:   unit.Dp(10),
								Right:  unit.Dp(10),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(th.MaterialTheme, unit.Sp(11), "Tombol Pilih File dikunci sementara. Simpan file hasil di bawah terlebih dahulu, atau klik BERSIHKAN.")
								lbl.Color = th.OnSurfaceVariant
								return lbl.Layout(gtx)
							})
						})
					}
					return DrawM3Button(gtx, th, &fv.PickFileBtn, pickBtnText, th.Primary, th.OnPrimary, unit.Dp(12))
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if hasProcessed {
						if isAndroid {
							return layout.Flex{
								Axis: layout.Vertical,
							}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return DrawM3Button(gtx, th, &fv.SaveFileBtn, "SIMPAN FILE (PILIH FOLDER HP)", th.Secondary, th.OnSecondary, unit.Dp(12))
								}),
								layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return DrawM3Button(gtx, th, &fv.CopyBase64Btn, "SALIN SEBAGAI TEKS (BASE64)", th.SurfaceContainerHigh, th.Tertiary, unit.Dp(10))
								}),
							)
						} else {
							return layout.Flex{
								Axis: layout.Vertical,
							}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return DrawM3Button(gtx, th, &fv.QuickSaveBtn, "SIMPAN KE FOLDER DOWNLOADS", th.Secondary, th.OnSecondary, unit.Dp(12))
								}),
								layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{
										Axis: layout.Horizontal,
									}.Layout(gtx,
										layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
											return DrawM3Button(gtx, th, &fv.SaveFileBtn, "SIMPAN KE FOLDER LAIN", th.SurfaceContainerHigh, th.Primary, unit.Dp(10))
										}),
										layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
										layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
											return DrawM3Button(gtx, th, &fv.CopyBase64Btn, "SALIN BASE64", th.SurfaceContainerHigh, th.Tertiary, unit.Dp(10))
										}),
									)
								}),
							)
						}
					}
					return layout.Dimensions{}
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if fv.LoadedFilePath != "" {
						return DrawCard(gtx, th.SurfaceContainerHigh, unit.Dp(8), func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Top:    unit.Dp(6),
								Bottom: unit.Dp(6),
								Left:   unit.Dp(10),
								Right:  unit.Dp(10),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(th.MaterialTheme, unit.Sp(11), fmt.Sprintf("File Terpilih: %s (%d byte)", fv.LoadedFileName, fv.LoadedFileSize))
								lbl.Color = th.Secondary
								lbl.Font.Weight = font.SemiBold
								return lbl.Layout(gtx)
							})
						})
					}
					lbl := material.Label(th.MaterialTheme, unit.Sp(11), helperDesc)
					lbl.Color = th.OnSurfaceVariant
					return lbl.Layout(gtx)
				}),
			)
		})
	})
}

func (fv *FileView) layoutKeySection(gtx layout.Context, th *M3Theme, algo *crypto.AlgorithmInfo) layout.Dimensions {
	if algo != nil && !algo.NeedsKey {
		return layout.Dimensions{}
	}

	isAndroid := IsKeyStoreSupported()

	if fv.UseDeviceKey {
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
						lbl := material.Label(th.MaterialTheme, unit.Sp(11), "File diamankan hardware KeyStore & Lockscreen HP. Memerlukan Biometrik atau Master Password.")
						lbl.Color = th.OnSecondaryContainer
						return lbl.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return DrawM3Button(gtx, th, &fv.SwitchToCustomPassBtn, "GANTI KE KATA SANDI MANUAL", th.SurfaceContainerHigh, th.OnSurface, unit.Dp(10))
					}),
				)
			})
		})
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
					return DrawM3Chip(gtx, th, &fv.ClearBtn, "Bersihkan", false)
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
					ed := material.Editor(th.MaterialTheme, &fv.KeyEditor, "Ketik kata sandi rahasia Anda di sini...")
					ed.Color = th.OnSurface
					ed.HintColor = th.Outline
					return ed.Layout(gtx)
				})
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if isAndroid {
				return layout.Inset{Top: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return DrawM3Button(gtx, th, &fv.KeyStoreBtn, "GUNAKAN KUNCI PERANGKAT (KEYSTORE HP)", th.SecondaryContainer, th.OnSecondaryContainer, unit.Dp(10))
				})
			}
			return layout.Dimensions{}
		}),
	)
}

func (fv *FileView) layoutBiometricModalContent(gtx layout.Context, th *M3Theme) layout.Dimensions {
	actionName := "Mengunci File"
	if fv.PendingAction == "decrypt" {
		actionName = "Membuka File"
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
			return DrawM3Button(gtx, th, &fv.ConfirmBiometricBtn, "VERIFIKASI BIOMETRIK / LOCKSCREEN", th.Secondary, th.OnSecondary, unit.Dp(10))
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return DrawM3Button(gtx, th, &fv.OpenMasterPassBtn, "GUNAKAN MASTER PASSWORD", th.SurfaceContainer, th.OnSurface, unit.Dp(10))
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return DrawM3Button(gtx, th, &fv.CancelBiometricBtn, "BATAL", th.SurfaceContainer, th.Error, unit.Dp(10))
		}),
	)
}

func (fv *FileView) layoutMasterPassModalContent(gtx layout.Context, th *M3Theme) layout.Dimensions {
	hasMaster := HasMasterPassword()

	if fv.ShowChangeMasterPass {
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
						ed := material.Editor(th.MaterialTheme, &fv.OldMasterPassEditor, "Master Password Lama...")
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
						ed := material.Editor(th.MaterialTheme, &fv.NewMasterPassEditor, "Master Password Baru...")
						ed.Color = th.OnSurface
						ed.HintColor = th.Outline
						return ed.Layout(gtx)
					})
				})
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DrawM3Button(gtx, th, &fv.SubmitChangeMasterPassBt, "SIMPAN PASSWORD BARU", th.Primary, th.OnPrimary, unit.Dp(10))
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DrawM3Button(gtx, th, &fv.CancelChangeMasterPassBt, "KEMBALI", th.SurfaceContainer, th.OnSurfaceVariant, unit.Dp(10))
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
						return DrawM3Chip(gtx, th, &fv.OpenChangeMasterPassBtn, "Ganti Sandi", false)
					}
					return layout.Dimensions{}
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return DrawOutlinedCard(gtx, th.SurfaceContainer, th.OutlineVariant, unit.Dp(8), unit.Dp(1), func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8), Left: unit.Dp(10), Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					ed := material.Editor(th.MaterialTheme, &fv.MasterPassEditor, hint)
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
							ed := material.Editor(th.MaterialTheme, &fv.ConfirmMasterPassEditor, "Konfirmasi Master Password baru...")
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
			return DrawM3Button(gtx, th, &fv.SubmitMasterPassBtn, "KONFIRMASI MASTER PASSWORD", th.Secondary, th.OnSecondary, unit.Dp(10))
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return DrawM3Button(gtx, th, &fv.CancelMasterPassBtn, "BATAL", th.SurfaceContainer, th.Error, unit.Dp(10))
		}),
	)
}

func (fv *FileView) layoutActionButtons(gtx layout.Context, th *M3Theme) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Horizontal,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			label := "KUNCI / ENKRIPSI FILE"
			if fv.IsProcessing {
				label = "SEDANG MEMPROSES..."
			}
			return DrawM3Button(gtx, th, &fv.EncryptBtn, label, th.Primary, th.OnPrimary, unit.Dp(12))
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			label := "BUKA / DEKRIPSI FILE"
			if fv.IsProcessing {
				label = "SEDANG MEMPROSES..."
			}
			return DrawM3Button(gtx, th, &fv.DecryptBtn, label, th.Secondary, th.OnSecondary, unit.Dp(12))
		}),
	)
}

func (fv *FileView) layoutProgressSection(gtx layout.Context, th *M3Theme) layout.Dimensions {
	if !fv.IsProcessing && fv.Progress <= 0 && fv.StatusMessage == "" {
		return layout.Dimensions{}
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if fv.IsProcessing || fv.Progress > 0 {
				bar := material.ProgressBar(th.MaterialTheme, fv.Progress)
				bar.Color = th.Primary
				bar.TrackColor = th.SurfaceContainerHighest
				return bar.Layout(gtx)
			}
			return layout.Dimensions{}
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if fv.StatusMessage == "" {
				return layout.Dimensions{}
			}
			var bg, fg color.NRGBA
			if fv.IsError {
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
					lbl := material.Label(th.MaterialTheme, unit.Sp(12), fv.StatusMessage)
					lbl.Color = fg
					lbl.Font.Weight = font.Medium
					return lbl.Layout(gtx)
				})
			})
		}),
	)
}

func (fv *FileView) layoutResultSection(gtx layout.Context, th *M3Theme) layout.Dimensions {
	if fv.LastResult == nil {
		return layout.Dimensions{}
	}

	res := fv.LastResult
	return DrawOutlinedCard(gtx, th.SurfaceContainer, th.OutlineVariant, unit.Dp(12), unit.Dp(1), func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(12),
			Bottom: unit.Dp(12),
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
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(th.MaterialTheme, unit.Sp(13), "FILE SUKSES DIPROSES")
							lbl.Color = th.Secondary
							lbl.Font.Weight = font.Bold
							return lbl.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if res.OriginalExt != "" {
								return DrawPill(gtx, th.PrimaryContainer, th.OnPrimaryContainer, "Format: "+res.OriginalExt, unit.Sp(10), th)
							}
							return layout.Dimensions{}
						}),
					)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.MaterialTheme, unit.Sp(11), fmt.Sprintf("Nama File: %s", fv.ProcessedFileName))
					lbl.Color = th.OnSurface
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(2)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if res.OriginalExt != "" {
						lbl := material.Label(th.MaterialTheme, unit.Sp(11), fmt.Sprintf("Format Asli File: %s", res.OriginalExt))
						lbl.Color = th.Primary
						lbl.Font.Weight = font.SemiBold
						return lbl.Layout(gtx)
					}
					return layout.Dimensions{}
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(2)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.MaterialTheme, unit.Sp(11), fmt.Sprintf("Ukuran Asal: %d byte  |  Ukuran Hasil: %d byte", res.InputSize, res.OutputSize))
					lbl.Color = th.OnSurfaceVariant
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(2)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if fv.SavedLocation != "" {
						return DrawCard(gtx, th.SecondaryContainer, unit.Dp(8), func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Top:    unit.Dp(6),
								Bottom: unit.Dp(6),
								Left:   unit.Dp(10),
								Right:  unit.Dp(10),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(th.MaterialTheme, unit.Sp(11), fmt.Sprintf("Tersimpan di: %s", fv.SavedLocation))
								lbl.Color = th.OnSecondaryContainer
								lbl.Font.Weight = font.SemiBold
								return lbl.Layout(gtx)
							})
						})
					}
					return layout.Dimensions{}
				}),
			)
		})
	})
}
