// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package tenant

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	testutils "github.com/open-edge-platform/cluster-api-provider-intel/pkg/testing"
)

var (
	validProjectId1  = "7642fcd0-d997-4ad4-b7a3-95e2cf8e3095"
	validProjectId2  = "df95e679-bac4-460c-99ff-e0b17b8562c4"
	invalidProjectId = "invalid-id"

	validRole   = roleProjectIdSeparator + m2mClientRole
	invalidRole = roleProjectIdSeparator + "cluster-read-role"
)

func ctxWithJwtRoles(t *testing.T, roles ...string) context.Context {
	_, jwt, err := testutils.CreateJWT(t)
	require.NoError(t, err)

	return metadata.NewIncomingContext(context.Background(), metadata.MD{
		testutils.AuthKey: []string{testutils.BearerPrefixLowercase + jwt},
		jwtRolesKey:       roles,
	})
}

func ctxWithProjectIdKey(projectId string) context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.MD{
		ActiveProjectIdHeaderKey: []string{projectId},
	})
}
