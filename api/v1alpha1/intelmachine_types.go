// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// FreeInstanceFinalizer allows ReconcileIntelMachine to remove the instance from the Workload in the Inventory
	FreeInstanceFinalizer = "intelmachine.infrastructure.cluster.x-k8s.io/free-instance"
	// HostCleanupFinalizer allows ReconcileIntelMachine to trigger a cleanup on the host
	HostCleanupFinalizer = "intelmachine.infrastructure.cluster.x-k8s.io/host-cleanup"
	// DeauthFinalizer was used for deauthorizing the host from the inventory before deletion in 3.1 release.
	// It is kept here to unblock deletion of clusters created with that version.
	DeauthFinalizer = "intelmachine.infrastructure.cluster.x-k8s.io/deauth-host"

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
	// ready denotes that the Intel machine infrastructure is fully provisioned.
	// NOTE: this field is part of the Cluster API contract and it is used to orchestrate provisioning.
	// The value of this field is never updated after provisioning is completed. Please use conditions
	// to check the operational state of the infra machine.
	// +optional
	Ready bool `json:"ready,omitempty"`

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
// +kubebuilder:metadata:labels="cluster.x-k8s.io/v1beta2=v1alpha1"

// IntelMachine is the Schema for the intelmachines API.
type IntelMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IntelMachineSpec   `json:"spec,omitempty"`
	Status IntelMachineStatus `json:"status,omitempty"`
}

// GetConditions returns the set of conditions for this object.
// This implements the v1beta2 Getter interface.
func (c *IntelMachine) GetConditions() []metav1.Condition {
	return c.Status.Conditions
}

// SetConditions sets the conditions on this object.
// This implements the v1beta2 Setter interface.
func (c *IntelMachine) SetConditions(conditions []metav1.Condition) {
	c.Status.Conditions = conditions
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
