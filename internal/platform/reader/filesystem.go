package reader

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"os"

	"github.com/bool64/ctxd"
)

// dataChBuf is the buffer size for the data channel.
const dataChBuf = 1000

// FileSystem is a storage that save/loads data to/from a file.
type FileSystem struct {
	file string

	logger ctxd.Logger
}

// NewFileSystem creates a new file storage.
func NewFileSystem(file string, logger ctxd.Logger) *FileSystem {
	return &FileSystem{
		file:   file,
		logger: logger,
	}
}

// ReadGeolocationData reads geolocation data from a file.
func (f *FileSystem) ReadGeolocationData(ctx context.Context) (<-chan []string, error) {
	file, err := os.Open(f.file)
	if err != nil {
		return nil, ctxd.NewError(ctx, "opening file", "error", err)
	}

	reader := csv.NewReader(file)

	// skip header
	_, err = reader.Read()
	if err != nil {
		return nil, ctxd.WrapError(ctx, err, "reading header")
	}

	dataCh := make(chan []string, dataChBuf)

	go func() {
		defer func() {
			close(dataCh)

			defer file.Close() //nolint:errcheck
		}()

		for {
			record, err := reader.Read()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return // End.
				}

				f.logger.Error(ctx, "reading record", "error", err)

				return
			}

			dataCh <- record
		}
	}()

	return dataCh, nil
}
