package auth

import "testing"

func TestRegisterRequestRejectsMissingEncryptedPassword(t *testing.T) {
	req := registerRequest{Name: "User", Email: "user@example.com", PasswordAlg: "RSA-OAEP-SHA256", Gender: "male"}
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

func TestPatchUserRequestValidation(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		name := "Alice"
		height := 165
		weight := 58.5
		gender := "female"
		req := patchUserRequest{
			Name:     &name,
			HeightCM: &height,
			WeightKG: &weight,
			Gender:   &gender,
		}
		if err := req.validate(); err != nil {
			t.Fatalf("unexpected validation error: %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		name := " "
		req := patchUserRequest{Name: &name}
		if err := req.validate(); err == nil || err.Error() != "name cannot be empty" {
			t.Fatalf("expected name validation error, got %v", err)
		}
	})

	t.Run("negative height", func(t *testing.T) {
		height := 0
		req := patchUserRequest{HeightCM: &height}
		if err := req.validate(); err == nil || err.Error() != "height_cm must be greater than 0" {
			t.Fatalf("expected height validation error, got %v", err)
		}
	})

	t.Run("invalid gender", func(t *testing.T) {
		gender := "unknown"
		req := patchUserRequest{Gender: &gender}
		if err := req.validate(); err == nil || err.Error() != "gender must be either 'male' or 'female'" {
			t.Fatalf("expected gender validation error, got %v", err)
		}
	})
}
