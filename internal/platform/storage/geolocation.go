package storage

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/bool64/ctxd"
	"github.com/bool64/sqluct"
	"github.com/dohernandez/vio/internal/domain/model"
	"github.com/dohernandez/vio/pkg/database"
	"github.com/dohernandez/vio/pkg/database/pgx"
)

// GeolocationTable is the table name for geolocation.
const GeolocationTable = "geolocation"

// Geolocation represents a Geolocation repository.
type Geolocation struct {
	storage *sqluct.Storage

	colIPAddress string
}

// NewGeolocation returns instance of Geolocation repository.
func NewGeolocation(storage *sqluct.Storage) *Geolocation {
	var geoLocation model.Geolocation

	return &Geolocation{
		storage:      storage,
		colIPAddress: storage.Mapper.Col(&geoLocation, &geoLocation.IPAddress),
	}
}

// SaveGeolocation store the geolocation data.
func (s *Geolocation) SaveGeolocation(ctx context.Context, geos []*model.Geolocation) error {
	errMsg := "storage.Geolocation: failed to save Geolocation"

	q := s.storage.InsertStmt(GeolocationTable, geos)

	_, err := s.storage.Exec(ctx, q)
	if err == nil {
		return nil
	}

	if pgx.IsUniqueViolation(err) {
		return ctxd.WrapError(ctx, database.ErrAlreadyExists, errMsg)
	}

	return ctxd.WrapError(ctx, err, errMsg)
}

// FindGeolocationByIP get the geolocation data by IP.
func (s *Geolocation) FindGeolocationByIP(ctx context.Context, ip string) (model.Geolocation, error) {
	errMsg := "storage.Geolocation: failed to get Geolocation by IP"

	var geo model.Geolocation

	q := s.storage.SelectStmt(GeolocationTable, geo).
		Where(squirrel.Eq{s.colIPAddress: ip})

	err := s.storage.Select(ctx, q, &geo)
	if err != nil {
		if pgx.IsNoRows(err) {
			return geo, ctxd.WrapError(ctx, database.ErrNotFound, errMsg)
		}

		return geo, ctxd.WrapError(ctx, err, errMsg)
	}

	return geo, nil
}
