// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package tenant

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddActiveProjectIdToContext(t *testing.T) {
	cases := []struct {
		name           string
		inputCtx       context.Context
		inputProjectId string
	}{
		{
			name:           "add valid project id to empty context",
			inputCtx:       context.Background(),
			inputProjectId: validProjectId1,
		},
		{
			name:           "add valid project id to existing project context",
			inputCtx:       context.WithValue(context.Background(), ActiveProjectIdContextKey, validProjectId1),
			inputProjectId: validProjectId2,
		},
		{
			name:           "add invalid project id to context",
			inputCtx:       context.Background(),
			inputProjectId: invalidProjectId,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.inputProjectId == invalidProjectId {
				assert.Panics(t, func() {
					_ = AddActiveProjectIdToContext(tc.inputCtx, tc.inputProjectId)
				})
				return
			}

			ctx := AddActiveProjectIdToContext(tc.inputCtx, tc.inputProjectId)
			assert.Equal(t, tc.inputProjectId, ctx.Value(ActiveProjectIdContextKey))
		})
	}
}

func TestGetActiveProjectIdFromContext(t *testing.T) {
	cases := []struct {
		name              string
		inputCtx          context.Context
		expectedProjectId string
	}{
		{
			name:              "valid project id value in context",
			inputCtx:          context.WithValue(context.Background(), ActiveProjectIdContextKey, validProjectId1),
			expectedProjectId: validProjectId1,
		},
		{
			name:              "invalid project id value in context",
			inputCtx:          context.WithValue(context.Background(), ActiveProjectIdContextKey, 7),
			expectedProjectId: DefaultProjectId,
		},
		{
			name:              "no project id value in context",
			inputCtx:          context.Background(),
			expectedProjectId: DefaultProjectId,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectedProjectId == DefaultProjectId {
				assert.Panics(t, func() {
					_ = GetActiveProjectIdFromContext(tc.inputCtx)
				})
				return
			}

			assert.Equal(t, tc.expectedProjectId, GetActiveProjectIdFromContext(tc.inputCtx))
		})
	}
}

func TestIsValidUuid(t *testing.T) {
	cases := []struct {
		name      string
		inputUuid string
		expected  bool
	}{
		{
			name:      "valid uuid 1",
			inputUuid: validProjectId1,
			expected:  true,
		},
		{
			name:      "valid uuid 2",
			inputUuid: validProjectId2,
			expected:  true,
		},
		{
			name:      "invalid uuid",
			inputUuid: invalidProjectId,
			expected:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, isValidUuid(tc.inputUuid))
		})
	}
}
