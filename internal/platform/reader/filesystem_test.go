package reader

import (
	"context"
	"testing"

	"github.com/bool64/ctxd"
	"github.com/stretchr/testify/require"
)

func TestFileSystem_ReadGeolocationData(t *testing.T) {
	t.Parallel()

	file := "../../../resources/sample_data/test_data.csv"
	logger := &ctxd.LoggerMock{}

	fs := NewFileSystem(file, logger)

	dataCh, err := fs.ReadGeolocationData(context.Background())
	require.NoError(t, err)

	var data [][]string //nolint:prealloc

	for d := range dataCh {
		data = append(data, d)
	}

	require.Len(t, data, 5)
}
