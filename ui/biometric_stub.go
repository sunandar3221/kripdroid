//go:build !android

package ui

import "errors"

func NativeBiometricPrompt(title, desc string) error {
	return errors.New("Biometrik asli hanya didukung pada platform Android")
}
