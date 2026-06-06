package auth

import (
	"errors"
	"fmt"
	"strings"

	"absensi-cn-api/internal/modules/user"
	"absensi-cn-api/pkg/password"
	"absensi-cn-api/pkg/token"

	"gorm.io/gorm"
)

type Service struct {
	jwtManager *token.JWTManager
	db         *gorm.DB
}

func NewService(jwtManager *token.JWTManager, db *gorm.DB) *Service {
	return &Service{
		jwtManager: jwtManager,
		db:         db,
	}
}

func (s *Service) Login(input LoginRequest) (*LoginResponse, error) {
	account, err := s.findUser(input)
	if err != nil {
		return nil, err
	}

	if err := password.Compare(account.PasswordHash, input.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	user := UserPayload{
		ID:       account.ID,
		Name:     account.Name,
		Role:     account.Role,
		Portal:   input.Portal,
		NIS:      account.NIS,
		Username: account.Username,
	}

	accessToken, err := s.jwtManager.Generate(user.ID, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &LoginResponse{
		AccessToken: accessToken,
		User:        user,
	}, nil
}

func (s *Service) findUser(input LoginRequest) (*user.User, error) {
	if s.db == nil {
		return nil, ErrAuthDataUnavailable
	}

	var account user.User
	var err error

	switch input.Portal {
	case "student":
		err = s.db.Where("nis = ?", strings.TrimSpace(input.NIS)).First(&account).Error
	case "staff":
		err = s.db.Where("LOWER(username) = ?", strings.ToLower(strings.TrimSpace(input.Username))).First(&account).Error
	default:
		return nil, ErrInvalidCredentials
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf("find user: %w", err)
	}

	switch input.Portal {
	case "student":
		if account.Role != user.RoleStudent {
			return nil, ErrStaffPortalOnly
		}
	case "staff":
		if account.Role == user.RoleStudent {
			return nil, ErrStudentPortalOnly
		}
	}

	return &account, nil
}
