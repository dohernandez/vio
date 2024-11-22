package service

import (
	"github.com/dohernandez/kit-template/pkg/grpc/rest"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

// RegisterHandlerService registers the service implementation to mux.
func (s *KitTemplateRESTService) RegisterHandlerService(mux *runtime.ServeMux) error {
	// register rest service
	// return api.RegisterKitTemplateServiceHandlerServer(context.Background(), mux, s)
	return nil
}

// WithUnaryServerInterceptor set the UnaryServerInterceptor for the REST service.
func (s *KitTemplateRESTService) WithUnaryServerInterceptor(i grpc.UnaryServerInterceptor) rest.ServiceServer {
	s.unaryInt = i

	return s
}
