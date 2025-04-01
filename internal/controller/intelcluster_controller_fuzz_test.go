// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrastructurev1alpha1 "github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha1"
	utils "github.com/open-edge-platform/cluster-api-provider-intel/test/utils"
	ccgv1 "github.com/open-edge-platform/cluster-connect-gateway/api/v1alpha1"
)

func FuzzClusterReconcile(f *testing.F) {
	const (
		namespace        = "default"
		clusterName      = "test-cluster"
		intelClusterName = "test-intelcluster"
		providerID       = "provider-id"
	)
	scheme := runtime.NewScheme()
	if err := infrastructurev1alpha1.AddToScheme(scheme); err != nil {
		f.Fatalf("infrastructurev1alpha1.AddToScheme: %v", err)
	}
	if err := clusterv1.AddToScheme(scheme); err != nil {
		f.Fatalf("clusterv1.AddToScheme: %v", err)
	}
	if err := ccgv1.AddToScheme(scheme); err != nil {
		f.Fatalf("ccgv1.AddToScheme: %v", err)
	}
	cluster := utils.NewCluster(namespace, clusterName)
	intelcluster := utils.NewIntelCluster(namespace, intelClusterName, providerID, cluster)
	cluster.Spec.InfrastructureRef = utils.GetObjectRef(&intelcluster.ObjectMeta, "IntelCluster")
	cluster.Spec.ControlPlaneRef = &corev1.ObjectReference{Name: cluster.Name + "-controlplane"}
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(cluster, intelcluster).
		Build()
	reconciler := &IntelClusterReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}
	f.Add("abc", "def")
	f.Fuzz(func(t *testing.T, name, ns string) {
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      name,
				Namespace: ns,
			},
		}
		_, _ = reconciler.Reconcile(context.Background(), req)
	})
}
