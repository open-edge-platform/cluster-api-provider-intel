// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package tracing

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/logging"
)

var log = logging.GetLogger("tracing")

// newTraceResources sets the resource attributes to be traces. It takes the service name and optional map of
// attributes to traces (both key and value being string).
func newTraceResources(service string, attribs map[string]string) (*resource.Resource, error) {
	if service == "" {
		return nil, fmt.Errorf("service name is required")
	}
	attributes := []attribute.KeyValue{
		attribute.String("service.name", service),
		attribute.String("library.language", "go"),
	}

	for attribKey, attribValue := range attribs {
		if attribKey == "" || attribValue == "" {
			continue
		}
		attributes = append(attributes, attribute.String(attribKey, attribValue))
	}

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attributes...,
		),
	)
	if err != nil {
		log.Warn().Err(err).Msg("Could not set trace resources")
	}

	return resources, err
}

// newTraceExporter sets up the exported for the open-telemetry client.
// It takes in the otlptrace.Client and the resource object as arguments to set up the trace exporter.
func newTraceExporter(client otlptrace.Client, resources *resource.Resource) (func(context.Context) error, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}

	if resources == nil {
		return nil, fmt.Errorf("resources is nil")
	}
	exporter, err := otlptrace.New(
		context.Background(),
		client,
	)
	if err != nil {
		log.Warn().Err(err).Msg("Could not create trace exporter")
		return nil, err
	}

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(resources),
		),
	)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return exporter.Shutdown, nil
}

// NewTraceExporterHTTP creates and starts a new exporter of traces to the provided otelURL using HTTP
// The tracing definition is annotated with the service string attribute.
// Other trace attributes can be provided by the attribs map.
// The function returns a function to call the shutdown of the trace.
func NewTraceExporterHTTP(otelURL, service string, attribs map[string]string) (func(context.Context) error, error) {
	if otelURL == "" {
		return nil, fmt.Errorf("otelURL is required")
	}

	if service == "" {
		return nil, fmt.Errorf("service is required")
	}
	secureOption := otlptracehttp.WithInsecure()
	traceClient := otlptracehttp.NewClient(
		secureOption,
		otlptracehttp.WithEndpoint(otelURL),
	)
	resources, err := newTraceResources(service, attribs)
	if err != nil {
		return nil, err
	}
	exporterShutdown, err := newTraceExporter(traceClient, resources)
	if err != nil {
		return nil, err
	}

	return exporterShutdown, nil
}

// NewTraceExporterGRPC creates and start a new exporter of traces to the provided otelgrpcURL using gRPC
// The tracing definition is annotated with the service string attribute.
// Other trace attributes can be provided by the attribs map.
// The function returns a function to call the shutdown of the trace.
func NewTraceExporterGRPC(otelgrpcURL, service string, attribs map[string]string) (func(context.Context) error, error) {
	if otelgrpcURL == "" {
		return nil, fmt.Errorf("otelgrpcURL is required")
	}

	if service == "" {
		return nil, fmt.Errorf("service is required")
	}
	secureOption := otlptracegrpc.WithInsecure()
	traceClient := otlptracegrpc.NewClient(
		secureOption,
		otlptracegrpc.WithEndpoint(otelgrpcURL),
	)

	resources, err := newTraceResources(service, attribs)
	if err != nil {
		return nil, err
	}
	exporterShutdown, err := newTraceExporter(traceClient, resources)
	if err != nil {
		return nil, err
	}

	return exporterShutdown, nil
}

// StartTrace starts a new trace and a child span of it.
// It must be used with StopTrace function to end the child span.
// StopTrace can be used with defer statement after StartTrace call.
func StartTrace(ctx context.Context, servicename, traceName string) context.Context {
	ctx, _ = otel.Tracer(servicename).Start(ctx, traceName)
	return ctx
}

func StartTraceFromRemote(ctx context.Context, servicename, traceName string) (context.Context, error) {
	tracer := otel.Tracer(servicename)
	var span trace.Span
	ctx, span = tracer.Start(
		trace.ContextWithRemoteSpanContext(ctx, trace.SpanContextFromContext(ctx)),
		traceName,
		trace.WithSpanKind(trace.SpanKindClient),
	)
	if span == nil {
		return ctx, fmt.Errorf("failed to start trace span")
	}
	return ctx, nil
}

// StopTrace ends a span trace from the provided context.
// StopTrace can be used with defer statement after StartTrace call.
func StopTrace(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	span.End()
}

// EnableGrpcClientTracing adds automated tracing instrumentation
// to client options.
func EnableGrpcClientTracing(opts []grpc.DialOption) []grpc.DialOption {
	return append(opts,
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		// grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()), // deprecated
		// grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor())) // deprecated
	)
}

// EnableGrpcServerTracing adds automated tracing instrumentation
// to server options.
func EnableGrpcServerTracing(opts []grpc.ServerOption) []grpc.ServerOption {
	return append(opts,
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		// grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()), // deprecated
		// grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor())) // deprecated
	)
}

// EnableEchoAutoTracing adds opentelemetry middleware
// to echo server for automated tracing instrumentation.
func EnableEchoAutoTracing(e *echo.Echo, name string) {
	e.Use(otelecho.Middleware(name))
}
