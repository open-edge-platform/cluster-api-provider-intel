// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_options

import (
	"context"

	grpcmw "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	mclog "github.com/open-edge-platform/cluster-api-provider-intel/pkg/logging"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/tenant"
)

// InterceptorLogger logs the traceID and spanID along with whatever message is passed
func InterceptorLogger(l mclog.MCLogger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, _ ...any) {
		method, _ := grpc.Method(ctx)
		l.Debug().
			Str("method", method).
			Str("trace", trace.SpanFromContext(ctx).SpanContext().TraceID().String()).
			Str("span", trace.SpanFromContext(ctx).SpanContext().SpanID().String()).
			Msg(msg)
	})
}

func ExemptPathUnaryInterceptor(exemptPaths []string, interceptor grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Check if the method is in the exempt paths
		for _, path := range exemptPaths {
			if info.FullMethod == path {
				// Skip the interceptor and directly call the handler
				return handler(ctx, req)
			}
		}
		// Apply the actual interceptor
		return interceptor(ctx, req, info, handler)
	}
}

func GetGrpcServerOpts(enableTracing bool) []grpc.ServerOption {
	var serverOptions []grpc.ServerOption

	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor

	if enableTracing {
		// Generate an event at the start and end of the call
		opts := []logging.Option{logging.WithLogOnEvents(logging.StartCall, logging.FinishCall)}

		// Enable options to insert otel context as well as log the trace and span ID at entry/exit of the call
		unaryInterceptors = append(unaryInterceptors, logging.UnaryServerInterceptor(InterceptorLogger(log), opts...))
		streamInterceptors = append(streamInterceptors, logging.StreamServerInterceptor(InterceptorLogger(log), opts...))
		serverOptions = append(serverOptions, grpc.StatsHandler(otelgrpc.NewServerHandler()))
	}

	// Add the tenant interceptor
	tenantInterceptor := tenant.ActiveProjectIdGrpcInterceptor()

	// Wrap the tenant interceptor to exempt specific paths
	exemptPaths := []string{"/cluster_orchestrator_southbound_proto.ClusterOrchestratorSouthbound"}
	wrappedTenantInterceptor := ExemptPathUnaryInterceptor(exemptPaths, tenantInterceptor)
	unaryInterceptors = append(unaryInterceptors, wrappedTenantInterceptor)

	serverOptions = append(serverOptions,
		grpc.UnaryInterceptor(grpcmw.ChainUnaryServer(unaryInterceptors...)),
		grpc.StreamInterceptor(grpcmw.ChainStreamServer(streamInterceptors...)))

	return serverOptions
}
