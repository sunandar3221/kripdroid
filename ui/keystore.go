package ui

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/argon2"
)

var keyStoreAlias = "KripDroidMasterKey"
var DeviceKeySentinel = "[KUNCI-PERANGKAT-ANDROID-KEYSTORE]"

type MasterPassStore struct {
	Salt     string `json:"salt"`
	Hash     string `json:"hash"`
	SeedSalt string `json:"seed_salt"`
}

func GetMasterPassFilePath() string {
	dir := GetSafeTempDir()
	return filepath.Join(dir, ".krip_masterpass.json")
}

func HasMasterPassword() bool {
	path := GetMasterPassFilePath()
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return false
	}
	var store MasterPassStore
	if err := json.Unmarshal(data, &store); err != nil {
		return false
	}
	return store.Hash != "" && store.Salt != ""
}

func hashMasterPassword(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, 2, 64*1024, 4, 32)
}

func SetMasterPassword(password string) error {
	trimmed := strings.TrimSpace(password)
	if len(trimmed) < 4 {
		return errors.New("kata sandi master minimal 4 karakter")
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		salt = []byte("KripDroidSalt2026")
	}

	seedSalt := make([]byte, 16)
	if _, err := rand.Read(seedSalt); err != nil {
		seedSalt = []byte("KripDroidSeed2026")
	}

	hash := hashMasterPassword(trimmed, salt)

	store := MasterPassStore{
		Salt:     hex.EncodeToString(salt),
		Hash:     hex.EncodeToString(hash),
		SeedSalt: hex.EncodeToString(seedSalt),
	}

	data, err := json.Marshal(store)
	if err != nil {
		return err
	}

	return os.WriteFile(GetMasterPassFilePath(), data, 0600)
}

func VerifyMasterPassword(password string) bool {
	path := GetMasterPassFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	var store MasterPassStore
	if err := json.Unmarshal(data, &store); err != nil {
		return false
	}

	salt, err := hex.DecodeString(store.Salt)
	if err != nil {
		return false
	}

	expectedHash, err := hex.DecodeString(store.Hash)
	if err != nil {
		return false
	}

	actualHash := hashMasterPassword(strings.TrimSpace(password), salt)
	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1
}

func ChangeMasterPassword(oldPass, newPass string) error {
	if !VerifyMasterPassword(oldPass) {
		return errors.New("master password lama salah")
	}
	trimmedNew := strings.TrimSpace(newPass)
	if len(trimmedNew) < 4 {
		return errors.New("master password baru minimal 4 karakter")
	}
	return SetMasterPassword(trimmedNew)
}

func GetDeviceMasterKey(masterPass string) (string, error) {
	if !VerifyMasterPassword(masterPass) {
		return "", errors.New("master password salah atau belum diatur")
	}

	path := GetMasterPassFilePath()
	data, _ := os.ReadFile(path)
	var store MasterPassStore
	_ = json.Unmarshal(data, &store)

	seedSalt, _ := hex.DecodeString(store.SeedSalt)
	derived := argon2.IDKey([]byte(masterPass), seedSalt, 2, 64*1024, 4, 32)
	return hex.EncodeToString(derived), nil
}

func IsKeyStoreSupported() bool {
	return IsAndroidPlatform()
}

func IsDeviceKey(keyStr string) bool {
	trimmed := strings.TrimSpace(keyStr)
	return trimmed == DeviceKeySentinel || strings.HasPrefix(trimmed, "[KUNCI-PERANGKAT")
}

func LaunchNativeAndroidBiometricPrompt(title, desc string) {
	// The native AM start command is non-blocking and deprecated.
	// We handle this via simulated in-app modal instead.
}

func GetDeviceHardwareKey() (string, error) {
	if !IsKeyStoreSupported() {
		return "", fmt.Errorf("android keystore hanya tersedia pada perangkat android")
	}

	dir := GetSafeTempDir()
	keyFile := filepath.Join(dir, ".krip_keystore_seed")
	if data, err := os.ReadFile(keyFile); err == nil && len(data) > 0 {
		return string(data), nil
	}

	seedRaw := fmt.Sprintf("AndroidKeyStore_%s_%s", keyStoreAlias, dir)
	h := sha256.Sum256([]byte(seedRaw))
	keyHex := hex.EncodeToString(h[:])

	_ = os.WriteFile(keyFile, []byte(keyHex), 0600)
	return keyHex, nil
}

func ResolveEffectiveKey(keyStr string) (string, error) {
	if IsDeviceKey(keyStr) {
		if !IsKeyStoreSupported() {
			return "", fmt.Errorf("kunci perangkat hanya didukung pada android")
		}
		hwKey, err := GetDeviceHardwareKey()
		if err != nil {
			return "", err
		}
		return hwKey, nil
	}
	return keyStr, nil
}
