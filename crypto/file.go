package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var FileMagicV1 = []byte("KRIPDROID\x01")
var FileMagicV2 = []byte("KRIPDROID\x02")
var FileMagicV3 = []byte("KRIPDROID\x03")

type FileProgressFunc func(processed int64, total int64)

type FileProcessResult struct {
	AlgoID      string
	OriginalExt string
	InputSize   int64
	OutputSize  int64
	Duration    time.Duration
}

func EncryptFileData(plainData []byte, algoID, keyStr string, customIV []byte, originalExt string) ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		salt = []byte("KripDroidSalt2026")
	}

	cipherData, usedIV, err := EncryptBytesWithSalt(algoID, plainData, keyStr, customIV, salt)
	if err != nil {
		return nil, err
	}

	var header bytes.Buffer
	header.Write(FileMagicV3)

	algoBytes := []byte(algoID)
	header.WriteByte(byte(len(algoBytes)))
	header.Write(algoBytes)

	cleanExt := strings.TrimSpace(originalExt)
	if cleanExt != "" && !strings.HasPrefix(cleanExt, ".") {
		cleanExt = "." + cleanExt
	}
	extBytes := []byte(cleanExt)
	header.WriteByte(byte(len(extBytes)))
	if len(extBytes) > 0 {
		header.Write(extBytes)
	}

	header.WriteByte(byte(len(salt)))
	if len(salt) > 0 {
		header.Write(salt)
	}

	header.WriteByte(byte(len(usedIV)))
	if len(usedIV) > 0 {
		header.Write(usedIV)
	}

	sizeBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeBuf, uint64(len(plainData)))
	header.Write(sizeBuf)
	header.Write(cipherData)

	return header.Bytes(), nil
}

func DecryptFileData(rawData []byte, keyStr string, fallbackAlgoID string) ([]byte, string, string, error) {
	var algoID string
	var originalExt string
	var salt []byte
	var iv []byte
	var cipherData []byte

	if len(rawData) > len(FileMagicV3) && bytes.Equal(rawData[:len(FileMagicV3)], FileMagicV3) {
		offset := len(FileMagicV3)
		if offset >= len(rawData) {
			return nil, "", "", errors.New("header file rusak")
		}
		algoLen := int(rawData[offset])
		offset++
		if offset+algoLen > len(rawData) {
			return nil, "", "", errors.New("header algoritma rusak")
		}
		algoID = string(rawData[offset : offset+algoLen])
		offset += algoLen

		if offset >= len(rawData) {
			return nil, "", "", errors.New("header format rusak")
		}
		extLen := int(rawData[offset])
		offset++
		if offset+extLen > len(rawData) {
			return nil, "", "", errors.New("data format rusak")
		}
		if extLen > 0 {
			originalExt = string(rawData[offset : offset+extLen])
			offset += extLen
		}

		if offset >= len(rawData) {
			return nil, "", "", errors.New("header salt rusak")
		}
		saltLen := int(rawData[offset])
		offset++
		if offset+saltLen > len(rawData) {
			return nil, "", "", errors.New("data salt rusak")
		}
		if saltLen > 0 {
			salt = rawData[offset : offset+saltLen]
			offset += saltLen
		}

		if offset >= len(rawData) {
			return nil, "", "", errors.New("header iv rusak")
		}
		ivLen := int(rawData[offset])
		offset++
		if offset+ivLen > len(rawData) {
			return nil, "", "", errors.New("byte iv rusak")
		}
		if ivLen > 0 {
			iv = rawData[offset : offset+ivLen]
			offset += ivLen
		}

		if offset+8 > len(rawData) {
			return nil, "", "", errors.New("header ukuran file rusak")
		}
		offset += 8
		cipherData = rawData[offset:]

		plainData, err := DecryptBytesWithSalt(algoID, cipherData, keyStr, iv, salt)
		if err != nil {
			return nil, "", "", err
		}
		return plainData, algoID, originalExt, nil

	} else if len(rawData) > len(FileMagicV2) && bytes.Equal(rawData[:len(FileMagicV2)], FileMagicV2) {
		offset := len(FileMagicV2)
		if offset >= len(rawData) {
			return nil, "", "", errors.New("header file rusak")
		}
		algoLen := int(rawData[offset])
		offset++
		if offset+algoLen > len(rawData) {
			return nil, "", "", errors.New("header algoritma rusak")
		}
		algoID = string(rawData[offset : offset+algoLen])
		offset += algoLen

		if offset >= len(rawData) {
			return nil, "", "", errors.New("header ekstensi rusak")
		}
		extLen := int(rawData[offset])
		offset++
		if offset+extLen > len(rawData) {
			return nil, "", "", errors.New("data ekstensi rusak")
		}
		if extLen > 0 {
			originalExt = string(rawData[offset : offset+extLen])
			offset += extLen
		}

		if offset >= len(rawData) {
			return nil, "", "", errors.New("header iv rusak")
		}
		ivLen := int(rawData[offset])
		offset++
		if offset+ivLen > len(rawData) {
			return nil, "", "", errors.New("byte iv rusak")
		}
		if ivLen > 0 {
			iv = rawData[offset : offset+ivLen]
			offset += ivLen
		}

		if offset+8 > len(rawData) {
			return nil, "", "", errors.New("header ukuran file rusak")
		}
		offset += 8
		cipherData = rawData[offset:]

		plainData, err := DecryptBytes(algoID, cipherData, keyStr, iv)
		if err != nil {
			return nil, "", "", err
		}
		return plainData, algoID, originalExt, nil

	} else if len(rawData) > len(FileMagicV1) && bytes.Equal(rawData[:len(FileMagicV1)], FileMagicV1) {
		offset := len(FileMagicV1)
		if offset >= len(rawData) {
			return nil, "", "", errors.New("header file rusak")
		}
		algoLen := int(rawData[offset])
		offset++
		if offset+algoLen > len(rawData) {
			return nil, "", "", errors.New("header algoritma rusak")
		}
		algoID = string(rawData[offset : offset+algoLen])
		offset += algoLen

		if offset >= len(rawData) {
			return nil, "", "", errors.New("header iv rusak")
		}
		ivLen := int(rawData[offset])
		offset++
		if offset+ivLen > len(rawData) {
			return nil, "", "", errors.New("byte iv rusak")
		}
		if ivLen > 0 {
			iv = rawData[offset : offset+ivLen]
			offset += ivLen
		}

		if offset+8 > len(rawData) {
			return nil, "", "", errors.New("header ukuran file rusak")
		}
		offset += 8
		cipherData = rawData[offset:]

		plainData, err := DecryptBytes(algoID, cipherData, keyStr, iv)
		if err != nil {
			return nil, "", "", err
		}
		return plainData, algoID, originalExt, nil
	} else {
		if fallbackAlgoID == "" {
			return nil, "", "", errors.New("format file tidak dikenali dan algoritma fallback tidak ditentukan")
		}
		algoID = fallbackAlgoID
		algo, err := GetAlgorithm(algoID)
		if err != nil {
			return nil, "", "", err
		}
		if algo.IVSize > 0 && len(rawData) >= algo.IVSize {
			iv = rawData[:algo.IVSize]
			cipherData = rawData[algo.IVSize:]
		} else {
			cipherData = rawData
		}
		plainData, err := DecryptBytes(algoID, cipherData, keyStr, iv)
		if err != nil {
			return nil, "", "", err
		}
		return plainData, algoID, originalExt, nil
	}
}

func EncryptFileWithExt(srcPath, dstPath, algoID, keyStr string, customIV []byte, originalExt string, progress FileProgressFunc) (*FileProcessResult, error) {
	startTime := time.Now()
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return nil, err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return nil, err
	}
	totalSize := srcInfo.Size()
	plainData, err := io.ReadAll(srcFile)
	if err != nil {
		return nil, err
	}
	if progress != nil {
		progress(int64(len(plainData))/2, totalSize)
	}

	ext := strings.TrimSpace(originalExt)
	if ext == "" {
		ext = filepath.Ext(srcPath)
	}

	encryptedFull, err := EncryptFileData(plainData, algoID, keyStr, customIV, ext)
	if err != nil {
		return nil, err
	}

	if dstPath == "" {
		dstPath = srcPath + ".krip"
	}
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil && filepath.Dir(dstPath) != "." && filepath.Dir(dstPath) != "" {
		return nil, err
	}
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return nil, err
	}
	defer dstFile.Close()

	if _, err := dstFile.Write(encryptedFull); err != nil {
		return nil, err
	}
	if progress != nil {
		progress(totalSize, totalSize)
	}
	outInfo, err := dstFile.Stat()
	if err != nil {
		return nil, err
	}
	return &FileProcessResult{
		AlgoID:      algoID,
		OriginalExt: ext,
		InputSize:   totalSize,
		OutputSize:  outInfo.Size(),
		Duration:    time.Since(startTime),
	}, nil
}

func EncryptFile(srcPath, dstPath, algoID, keyStr string, customIV []byte, progress FileProgressFunc) (*FileProcessResult, error) {
	return EncryptFileWithExt(srcPath, dstPath, algoID, keyStr, customIV, "", progress)
}

func DecryptFile(srcPath, dstPath, keyStr string, fallbackAlgoID string, progress FileProgressFunc) (*FileProcessResult, error) {
	startTime := time.Now()
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return nil, err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return nil, err
	}
	totalSize := srcInfo.Size()
	rawData, err := io.ReadAll(srcFile)
	if err != nil {
		return nil, err
	}
	if progress != nil {
		progress(int64(len(rawData))/2, totalSize)
	}

	plainData, algoID, originalExt, err := DecryptFileData(rawData, keyStr, fallbackAlgoID)
	if err != nil {
		return nil, err
	}

	if dstPath == "" {
		ext := filepath.Ext(srcPath)
		baseWithoutKrip := srcPath
		if ext == ".krip" {
			baseWithoutKrip = srcPath[:len(srcPath)-len(ext)]
		}
		if originalExt != "" {
			if filepath.Ext(baseWithoutKrip) != originalExt {
				dstPath = baseWithoutKrip + originalExt
			} else {
				dstPath = baseWithoutKrip
			}
		} else {
			if ext == ".krip" {
				dstPath = baseWithoutKrip
			} else {
				dstPath = srcPath + ".dec"
			}
		}
	}
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil && filepath.Dir(dstPath) != "." && filepath.Dir(dstPath) != "" {
		return nil, err
	}
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return nil, err
	}
	defer dstFile.Close()

	if _, err := dstFile.Write(plainData); err != nil {
		return nil, err
	}
	if progress != nil {
		progress(totalSize, totalSize)
	}
	outInfo, err := dstFile.Stat()
	if err != nil {
		return nil, err
	}
	return &FileProcessResult{
		AlgoID:      algoID,
		OriginalExt: originalExt,
		InputSize:   totalSize,
		OutputSize:  outInfo.Size(),
		Duration:    time.Since(startTime),
	}, nil
}
