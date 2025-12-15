package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/parlorhub/api-core/internal/database/sqlc"
)

var (
	ErrServiceNotFound  = errors.New("service not found")
	ErrServiceForbidden = errors.New("forbidden")
	ErrSalonNotFound    = errors.New("salon not found")
)

type ServicesService struct {
	queries *sqlc.Queries
}

func NewServicesService(db *sql.DB) *ServicesService {
	return &ServicesService{queries: sqlc.New(db)}
}

// EnsureSalonOwnership returns the salon when caller is owner or admin.
func (s *ServicesService) EnsureSalonOwnership(ctx context.Context, salonID, userID uuid.UUID, role string) (sqlc.Salon, error) {
	salon, err := s.queries.GetSalonByID(ctx, salonID)
	if err != nil {
		if err == sql.ErrNoRows {
			return sqlc.Salon{}, ErrSalonNotFound
		}
		return sqlc.Salon{}, err
	}

	if role == string(sqlc.UserRoleAdmin) {
		return salon, nil
	}

	if salon.OwnerID != userID {
		return sqlc.Salon{}, ErrServiceForbidden
	}

	return salon, nil
}

func (s *ServicesService) CreateService(ctx context.Context, req CreateServiceRequest) (sqlc.Service, error) {
	params := sqlc.CreateServiceParams{
		SalonID:  req.SalonID,
		Name:     req.Name,
		Duration: req.Duration,
		Price:    req.Price,
		Active:   true,
	}

	if req.Active != nil {
		params.Active = *req.Active
	}

	if req.UserID != nil {
		params.UserID = uuid.NullUUID{UUID: *req.UserID, Valid: true}
	}

	return s.queries.CreateService(ctx, params)
}

func (s *ServicesService) GetServiceByID(ctx context.Context, id uuid.UUID) (sqlc.Service, error) {
	service, err := s.queries.GetServiceByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return sqlc.Service{}, ErrServiceNotFound
		}
		return sqlc.Service{}, err
	}
	return service, nil
}

func (s *ServicesService) ListServicesBySalonID(ctx context.Context, salonID uuid.UUID) ([]sqlc.Service, error) {
	return s.queries.ListServicesBySalonID(ctx, salonID)
}

func (s *ServicesService) ListActiveServicesBySalonID(ctx context.Context, salonID uuid.UUID) ([]sqlc.Service, error) {
	return s.queries.ListActiveServicesBySalonID(ctx, salonID)
}

func (s *ServicesService) ListServicesByUserID(ctx context.Context, userID uuid.UUID) ([]sqlc.Service, error) {
	return s.queries.ListServicesByUserID(ctx, uuid.NullUUID{UUID: userID, Valid: true})
}

func (s *ServicesService) ListServicesBySalonAndUser(ctx context.Context, salonID, userID uuid.UUID) ([]sqlc.Service, error) {
	params := sqlc.ListServicesBySalonAndUserParams{
		SalonID: salonID,
		UserID:  uuid.NullUUID{UUID: userID, Valid: true},
	}
	return s.queries.ListServicesBySalonAndUser(ctx, params)
}

func (s *ServicesService) UpdateService(ctx context.Context, id uuid.UUID, req UpdateServiceRequest) (sqlc.Service, error) {
	params := sqlc.UpdateServiceParams{ID: id}

	if req.Name != nil {
		params.Name = sql.NullString{String: *req.Name, Valid: true}
	}

	if req.Duration != nil {
		params.Duration = sql.NullInt32{Int32: *req.Duration, Valid: true}
	}

	if req.Price != nil {
		params.Price = sql.NullInt32{Int32: *req.Price, Valid: true}
	}

	if req.Active != nil {
		params.Active = sql.NullBool{Bool: *req.Active, Valid: true}
	}

	if req.UserID != nil {
		params.UserID = uuid.NullUUID{UUID: *req.UserID, Valid: true}
	}

	service, err := s.queries.UpdateService(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return sqlc.Service{}, ErrServiceNotFound
		}
		return sqlc.Service{}, err
	}

	return service, nil
}

func (s *ServicesService) DeleteService(ctx context.Context, id uuid.UUID) error {
	err := s.queries.DeleteService(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrServiceNotFound
		}
		return err
	}

	return nil
}
