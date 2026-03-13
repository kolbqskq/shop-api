package payment

import (
	"context"

	"github.com/google/uuid"
)

type Service struct{}

func NewService() *Service{
	return &Service{}
}

func (s *Service) Pay(ctx context.Context, orderID uuid.UUID, amount int64) error {
	return nil
}
