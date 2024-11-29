package storage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bool64/sqluct"
	"github.com/dohernandez/vio/internal/domain/model"
	"github.com/dohernandez/vio/internal/platform/helpers"
	"github.com/dohernandez/vio/internal/platform/storage"
	"github.com/dohernandez/vio/pkg/database"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestGeolocation_SaveGeolocation_success(t *testing.T) {
	t.Parallel()

	// Load sample data
	// 200.106.141.15,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346
	data, err := helpers.LoadSampleData(1, 0)
	require.NoError(t, err)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	geo, err := model.DecodeGeolocation(data[0])
	require.NoError(t, err)

	mock.ExpectExec(`
		INSERT INTO geolocation (ip_address,country_code,country,city,latitude,longitude,mystery_value) 
			VALUES ($1,$2,$3,$4,$5,$6,$7)
		`).
		WithArgs(
			geo.IPAddress,
			geo.CountryCode,
			geo.Country,
			geo.City,
			geo.Latitude,
			geo.Longitude,
			geo.MysteryValue,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	st := sqluct.NewStorage(sqlx.NewDb(db, "sqlmock"))

	s := storage.NewGeolocation(st)

	err = s.SaveGeolocation(context.Background(), []*model.Geolocation{&geo})
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGeolocation_SaveGeolocation_failure(t *testing.T) {
	t.Parallel()

	// Load sample data
	// 200.106.141.15,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346
	data, err := helpers.LoadSampleData(1, 0)
	require.NoError(t, err)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	geo, err := model.DecodeGeolocation(data[0])
	require.NoError(t, err)

	mock.ExpectExec(`
		INSERT INTO geolocation (ip_address,country_code,country,city,latitude,longitude,mystery_value) 
			VALUES ($1,$2,$3,$4,$5,$6,$7)
		`).
		WithArgs(
			geo.IPAddress,
			geo.CountryCode,
			geo.Country,
			geo.City,
			geo.Latitude,
			geo.Longitude,
			geo.MysteryValue,
		).
		WillReturnError(errors.New("error"))

	st := sqluct.NewStorage(sqlx.NewDb(db, "sqlmock"))

	s := storage.NewGeolocation(st)

	err = s.SaveGeolocation(context.Background(), []*model.Geolocation{&geo})
	require.Error(t, err)
	require.ErrorContains(t, err, "error")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGeolocation_SaveGeolocation_batch(t *testing.T) {
	t.Parallel()

	// Load sample data
	data, err := helpers.LoadSampleData(3, 0)
	require.NoError(t, err)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	var geos []*model.Geolocation

	geo1, err := model.DecodeGeolocation(data[0])
	require.NoError(t, err)

	geos = append(geos, &geo1)

	geo2, err := model.DecodeGeolocation(data[1])
	require.NoError(t, err)

	geos = append(geos, &geo2)

	geo3, err := model.DecodeGeolocation(data[2])
	require.NoError(t, err)

	geos = append(geos, &geo3)

	mock.ExpectExec(`
		INSERT INTO geolocation (ip_address,country_code,country,city,latitude,longitude,mystery_value) 
			VALUES ($1,$2,$3,$4,$5,$6,$7),($8,$9,$10,$11,$12,$13,$14),($15,$16,$17,$18,$19,$20,$21)
		`).
		WithArgs(
			geo1.IPAddress,
			geo1.CountryCode,
			geo1.Country,
			geo1.City,
			geo1.Latitude,
			geo1.Longitude,
			geo1.MysteryValue,
			geo2.IPAddress,
			geo2.CountryCode,
			geo2.Country,
			geo2.City,
			geo2.Latitude,
			geo2.Longitude,
			geo2.MysteryValue,
			geo3.IPAddress,
			geo3.CountryCode,
			geo3.Country,
			geo3.City,
			geo3.Latitude,
			geo3.Longitude,
			geo3.MysteryValue,
		).
		WillReturnResult(sqlmock.NewResult(3, 3))

	st := sqluct.NewStorage(sqlx.NewDb(db, "sqlmock"))

	s := storage.NewGeolocation(st)

	err = s.SaveGeolocation(context.Background(), geos)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGeolocation_FindGeolocationByIP_success(t *testing.T) {
	t.Parallel()

	// Load sample data
	// 200.106.141.15,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346
	data, err := helpers.LoadSampleData(1, 0)
	require.NoError(t, err)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	geo, err := model.DecodeGeolocation(data[0])
	require.NoError(t, err)

	meQuery := mock.ExpectQuery(`
				SELECT ip_address, country_code, country, city, latitude, longitude, mystery_value 
				FROM geolocation
				WHERE ip_address = $1
			`).
		WithArgs(
			geo.IPAddress,
		)

	rows := sqlmock.NewRows([]string{"ip_address", "country_code", "country", "city", "latitude", "longitude", "mystery_value"})

	rows.AddRow(
		geo.IPAddress,
		geo.CountryCode,
		geo.Country,
		geo.City,
		geo.Latitude,
		geo.Longitude,
		geo.MysteryValue,
	)

	meQuery.WillReturnRows(rows)

	st := sqluct.NewStorage(sqlx.NewDb(db, "sqlmock"))

	s := storage.NewGeolocation(st)

	g, err := s.FindGeolocationByIP(context.Background(), geo.IPAddress)
	require.NoError(t, err)

	require.Equal(t, geo, g)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGeolocation_FindGeolocationByIP_failure(t *testing.T) {
	t.Parallel()

	ip := "200.106.141.15"

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	_ = mock.ExpectQuery(`
				SELECT ip_address, country_code, country, city, latitude, longitude, mystery_value 
				FROM geolocation
				WHERE ip_address = $1
			`).
		WithArgs(
			ip,
		).
		WillReturnError(database.ErrNotFound)

	st := sqluct.NewStorage(sqlx.NewDb(db, "sqlmock"))

	s := storage.NewGeolocation(st)

	_, err = s.FindGeolocationByIP(context.Background(), ip)
	require.Error(t, err)
	require.ErrorContains(t, err, "not found")

	require.NoError(t, mock.ExpectationsWereMet())
}
