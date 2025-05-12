// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// Package rbac implements utility functions for Role-Based Access Control
package rbac

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/open-edge-platform/orch-library/go/pkg/auth"

	"github.com/bnkamalesh/errors"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	grpcauth "github.com/open-edge-platform/orch-library/go/pkg/grpc/auth"
	"github.com/open-policy-agent/opa/v1/rego"
	"k8s.io/utils/strings/slices"

	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/logging"
	"github.com/rs/zerolog"
)

const (
	// global variables
	ClusterRoleWrite = "clusters-write-role"
	ClusterRoleRead  = "clusters-read-role"

	TemplateRoleWrite = "cluster-templates-write-role"
	TemplateRoleRead  = "cluster-templates-read-role"

	RoleRancherHeader = "rancher-role-header"

	RbacDirectory = "/rego/authz.rego"

	// internal variables
	RoleRancherAdmin     = "restricted-role"
	RoleRancherReadWrite = "standard-role"
	RoleRancherReadOnly  = "base-role"

	MethodPost     = "POST"
	MethodPut      = "PUT"
	MethodDelete   = "DELETE"
	MethodGet      = "GET"
	MethodPatch    = "PATCH"
	MethodUpdate   = "UPDATE"
	MethodRegister = "REGISTER"

	resourceAccess = "resource_access"
	rolesKey       = "roles"

	adminAccess = "data.authz.hasAdminAccess"
	writeAccess = "data.authz.hasWriteAccess"
	readAccess  = "data.authz.hasReadAccess"

	// contextMetadataBearerKeyLower metadata JWT token key.
	contextMetadataBearerKeyLower = "bearer"
	// contextMetadataBearerKeyCamel metadata JWT token key.
	contextMetadataBearerKeyCamel = "Bearer"
	allowMissingAuthClients       = "ALLOW_MISSING_AUTH_CLIENTS"
	clientKeyLower                = "client"
	clientKeyCamel                = "Client"
	authKey                       = "authorization"
)

var (
	log      = logging.GetLogger("CO-RBAC")
	adminKey = "admin"
	rwKey    = "write"
	roKey    = "read"

	allRancherValidRoles = []string{RoleRancherAdmin, RoleRancherReadWrite, RoleRancherReadOnly}
	Policies             *Policy
	PolicyExistFlag      bool
)

type Policy struct {
	queries map[string]*rego.PreparedEvalQuery
}

func loadQuery(ctx context.Context, policies *Policy, ruleDirectory, queryName, queryCode string) error {
	query, err := rego.New(
		rego.Query(queryCode),
		rego.Load([]string{ruleDirectory}, nil),
	).PrepareForEval(ctx)
	if err != nil {
		return errors.InternalErr(err, fmt.Sprintf("can't load %s query", queryName))
	}
	customLog(zerolog.DebugLevel, fmt.Sprintf("loadQuery queryName is %s, query is %v, query module is %v", queryName, query, query.Modules()))
	policies.queries[queryName] = &query
	return nil
}

func New(ruleDirectory string) (*Policy, error) {
	ctx := context.Background()

	policies := Policy{
		queries: make(map[string]*rego.PreparedEvalQuery),
	}

	if err := loadQuery(ctx, &policies, ruleDirectory, adminKey, adminAccess); err != nil {
		return nil, err
	}

	if err := loadQuery(ctx, &policies, ruleDirectory, rwKey, writeAccess); err != nil {
		return nil, err
	}

	if err := loadQuery(ctx, &policies, ruleDirectory, roKey, readAccess); err != nil {
		return nil, err
	}

	return &policies, nil
}

func (p *Policy) Verify(claims metautils.NiceMD, operation string) error {
	switch operation {
	case MethodGet:
		if err := p.evaluateQuery(roKey, claims, operation); err != nil {
			return err
		}

		return nil
	case MethodPost, MethodPut, MethodDelete, MethodPatch, MethodRegister, MethodUpdate:
		if err := p.evaluateQuery(rwKey, claims, operation); err != nil {
			return err
		}
		return nil
	default:
		return errors.Internalf("permission denied, authorization failed - unsupported operation: %s", operation)
	}
}

func (p *Policy) evaluateQuery(queryKey string, claims metautils.NiceMD, operation string) error {
	log.Trace().Msgf("evaluateQuery, queryKey is %v, operation is %v", queryKey, operation)
	query, ok := p.queries[queryKey]
	if !ok {
		return errors.Internalf("permission denied, can not get query for %s", operation)
	}
	result, err := query.Eval(context.Background(), rego.EvalInput(claims))
	if err != nil {
		return errors.Internalf("permission denied, received error: %s for %s", err.Error(), operation)
	}
	log.Trace().Msgf("evaluateQuery, result.Allowed() is %v", result.Allowed())
	if !result.Allowed() {
		return errors.Internalf("permission denied, %s is not allowed by OPA", operation)
	}
	return nil
}

// RequestIsAuthorized function validates the JWT token included in a context.
// It also starts the OPA instance and performs the RBAC authorization of the call.
func (p *Policy) RequestIsAuthorized(ctx context.Context, operation string) bool {
	// check if client is the one which is set to bypass authorization
	if clientCanBypassAuthN(ctx) {
		return true
	}

	// assuming that the client should not bypass authorization
	md, err := VerifyContextClaims(ctx)
	if err != nil {
		return false
	}
	// performing RBAC authorization
	err = p.Verify(md, operation)
	return err == nil
}

// clientCanBypassAuthN checks if user can bypass AuthN (authentication + authorization) by
// checking the environmental variable.
func clientCanBypassAuthN(ctx context.Context) bool {
	niceMd := metautils.ExtractIncoming(ctx)
	acceptNoAuth := os.Getenv(allowMissingAuthClients)
	if acceptNoAuth == "" {
		// no clients to bypass AuthN specified
		return false
	}
	allowedMissingClients := strings.Split(acceptNoAuth, ",")
	requestClient := niceMd.Get(clientKeyLower)
	if requestClient == "" {
		// re-try to read with the other key
		requestClient = niceMd.Get(clientKeyCamel)
		if requestClient == "" {
			// no client name specified in the context, AuthN should be performed
			return false
		}
	}
	var foundMissingAuthClient bool
	for _, amc := range allowedMissingClients {
		if strings.ToLower(requestClient) == strings.TrimSpace(strings.ToLower(amc)) {
			foundMissingAuthClient = true
			break
		}
	}
	if foundMissingAuthClient {
		customLog(zerolog.WarnLevel, fmt.Sprintf("Allowing unauthenticated gRPC request from client: %s", niceMd.Get("client")))
		return true
	}

	log.Trace().Msgf("Client %s is not allowed to bypass authorization", requestClient)
	return false
}

func VerifyContextClaims(ctx context.Context) (metautils.NiceMD, error) {
	niceMd := metautils.ExtractIncoming(ctx)

	// Extract token from metadata in the context
	token, err := ExtractAuthorizationFromMd(ctx)
	if err != nil {
		log.Error().Err(err).Msgf("failed to extract jwt from context")
		return nil, err
	}

	// Authenticate the jwt token
	jwtAuth := new(auth.JwtAuthenticator)
	authClaimsIf, err := jwtAuth.ParseAndValidate(token)
	if err != nil {
		log.Error().Msgf("Failed to parse and validate JWT token: %v", err)
		return nil, err
	}

	authClaims, isMap := authClaimsIf.(jwt.MapClaims)
	if !isMap {
		err = errors.Internal("error converting claims to a map")
		log.Error().Msgf("Failed to convert claims into a map: %v", err)
		return nil, err
	}
	for k, v := range authClaims {
		err = grpcauth.HandleClaim(&niceMd, []string{k}, v)
		if err != nil {
			log.Error().Msgf("Failed to handle claim in JWT token: %v", err)
			return niceMd, err
		}
	}

	log.Trace().Msg("JWT token is valid, proceeding to RBAC")

	return niceMd, nil
}

func SetOPAPolicies() {
	var err error
	PolicyExistFlag = true
	Policies, err = New(RbacDirectory)
	if err != nil {
		log.Error().Msgf("Can't upload RBAC realm policies to OPA package: %v", err)
		PolicyExistFlag = false
	}
}

func GetResourceRole(claims map[string]interface{}) (string, error) {
	var resourceAccessRole string
	var err error
	if claims == nil {
		return "", errors.Internal("get role failed due to claims is nil")
	}

	valResource := claims[resourceAccess]
	if valResource != nil {
		resourceAccessRole, err = getRancherResourceRole(valResource)
	}

	return resourceAccessRole, err

}

func ExtractAuthorizationFromMd(ctx context.Context) (string, error) {
	token, err := grpc_auth.AuthFromMD(ctx, contextMetadataBearerKeyLower)
	if err != nil {
		token, err = grpc_auth.AuthFromMD(ctx, contextMetadataBearerKeyCamel)
		if err != nil {
			return "", err
		}

		return token, nil
	}

	return token, nil
}

func getRancherResourceRole(resourceAccessMap interface{}) (string, error) {
	var role, rolePre string
	resources, ok := resourceAccessMap.(map[string]interface{})
	if !ok {
		return "", errors.Internal("invalid role format")
	}

	rancherOIDCKey := os.Getenv("OIDC_CLIENT_ID")

	// Check specifically for resource roles
	rancherOidcRoles, ok := resources[rancherOIDCKey]
	if !ok {
		// No resource roles found
		return "", errors.Internal("no resouce role found")
	}

	roleOidcObjs, ok := rancherOidcRoles.(map[string]interface{})
	if !ok {
		return "", errors.Internal("invalid resouce role format")
	}

	roleObjs, ok := roleOidcObjs[rolesKey].([]interface{})
	if !ok {
		return "", errors.Internal("invalid roles format in resource roles")
	}

	for _, roleVal := range roleObjs {
		var roleObj string
		roleObj, ok = roleVal.(string)
		if !ok || !slices.Contains(allRancherValidRoles, roleObj) {
			continue
		}
		switch roleObj {
		case RoleRancherAdmin:
			return RoleRancherAdmin, nil
		case RoleRancherReadWrite, RoleRancherReadOnly:
			if roleObj == RoleRancherReadOnly && rolePre == RoleRancherReadWrite {
				continue
			}
			rolePre = roleObj
		}
		role = roleObj
	}

	if len(role) == 0 {
		return "", errors.Internal("can't get resource role")
	}

	return role, nil

}

// customLog logs messages based on the global log level
func customLog(level zerolog.Level, msg string) {
	if level == zerolog.GlobalLevel() {
		log.WithLevel(level).Msg(msg)
	}
}
