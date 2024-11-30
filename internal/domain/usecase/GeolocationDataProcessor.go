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
func (p *GeolocationDataProcessor) Process(ctx context.Context, reader GeolocationDataReader, inParallel uint) error {
	startTime := time.Now()

	data, err := reader.ReadGeolocationData(ctx)
	if err != nil {
		return err
	}

	eg, egctx := errgroup.WithContext(ctx)

	var (
		report = &reporter{
			eg: eg,
		}
		dupl = &duplication{}

		processWorker = inParallel
		saverWorker   = 15
		// ready is a channel to send geolocation data to be saved.
		// The buffer size is twice the number of saver workers.
		// This is to ensure that the processor worker can continue to process the data while the saver worker(s) is/are
		// still saving the data.
		ready = make(chan *model.Geolocation, batchBuffer*saverWorker*2)
	)

	// Start the saver worker(s).
	eg.Go(func() error {
		var wg sync.WaitGroup

		for range saverWorker {
			wg.Add(1)

			go func() {
				defer wg.Done()

				p.save(egctx, ready, report)
			}()
		}

		wg.Wait()

		return nil
	})

	// Start the processor worker(s).
	eg.Go(func() error {
		defer func() {
			close(ready)
		}()

		var wg sync.WaitGroup

		for range processWorker {
			wg.Add(1)

			go func() {
				defer wg.Done()

				p.process(egctx, data, ready, dupl, report)
			}()
		}

		wg.Wait()

		return nil
	})

	if err := eg.Wait(); err != nil {
		return ctxd.NewError(ctx, "processing geolocation data", "error", err)
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

func (p *GeolocationDataProcessor) process(
	ctx context.Context,
	data <-chan []string,
	ready chan<- *model.Geolocation,
	dupl *duplication,
	r *reporter,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-data:
			if !ok {
				return
			}

			geo, err := model.DecodeGeolocation(d) //nolint:contextcheck
			if err != nil {
				r.failed(err)

				p.logger.Debug(ctx, "decode geolocation data", "error", err)

				continue
			}

			// Validate geolocation data before saving.
			if err = geo.IsValid(); err != nil { //nolint:contextcheck
				r.failed(err)

				p.logger.Debug(ctx, "invalid geolocation data", "error", err)

				continue
			}

			err = dupl.check(&geo)
			if err != nil {
				r.failed(err)

				p.logger.Debug(ctx, "geolocation data exists", "error", err)

				continue
			}

			ready <- &geo
		}
	}
}

func (p *GeolocationDataProcessor) save(
	ctx context.Context,
	ready <-chan *model.Geolocation,
	r *reporter,
) {
	buf := make([]*model.Geolocation, 0, batchBuffer)

	defer func() {
		if len(buf) == 0 {
			return
		}

		if err := p.storage.SaveGeolocation(ctx, buf); err != nil {
			r.failed(err)

			p.logger.Debug(ctx, "save geolocation data", "error", err)

			return
		}

		r.succeed(len(buf))
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case geo, ok := <-ready:
			if !ok {
				return
			}

			buf = append(buf, geo)

			if len(buf) < batchBuffer {
				continue
			}
		}

		if err := p.storage.SaveGeolocation(ctx, buf); err != nil {
			r.failed(err)

			p.logger.Debug(ctx, "save geolocation data", "error", err)

			continue
		}

		r.succeed(len(buf))

		// Reset buffer
		buf = make([]*model.Geolocation, 0, batchBuffer)
	}
}

// reporter is a helper to report the processing result.
type reporter struct {
	accepted         int
	discarded        int
	discardedReasons map[string]uint

	// eg...
	eg *errgroup.Group

	smA sync.Mutex
	smD sync.Mutex
}

func (r *reporter) succeed(a int) {
	r.eg.Go(func() error {
		r.smA.Lock()
		defer r.smA.Unlock()

		r.accepted += a

		return nil
	})
}

func (r *reporter) failed(err error) {
	r.eg.Go(func() error {
		r.smD.Lock()
		defer r.smD.Unlock()

		r.discarded++

		if r.discardedReasons == nil {
			r.discardedReasons = make(map[string]uint)
		}

		errMsg := err.Error()

		r.discardedReasons[errMsg]++

		return nil
	})
}

// duplication is a helper to check the duplication of geolocation data loaded.
// It keeps the uniqueness of the geolocation data by IP address.
type duplication struct {
	uniqueness map[string]bool

	sm sync.Mutex
}

func (d *duplication) check(geo *model.Geolocation) error {
	d.sm.Lock()
	defer d.sm.Unlock()

	if d.uniqueness == nil {
		d.uniqueness = make(map[string]bool)
	}

	if _, ok := d.uniqueness[geo.IPAddress]; ok {
		return model.ErrGeolocationAlreadyExists
	}

	d.uniqueness[geo.IPAddress] = true

	return nil
}
