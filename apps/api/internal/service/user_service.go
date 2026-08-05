package service

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

type UserCreateInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UserRoleInput struct {
	Role string `json:"role"`
}

type UserStatusInput struct {
	IsActive bool `json:"is_active"`
}

type UserPasswordInput struct {
	Password string `json:"password"`
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) List() ([]domain.ManagedUser, error) {
	return s.userRepo.List()
}

func (s *UserService) GetByID(id int64) (domain.ManagedUser, error) {
	return s.userRepo.GetManagedByID(id)
}

func (s *UserService) Create(input UserCreateInput) (int64, error) {
	username := strings.TrimSpace(input.Username)
	password := strings.TrimSpace(input.Password)
	role := strings.TrimSpace(input.Role)

	if username == "" || password == "" {
		return 0, errors.New("username and password are required")
	}

	if len(password) < 4 {
		return 0, errors.New("password must be at least 4 characters")
	}

	if !isSupportedRole(role) {
		return 0, errors.New("invalid user role")
	}

	if _, err := s.userRepo.FindCredentialByUsername(username); err == nil {
		return 0, errors.New("username already exists")
	} else if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	return s.userRepo.Create(repository.UserCreateParams{
		Username:     username,
		PasswordHash: hashUserPassword(password),
		Role:         role,
	})
}

func (s *UserService) UpdateRole(id int64, input UserRoleInput, operatorID int64) error {
	role := strings.TrimSpace(input.Role)
	if !isSupportedRole(role) {
		return errors.New("invalid user role")
	}

	if id == operatorID {
		return errors.New("cannot change your own role")
	}

	if _, err := s.userRepo.FindByID(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("user not found")
		}
		return err
	}

	return s.userRepo.UpdateRole(id, role)
}

func (s *UserService) UpdateStatus(id int64, input UserStatusInput, operatorID int64) error {
	if id == operatorID {
		return errors.New("cannot update your own active status")
	}

	if _, err := s.userRepo.FindByID(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("user not found")
		}
		return err
	}

	return s.userRepo.UpdateStatus(id, input.IsActive)
}

func (s *UserService) ResetPassword(id int64, input UserPasswordInput) error {
	password := strings.TrimSpace(input.Password)
	if len(password) < 4 {
		return errors.New("password must be at least 4 characters")
	}

	if _, err := s.userRepo.FindByID(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("user not found")
		}
		return err
	}

	return s.userRepo.UpdatePassword(id, hashUserPassword(password))
}

func isSupportedRole(role string) bool {
	switch role {
	case "admin", "editor", "ops":
		return true
	default:
		return false
	}
}

func hashUserPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
