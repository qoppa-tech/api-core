package clients

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/parlorhub/api-core/internal/database/sqlc"
)

var ErrClientNotFound = errors.New("client not found")

type ClientsService struct {
	queries *sqlc.Queries
}

func NewClientsService(db *sql.DB) *ClientsService {
	return &ClientsService{queries: sqlc.New(db)}
}

func (s *ClientsService) CreateClient(ctx context.Context, req CreateClientRequest) (sqlc.Client, error) {
	params := sqlc.CreateClientParams{
		SalonID: req.SalonID,
		Name:    req.Name,
		Phone:   req.Phone,
	}

	if req.Email != nil {
		params.Email = sql.NullString{String: *req.Email, Valid: true}
	}

	if req.Birthday != nil {
		params.Birthday = sql.NullTime{Time: *req.Birthday, Valid: true}
	}

	return s.queries.CreateClient(ctx, params)
}

func (s *ClientsService) GetClientByID(ctx context.Context, id uuid.UUID) (sqlc.Client, error) {
	client, err := s.queries.GetClientByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return sqlc.Client{}, ErrClientNotFound
		}
		return sqlc.Client{}, err
	}
	return client, nil
}

func (s *ClientsService) ListClientsBySalonID(ctx context.Context, salonID uuid.UUID) ([]sqlc.Client, error) {
	return s.queries.ListClientsBySalonID(ctx, salonID)
}

func (s *ClientsService) SearchClientsByName(ctx context.Context, salonID uuid.UUID, name string) ([]sqlc.Client, error) {
	params := sqlc.SearchClientsByNameParams{
		SalonID: salonID,
		Column2: sql.NullString{String: name, Valid: true},
	}
	return s.queries.SearchClientsByName(ctx, params)
}

func (s *ClientsService) SearchClientsByPhone(ctx context.Context, salonID uuid.UUID, phone string) ([]sqlc.Client, error) {
	params := sqlc.SearchClientsByPhoneParams{
		SalonID: salonID,
		Column2: sql.NullString{String: phone, Valid: true},
	}
	return s.queries.SearchClientsByPhone(ctx, params)
}

func (s *ClientsService) UpdateClient(ctx context.Context, id uuid.UUID, req UpdateClientRequest) (sqlc.Client, error) {
	params := sqlc.UpdateClientParams{ID: id}

	if req.Name != nil {
		params.Name = sql.NullString{String: *req.Name, Valid: true}
	}

	if req.Phone != nil {
		params.Phone = sql.NullString{String: *req.Phone, Valid: true}
	}

	if req.Email != nil {
		params.Email = sql.NullString{String: *req.Email, Valid: true}
	}

	if req.Birthday != nil {
		params.Birthday = sql.NullTime{Time: *req.Birthday, Valid: true}
	}

	client, err := s.queries.UpdateClient(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return sqlc.Client{}, ErrClientNotFound
		}
		return sqlc.Client{}, err
	}

	return client, nil
}

func (s *ClientsService) DeleteClient(ctx context.Context, id uuid.UUID) error {
	if _, err := s.GetClientByID(ctx, id); err != nil {
		return err
	}

	if err := s.queries.DeleteClient(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			return ErrClientNotFound
		}
		return err
	}

	return nil
}

func (s *ClientsService) GetClientByPhone(ctx context.Context, salonID uuid.UUID, phone string) (sqlc.Client, error) {
	params := sqlc.GetClientByPhoneParams{
		SalonID: salonID,
		Phone:   phone,
	}

	client, err := s.queries.GetClientByPhone(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return sqlc.Client{}, ErrClientNotFound
		}
		return sqlc.Client{}, err
	}

	return client, nil
}
