// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package scope

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"

	infrav1alpha2 "github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha2"
	"github.com/open-edge-platform/cluster-api-provider-intel/mocks/m_client"
	"github.com/open-edge-platform/cluster-api-provider-intel/test/utils"
)

const (
	namespaceName = "275ecb36-5aa8-4c2a-9c47-000000000000"
	clusterName   = "test-cluster"
)

var (
	testLogger       = ctrl.LoggerFrom(context.TODO())
	testCluster      = utils.NewCluster(namespaceName, clusterName)
	testIntelCluster = utils.NewIntelClusterNoSpec(testCluster)

	ErrInvalidPatch = errors.New("failed to patch object")
)

func TestClusterScopeWithValidBuild(t *testing.T) {
	context := context.TODO()
	fakeClient := &m_client.MockClient{}
	testScheme := runtime.NewScheme()
	err := infrav1alpha2.AddToScheme(testScheme)
	require.Nil(t, err)
	fakeClient.On("Scheme").Return(testScheme).Times(2)

	successfulClusterScope, _ := NewClusterReconcileScopeBuilder().
		WithContext(context).
		WithLog(&testLogger).
		WithClient(fakeClient).
		WithCluster(testCluster).
		WithIntelCluster(testIntelCluster).
		Build()

	cases := []struct {
		name        string
		builder     ClusterReconcileScopeBuilder
		scopeOutput *ClusterReconcileScope
	}{
		{
			name: "valid scope",
			builder: NewClusterReconcileScopeBuilder().
				WithContext(context).
				WithLog(&testLogger).
				WithClient(fakeClient).
				WithCluster(testCluster).
				WithIntelCluster(testIntelCluster),
			scopeOutput: successfulClusterScope,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scope, err := tc.builder.Build()
			assert.Equal(t, tc.scopeOutput, scope)
			assert.Nil(t, err)
		})
	}
}

func TestClusterScopeBuildWithValidationErrors(t *testing.T) {
	context := context.TODO()
	fakeClient := &m_client.MockClient{}
	fakeClient.On("Scheme").Return(scheme.Scheme)

	cases := []struct {
		name              string
		builder           ClusterReconcileScopeBuilder
		builderErr        error
		returnsWrappedErr bool
	}{
		{
			name:       "empty context",
			builder:    NewClusterReconcileScopeBuilder(),
			builderErr: ErrInvalidScopeWithContext,
		},
		{
			name: "empty log",
			builder: NewClusterReconcileScopeBuilder().
				WithContext(context),
			builderErr: ErrInvalidScopeWithLogger,
		},
		{
			name: "empty client",
			builder: NewClusterReconcileScopeBuilder().
				WithContext(context).
				WithLog(&testLogger).
				WithClient(nil),
			builderErr: ErrInvalidScopeWithClient,
		}, {
			name: "empty cluster",
			builder: NewClusterReconcileScopeBuilder().
				WithContext(context).
				WithLog(&testLogger).
				WithClient(fakeClient).
				WithCluster(nil),
			builderErr: ErrInvalidScopeWithCluster,
		}, {
			name: "empty intelcluster",
			builder: NewClusterReconcileScopeBuilder().
				WithContext(context).
				WithLog(&testLogger).
				WithClient(fakeClient).
				WithCluster(testCluster),
			builderErr: ErrInvalidScopeWithIntelCluster,
		}, {
			name: "empty patch helper",
			builder: NewClusterReconcileScopeBuilder().
				WithContext(context).
				WithLog(&testLogger).
				WithClient(fakeClient).
				WithCluster(testCluster).
				WithIntelCluster(testIntelCluster),
			builderErr:        errors.New("failed to create patch helper"),
			returnsWrappedErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scope, err := tc.builder.Build()
			assert.Nil(t, scope)
			assert.NotNil(t, err)
			if !tc.returnsWrappedErr {
				assert.Equal(t, tc.builderErr, err)
			} else {
				assert.Contains(t, err.Error(), tc.builderErr.Error())
			}
		})
	}
}

func TestClusterScopeClose(t *testing.T) {
	context := context.TODO()
	fakeClient := &m_client.MockClient{}
	testScheme := runtime.NewScheme()
	err := infrav1alpha2.AddToScheme(testScheme)
	require.Nil(t, err)
	defer func() {
		fakeClient.AssertExpectations(t)
	}()

	fakeClient.On("Scheme").Return(testScheme)

	scope, err := NewClusterReconcileScopeBuilder().
		WithContext(context).
		WithLog(&testLogger).
		WithClient(fakeClient).
		WithCluster(testCluster).
		WithIntelCluster(testIntelCluster).
		Build()
	require.NotNil(t, scope)
	require.Nil(t, err)
	err = scope.Close()
	assert.Nil(t, err)

}

func TestClusterScopeCloseWithError(t *testing.T) {
	context := context.TODO()
	fakeClient := &m_client.MockClient{}
	testScheme := runtime.NewScheme()
	err := infrav1alpha2.AddToScheme(testScheme)
	require.Nil(t, err)
	subRW := &m_client.MockSubResourceWriter{}
	defer func() {
		fakeClient.AssertExpectations(t)
		subRW.AssertExpectations(t)
	}()
	subRW.On("Patch", context, mock.Anything, mock.Anything).Return(ErrInvalidPatch).Times(2)

	fakeClient.On("Scheme").Return(testScheme)
	fakeClient.On("Get", context, types.NamespacedName{Namespace: testIntelCluster.Namespace, Name: testIntelCluster.Name},
		mock.Anything).Return(nil)
	fakeClient.On("Status").Return(subRW)

	conditions.MarkTrue(testIntelCluster, infrav1alpha2.WorkloadCreatedReadyCondition)
	scope, err := NewClusterReconcileScopeBuilder().
		WithContext(context).
		WithLog(&testLogger).
		WithClient(fakeClient).
		WithCluster(testCluster).
		WithIntelCluster(testIntelCluster).
		Build()
	require.NotNil(t, scope)
	require.Nil(t, err)
	err = scope.Close()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), ErrInvalidPatch.Error())
}
