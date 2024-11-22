package metrics

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/bool64/ctxd"
	"github.com/dohernandez/kit-template/pkg/servicing"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var serverType = "Metrics"

// Option sets up a metrics server.
type Option func(metrics *Server)

// WithAddress sets metrics server address.
func WithAddress(addr string) Option {
	return func(srv *Server) {
		srv.config.addr = addr
	}
}

// WithPort sets metrics server address port.
func WithPort(port int) Option {
	return func(srv *Server) {
		srv.config.addr = fmt.Sprintf(":%d", port)
	}
}

// WithListener sets the listener. Metrics server does not need to start a new one.
func WithListener(l net.Listener, shouldCloseListener bool) Option {
	return func(srv *Server) {
		srv.listener = l
		srv.config.shouldCloseListener = shouldCloseListener
	}
}

// WithAddrAssigned sets service to ask for listener assigned address. Mainly used when the port to the listener is assigned dynamically.
func WithAddrAssigned() Option {
	return func(srv *Server) {
		srv.AddrAssigned = make(chan string, 1)
	}
}

type config struct {
	addr                string
	shouldCloseListener bool
}

// Server is a wrapper around http.Server for a metrics.
type Server struct {
	config config

	listener     net.Listener
	server       *http.Server
	AddrAssigned chan string

	serverMetrics *grpcPrometheus.ServerMetrics

	shutdownSignal <-chan struct{}
	shutdownDone   chan<- struct{}

	listeningError chan error
}

// NewServer initiates a new metrics server.
func NewServer(opts ...Option) *Server {
	srv := &Server{
		config: defaultConfig(),
	}

	for _, o := range opts {
		o(srv)
	}

	// Create some standard server metrics.
	srv.serverMetrics = grpcPrometheus.NewServerMetrics()

	srv.server = &http.Server{}

	return srv
}

func defaultConfig() config {
	return config{
		addr:                ":0",
		shouldCloseListener: true,
	}
}

// Listener returns the listener.
func (s *Server) Listener() net.Listener {
	return s.listener
}

// ServerMetrics returns the grpc server metrics.
func (s *Server) ServerMetrics() *grpcPrometheus.ServerMetrics {
	return s.serverMetrics
}

// WithShutdownSignal adds channels to wait for shutdown and to report shutdown finished.
func (s *Server) WithShutdownSignal(shutdown <-chan struct{}, done chan<- struct{}) servicing.Service {
	s.shutdownSignal = shutdown
	s.shutdownDone = done

	return s
}

// Start begins listening and serving.
func (s *Server) Start() error {
	if err := s.listen(); err != nil {
		return ctxd.WrapError(context.Background(), err, "failed to start server server",
			"addr", s.config.addr)
	}

	// Create a metrics registry.
	reg := prom.NewRegistry()
	reg.MustRegister(s.serverMetrics)

	reg.MustRegister(collectors.NewBuildInfoCollector())
	reg.MustRegister(collectors.NewGoCollector())

	// Initialize opencensus prometheus exporter
	promExporter, err := prometheus.NewExporter(prometheus.Options{
		Registry:  reg,
		Namespace: "",
	})
	if err != nil {
		return err
	}

	s.server.Handler = promExporter

	s.handleServerShutdown()

	s.listeningError = make(chan error)

	go func() {
		defer func() {
			close(s.listeningError)

			if s.config.shouldCloseListener {
				_ = s.listener.Close() // nolint: errcheck
			}
		}()

		if err := s.server.Serve(s.listener); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				// server is shutting down, handleServerShutdown should have handled this already
				s.listeningError <- nil

				return
			}
			s.listeningError <- err

			return
		}
	}()

	// the server is being asked for the dynamical address assigned.
	if s.AddrAssigned != nil {
		s.AddrAssigned <- s.listener.Addr().String()
	}

	if err := <-s.listeningError; err != nil {
		return ctxd.WrapError(context.Background(), err, "Metrics server failed",
			"addr", s.listener.Addr().String())
	}

	return nil
}

func (s *Server) listen() (err error) {
	if s.listener != nil {
		return nil
	}

	s.listener, err = net.Listen("tcp", s.config.addr)

	return err
}

// handleServerShutdown will handle the shutdown signal that comes to the server
// and gracefully shutdown the server.
func (s *Server) handleServerShutdown() {
	if s.shutdownSignal == nil {
		return
	}

	go func() {
		<-s.shutdownSignal

		if err := s.server.Shutdown(context.Background()); err != nil {
			_ = s.server.Close() // nolint: errcheck
		}

		close(s.shutdownDone)
	}()
}

// Name Service name.
func (s *Server) Name() string {
	return serverType
}

// Addr service address.
func (s *Server) Addr() string {
	return s.listener.Addr().String()
}
