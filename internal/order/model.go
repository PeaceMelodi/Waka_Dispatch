package order

import "time"

type Order struct {
	ID          string    `db:"id" json:"id"`
	CustomerName string   `db:"customer_name" json:"customer_name"`
	PickupLat   float64   `db:"pickup_lat" json:"pickup_lat"`
	PickupLng   float64   `db:"pickup_lng" json:"pickup_lng"`
	DropoffLat  float64   `db:"dropoff_lat" json:"dropoff_lat"`
	DropoffLng  float64   `db:"dropoff_lng" json:"dropoff_lng"`
	Status      string    `db:"status" json:"status"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}