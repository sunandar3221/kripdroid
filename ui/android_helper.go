package ui

import (
	"io"
	"os"
	"path/filepath"
	"runtime"

	"gioui.org/app"
)

func IsAndroidPlatform() bool {
	return runtime.GOOS == "android" || os.Getenv("ANDROID_DATA") != "" || os.Getenv("ANDROID_ROOT") != ""
}

func GetSafeTempDir() string {
	if dir, err := app.DataDir(); err == nil && dir != "" {
		tmpDir := filepath.Join(dir, "krip_temp")
		if err := os.MkdirAll(tmpDir, 0755); err == nil {
			return tmpDir
		}
		return dir
	}
	candidates := []string{
		os.Getenv("TMPDIR"),
		os.TempDir(),
		".",
	}
	for _, d := range candidates {
		if d != "" {
			tmpDir := filepath.Join(d, "krip_temp")
			if err := os.MkdirAll(tmpDir, 0755); err == nil {
				return tmpDir
			}
		}
	}
	return "."
}

func CreateSafeTempFile(pattern string) (*os.File, error) {
	dir := GetSafeTempDir()
	return os.CreateTemp(dir, pattern)
}

func GetStateFilePath() string {
	if dir, err := app.DataDir(); err == nil && dir != "" {
		return filepath.Join(dir, "krip_state.json")
	}
	return filepath.Join(GetSafeTempDir(), "krip_state.json")
}

func SaveFileFromPathToWindowsDownloads(filename string, srcPath string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dlDir := filepath.Join(home, "Downloads")
	if err := os.MkdirAll(dlDir, 0755); err != nil {
		return "", err
	}

	target := filepath.Join(dlDir, filename)
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(target)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	buf := make([]byte, 64*1024)
	if _, err := io.CopyBuffer(dstFile, srcFile, buf); err != nil {
		return "", err
	}

	return target, nil
}
