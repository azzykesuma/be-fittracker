package auth

import (
	"time"

	"be-fittracker/internal/utils"
)

type registerRequest struct {
	Name              string `json:"name"`
	Email             string `json:"email"`
	PasswordEncrypted string `json:"password_encrypted"`
	PasswordAlg       string `json:"password_alg"`
}

func (req registerRequest) validateEncryptedPassword() error {
	return validateEncryptedPassword(req.PasswordEncrypted, req.PasswordAlg)
}

type loginRequest struct {
	Email             string `json:"email"`
	PasswordEncrypted string `json:"password_encrypted"`
	PasswordAlg       string `json:"password_alg"`
}

func (req loginRequest) validateEncryptedPassword() error {
	return validateEncryptedPassword(req.PasswordEncrypted, req.PasswordAlg)
}

func validateEncryptedPassword(ciphertextBase64 string, alg string) error {
	if ciphertextBase64 == "" || alg != utils.PasswordAlgorithmRSAOAEPWithSHA256 {
		return ErrInvalidAuthRequest
	}
	return nil
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type userResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type tokenResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         userResponse `json:"user"`
}
