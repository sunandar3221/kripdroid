package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

func AesCbcEncrypt(plainText, key, iv []byte) ([]byte, []byte, error) {
	if len(key) != 16 && len(key) != 32 {
		return nil, nil, errors.New("aes key must be 16 or 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	if len(iv) == 0 {
		iv = make([]byte, aes.BlockSize)
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return nil, nil, err
		}
	} else if len(iv) != aes.BlockSize {
		return nil, nil, errors.New("aes iv must be 16 bytes")
	}
	padded := PKCS7Padding(plainText, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)
	return ciphertext, iv, nil
}

func AesCbcDecrypt(ciphertext, key, iv []byte) ([]byte, error) {
	if len(key) != 16 && len(key) != 32 {
		return nil, errors.New("aes key must be 16 or 32 bytes")
	}
	if len(iv) != aes.BlockSize {
		return nil, errors.New("aes iv must be 16 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(ciphertext)%aes.BlockSize != 0 || len(ciphertext) == 0 {
		return nil, errors.New("invalid aes ciphertext length")
	}
	plain := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plain, ciphertext)
	return PKCS7Unpadding(plain, aes.BlockSize)
}

func AesGcmEncrypt(plainText, key, nonce []byte) ([]byte, []byte, error) {
	if len(key) != 16 && len(key) != 32 {
		return nil, nil, errors.New("aes key must be 16 or 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	if len(nonce) == 0 {
		nonce = make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return nil, nil, err
		}
	} else if len(nonce) != gcm.NonceSize() {
		return nil, nil, errors.New("invalid aes gcm nonce size")
	}
	ciphertext := gcm.Seal(nil, nonce, plainText, nil)
	return ciphertext, nonce, nil
}

func AesGcmDecrypt(ciphertext, key, nonce []byte) ([]byte, error) {
	if len(key) != 16 && len(key) != 32 {
		return nil, errors.New("aes key must be 16 or 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, errors.New("invalid aes gcm nonce size")
	}
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func ChaCha20Poly1305Encrypt(plainText, key, nonce []byte) ([]byte, []byte, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, nil, errors.New("chacha20poly1305 key must be 32 bytes")
	}
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, nil, err
	}
	if len(nonce) == 0 {
		nonce = make([]byte, chacha20poly1305.NonceSize)
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return nil, nil, err
		}
	} else if len(nonce) != chacha20poly1305.NonceSize {
		return nil, nil, errors.New("invalid chacha20poly1305 nonce size")
	}
	ciphertext := aead.Seal(nil, nonce, plainText, nil)
	return ciphertext, nonce, nil
}

func ChaCha20Poly1305Decrypt(ciphertext, key, nonce []byte) ([]byte, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, errors.New("chacha20poly1305 key must be 32 bytes")
	}
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}
	if len(nonce) != chacha20poly1305.NonceSize {
		return nil, errors.New("invalid chacha20poly1305 nonce size")
	}
	return aead.Open(nil, nonce, ciphertext, nil)
}
