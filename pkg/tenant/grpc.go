// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package tenant

import (
	"context"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/rbac"
)

// ActiveProjectIdGrpcInterceptor returns an interceptor to extract the active project id from jwt and provide it in the context.
func ActiveProjectIdGrpcInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		projectId, err := extractProjectIdFromJwtRoles(ctx)
		if err != nil {
			return nil, err
		}

		log.Trace().Msgf("project id intercepted in grpc request: '%s'", projectId)
		return handler(AddActiveProjectIdToContext(ctx, projectId), req)
	}
}

// extractProjectIdFromJwtRoles infers the project id from the m2mClientRole in the jwt found in the context.
func extractProjectIdFromJwtRoles(ctx context.Context) (string, error) {
	projectId := DefaultProjectId

	// error ignored to handle case from ctm where jwt not supplied
	token, _ := rbac.ExtractAuthorizationFromMd(ctx)
	if token == "" {
		log.Debug().Msg("jwt token not found in grpc request, extracting project id from context metadata")
		projectId = metautils.ExtractIncoming(ctx).Get(ActiveProjectIdHeaderKey)
		if projectId == "" || !isValidUuid(projectId) {
			return DefaultProjectId, status.New(codes.Unauthenticated, "project id not supplied").Err()
		}
		return projectId, nil
	}

	// TODO! refactor to avoid double authentication
	md, _ := rbac.VerifyContextClaims(ctx)
	roles, ok := md[jwtRolesKey]
	if !ok {
		return DefaultProjectId, status.New(codes.Unauthenticated, "no roles found in jwt").Err()
	}

	for _, role := range roles {
		if !strings.Contains(role, roleProjectIdSeparator+m2mClientRole) {
			continue
		}

		tid := strings.Split(role, roleProjectIdSeparator)[0]
		if !isValidUuid(tid) {
			continue
		}

		if projectId == DefaultProjectId {
			projectId = tid
		}

		if projectId != tid {
			return DefaultProjectId, status.New(codes.Unauthenticated, "mismatched project ids found in jwt roles").Err()
		}
	}

	if projectId == DefaultProjectId {
		return DefaultProjectId, status.New(codes.Unauthenticated, "project id not available in request context").Err()
	}

	return projectId, nil
}
