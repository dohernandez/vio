package rest

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

// ServiceServer is an interface for a server that provides services.
type ServiceServer interface {
	RegisterHandlerService(mux *runtime.ServeMux) error
	WithUnaryServerInterceptor(i grpc.UnaryServerInterceptor) ServiceServer
}

// ServiceHandlerServerFunc is the function to register the http handlers for service to "mux".
type ServiceHandlerServerFunc func(mux *runtime.ServeMux) error

// HandlerPathFunc allows users to configure custom path handlers for mux service.
type HandlerPathFunc func(mux *runtime.ServeMux) error

// HandlerPathOption allows users to configure custom path handlers. The struct is used along with HandlerPathFunc when
// configuring custom path handlers for the mux service.
type HandlerPathOption struct {
	// Method is the http method used in the route/path.
	Method string
	// PathPattern the http route path.
	PathPattern string
	// Handler is what to do when the request fulfill the method and path pattern.
	Handler func(w http.ResponseWriter, r *http.Request, _ map[string]string)
}
