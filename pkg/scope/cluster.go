// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package scope

import (
	"context"
	"errors"

	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	infrav1 "github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

var (
	ErrInvalidScopeWithCluster      = errors.New("failed to generate scope with empty cluster")
	ErrInvalidScopeWithLogger       = errors.New("failed to generate scope with empty logger")
	ErrInvalidScopeWithContext      = errors.New("failed to generate scope with nil context")
	ErrInvalidScopeWithIntelCluster = errors.New("failed to generate scope with empty intelcluster")
	ErrInvalidScopeWithClient       = errors.New("failed to generate scope with empty client")
)

type ClusterScope interface {
	Close() error
}
type ClusterReconcileScope struct {
	patchHelper  *patch.Helper
	client       client.Client
	Ctx          context.Context
	Log          *logr.Logger
	Cluster      *clusterv1.Cluster
	IntelCluster *infrav1.IntelCluster
}

type clusterReconcileScopeBuilder struct {
	clusterReconcileScope *ClusterReconcileScope
}

type ClusterReconcileScopeBuilder interface {
	WithContext(ctx context.Context) ClusterReconcileScopeBuilder
	WithLog(log *logr.Logger) ClusterReconcileScopeBuilder
	WithClient(client client.Client) ClusterReconcileScopeBuilder
	WithCluster(cluster *clusterv1.Cluster) ClusterReconcileScopeBuilder
	WithIntelCluster(intelCluster *infrav1.IntelCluster) ClusterReconcileScopeBuilder
	Build() (*ClusterReconcileScope, error)
}

func NewClusterReconcileScopeBuilder() ClusterReconcileScopeBuilder {
	return &clusterReconcileScopeBuilder{
		clusterReconcileScope: &ClusterReconcileScope{},
	}
}

func (b *clusterReconcileScopeBuilder) WithContext(ctx context.Context) ClusterReconcileScopeBuilder {
	b.clusterReconcileScope.Ctx = ctx
	return b
}

func (b *clusterReconcileScopeBuilder) WithLog(log *logr.Logger) ClusterReconcileScopeBuilder {
	b.clusterReconcileScope.Log = log
	return b
}

func (b *clusterReconcileScopeBuilder) WithClient(client client.Client) ClusterReconcileScopeBuilder {
	b.clusterReconcileScope.client = client
	return b
}

func (b *clusterReconcileScopeBuilder) WithCluster(cluster *clusterv1.Cluster) ClusterReconcileScopeBuilder {
	b.clusterReconcileScope.Cluster = cluster
	return b
}

func (b *clusterReconcileScopeBuilder) WithIntelCluster(
	intelCluster *infrav1.IntelCluster) ClusterReconcileScopeBuilder {
	b.clusterReconcileScope.IntelCluster = intelCluster
	return b
}

func (b *clusterReconcileScopeBuilder) Build() (*ClusterReconcileScope, error) {
	if b.clusterReconcileScope.Ctx == nil {
		return nil, ErrInvalidScopeWithContext
	}

	if b.clusterReconcileScope.Log == nil {
		return nil, ErrInvalidScopeWithLogger
	}

	if b.clusterReconcileScope.client == nil {
		return nil, ErrInvalidScopeWithClient
	}

	if b.clusterReconcileScope.Cluster == nil {
		return nil, ErrInvalidScopeWithCluster
	}

	if b.clusterReconcileScope.IntelCluster == nil {
		return nil, ErrInvalidScopeWithIntelCluster
	}

	patchHelper, err := patch.NewHelper(b.clusterReconcileScope.IntelCluster, b.clusterReconcileScope.client)
	if err != nil {
		b.clusterReconcileScope.Log.Error(err, "failed to create cluster patch helper")
		return nil, err
	}
	b.clusterReconcileScope.patchHelper = patchHelper

	return b.clusterReconcileScope, nil
}

func (s *ClusterReconcileScope) Close() error {
	// always update the readyCondition by summarizing the state of other conditions
	// a step counter is added to represent progress during the provisioning process
	// (instead we are hiding it during the deletion process)
	conditions.SetSummary(s.IntelCluster,
		conditions.WithConditions(
			infrav1.ControlPlaneEndpointReadyCondition,
			infrav1.WorkloadCreatedReadyCondition,
			infrav1.SecureTunnelEstablishedCondition,
		),
		conditions.WithStepCounterIf(s.IntelCluster.ObjectMeta.DeletionTimestamp.IsZero()),
	)

	// patch the object, ignoring conflicts on the conditions owned by this controller
	return s.patchHelper.Patch(
		// it seems like for now we're not hitting any timeouts; once we do
		// we could use an empty context for this call
		s.Ctx,
		s.IntelCluster,
		patch.WithOwnedConditions{Conditions: []clusterv1.ConditionType{
			clusterv1.ReadyCondition,
			infrav1.ControlPlaneEndpointReadyCondition,
			infrav1.WorkloadCreatedReadyCondition,
			infrav1.SecureTunnelEstablishedCondition,
		}},
	)
}
