package rider

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

func (r *Repository) Create(ctx context.Context, rider *Rider) error {
	query := `
		INSERT INTO riders (name, phone, status)
		VALUES ($1, $2, $3)
		RETURNING id, updated_at, created_at
	`

	err := r.db.QueryRowxContext(ctx, query, rider.Name, rider.Phone, rider.Status).Scan(
		&rider.ID, &rider.UpdatedAt, &rider.CreatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Rider, error) {
	var rider Rider
	query := `SELECT * FROM riders WHERE id = $1`

	err := r.db.GetContext(ctx, &rider, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rider, nil
}

func (r *Repository) UpdateLocation(ctx context.Context, id string, lat float64, lng float64) error {
	query := `UPDATE riders SET current_lat = $1, current_lng = $2, updated_at = now() WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, lat, lng, id)
	return err
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE riders SET status = $1, updated_at = now() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *Repository) ListAvailable(ctx context.Context) ([]Rider, error) {
	var riders []Rider
	query := `SELECT * FROM riders WHERE status = 'available'`

	err := r.db.SelectContext(ctx, &riders, query)
	if err != nil {
		return nil, err
	}
	return riders, nil
}

func (r *Repository) SetBusy(ctx context.Context, id string) error {
	query := `UPDATE riders SET status = 'busy', updated_at = now() WHERE id = $1 AND status = 'available'`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("rider is not available for assignment")
	}
	return nil
}

func (r *Repository) SetAvailable(ctx context.Context, id string) error {
	query := `UPDATE riders SET status = 'available', updated_at = now() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *Repository) ListAll(ctx context.Context) ([]Rider, error) {
	var riders []Rider
	query := `SELECT * FROM riders ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &riders, query)
	if err != nil {
		return nil, err
	}
	return riders, nil
}