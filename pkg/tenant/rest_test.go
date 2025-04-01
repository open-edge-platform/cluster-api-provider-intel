// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package tenant

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetActiveProjectIdFromRequest(t *testing.T) {
	cases := []struct {
		name         string
		inputRequest *http.Request
		expectedStr  string
		expectedErr  error
	}{
		{
			name:         "valid project id header in request",
			inputRequest: &http.Request{Header: http.Header{ActiveProjectIdHeaderKey: []string{validProjectId1}}},
			expectedStr:  validProjectId1,
		},
		{
			name:         "multiple project id headers in request valid",
			inputRequest: &http.Request{Header: http.Header{ActiveProjectIdHeaderKey: []string{validProjectId1, invalidProjectId}}},
			expectedStr:  validProjectId1,
		},
		{
			name:         "invalid project id header in request",
			inputRequest: &http.Request{Header: http.Header{ActiveProjectIdHeaderKey: []string{invalidProjectId}}},
			expectedStr:  DefaultProjectId,
			expectedErr:  projectIdInvalidError,
		},
		{
			name:         "multiple project id headers in request invalid",
			inputRequest: &http.Request{Header: http.Header{ActiveProjectIdHeaderKey: []string{invalidProjectId, validProjectId1}}},
			expectedStr:  DefaultProjectId,
			expectedErr:  projectIdInvalidError,
		},
		{
			name:         "no project id header in request",
			inputRequest: &http.Request{},
			expectedStr:  DefaultProjectId,
			expectedErr:  projectIdNotProvidedError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			projectId, err := getActiveProjectIdFromRequest(tc.inputRequest)
			assert.Equal(t, tc.expectedStr, projectId)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestActiveProjectIdEchoMiddleware(t *testing.T) {
	cases := []struct {
		name                   string
		headerActiveProjectIds []string
		expectedProjectId      string
		expectedErr            error
	}{
		{
			name:                   "successfully extract project id from request with single active project id",
			headerActiveProjectIds: []string{validProjectId1},
			expectedProjectId:      validProjectId1,
		},
		{
			name:                   "successfully extract project id from request with multiple active project ids",
			headerActiveProjectIds: []string{validProjectId2, validProjectId1, invalidProjectId},
			expectedProjectId:      validProjectId2,
		},
		{
			name:                   "fail to extract project id from request with single invalid active project id",
			headerActiveProjectIds: []string{invalidProjectId},
			expectedErr:            projectIdInvalidError,
		},
		{
			name:                   "fail to extract project id from request with multiple active project ids where the first id is invalid",
			headerActiveProjectIds: []string{invalidProjectId, validProjectId1, validProjectId2},
			expectedErr:            projectIdInvalidError,
		},
		{
			name:                   "fail to extract project id from request without active project id header",
			headerActiveProjectIds: []string{},
			expectedErr:            projectIdNotProvidedError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := echo.New().NewContext(&http.Request{Header: map[string][]string{ActiveProjectIdHeaderKey: tc.headerActiveProjectIds}}, nil)

			middleware := ActiveProjectIdEchoMiddleware()
			handler := middleware(func(e echo.Context) error {
				assert.Equal(t, tc.expectedProjectId, e.Request().Context().Value(ActiveProjectIdContextKey))
				return nil
			})

			err := handler(ctx)
			if tc.expectedErr != nil {
				assert.Equal(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
