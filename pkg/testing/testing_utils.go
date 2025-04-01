// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package testing_utils

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/stretchr/testify/require"
)

const (
	// SharedSecretKey environment variable name for shared secret key for signing a token.
	SharedSecretKey       = "SHARED_SECRET_KEY"
	secretKey             = "randomSecretKey"
	readRole              = "clusters-read-role"
	writeRole             = "clusters-write-role"
	nodeAgentRole         = "node-agent-readwrite-role"
	AuthKey               = "authorization"
	BearerPrefixLowercase = "bearer "
)

// CreateJWT returns random signing key and JWT token (HS256 encoded) in a string with both roles, read and write.
// Only 1 token can persist in the system (otherwise, env variable holding secret key would be re-written).
func CreateJWT(tb testing.TB) (string, string, error) {
	tb.Helper()
	claims := &jwt.MapClaims{
		"iss": "https://keycloak.kind.internal/realms/master",
		"exp": time.Now().Add(time.Hour).Unix(),
		"typ": "Bearer",
		"realm_access": map[string]interface{}{
			"roles": []string{
				writeRole,
				readRole,
				nodeAgentRole,
			},
		},
	}

	return CreateJWTWithClaims(tb, claims)
}

// CreateJWTWithReadRole returns random signing key and JWT token (HS256 encoded) in a string with only write role.
// Only 1 token can persist in the system (otherwise, env variable holding secret key would be re-written).
func CreateJWTWithReadRole(tb testing.TB) (string, string, error) {
	tb.Helper()
	claims := &jwt.MapClaims{
		"iss": "https://keycloak.kind.internal/realms/master",
		"exp": time.Now().Add(time.Hour).Unix(),
		"typ": "Bearer",
		"realm_access": map[string]interface{}{
			"roles": []string{
				readRole,
			},
		},
	}

	return CreateJWTWithClaims(tb, claims)
}

// CreateJWTWithWriteRole returns random signing key and JWT token (HS256 encoded) in a string with only write role.
// Only 1 token can persist in the system (otherwise, env variable holding secret key would be re-written).
func CreateJWTWithWriteRole(tb testing.TB) (string, string, error) {
	tb.Helper()
	claims := &jwt.MapClaims{
		"iss": "https://keycloak.kind.internal/realms/master",
		"exp": time.Now().Add(time.Hour).Unix(),
		"typ": "Bearer",
		"realm_access": map[string]interface{}{
			"roles": []string{
				writeRole,
			},
		},
	}

	return CreateJWTWithClaims(tb, claims)
}

// CreateJWTWithClaims returns random signing key and JWT token (HS256 encoded) in a string with defined claims.
func CreateJWTWithClaims(tb testing.TB, claims *jwt.MapClaims) (string, string, error) {
	tb.Helper()

	tb.Setenv(SharedSecretKey, secretKey)
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims)
	jwtStr, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", "", err
	}
	return secretKey, jwtStr, nil
}

// CreateContextWithJWT can be used only with test clients, which send the request to the server.
func CreateContextWithJWT(tb testing.TB) (context.Context, context.CancelFunc) {
	tb.Helper()
	return CreateContextWithTimeoutWithJWT(tb, time.Second)
}

func AddJWTToTheOutgoingContext(ctx context.Context, jwtToken string) context.Context {
	niceMD := metautils.NiceMD{}
	niceMD.Add(AuthKey, "Bearer "+jwtToken)
	return niceMD.ToOutgoing(ctx)
}

func AddJWTToTheIncomingContext(ctx context.Context, jwtToken string) context.Context {
	niceMD := metautils.NiceMD{}
	niceMD.Add(AuthKey, "Bearer "+jwtToken)
	return niceMD.ToIncoming(ctx)
}

// CreateContextWithTimeoutWithJWT can be used only with test clients, which send the request to the server.
func CreateContextWithTimeoutWithJWT(tb testing.TB, timeout time.Duration) (context.Context, context.CancelFunc) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// adding JWT token to the context
	_, jwtToken, err := CreateJWT(tb)
	require.NoError(tb, err)
	return AddJWTToTheOutgoingContext(ctx, jwtToken), cancel
}

// CreateIncomingContextWithJWT can be used to test the actual call, which performs the authorization.
func CreateIncomingContextWithJWT(tb testing.TB) (context.Context, context.CancelFunc) {
	tb.Helper()
	return CreateIncomingContextWithTimeoutWithJWT(tb, time.Second)
}

// CreateIncomingContextWithTimeoutWithJWT can be used to test the actual call, which performs the authorization.
func CreateIncomingContextWithTimeoutWithJWT(tb testing.TB, timeout time.Duration) (context.Context, context.CancelFunc) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// adding JWT token to the context
	_, jwtToken, err := CreateJWT(tb)
	require.NoError(tb, err)
	return AddJWTToTheIncomingContext(ctx, jwtToken), cancel
}
