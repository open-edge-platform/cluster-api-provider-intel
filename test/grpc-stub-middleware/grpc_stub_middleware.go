// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package grpc_stub_middleware

import (
	"context"
	"os"

	grpcmw "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/logging"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/tenant"
	"google.golang.org/grpc"
)

var (
	log = logging.GetLogger("grpc-stub-middleware")
)

func GetStubGrpcOptions() []grpc.ServerOption {
	var serverOptions []grpc.ServerOption

	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor

	unaryInterceptors = append(unaryInterceptors, StubActiveProjectIdGrpcInterceptor())

	serverOptions = append(serverOptions,
		grpc.UnaryInterceptor(grpcmw.ChainUnaryServer(unaryInterceptors...)),
		grpc.StreamInterceptor(grpcmw.ChainStreamServer(streamInterceptors...)))

	return serverOptions
}

// StubActiveProjectIdGrpcInterceptor returns an interceptor that inserts a dummy active project id.
func StubActiveProjectIdGrpcInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Get the tenant ID from the environment. Note that tenantId and projectId are used interchangeably.
		projectId := os.Getenv("TENANT_ID")
		if projectId == "" {
			projectId = "53cd37b9-66b2-4cc8-b080-3722ed7af64a" // Fallback to a default tenant ID if not set
		}
		log.Trace().Msgf("project id intercepted in grpc request: '%s'", projectId)
		return handler(tenant.AddActiveProjectIdToContext(ctx, projectId), req)
	}
}
