package server

import "google.golang.org/grpc"

// ServiceServer is an interface for a server that provides services.
type ServiceServer interface {
	RegisterService(s grpc.ServiceRegistrar)
}

// ServiceServerFunc is the function to register service to a service registrar.
type ServiceServerFunc func(s grpc.ServiceRegistrar)
