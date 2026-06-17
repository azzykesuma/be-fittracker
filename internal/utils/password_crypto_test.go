package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"
)

func TestDecryptPasswordRoundTrip(t *testing.T) {
	privateKey := testPrivateKey(t)
	t.Setenv("AUTH_PASSWORD_PRIVATE_KEY", testPrivateKeyPEM(t, privateKey))

	ciphertext := testEncryptPassword(t, &privateKey.PublicKey, "password123")
	password, err := DecryptPassword(ciphertext, PasswordAlgorithmRSAOAEPWithSHA256)
	if err != nil {
		t.Fatalf("DecryptPassword returned error: %v", err)
	}
	if password != "password123" {
		t.Fatalf("expected decrypted password to match")
	}
}

func TestDecryptPasswordRoundTripWithQuotedEnv(t *testing.T) {
	privateKey := testPrivateKey(t)
	t.Setenv("AUTH_PASSWORD_PRIVATE_KEY", `"`+testPrivateKeyPEM(t, privateKey)+`"`)

	ciphertext := testEncryptPassword(t, &privateKey.PublicKey, "password123")
	password, err := DecryptPassword(ciphertext, PasswordAlgorithmRSAOAEPWithSHA256)
	if err != nil {
		t.Fatalf("DecryptPassword returned error: %v", err)
	}
	if password != "password123" {
		t.Fatalf("expected decrypted password to match")
	}
}

func TestDecryptPasswordInvalidAlgorithm(t *testing.T) {
	privateKey := testPrivateKey(t)
	t.Setenv("AUTH_PASSWORD_PRIVATE_KEY", testPrivateKeyPEM(t, privateKey))

	_, err := DecryptPassword("ciphertext", "RSA-OAEP")
	if err != ErrInvalidEncryptedPassword {
		t.Fatalf("expected ErrInvalidEncryptedPassword, got %v", err)
	}
}

func TestDecryptPasswordInvalidBase64(t *testing.T) {
	privateKey := testPrivateKey(t)
	t.Setenv("AUTH_PASSWORD_PRIVATE_KEY", testPrivateKeyPEM(t, privateKey))

	_, err := DecryptPassword("not base64", PasswordAlgorithmRSAOAEPWithSHA256)
	if err != ErrInvalidEncryptedPassword {
		t.Fatalf("expected ErrInvalidEncryptedPassword, got %v", err)
	}
}

func TestDecryptPasswordWrongPrivateKey(t *testing.T) {
	encryptionKey := testPrivateKey(t)
	wrongKey := testPrivateKey(t)
	t.Setenv("AUTH_PASSWORD_PRIVATE_KEY", testPrivateKeyPEM(t, wrongKey))

	ciphertext := testEncryptPassword(t, &encryptionKey.PublicKey, "password123")
	_, err := DecryptPassword(ciphertext, PasswordAlgorithmRSAOAEPWithSHA256)
	if err != ErrInvalidEncryptedPassword {
		t.Fatalf("expected ErrInvalidEncryptedPassword, got %v", err)
	}
}

func testPrivateKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey returned error: %v", err)
	}
	return privateKey
}

func testPrivateKeyPEM(t *testing.T, privateKey *rsa.PrivateKey) string {
	t.Helper()
	bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey returned error: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: bytes}))
}

func testEncryptPassword(t *testing.T, publicKey *rsa.PublicKey, password string) string {
	t.Helper()
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, []byte(password), nil)
	if err != nil {
		t.Fatalf("EncryptOAEP returned error: %v", err)
	}
	return base64.StdEncoding.EncodeToString(ciphertext)
}
