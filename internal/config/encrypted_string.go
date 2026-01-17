package config

import (
	"bytes"
	"encoding"
	"fmt"

	"dario.lol/cf/internal/crypt"
	"filippo.io/age"
)

type EncryptedString string

func (s *EncryptedString) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*s = ""
		return nil
	}

	if bytes.HasPrefix(text, []byte("age-encryption.org/")) ||
		bytes.HasPrefix(text, []byte("-----BEGIN AGE ENCRYPTED FILE-----")) {

		var identities []age.Identity

		// Always try to load the new optimized X25519 identity
		x25519Identity, err := crypt.GetOrGenerateX25519Identity()
		if err != nil {
			return fmt.Errorf("could not get identity from keyring: %w", err)
		}
		identities = append(identities, x25519Identity)

		// Try to load legacy identity for backward compatibility
		legacyIdentity, err := crypt.GetLegacyScryptIdentity()
		if err != nil {
			return fmt.Errorf("could not get legacy identity from keyring: %w", err)
		}
		if legacyIdentity != nil {
			identities = append(identities, legacyIdentity)
		}

		decryptedBytes, err := crypt.Decrypt(text, identities...)
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
	identity, err := crypt.GetOrGenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("could not get identity from keyring for encryption: %w", err)
	}

	encryptedBytes, err := crypt.Encrypt([]byte(s), identity.Recipient())
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt field for saving: %w", err)
	}

	return encryptedBytes, nil
}

var _ encoding.TextUnmarshaler = (*EncryptedString)(nil)
var _ encoding.TextMarshaler = (*EncryptedString)(nil)
