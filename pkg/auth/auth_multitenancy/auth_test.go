// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package authmultitenancy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// TestGetPublicKeyForIssuer tests the getPublicKeyForIssuer function using table-driven tests
func TestGetPublicKeyForIssuer(t *testing.T) {
	// Define your test cases
	testCases := []struct {
		name          string
		issuer        string
		kid           string
		jwksResponse  string
		expectSuccess bool
		token         jwt.Token
	}{
		{
			name:          "Valid kid",
			issuer:        "https://keycloak.example.com",
			kid:           "xiUPwPMCW0goopjfE-yapRF-c2YT_qu0YiL8CPJ6eLM",
			jwksResponse:  `{ "keys": [ { "kid": "xiUPwPMCW0goopjfE-yapRF-c2YT_qu0YiL8CPJ6eLM", "kty": "RSA", "alg": "RS256", "use": "sig", "n": "o92ECpv_sE8lnyIrHQzYXXJYtghzb4k2ABg3LqQBnhfc5RudiIx0cI5Dv8ErStgLbE843hFIsZ3QRPEQNmKYbtcFFuzuStZPn-3LBy1aSCslILOaZmFMR5uPawH6YZcAF2QQML2Ew_xSankQtFhKHsiV7cR1lcmaeDPKISlZk1vJ36VrdiJ0rqS5PtsKIxx2YqMosOwEnP5-gCPEsx77FCzGBKMw8UrK7m_nsrF-GQzICX3Tvc9EHoFf8UubxhL21JNkqMx1cbinxnU6mnypVoNbtGOH2DmbkrETmefOTErd0oXD3DZM2l0jgITNnFWp6bkwQ1PTiumChQLmKupTRQ", "e": "AQAB", "x5c": [ "MIICmzCCAYMCBgGSCpgNVjANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDDAZtYXN0ZXIwHhcNMjQwOTE5MTQwMTMwWhcNMzQwOTE5MTQwMzEwWjARMQ8wDQYDVQQDDAZtYXN0ZXIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCj3YQKm/+wTyWfIisdDNhdcli2CHNviTYAGDcupAGeF9zlG52IjHRwjkO/wStK2AtsTzjeEUixndBE8RA2Yphu1wUW7O5K1k+f7csHLVpIKyUgs5pmYUxHm49rAfphlwAXZBAwvYTD/FJqeRC0WEoeyJXtxHWVyZp4M8ohKVmTW8nfpWt2InSupLk+2wojHHZioyiw7ASc/n6AI8SzHvsULMYEozDxSsrub+eysX4ZDMgJfdO9z0QegV/xS5vGEvbUk2SozHVxuKfGdTqafKlWg1u0Y4fYOZuSsROZ585MSt3ShcPcNkzaXSOAhM2cVanpuTBDU9OK6YKFAuYq6lNFAgMBAAEwDQYJKoZIhvcNAQELBQADggEBAIvHcFE9eZxAC0+YyCiDqM9bl5LI/J7/NN+vWlVZcTxODhJkWHzukDksYOIw8YyFDK9chDUIfJgkqzQAlGvSGnvf4r0qd2VSqC7p+eFQYGuLbWkTL3AkAY0dRKW0i/3YGLApSE7p7uqn649Dvk0/yIZKxx2KWqAiSY0dyQSWtrzOCw/D+OFnILrzMnVDUxNP70H56BQtMG/43ADFcaYJZy7W/0O7kd1ltuO085PlVygVh7nEJjZHMw0geYD5cRkHOjgSoddUJra4MSArVC/PZ1pcQ2Sh+GVokWl/6MELevdU3FLFGM2xS732DN2Ml6tJn3tT77l7xm9VOkU+5faSPNc=" ], "x5t": "qamdDK3h6Sl3-zYPuuwdKoLOytI", "x5t#S256": "0PjqrPSPuDllMQokVxXls54PedXAkwnfOb7XDdSauhM" } ] }`,
			expectSuccess: true,
		},
		{
			name:          "Invalid kid",
			issuer:        "https://untrusted.example.com",
			kid:           "hehe",
			jwksResponse:  `{ "keys": [ { "kid": "xiUPwPMCW0goopjfE-yapRF-c2YT_qu0YiL8CPJ6eLM", "kty": "RSA", "alg": "RS256", "use": "sig", "n": "o92ECpv_sE8lnyIrHQzYXXJYtghzb4k2ABg3LqQBnhfc5RudiIx0cI5Dv8ErStgLbE843hFIsZ3QRPEQNmKYbtcFFuzuStZPn-3LBy1aSCslILOaZmFMR5uPawH6YZcAF2QQML2Ew_xSankQtFhKHsiV7cR1lcmaeDPKISlZk1vJ36VrdiJ0rqS5PtsKIxx2YqMosOwEnP5-gCPEsx77FCzGBKMw8UrK7m_nsrF-GQzICX3Tvc9EHoFf8UubxhL21JNkqMx1cbinxnU6mnypVoNbtGOH2DmbkrETmefOTErd0oXD3DZM2l0jgITNnFWp6bkwQ1PTiumChQLmKupTRQ", "e": "AQAB", "x5c": [ "MIICmzCCAYMCBgGSCpgNVjANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDDAZtYXN0ZXIwHhcNMjQwOTE5MTQwMTMwWhcNMzQwOTE5MTQwMzEwWjARMQ8wDQYDVQQDDAZtYXN0ZXIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCj3YQKm/+wTyWfIisdDNhdcli2CHNviTYAGDcupAGeF9zlG52IjHRwjkO/wStK2AtsTzjeEUixndBE8RA2Yphu1wUW7O5K1k+f7csHLVpIKyUgs5pmYUxHm49rAfphlwAXZBAwvYTD/FJqeRC0WEoeyJXtxHWVyZp4M8ohKVmTW8nfpWt2InSupLk+2wojHHZioyiw7ASc/n6AI8SzHvsULMYEozDxSsrub+eysX4ZDMgJfdO9z0QegV/xS5vGEvbUk2SozHVxuKfGdTqafKlWg1u0Y4fYOZuSsROZ585MSt3ShcPcNkzaXSOAhM2cVanpuTBDU9OK6YKFAuYq6lNFAgMBAAEwDQYJKoZIhvcNAQELBQADggEBAIvHcFE9eZxAC0+YyCiDqM9bl5LI/J7/NN+vWlVZcTxODhJkWHzukDksYOIw8YyFDK9chDUIfJgkqzQAlGvSGnvf4r0qd2VSqC7p+eFQYGuLbWkTL3AkAY0dRKW0i/3YGLApSE7p7uqn649Dvk0/yIZKxx2KWqAiSY0dyQSWtrzOCw/D+OFnILrzMnVDUxNP70H56BQtMG/43ADFcaYJZy7W/0O7kd1ltuO085PlVygVh7nEJjZHMw0geYD5cRkHOjgSoddUJra4MSArVC/PZ1pcQ2Sh+GVokWl/6MELevdU3FLFGM2xS732DN2Ml6tJn3tT77l7xm9VOkU+5faSPNc=" ], "x5t": "qamdDK3h6Sl3-zYPuuwdKoLOytI", "x5t#S256": "0PjqrPSPuDllMQokVxXls54PedXAkwnfOb7XDdSauhM" } ] }`,
			expectSuccess: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test server with a handler that simulates the JWKS endpoint
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				configResponse := fmt.Sprintf(`{ "jwks_uri": "http://%s/protocol/openid-connect/certs" }`, r.Host)
				switch r.URL.Path {
				case "/.well-known/openid-configuration":
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(configResponse))
				default:
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tc.jwksResponse))
				}
			}))
			defer ts.Close()

			// Override the issuer with the test server URL for testing
			issuer := ts.URL

			// Create an http.Client
			client := ts.Client()
			// Call getPublicKeyForIssuer with the test server URL as the issuer
			key, err := getPublicKeyForIssuer(issuer, tc.kid, client)

			if tc.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, key)
			} else {
				assert.Error(t, err)
				assert.Nil(t, key)
			}
		})
	}
}
