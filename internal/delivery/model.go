package delivery

import "time"

type Delivery struct {
	ID          string     `db:"id" json:"id"`
	OrderID     string     `db:"order_id" json:"order_id"`
	RiderID     string     `db:"rider_id" json:"rider_id"`
	AssignedAt  time.Time  `db:"assigned_at" json:"assigned_at"`
	PickedUpAt  *time.Time `db:"picked_up_at" json:"picked_up_at,omitempty"`
	InTransitAt *time.Time `db:"in_transit_at" json:"in_transit_at,omitempty"`
	DeliveredAt *time.Time `db:"delivered_at" json:"delivered_at,omitempty"`
	CancelledAt *time.Time `db:"cancelled_at" json:"cancelled_at,omitempty"`
}