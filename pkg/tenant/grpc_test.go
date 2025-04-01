// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package tenant

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActiveProjectIdGrpcInterceptor(t *testing.T) {
	cases := []struct {
		name              string
		incomingContext   context.Context
		expectedProjectId string
		expectedErr       string
	}{
		{
			name:              "successfully extract project id from single role",
			incomingContext:   ctxWithJwtRoles(t, validProjectId1+validRole),
			expectedProjectId: validProjectId1,
		},
		{
			name:              "successfully extract project id from multiple roles", // this case is not expected in production
			incomingContext:   ctxWithJwtRoles(t, validProjectId1+validRole, validProjectId1+validRole),
			expectedProjectId: validProjectId1,
		},
		{
			name:              "successfully extract project id from multiple roles with different project ids", // this case is not expected in production
			incomingContext:   ctxWithJwtRoles(t, validProjectId1+validRole, validProjectId2+invalidRole),
			expectedProjectId: validProjectId1,
		},
		{
			name:              "no roles contain a valid project id or role",
			incomingContext:   ctxWithJwtRoles(t, validProjectId1+invalidRole, validProjectId1+"missing-separator", "_no-project-id", "invalid-project-id_role-irrelevant"),
			expectedProjectId: DefaultProjectId,
			expectedErr:       "rpc error: code = Unauthenticated desc = project id not available in request context",
		},
		{
			name:              "unauthorized response with mismatched project ids across multiple roles",
			incomingContext:   ctxWithJwtRoles(t, validProjectId1+validRole, validProjectId2+validRole),
			expectedProjectId: DefaultProjectId,
			expectedErr:       "rpc error: code = Unauthenticated desc = mismatched project ids found in jwt roles",
		},
		{
			name:              "successfully extract project id from metadata",
			incomingContext:   ctxWithProjectIdKey(validProjectId1),
			expectedProjectId: validProjectId1,
		},
		{
			name:              "fail to extract project id from metadata with invalid id",
			incomingContext:   ctxWithProjectIdKey(invalidProjectId),
			expectedProjectId: DefaultProjectId,
			expectedErr:       "rpc error: code = Unauthenticated desc = project id not supplied",
		},
		{
			name:              "fail to extract project id from empty context",
			incomingContext:   context.TODO(),
			expectedProjectId: DefaultProjectId,
			expectedErr:       "rpc error: code = Unauthenticated desc = project id not supplied",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotNil(t, tc.incomingContext)

			interceptor := ActiveProjectIdGrpcInterceptor()
			handler := func(ctx context.Context, req any) (any, error) {
				assert.Equal(t, tc.expectedProjectId, ctx.Value(ActiveProjectIdContextKey))
				return nil, nil
			}

			_, err := interceptor(tc.incomingContext, nil, nil, handler)
			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtractProjectIdFromJwtRoles(t *testing.T) {
	cases := []struct {
		name              string
		incomingContext   context.Context
		expectedProjectId string
		expectedErr       string
	}{
		{
			name:              "successfully extract project id from single role",
			incomingContext:   ctxWithJwtRoles(t, validProjectId1+validRole),
			expectedProjectId: validProjectId1,
		},
		{
			name:              "successfully extract project id from multiple roles", // this case is not expected in production
			incomingContext:   ctxWithJwtRoles(t, validProjectId1+validRole, validProjectId1+validRole),
			expectedProjectId: validProjectId1,
		},
		{
			name:              "successfully extract project id from multiple roles with different project ids", // this case is not expected in production
			incomingContext:   ctxWithJwtRoles(t, validProjectId1+validRole, validProjectId2+invalidRole),
			expectedProjectId: validProjectId1,
		},
		{
			name:              "no roles contain a valid project id or role",
			incomingContext:   ctxWithJwtRoles(t, validProjectId1+invalidRole, validProjectId1+"missing-separator", "_no-project-id", "invalid-project-id_role-irrelevant"),
			expectedProjectId: DefaultProjectId,
			expectedErr:       "rpc error: code = Unauthenticated desc = project id not available in request context",
		},
		{
			name:              "unauthorized response with mismatched tenproject ids across multiple roles",
			incomingContext:   ctxWithJwtRoles(t, validProjectId1+validRole, validProjectId2+validRole),
			expectedProjectId: DefaultProjectId,
			expectedErr:       "rpc error: code = Unauthenticated desc = mismatched project ids found in jwt roles",
		},
		{
			name:              "successfully extract project id from metadata",
			incomingContext:   ctxWithProjectIdKey(validProjectId1),
			expectedProjectId: validProjectId1,
		},
		{
			name:              "fail to extract project id from metadata with invalid id",
			incomingContext:   ctxWithProjectIdKey(invalidProjectId),
			expectedProjectId: DefaultProjectId,
			expectedErr:       "rpc error: code = Unauthenticated desc = project id not supplied",
		},
		{
			name:              "fail to extract project id from empty context",
			incomingContext:   context.TODO(),
			expectedProjectId: DefaultProjectId,
			expectedErr:       "rpc error: code = Unauthenticated desc = project id not supplied",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotNil(t, tc.incomingContext)

			projectId, err := extractProjectIdFromJwtRoles(tc.incomingContext)
			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedProjectId, projectId)
		})
	}
}
