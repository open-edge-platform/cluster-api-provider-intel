// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/undefinedlabs/go-mpatch"
)

var (
	errHttpDo = errors.New("errHttpDo error")
)

//nolint:unparam
func patchClientDoResp(t *testing.T, statusCode int, fail bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error
	var c *http.Client
	readCloser := io.NopCloser(strings.NewReader("Hello GoLinuxCloud Members!"))
	resptest := &http.Response{
		Status:           "",
		StatusCode:       statusCode,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           map[string][]string{},
		Body:             readCloser,
		ContentLength:    0,
		TransferEncoding: []string{},
		Close:            false,
		Uncompressed:     false,
		Trailer:          map[string][]string{},
		Request:          &http.Request{},
		TLS:              &tls.ConnectionState{},
	}
	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(c), "Do", func(c *http.Client, req *http.Request) (*http.Response, error) {
		if fail {
			return resptest, errHttpDo
		} else {
			return resptest, nil
		}
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
	return patch
}

//nolint:unused
type nopCloser struct {
	io.Reader
}

//nolint:unused
func (nopCloser) Close() error {
	return nil
}

//nolint:unused
type errorCloser struct {
	io.Reader
}

//nolint:unused
func (errorCloser) Close() error {
	return errors.New("close error")
}

func Test_closeHttpClient(t *testing.T) {
	type args struct {
		resp *http.Response
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			"with nil response",
			args{
				resp: nil,
			},
		},
		{
			"with response but nil body",
			args{
				resp: &http.Response{},
			},
		},
		{
			"fails to close",
			args{
				resp: &http.Response{
					Body: &errorCloser{},
				},
			},
		},
		{
			"successful close",
			args{
				resp: &http.Response{
					Body: &nopCloser{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CloseHttpClient(tt.args.resp)
		})
	}
}

// Custom failingReadCloser that always returns an error when Read is called
//
//nolint:unused
type failingReadCloser struct{}

//nolint:unused
func (failingReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("error reading response body")
}

//nolint:unused
func (f failingReadCloser) Close() error {
	return nil // Implement the Close method to satisfy the io.ReadCloser interface
}

func Test_setHttpHeaders(t *testing.T) {
	type args struct {
		req     *http.Request
		headers map[string]string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "Test Case 1: Single Header",
			args: args{
				req: &http.Request{
					Header: http.Header{},
				},
				headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
		},
		{
			name: "Test Case 2: Multiple Headers",
			args: args{
				req: &http.Request{
					Header: http.Header{},
				},
				headers: map[string]string{
					"Authorization":   "Bearer token",
					"X-Custom-Header": "value",
				},
			},
		},
		{
			name: "Test Case 3: Empty Headers",
			args: args{
				req: &http.Request{
					Header: http.Header{},
				},
				headers: map[string]string{},
			},
		},
		{
			name: "Header Value with Special Characters",
			args: args{
				req: &http.Request{
					Header: http.Header{},
				},
				headers: map[string]string{
					"X-Header": "Value with !@#$%^&*() special characters",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetHttpHeaders(tt.args.req, tt.args.headers)
		})
	}
}

func Test_doHttpRequestWithRetry(t *testing.T) {

	type args struct {
		req        *http.Request
		httpClient *http.Client
		maxRetries uint8
	}
	tests := []struct {
		name           string
		args           args
		want           *http.Response
		wantErr        bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		// TODO: Add test cases.
		{
			name: "Test Case 1: Successful request without retries",
			args: args{
				req: &http.Request{
					URL: &url.URL{
						Scheme: "https",
						Host:   "example.com",
						// Add the additional URL components as needed
					},
					Header: http.Header{},
				},
				httpClient: &http.Client{},
				maxRetries: 0,
			},
			want: &http.Response{
				StatusCode: 200,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader("Hello GoLinuxCloud Members!")),
			},
			wantErr: false,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchClientDoResp(t, 200, false)
				return []*mpatch.Patch{patch1}
			},
		},
		{
			name: "Test Case 2: Successful request with retries",
			args: args{
				req: &http.Request{
					URL: &url.URL{
						Scheme: "https",
						Host:   "example.com",
					},
					Header: http.Header{},
				},
				httpClient: &http.Client{},
				maxRetries: 3,
			},
			want: &http.Response{
				StatusCode: 200,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader("Hello GoLinuxCloud Members!")),
			},
			wantErr: false,
			funcBeforeTest: func() []*mpatch.Patch {
				// Return a successful response on the first attempt, and fail on subsequent attempts
				patch1 := patchClientDoResp(t, 200, false)
				return []*mpatch.Patch{patch1}
			},
		},
		{
			name: "Test Case 3: All retries failed",
			args: args{
				req: &http.Request{
					URL: &url.URL{
						Scheme: "https",
						Host:   "example.com",
					},
					Header: http.Header{},
				},
				httpClient: &http.Client{},
				maxRetries: 3,
			},
			want: &http.Response{
				StatusCode: 500,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader("Hello GoLinuxCloud Members!")),
			},
			wantErr: true,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchClientDoResp(t, 500, true)
				return []*mpatch.Patch{patch1}
			},
		},
		{
			name: "Test Case 4: Invalid URL",
			args: args{
				req: &http.Request{
					Method: http.MethodGet,
					URL:    &url.URL{Scheme: "invalid", Host: "example.com"},
				},
				httpClient: &http.Client{},
				maxRetries: 0,
			},
			want: &http.Response{
				StatusCode: 0,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader("Hello GoLinuxCloud Members!")),
			},
			wantErr: true,
			funcBeforeTest: func() []*mpatch.Patch {
				return []*mpatch.Patch{}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pList := []*mpatch.Patch{}
			if tt.funcBeforeTest != nil {
				pList = tt.funcBeforeTest()
			}
			defer unpatchAll(t, pList) // Unpatch the methods at the end of the test case
			got, err := DoHttpRequestWithRetry(tt.args.req, tt.args.httpClient, tt.args.maxRetries)
			if (err != nil) != tt.wantErr {
				t.Errorf("doHttpRequestWithRetry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != nil && got.StatusCode != tt.want.StatusCode {
				t.Errorf("doHttpRequestWithRetry() = %v, want %v", got, tt.want)
			}
			// unpatchAll(t, pList)
		})
	}
}

func TestCreateHttpHeaderReq(t *testing.T) {
	type args struct {
		method  string
		url     string
		body    *bytes.Reader
		headers map[string]string
	}
	body := bytes.NewReader([]byte("request body"))
	tests := []struct {
		name          string
		args          args
		wantHost      string
		wantErr       bool
		wantEmptyBody bool
	}{
		{
			name:          "GET request with JSON content type",
			args:          args{"GET", "http://example.com", body, map[string]string{"Content-Type": "application/json"}},
			wantHost:      "example.com",
			wantErr:       false,
			wantEmptyBody: false,
		},
		{
			name:          "POST request with Form content type",
			args:          args{"POST", "http://example.com/login", body, map[string]string{"Content-Type": "application/x-www-form-urlencoded"}},
			wantHost:      "example.com",
			wantErr:       false,
			wantEmptyBody: false,
		},
		{
			name:    "Invalid URL",
			args:    args{"GET", ":foo", body, map[string]string{"Content-Type": "application/json"}},
			wantErr: true,
		},
		{
			name:    "Empty method",
			args:    args{"", "http://example.com", body, map[string]string{"Content-Type": "application/json"}},
			wantErr: true,
		},
		{
			name:          "GET request with empty body",
			args:          args{"GET", "http://example.com", nil, map[string]string{"Content-Type": "application/json"}},
			wantHost:      "example.com",
			wantErr:       false,
			wantEmptyBody: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateHttpHeaderReq(tt.args.method, tt.args.url, tt.args.body, tt.args.headers)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateHttpHeaderReq() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if got.Method != tt.args.method {
				t.Errorf("CreateHttpHeaderReq() got method = %v, want %v", got.Method, tt.args.method)
			}

			if got.Host != tt.wantHost {
				t.Errorf("CreateHttpHeaderReq() got host = %v, want %v", got.Host, tt.wantHost)
			}

			if got.Header.Get("Content-Type") != tt.args.headers["Content-Type"] {
				t.Errorf("CreateHttpHeaderReq() got content type = %v, want %v", got.Header.Get("Content-Type"), tt.args.headers["Content-Type"])
			}

			if got.Body == nil && !tt.wantEmptyBody {
				t.Errorf("CreateHttpHeaderReq() got no body, want body")
			}
		})
	}
}
func TestCreateHttpHeaderReqCtx(t *testing.T) {
	type args struct {
		ctx     context.Context
		method  string
		url     string
		body    *bytes.Reader
		headers map[string]string
	}
	body := bytes.NewReader([]byte("request body"))
	tests := []struct {
		name          string
		args          args
		wantHost      string
		wantErr       bool
		wantEmptyBody bool
	}{
		// TODO: Add test cases.
		{
			name:          "GET request with JSON content type",
			args:          args{context.Background(), "GET", "http://example.com", body, map[string]string{"Content-Type": "application/json"}},
			wantHost:      "example.com",
			wantErr:       false,
			wantEmptyBody: false,
		},
		{
			name:          "POST request with Form content type",
			args:          args{context.Background(), "POST", "http://example.com/login", body, map[string]string{"Content-Type": "application/x-www-form-urlencoded"}},
			wantHost:      "example.com",
			wantErr:       false,
			wantEmptyBody: false,
		},
		{
			name:    "Invalid URL",
			args:    args{context.Background(), "GET", ":foo", body, map[string]string{"Content-Type": "application/json"}},
			wantErr: true,
		},
		{
			name:    "Empty method",
			args:    args{context.Background(), "", "http://example.com", body, map[string]string{"Content-Type": "application/json"}},
			wantErr: true,
		},
		{
			name:          "GET request with empty body",
			args:          args{context.Background(), "GET", "http://example.com", nil, map[string]string{"Content-Type": "application/json"}},
			wantHost:      "example.com",
			wantErr:       false,
			wantEmptyBody: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateHttpHeaderReqCtx(tt.args.ctx, tt.args.method, tt.args.url, tt.args.body, tt.args.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateHttpHeaderReq() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if got.Method != tt.args.method {
				t.Errorf("CreateHttpHeaderReq() got method = %v, want %v", got.Method, tt.args.method)
			}

			if got.Host != tt.wantHost {
				t.Errorf("CreateHttpHeaderReq() got host = %v, want %v", got.Host, tt.wantHost)
			}

			if got.Header.Get("Content-Type") != tt.args.headers["Content-Type"] {
				t.Errorf("CreateHttpHeaderReq() got content type = %v, want %v", got.Header.Get("Content-Type"), tt.args.headers["Content-Type"])
			}

			if got.Body == nil && !tt.wantEmptyBody {
				t.Errorf("CreateHttpHeaderReq() got no body, want body")
			}
		})
	}
}

func TestCreateRequestBody(t *testing.T) {
	type testStruct struct {
		A string `json:"a"`
		B int    `json:"b"`
	}

	tests := []struct {
		name    string
		bodyReq interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "StructToJSON",
			bodyReq: testStruct{A: "Test", B: 123},
			want:    `{"a":"Test","b":123}`,
			wantErr: false,
		},
		{
			name:    "NilInput",
			bodyReq: nil,
			want:    "null",
			wantErr: false,
		},
		{
			name:    "UnsupportedTypeInput",
			bodyReq: make(chan int),
			wantErr: true,
		},
		{
			name:    "EmptyStructInput",
			bodyReq: struct{}{},
			want:    "{}",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateRequestBody(tt.bodyReq)

			// Check expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRequestBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// No need to further test if wantErr is true
			if tt.wantErr {
				return
			}

			// Read content from bytes.Reader
			var gotBytes bytes.Buffer
			if _, err := gotBytes.ReadFrom(got); err != nil {
				t.Fatalf("CreateRequestBody() reading = %v", err)
			}

			// Compare content
			if gotStr := gotBytes.String(); gotStr != tt.want {
				t.Errorf("CreateRequestBody() = %v, want %v", gotStr, tt.want)
			}
		})
	}
}
