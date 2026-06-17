package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"os"
	"strings"
)

const PasswordAlgorithmRSAOAEPWithSHA256 = "RSA-OAEP-SHA256"

var ErrInvalidEncryptedPassword = errors.New("invalid encrypted password")

func LoadPasswordPrivateKeyFromEnv() (*rsa.PrivateKey, error) {
	keyPEM := strings.TrimSpace(os.Getenv("AUTH_PASSWORD_PRIVATE_KEY"))
	if keyPEM == "" {
		return nil, ErrInvalidEncryptedPassword
	}

	keyPEM = strings.Trim(keyPEM, `"'`)
	keyPEM = strings.ReplaceAll(keyPEM, `\n`, "\n")
	block, _ := pem.Decode([]byte(keyPEM))
	if block == nil {
		return nil, ErrInvalidEncryptedPassword
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, ErrInvalidEncryptedPassword
	}

	privateKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrInvalidEncryptedPassword
	}

	return privateKey, nil
}

func DecryptPassword(ciphertextBase64 string, alg string) (string, error) {
	if strings.TrimSpace(alg) != PasswordAlgorithmRSAOAEPWithSHA256 || strings.TrimSpace(ciphertextBase64) == "" {
		return "", ErrInvalidEncryptedPassword
	}

	privateKey, err := LoadPasswordPrivateKeyFromEnv()
	if err != nil {
		return "", ErrInvalidEncryptedPassword
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", ErrInvalidEncryptedPassword
	}

	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, nil)
	if err != nil {
		return "", ErrInvalidEncryptedPassword
	}
	defer zeroBytes(plaintext)

	return string(plaintext), nil
}

func zeroBytes(value []byte) {
	for i := range value {
		value[i] = 0
	}
}
