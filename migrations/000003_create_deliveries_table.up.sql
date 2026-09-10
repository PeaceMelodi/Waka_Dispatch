-- Join table linking an order to its assigned rider, with lifecycle timestamps
CREATE TABLE deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id),
    rider_id UUID NOT NULL REFERENCES riders(id),
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    picked_up_at TIMESTAMPTZ,
    in_transit_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ
);