package crypto

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestAllAlgorithmsText(t *testing.T) {
	testKey := "SuperSecretKey123!@#"
	samplePlain := "Hello, KripDroid Security World! 12345 ABC xyz."

	for _, algo := range AvailableAlgorithms {
		t.Run(algo.ID, func(t *testing.T) {
			enc, iv, err := EncryptText(algo.ID, samplePlain, testKey, "")
			if err != nil {
				t.Fatalf("encryption failed for %s: %v", algo.ID, err)
			}
			if len(enc) == 0 {
				t.Fatalf("encrypted output is empty for %s", algo.ID)
			}

			dec, err := DecryptText(algo.ID, enc, testKey, "")
			if err != nil {
				t.Fatalf("decryption failed for %s: %v (iv=%s)", algo.ID, err, iv)
			}
			if dec != samplePlain {
				t.Fatalf("decrypted mismatch for %s: expected %q, got %q", algo.ID, samplePlain, dec)
			}
		})
	}
}

func TestClassicSpecifics(t *testing.T) {
	cEnc, err := CaesarEncrypt("Hello World", "3")
	if err != nil {
		t.Fatal(err)
	}
	if cEnc != "Khoor Zruog" {
		t.Fatalf("unexpected caesar enc: %s", cEnc)
	}
	cDec, err := CaesarDecrypt(cEnc, "3")
	if err != nil {
		t.Fatal(err)
	}
	if cDec != "Hello World" {
		t.Fatalf("unexpected caesar dec: %s", cDec)
	}

	rEnc := Rot13("Hello World")
	if rEnc != "Uryyb Jbeyq" {
		t.Fatalf("unexpected rot13 enc: %s", rEnc)
	}
	rDec := Rot13(rEnc)
	if rDec != "Hello World" {
		t.Fatalf("unexpected rot13 dec: %s", rDec)
	}

	aEnc := Atbash("Hello World")
	if aEnc != "Svool Dliow" {
		t.Fatalf("unexpected atbash enc: %s", aEnc)
	}
	aDec := Atbash(aEnc)
	if aDec != "Hello World" {
		t.Fatalf("unexpected atbash dec: %s", aDec)
	}

	vEnc, err := VigenereEncrypt("ATTACKATDAWN", "LEMON")
	if err != nil {
		t.Fatal(err)
	}
	if vEnc != "LXFOPVEFRNHR" {
		t.Fatalf("unexpected vigenere enc: %s", vEnc)
	}
	vDec, err := VigenereDecrypt(vEnc, "LEMON")
	if err != nil {
		t.Fatal(err)
	}
	if vDec != "ATTACKATDAWN" {
		t.Fatalf("unexpected vigenere dec: %s", vDec)
	}
}

func TestFileEncryptionDecryption(t *testing.T) {
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "video_presentation.mp4")
	encPath := filepath.Join(tempDir, "video_presentation.mp4.krip")
	decPath := filepath.Join(tempDir, "video_recovered.mp4")

	originalData := []byte("Video MP4 stream binary simulated payload test data.")
	if err := os.WriteFile(srcPath, originalData, 0644); err != nil {
		t.Fatal(err)
	}

	testAlgos := []string{"aes256gcm", "chacha20poly1305", "aes128cbc", "blowfish", "tripledes", "des", "xor"}
	for _, algoID := range testAlgos {
		t.Run(algoID, func(t *testing.T) {
			resEnc, err := EncryptFile(srcPath, encPath, algoID, "MasterPassphrase!123", nil, nil)
			if err != nil {
				t.Fatalf("file encryption failed: %v", err)
			}
			if resEnc.OutputSize == 0 {
				t.Fatal("encrypted file is empty")
			}
			if resEnc.OriginalExt != ".mp4" {
				t.Fatalf("expected ext .mp4, got %s", resEnc.OriginalExt)
			}

			resDec, err := DecryptFile(encPath, decPath, "MasterPassphrase!123", "", nil)
			if err != nil {
				t.Fatalf("file decryption failed: %v", err)
			}
			if resDec.OriginalExt != ".mp4" {
				t.Fatalf("expected recovered ext .mp4, got %s", resDec.OriginalExt)
			}
			if resDec.OutputSize != int64(len(originalData)) {
				t.Fatalf("file size mismatch: expected %d, got %d", len(originalData), resDec.OutputSize)
			}

			decData, err := os.ReadFile(decPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(decData, originalData) {
				t.Fatalf("recovered file content mismatch for %s", algoID)
			}
		})
	}
}

func TestFileDataExtensionPreservation(t *testing.T) {
	plain := []byte("secret document content")
	encBytes, err := EncryptFileData(plain, "aes256gcm", "key123", nil, ".pdf")
	if err != nil {
		t.Fatal(err)
	}

	decBytes, algo, ext, err := DecryptFileData(encBytes, "key123", "")
	if err != nil {
		t.Fatal(err)
	}
	if algo != "aes256gcm" {
		t.Fatalf("expected algo aes256gcm, got %s", algo)
	}
	if ext != ".pdf" {
		t.Fatalf("expected ext .pdf, got %s", ext)
	}
	if !bytes.Equal(decBytes, plain) {
		t.Fatal("plain content mismatch")
	}
}

func TestTempFileEncryptionWithOriginalExt(t *testing.T) {
	tempDir := t.TempDir()
	srcTemp := filepath.Join(tempDir, "krip_in_987654")
	encTemp := filepath.Join(tempDir, "krip_out_987654")
	decTemp := filepath.Join(tempDir, "krip_dec_987654")

	data := []byte("Audio MP3 music test stream binary payload")
	if err := os.WriteFile(srcTemp, data, 0644); err != nil {
		t.Fatal(err)
	}

	resEnc, err := EncryptFileWithExt(srcTemp, encTemp, "aes256gcm", "SecretKey123!", nil, ".mp3", nil)
	if err != nil {
		t.Fatal(err)
	}
	if resEnc.OriginalExt != ".mp3" {
		t.Fatalf("expected .mp3, got %s", resEnc.OriginalExt)
	}

	resDec, err := DecryptFile(encTemp, decTemp, "SecretKey123!", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if resDec.OriginalExt != ".mp3" {
		t.Fatalf("expected recovered .mp3, got %s", resDec.OriginalExt)
	}

	decData, err := os.ReadFile(decTemp)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decData, data) {
		t.Fatal("recovered data mismatch")
	}
}
