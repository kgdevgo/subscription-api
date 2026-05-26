package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"subscription-api/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type SubscriptionRepository struct {
	db *pgxpool.DB
}

func NewSubscriptionRepository(db *pgxpool.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *domain.Subscription) error {
	const op = "repository.postgres.Create"

	query := `
		INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	var endDate interface{}
	if sub.EndDate != nil {
		endDate = sub.EndDate.Time
	} else {
		endDate = nil
	}

	_, err := r.db.Exec(ctx, query, sub.ID, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate.Time, endDate)
	if err != nil {
		slog.Error("database error", slog.String("op", op), slog.String("err", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	const op = "repository.postgres.GetByID"

	query := `
		SELECT id, service_name, price, user_id, start_date, end_date 
		FROM subscriptions 
		WHERE id = $1
	`

	var sub domain.Subscription
	var dbStartDate time.Time
	var dbEndDate *time.Time

	err := r.db.QueryRow(ctx, query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&dbStartDate,
		&dbEndDate,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		slog.Error("database error", slog.String("op", op), slog.String("err", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	sub.StartDate = domain.CustomDate{Time: dbStartDate}
	if dbEndDate != nil {
		sub.EndDate = &domain.CustomDate{Time: *dbEndDate}
	}

	return &sub, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub *domain.Subscription) error {
	const op = "repository.postgres.Update"

	query := `
		UPDATE subscriptions 
		SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5
		WHERE id = $6
	`

	var endDate interface{}
	if sub.EndDate != nil {
		endDate = sub.EndDate.Time
	} else {
		endDate = nil
	}

	cmdTag, err := r.db.Exec(ctx, query, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate.Time, endDate, sub.ID)
	if err != nil {
		slog.Error("database error", slog.String("op", op), slog.String("err", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "repository.postgres.Delete"

	query := `DELETE FROM subscriptions WHERE id = $1`

	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		slog.Error("database error", slog.String("op", op), slog.String("err", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *SubscriptionRepository) CalculateTotal(ctx context.Context, filter domain.Filter) (int64, error) {
	const op = "repository.postgres.CalculateTotal"

	query := `SELECT COALESCE(SUM(price), 0) FROM subscriptions WHERE 1=1`
	var args []interface{}
	argID := 1

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argID)
		args = append(args, *filter.UserID)
		argID++
	}

	if filter.ServiceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", argID)
		args = append(args, *filter.ServiceName)
		argID++
	}

	if filter.FromDate != nil {
		query += fmt.Sprintf(" AND (end_date IS NULL OR end_date >= $%d)", argID)
		args = append(args, filter.FromDate.Time)
		argID++
	}

	if filter.ToDate != nil {
		query += fmt.Sprintf(" AND start_date <= $%d", argID)
		args = append(args, filter.ToDate.Time)
		argID++
	}

	var total int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		slog.Error("database error during aggregation", slog.String("op", op), slog.String("err", err.Error()))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return total, nil
}
