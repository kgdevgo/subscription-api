package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"subscription-api/internal/domain"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var (
	ErrInvalidInput = errors.New("invalid input data")
	ErrDateConflict = errors.New("end date cannot be before start date")
)

type SubscriptionUseCase struct {
	repo     domain.SubscriptionRepository
	validate *validator.Validate
}

func NewSubscriptionUseCase(repo domain.SubscriptionRepository) *SubscriptionUseCase {
	return &SubscriptionUseCase{
		repo:     repo,
		validate: validator.New(),
	}
}

func (uc *SubscriptionUseCase) Create(ctx context.Context, sub *domain.Subscription) error {
	const op = "usecase.subscription.Create"

	if err := uc.validate.Struct(sub); err != nil {
		slog.Warn("validation failed", slog.String("op", op), slog.String("err", err.Error()))
		return fmt.Errorf("%w: %s", ErrInvalidInput, err.Error())
	}

	if sub.EndDate != nil && sub.EndDate.Before(sub.StartDate.Time) {
		return ErrDateConflict
	}

	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}

	return uc.repo.Create(ctx, sub)
}

func (uc *SubscriptionUseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidInput
	}
	return uc.repo.GetByID(ctx, id)
}

func (uc *SubscriptionUseCase) Update(ctx context.Context, sub *domain.Subscription) error {
	const op = "usecase.subscription.Update"

	if err := uc.validate.Struct(sub); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidInput, err.Error())
	}

	if sub.EndDate != nil && sub.EndDate.Before(sub.StartDate.Time) {
		return ErrDateConflict
	}

	return uc.repo.Update(ctx, sub)
}

func (uc *SubscriptionUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidInput
	}
	return uc.repo.Delete(ctx, id)
}

func (uc *SubscriptionUseCase) CalculateTotal(ctx context.Context, filter domain.Filter) (int64, error) {
	if filter.FromDate != nil && filter.ToDate != nil && filter.ToDate.Before(filter.FromDate.Time) {
		return 0, fmt.Errorf("%w: to_date cannot be before from_date", ErrInvalidInput)
	}
	return uc.repo.CalculateTotal(ctx, filter)
}
