package usecase

import (
	"context"
	"testing"

	"github.com/dohernandez/vio/internal/domain/model"
	"github.com/dohernandez/vio/internal/domain/usecase/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGeolocationByIPExposer_ExposeGeolocationByIP_success(t *testing.T) {
	t.Parallel()

	ip := "200.106.141.15"

	finder := mocks.NewGeolocationByIPFinder(t)
	finder.EXPECT().FindGeolocationByIP(mock.Anything, ip).Return(model.Geolocation{IPAddress: ip}, nil)

	exposer := NewGeolocationByIPExposer(finder)

	geo, err := exposer.ExposeGeolocationByIP(context.Background(), ip)
	require.NoError(t, err)
	require.Equal(t, model.Geolocation{IPAddress: ip}, geo)
}
