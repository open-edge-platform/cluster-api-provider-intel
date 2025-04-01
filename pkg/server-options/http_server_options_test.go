// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_options

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/tracing"
	"github.com/stretchr/testify/assert"

	"github.com/labstack/echo/v4"
)

var (
	mockEchoNew        = echo.New()
	mockdefaultBaseURL = "0.0.0.0:38080"
)

func TestServer_setRateLimiter(t *testing.T) {
	type fields struct {
		echo    *echo.Echo
		baseURL string
	}
	type args struct {
		e *echo.Echo
	}

	newEcho := mockEchoNew

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "set rate limiter",
			fields: fields{
				echo:    newEcho,
				baseURL: mockdefaultBaseURL,
			},
			args: args{newEcho},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetRateLimiter(tt.args.e)
		})
	}
}

func TestServer_setLimits(t *testing.T) {
	type fields struct {
		echo    *echo.Echo
		baseURL string
	}
	type args struct {
		e *echo.Echo
	}
	newEcho := mockEchoNew

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "setLimits sets the max size of a request body and the max size of header bytes limiter",
			fields: fields{
				echo:    newEcho,
				baseURL: mockdefaultBaseURL,
			},
			args: args{newEcho},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLimits(tt.args.e)
		})
	}
}

func TestServer_setTimeout(t *testing.T) {
	type fields struct {
		echo    *echo.Echo
		baseURL string
	}
	type args struct {
		e *echo.Echo
	}
	newEcho := mockEchoNew

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "set time out",
			fields: fields{
				echo:    newEcho,
				baseURL: mockdefaultBaseURL,
			},
			args: args{newEcho},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetTimeout(tt.args.e)
		})
	}
}

func TestServer_setOptions(t *testing.T) {
	type fields struct {
		echo    *echo.Echo
		baseURL string
	}
	type args struct {
		e    *echo.Echo
		cors string
	}

	newEcho := mockEchoNew

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "set optioins",
			fields: fields{
				echo:    newEcho,
				baseURL: mockdefaultBaseURL,
			},
			args: args{newEcho, "127.0.0.1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetOptions(tt.args.e, tt.args.cors)
		})
	}
}

func TestServer_setCors(t *testing.T) {
	type fields struct {
		echo    *echo.Echo
		baseURL string
	}
	type args struct {
		e    *echo.Echo
		cors string
	}
	newEcho := mockEchoNew

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "set Cors",
			fields: fields{
				echo:    newEcho,
				baseURL: mockdefaultBaseURL,
			},
			args: args{newEcho, "127.0.0.1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetCors(tt.args.e, tt.args.cors)
		})
	}
}

func TestServer_setMethodOverride(t *testing.T) {
	type fields struct {
		echo    *echo.Echo
		baseURL string
	}
	type args struct {
		e *echo.Echo
	}
	newEcho := mockEchoNew

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "set method override",
			fields: fields{
				echo:    newEcho,
				baseURL: mockdefaultBaseURL,
			},
			args: args{newEcho},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetMethodOverride(tt.args.e)
		})
	}
}

func TestServer_setSecureConfig(t *testing.T) {
	type fields struct {
		echo    *echo.Echo
		baseURL string
	}
	type args struct {
		e *echo.Echo
	}
	newEcho := mockEchoNew

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "set secure config",
			fields: fields{
				echo:    newEcho,
				baseURL: mockdefaultBaseURL,
			},
			args: args{newEcho},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetSecureConfig(tt.args.e)
		})
	}
}

func TestServer_setNoCache(t *testing.T) {
	type fields struct {
		echo    *echo.Echo
		baseURL string
	}
	type args struct {
		e *echo.Echo
	}
	newEcho := mockEchoNew

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "set no cache",
			fields: fields{
				echo:    newEcho,
				baseURL: mockdefaultBaseURL,
			},
			args: args{newEcho},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetNoCache(tt.args.e)
		})
	}
}

func Test_noCacheWithConfig(t *testing.T) {
	type args struct {
		config noCacheConfig
	}
	var tests []struct {
		name string
		args args
		want echo.MiddlewareFunc
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := noCacheWithConfig(tt.args.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("noCacheWithConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSetTracing tests tracing and trace logger which automatically logs trace and spanID
func TestSetTracing(t *testing.T) {
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
	e := echo.New()
	SetTracing(e, "test-service") // generates trace-id and span-id
	e.GET("/users/:id", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hi There!")
	})
	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	/*Now we should see traceID and spanID getting automatically logged.*/
}

// Fake skipper function that always returns false
func fakeSkipperFunc(c echo.Context) bool {
	return false
}

// TestNoCacheWithConfig tests the noCacheWithConfig function
func TestNoCacheWithConfig(t *testing.T) {
	expectedHeaders := noCacheHeaders
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	config := noCacheConfig{
		Skipper: fakeSkipperFunc,
	}

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}
	h := noCacheWithConfig(config)(handler)

	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "test", rec.Body.String())
		for k, v := range expectedHeaders {
			assert.Equal(t, v, rec.Header().Get(k))
		}
	}
}

func TestHttpReqHeaderProcess(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	newHeaders := map[string]string{
		"X-Custom-Header": "Header",
	}
	middleware := httpReqHeaderProcess(newHeaders)
	handler := middleware(func(c echo.Context) error {
		for k, v := range newHeaders {
			if assert.Equal(t, v, c.Request().Header.Get(k)) {
				t.Logf("Header was set correctly: %s: %s", k, v)
			} else {
				t.Errorf("Header was not set correctly: %s. Expected %s, got %s", k, v, c.Request().Header.Get(k))
			}
		}
		return c.NoContent(http.StatusOK)
	})

	if assert.NoError(t, handler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}
