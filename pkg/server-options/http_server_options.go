// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_options

import (
	"net/http"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echo_middleware "github.com/oapi-codegen/echo-middleware"
	"go.opentelemetry.io/otel/trace"

	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/auth/auth_echo"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/logging"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/tenant"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/tracing"
)

var log = logging.GetLogger("middleware")

const (
	corsMaxAge               = 600
	rateMaxRequestsPerSecond = 100
	rateExpirePeriod         = 3
	// While a 1 MB limit on HTTP request sizes is reasonable for most applications, it might restrict file upload speeds where applicable
	bodyLimitMax             = "100K"
	headerLimitMax           = http.DefaultMaxHeaderBytes
	serverDefaultTimeout     = 60
	serverDefaultIdleTimeout = 60
)

var (
	noCacheHeaders = map[string]string{
		"Expires":       "-1",
		"Cache-Control": "no-cache, no-store, max-age=0, must-revalidate",
		"Pragma":        "no-cache",
	}
	DefaultNoCacheConfig = noCacheConfig{
		Skipper: middleware.DefaultSkipper,
	}
)

type noCacheConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper middleware.Skipper
}

func SetCors(e *echo.Echo, cors string) {
	log.Debug().Msg("configure CORS")
	corsOrigins := strings.Split(cors, ",")
	if len(corsOrigins) > 0 {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: corsOrigins,
			AllowHeaders: []string{
				echo.HeaderAccessControlAllowOrigin,
				echo.HeaderContentType,
				echo.HeaderAuthorization,
				echo.HeaderAccept,
			},
			AllowMethods: []string{
				http.MethodGet,
				http.MethodHead,
				http.MethodPut,
				http.MethodPatch,
				http.MethodPost,
				http.MethodDelete,
				http.MethodOptions,
			},
			MaxAge: corsMaxAge,
		}))
	}
}

// SetMethodOverride set config that prevents method override in HTTP header.
func SetMethodOverride(e *echo.Echo) {
	log.Info().Msg("configure MethodOverride")
	e.Pre(middleware.MethodOverrideWithConfig(middleware.MethodOverrideConfig{
		Getter: nil,
	}))
}

// SetSecureConfig defines some of the secure config related to HTTP headers.
// Definitions are set based on:
// https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Headers_Cheat_Sheet.html
// https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html
func SetSecureConfig(e *echo.Echo) {
	log.Debug().Msg("Configure SecureConfig")
	secureConfig := middleware.SecureConfig{
		XFrameOptions:         "DENY",
		XSSProtection:         "0",
		ContentTypeNosniff:    "nosniff",
		ContentSecurityPolicy: "default-src 'self'; frame-ancestors 'self'; form-action 'self'",
		HSTSMaxAge:            1024000,
		HSTSExcludeSubdomains: false,
		Skipper: func(c echo.Context) bool {
			return false
		},
	}

	e.Use(middleware.SecureWithConfig(secureConfig))
}

// SetRateLimiter sets the rate limiter to the server.
func SetRateLimiter(e *echo.Echo) {
	log.Debug().Msg("Configure Rate Limiter")
	config := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      rateMaxRequestsPerSecond,
				Burst:     rateMaxRequestsPerSecond,
				ExpiresIn: rateExpirePeriod * time.Minute,
			},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return context.JSON(http.StatusForbidden, nil)
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(http.StatusTooManyRequests, nil)
		},
	}

	e.Use(middleware.RateLimiterWithConfig(config))
}

// SetLimits sets the max size of a request body
// and the max size of header bytes.
func SetLimits(e *echo.Echo) {
	log.Debug().Msg("Configure Header/Body")
	e.Use(middleware.BodyLimit(bodyLimitMax))
	e.Server.MaxHeaderBytes = headerLimitMax
}

// SetTimeout sets the timeout of a request.
func SetTimeout(e *echo.Echo) {
	log.Debug().Msg("Configure Timeout")
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		ErrorMessage: "request timeout",
		Timeout:      serverDefaultTimeout * time.Second,
	}))
	e.Server.ReadTimeout = serverDefaultTimeout * time.Second
	e.Server.WriteTimeout = serverDefaultTimeout * time.Second
	e.Server.IdleTimeout = serverDefaultIdleTimeout * time.Second
}

// NoCacheWithConfig returns a nocache middleware with config.
func noCacheWithConfig(config noCacheConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultNoCacheConfig.Skipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.Skipper(c) {
				return next(c)
			}
			// Set NoCache to headers
			res := c.Response()
			for k, v := range noCacheHeaders {
				res.Header().Set(k, v)
			}

			return next(c)
		}
	}
}

func SetNoCache(e *echo.Echo) {
	e.Use(noCacheWithConfig(DefaultNoCacheConfig))
}

func httpReqHeaderProcess(header map[string]string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			for k, v := range header {
				c.Request().Header.Set(k, v)
			}
			return next(c)
		}
	}
}

func SetHttpReqHeader(e *echo.Echo, header map[string]string) {
	if header != nil {
		e.Use(httpReqHeaderProcess(header))
	}
}

// SetOptions sets all options to echo.Echo defined in this file.
func SetOptions(e *echo.Echo, cors string) {
	log.Debug().Msg("Configure http server options")
	// NOTE the CORS middleware has to be the first one
	// if not OPTIONS pre-flights are denied by the OapiRequestValidator middleware
	SetCors(e, cors)
	SetSecureConfig(e)
	SetRateLimiter(e)
	SetLimits(e)
	SetTimeout(e)
	SetNoCache(e)
	SetProjectIdInterceptor(e)
	e.HideBanner = true
	e.HidePort = true
}

// SetTracing enables tracing and also starts the trace logger
func SetTracing(e *echo.Echo, serviceName string) {
	log.Debug().Msg("set tracing")
	tracing.EnableEchoAutoTracing(e, serviceName)
	SetTraceLogger(e)
}

func SetAuthInterceptor(e *echo.Echo, config auth_echo.AuthInterceptorConfig) {
	// enable authentication interceptor in the rest
	e.Use(auth_echo.AuthenticationInterceptor(config))
}

func SetRequestValidator(e *echo.Echo, openapi *openapi3.T) {
	// Validate incoming request against the OpenAPI Swagger Spec.
	e.Use(echo_middleware.OapiRequestValidator(openapi))
}

func SetProjectIdInterceptor(e *echo.Echo) {
	log.Debug().Msg("set project id interceptor")
	e.Use(tenant.ActiveProjectIdEchoMiddleware())
}

// SetTraceLogger logs trace and span IDs along with Method and URI. This is useful to know the traceID that
// we would like to check on the Grafana Tempo UI
func SetTraceLogger(e *echo.Echo) {
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod: true,
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			// log tracing information at debug level which otherwise would generate a lot of logs at Info level
			log.Debug().
				Str("Method", v.Method).
				Str("URI", v.URI).
				Int("status", v.Status).
				Msgf("traceID: %v spanID: %v",
					trace.SpanFromContext(c.Request().Context()).SpanContext().TraceID(),
					trace.SpanFromContext(c.Request().Context()).SpanContext().SpanID())
			return nil
		},
	}))
}
