// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0
package auth_echo

import (
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/labstack/echo/v4"
	"github.com/naughtygopher/errors"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/rbac"
	"github.com/stretchr/testify/assert"
	"github.com/undefinedlabs/go-mpatch"
	"go.uber.org/mock/gomock"
)

var (
	errGeneralErr = errors.New("general error")
)

func makeBearerToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("test-signing-key"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	return "Bearer " + signedToken
}

func makeValidAuthHeader(t *testing.T) string {
	t.Helper()

	return makeBearerToken(t, jwt.MapClaims{
		"sub": "test-user",
	})
}

func makeResourceAuthHeader(t *testing.T) string {
	t.Helper()

	return makeBearerToken(t, jwt.MapClaims{
		"sub": "test-user",
		"resource_access": map[string]interface{}{
			"test-client": map[string]interface{}{
				"roles": []interface{}{rbac.RoleRancherReadOnly},
			},
		},
	})
}

func unpatchAll(t *testing.T, pList []*mpatch.Patch) {
	for _, p := range pList {
		if p != nil {
			if err := p.Unpatch(); err != nil {
				t.Errorf("unpatch error: %v", err)
			}
		}
	}
}

func patchRBACNew(t *testing.T, fail bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error

	patch, patchErr = mpatch.PatchMethod(rbac.New, func(ruleDirectory string) (*rbac.Policy, error) {
		if fail {
			return nil, errGeneralErr
		} else {
			return nil, nil
		}
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
	return patch
}

func patchVerify(t *testing.T, fail bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error

	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(rbac.Policy{}), "Verify",
		func(p *rbac.Policy, claims metautils.NiceMD, operation string) error {
			if fail {
				return errGeneralErr
			} else {
				return nil
			}
		})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
	return patch
}

func TestAuthenticationInterceptor(t *testing.T) {
	rbacRealmDirectory := "test/authz.rego"
	tests := []struct {
		name           string
		authHeader     string
		expectedError  string
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:          "Valid Token",
			authHeader:    makeValidAuthHeader(t),
			expectedError: "",
			funcBeforeTest: func() []*mpatch.Patch {
				rbac.Policies, _ = rbac.New(rbacRealmDirectory)
				rbac.PolicyExistFlag = true
				patch1 := patchRBACNew(t, false)
				patch2 := patchVerify(t, false)
				return []*mpatch.Patch{patch1, patch2}
			},
		},
		{
			name:          "Missing Token",
			authHeader:    "Bearer ",
			expectedError: "code=401, message=wrong Authorization header definition, internal=wrong Authorization header definition",
			funcBeforeTest: func() []*mpatch.Patch {
				rbac.Policies, _ = rbac.New(rbacRealmDirectory)
				rbac.PolicyExistFlag = true
				patch1 := patchRBACNew(t, false)
				patch2 := patchVerify(t, true)
				return []*mpatch.Patch{patch1, patch2}
			},
		},
		{
			name:          "Invalid Token",
			authHeader:    "Bearer invalid-token",
			expectedError: "code=403, message=Forbidden",
			funcBeforeTest: func() []*mpatch.Patch {
				rbac.Policies, _ = rbac.New(rbacRealmDirectory)
				rbac.PolicyExistFlag = true
				patch1 := patchRBACNew(t, false)
				patch2 := patchVerify(t, true)
				return []*mpatch.Patch{patch1, patch2}
			},
		},
		{
			name:          "Invalid schema",
			authHeader:    "dsf invalid-token",
			expectedError: "code=401, message=wrong Authorization header definition, internal=wrong Authorization header definition. Expecting \"Bearer\" Scheme to be sent",
			funcBeforeTest: func() []*mpatch.Patch {
				rbac.Policies, _ = rbac.New(rbacRealmDirectory)
				rbac.PolicyExistFlag = true
				patch1 := patchRBACNew(t, false)
				patch2 := patchVerify(t, true)
				return []*mpatch.Patch{patch1, patch2}
			},
		},
		{
			name:          "RBAC policy file is empty",
			authHeader:    "dsf invalid-token",
			expectedError: "code=403, message=Can't upload RBAC realm policies to OPA package",
			funcBeforeTest: func() []*mpatch.Patch {
				rbac.Policies, _ = rbac.New("")
				rbac.PolicyExistFlag = false
				patch1 := patchRBACNew(t, true)
				patch2 := patchVerify(t, true)
				return []*mpatch.Patch{patch1, patch2}
			},
		},
		{
			name:          "empty auth header",
			authHeader:    "",
			expectedError: "",
			funcBeforeTest: func() []*mpatch.Patch {
				rbac.Policies, _ = rbac.New(rbacRealmDirectory)
				rbac.PolicyExistFlag = true
				patch1 := patchRBACNew(t, false)
				patch2 := patchVerify(t, true)
				return []*mpatch.Patch{patch1, patch2}
			},
		},
		{
			name:          "Valid Token with resouce verification",
			authHeader:    makeResourceAuthHeader(t),
			expectedError: "",
			funcBeforeTest: func() []*mpatch.Patch {
				rbac.Policies, _ = rbac.New(rbacRealmDirectory)
				rbac.PolicyExistFlag = true
				patch1 := patchRBACNew(t, false)
				patch2 := patchVerify(t, false)
				return []*mpatch.Patch{patch1, patch2}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.funcBeforeTest != nil {
				pList := tt.funcBeforeTest()
				defer unpatchAll(t, pList)
			} // Define your test configuration values
			testUserAgent := "TestUserAgent"
			testServiceName := "TestServiceName"

			// Create a configuration instance for testing
			testConfig := AuthInterceptorConfig{
				UserAgent:   testUserAgent,
				ServiceName: testServiceName,
			}
			// Set up the interceptor
			interceptor := AuthenticationInterceptor(testConfig)(func(c echo.Context) error {
				// This is the next handler in the chain, do any assertions or logic here if needed
				return nil
			})
			// Create a mock request with the authentication header
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", tt.authHeader)

			if tt.name == "Valid Token with resouce verification" {
				req.Header.Set(testUserAgent, testServiceName)
			}
			// Create a mock response recorder
			rec := httptest.NewRecorder()

			// Create a new Echo context
			e := echo.New()
			ctx := e.NewContext(req, rec)

			// Invoke the interceptor
			err := interceptor(ctx)

			// Assert the error and HTTP status code
			if len(tt.expectedError) > 0 {
				assert.ErrorContains(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetAuthRestConfig(t *testing.T) {
	patch_getAuthCfg := func(ctrl *gomock.Controller) []*mpatch.Patch {
		patch_getAuthCfgTrue, patchErr := mpatch.PatchMethod(os.Getenv, func(key string) string {
			return "true"
		})
		if patchErr != nil {
			t.Errorf("patch error: %v", patchErr)
		}

		return []*mpatch.Patch{patch_getAuthCfgTrue}
	}

	tests := []struct {
		name           string
		want           bool
		funcBeforeTest func(*gomock.Controller) []*mpatch.Patch
	}{
		// TODO: Add test cases.
		{
			name:           "Get auth config -- disable",
			want:           false,
			funcBeforeTest: nil,
		},
		// TODO: Add test cases.
		{
			name:           "Get auth config -- enable",
			want:           true,
			funcBeforeTest: patch_getAuthCfg,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			if tt.funcBeforeTest != nil {
				plist := tt.funcBeforeTest(ctrl)
				defer unpatchAll(t, plist)
			}
			if got := GetAuthRestConfig(); got != tt.want {
				t.Errorf("GetAuthRestConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
