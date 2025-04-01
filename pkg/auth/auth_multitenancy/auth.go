// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package authmultitenancy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/tenant"
	"github.com/open-policy-agent/opa/rego"
)

const (
	// OpenidConfiguration is the discovery point on the OIDC server
	OpenidConfiguration = ".well-known/openid-configuration"
)

var (
	jwksCache *jwk.Cache
)

func init() {
	// Initialize the JWKS cache
	jwksCache = jwk.NewCache(context.Background())
}

type AuthMiddlewareConfig struct {
	KeycloakEndpoint string
	RegosFilePaths   []string
	JwtSigningMethod string // IN 24.08 release is RS256
	Client           http.Client
	Skipper          middleware.Skipper
}

// Structure that represents the input to OPA policy.
type authzInput struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	User   struct {
		Permissions []string `json:"permissions"`
	} `json:"user"`
	ProjectID string `json:"project_id"`
}

type providerJSON struct {
	Issuer        string   `json:"issuer"`
	AuthURL       string   `json:"authorization_endpoint"`
	TokenURL      string   `json:"token_endpoint"`
	DeviceAuthURL string   `json:"device_authorization_endpoint"`
	JWKSURL       string   `json:"jwks_uri"`
	UserInfoURL   string   `json:"userinfo_endpoint"`
	Algorithms    []string `json:"id_token_signing_alg_values_supported"`
}

// getKeyFunc returns a jwt.Keyfunc that dynamically selects the key based on the issuer
func getKeyFunc(config AuthMiddlewareConfig) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, errors.New("invalid token claims")
		}

		_, ok = claims["iss"].(string)
		if !ok {
			return nil, errors.New("issuer claim 'iss' not found")
		}

		keyID, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("token header 'kid' not found")
		}

		return getPublicKeyForIssuer(config.KeycloakEndpoint, keyID, &config.Client)
	}
}

func getJKWSURL(issuer string, client *http.Client) (string, error) {
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("%s/%s", issuer, OpenidConfiguration), nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	resOpenIDConfig, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("error obtaining information from OIDC well-known URL: %w", err)
	}
	if resOpenIDConfig.Body != nil {
		defer resOpenIDConfig.Body.Close()
	}
	openIDConfigBody, readErr := io.ReadAll(resOpenIDConfig.Body)
	if readErr != nil {
		return "", fmt.Errorf("error reading Body of the OIDC configuration: %w", readErr)
	}
	var openIDprovider providerJSON
	jsonErr := json.Unmarshal(openIDConfigBody, &openIDprovider)
	if jsonErr != nil {
		return "", fmt.Errorf("error unmarshalling OIDC configuration: %w", jsonErr)
	}
	return openIDprovider.JWKSURL, nil
}

// getPublicKeyForIssuer fetches the public key from the JWKS endpoint of keycloak
func getPublicKeyForIssuer(issuer string, kid string, client *http.Client) (interface{}, error) {
	// Fetch the JWKS from the keycloak's endpoint
	jwksEndpoint, err := getJKWSURL(issuer, client)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWKS URL: %w", err)
	}

	ctx := context.Background()

	// Configure the cache to fetch keys from the JWKS endpoint
	_ = jwksCache.Register(jwksEndpoint, jwk.WithHTTPClient(client), jwk.WithMinRefreshInterval(15*time.Minute))

	// Fetch the JWKS set from the cache
	set, err := jwksCache.Get(ctx, jwksEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	// Look for the key that was used to sing jwt token key
	key, found := set.LookupKeyID(kid)
	if !found {
		return nil, errors.New("key not found")
	}

	// Convert the JWK to a public key
	var rawKey interface{}
	if err = key.Raw(&rawKey); err != nil {
		return nil, fmt.Errorf("failed to create public key: %w", err)
	}

	return rawKey, nil
}

func jwtErrorHandler(c echo.Context, err error) error {
	// Log the error
	c.Logger().Errorf("JWT validation error: %v URL %v", err, c.Request().RequestURI)

	// Respond with an appropriate error message
	return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired JWT")
}

// extractRolesFromToken extracts roles from the JWT token.
func extractRolesFromToken(user *jwt.Token) ([]string, error) {
	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	realmAccess, ok := claims["realm_access"].(map[string]interface{})
	if !ok {
		return nil, errors.New("realm_access claim is missing or invalid")
	}

	rolesInterface, ok := realmAccess["roles"].([]interface{})
	if !ok {
		return nil, errors.New("roles are missing or invalid in realm_access claim")
	}

	roles := make([]string, 0, len(rolesInterface))
	for _, role := range rolesInterface {
		roleStr, validString := role.(string)
		if !validString {
			continue // Ignore non-string roles
		}
		roles = append(roles, roleStr)
	}

	return roles, nil
}

// constructAuthzInput constructs the input for the OPA policy.
func constructAuthzInput(c echo.Context, roles []string) authzInput {
	requestProjectID := c.Request().Header.Get(tenant.ActiveProjectIdHeaderKey)
	return authzInput{
		Method: c.Request().Method,
		Path:   c.Request().URL.Path,
		User: struct {
			Permissions []string `json:"permissions"`
		}{Permissions: roles},
		ProjectID: requestProjectID,
	}
}

// evaluatePolicy evaluates the OPA policy with the given input.
func evaluatePolicy(query rego.PreparedEvalQuery, input authzInput) (bool, error) {
	results, err := query.Eval(context.Background(), rego.EvalInput(input))
	if err != nil {
		return false, err
	}

	if len(results) == 0 || len(results[0].Expressions) == 0 {
		return false, errors.New("no result from policy evaluation")
	}

	allowed, ok := results[0].Expressions[0].Value.(bool)
	if !ok {
		return false, errors.New("policy result is not a boolean")
	}

	return allowed, nil
}

// opaAuthzMiddleware is the OPA middleware to authorize (RBAC) requests.
func opaAuthzMiddleware(query rego.PreparedEvalQuery, skipper middleware.Skipper) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if the middleware should be skipped
			if skipper != nil && skipper(c) {
				return next(c)
			}

			user, ok := c.Get("user").(*jwt.Token)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "user token is missing or invalid")
			}

			roles, err := extractRolesFromToken(user)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
			}

			input := constructAuthzInput(c, roles)

			allowed, err := evaluatePolicy(query, input)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Error evaluating policy")
			}

			if allowed {
				return next(c)
			}

			return echo.NewHTTPError(http.StatusForbidden, "Access denied")
		}
	}
}

func SetupAuthMiddleware(e *echo.Echo, config AuthMiddlewareConfig) error {
	jwtMiddleware := echojwt.WithConfig(echojwt.Config{
		Skipper:       config.Skipper,
		SigningMethod: config.JwtSigningMethod,
		KeyFunc:       getKeyFunc(config),
		ErrorHandler:  jwtErrorHandler,
	})
	e.Use(jwtMiddleware)

	re := rego.New(
		rego.Query("data.httpapi.authz.allow"),
		rego.Load(config.RegosFilePaths, nil),
	)
	query, err := re.PrepareForEval(context.Background())
	if err != nil {
		return err
	}
	e.Use(opaAuthzMiddleware(query, config.Skipper))
	return nil
}
