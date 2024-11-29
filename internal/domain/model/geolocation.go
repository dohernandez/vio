package model

import (
	"context"
	"errors"
	"net"
	"strconv"

	"github.com/bool64/ctxd"
)

const (
	// InputFieldNum is the number of fields in the input data. It is used to validate the input data before normalizing.
	InputFieldNum = 7
)

// Geolocation errors.
var (
	// ErrGeolocationNotFound is returned when a geolocation is not found.
	ErrGeolocationNotFound      = errors.New("geolocation not found")
	ErrGeolocationAlreadyExists = errors.New("geolocation already exists")
)

// Geolocation represents a geolocation entity.
type Geolocation struct {
	IPAddress    string  `db:"ip_address"`
	CountryCode  string  `db:"country_code"`
	Country      string  `db:"country"`
	City         string  `db:"city"`
	Latitude     float64 `db:"latitude"`
	Longitude    float64 `db:"longitude"`
	MysteryValue float64 `db:"mystery_value"`
}

// DecodeGeolocation normalizes the input data into a geolocation entity.
func DecodeGeolocation(data []string) (Geolocation, error) {
	var geo Geolocation

	if len(data) != InputFieldNum {
		return geo, ctxd.NewError(context.Background(), "not enough fields in input", "input", data, "expected", InputFieldNum, "actual", len(data))
	}

	geo.IPAddress = data[0]
	geo.CountryCode = data[1]
	geo.Country = data[2]
	geo.City = data[3]

	var err error

	geo.Latitude, err = strconv.ParseFloat(data[4], 64)
	if err != nil {
		return geo, ctxd.NewError(context.Background(), "parsing latitude", "latitude", data[4], "error", err)
	}

	geo.Longitude, err = strconv.ParseFloat(data[5], 64)
	if err != nil {
		return geo, ctxd.NewError(context.Background(), "parsing longitude", "longitude", data[5], "error", err)
	}

	geo.MysteryValue, err = strconv.ParseFloat(data[6], 64)
	if err != nil {
		return geo, ctxd.NewError(context.Background(), "parsing mystery value", "mystery_value", data[6], "error", err)
	}

	return geo, nil
}

// IsValid validates the geolocation entity.
func (g Geolocation) IsValid() error {
	if g.IPAddress == "" {
		return ctxd.NewError(context.Background(), "missing ip address")
	}

	if net.ParseIP(g.IPAddress) == nil {
		return ctxd.NewError(context.Background(), "invalid ip address", "ip_address", g.IPAddress)
	}

	if g.CountryCode == "" {
		return ctxd.NewError(context.Background(), "missing country code")
	}

	if len(g.CountryCode) != 2 {
		return ctxd.NewError(context.Background(), "invalid country code length", "country_code", g.CountryCode)
	}

	if g.Country == "" {
		return ctxd.NewError(context.Background(), "missing country")
	}

	if g.City == "" {
		return ctxd.NewError(context.Background(), "missing city")
	}

	if g.Latitude < -90 || g.Latitude > 90 {
		return ctxd.NewError(context.Background(), "invalid latitude", "latitude", g.Latitude)
	}

	if g.Longitude < -180 || g.Longitude > 180 {
		return ctxd.NewError(context.Background(), "invalid longitude", "longitude", g.Longitude)
	}

	return nil
}
