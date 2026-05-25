package domain

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const DateFormat = "01-2006"

type CustomDate struct {
	time.Time
}

func (cd *CustomDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		return nil
	}
	t, err := time.Parse(DateFormat, s)
	if err != nil {
		return fmt.Errorf("invalid date format, expected MM-YYYY: %w", err)
	}
	cd.Time = t
	return nil
}

func (cd *CustomDate) MarshalJSON() ([]byte, error) {
	if cd == nil || cd.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", cd.Format(DateFormat))), nil
}

func (cd *CustomDate) Value() (driver.Value, error) {
	if cd == nil || cd.IsZero() {
		return nil, nil
	}
	return cd.Time, nil
}

func (cd *CustomDate) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	t, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("failed to scan CustomDate from value: %v", value)
	}
	cd.Time = t
	return nil
}

type Subscription struct {
	ID          uuid.UUID   `json:"id"`
	ServiceName string      `json:"service_name" validate:"required,min=1,max=255"`
	Price       int64       `json:"price" validate:"required,gte=0"`
	UserID      uuid.UUID   `json:"user_id" validate:"required"`
	StartDate   CustomDate  `json:"start_date" validate:"required"`
	EndDate     *CustomDate `json:"end_date,omitempty"`
}

type Filter struct {
	UserID      *uuid.UUID
	ServiceName *string
	FromDate    *CustomDate
	ToDate      *CustomDate
}

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*Subscription, error)
	Update(ctx context.Context, sub *Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	CalculateTotal(ctx context.Context, filter Filter) (int64, error)
}

type SubscriptionUseCase interface {
	Create(ctx context.Context, sub *Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*Subscription, error)
	Update(ctx context.Context, sub *Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	CalculateTotal(ctx context.Context, filter Filter) (int64, error)
}
