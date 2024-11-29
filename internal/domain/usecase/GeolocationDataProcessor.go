package usecase

import (
	"context"
	"sync"
	"time"

	"github.com/bool64/ctxd"
	"github.com/dohernandez/vio/internal/domain/model"
	"golang.org/x/sync/errgroup"
)

const batchBuffer = 500

//go:generate mockery --name=GeolocationDataReader --outpkg=mocks --output=mocks --filename=geolocation_data_reader.go --with-expecter

// GeolocationDataReader is the interface that provides the ability to read geolocation data.
// It returns a channel of string slice which represents the geolocation item data.
type GeolocationDataReader interface {
	ReadGeolocationData(ctx context.Context) (<-chan []string, error)
}

//go:generate mockery --name=GeolocationDataStorage --outpkg=mocks --output=mocks --filename=geolocation_data_storage.go --with-expecter

// GeolocationDataStorage is the interface that provides the ability to save geolocation data.
type GeolocationDataStorage interface {
	SaveGeolocation(ctx context.Context, geo []*model.Geolocation) error
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
//
//nolint:funlen,gocognit
func (p *GeolocationDataProcessor) Process(ctx context.Context, reader GeolocationDataReader, inParallel uint) error {
	startTime := time.Now()

	data, err := reader.ReadGeolocationData(ctx)
	if err != nil {
		return err
	}

	buf := make(chan []string, inParallel)

	report := &reporter{}

	eg, ctxt := errgroup.WithContext(ctx)

	var (
		batch = &geoBatch{}
		sm    sync.Mutex
	)

	for range inParallel {
		eg.Go(func() error {
			for d := range buf {
				geo, err := model.DecodeGeolocation(d) //nolint:contextcheck
				if err != nil {
					report.failed(err)

					p.logger.Debug(ctxt, "failed to decode geolocation data", "error", err)

					continue
				}

				// Validate geolocation data before saving.
				if err = geo.IsValid(); err != nil { //nolint:contextcheck
					report.failed(err)

					p.logger.Debug(ctxt, "invalid geolocation data", "error", err)

					continue
				}

				sm.Lock()

				if err = batch.add(&geo); err != nil {
					sm.Unlock()

					report.failed(err)

					p.logger.Debug(ctxt, "failed to add geolocation data to batch", "error", err)

					continue
				}

				report.succeed()

				size := batch.len()

				if size < batchBuffer {
					sm.Unlock()

					continue
				}

				if err := p.storage.SaveGeolocation(ctxt, batch.get()); err != nil {
					report.failed(err)

					p.logger.Debug(ctxt, "failed to save geolocation data", "error", err)

					continue
				}

				p.logger.Debug(ctxt, "geolocation data saved", "geolocation", size)

				batch.reset()

				sm.Unlock()
			}

			return nil
		})
	}

	for d := range data {
		select {
		case <-ctxt.Done():
			close(buf)

			return ctxt.Err()
		case buf <- d:
		}
	}

	close(buf)

	if err := eg.Wait(); err != nil {
		return ctxd.NewError(ctx, "processing geolocation data", "error", err)
	}

	if batch.len() > 0 {
		if err := p.storage.SaveGeolocation(ctx, batch.get()); err != nil {
			report.failed(err)

			p.logger.Debug(ctxt, "failed to save geolocation data", "error", err)
		}
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

// geoBatch is a buffer for geolocation data.
// goeBatch is not thread-safe, therefore it should be protected by a mutex when used in a concurrent environment.
type geoBatch struct {
	goes []*model.Geolocation

	uniqueness map[string]bool
}

func (b *geoBatch) add(geo *model.Geolocation) error {
	if b.uniqueness == nil {
		b.uniqueness = make(map[string]bool)
	}

	if _, ok := b.uniqueness[geo.IPAddress]; ok {
		return model.ErrGeolocationAlreadyExists
	}

	b.uniqueness[geo.IPAddress] = true

	b.goes = append(b.goes, geo)

	return nil
}

func (b *geoBatch) len() int {
	return len(b.goes)
}

func (b *geoBatch) reset() {
	b.goes = make([]*model.Geolocation, 0, batchBuffer)
}

func (b *geoBatch) get() []*model.Geolocation {
	return b.goes
}
