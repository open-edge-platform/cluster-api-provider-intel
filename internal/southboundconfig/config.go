// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package southboundconfig

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/logging"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/utils"
)

const (
	defaultGrpcAddr             = "0.0.0.0"
	defaultGrpcPort             = "50020"
	defaultReadinessProbeGrpcEP = "127.0.0.1:50020"
	defaultTraceURL             = "observability-opentelemetry-collector.orch-platform.svc.cluster.local:4317"
	TracingServiceName          = "mc-ecm"
)

var (
	log           = logging.GetLogger("config")
	EnableTracing bool
)

type Config struct {
	GrpcAddr              string
	GrpcPort              string
	ReadinessProbeGrpcEP  string
	EnableTracing         bool
	TraceURL              string
	UseGrpcStubMiddleware bool
}

// ValidateConfig validates the parsed Configuration
func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("configuration is nil")
	}
	if err := utils.IsValidHost(cfg.GrpcAddr); err != nil {
		return err
	}
	if err := utils.IsValidPort(cfg.GrpcPort); err != nil {
		return err
	}
	if err := utils.IsValidIPV4Port(cfg.ReadinessProbeGrpcEP); err != nil {
		return err
	}
	if cfg.EnableTracing {
		if !utils.IsValidHostnamePort(cfg.TraceURL) {
			return fmt.Errorf("not a valid hostname:port: %v", cfg.TraceURL)
		}
	}

	return nil
}

func ParseBooleanEnvConfigVar(envName string) bool {
	boolEnv, err := strconv.ParseBool(os.Getenv(envName))
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to parse boolean value from env var '%s'", envName)
	}
	return boolEnv
}

// ParseInputArg parses the input flags and returns a *Config struct
func ParseInputArg() *Config {
	grpcAddr := flag.String("grpcAddr", defaultGrpcAddr, "Address for grpc server")
	grpcPort := flag.String("grpcPort", defaultGrpcPort, "Port for grpc server")
	readinessProbeGrpcEP := flag.String("readinessProbeGrpcEP", defaultReadinessProbeGrpcEP, "Healthz port for grpc server")
	enableTracing := flag.Bool("enableTracing", false, "Flag to enable tracing")
	traceURL := flag.String("traceURL", defaultTraceURL, "Tracing URL for OTLP protocol")
	useGrpcStubMiddleware := flag.Bool("useGrpcStubMiddleware", false, "Flag to enable gRPC stub middleware. Use for CO E2E testing only")

	flag.Parse()

	// Set the global tracing variable to be used by other modules in the package for configuring tracing.
	// This is a dirty hack as config information isn't passed around to various modules.
	EnableTracing = *enableTracing

	config := &Config{
		GrpcAddr:              *grpcAddr,
		GrpcPort:              *grpcPort,
		ReadinessProbeGrpcEP:  *readinessProbeGrpcEP,
		EnableTracing:         *enableTracing,
		TraceURL:              *traceURL,
		UseGrpcStubMiddleware: *useGrpcStubMiddleware,
	}

	return config
}
