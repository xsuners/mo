package xgrpc

import (
	"context"

	"github.com/xsuners/mo/log"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type Checker struct{}

// Check .
func (c *Checker) Check(ctx context.Context, in *grpc_health_v1.HealthCheckRequest) (out *grpc_health_v1.HealthCheckResponse, err error) {
	log.Infos("Check")
	out = new(grpc_health_v1.HealthCheckResponse)
	out.Status = grpc_health_v1.HealthCheckResponse_SERVING
	return
}

// Watch .
func (c *Checker) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	return nil
}
