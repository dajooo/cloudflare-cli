package config

import (
	"bytes"
	"encoding"
	"fmt"

	"github.com/dajooo/cloudflare-cli/internal/crypt"
)

type EncryptedString string

func (s *EncryptedString) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*s = ""
		return nil
	}

	if bytes.HasPrefix(text, []byte("age-encryption.org/")) ||
		bytes.HasPrefix(text, []byte("-----BEGIN AGE ENCRYPTED FILE-----")) {
		password, err := crypt.GetOrStoreNewlyGeneratedPassword()
		if err != nil {
			return fmt.Errorf("could not get password from keyring: %w", err)
		}

		decryptedBytes, err := crypt.Decrypt(text, password)
		if err != nil {
			return fmt.Errorf("failed to decrypt field: %w", err)
		}
		*s = EncryptedString(decryptedBytes)
		return nil
	}

	*s = EncryptedString(text)
	return nil
}

func (s EncryptedString) MarshalText() ([]byte, error) {
	if len(s) == 0 {
		return nil, nil
	}
	password, err := crypt.GetOrStoreNewlyGeneratedPassword()
	if err != nil {
		return nil, fmt.Errorf("could not get password from keyring for encryption: %w", err)
	}

	encryptedBytes, err := crypt.Encrypt([]byte(s), password)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt field for saving: %w", err)
	}

	return encryptedBytes, nil
}

var _ encoding.TextUnmarshaler = (*EncryptedString)(nil)
var _ encoding.TextMarshaler = (*EncryptedString)(nil)
