package crypt

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"dario.lol/cf/internal/constants"
	"filippo.io/age"
	"github.com/zalando/go-keyring"
)

func Encrypt(data []byte, recipients ...age.Recipient) ([]byte, error) {
	var encrypted bytes.Buffer
	w, err := age.Encrypt(&encrypted, recipients...)
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

func Decrypt(encryptedData []byte, identities ...age.Identity) ([]byte, error) {
	r, err := age.Decrypt(bytes.NewReader(encryptedData), identities...)
	if err != nil {
		return nil, fmt.Errorf("failed to Decrypt data: %w", err)
	}

	decrypted, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read decrypted data: %w", err)
	}

	return decrypted, nil
}

func GetOrGenerateX25519Identity() (*age.X25519Identity, error) {
	identityStr, err := keyring.Get(constants.ServiceName, "identity")
	if err == nil {
		return age.ParseX25519Identity(identityStr)
	}
	if !errors.Is(err, keyring.ErrNotFound) {
		return nil, fmt.Errorf("failed to get identity from keyring: %w", err)
	}

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("failed to generate identity: %w", err)
	}

	err = keyring.Set(constants.ServiceName, "identity", identity.String())
	if err != nil {
		return nil, fmt.Errorf("failed to set identity in keyring: %w", err)
	}
	return identity, nil
}

func GetLegacyScryptIdentity() (*age.ScryptIdentity, error) {
	password, err := keyring.Get(constants.ServiceName, "user")
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil, nil // No legacy identity found
		}
		return nil, fmt.Errorf("failed to get password from keyring: %w", err)
	}

	identity, err := age.NewScryptIdentity(password)
	if err != nil {
		return nil, fmt.Errorf("failed to create scrypt identity: %w", err)
	}
	return identity, nil
}
