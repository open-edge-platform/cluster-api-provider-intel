// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
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

// IntelClusterV1Beta2Status groups all the fields that will be added or modified in IntelCluster with the V1Beta2 version.
// See https://github.com/kubernetes-sigs/cluster-api/blob/main/docs/proposals/20240916-improve-status-in-CAPI-resources.md for more context.
type IntelClusterV1Beta2Status struct {
	// conditions represents the observations of an IntelCluster's current state.
	// Known condition types are Ready, Provisioned, BootstrapExecSucceeded, Deleting, Paused.
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=32
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// IntelClusterStatus defines the observed state of IntelCluster
type IntelClusterStatus struct {
	// ready denotes that the Intel cluster infrastructure is fully provisioned
	// NOTE: this field is part of the Cluster API contract and it is used to orchestrate provisioning.
	// The value of this field is never updated after provisioning is completed. Please use conditions
	// to check the operational state of the infa cluster.
	// +optional
	Ready      bool                 `json:"ready"`
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`
	// v1beta2 groups all the fields that will be added or modified in IntelCluster's status with the V1Beta2 version.
	// +optional
	V1Beta2 *IntelClusterV1Beta2Status `json:"v1beta2,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=intelclusters,scope=Namespaced,categories=cluster-api
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels['cluster\\.x-k8s\\.io/cluster-name']",description="Cluster"
// +kubebuilder:printcolumn:name="ProviderId",type="string",JSONPath=".spec.providerId",description="ProviderId associated with the cluster"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="IntelCluster is ready for IntelMachine"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of IntelCluster"
// +kubebuilder:metadata:labels="cluster.x-k8s.io/v1beta1=v1alpha1"

// IntelCluster is the Schema for the intelclusters API
type IntelCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IntelClusterSpec   `json:"spec,omitempty"`
	Status IntelClusterStatus `json:"status,omitempty"`
}

// GetConditions returns the observations of the operational state of the IntelCluster resource.
func (r *IntelCluster) GetConditions() clusterv1.Conditions {
	return r.Status.Conditions
}

// SetConditions sets the underlying service state of the IntelCluster to the predescribed clusterv1.Conditions.
func (r *IntelCluster) SetConditions(conditions clusterv1.Conditions) {
	r.Status.Conditions = conditions
}

// GetV1Beta2Conditions returns the set of conditions for this object.
func (c *IntelCluster) GetV1Beta2Conditions() []metav1.Condition {
	if c.Status.V1Beta2 == nil {
		return nil
	}
	return c.Status.V1Beta2.Conditions
}

// SetV1Beta2Conditions sets conditions for an API object.
func (c *IntelCluster) SetV1Beta2Conditions(conditions []metav1.Condition) {
	if c.Status.V1Beta2 == nil {
		c.Status.V1Beta2 = &IntelClusterV1Beta2Status{}
	}
	c.Status.V1Beta2.Conditions = conditions
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
