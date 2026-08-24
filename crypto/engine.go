package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type SecurityCategory string

const (
	CategoryClassic SecurityCategory = "Klasik"
	CategoryLegacy  SecurityCategory = "Legacy"
	CategoryModern  SecurityCategory = "Modern"
)

type AlgorithmInfo struct {
	ID             string
	Name           string
	Category       SecurityCategory
	SecurityLevel  string
	SecurityRating int
	KeyDescription string
	NeedsKey       bool
	NeedsIV        bool
	KeySize        int
	IVSize         int
	Description    string
}

var AvailableAlgorithms = []AlgorithmInfo{
	{
		ID:             "caesar",
		Name:           "Caesar Cipher",
		Category:       CategoryClassic,
		SecurityLevel:  "Lemah / Edukasi",
		SecurityRating: 1,
		KeyDescription: "Angka Pergeseran (cth: 3) atau Kata Sandi",
		NeedsKey:       true,
		NeedsIV:        false,
		KeySize:        0,
		IVSize:         0,
		Description:    "Cipher substitusi monoalfabetik dengan menggeser huruf sesuai nilai kunci integer.",
	},
	{
		ID:             "rot13",
		Name:           "ROT13",
		Category:       CategoryClassic,
		SecurityLevel:  "Lemah / Edukasi",
		SecurityRating: 1,
		KeyDescription: "Tanpa Kunci (Pergeseran Tetap 13 Posisi)",
		NeedsKey:       false,
		NeedsIV:        false,
		KeySize:        0,
		IVSize:         0,
		Description:    "Varian khusus Caesar Cipher dengan rotasi alfabet tetap sebesar 13 langkah.",
	},
	{
		ID:             "atbash",
		Name:           "Atbash Cipher",
		Category:       CategoryClassic,
		SecurityLevel:  "Lemah / Edukasi",
		SecurityRating: 1,
		KeyDescription: "Tanpa Kunci (Pembalikan Urutan Alfabet)",
		NeedsKey:       false,
		NeedsIV:        false,
		KeySize:        0,
		IVSize:         0,
		Description:    "Cipher substitusi kuno yang membalik susunan alfabet (A-Z, B-Y, C-X).",
	},
	{
		ID:             "vigenere",
		Name:           "Vigenere Cipher",
		Category:       CategoryClassic,
		SecurityLevel:  "Lemah / Edukasi",
		SecurityRating: 2,
		KeyDescription: "Kata Kunci / Frasa (cth: RAHASIA)",
		NeedsKey:       true,
		NeedsIV:        false,
		KeySize:        0,
		IVSize:         0,
		Description:    "Cipher substitusi polialfabetik menggunakan tabel pergeseran kata kunci berulang.",
	},
	{
		ID:             "xor",
		Name:           "Simple XOR Cipher",
		Category:       CategoryClassic,
		SecurityLevel:  "Lemah / Edukasi",
		SecurityRating: 2,
		KeyDescription: "Kunci / Kata Sandi Bebas",
		NeedsKey:       true,
		NeedsIV:        false,
		KeySize:        0,
		IVSize:         0,
		Description:    "Operasi bitwise Exclusive-OR (XOR) biner cepat dengan pengulangan kunci.",
	},
	{
		ID:             "des",
		Name:           "DES (Data Encryption Standard)",
		Category:       CategoryLegacy,
		SecurityLevel:  "Menengah / Legacy",
		SecurityRating: 3,
		KeyDescription: "Kunci 56-bit (8 karakter / Kata Sandi)",
		NeedsKey:       true,
		NeedsIV:        true,
		KeySize:        8,
		IVSize:         8,
		Description:    "Cipher blok 64-bit simetris dalam mode CBC dengan bantalan PKCS7.",
	},
	{
		ID:             "tripledes",
		Name:           "Triple DES (3DES / EDE)",
		Category:       CategoryLegacy,
		SecurityLevel:  "Menengah / Legacy",
		SecurityRating: 4,
		KeyDescription: "Kunci 168-bit (24 karakter / Kata Sandi)",
		NeedsKey:       true,
		NeedsIV:        true,
		KeySize:        24,
		IVSize:         8,
		Description:    "Penerapan algoritma DES tiga kali berurutan per blok data dalam mode CBC.",
	},
	{
		ID:             "blowfish",
		Name:           "Blowfish",
		Category:       CategoryLegacy,
		SecurityLevel:  "Menengah / Legacy",
		SecurityRating: 4,
		KeyDescription: "Kunci Variabel 32-448 bit (Kata Sandi)",
		NeedsKey:       true,
		NeedsIV:        true,
		KeySize:        16,
		IVSize:         8,
		Description:    "Cipher blok 64-bit berkecepatan tinggi rancangan Bruce Schneier.",
	},
	{
		ID:             "aes128cbc",
		Name:           "AES-128 (Mode CBC)",
		Category:       CategoryModern,
		SecurityLevel:  "Standar Keamanan Tinggi",
		SecurityRating: 5,
		KeyDescription: "Kunci 128-bit (16 karakter / Kata Sandi)",
		NeedsKey:       true,
		NeedsIV:        true,
		KeySize:        16,
		IVSize:         16,
		Description:    "Advanced Encryption Standard dengan kunci 128-bit dalam mode Cipher Block Chaining.",
	},
	{
		ID:             "aes128gcm",
		Name:           "AES-128 (GCM AEAD)",
		Category:       CategoryModern,
		SecurityLevel:  "Standar Keamanan Tinggi",
		SecurityRating: 5,
		KeyDescription: "Kunci 128-bit (16 karakter / Kata Sandi)",
		NeedsKey:       true,
		NeedsIV:        true,
		KeySize:        16,
		IVSize:         12,
		Description:    "Mode Galois/Counter dengan otentikasi data (AEAD) dan throughput sangat tinggi.",
	},
	{
		ID:             "aes256cbc",
		Name:           "AES-256 (Mode CBC)",
		Category:       CategoryModern,
		SecurityLevel:  "Standar Militer & Rahasia",
		SecurityRating: 5,
		KeyDescription: "Kunci 256-bit (32 karakter / Kata Sandi)",
		NeedsKey:       true,
		NeedsIV:        true,
		KeySize:        32,
		IVSize:         16,
		Description:    "Cipher blok 256-bit standar industri dengan ketahanan maksimal terhadap brute-force.",
	},
	{
		ID:             "aes256gcm",
		Name:           "AES-256 (GCM AEAD)",
		Category:       CategoryModern,
		SecurityLevel:  "Standar Militer & Rahasia",
		SecurityRating: 5,
		KeyDescription: "Kunci 256-bit (32 karakter / Kata Sandi)",
		NeedsKey:       true,
		NeedsIV:        true,
		KeySize:        32,
		IVSize:         12,
		Description:    "Enkripsi terotentikasi 256-bit AEAD menjamin kerahasiaan dan integritas data penuh.",
	},
	{
		ID:             "chacha20poly1305",
		Name:           "ChaCha20-Poly1305",
		Category:       CategoryModern,
		SecurityLevel:  "Standar Militer & TLS 1.3",
		SecurityRating: 5,
		KeyDescription: "Kunci 256-bit (32 karakter / Kata Sandi)",
		NeedsKey:       true,
		NeedsIV:        true,
		KeySize:        32,
		IVSize:         12,
		Description:    "Stream cipher modern berkecepatan tinggi dengan autentikator Poly1305 (RFC 8439).",
	},
}

func GetAlgorithm(id string) (*AlgorithmInfo, error) {
	for i := range AvailableAlgorithms {
		if AvailableAlgorithms[i].ID == id {
			return &AvailableAlgorithms[i], nil
		}
	}
	return nil, fmt.Errorf("algoritma tidak dikenal: %s", id)
}

func DeriveKeyArgon2id(passphrase string, salt []byte, targetBytes int) []byte {
	if targetBytes <= 0 {
		return []byte(passphrase)
	}
	if len(salt) == 0 {
		salt = []byte("KripDroidArgon2idSalt2026")
	}
	return argon2.IDKey([]byte(passphrase), salt, 2, 64*1024, 4, uint32(targetBytes))
}

func DeriveKey(passphrase string, targetBytes int) []byte {
	if targetBytes <= 0 {
		return []byte(passphrase)
	}
	trimmed := strings.TrimSpace(passphrase)
	if len(trimmed) == targetBytes*2 {
		if hx, err := hex.DecodeString(trimmed); err == nil && len(hx) == targetBytes {
			return hx
		}
	}
	if len(passphrase) == targetBytes {
		return []byte(passphrase)
	}
	if targetBytes <= 32 {
		hash := sha256.Sum256([]byte(passphrase))
		return hash[:targetBytes]
	}
	hash := sha512.Sum512([]byte(passphrase))
	return hash[:targetBytes]
}

func ParseIV(ivStr string, targetBytes int) ([]byte, error) {
	if targetBytes <= 0 {
		return nil, nil
	}
	trimmed := strings.TrimSpace(ivStr)
	if trimmed == "" {
		return nil, nil
	}
	if hx, err := hex.DecodeString(trimmed); err == nil && len(hx) == targetBytes {
		return hx, nil
	}
	if len(trimmed) == targetBytes {
		return []byte(trimmed), nil
	}
	hash := sha256.Sum256([]byte(trimmed))
	return hash[:targetBytes], nil
}

func EncryptBytesWithSalt(algoID string, plain []byte, keyStr string, customIV []byte, salt []byte) ([]byte, []byte, error) {
	algo, err := GetAlgorithm(algoID)
	if err != nil {
		return nil, nil, err
	}
	if algo.NeedsKey && strings.TrimSpace(keyStr) == "" {
		return nil, nil, errors.New("kunci rahasia wajib diisi untuk algoritma ini")
	}

	var key []byte
	if algo.NeedsKey {
		if len(salt) > 0 {
			key = DeriveKeyArgon2id(keyStr, salt, algo.KeySize)
		} else {
			key = DeriveKey(keyStr, algo.KeySize)
		}
	}

	switch algoID {
	case "caesar":
		res, err := CaesarBytes(plain, keyStr, false)
		return res, nil, err
	case "rot13":
		return Rot13Bytes(plain), nil, nil
	case "atbash":
		return AtbashBytes(plain), nil, nil
	case "vigenere":
		res, err := VigenereBytes(plain, keyStr, false)
		return res, nil, err
	case "xor":
		res, err := XorCipher(plain, []byte(keyStr))
		return res, nil, err
	case "des":
		return DesEncrypt(plain, key, customIV)
	case "tripledes":
		return TripleDesEncrypt(plain, key, customIV)
	case "blowfish":
		return BlowfishEncrypt(plain, key, customIV)
	case "aes128cbc":
		return AesCbcEncrypt(plain, key, customIV)
	case "aes128gcm":
		return AesGcmEncrypt(plain, key, customIV)
	case "aes256cbc":
		return AesCbcEncrypt(plain, key, customIV)
	case "aes256gcm":
		return AesGcmEncrypt(plain, key, customIV)
	case "chacha20poly1305":
		return ChaCha20Poly1305Encrypt(plain, key, customIV)
	default:
		return nil, nil, fmt.Errorf("algoritma belum didukung: %s", algoID)
	}
}

func EncryptBytes(algoID string, plain []byte, keyStr string, customIV []byte) ([]byte, []byte, error) {
	return EncryptBytesWithSalt(algoID, plain, keyStr, customIV, nil)
}

func DecryptBytesWithSalt(algoID string, cipherData []byte, keyStr string, iv []byte, salt []byte) ([]byte, error) {
	algo, err := GetAlgorithm(algoID)
	if err != nil {
		return nil, err
	}
	if algo.NeedsKey && strings.TrimSpace(keyStr) == "" {
		return nil, errors.New("kunci rahasia wajib diisi untuk algoritma ini")
	}

	var key []byte
	if algo.NeedsKey {
		if len(salt) > 0 {
			key = DeriveKeyArgon2id(keyStr, salt, algo.KeySize)
		} else {
			key = DeriveKey(keyStr, algo.KeySize)
		}
	}

	switch algoID {
	case "caesar":
		return CaesarBytes(cipherData, keyStr, true)
	case "rot13":
		return Rot13Bytes(cipherData), nil
	case "atbash":
		return AtbashBytes(cipherData), nil
	case "vigenere":
		return VigenereBytes(cipherData, keyStr, true)
	case "xor":
		return XorCipher(cipherData, []byte(keyStr))
	case "des":
		return DesDecrypt(cipherData, key, iv)
	case "tripledes":
		return TripleDesDecrypt(cipherData, key, iv)
	case "blowfish":
		return BlowfishDecrypt(cipherData, key, iv)
	case "aes128cbc":
		return AesCbcDecrypt(cipherData, key, iv)
	case "aes128gcm":
		return AesGcmDecrypt(cipherData, key, iv)
	case "aes256cbc":
		return AesCbcDecrypt(cipherData, key, iv)
	case "aes256gcm":
		return AesGcmDecrypt(cipherData, key, iv)
	case "chacha20poly1305":
		return ChaCha20Poly1305Decrypt(cipherData, key, iv)
	default:
		return nil, fmt.Errorf("algoritma belum didukung: %s", algoID)
	}
}

func DecryptBytes(algoID string, cipherData []byte, keyStr string, iv []byte) ([]byte, error) {
	return DecryptBytesWithSalt(algoID, cipherData, keyStr, iv, nil)
}

func EncryptText(algoID string, plainText string, keyStr string, ivStr string) (string, string, error) {
	algo, err := GetAlgorithm(algoID)
	if err != nil {
		return "", "", err
	}
	if algo.Category == CategoryClassic && algoID != "xor" {
		switch algoID {
		case "caesar":
			res, err := CaesarEncrypt(plainText, keyStr)
			return res, "", err
		case "rot13":
			return Rot13(plainText), "", nil
		case "atbash":
			return Atbash(plainText), "", nil
		case "vigenere":
			res, err := VigenereEncrypt(plainText, keyStr)
			return res, "", err
		}
	}
	ivBytes, err := ParseIV(ivStr, algo.IVSize)
	if err != nil {
		return "", "", err
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		salt = []byte("KripDroidSalt2026")
	}

	cipherBytes, usedIV, err := EncryptBytesWithSalt(algoID, []byte(plainText), keyStr, ivBytes, salt)
	if err != nil {
		return "", "", err
	}
	ivHex := hex.EncodeToString(usedIV)
	if algoID == "xor" {
		return base64.StdEncoding.EncodeToString(cipherBytes), "", nil
	}

	payload := make([]byte, 0, 1+len(salt)+len(usedIV)+len(cipherBytes))
	payload = append(payload, byte(len(salt)))
	payload = append(payload, salt...)
	payload = append(payload, usedIV...)
	payload = append(payload, cipherBytes...)

	return base64.StdEncoding.EncodeToString(payload), ivHex, nil
}

func DecryptText(algoID string, cipherText string, keyStr string, ivStr string) (string, error) {
	algo, err := GetAlgorithm(algoID)
	if err != nil {
		return "", err
	}
	if algo.Category == CategoryClassic && algoID != "xor" {
		switch algoID {
		case "caesar":
			return CaesarDecrypt(cipherText, keyStr)
		case "rot13":
			return Rot13(cipherText), nil
		case "atbash":
			return Atbash(cipherText), nil
		case "vigenere":
			return VigenereDecrypt(cipherText, keyStr)
		}
	}
	trimmed := strings.TrimSpace(cipherText)
	var rawData []byte
	if b64, err := base64.StdEncoding.DecodeString(trimmed); err == nil {
		rawData = b64
	} else if hx, err := hex.DecodeString(trimmed); err == nil {
		rawData = hx
	} else {
		rawData = []byte(trimmed)
	}
	if algoID == "xor" {
		dec, err := DecryptBytes(algoID, rawData, keyStr, nil)
		if err != nil {
			return "", err
		}
		return string(dec), nil
	}

	if len(rawData) > 1+16+algo.IVSize && rawData[0] == 16 {
		salt := rawData[1:17]
		offset := 17
		var ivBytes []byte
		var cipherBytes []byte
		if strings.TrimSpace(ivStr) != "" {
			parsedIV, err := ParseIV(ivStr, algo.IVSize)
			if err != nil {
				return "", err
			}
			ivBytes = parsedIV
			cipherBytes = rawData[offset+algo.IVSize:]
		} else {
			ivBytes = rawData[offset : offset+algo.IVSize]
			cipherBytes = rawData[offset+algo.IVSize:]
		}

		dec, err := DecryptBytesWithSalt(algoID, cipherBytes, keyStr, ivBytes, salt)
		if err == nil {
			return string(dec), nil
		}
	}

	var ivBytes []byte
	var cipherBytes []byte
	if strings.TrimSpace(ivStr) != "" {
		parsedIV, err := ParseIV(ivStr, algo.IVSize)
		if err != nil {
			return "", err
		}
		ivBytes = parsedIV
		cipherBytes = rawData
	} else {
		if len(rawData) < algo.IVSize {
			return "", errors.New("ciphertext terlalu pendek untuk mengekstrak iv/nonce")
		}
		ivBytes = rawData[:algo.IVSize]
		cipherBytes = rawData[algo.IVSize:]
	}

	dec, err := DecryptBytes(algoID, cipherBytes, keyStr, ivBytes)
	if err != nil {
		return "", err
	}
	return string(dec), nil
}
