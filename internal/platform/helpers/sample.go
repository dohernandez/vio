package helpers

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basePath   = filepath.Dir(b)
)

// LoadSampleData loads sample data from a CSV file.
//
// When limit is -1, it loads all the data.
// When offset is -1, it loads the header too.
func LoadSampleData(limit, offset int) ([][]string, error) {
	// Open the CSV file.
	file, err := os.Open(filepath.Join(basePath, "../../../resources/sample_data/test_data.csv")) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	defer file.Close() //nolint:errcheck

	// Create a new CSV reader.
	reader := csv.NewReader(file)

	cursor := 0

	// When offset is not -1, keep the header.
	if offset != -1 {
		for {
			// Read each record (line) from the CSV.
			_, err := reader.Read()
			if errors.Is(err, io.EOF) {
				break // End of file
			}

			if err != nil {
				return nil, fmt.Errorf("moving cursor: %w", err)
			}

			if cursor >= offset {
				break
			}

			cursor++
		}
	}

	if offset == -1 {
		offset = 0
	}

	buf := limit

	// If limit is -1, then start with 100.
	if buf == -1 {
		buf = 100
	}

	records := make([][]string, 0, buf)

	for {
		if limit != -1 && cursor >= offset+limit {
			break
		}

		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break // End of file
		}

		if err != nil {
			return nil, fmt.Errorf("reading record: %w", err)
		}

		records = append(records, record)

		cursor++
	}

	return records, nil
}

// LoadAllSampleData loads all the sample data from a CSV file.
func LoadAllSampleData() ([][]string, error) {
	return LoadSampleData(-1, 0)
}
