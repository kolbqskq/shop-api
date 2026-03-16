package auth

import (
	"context"
	"errors"
	"shop-api/internal/core/errs"
	"shop-api/internal/features/user"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Service struct {
	userRepo   IUserRepository
	jwtService IJWTService
	txManager  ITxManager
	logger     zerolog.Logger
}
type ServiceDeps struct {
	UserRepository IUserRepository
	JWTService     IJWTService
	TxManager      ITxManager
	Logger         zerolog.Logger
}

func NewService(deps ServiceDeps) *Service {
	return &Service{
		userRepo:   deps.UserRepository,
		jwtService: deps.JWTService,
		txManager:  deps.TxManager,
		logger:     deps.Logger,
	}
}

func (s *Service) Register(ctx context.Context, email, password string) error {

	return s.createUser(ctx, email, password, user.Default)
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

	accessToken, err := s.jwtService.CreateAccessToken(u.ID(), string(u.Role()))
	if err != nil {
		return "", "", err
	}

	var refreshToken string
	err = s.txManager.WithTx(ctx, func(ctx context.Context) error {
		if err := s.userRepo.Save(ctx, u); err != nil {
			s.logger.Warn().Err(err).Msg("failed to update last_login_at")
		}

		if err := s.jwtService.DeleteRefreshByUserID(ctx, u.ID()); err != nil {
			return err
		}

		refreshToken, err = s.jwtService.CreateRefreshToken(ctx, u.ID())
		if err != nil {
			return err
		}

		return nil
	})
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
	return s.createUser(ctx, email, password, user.Admin)
}

func (s *Service) createUser(ctx context.Context, email, password string, role user.Role) error {
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
	user, err := user.NewUser(id, e, hashedPassword, role)
	if err != nil {
		return err
	}

	return s.userRepo.Create(ctx, user)
}
