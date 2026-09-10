package delivery

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, delivery *Delivery) error {
	query := `
		INSERT INTO deliveries (order_id, rider_id)
		VALUES ($1, $2)
		RETURNING id, assigned_at
	`

	err := r.db.QueryRowxContext(ctx, query, delivery.OrderID, delivery.RiderID).Scan(
		&delivery.ID, &delivery.AssignedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Delivery, error) {
	var delivery Delivery
	query := `SELECT * FROM deliveries WHERE id = $1`

	err := r.db.GetContext(ctx, &delivery, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (r *Repository) GetByOrderID(ctx context.Context, orderID string) (*Delivery, error) {
	var delivery Delivery
	query := `SELECT * FROM deliveries WHERE order_id = $1`

	err := r.db.GetContext(ctx, &delivery, query, orderID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, status string) error {
	switch status {
	case "picked_up":
		query := `UPDATE deliveries SET picked_up_at = now() WHERE id = $1`
		_, err := r.db.ExecContext(ctx, query, id)
		return err
	case "in_transit":
		query := `UPDATE deliveries SET in_transit_at = now() WHERE id = $1`
		_, err := r.db.ExecContext(ctx, query, id)
		return err
	case "delivered":
		query := `UPDATE deliveries SET delivered_at = now() WHERE id = $1`
		_, err := r.db.ExecContext(ctx, query, id)
		return err
	case "cancelled":
		query := `UPDATE deliveries SET cancelled_at = now() WHERE id = $1`
		_, err := r.db.ExecContext(ctx, query, id)
		return err
	default:
		return errors.New("invalid delivery status")
	}
}

func (r *Repository) ListActive(ctx context.Context) ([]Delivery, error) {
	var deliveries []Delivery
	query := `
		SELECT * FROM deliveries 
		WHERE delivered_at IS NULL AND cancelled_at IS NULL
		ORDER BY assigned_at DESC
	`

	err := r.db.SelectContext(ctx, &deliveries, query)
	if err != nil {
		return nil, err
	}
	return deliveries, nil
}