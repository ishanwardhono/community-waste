package household

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ishanwardhono/community-waste/pkg/httpres"
	"github.com/ishanwardhono/community-waste/pkg/logger"
)

//go:generate go tool mockgen -source=service.go -destination=../../test/mocks/household/service.go -package=mockhousehold
type Service interface {
	Create(ctx context.Context, req CreateRequest) (Household, error)
	List(ctx context.Context, page httpres.Page) ([]Household, int64, error)
	Get(ctx context.Context, id uuid.UUID) (Household, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (Household, error) {
	now := time.Now()
	h := Household{
		ID:        uuid.Must(uuid.NewV7()),
		OwnerName: strings.TrimSpace(req.OwnerName),
		Address:   strings.TrimSpace(req.Address),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Insert(ctx, h); err != nil {
		logger.Errorf(ctx, "insert household: %v", err)
		return Household{}, err
	}
	return h, nil
}

func (s *service) List(ctx context.Context, page httpres.Page) ([]Household, int64, error) {
	return s.repo.List(ctx, page.Limit, page.Offset())
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (Household, error) {
	return s.repo.Get(ctx, id)
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
