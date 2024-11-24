package model

import (
	"testing"

	"github.com/dohernandez/vio/internal/platform/helpers"
	"github.com/stretchr/testify/require"
)

func TestDecodeGeolocation_success(t *testing.T) {
	t.Parallel()

	// Load sample data
	// 200.106.141.15,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346
	data, err := helpers.LoadSampleData(1, 0)
	require.NoError(t, err)

	geolocation, err := DecodeGeolocation(data[0])
	require.NoError(t, err)

	require.Equal(t, "200.106.141.15", geolocation.IPAddress)
	require.Equal(t, "SI", geolocation.CountryCode)
	require.Equal(t, "Nepal", geolocation.Country)
	require.Equal(t, "DuBuquemouth", geolocation.City)
	require.InEpsilon(t, -84.87503094689836, geolocation.Latitude, 0)
	require.InEpsilon(t, 7.206435933364332, geolocation.Longitude, 0)
	require.InEpsilon(t, float64(7823011346), geolocation.MysteryValue, 0)
}

func TestDecodeGeolocation_error_no_enough(t *testing.T) {
	t.Parallel()

	// Load sample data
	// 200.106.141.15,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346
	data, err := helpers.LoadSampleData(1, 0)
	require.NoError(t, err)

	// Test with missing no enough field
	_, err = DecodeGeolocation(data[0][:4])
	require.Error(t, err)
}

func TestDecodeGeolocation_error_latitude(t *testing.T) {
	t.Parallel()

	// Load sample data
	// 200.106.141.15,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346
	data, err := helpers.LoadSampleData(1, 0)
	require.NoError(t, err)

	// Test with missing longitude field
	data[0][4] = ""

	_, err = DecodeGeolocation(data[0])
	require.Error(t, err)
	require.ErrorContains(t, err, "parsing latitude")
}

func TestDecodeGeolocation_error_longitude(t *testing.T) {
	t.Parallel()

	// Load sample data
	// 200.106.141.15,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346
	data, err := helpers.LoadSampleData(1, 0)
	require.NoError(t, err)

	// Test with missing longitude field
	data[0][5] = ""

	_, err = DecodeGeolocation(data[0])
	require.Error(t, err)
	require.ErrorContains(t, err, "parsing longitude")
}

func TestGeolocation_IsValid_success(t *testing.T) {
	t.Parallel()

	// Load sample data
	// 200.106.141.15,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346
	data, err := helpers.LoadSampleData(1, 0)
	require.NoError(t, err)

	geolocation, err := DecodeGeolocation(data[0])
	require.NoError(t, err)

	require.NoError(t, geolocation.IsValid())
}

func TestGeolocation_IsValid_failure(t *testing.T) {
	t.Parallel()

	// Load sample data
	// 200.106.141.15,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346
	data, err := helpers.LoadSampleData(1, 3)
	require.NoError(t, err)

	geolocation, err := DecodeGeolocation(data[0])
	require.NoError(t, err)

	err = geolocation.IsValid()
	require.Error(t, err)
	require.ErrorContains(t, err, "missing ip address")
}
