package order

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

func (r *Repository) Create(ctx context.Context, order *Order) error {
	query := `
		INSERT INTO orders (customer_name, pickup_lat, pickup_lng, dropoff_lat, dropoff_lng)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, status, created_at
	`

	err := r.db.QueryRowxContext(ctx, query, order.CustomerName, order.PickupLat, order.PickupLng, order.DropoffLat, order.DropoffLng).Scan(
		&order.ID, &order.Status, &order.CreatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Order, error) {
	var order Order
	query := `SELECT * FROM orders WHERE id = $1`

	err := r.db.GetContext(ctx, &order, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *Repository) List(ctx context.Context, status string) ([]Order, error) {
	var orders []Order

	if status == "" {
		query := `SELECT * FROM orders ORDER BY created_at DESC`
		err := r.db.SelectContext(ctx, &orders, query)
		if err != nil {
			return nil, err
		}
		return orders, nil
	}

	query := `SELECT * FROM orders WHERE status = $1 ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &orders, query, status)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM orders WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}