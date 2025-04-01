// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IntelMachineBindingSpec defines the desired state of IntelMachineBinding.
type IntelMachineBindingSpec struct {
	// NodeGUID contains the GUID of the node
	NodeGUID string `json:"nodeGUID"`

	// ClusterName contains the name of the cluster to which the node is bound
	ClusterName string `json:"clusterName"`

	// IntelMachineTemplateName contains the name of the IntelMachineTemplate for the node
	IntelMachineTemplateName string `json:"intelMachineTemplateName"`
}

// IntelMachineBindingStatus defines the observed state of IntelMachineBinding.
type IntelMachineBindingStatus struct {
	// Allocated denotes that the node has been allocated to the cluster
	Allocated bool `json:"allocated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Cluster Name",type=string,JSONPath=`.spec.clusterName`
// +kubebuilder:printcolumn:name="Node GUID",type=string,JSONPath=`.spec.nodeGUID`
// +kubebuilder:printcolumn:name="Template Name",type=string,JSONPath=`.spec.intelMachineTemplateName`
// +kubebuilder:printcolumn:name="Allocated",type=boolean,JSONPath=`.status.allocated`

// IntelMachineBinding is the Schema for the intelmachinebindings API.
type IntelMachineBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IntelMachineBindingSpec   `json:"spec,omitempty"`
	Status IntelMachineBindingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IntelMachineBindingList contains a list of IntelMachineBinding.
type IntelMachineBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IntelMachineBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IntelMachineBinding{}, &IntelMachineBindingList{})
}
