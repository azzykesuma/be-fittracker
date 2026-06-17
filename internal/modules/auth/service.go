package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"be-fittracker/internal/utils"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidRefreshToken = errors.New("invalid refresh token")
var ErrInvalidAuthRequest = errors.New("invalid auth request")

type Service struct {
	repo      *Repository
	jwtSecret string
}

func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret}
}

func (svc *Service) Register(ctx context.Context, req registerRequest) (tokenResponse, error) {
	name := strings.TrimSpace(req.Name)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if name == "" || email == "" {
		return tokenResponse{}, ErrInvalidAuthRequest
	}

	if err := req.validateEncryptedPassword(); err != nil {
		return tokenResponse{}, ErrInvalidAuthRequest
	}

	password, err := decryptRequestPassword(req.PasswordEncrypted, req.PasswordAlg)
	if err != nil {
		return tokenResponse{}, ErrInvalidAuthRequest
	}
	defer func() { password = "" }()

	if len(password) < 8 {
		return tokenResponse{}, ErrInvalidAuthRequest
	}

	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return tokenResponse{}, err
	}

	user, err := svc.repo.CreateUser(ctx, uuid.NewString(), name, email, passwordHash)
	if err != nil {
		return tokenResponse{}, err
	}

	return svc.issueTokens(ctx, user)
}

func (svc *Service) Login(ctx context.Context, req loginRequest) (tokenResponse, error) {
	if err := req.validateEncryptedPassword(); err != nil {
		return tokenResponse{}, ErrInvalidAuthRequest
	}

	password, err := decryptRequestPassword(req.PasswordEncrypted, req.PasswordAlg)
	if err != nil {
		return tokenResponse{}, ErrInvalidAuthRequest
	}
	defer func() { password = "" }()

	user, err := svc.repo.FindUserByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return tokenResponse{}, ErrInvalidCredentials
		}
		return tokenResponse{}, err
	}

	if !utils.CheckPassword(user.PasswordHash, password) {
		return tokenResponse{}, ErrInvalidCredentials
	}

	return svc.issueTokens(ctx, user)
}

func decryptRequestPassword(ciphertextBase64 string, alg string) (string, error) {
	password, err := utils.DecryptPassword(ciphertextBase64, alg)
	if err != nil {
		return "", ErrInvalidAuthRequest
	}
	return password, nil
}

func (svc *Service) Me(ctx context.Context, userID string) (userResponse, error) {
	user, err := svc.repo.FindUserByID(ctx, userID)
	if err != nil {
		return userResponse{}, err
	}
	return toUserResponse(user), nil
}

func (svc *Service) Refresh(ctx context.Context, token string) (tokenResponse, error) {
	if strings.TrimSpace(token) == "" {
		return tokenResponse{}, ErrInvalidRefreshToken
	}

	tokenHash := hashToken(token)
	userID, expiresAt, revoked, err := svc.repo.FindRefreshSession(ctx, tokenHash)
	if err != nil || revoked || time.Now().After(expiresAt) {
		return tokenResponse{}, ErrInvalidRefreshToken
	}

	if err := svc.repo.RevokeRefreshSession(ctx, tokenHash); err != nil {
		return tokenResponse{}, err
	}

	user, err := svc.repo.FindUserByID(ctx, userID)
	if err != nil {
		return tokenResponse{}, err
	}

	return svc.issueTokens(ctx, user)
}

func (svc *Service) Logout(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return nil
	}
	return svc.repo.RevokeRefreshSession(ctx, hashToken(token))
}

func (svc *Service) issueTokens(ctx context.Context, user userRecord) (tokenResponse, error) {
	accessToken, err := utils.SignAccessToken(svc.jwtSecret, user.ID, 15*time.Minute)
	if err != nil {
		return tokenResponse{}, err
	}

	refreshToken, err := randomToken()
	if err != nil {
		return tokenResponse{}, err
	}

	if err := svc.repo.CreateRefreshSession(ctx, uuid.NewString(), user.ID, hashToken(refreshToken), time.Now().Add(30*24*time.Hour)); err != nil {
		return tokenResponse{}, err
	}

	return tokenResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: toUserResponse(user)}, nil
}

func toUserResponse(user userRecord) userResponse {
	return userResponse{ID: user.ID, Name: user.Name, Email: user.Email, CreatedAt: user.CreatedAt}
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
