// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package tenant

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

var (
	projectIdNotProvidedError = echo.NewHTTPError(http.StatusUnauthorized, "request unauthorized, active project id not provided")
	projectIdInvalidError     = echo.NewHTTPError(http.StatusUnauthorized, "request unauthorized, active project id invalid")
)

// ActiveProjectIdEchoMiddleware returns an echo middleware function that extracts the active project id from the request
// header and adds it to the echo request context.
func ActiveProjectIdEchoMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// skip project id check for healthz route
			if c.Path() == "/v1/healthz" {
				return next(c)
			}

			projectId, err := getActiveProjectIdFromRequest(c.Request())
			if err != nil {
				return err
			}

			ctx := AddActiveProjectIdToContext(c.Request().Context(), projectId)
			c.SetRequest(c.Request().WithContext(ctx))

			log.Trace().Msgf("project id intercepted in http request: '%s'", projectId)
			return next(c)
		}
	}
}

// getActiveProjectIdFromRequest extracts the active project id from the request header using the ActiveProjectIdHeaderKey.
func getActiveProjectIdFromRequest(request *http.Request) (string, error) {
	activeProjectIds := request.Header.Values(ActiveProjectIdHeaderKey)
	switch len(activeProjectIds) {
	case 0:
		log.Error().Msgf("received request without an active project id")
		return DefaultProjectId, projectIdNotProvidedError
	case 1:
	default:
		log.Warn().Msgf("received request with multiple active project ids %s, using: %s", activeProjectIds, activeProjectIds[0])
	}

	if !isValidUuid(activeProjectIds[0]) {
		log.Error().Msgf("received request with invalid active project id: '%s'", activeProjectIds[0])
		return DefaultProjectId, projectIdInvalidError
	}

	return activeProjectIds[0], nil
}
