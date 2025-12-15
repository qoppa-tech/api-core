package appointment

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/parlorhub/api-core/internal/database/sqlc"
)

var ErrAppointmentNotFound = errors.New("appointment not found")

type AppointmentService struct {
	queries *sqlc.Queries
}

func NewAppointmentService(db *sql.DB) *AppointmentService {
	return &AppointmentService{queries: sqlc.New(db)}
}

func (s *AppointmentService) CreateAppointment(ctx context.Context, params sqlc.CreateAppointmentParams) (sqlc.Appointment, error) {
	return s.queries.CreateAppointment(ctx, params)
}

func (s *AppointmentService) GetAppointmentByID(ctx context.Context, id uuid.UUID) (sqlc.Appointment, error) {
	appt, err := s.queries.GetAppointmentByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return sqlc.Appointment{}, ErrAppointmentNotFound
		}
		return sqlc.Appointment{}, err
	}
	return appt, nil
}

func (s *AppointmentService) ListAppointmentsBySalonID(ctx context.Context, salonID uuid.UUID) ([]sqlc.Appointment, error) {
	return s.queries.ListAppointmentsBySalonID(ctx, salonID)
}

func (s *AppointmentService) ListAppointmentsByDate(ctx context.Context, params sqlc.ListAppointmentsByDateParams) ([]sqlc.Appointment, error) {
	return s.queries.ListAppointmentsByDate(ctx, params)
}

func (s *AppointmentService) ListAppointmentsByDateRange(ctx context.Context, params sqlc.ListAppointmentsByDateRangeParams) ([]sqlc.Appointment, error) {
	return s.queries.ListAppointmentsByDateRange(ctx, params)
}

func (s *AppointmentService) ListAppointmentsByUser(ctx context.Context, params sqlc.ListAppointmentsByUserParams) ([]sqlc.Appointment, error) {
	return s.queries.ListAppointmentsByUser(ctx, params)
}

func (s *AppointmentService) ListAppointmentsByStatus(ctx context.Context, params sqlc.ListAppointmentsByStatusParams) ([]sqlc.Appointment, error) {
	return s.queries.ListAppointmentsByStatus(ctx, params)
}

func (s *AppointmentService) ListAppointmentsByClientPhone(ctx context.Context, params sqlc.ListAppointmentsByClientPhoneParams) ([]sqlc.Appointment, error) {
	return s.queries.ListAppointmentsByClientPhone(ctx, params)
}

func (s *AppointmentService) UpdateAppointment(ctx context.Context, params sqlc.UpdateAppointmentParams) (sqlc.Appointment, error) {
	appt, err := s.queries.UpdateAppointment(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return sqlc.Appointment{}, ErrAppointmentNotFound
		}
		return sqlc.Appointment{}, err
	}
	return appt, nil
}

func (s *AppointmentService) UpdateAppointmentStatus(ctx context.Context, params sqlc.UpdateAppointmentStatusParams) (sqlc.Appointment, error) {
	appt, err := s.queries.UpdateAppointmentStatus(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return sqlc.Appointment{}, ErrAppointmentNotFound
		}
		return sqlc.Appointment{}, err
	}
	return appt, nil
}

func (s *AppointmentService) DeleteAppointment(ctx context.Context, id uuid.UUID) error {
	if _, err := s.GetAppointmentByID(ctx, id); err != nil {
		return err
	}

	if err := s.queries.DeleteAppointment(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			return ErrAppointmentNotFound
		}
		return err
	}

	return nil
}

func (s *AppointmentService) CountAppointmentsByUserAndDate(ctx context.Context, params sqlc.CountAppointmentsByUserAndDateParams) (int64, error) {
	return s.queries.CountAppointmentsByUserAndDate(ctx, params)
}
