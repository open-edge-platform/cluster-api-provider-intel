// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/naughtygopher/errors"
)

const (
	// back off delay time interval default value is 500ms
	delayTimeIntervalMS = 500 * time.Millisecond
)

// createRequestBody creates a request body from the given input using JSON marshaling
func CreateRequestBody(bodyReq interface{}) (*bytes.Reader, error) {
	payloadBytes, err := json.Marshal(bodyReq)
	if err != nil {
		return nil, errors.InternalErr(err, "Failed to create http request body")
	}

	body := bytes.NewReader(payloadBytes)
	return body, nil
}

/*
Creates an HTTP request with the specified method, URL, request body, and headers

	headers := map[string]string{
		"Authorization": rc.AdminBearerToken,
		"Accept": "application/json",
		"Content-Type": "application/json",
		// you can add more headers here
	}
*/
func CreateHttpHeaderReq(method, url string, body *bytes.Reader, headers map[string]string) (*http.Request, error) {
	if method == "" || url == "" {
		return nil, errors.Internal("method and url are required")
	}
	var req *http.Request
	var err error
	// this check help to avoid null pointer dereference errors when calling the function with a nil body
	// If body is nil, the http.NewRequest function is called with nil as the third parameter.
	// If body is not nil, it proceeds as it did previously
	if body == nil {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, body)
	}
	if err != nil {
		return nil, errors.InternalErr(err, "failed to create http request")
	}
	// set http request header
	SetHttpHeaders(req, headers)
	return req, nil
}

/*
Creates an HTTP request with the specified context, method, URL, request body, and headers

	headers := map[string]string{
		"Authorization": rc.AdminBearerToken,
		"Accept": "application/json",
		"Content-Type": "application/json",
		// you can add more headers here
	}
*/
func CreateHttpHeaderReqCtx(ctx context.Context, method, url string, body *bytes.Reader, headers map[string]string) (*http.Request, error) {
	if ctx.Err() != nil {
		return nil, errors.InternalErr(ctx.Err(), "context is not alive")
	}
	if method == "" || url == "" {
		return nil, errors.Internal("method and url are required")
	}
	var req *http.Request
	var err error
	// this check help to avoid null pointer dereference errors when calling the function with a nil body
	// If body is nil, the http.NewRequest function is called with nil as the third parameter.
	// If body is not nil, it proceeds as it did previously
	if body == nil {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, body)
	}

	if err != nil {
		return nil, errors.InternalErr(err, "failed to create http request")
	}
	// set http request header
	SetHttpHeaders(req, headers)
	return req, nil
}

/*
Sets the headers of an HTTP request using the provided key-value pairs
input paramter example:

	headers := map[string]string{
		"Authorization": rc.AdminBearerToken,
		"Accept": "application/json",
		"Content-Type": "application/json",
		// you can add more headers here
	}
*/
func SetHttpHeaders(req *http.Request, headers map[string]string) {
	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

// Calculates the backoff duration for exponential backoff based on the retry count
func CalculateBackoff(retryCount uint8) time.Duration {
	// Exponential backoff formula: 2^retryCount * 500 milliseconds
	backoffDuration := time.Duration(1<<retryCount) * delayTimeIntervalMS
	return backoffDuration
}

/*
The function attempts to send the request using the client, with the ability to retry multiple times in case of failure.
*/
func DoHttpRequestWithRetry(req *http.Request, httpClient *http.Client, maxRetries uint8) (*http.Response, error) {
	if req == nil || httpClient == nil {
		return nil, errors.Internal("Invalid input parameters of doHttpRequestWithRetry")
	}

	var err error
	var resp *http.Response

	for i := uint8(0); i <= maxRetries; i++ {
		resp, err = httpClient.Do(req)
		if err == nil {
			return resp, nil
		}

		// If we haven't reached max retries, calculate backoff duration and sleep
		if i < maxRetries {
			backoff := CalculateBackoff(i)
			time.Sleep(backoff)
		}
	}

	return nil, errors.Internal("All retries failed")
}

// closing an HTTP client's response body
func CloseHttpClient(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error().Msgf("error closing http client: %v", err)
		}
	}
}
