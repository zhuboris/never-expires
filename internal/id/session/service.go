package session

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zhuboris/never-expires/internal/shared/servicechecker"
)

type repository interface {
	add(ctx context.Context, session Session) (pgtype.UUID, error)
	deactivate(ctx context.Context, opts option) error
	contains(ctx context.Context, session Session) error
	isDeviceNewWhenUserHadSessionsBefore(ctx context.Context, session Session) (bool, error)
	Ping(ctx context.Context) error
}

type Service struct {
	repo         repository
	statusMetric servicechecker.StatusDisplay
}

func NewService(repo repository, statusDisplay servicechecker.StatusDisplay) *Service {
	return &Service{
		repo:         repo,
		statusMetric: statusDisplay,
	}
}

func (s Service) Add(ctx context.Context, session Session) (pgtype.UUID, error) {
	return s.repo.add(ctx, session)
}

func (s Service) Deactivate(ctx context.Context, sessionID pgtype.UUID) error {
	return s.repo.deactivate(ctx, bySession(sessionID))
}

func (s Service) DeactivateAll(ctx context.Context) error {
	return s.repo.deactivate(ctx, byUser())
}

func (s Service) Contains(ctx context.Context, session Session) error {
	return s.repo.contains(ctx, session)
}

func (s Service) IsDeviceNewWhenUserHadSessionsBefore(ctx context.Context, session Session) (bool, error) {
	return s.repo.isDeviceNewWhenUserHadSessionsBefore(ctx, session)
}

func (s Service) Status(ctx context.Context) error {
	return servicechecker.Ping(ctx, s.repo, s.statusMetric, "sessionRepository")

}
