package crypto

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/blowfish"
)

func PKCS7Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padText...)
}

func PKCS7Unpadding(src []byte, blockSize int) ([]byte, error) {
	length := len(src)
	if length == 0 || length%blockSize != 0 {
		return nil, errors.New("invalid ciphertext block size")
	}
	unpadding := int(src[length-1])
	if unpadding == 0 || unpadding > blockSize || unpadding > length {
		return nil, errors.New("invalid padding value")
	}
	for i := 0; i < unpadding; i++ {
		if src[length-1-i] != byte(unpadding) {
			return nil, errors.New("invalid padding bytes")
		}
	}
	return src[:(length - unpadding)], nil
}

func DesEncrypt(plainText, key, iv []byte) ([]byte, []byte, error) {
	if len(key) != 8 {
		return nil, nil, errors.New("des key must be exactly 8 bytes")
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	if len(iv) == 0 {
		iv = make([]byte, des.BlockSize)
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return nil, nil, err
		}
	} else if len(iv) != des.BlockSize {
		return nil, nil, errors.New("des iv must be 8 bytes")
	}
	padded := PKCS7Padding(plainText, des.BlockSize)
	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)
	return ciphertext, iv, nil
}

func DesDecrypt(ciphertext, key, iv []byte) ([]byte, error) {
	if len(key) != 8 {
		return nil, errors.New("des key must be exactly 8 bytes")
	}
	if len(iv) != des.BlockSize {
		return nil, errors.New("des iv must be 8 bytes")
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(ciphertext)%des.BlockSize != 0 || len(ciphertext) == 0 {
		return nil, errors.New("invalid des ciphertext length")
	}
	plain := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plain, ciphertext)
	return PKCS7Unpadding(plain, des.BlockSize)
}

func TripleDesEncrypt(plainText, key, iv []byte) ([]byte, []byte, error) {
	if len(key) != 24 {
		return nil, nil, errors.New("triple des key must be exactly 24 bytes")
	}
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, nil, err
	}
	if len(iv) == 0 {
		iv = make([]byte, des.BlockSize)
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return nil, nil, err
		}
	} else if len(iv) != des.BlockSize {
		return nil, nil, errors.New("triple des iv must be 8 bytes")
	}
	padded := PKCS7Padding(plainText, des.BlockSize)
	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)
	return ciphertext, iv, nil
}

func TripleDesDecrypt(ciphertext, key, iv []byte) ([]byte, error) {
	if len(key) != 24 {
		return nil, errors.New("triple des key must be exactly 24 bytes")
	}
	if len(iv) != des.BlockSize {
		return nil, errors.New("triple des iv must be 8 bytes")
	}
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	if len(ciphertext)%des.BlockSize != 0 || len(ciphertext) == 0 {
		return nil, errors.New("invalid triple des ciphertext length")
	}
	plain := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plain, ciphertext)
	return PKCS7Unpadding(plain, des.BlockSize)
}

func BlowfishEncrypt(plainText, key, iv []byte) ([]byte, []byte, error) {
	if len(key) < 1 || len(key) > 56 {
		return nil, nil, errors.New("blowfish key must be between 1 and 56 bytes")
	}
	block, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	blockSize := blowfish.BlockSize
	if len(iv) == 0 {
		iv = make([]byte, blockSize)
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return nil, nil, err
		}
	} else if len(iv) != blockSize {
		return nil, nil, errors.New("blowfish iv must be 8 bytes")
	}
	padded := PKCS7Padding(plainText, blockSize)
	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)
	return ciphertext, iv, nil
}

func BlowfishDecrypt(ciphertext, key, iv []byte) ([]byte, error) {
	if len(key) < 1 || len(key) > 56 {
		return nil, errors.New("blowfish key must be between 1 and 56 bytes")
	}
	blockSize := blowfish.BlockSize
	if len(iv) != blockSize {
		return nil, errors.New("blowfish iv must be 8 bytes")
	}
	block, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(ciphertext)%blockSize != 0 || len(ciphertext) == 0 {
		return nil, errors.New("invalid blowfish ciphertext length")
	}
	plain := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plain, ciphertext)
	return PKCS7Unpadding(plain, blockSize)
}
