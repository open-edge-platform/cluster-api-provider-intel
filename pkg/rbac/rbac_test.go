// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0
package rbac

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	testing_utils "github.com/open-edge-platform/cluster-api-provider-intel/pkg/testing"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/undefinedlabs/go-mpatch"
)

const (
	writeRole  = "clusters-write-role"
	readRole   = "clusters-read-role"
	clientName = "testClient"
)

func unpatchAll(t *testing.T, pList []*mpatch.Patch) {
	for _, p := range pList {
		if p != nil {
			if err := p.Unpatch(); err != nil {
				t.Errorf("unpatch error: %v", err)
			}
		}
	}
}

func Test_getRole(t *testing.T) {

	claimMap2 := map[string]interface{}{
		"realm_access": map[string]interface{}{
			"roles": []interface{}{
				"default-roles-master",
				"offline_access",
				ClusterRoleWrite,
				"uma_authorization",
			},
		},
	}

	claimMap3 := map[string]interface{}{
		"realm_access": map[string]interface{}{
			"roles": []interface{}{
				"default-roles-master",
				"offline_access",
				ClusterRoleWrite,
				ClusterRoleRead,
				"uma_authorization",
			},
		},
	}

	claimMap6 := map[string]interface{}{
		"realm_access": map[string]interface{}{
			"roles": []interface{}{
				"default-roles-master",
				"admin",
				"uma_authorization",
			},
		},
	}

	claimMap7 := map[string]interface{}{
		"realm_access": map[string]interface{}{
			"roles": []interface{}{
				"default-roles-master",
				"admin",
				ClusterRoleRead,
				"uma_authorization",
			},
		},
	}
	claimMap10 := map[string]interface{}{
		"realm_access": map[string]interface{}{
			"roles": []interface{}{
				"default-roles-master",
				"offline_access",
				ClusterRoleWrite,
				"uma_authorization",
			},
		},
		"resource_access": map[string]interface{}{
			"ioep-rancher": map[string]interface{}{
				"roles": []interface{}{
					"default-roles-master",
					"offline_access",
					RoleRancherReadWrite,
					"uma_authorization",
				},
			},
		},
	}

	claimMap11 := map[string]interface{}{
		"realm_access": map[string]interface{}{
			"roles": []interface{}{
				"default-roles-master",
				"offline_access",
				ClusterRoleWrite,
				ClusterRoleRead,
				"uma_authorization",
			},
		},
		"resource_access": map[string]interface{}{
			"ioep-rancher": map[string]interface{}{
				"roles": []interface{}{
					"default-roles-master",
					"offline_access",
					RoleRancherReadWrite,
					RoleRancherReadOnly,
					"uma_authorization",
				},
			},
		},
	}

	claimMap14 := map[string]interface{}{
		"realm_access": map[string]interface{}{
			"roles": []interface{}{
				"default-roles-master",
				"admin",
				"uma_authorization",
			},
		},
		"resource_access": map[string]interface{}{
			"ioep-rancher": map[string]interface{}{
				"roles": []interface{}{
					"default-roles-master",
					"admin",
					"uma_authorization",
				},
			},
		},
	}

	claimMap15 := map[string]interface{}{
		"realm_access": map[string]interface{}{
			"roles": []interface{}{
				"default-roles-master",
				"admin",
				ClusterRoleRead,
				"uma_authorization",
			},
		},
		"resource_access": map[string]interface{}{
			"ioep-rancher": map[string]interface{}{
				"roles": []interface{}{
					"default-roles-master",
					"admin",
					RoleRancherReadOnly,
					"uma_authorization",
				},
			},
		},
	}

	claimMap16 := map[string]interface{}{
		"realm_access": map[string]interface{}{
			"roles": []interface{}{
				"default-roles-master",
				"admin",
				ClusterRoleRead,
				"uma_authorization",
			},
		},
		"resource_access": map[string]interface{}{
			"ioep-rancher": map[string]interface{}{
				"roles": []interface{}{
					"default-roles-master",
					"admin",
					RoleRancherAdmin,
					"uma_authorization",
				},
			},
		},
	}

	claimMapErr2 := map[string]interface{}{
		"realm_access": map[string]interface{}{
			"roles": []interface{}{
				"default-roles-master",
				"offline_access",
				"uma_authorization",
			},
		},
		"resource_access": map[string]interface{}{
			"ioep-rancher": map[string]interface{}{
				"roles": []interface{}{
					"default-roles-master",
					"offline_access",
					"uma_authorization",
				},
			},
		},
	}

	type args struct {
		claims map[string]interface{}
	}
	tests := []struct {
		name           string
		args           args
		want           string
		wantErr        bool
		funcBeforeTest func(*gomock.Controller) []*mpatch.Patch
	}{
		// TODO: Add test cases.
		{
			name:           "Get an empty role",
			args:           args{nil},
			want:           "",
			wantErr:        true,
			funcBeforeTest: nil,
		},
		{
			name:           "Get read write role",
			args:           args{claimMap2},
			want:           "",
			wantErr:        false,
			funcBeforeTest: nil,
		},
		{
			name:           "Get read write role",
			args:           args{claimMap3},
			want:           "",
			wantErr:        false,
			funcBeforeTest: nil,
		},
		{
			name:           "No valid role",
			args:           args{claimMap6},
			want:           "",
			wantErr:        false,
			funcBeforeTest: nil,
		},
		{
			name:           "Get read only role",
			args:           args{claimMap7},
			want:           "",
			wantErr:        false,
			funcBeforeTest: nil,
		},
		{
			name:           "Get Rancher read write role 1",
			args:           args{claimMap10},
			want:           RoleRancherReadWrite,
			wantErr:        false,
			funcBeforeTest: nil,
		},
		{
			name:           "Get Rancher read write role 2",
			args:           args{claimMap11},
			want:           RoleRancherReadWrite,
			wantErr:        false,
			funcBeforeTest: nil,
		},
		{
			name:           "No valid Rancher role",
			args:           args{claimMap14},
			want:           "",
			wantErr:        true,
			funcBeforeTest: nil,
		},
		{
			name:           "Get Rancher read only role 1",
			args:           args{claimMap15},
			want:           RoleRancherReadOnly,
			wantErr:        false,
			funcBeforeTest: nil,
		},
		{
			name:           "Get Rancher RoleRancherAdmin role ",
			args:           args{claimMap16},
			want:           RoleRancherAdmin,
			wantErr:        false,
			funcBeforeTest: nil,
		},
		{
			name:           "Get Rancher role fail ed",
			args:           args{claimMapErr2},
			want:           "",
			wantErr:        true,
			funcBeforeTest: nil,
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
			_ = os.Setenv("OIDC_CLIENT_ID", "ioep-rancher")
			got1, err := GetResourceRole(tt.args.claims)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got1 != tt.want {
				t.Errorf("getRole() role1 = %v, want %v", got1, tt.want)
			}
			_ = os.Unsetenv("OIDC_CLIENT_ID")
		})
	}
}

// nolint
func Test_New(t *testing.T) {
	ruleDirectory := "test/authz.rego"

	policies, err := New(ruleDirectory)
	if err != nil {
		t.Errorf("New() error = %v", err)
	}

	if policies == nil {
		t.Errorf("New() policies is nil")
	}

	if len(policies.queries) != 3 {
		t.Errorf("New() invalid number of queries")
	}

	ruleDirectory = ""

	New(ruleDirectory)
	// if err != nil {
	// 	assert.ErrorContains(t, err, errors.New("can't load admin query").Error())
	// }
}

func TestSetOPAPolicies(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
		{
			name: "set OPA policies",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetOPAPolicies()
		})
	}
}

func TestPolicy_Verify(t *testing.T) {
	type fields struct {
		queries map[string]*rego.PreparedEvalQuery
	}
	type args struct {
		claims    metautils.NiceMD
		operation string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Test case 1: verify GET",
			fields: fields{
				queries: map[string]*rego.PreparedEvalQuery{},
			},
			args: args{
				claims:    metautils.NiceMD{},
				operation: "GET",
			},
			wantErr: true,
		},
		{
			name: "Test case 2: verify PATCH",
			fields: fields{
				queries: map[string]*rego.PreparedEvalQuery{},
			},
			args: args{
				claims:    metautils.NiceMD{},
				operation: "PATCH",
			},
			wantErr: true,
		},
		{
			name: "Test case 3: verify unknow methnod",
			fields: fields{
				queries: map[string]*rego.PreparedEvalQuery{},
			},
			args: args{
				claims:    metautils.NiceMD{},
				operation: "unknow",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Policy{
				queries: tt.fields.queries,
			}
			if err := p.Verify(tt.args.claims, tt.args.operation); (err != nil) != tt.wantErr {
				t.Errorf("Policy.Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func prepareQueryForTesting(t *testing.T, regoPolicy string) *rego.PreparedEvalQuery {
	ctx := context.Background()
	r := rego.New(
		rego.Query("data.test.allow"),
		rego.Module("", regoPolicy),
	)
	query, err := r.PrepareForEval(ctx)
	if err != nil {
		t.Fatalf("Failed to prepare query for testing: %v", err)
	}
	return &query
}

func TestPolicy_evaluateQuery(t *testing.T) {
	type fields struct {
		queries map[string]*rego.PreparedEvalQuery
	}
	type args struct {
		queryKey  string
		claims    metautils.NiceMD
		operation string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Test case 1: Query does not exist",
			fields: fields{
				queries: map[string]*rego.PreparedEvalQuery{},
			},
			args: args{
				queryKey:  "nonexistent",
				claims:    metautils.NiceMD{},
				operation: "GET",
			},
			wantErr: true,
		},
		{
			name: "Test case 2: Query exists and evaluation does not return an error",
			fields: fields{
				queries: map[string]*rego.PreparedEvalQuery{
					"existing": prepareQueryForTesting(t, "package test\n\ndefault allow = true"),
				},
			},
			args: args{
				queryKey:  "existing",
				claims:    metautils.NiceMD{},
				operation: "GET",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Policy{
				queries: tt.fields.queries,
			}
			if err := p.evaluateQuery(tt.args.queryKey, tt.args.claims, tt.args.operation); (err != nil) != tt.wantErr {
				t.Errorf("Policy.evaluateQuery() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_RequestIsAuthorized(t *testing.T) {
	// p, err := loadPolicyBundle(regoPath)
	p, err := New("test/authz.rego")
	require.NoError(t, err)

	// creating a JWT with read and write roles
	_, jwtToken, err := testing_utils.CreateJWT(t)
	require.NoError(t, err)

	niceMD := metautils.NiceMD{}
	niceMD.Add(authKey, "bearer "+jwtToken)
	ctx := niceMD.ToIncoming(context.Background())

	res := p.RequestIsAuthorized(ctx, MethodPost)
	assert.True(t, res)

	res = p.RequestIsAuthorized(ctx, MethodGet)
	assert.True(t, res)

	res = p.RequestIsAuthorized(ctx, "UnknownOp")
	assert.False(t, res)

	// creating a JWT with read only role
	_, jwtToken, err = testing_utils.CreateJWTWithReadRole(t)
	require.NoError(t, err)

	niceMD1 := metautils.NiceMD{}
	niceMD1.Add(authKey, "bearer "+jwtToken)
	ctx1 := niceMD1.ToIncoming(context.Background())

	res1 := p.RequestIsAuthorized(ctx1, MethodPost)
	assert.False(t, res1)

	res1 = p.RequestIsAuthorized(ctx1, MethodGet)
	assert.True(t, res1)

	res1 = p.RequestIsAuthorized(ctx1, "UnknownOp")
	assert.False(t, res1)

	// creating a JWT with write only role
	_, jwtToken, err = testing_utils.CreateJWTWithWriteRole(t)
	require.NoError(t, err)

	niceMD2 := metautils.NiceMD{}
	niceMD2.Add(authKey, "bearer "+jwtToken)
	ctx2 := niceMD2.ToIncoming(context.Background())

	res2 := p.RequestIsAuthorized(ctx2, MethodPost)
	assert.True(t, res2)

	res2 = p.RequestIsAuthorized(ctx2, MethodGet)
	assert.False(t, res2)

	res2 = p.RequestIsAuthorized(ctx2, "UnknownOp")
	assert.False(t, res2)
}

func Test_ClientCanBypassAuthN(t *testing.T) {
	p, err := New("test/authz.rego")
	require.NoError(t, err)

	t.Setenv(allowMissingAuthClients, clientName)

	niceMD := metautils.NiceMD{}
	ctx := niceMD.ToIncoming(context.Background())

	res := p.RequestIsAuthorized(ctx, MethodPost)
	assert.False(t, res)

	niceMD.Add(clientKeyLower, clientName)
	ctx = niceMD.ToIncoming(context.Background())

	res = p.RequestIsAuthorized(ctx, MethodPost)
	assert.True(t, res)
}

func Test_AddJWTToTheOutgoingContext(t *testing.T) {
	// creating a JWT with read and write roles
	_, jwtToken, err := testing_utils.CreateJWT(t)
	require.NoError(t, err)

	ctx := context.Background()
	niceMD := metautils.NiceMD{}
	niceMD.Add(authKey, "Bearer "+jwtToken)
	ctx = niceMD.ToOutgoing(ctx)

	niceMD = metautils.ExtractOutgoing(ctx)
	retAuth := niceMD.Get(authKey)
	retAuthTokens := strings.Split(retAuth, " ")
	require.Equal(t, 2, len(retAuthTokens))
	assert.Equal(t, jwtToken, retAuthTokens[1])
}

func Test_AddJWTToTheIncomingContext(t *testing.T) {
	// creating a JWT with read and write roles
	_, jwtToken, err := testing_utils.CreateJWT(t)
	require.NoError(t, err)

	ctx := context.Background()

	niceMD := metautils.NiceMD{}
	niceMD.Add(authKey, "Bearer "+jwtToken)

	niceMD = metautils.ExtractIncoming(niceMD.ToIncoming(ctx))
	retAuth := niceMD.Get(authKey)
	retAuthTokens := strings.Split(retAuth, " ")
	require.Equal(t, 2, len(retAuthTokens))
	assert.Equal(t, jwtToken, retAuthTokens[1])
}
