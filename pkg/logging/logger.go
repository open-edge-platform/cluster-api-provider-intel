// SPDX-FileCopyrightText: (C) 2022 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

// RFC3339Micro represents a time format similar to time.RFC3339Nano but only accurate to microseconds and keeping trailing zeros
const RFC3339Micro = "2006-01-02T15:04:05.000000Z07:00"

//nolint:gochecknoinits // Using init for defining flags is a valid exception.
func init() {
	flag.Func(
		"globalLogLevel",
		"Sets the application-wide logging level. Must be a valid zerolog.Level. Defaults to 'info'",
		handleLogLevel,
	)

	zerolog.TimeFieldFormat = RFC3339Micro
	zerolog.TimestampFieldName = "timestamp"

	// use UTC time
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}

	// Use shorter file name
	// For instance without below CallerMarshalFunc you will see file name like below
	// -> "/Users/$USER/Workspace/$PACKAGE_NAME/pkg/$PATH_TO_FILE/http_server_options.go:219"
	// But with the below  CallerMarshalFunc, the file name will be
	// -> http_server_options.go:219
	// Since we also log the 'component' name in the logs we already have isolation if there are multiple files with the same
	// name across different components.
	// Logging shorter filename consumes less space and resource and also makes the log look neat!
	zerolog.CallerMarshalFunc = callerMarshal
}

func handleLogLevel(l string) error {
	level, err := zerolog.ParseLevel(l)
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(level)
	return nil
}

type MCLogger struct {
	zerolog.Logger
}

type MCCtxLogger struct {
	zerolog.Logger
}

type spanlogHook struct {
	span trace.Span
}

func (h spanlogHook) Run(_ *zerolog.Event, _ zerolog.Level, msg string) {
	if h.span.IsRecording() {
		h.span.AddEvent(msg)
	}
}

func callerMarshal(_ uintptr, file string, line int) string {
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	file = short
	return fmt.Sprintf("%s:%d", short, line)
}

func GetLogger(component string) MCLogger {
	var logger zerolog.Logger

	// When 'HUMAN' env is set it dumps logs in human friendly readable format.
	// More details here https://betterstack.com/community/guides/logging/zerolog/#prettifying-your-logs-in-development.
	if _, present := os.LookupEnv("HUMAN"); present {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339Nano})
	} else {
		logger = zerolog.New(os.Stdout)
	}

	// Overwrite the default log level set in 'zerolog.New' to whatever is the Global Level
	logger = logger.Level(zerolog.GlobalLevel())
	logger = logger.With().Caller().Timestamp().Str("component", component).Logger()
	return MCLogger{logger}
}

func (l MCLogger) TraceCtx(ctx context.Context) MCCtxLogger {
	span := trace.SpanFromContext(ctx)
	newlogger := l.With().
		Str("span_id", span.SpanContext().SpanID().String()).
		Str("trace_id", span.SpanContext().TraceID().String()).
		Logger()
	newlogger = newlogger.Hook(spanlogHook{span})
	return MCCtxLogger{newlogger}
}

// McSec is a logging decorator for MCLogger intended to be used for security related events.
func (l *MCLogger) McSec() *MCLogger {
	return &MCLogger{l.With().Str("MCSec", "true").Logger()}
}

// McSec is a logging decorator MCCtxLogger intended to be used for security related events.
func (l *MCCtxLogger) McSec() *MCCtxLogger {
	return &MCCtxLogger{l.With().Str("MCSec", "true").Logger()}
}

// McErr is an extension for MCLogger intended to be used for error logging.
func (l *MCLogger) McErr(err error) *zerolog.Event {
	mcLogger := &MCLogger{l.With().Err(err).Logger()}
	return mcLogger.Error()
}

// McErr is an extension for MCCtxLogger intended to be used for error logging.
func (l *MCCtxLogger) McErr(err error) *zerolog.Event {
	mcLogger := &MCCtxLogger{l.With().Err(err).Logger()}
	return mcLogger.Error()
}

// McError is an extension for MCLogger intended to be used for logging of inline errors.
func (l *MCLogger) McError(format string, args ...interface{}) *zerolog.Event {
	logger := &MCLogger{l.With().Str("error", fmt.Sprintf(format, args...)).Logger()}
	return logger.Error()
}

// McError is an extension for MCCtxLogger intended to be used for logging of inline errors.
func (l *MCCtxLogger) McError(format string, args ...interface{}) *zerolog.Event {
	logger := &MCCtxLogger{l.With().Str("error", fmt.Sprintf(format, args...)).Logger()}
	return logger.Error()
}
