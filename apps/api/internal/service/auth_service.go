package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
)

const LegacyDemoToken = "phase1-demo-token"

const authTokenSecret = "voidlab-phase3-auth-secret"

var ErrInvalidToken = errors.New("invalid token")

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Login(username, password string) (domain.AuthSession, error) {
	if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
		return domain.AuthSession{}, errors.New("username and password are required")
	}

	user, err := s.userRepo.FindCredentialByUsername(strings.TrimSpace(username))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.AuthSession{}, errors.New("invalid username or password")
		}
		return domain.AuthSession{}, err
	}

	if !user.IsActive {
		return domain.AuthSession{}, errors.New("user is inactive")
	}

	if hashPassword(password) != user.PasswordHash {
		return domain.AuthSession{}, errors.New("invalid username or password")
	}

	currentUser := buildUserProfile(user.ID, user.Username, user.Role)

	return domain.AuthSession{
		Token: signToken(currentUser.ID),
		User:  currentUser,
	}, nil
}

func (s *AuthService) ResolveToken(token string) (domain.User, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return domain.User{}, ErrInvalidToken
	}

	if token == LegacyDemoToken {
		user, err := s.userRepo.FindDefaultAdmin()
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.User{}, ErrInvalidToken
			}
			return domain.User{}, err
		}

		return user, nil
	}

	userID, err := parseToken(token)
	if err != nil {
		return domain.User{}, err
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, ErrInvalidToken
		}
		return domain.User{}, err
	}

	if !user.IsActive {
		return domain.User{}, ErrInvalidToken
	}

	return user, nil
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func signToken(userID int64) string {
	payload := strconv.FormatInt(userID, 10)
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))
	return fmt.Sprintf("%s.%s", encodedPayload, signPayload(payload))
}

func parseToken(token string) (int64, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, ErrInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, ErrInvalidToken
	}

	payload := string(payloadBytes)
	if !hmac.Equal([]byte(parts[1]), []byte(signPayload(payload))) {
		return 0, ErrInvalidToken
	}

	userID, err := strconv.ParseInt(payload, 10, 64)
	if err != nil || userID <= 0 {
		return 0, ErrInvalidToken
	}

	return userID, nil
}

func signPayload(payload string) string {
	mac := hmac.New(sha256.New, []byte(authTokenSecret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func buildUserProfile(id int64, username, role string) domain.User {
	displayName := "VOIDLAB Operator"

	switch role {
	case "admin":
		displayName = "VOIDLAB Admin"
	case "editor":
		displayName = "VOIDLAB Editor"
	case "ops":
		displayName = "VOIDLAB Ops"
	}

	if strings.TrimSpace(username) == "" {
		username = "operator"
	}

	return domain.User{
		ID:          id,
		Username:    username,
		Role:        role,
		DisplayName: displayName,
		IsActive:    true,
	}
}
