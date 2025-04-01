// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package auth_echo

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bnkamalesh/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/labstack/echo/v4"

	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/logging"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/rbac"
)

type key string

// making it a constant to satisfy go-mnd linter.
const (
	authPairLen     = 2
	authKey     key = "authorization"
	bearer          = "bearer"
)

var (
	log = logging.GetLogger("auth_echo")
)

type AuthInterceptorConfig struct {
	UserAgent   string
	ServiceName string
}

/*
This function provides a convenient way to check if the authentication and authorization configuration is enabled for the REST API,
allowing for conditional logic based on the configuration status.
*/
func GetAuthRestConfig() bool {
	boolAuthRestEnabled := false
	authRestEnabled := os.Getenv("authRestEnabled")
	if authRestEnabled != "" {
		boolAuthRestEnabled, _ = strconv.ParseBool(authRestEnabled)
	}
	return boolAuthRestEnabled
}

func AuthenticationInterceptor(config AuthInterceptorConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !rbac.PolicyExistFlag {
				log.Error().Msgf("Can't upload RBAC realm policies to OPA package")

				return &echo.HTTPError{
					Code:     http.StatusForbidden,
					Message:  "Can't upload RBAC realm policies to OPA package",
					Internal: errors.New("Can't upload RBAC realm policies to OPA package"),
				}
			}

			authHeader := getAuthHeader(c)

			if authHeader == "" {
				// For internal service, there is no authHeader
				// For external service, validate-jwt middleware will do authentication
				return next(c)
			}

			authScheme, authToken, err := parseAuthHeader(authHeader)
			if err != nil {
				log.Error().Err(err).Msg("parse auth header failed")
				return err
			}

			claims := jwt.MapClaims{}
			_, _ = jwt.ParseWithClaims(authToken, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte("parse token claims"), nil
			})

			niceMd := metautils.ExtractIncoming(c.Request().Context())

			err = handleClaims(&niceMd, claims)
			if err != nil {
				log.Error().Err(err).Msg("convert claims to meta data failed")
				return err
			}

			err = rbac.Policies.Verify(niceMd, c.Request().Method)
			if err != nil {
				log.Error().Err(err).Msgf("request can't be authorized")
				return &echo.HTTPError{
					Code:     http.StatusForbidden,
					Message:  http.StatusText(http.StatusForbidden),
					Internal: errors.New(c.Request().Method + " request can't be authorized, error:" + err.Error()),
				}
			}

			if c.Request().Header.Get(config.UserAgent) == config.ServiceName {
				SetResourceRole(&niceMd, claims, c)
			}

			log.Debug().Msgf("JWT token is valid, proceeding with processing")
			// including token to the message metadata
			c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), authKey,
				strings.ToLower(authScheme)+" "+authToken)))

			return next(c)
		}
	}
}

func handleClaims(niceMd *metautils.NiceMD, claims jwt.MapClaims) error {
	for k, v := range claims {
		err := convertClaim(niceMd, []string{k}, v)
		if err != nil {
			return &echo.HTTPError{
				Code:     http.StatusForbidden,
				Message:  "convert claims to meta data failed",
				Internal: errors.New("error handling claim, error:" + err.Error()),
			}
		}
	}

	return nil
}

func convertClaim(niceMd *metautils.NiceMD, key []string, value interface{}) error {
	k := strings.Join(key, "/")
	switch vt := value.(type) {
	case string:
		niceMd.Set(k, vt)
	case float64:
		niceMd.Set(k, fmt.Sprintf("%v", vt))
	case bool:
		if vt {
			niceMd.Set(k, "true")
		} else {
			niceMd.Set(k, "false")
		}
	case []interface{}:
		for _, item := range vt {
			niceMd.Add(k, fmt.Sprintf("%v", item))
		}
	case map[string]interface{}:
		for k, v := range vt {
			err := convertClaim(niceMd, append(key, k), v)
			if err != nil {
				return err
			}
		}
	default:
		return errors.Validationf("metadata unhandled type %T", vt)
	}
	return nil
}

func SetResourceRole(niceMd *metautils.NiceMD, claims jwt.MapClaims, c echo.Context) {
	resourceAccessRole, _ := rbac.GetResourceRole(claims)
	role := resourceAccessRole
	log.Debug().Msgf("resource role: %v", role)
	c.Request().Header.Set(rbac.RoleRancherHeader, role)
}

func getAuthHeader(c echo.Context) string {
	authHeader := c.Request().Header.Get("authorization")
	if authHeader == "" {
		// re-try if the extraction is case-sensitive
		authHeader = c.Request().Header.Get("Authorization")
	}

	return authHeader
}

func parseAuthHeader(authHeader string) (string, string, error) {
	log.Debug().Msgf("parsing authorization header")
	authPair := strings.Split(strings.TrimSpace(authHeader), " ")
	if len(authPair) != authPairLen {
		return "", "", &echo.HTTPError{
			Code:     http.StatusUnauthorized,
			Message:  "wrong Authorization header definition",
			Internal: errors.New("wrong Authorization header definition"),
		}
	}

	authScheme := authPair[0]
	authToken := authPair[1]

	if !strings.EqualFold(authScheme, bearer) {
		return "", "", &echo.HTTPError{
			Code:    http.StatusUnauthorized,
			Message: "wrong Authorization header definition",
			Internal: errors.New("wrong Authorization header definition. " +
				"Expecting \"Bearer\" Scheme to be sent"),
		}
	}

	return authScheme, authToken, nil
}
