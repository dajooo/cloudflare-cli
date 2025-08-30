package crypt

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"dario.lol/cf/internal/constants"
	password2 "dario.lol/gotils/pkg/password"
	"filippo.io/age"
	"github.com/zalando/go-keyring"
)

func Encrypt(data []byte, password string) ([]byte, error) {
	recipient, err := age.NewScryptRecipient(password)
	if err != nil {
		return nil, fmt.Errorf("failed to create recipient: %w", err)
	}

	var encrypted bytes.Buffer
	w, err := age.Encrypt(&encrypted, recipient)
	if err != nil {
		return nil, fmt.Errorf("failed to create encrypted writer: %w", err)
	}
	if _, err := w.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write to encrypted writer: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close encrypted writer: %w", err)
	}

	return encrypted.Bytes(), nil
}

func Decrypt(encryptedData []byte, password string) ([]byte, error) {
	identity, err := age.NewScryptIdentity(password)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity: %w", err)
	}

	r, err := age.Decrypt(bytes.NewReader(encryptedData), identity)
	if err != nil {
		return nil, fmt.Errorf("failed to Decrypt data: %w", err)
	}

	decrypted, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read decrypted data: %w", err)
	}

	return decrypted, nil
}

func GetOrStoreNewlyGeneratedPassword() (string, error) {
	password, err := keyring.Get(constants.ServiceName, "user")
	if err == nil {
		return password, nil
	}
	if !errors.Is(err, keyring.ErrNotFound) {
		return "", fmt.Errorf("failed to get password from keyring: %w", err)
	}

	password, err = password2.Generate(password2.GenerateWithLengthOption(64))
	if err != nil {
		return "", fmt.Errorf("failed to generate password: %w", err)
	}
	err = keyring.Set(constants.ServiceName, "user", password)
	if err != nil {
		return "", fmt.Errorf("failed to set password in keyring: %w", err)
	}
	return password, nil
}
