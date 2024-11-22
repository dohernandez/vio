package rest_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/dohernandez/kit-template/pkg/grpc/rest"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/test/bufconn"
)

func TestWithPort(t *testing.T) {
	t.Parallel()

	runShutdownTest(t, rest.WithPort(9800))
}

func runShutdownTest(t *testing.T, opts ...rest.Option) {
	t.Helper()

	serverMuxOpts := append(([]rest.Option)(nil), opts...)

	srv, err := rest.NewServer(context.Background(), serverMuxOpts...)
	require.NoError(t, err)

	shutdownCh := make(chan struct{})
	shutdownDoneCh := make(chan struct{})
	errCh := make(chan error, 1)

	go func() {
		errCh <- srv.WithShutdownSignal(shutdownCh, shutdownDoneCh).Start()
	}()

	time.Sleep(time.Millisecond * 20)

	close(shutdownCh)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1000)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("context deadline exceeded")

		case err := <-errCh:
			require.NoError(t, err)

		case <-shutdownDoneCh:
			return
		}
	}
}

func TestWithAddress(t *testing.T) {
	t.Parallel()

	runShutdownTest(t, rest.WithAddress(":9801"))
}

func TestWithListener_Port(t *testing.T) {
	t.Parallel()

	l, err := net.Listen("tcp", ":9802") // nolint: gosec
	require.NoError(t, err)

	defer l.Close() // nolint: errcheck

	runShutdownTest(t, rest.WithListener(l, false))
}

func TestWithListener_Buf(t *testing.T) {
	t.Parallel()

	buf := bufconn.Listen(1024 * 1024)
	defer buf.Close() // nolint: errcheck

	runShutdownTest(t, rest.WithListener(buf, false))
}
