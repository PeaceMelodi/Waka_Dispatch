package dispatch

import (
	"context"
	"errors"
	"math"

	"github.com/PeaceMelodi/waka-dispatch/internal/rider"
)

func Haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371 // kilometers

	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLng := (lng2 - lng1) * (math.Pi / 180)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180))*math.Cos(lat2*(math.Pi/180))*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

type Matcher struct {
	riderRepo *rider.Repository
}

func NewMatcher(riderRepo *rider.Repository) *Matcher {
	return &Matcher{riderRepo: riderRepo}
}

func (m *Matcher) FindNearestRider(ctx context.Context, pickupLat, pickupLng float64) (*rider.Rider, error) {
	availableRiders, err := m.riderRepo.ListAvailable(ctx)
	if err != nil {
		return nil, err
	}

	if len(availableRiders) == 0 {
		return nil, errors.New("no available riders found")
	}

	var nearestRider *rider.Rider
	minDistance := math.Inf(1)

	for i := range availableRiders {
		r := &availableRiders[i]
		if r.CurrentLat == nil || r.CurrentLng == nil {
			continue 
		}

		distance := Haversine(pickupLat, pickupLng, *r.CurrentLat, *r.CurrentLng)
		if distance < minDistance {
			minDistance = distance
			nearestRider = r
		}
	}

	if nearestRider == nil {
		return nil, errors.New("no riders with location available")
	}

	return nearestRider, nil
}