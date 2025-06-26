// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	// FreeInstanceFinalizer allows ReconcileIntelMachine to remove the instance from the Workload in the Inventory
	FreeInstanceFinalizer = "intelmachine.infrastructure.cluster.x-k8s.io/free-instance"
	// HostCleanupFinalizer - no longer used, but kept for backward compatibility
	HostCleanupFinalizer = "intelmachine.infrastructure.cluster.x-k8s.io/host-cleanup"
	// DeauthHostFinalizer allows ReconcileIntelMachine to deauthorize the host in the Inventory
	DeauthHostFinalizer = "intelmachine.infrastructure.cluster.x-k8s.io/deauth-host"

	// HostState is used by the SB Handler to report the cluster status posted by the agent as an annotation.
	HostStateAnnotation = "intelmachine.infrastructure.cluster.x-k8s.io/agent-status"
	HostStateActive     = "active"
	HostStateInactive   = "inactive"
	HostStateError      = "error"
	HostStateInProgress = "in-progress"

	HostIdAnnotation = "intelmachine.infrastructure.cluster.x-k8s.io/host-id"

	// NodeGUID label key
	NodeGUIDKey = "NodeGUID"
)

// IntelMachineSpec defines the desired state of IntelMachine.
type IntelMachineSpec struct {
	// ProviderID must match the provider ID as seen on the node object corresponding to this machine.
	// +optional
	ProviderID *string `json:"providerID,omitempty"`

	// NodeGUID contains the GUID of the node.
	// +optional
	NodeGUID string `json:"nodeGUID,omitempty"`
}

// IntelMachineStatus defines the observed state of IntelMachine.
type IntelMachineStatus struct {
	// Conditions is a list of conditions that describe the state of the IntelMachine.
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`

	// v1beta2 groups all the fields that will be added or modified in IntelMachine's status with the V1Beta2 version.
	// +optional
	V1Beta2 *IntelMachineV1Beta2Status `json:"v1beta2,omitempty"`

	// ready denotes that the Intel machine infrastructure is fully provisioned.
	// NOTE: this field is part of the Cluster API contract and it is used to orchestrate provisioning.
	// The value of this field is never updated after provisioning is completed. Please use conditions
	// to check the operational state of the infra machine.
	// +optional
	Ready bool `json:"ready,omitempty"`
}

// IntelMachineV1Beta2Status groups all the fields that will be added or modified in IntelMachine with the V1Beta2 version.
// See https://github.com/kubernetes-sigs/cluster-api/blob/main/docs/proposals/20240916-improve-status-in-CAPI-resources.md for more context.
type IntelMachineV1Beta2Status struct {
	// conditions represents the observations of an IntelMachine's current state.
	// Known condition types are Ready, Provisioned, BootstrapExecSucceeded, Deleting, Paused.
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=32
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Provider ID",type=string,JSONPath=`.spec.providerID`
// +kubebuilder:printcolumn:name="Node GUID",type=string,JSONPath=`.spec.nodeGUID`
// +kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:metadata:labels="cluster.x-k8s.io/v1beta1=v1alpha1"

// IntelMachine is the Schema for the intelmachines API.
type IntelMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IntelMachineSpec   `json:"spec,omitempty"`
	Status IntelMachineStatus `json:"status,omitempty"`
}

// GetConditions returns the set of conditions for this object.
func (c *IntelMachine) GetConditions() clusterv1.Conditions {
	return c.Status.Conditions
}

// SetConditions sets the conditions on this object.
func (c *IntelMachine) SetConditions(conditions clusterv1.Conditions) {
	c.Status.Conditions = conditions
}

// GetV1Beta2Conditions returns the set of conditions for this object.
func (c *IntelMachine) GetV1Beta2Conditions() []metav1.Condition {
	if c.Status.V1Beta2 == nil {
		return nil
	}
	return c.Status.V1Beta2.Conditions
}

// SetV1Beta2Conditions sets conditions for an API object.
func (c *IntelMachine) SetV1Beta2Conditions(conditions []metav1.Condition) {
	if c.Status.V1Beta2 == nil {
		c.Status.V1Beta2 = &IntelMachineV1Beta2Status{}
	}
	c.Status.V1Beta2.Conditions = conditions
}

// +kubebuilder:object:root=true

// IntelMachineList contains a list of IntelMachine.
type IntelMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IntelMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IntelMachine{}, &IntelMachineList{})
}
