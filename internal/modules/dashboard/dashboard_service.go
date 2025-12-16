package dashboard

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/parlorhub/api-core/internal/database/sqlc"
)

type AppointmentStats struct {
	Total     int64 `json:"total"`
	Done      int64 `json:"done"`
	Cancelled int64 `json:"cancelled"`
	NoShow    int64 `json:"no_show"`
	Pending   int64 `json:"pending"`
	Confirmed int64 `json:"confirmed"`
}

type ServiceCount struct {
	ServiceID uuid.UUID `json:"service_id"`
	Name      string    `json:"name"`
	Count     int64     `json:"count"`
}

type ServicesStats struct {
	Total      int64          `json:"total"`
	PerService []ServiceCount `json:"per_service"`
}

type ClientsStats struct {
	NewClients       int64 `json:"new_clients"`
	UniqueThisMonth  int64 `json:"unique_this_month"`
	Engaged90Days    int64 `json:"engaged_last_90_days"`
	ActiveLast90Days int64 `json:"active_last_90_days"`
}

type Summary struct {
	Month        time.Time        `json:"month"`
	Appointments AppointmentStats `json:"appointments"`
	Services     ServicesStats    `json:"services"`
	Clients      ClientsStats     `json:"clients"`
}

type Service struct {
	queries *sqlc.Queries
}

var (
	ErrSalonNotFound  = errors.New("salon not found")
	ErrSalonForbidden = errors.New("forbidden")
)

func NewService(db *sql.DB) *Service {
	return &Service{queries: sqlc.New(db)}
}

func (s *Service) ensureSalonOwnership(ctx context.Context, salonID, ownerID uuid.UUID) error {
	salon, err := s.queries.GetSalonByID(ctx, salonID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrSalonNotFound
		}
		return err
	}

	if salon.OwnerID != ownerID {
		return ErrSalonForbidden
	}

	return nil
}

func monthBounds(month time.Time) (time.Time, time.Time) {
	start := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	return start, end
}

func (s *Service) getAppointmentStats(ctx context.Context, salonID uuid.UUID, month time.Time) (AppointmentStats, error) { //nolint:revive
	start, end := monthBounds(month)
	stats := AppointmentStats{}

	rows, err := s.queries.DashboardAppointmentStatus(ctx, sqlc.DashboardAppointmentStatusParams{
		SalonID: salonID,
		Date:    start,
		Date_2:  end,
	})
	if err != nil {
		return stats, err
	}

	for _, row := range rows {
		stats.Total += row.Count
		switch row.Status {
		case sqlc.AppointmentStatusDone:
			stats.Done = row.Count
		case sqlc.AppointmentStatusCancelled:
			stats.Cancelled = row.Count
		case sqlc.AppointmentStatusNoShow:
			stats.NoShow = row.Count
		case sqlc.AppointmentStatusPending:
			stats.Pending = row.Count
		case sqlc.AppointmentStatusConfirmed:
			stats.Confirmed = row.Count
		}
	}

	return stats, nil
}

func (s *Service) getServicesStats(ctx context.Context, salonID uuid.UUID, month time.Time) (ServicesStats, error) { //nolint:revive
	start, end := monthBounds(month)
	stats := ServicesStats{}

	rows, err := s.queries.DashboardServiceCounts(ctx, sqlc.DashboardServiceCountsParams{
		SalonID: salonID,
		Date:    start,
		Date_2:  end,
	})
	if err != nil {
		return stats, err
	}

	for _, row := range rows {
		sc := ServiceCount{ServiceID: row.ServiceID, Name: row.Name, Count: row.Count}
		stats.Total += row.Count
		stats.PerService = append(stats.PerService, sc)
	}

	return stats, nil
}

func (s *Service) getClientsStats(ctx context.Context, salonID uuid.UUID, month time.Time) (ClientsStats, error) { //nolint:revive
	start, end := monthBounds(month)
	ninetyAgo := end.AddDate(0, 0, -90)
	stats := ClientsStats{}

	var err error
	if stats.NewClients, err = s.queries.DashboardNewClients(ctx, sqlc.DashboardNewClientsParams{
		SalonID:     salonID,
		CreatedAt:   start,
		CreatedAt_2: end,
	}); err != nil {
		return stats, err
	}

	if stats.UniqueThisMonth, err = s.queries.DashboardUniqueClients(ctx, sqlc.DashboardUniqueClientsParams{
		SalonID: salonID,
		Date:    start,
		Date_2:  end,
	}); err != nil {
		return stats, err
	}

	if stats.ActiveLast90Days, err = s.queries.DashboardActiveClients(ctx, sqlc.DashboardActiveClientsParams{
		SalonID: salonID,
		Date:    ninetyAgo,
		Date_2:  end,
	}); err != nil {
		return stats, err
	}

	if stats.Engaged90Days, err = s.queries.DashboardEngagedClients(ctx, sqlc.DashboardEngagedClientsParams{
		SalonID: salonID,
		Date:    ninetyAgo,
		Date_2:  end,
	}); err != nil {
		return stats, err
	}

	return stats, nil
}

func (s *Service) GetSummary(ctx context.Context, salonID uuid.UUID, month time.Time) (Summary, error) {
	summary := Summary{Month: time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)}

	var err error
	if summary.Appointments, err = s.getAppointmentStats(ctx, salonID, month); err != nil {
		return Summary{}, err
	}

	if summary.Services, err = s.getServicesStats(ctx, salonID, month); err != nil {
		return Summary{}, err
	}

	if summary.Clients, err = s.getClientsStats(ctx, salonID, month); err != nil {
		return Summary{}, err
	}

	return summary, nil
}
