package auth

import "testing"

func TestRegisterRequestRejectsMissingEncryptedPassword(t *testing.T) {
	req := registerRequest{Name: "User", Email: "user@example.com", PasswordAlg: "RSA-OAEP-SHA256"}
	if err := req.validateEncryptedPassword(); err != ErrInvalidAuthRequest {
		t.Fatalf("expected ErrInvalidAuthRequest, got %v", err)
	}
}

func TestLoginRequestRejectsMissingEncryptedPassword(t *testing.T) {
	req := loginRequest{Email: "user@example.com", PasswordAlg: "RSA-OAEP-SHA256"}
	if err := req.validateEncryptedPassword(); err != ErrInvalidAuthRequest {
		t.Fatalf("expected ErrInvalidAuthRequest, got %v", err)
	}
}

func TestAuthRequestRejectsInvalidPasswordAlgorithm(t *testing.T) {
	req := loginRequest{Email: "user@example.com", PasswordEncrypted: "abc", PasswordAlg: "RSA-OAEP"}
	if err := req.validateEncryptedPassword(); err != ErrInvalidAuthRequest {
		t.Fatalf("expected ErrInvalidAuthRequest, got %v", err)
	}
}
