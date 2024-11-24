package usecase

import (
	"context"
	"sync"
	"time"

	"github.com/bool64/ctxd"
	"github.com/dohernandez/vio/internal/domain/model"
	"golang.org/x/sync/errgroup"
)

//go:generate mockery --name=GeolocationDataReader --outpkg=mocks --output=mocks --filename=geolocation_data_reader.go --with-expecter

// GeolocationDataReader is the interface that provides the ability to read geolocation data.
// It returns a channel of string slice which represents the geolocation item data.
type GeolocationDataReader interface {
	ReadGeolocationData(ctx context.Context) (<-chan []string, error)
}

//go:generate mockery --name=GeolocationDataStorage --outpkg=mocks --output=mocks --filename=geolocation_data_storage.go --with-expecter

// GeolocationDataStorage is the interface that provides the ability to save geolocation data.
type GeolocationDataStorage interface {
	SaveGeolocation(ctx context.Context, geo model.Geolocation) error
}

// GeolocationDataProcessor processes the geolocation data.
type GeolocationDataProcessor struct {
	storage GeolocationDataStorage

	logger ctxd.Logger
}

// NewParseGeolocationData creates a new GeolocationDataProcessor.
func NewParseGeolocationData(storage GeolocationDataStorage, logger ctxd.Logger) *GeolocationDataProcessor {
	return &GeolocationDataProcessor{
		storage: storage,
		logger:  logger,
	}
}

// Process processes the geolocation data from the given reader in parallel.
func (p *GeolocationDataProcessor) Process(ctx context.Context, reader GeolocationDataReader, inParallel uint) error {
	startTime := time.Now()

	data, err := reader.ReadGeolocationData(ctx)
	if err != nil {
		return err
	}

	buf := make(chan []string, inParallel)

	report := &reporter{}

	eg, ctx := errgroup.WithContext(ctx)

	for range inParallel {
		eg.Go(func() error {
			for d := range buf {
				geo, err := model.DecodeGeolocation(d) //nolint:contextcheck
				if err != nil {
					report.failed(err)

					p.logger.Debug(ctx, "failed to decode geolocation data", "error", err)

					continue
				}

				// Validate geolocation data before saving.
				if err = geo.IsValid(); err != nil { //nolint:contextcheck
					report.failed(err)

					p.logger.Debug(ctx, "invalid geolocation data", "error", err)

					continue
				}

				if err := p.storage.SaveGeolocation(ctx, geo); err != nil {
					report.failed(err)

					p.logger.Debug(ctx, "failed to save geolocation data", "error", err)

					continue
				}

				report.succeed()

				p.logger.Debug(ctx, "geolocation data saved", "geolocation", geo)
			}

			return nil
		})
	}

	for d := range data {
		select {
		case <-ctx.Done():
			close(buf)

			return ctx.Err()
		case buf <- d:
		}
	}

	close(buf)

	if err := eg.Wait(); err != nil {
		return ctxd.NewError(context.Background(), "processing geolocation data", "error", err) //nolint:contextcheck
	}

	endTime := time.Since(startTime)

	p.logger.Important(ctx, "geolocation data processed",
		"accepted", report.accepted,
		"discarded", report.discarded,
		"discarded_reasons", report.discardedReasons,
		"duration_s", endTime.Seconds(),
	)

	return nil
}

type reporter struct {
	accepted         int
	discarded        int
	discardedReasons map[string]uint

	smA sync.Mutex
	smD sync.Mutex
}

func (r *reporter) succeed() {
	r.smA.Lock()
	defer r.smA.Unlock()

	r.accepted++
}

func (r *reporter) failed(err error) {
	r.smD.Lock()
	defer r.smD.Unlock()

	r.discarded++

	if r.discardedReasons == nil {
		r.discardedReasons = make(map[string]uint)
	}

	errMsg := err.Error()

	r.discardedReasons[errMsg]++
}
