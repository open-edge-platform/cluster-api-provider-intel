// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package southboundserver

import (
	"context"
	"net"

	grpc_stub_middleware "github.com/open-edge-platform/cluster-api-provider-intel/test/grpc-stub-middleware"

	"github.com/naughtygopher/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	config "github.com/open-edge-platform/cluster-api-provider-intel/internal/southboundconfig"
	sbhandler "github.com/open-edge-platform/cluster-api-provider-intel/internal/southboundhandler"
	pb "github.com/open-edge-platform/cluster-api-provider-intel/pkg/api/proto"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/logging"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/rbac"
	serverOptions "github.com/open-edge-platform/cluster-api-provider-intel/pkg/server-options"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/tenant"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/utils"
)

type server struct {
	pb.UnimplementedClusterOrchestratorSouthboundServer
	listen  string
	rbac    *rbac.Policy
	handler sbhandler.SouthboundHandler
}

var (
	log        = logging.GetLogger("grpc")
	sampledLog = log.Sample(&zerolog.BasicSampler{N: 10})
)

func NewGrpcServer(cfg config.Config, regoFilePath string) (*grpc.Server, net.Listener) {
	// start OPA server with policies
	opaPolicy, err := rbac.New(regoFilePath)
	if err != nil {
		log.Fatal().Msgf("Failed to start RBAC OPA server: %v", err)
	}

	handler, err := sbhandler.NewHandler()
	if err != nil {
		log.Fatal().Msgf("Failed to create handler: %v", err)
	}

	s := &server{
		listen:  cfg.GrpcAddr + ":" + cfg.GrpcPort,
		rbac:    opaPolicy,
		handler: handler,
	}

	log.Info().Msgf("Control listening on %v", s.listen)
	lis, err := net.Listen("tcp", s.listen)
	if err != nil {
		log.Error().Msgf("failed to listen on port: %v", cfg.GrpcPort)
		return nil, nil
	}
	var srvOpts []grpc.ServerOption
	if cfg.UseGrpcStubMiddleware {
		srvOpts = grpc_stub_middleware.GetStubGrpcOptions()
	} else {
		srvOpts = serverOptions.GetGrpcServerOpts(cfg.EnableTracing)
	}
	grpcServer := grpc.NewServer(srvOpts...)
	// Enable gRPC reflection
	reflection.Register(grpcServer)
	pb.RegisterClusterOrchestratorSouthboundServer(grpcServer, s)

	// Create health service
	healthService := health.NewServer()

	// Register the health service with the gRPC server
	grpc_health_v1.RegisterHealthServer(grpcServer, healthService)

	// Set the service health status
	healthService.SetServingStatus("/cluster_orchestrator_southbound_proto.ClusterOrchestratorSouthbound", grpc_health_v1.HealthCheckResponse_SERVING)
	return grpcServer, lis
}

func RunGrpcServer(server *grpc.Server, lis net.Listener) error {
	if err := server.Serve(lis); err != nil {
		return err
	}
	return nil
}

func (s *server) RegisterCluster(ctx context.Context, in *pb.RegisterClusterRequest) (*pb.RegisterClusterResponse, error) {
	if !s.rbac.RequestIsAuthorized(ctx, rbac.MethodRegister) {
		log.Error().Msg("Request RegisterCluster is not authenticated")
		return nil, status.Error(codes.Unauthenticated, "Request is blocked by RBAC")
	}
	if err := in.Validate(); err != nil {
		log.Error().Msgf("error validating incoming request: %v", err)
		return nil, err
	}
	log.Info().Msgf("Get register cluster request from cluster agent in node: %v", in.GetNodeGuid())

	installCommand, uninstallCommand, res, _ := s.handler.Register(ctx, in.GetNodeGuid())

	return &pb.RegisterClusterResponse{
		InstallCmd:   installCommand,
		UninstallCmd: uninstallCommand,
		Res:          res,
	}, nil
}

func (s *server) UpdateClusterStatus(ctx context.Context, in *pb.UpdateClusterStatusRequest) (*pb.UpdateClusterStatusResponse, error) {
	if !s.rbac.RequestIsAuthorized(ctx, rbac.MethodUpdate) {
		log.Error().Msg("request to update cluster status is not authenticated")
		return nil, status.Error(codes.Unauthenticated, "request is blocked by rbac")
	}

	if err := in.Validate(); err != nil {
		log.Error().Msgf("error validating incoming request: %v", err)
		return nil, err
	}

	handleFrequentLog(ctx, "cluster-agent status update, '%s': '%s'", in.GetNodeGuid(), in.GetCode())
	actionReq, err := s.handler.UpdateStatus(ctx, in.GetNodeGuid(), in.GetCode())

	if err != nil {
		st, msg, _ := errors.HTTPStatusCodeMessage(err)
		log.Error().
			Any("status", utils.HttpToGrpcStatusCode(st)).
			Str("msg", msg).
			Str("stack", errors.Stacktrace(err)).
			Msgf("error updating cluster status for node: %v", in.GetNodeGuid())

		// note that we are converting from http status code grpc status code
		return nil, status.Error(utils.HttpToGrpcStatusCode(st), msg)
	}
	return &pb.UpdateClusterStatusResponse{
		ActionRequest: actionReq,
	}, nil
}

// local functions

// handleFrequentLog can be used to log frequent logs
// If this function finds greater utility across many other areas, we can move this to common repo later.
func handleFrequentLog(ctx context.Context, msg string, opts ...interface{}) {
	// This particular log is logged periodically every ~10s per edge node and can generate a lot of logs when there are
	// 100s of edge nodes. What we do here is below
	// 1. If DebugLevel (or lower) log everytime
	// 2. Else sample and log every 10th instance of it
	// If this logic isn't helping for some reason, we can refine it.
	// This is a periodic status message and ideally does not make sense logging it at Info level everytime.
	if zerolog.GlobalLevel() <= zerolog.DebugLevel {
		log.Debug().Str("project", tenant.GetActiveProjectIdFromContext(ctx)).Msgf(msg, opts...)
	} else {
		sampledLog.Info().Msgf(msg, opts...)
		/*
			There is also Random sampler we can explore later if above logic only logs for particular edge node
			Below will randomly sample log every ~ 10 events.
			-> log.Sample(zerolog.RandomSampler(10)).Info().Msgf("Get cluster status update from cluster agent in node: %v, Status code: %v", in.NodeGuid, in.Code)
		*/
	}
}
