// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_options

import (
	"context"
	"net"
	"testing"
	"time"

	example "github.com/gogo/grpc-example/proto"
	"github.com/gogo/grpc-example/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/tracing"
)

func TestGrpcServerOptions(t *testing.T) {
	cleanup, exportErr := tracing.NewTraceExporterGRPC(
		"observability-opentelemetry-collector.orch-platform.svc.cluster.local:4317",
		"test-service", nil,
	)
	if exportErr != nil {
		t.Errorf("Error creating trace exporter: %v", exportErr)
		return
	}
	if cleanup != nil {
		t.Log("Tracing enabled")
	} else {
		t.Errorf("Tracing could not be enabled")
		return
	}
	defer func() {
		_ = cleanup(context.Background())
	}()
	srvOpts := GetGrpcServerOpts(true)
	if len(srvOpts) == 0 {
		t.Errorf("no serveroptions were returned")
	}
	buffer := 1024 * 1024
	lis := bufconn.Listen(buffer)
	baseServer := grpc.NewServer(srvOpts...)
	example.RegisterUserServiceServer(baseServer, server.New())
	go func() {
		if err := baseServer.Serve(lis); err != nil {
			t.Errorf("error serving server: %v", err)
			return
		}
	}()
	conn, err := grpc.DialContext(context.Background(), "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Errorf("error connecting to server: %v", err)
		return
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			t.Errorf("error closing listener: %v", err)
			return
		}
		baseServer.Stop()
	}
	defer closer()
	client := example.NewUserServiceClient(conn)
	// The below API call will log the trace and span ID at entry/exist of the call.
	_, err = client.ListUsers(context.Background(), &example.ListUsersRequest{
		CreatedSince: nil,
		OlderThan:    nil,
	})
	t.Logf("return val: %v", err)

	// Extra time allowed for the logger to complete its task
	time.Sleep(5 * time.Second)

}
