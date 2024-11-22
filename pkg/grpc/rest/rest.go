package rest

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/bool64/ctxd"
	"github.com/dohernandez/kit-template/pkg/servicing"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

var serverType = "REST"

// Option sets up a server mux.
type Option func(mux *Server)

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
		srv.config.services = append(srv.config.services, s.RegisterHandlerService)
	}
}

// WithServerMuxOption sets the options for the mux server.
func WithServerMuxOption(opts ...runtime.ServeMuxOption) Option {
	return func(srv *Server) {
		srv.config.muxOpts = append(srv.config.muxOpts, opts...)
	}
}

// WithHandlerPathOption sets the options for custom path handlers to the mux server.
func WithHandlerPathOption(opts ...HandlerPathFunc) Option {
	return func(srv *Server) {
		srv.config.handlerPaths = append(srv.config.handlerPaths, opts...)
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
	muxOpts             []runtime.ServeMuxOption
	services            []ServiceHandlerServerFunc
	handlerPaths        []HandlerPathFunc
}

// Server is a wrapper around runtime.Server.
type Server struct {
	config config

	listener     net.Listener
	mux          *runtime.ServeMux
	server       *http.Server
	AddrAssigned chan string

	shutdownSignal <-chan struct{}
	shutdownDone   chan<- struct{}

	listeningError chan error
}

// NewServer initiates a new wrapped mux server.
func NewServer(ctx context.Context, opts ...Option) (*Server, error) {
	srv := &Server{
		config: defaultConfig(),
	}

	for _, o := range opts {
		o(srv)
	}

	// Init REST Server.
	mux := runtime.NewServeMux(srv.config.muxOpts...)

	for _, register := range srv.config.services {
		err := register(mux)
		if err != nil {
			return nil, ctxd.WrapError(ctx, err, "failed to register handler service")
		}
	}

	for _, handlerPath := range srv.config.handlerPaths {
		err := handlerPath(mux)
		if err != nil {
			return nil, ctxd.WrapError(ctx, err, "failed to set handler path to service")
		}
	}

	srv.mux = mux

	srv.server = &http.Server{
		Handler: srv.mux,
	}

	return srv, nil
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

// WithShutdownSignal adds channels to wait for shutdown and to report shutdown finished.
func (s *Server) WithShutdownSignal(shutdown <-chan struct{}, done chan<- struct{}) servicing.Service {
	s.shutdownSignal = shutdown
	s.shutdownDone = done

	return s
}

// Start begins listening and serving.
func (s *Server) Start() error {
	if err := s.listen(); err != nil {
		return ctxd.WrapError(context.Background(), err, "failed to start REST server",
			"addr", s.config.addr)
	}

	s.handleServerShutdown()

	s.listeningError = make(chan error, 1)

	go func() {
		defer close(s.listeningError)

		defer func() {
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
		return ctxd.WrapError(context.Background(), err, "REST server failed",
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
