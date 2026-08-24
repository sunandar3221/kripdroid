package crypto

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

func CaesarEncrypt(plainText string, key string) (string, error) {
	shift, err := parseShift(key)
	if err != nil {
		return "", err
	}
	shift = (shift%26 + 26) % 26
	var sb strings.Builder
	for _, r := range plainText {
		if unicode.IsUpper(r) {
			sb.WriteRune('A' + (r-'A'+rune(shift))%26)
		} else if unicode.IsLower(r) {
			sb.WriteRune('a' + (r-'a'+rune(shift))%26)
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String(), nil
}

func CaesarDecrypt(cipherText string, key string) (string, error) {
	shift, err := parseShift(key)
	if err != nil {
		return "", err
	}
	shift = (shift%26 + 26) % 26
	invShift := (26 - shift) % 26
	var sb strings.Builder
	for _, r := range cipherText {
		if unicode.IsUpper(r) {
			sb.WriteRune('A' + (r-'A'+rune(invShift))%26)
		} else if unicode.IsLower(r) {
			sb.WriteRune('a' + (r-'a'+rune(invShift))%26)
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String(), nil
}

func CaesarBytes(data []byte, key string, decrypt bool) ([]byte, error) {
	shift, err := parseShift(key)
	if err != nil {
		return nil, err
	}
	if decrypt {
		shift = -shift
	}
	s := byte((shift%256 + 256) % 256)
	res := make([]byte, len(data))
	for i, b := range data {
		res[i] = b + s
	}
	return res, nil
}

func parseShift(key string) (int, error) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return 3, nil
	}
	if n, err := strconv.Atoi(trimmed); err == nil {
		return n, nil
	}
	total := 0
	for _, r := range trimmed {
		total += int(r)
	}
	if total == 0 {
		return 3, nil
	}
	return total, nil
}

func Rot13(input string) string {
	var sb strings.Builder
	for _, r := range input {
		if r >= 'A' && r <= 'Z' {
			sb.WriteRune('A' + (r-'A'+13)%26)
		} else if r >= 'a' && r <= 'z' {
			sb.WriteRune('a' + (r-'a'+13)%26)
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func Rot13Bytes(data []byte) []byte {
	res := make([]byte, len(data))
	for i, b := range data {
		if b >= 'A' && b <= 'Z' {
			res[i] = 'A' + (b-'A'+13)%26
		} else if b >= 'a' && b <= 'z' {
			res[i] = 'a' + (b-'a'+13)%26
		} else {
			res[i] = b + 13
		}
	}
	return res
}

func Atbash(input string) string {
	var sb strings.Builder
	for _, r := range input {
		if unicode.IsUpper(r) {
			sb.WriteRune('Z' - (r - 'A'))
		} else if unicode.IsLower(r) {
			sb.WriteRune('z' - (r - 'a'))
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func AtbashBytes(data []byte) []byte {
	res := make([]byte, len(data))
	for i, b := range data {
		if b >= 'A' && b <= 'Z' {
			res[i] = 'Z' - (b - 'A')
		} else if b >= 'a' && b <= 'z' {
			res[i] = 'z' - (b - 'a')
		} else {
			res[i] = 255 - b
		}
	}
	return res
}

func VigenereEncrypt(plainText string, key string) (string, error) {
	k := strings.TrimSpace(key)
	if k == "" {
		return "", errors.New("vigenere key cannot be empty")
	}
	cleanKey := make([]rune, 0, len(k))
	for _, r := range k {
		if unicode.IsLetter(r) {
			cleanKey = append(cleanKey, unicode.ToUpper(r))
		}
	}
	if len(cleanKey) == 0 {
		cleanKey = []rune(strings.ToUpper(k))
	}
	var sb strings.Builder
	keyIdx := 0
	for _, r := range plainText {
		if unicode.IsUpper(r) {
			shift := cleanKey[keyIdx%len(cleanKey)] - 'A'
			sb.WriteRune('A' + (r-'A'+shift)%26)
			keyIdx++
		} else if unicode.IsLower(r) {
			shift := cleanKey[keyIdx%len(cleanKey)] - 'A'
			sb.WriteRune('a' + (r-'a'+shift)%26)
			keyIdx++
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String(), nil
}

func VigenereDecrypt(cipherText string, key string) (string, error) {
	k := strings.TrimSpace(key)
	if k == "" {
		return "", errors.New("vigenere key cannot be empty")
	}
	cleanKey := make([]rune, 0, len(k))
	for _, r := range k {
		if unicode.IsLetter(r) {
			cleanKey = append(cleanKey, unicode.ToUpper(r))
		}
	}
	if len(cleanKey) == 0 {
		cleanKey = []rune(strings.ToUpper(k))
	}
	var sb strings.Builder
	keyIdx := 0
	for _, r := range cipherText {
		if unicode.IsUpper(r) {
			shift := (cleanKey[keyIdx%len(cleanKey)] - 'A') % 26
			sb.WriteRune('A' + (r-'A'-shift+26)%26)
			keyIdx++
		} else if unicode.IsLower(r) {
			shift := (cleanKey[keyIdx%len(cleanKey)] - 'A') % 26
			sb.WriteRune('a' + (r-'a'-shift+26)%26)
			keyIdx++
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String(), nil
}

func VigenereBytes(data []byte, key string, decrypt bool) ([]byte, error) {
	k := []byte(key)
	if len(k) == 0 {
		return nil, errors.New("vigenere key cannot be empty")
	}
	res := make([]byte, len(data))
	for i, b := range data {
		shift := k[i%len(k)]
		if decrypt {
			res[i] = b - shift
		} else {
			res[i] = b + shift
		}
	}
	return res, nil
}

func XorCipher(data []byte, key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("xor key cannot be empty")
	}
	res := make([]byte, len(data))
	for i, b := range data {
		res[i] = b ^ key[i%len(key)]
	}
	return res, nil
}
