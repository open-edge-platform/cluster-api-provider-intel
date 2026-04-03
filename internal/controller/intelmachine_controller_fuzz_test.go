// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrastructurev1alpha1 "github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha1"
	"github.com/open-edge-platform/cluster-api-provider-intel/mocks/m_inventory"
	inventory "github.com/open-edge-platform/cluster-api-provider-intel/pkg/inventory"
	utils "github.com/open-edge-platform/cluster-api-provider-intel/test/utils"
)

func FuzzMachineReconcile(f *testing.F) {
	const (
		namespace               = "default"
		clusterName             = "test-cluster"
		intelClusterName        = "test-intelcluster"
		machineName             = "test-machine"
		intelMachineName        = "test-intelmachine"
		intelMachineBindingName = "test-intelmachinebinding"
		machineTemplateName     = "test-machinetemplate"
		workloadId              = "test-workload-id"
		instanceId              = "test-instance-id"
		bootstrapKind           = "RKE2Config"
		nodeGUID                = "1234567890"
	)
	scheme := runtime.NewScheme()
	if err := infrastructurev1alpha1.AddToScheme(scheme); err != nil {
		f.Fatalf("infrastructurev1alpha1.AddToScheme: %v", err)
	}
	if err := clusterv1.AddToScheme(scheme); err != nil {
		f.Fatalf("clusterv1.AddToScheme: %v", err)
	}
	cluster := utils.NewCluster(namespace, clusterName)
	conditions.Set(cluster, metav1.Condition{
		Type:   string(clusterv1.InfrastructureReadyCondition),
		Status: metav1.ConditionTrue,
		Reason: "Ready",
	})
	intelcluster := utils.NewIntelCluster(namespace, intelClusterName, workloadId, cluster)
	machine := utils.NewMachine(namespace, clusterName, machineName, bootstrapKind)
	intelmachinebinding := utils.NewIntelMachineBinding(namespace, intelMachineBindingName, nodeGUID, clusterName, machineTemplateName)
	intelmachine := utils.NewIntelMachine(namespace, intelMachineName, machine)
	intelmachine.Annotations = map[string]string{
		clusterv1.TemplateClonedFromNameAnnotation: machineTemplateName,
	}
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(cluster, intelcluster, machine, intelmachine, intelmachinebinding).
		WithStatusSubresource(intelmachine, intelmachinebinding).
		WithIndex(
			&infrastructurev1alpha1.IntelMachineBinding{},
			intelMachineBindingKey,
			intelMachineBindingIdxFunc).
		Build()
	inventoryClient := &m_inventory.MockInfrastructureProvider{}
	reconciler := &IntelMachineReconciler{
		Client:          fakeClient,
		Scheme:          scheme,
		InventoryClient: inventoryClient,
	}
	cluster.Spec.InfrastructureRef = clusterv1.ContractVersionedObjectReference{
		Kind:     "IntelCluster",
		Name:     intelcluster.Name,
		APIGroup: infrastructurev1alpha1.GroupVersion.Group,
	}
	if err := fakeClient.Update(ctx, cluster); err != nil {
		f.Fatalf("fakeClient.Update: %v", err)
	}
	machine.Spec.InfrastructureRef = clusterv1.ContractVersionedObjectReference{
		Kind:     "IntelMachine",
		Name:     intelmachine.Name,
		APIGroup: infrastructurev1alpha1.GroupVersion.Group,
	}
	if err := fakeClient.Update(ctx, machine); err != nil {
		f.Fatalf("fakeClient.Update: %v", err)
	}
	instance := &inventory.Instance{
		Id: instanceId,
	}
	inventoryClient.On("GetInstanceByMachineId", mock.Anything).
		Return(inventory.GetInstanceByMachineIdOutput{Instance: instance, Err: nil}).
		Maybe()
	inventoryClient.On("AddInstanceToWorkload", mock.Anything).
		Return(inventory.AddInstanceToWorkloadOutput{Err: nil}).
		Maybe()
	inventoryClient.On("DeleteInstanceFromWorkload", mock.Anything).
		Return(inventory.DeleteInstanceFromWorkloadOutput{Err: nil}).
		Maybe()
	inventoryClient.On("DeleteWorkload", mock.Anything).
		Return(inventory.DeleteWorkloadOutput{Err: nil}).
		Maybe()

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
