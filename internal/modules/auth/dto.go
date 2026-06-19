package auth

import (
	"errors"
	"strings"
	"time"

	"be-fittracker/internal/utils"
)

type registerRequest struct {
	Name              string `json:"name"`
	Email             string `json:"email"`
	PasswordEncrypted string `json:"password_encrypted"`
	PasswordAlg       string `json:"password_alg"`
	Gender            string `json:"gender"`
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
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	FitnessGoal *string   `json:"fitness_goal,omitempty"`
	HeightCM    *int      `json:"height_cm,omitempty"`
	WeightKG    *float64  `json:"weight_kg,omitempty"`
	Gender      string    `json:"gender"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type patchUserRequest struct {
	Name        *string  `json:"name"`
	FitnessGoal *string  `json:"fitness_goal"`
	HeightCM    *int     `json:"height_cm"`
	WeightKG    *float64 `json:"weight_kg"`
	Gender      *string  `json:"gender"`
}

func (req patchUserRequest) validate() error {
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return errors.New("name cannot be empty")
	}
	if req.HeightCM != nil && *req.HeightCM <= 0 {
		return errors.New("height_cm must be greater than 0")
	}
	if req.WeightKG != nil && *req.WeightKG <= 0 {
		return errors.New("weight_kg must be greater than 0")
	}
	if req.Gender != nil {
		g := strings.TrimSpace(*req.Gender)
		if g != "male" && g != "female" {
			return errors.New("gender must be either 'male' or 'female'")
		}
	}
	return nil
}

type tokenResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         userResponse `json:"user"`
}
