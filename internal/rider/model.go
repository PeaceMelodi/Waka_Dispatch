package rider

import "time"

type Rider struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Phone       string    `db:"phone" json:"phone"`
	Status      string    `db:"status" json:"status"`
	CurrentLat  *float64  `db:"current_lat" json:"current_lat,omitempty"`
	CurrentLng  *float64  `db:"current_lng" json:"current_lng,omitempty"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}