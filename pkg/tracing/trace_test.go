// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"google.golang.org/grpc"
)

var (
	tracingAddressNil = ""
	tracingAddress    = "localhost:2348"
	tracingServiceNil = ""
	tracingService    = "test"
	tracingSpanName   = "span"
	tracingAttribs    = map[string]string{
		"key": "value",
	}
	traceClient = otlptracehttp.NewClient()
)

func TestTracing(t *testing.T) {
	ctx := context.Background()

	shutdownTraceHTTP, err := NewTraceExporterHTTP(tracingAddressNil, tracingService, tracingAttribs)
	require.Error(t, err)
	assert.Nil(t, shutdownTraceHTTP)
	shutdownTraceHTTP, err = NewTraceExporterHTTP(tracingAddress, tracingServiceNil, tracingAttribs)
	require.Error(t, err)
	assert.Nil(t, shutdownTraceHTTP)
	shutdownTraceHTTP, err = NewTraceExporterHTTP(tracingAddress, tracingService, tracingAttribs)
	require.NoError(t, err)
	assert.NotNil(t, shutdownTraceHTTP)
	err = shutdownTraceHTTP(ctx)
	require.NoError(t, err)
	assert.NotNil(t, shutdownTraceHTTP)

	shutdownTraceHTTP, err = NewTraceExporterHTTP(tracingAddress, tracingService, nil)
	require.NoError(t, err)
	assert.NotNil(t, shutdownTraceHTTP)
	err = shutdownTraceHTTP(ctx)
	require.NoError(t, err)
	assert.NotNil(t, shutdownTraceHTTP)

	shutdownTraceGRPC, err := NewTraceExporterGRPC(tracingAddressNil, tracingService, tracingAttribs)
	assert.Error(t, err)
	assert.Nil(t, shutdownTraceGRPC)
	shutdownTraceGRPC, err = NewTraceExporterGRPC(tracingAddress, tracingServiceNil, tracingAttribs)
	assert.Error(t, err)
	assert.Nil(t, shutdownTraceGRPC)
	shutdownTraceGRPC, err = NewTraceExporterGRPC(tracingAddress, tracingService, tracingAttribs)
	assert.NoError(t, err)
	assert.NotNil(t, shutdownTraceGRPC)

	shutdownTraceGRPC, err = NewTraceExporterGRPC(tracingAddress, tracingService, nil)
	assert.NoError(t, err)
	assert.NotNil(t, shutdownTraceGRPC)

	shutdownTrace, err := newTraceExporter(nil, nil)
	assert.Error(t, err)
	assert.Nil(t, shutdownTrace)

	shutdownTrace, err = newTraceExporter(traceClient, nil)
	assert.Error(t, err)
	assert.Nil(t, shutdownTrace)

	StartTrace(ctx, tracingService, tracingSpanName)

	_, error := StartTraceFromRemote(ctx, tracingService, tracingSpanName)
	assert.NoError(t, error)

	defer StopTrace(ctx)

	var optsClient []grpc.DialOption
	optsClient = EnableGrpcClientTracing(optsClient)
	assert.NotNil(t, optsClient)

	var optsServer []grpc.ServerOption
	optsServer = EnableGrpcServerTracing(optsServer)
	assert.NotNil(t, optsServer)
}
