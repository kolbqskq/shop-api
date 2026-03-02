package auth

import (
	"context"
	"errors"
	"shop-api/internal/errs"
	"shop-api/internal/user"

	"github.com/google/uuid"
)

type Service struct {
	userRepo IUserRepository
}
type ServiceDeps struct {
	UserRepository IUserRepository
}

func NewService(deps ServiceDeps) *Service {
	return &Service{
		userRepo: deps.UserRepository,
	}
}

func (s *Service) Register(ctx context.Context, email, password string) error {
	e, err := user.NewEmail(email)
	if err != nil {
		return err
	}
	_, err = s.userRepo.GetByEmail(ctx, e)
	if err == nil {
		return errs.ErrUserAlreadyExists
	}
	if !errors.Is(err, errs.ErrUserNotFound) {
		return err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	hashedPassword, err := user.NewPasswordHash(password)
	if err != nil {
		return err
	}
	user, err := user.NewUser(id, e, hashedPassword, user.Default)
	return s.userRepo.Create(ctx, user)
}

func (s *Service) Login(ctx context.Context, email, password string) error {
	e, err := user.NewEmail(email)
	if err != nil {
		return errs.ErrInvalidCredentials
	}
	u, err := s.userRepo.GetByEmail(ctx, e)
	if err != nil {
		return errs.ErrInvalidCredentials
	}

	if err := u.Login(password); err != nil {
		return err
	}

	return s.userRepo.Save(ctx, u)
}
