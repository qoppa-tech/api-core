package salons

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/parlorhub/api-core/internal/database/sqlc"
	"github.com/sqlc-dev/pqtype"
)

var (
	ErrSalonNotFound  = errors.New("salon not found")
	ErrSalonForbidden = errors.New("forbidden")
)

type SalonsService struct {
	queries *sqlc.Queries
}

func NewSalonsService(db *sql.DB) *SalonsService {
	return &SalonsService{queries: sqlc.New(db)}
}

func (s *SalonsService) CreateSalon(ctx context.Context, req CreateSalonRequest, ownerID uuid.UUID) (sqlc.Salon, error) {
	params := sqlc.CreateSalonParams{
		Name:    req.Name,
		Slug:    req.Slug,
		OwnerID: ownerID,
		Address: req.Address,
	}

	if req.Whatsapp != nil {
		params.Whatsapp = sql.NullString{String: *req.Whatsapp, Valid: true}
	}

	if req.BusinessHours != nil {
		params.BusinessHours = req.BusinessHours
	} else {
		params.BusinessHours = json.RawMessage(`{}`)
	}

	if req.LogoUrl != nil {
		params.LogoUrl = sql.NullString{String: *req.LogoUrl, Valid: true}
	}

	return s.queries.CreateSalon(ctx, params)
}

func (s *SalonsService) GetSalonByID(ctx context.Context, id uuid.UUID) (sqlc.Salon, error) {
	salon, err := s.queries.GetSalonByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return sqlc.Salon{}, ErrSalonNotFound
		}
		return sqlc.Salon{}, err
	}
	return salon, nil
}

func (s *SalonsService) GetSalonBySlug(ctx context.Context, slug string) (sqlc.Salon, error) {
	salon, err := s.queries.GetSalonBySlug(ctx, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return sqlc.Salon{}, ErrSalonNotFound
		}
		return sqlc.Salon{}, err
	}
	return salon, nil
}

func (s *SalonsService) GetSalonByOwnerID(ctx context.Context, ownerID uuid.UUID) (sqlc.Salon, error) {
	salon, err := s.queries.GetSalonByOwnerID(ctx, ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return sqlc.Salon{}, ErrSalonNotFound
		}
		return sqlc.Salon{}, err
	}
	return salon, nil
}

func (s *SalonsService) ListSalons(ctx context.Context) ([]sqlc.Salon, error) {
	return s.queries.ListSalons(ctx)
}

func (s *SalonsService) UpdateSalon(ctx context.Context, id uuid.UUID, req UpdateSalonRequest) (sqlc.Salon, error) {
	params := sqlc.UpdateSalonParams{ID: id}

	if req.Name != nil {
		params.Name = sql.NullString{String: *req.Name, Valid: true}
	}

	if req.Slug != nil {
		params.Slug = sql.NullString{String: *req.Slug, Valid: true}
	}

	if req.Address != nil {
		params.Address = sql.NullString{String: *req.Address, Valid: true}
	}

	if req.Whatsapp != nil {
		params.Whatsapp = sql.NullString{String: *req.Whatsapp, Valid: true}
	}

	if req.BusinessHours != nil {
		params.BusinessHours = pqtype.NullRawMessage{RawMessage: req.BusinessHours, Valid: true}
	}

	if req.LogoUrl != nil {
		params.LogoUrl = sql.NullString{String: *req.LogoUrl, Valid: true}
	}

	salon, err := s.queries.UpdateSalon(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return sqlc.Salon{}, ErrSalonNotFound
		}
		return sqlc.Salon{}, err
	}

	return salon, nil
}

func (s *SalonsService) DeleteSalon(ctx context.Context, id uuid.UUID) error {
	err := s.queries.DeleteSalon(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrSalonNotFound
		}
		return err
	}
	return nil
}

// EnsureOwnership returns the salon if the caller is owner or admin; otherwise ErrSalonForbidden/ErrSalonNotFound.
func (s *SalonsService) EnsureOwnership(ctx context.Context, salonID, userID uuid.UUID, role string) (sqlc.Salon, error) {
	salon, err := s.GetSalonByID(ctx, salonID)
	if err != nil {
		return sqlc.Salon{}, err
	}

	if role == string(sqlc.UserRoleAdmin) {
		return salon, nil
	}

	if salon.OwnerID != userID {
		return sqlc.Salon{}, ErrSalonForbidden
	}

	return salon, nil
}
