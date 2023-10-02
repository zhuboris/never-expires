package apn

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zhuboris/never-expires/internal/shared/servicechecker"
	"github.com/zhuboris/never-expires/internal/shared/usrctx"
)

type (
	deviceRepo interface {
		addDeviceToken(ctx context.Context, userID pgtype.UUID, token string) error
		Ping(ctx context.Context) error
	}
	userIDDecoder interface {
		Decode(ctx context.Context) (pgtype.UUID, error)
	}
)

type DeviceService struct {
	repo         deviceRepo
	usrID        userIDDecoder
	statusMetric servicechecker.StatusDisplay
}

func NewDeviceService(repo deviceRepo, statusDisplay servicechecker.StatusDisplay) *DeviceService {
	return &DeviceService{
		repo:         repo,
		usrID:        usrctx.ID{},
		statusMetric: statusDisplay,
	}
}

func (s DeviceService) AddDeviceToken(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("token string is empty")
	}

	userID, err := s.usrID.Decode(ctx)
	if err != nil {
		return err
	}

	return s.repo.addDeviceToken(ctx, userID, token)
}

func (s DeviceService) Status(ctx context.Context) error {
	return servicechecker.Ping(ctx, s.repo, s.statusMetric, "apnsRepository")
}
