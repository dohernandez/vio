package service

import (
	"google.golang.org/grpc"
)

// KitTemplateRESTService Wrapper on top of the GRPC server to be able to use the interceptor for
// REST request as it is used for grpc request.
type KitTemplateRESTService struct {
	*KitTemplateService

	unaryInt grpc.UnaryServerInterceptor
}

// NewKitTemplateRESTService creates an instance of QontoService.
func NewKitTemplateRESTService(service *KitTemplateService) *KitTemplateRESTService {
	return &KitTemplateRESTService{
		KitTemplateService: service,
	}
}

/*
// PostFuncName is wrapper on the unary RPC to ... for REST calls.
func (s *KitTemplateRESTService) PostFuncName(ctx context.Context, req interface{}) (interface{}, error) {
	info := &grpc.UnaryServerInfo{
		Server:     s.KitTemplateService,
		FullMethod: "/kit.template.Service/PostFuncName",
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.KitTemplateService.PostFuncName(ctx, req)
	}

	resp, err := s.unaryInt(ctx, req, info, handler)

	return resp, err
}
*/
