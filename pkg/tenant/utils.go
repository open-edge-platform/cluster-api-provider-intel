// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package tenant

import (
	"context"
)

func AddActiveProjectIdToContext(ctx context.Context, activeProjectId string) context.Context {
	if !isValidUuid(activeProjectId) {
		log.Panic().Msgf("cannot add invalid active project id to context: '%s'", activeProjectId)
	}

	return context.WithValue(ctx, ActiveProjectIdContextKey, activeProjectId)
}

func GetActiveProjectIdFromContext(ctx context.Context) string {
	activeProjectId, ok := ctx.Value(ActiveProjectIdContextKey).(string)
	if !ok || !isValidUuid(activeProjectId) {
		log.Panic().Msg("no valid active project id found in context")
	}

	return activeProjectId
}

func isValidUuid(uuid string) bool {
	return uuidRegex.MatchString(uuid)
}
