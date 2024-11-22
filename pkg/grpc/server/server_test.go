package server_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/dohernandez/kit-template/pkg/grpc/server"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/test/bufconn"
)

func TestWithPort(t *testing.T) {
	t.Parallel()

	runShutdownTest(t, server.WithPort(8800))
}

func runShutdownTest(t *testing.T, opts ...server.Option) {
	t.Helper()

	serverOpts := append(([]server.Option)(nil), opts...)
	serverOpts = append(serverOpts,
		server.ChainUnaryInterceptor(grpcRecovery.UnaryServerInterceptor()),
		server.ChainStreamInterceptor(grpcRecovery.StreamServerInterceptor()),
	)

	srv := server.NewServer(serverOpts...)

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

	runShutdownTest(t, server.WithAddress(":8801"))
}

func TestWithListener_Port(t *testing.T) {
	t.Parallel()

	l, err := net.Listen("tcp", ":8802") // nolint: gosec
	require.NoError(t, err)

	defer l.Close() // nolint: errcheck

	runShutdownTest(t, server.WithListener(l, false))
}

func TestWithListener_Buf(t *testing.T) {
	t.Parallel()

	buf := bufconn.Listen(1024 * 1024)
	defer buf.Close() // nolint: errcheck

	runShutdownTest(t, server.WithListener(buf, false))
}
