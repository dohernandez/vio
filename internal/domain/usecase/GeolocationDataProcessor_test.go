package usecase

import (
	"context"
	"reflect"
	"testing"

	"github.com/bool64/ctxd"
	"github.com/dohernandez/vio/internal/domain/model"
	"github.com/dohernandez/vio/internal/domain/usecase/mocks"
	"github.com/dohernandez/vio/internal/platform/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGeolocationDataProcessor_Process_all_success(t *testing.T) {
	t.Parallel()

	// Load sample data
	data, err := helpers.LoadSampleData(3, 0)
	require.NoError(t, err)

	// reader
	dataCh := make(chan []string, len(data))

	reader := mocks.NewGeolocationDataReader(t)
	reader.EXPECT().ReadGeolocationData(mock.Anything).Run(func(_ context.Context) {
		go func() {
			for _, d := range data {
				dataCh <- d
			}
			close(dataCh)
		}()
	}).Return(dataCh, nil)

	// storage
	storage := mocks.NewGeolocationDataStorage(t)
	storage.EXPECT().SaveGeolocation(mock.Anything, mock.AnythingOfType(reflect.TypeOf(model.Geolocation{}).String())).Return(nil).Times(len(data))

	logger := &ctxd.LoggerMock{}

	processor := NewParseGeolocationData(storage, logger)

	// Process with 3 parallel processes
	err = processor.Process(context.Background(), reader, 3)
	require.NoError(t, err)

	reportLog := logger.LoggedEntries[len(logger.LoggedEntries)-1]

	assert.Equal(t, "geolocation data processed", reportLog.Message)

	reportLogData := make(map[string]interface{}, len(reportLog.Data)-1)

	for k, v := range reportLog.Data {
		if k == "duration_s" {
			continue
		}

		reportLogData[k] = v
	}

	assert.Equal(t, map[string]interface{}{
		"accepted":          3,
		"discarded":         0,
		"discarded_reasons": map[string]uint(nil),
	}, reportLogData)
}

func TestGeolocationDataProcessor_Process_all_failure(t *testing.T) {
	t.Parallel()

	// Load sample data
	data, err := helpers.LoadAllSampleData()
	require.NoError(t, err)

	// reader
	dataCh := make(chan []string, len(data))

	reader := mocks.NewGeolocationDataReader(t)
	reader.EXPECT().ReadGeolocationData(mock.Anything).Run(func(_ context.Context) {
		go func() {
			for _, d := range data {
				// Not enough fields.
				dataCh <- d[:4]
			}
			close(dataCh)
		}()
	}).Return(dataCh, nil)

	// storage
	storage := mocks.NewGeolocationDataStorage(t)

	logger := &ctxd.LoggerMock{}

	processor := NewParseGeolocationData(storage, logger)

	// Process with 3 parallel processes
	err = processor.Process(context.Background(), reader, 3)
	require.NoError(t, err)

	reportLog := logger.LoggedEntries[len(logger.LoggedEntries)-1]

	assert.Equal(t, "geolocation data processed", reportLog.Message)

	reportLogData := make(map[string]interface{}, len(reportLog.Data)-1)

	for k, v := range reportLog.Data {
		if k == "duration_s" {
			continue
		}

		reportLogData[k] = v
	}

	assert.Equal(t, map[string]interface{}{
		"accepted":          0,
		"discarded":         5,
		"discarded_reasons": map[string]uint{"not enough fields in input": 0x5},
	}, reportLogData)
}

func TestGeolocationDataProcessor_Process(t *testing.T) {
	t.Parallel()

	// Load sample data
	data, err := helpers.LoadAllSampleData()
	require.NoError(t, err)

	// reader
	dataCh := make(chan []string, len(data))

	reader := mocks.NewGeolocationDataReader(t)
	reader.EXPECT().ReadGeolocationData(mock.Anything).Run(func(_ context.Context) {
		go func() {
			for _, d := range data {
				dataCh <- d
			}
			close(dataCh)
		}()
	}).Return(dataCh, nil)

	// storage
	storage := mocks.NewGeolocationDataStorage(t)
	storage.EXPECT().SaveGeolocation(mock.Anything, mock.AnythingOfType(reflect.TypeOf(model.Geolocation{}).String())).Return(nil)

	logger := &ctxd.LoggerMock{}

	processor := NewParseGeolocationData(storage, logger)

	// Process with 3 parallel processes
	err = processor.Process(context.Background(), reader, 3)
	require.NoError(t, err)

	reportLog := logger.LoggedEntries[len(logger.LoggedEntries)-1]

	assert.Equal(t, "geolocation data processed", reportLog.Message)

	reportLogData := make(map[string]interface{}, len(reportLog.Data)-1)

	for k, v := range reportLog.Data {
		if k == "duration_s" {
			continue
		}

		reportLogData[k] = v
	}

	assert.Equal(t, map[string]interface{}{
		"accepted":          4,
		"discarded":         1,
		"discarded_reasons": map[string]uint{"missing ip address": 0x1},
	}, reportLogData)
}
