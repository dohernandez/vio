package server

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/bool64/ctxd"
	"github.com/dohernandez/kit-template/pkg/servicing"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var serverType = "GRPC"

// Option sets up a server.
type Option func(srv *Server)

// WithAddress sets server address.
func WithAddress(addr string) Option {
	return func(srv *Server) {
		srv.config.addr = addr
	}
}

// WithPort sets server address port.
func WithPort(port int) Option {
	return func(srv *Server) {
		srv.config.addr = fmt.Sprintf(":%d", port)
	}
}

// WithListener sets the listener. Server does not need to start a new one.
func WithListener(l net.Listener, shouldCloseListener bool) Option {
	return func(srv *Server) {
		srv.listener = l
		srv.config.shouldCloseListener = shouldCloseListener
	}
}

// WithService registers a service.
func WithService(s ServiceServer) Option {
	return func(srv *Server) {
		srv.config.services = append(srv.config.services, s.RegisterService)
	}
}

// ChainUnaryInterceptor sets the server interceptors for unary.
func ChainUnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) Option {
	return WithServerOption(grpc.ChainUnaryInterceptor(interceptors...))
}

// ChainStreamInterceptor sets the server interceptors for stream.
func ChainStreamInterceptor(interceptors ...grpc.StreamServerInterceptor) Option {
	return WithServerOption(grpc.ChainStreamInterceptor(interceptors...))
}

// WithServerOption sets the options for the grpc server.
func WithServerOption(opts ...grpc.ServerOption) Option {
	return func(srv *Server) {
		srv.config.serverOpts = append(srv.config.serverOpts, opts...)
	}
}

// WithReflective sets service reflective so that APIs can be discovered.
func WithReflective() Option {
	return func(srv *Server) {
		srv.reflective = true
	}
}

// WithAddrAssigned sets service to ask for listener assigned address. Mainly used when the port to the listener is assigned dynamically.
func WithAddrAssigned() Option {
	return func(srv *Server) {
		srv.AddrAssigned = make(chan string, 1)
	}
}

// WithMetrics sets the metrics, metrics handler and metrics listener. Used to initialize all metrics and create http.Serve
// for /metrics endpoint.
func WithMetrics(metrics *grpcPrometheus.ServerMetrics) Option {
	return func(srv *Server) {
		srv.metrics = metrics
	}
}

// Server is a wrapper around grpc.Server.
type Server struct {
	config config

	listener     net.Listener
	server       *grpc.Server
	AddrAssigned chan string

	shutdownSignal <-chan struct{}
	shutdownDone   chan<- struct{}

	listeningError chan error

	reflective bool

	metrics *grpcPrometheus.ServerMetrics
}

type config struct {
	addr                string
	shouldCloseListener bool
	serverOpts          []grpc.ServerOption
	services            []ServiceServerFunc
}

// Listener returns the listener.
func (s *Server) Listener() net.Listener {
	return s.listener
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
		return ctxd.WrapError(context.Background(), err, "failed to start GRPC server",
			"addr", s.config.addr)
	}

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
			if errors.Is(err, grpc.ErrServerStopped) {
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
		return ctxd.WrapError(context.Background(), err, "GRPC server failed",
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

		s.server.GracefulStop()

		close(s.shutdownDone)
	}()
}

// NewServer initiates a new wrapped grpc server.
func NewServer(opts ...Option) *Server {
	srv := &Server{
		config: defaultConfig(),
	}

	for _, o := range opts {
		o(srv)
	}

	// Init GRPC Server.
	grpcSrv := grpc.NewServer(srv.config.serverOpts...)

	for _, register := range srv.config.services {
		register(grpcSrv)
	}

	// Make the service reflective so that APIs can be discovered.
	if srv.reflective {
		reflection.Register(grpcSrv)
	}

	srv.server = grpcSrv

	if srv.metrics != nil {
		// Initialize all metrics.
		srv.metrics.InitializeMetrics(srv.server)
	}

	return srv
}

func defaultConfig() config {
	return config{
		addr:                ":0",
		shouldCloseListener: true,
	}
}

// Name Service name.
func (s *Server) Name() string {
	return serverType
}

// Addr service address.
func (s *Server) Addr() string {
	return s.listener.Addr().String()
}
