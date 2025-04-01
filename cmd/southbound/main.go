// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	config "github.com/open-edge-platform/cluster-api-provider-intel/internal/southboundconfig"
	grpcserver "github.com/open-edge-platform/cluster-api-provider-intel/internal/southboundserver"

	"google.golang.org/grpc"

	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/logging"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/tracing"
)

var log = logging.GetLogger("main")
var traceCleanupFunc func(context.Context) error

const (
	RbacRealmDirectory = "./internal/rego/authz.rego"
)

func handleSignal(grpcServer *grpc.Server) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	sig := <-sigCh
	log.Info().Msgf("received signal: %v, stopping gRPC Southbound Handler", sig)
	grpcServer.Stop()
	if traceCleanupFunc != nil {
		err := traceCleanupFunc(context.Background())
		if err != nil {
			log.Err(err).Msg("Error in tracing cleanup")
		}
	}
}

func setTracing(traceURL string) func(context.Context) error {
	cleanup, exportErr := tracing.NewTraceExporterGRPC(traceURL, config.TracingServiceName, nil)
	if exportErr != nil {
		log.Err(exportErr).Msg("Error creating trace exporter")
	}
	if cleanup != nil {
		log.Info().Msg("Tracing enabled")
	} else {
		log.Info().Msg("Tracing disabled")
	}
	return cleanup
}

func main() {
	cfg := config.ParseInputArg()
	if err := config.ValidateConfig(cfg); err != nil {
		log.Fatal().Err(err).Msgf("error validating config")
	}

	if cfg.EnableTracing {
		traceCleanupFunc = setTracing(cfg.TraceURL)
	}

	// Start gRPC server to handle RPCs from Edge Node
	gServer, listener := grpcserver.NewGrpcServer(*cfg, RbacRealmDirectory)
	if gServer == nil {
		log.Fatal().Msg("Failed to create grpc server")
	}
	log.Info().Msg("Starting gRPC Southbound Handler")
	go func() {
		if err := grpcserver.RunGrpcServer(gServer, listener); err != nil {
			log.Fatal().Msgf("Failed to run grpc server: %v", err)
		}
	}()

	// Start routine to handle any interrupt signals
	handleSignal(gServer)
	log.Info().Msg("Exiting gRPC Southbound Handler")
}
