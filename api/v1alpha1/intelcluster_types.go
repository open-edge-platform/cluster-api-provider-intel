// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

const (
	ClusterFinalizer = "intelcluster.infrastructure.cluster.x-k8s.io"
)

// IntelClusterSpec defines the desired state of IntelCluster
type IntelClusterSpec struct {
	// controlPlaneEndpoint represents the endpoint used to communicate with the control plane
	// +optional
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint"`
	// providerId represents the id the inventory manager assigns to the cluster at creation time
	// +optional
	ProviderId string `json:"providerId"`
}

// IntelClusterStatus defines the observed state of IntelCluster
type IntelClusterStatus struct {
	// ready denotes that the Intel cluster infrastructure is fully provisioned
	// NOTE: this field is part of the Cluster API contract and it is used to orchestrate provisioning.
	// The value of this field is never updated after provisioning is completed. Please use conditions
	// to check the operational state of the infa cluster.
	// +optional
	Ready bool `json:"ready"`

	// conditions represents the observations of an IntelCluster's current state.
	// Known condition types are Ready, Provisioned, Deleting, Paused.
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=32
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=intelclusters,scope=Namespaced,categories=cluster-api
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels['cluster\\.x-k8s\\.io/cluster-name']",description="Cluster"
// +kubebuilder:printcolumn:name="ProviderId",type="string",JSONPath=".spec.providerId",description="ProviderId associated with the cluster"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="IntelCluster is ready for IntelMachine"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of IntelCluster"
// +kubebuilder:metadata:labels="cluster.x-k8s.io/v1beta2=v1alpha1"

// IntelCluster is the Schema for the intelclusters API
type IntelCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IntelClusterSpec   `json:"spec,omitempty"`
	Status IntelClusterStatus `json:"status,omitempty"`
}

// GetConditions returns the observations of the operational state of the IntelCluster resource.
// This implements the v1beta2 Getter interface.
func (c *IntelCluster) GetConditions() []metav1.Condition {
	return c.Status.Conditions
}

// SetConditions sets the underlying service state of the IntelCluster to the predescribed conditions.
// This implements the v1beta2 Setter interface.
func (c *IntelCluster) SetConditions(conditions []metav1.Condition) {
	c.Status.Conditions = conditions
}

// +kubebuilder:object:root=true

// IntelClusterList contains a list of IntelCluster
type IntelClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IntelCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IntelCluster{}, &IntelClusterList{})
}
