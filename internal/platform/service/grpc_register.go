package service

import (
	"google.golang.org/grpc"
)

// RegisterService registers the service implementation to grpc service.
func (s *KitTemplateService) RegisterService(sr grpc.ServiceRegistrar) {
	// register grpc service
}
