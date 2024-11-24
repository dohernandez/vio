package usecase

import (
	"context"

	"github.com/bool64/ctxd"
	"github.com/dohernandez/vio/internal/domain/model"
)

//go:generate mockery --name=GeolocationByIPFinder --outpkg=mocks --output=mocks --filename=geolocation_by_ip_finder.go --with-expecter

// GeolocationByIPFinder is the interface that provides the ability to find geolocation by IP.
//
// Returns ErrGeolocationNotFound if the geolocation is not found.
type GeolocationByIPFinder interface {
	FindGeolocationByIP(ctx context.Context, ip string) (model.Geolocation, error)
}

// GeolocationByIPExposer exposes the geolocation by IP.
type GeolocationByIPExposer struct {
	finder GeolocationByIPFinder
}

// NewGeolocationByIPExposer creates a new GeolocationByIPExposer.
func NewGeolocationByIPExposer(finder GeolocationByIPFinder) *GeolocationByIPExposer {
	return &GeolocationByIPExposer{
		finder: finder,
	}
}

// ExposeGeolocationByIP exposes the geolocation by IP.
//
// Returns ErrGeolocationNotFound if the geolocation is not found.
func (e *GeolocationByIPExposer) ExposeGeolocationByIP(ctx context.Context, ip string) (model.Geolocation, error) {
	geo, err := e.finder.FindGeolocationByIP(ctx, ip)
	if err != nil {
		return model.Geolocation{}, ctxd.WrapError(ctx, model.ErrGeolocationNotFound, err.Error())
	}

	return geo, nil
}
