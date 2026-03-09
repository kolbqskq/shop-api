package auth

import (
	"context"
	"errors"
	"log"
	"shop-api/internal/errs"
	"shop-api/internal/user"

	"github.com/google/uuid"
)

type Service struct {
	userRepo   IUserRepository
	jwtService IJWTService
}
type ServiceDeps struct {
	UserRepository IUserRepository
	JWTService     IJWTService
}

func NewService(deps ServiceDeps) *Service {
	return &Service{
		userRepo:   deps.UserRepository,
		jwtService: deps.JWTService,
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

func (s *Service) Login(ctx context.Context, email, password string) (access, refresh string, err error) {
	e, err := user.NewEmail(email)
	if err != nil {
		return "", "", errs.ErrInvalidCredentials
	}
	u, err := s.userRepo.GetByEmail(ctx, e)
	if err != nil {
		return "", "", errs.ErrInvalidCredentials
	}

	if err := u.Login(password); err != nil {
		return "", "", err
	}

	if err := s.userRepo.Save(ctx, u); err != nil {
		log.Default().Println("failed to save") //поменять лог
	}

	accessToken, err := s.jwtService.CreateAccessToken(u.ID(), string(u.Role()))
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.jwtService.CreateRefreshToken(ctx, u.ID())
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.jwtService.DeleteRefresh(ctx, refreshToken)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (access, refresh string, err error) {
	return s.jwtService.Refresh(ctx, refreshToken)
}

func (s *Service) CreateAdmin(ctx context.Context, email, password string) error {
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
	user, err := user.NewUser(id, e, hashedPassword, user.Admin)
	return s.userRepo.Create(ctx, user)
}
