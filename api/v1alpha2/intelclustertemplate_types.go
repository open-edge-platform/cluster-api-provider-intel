// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// IntelClusterTemplateSpec defines the desired state of IntelClusterTemplate
type IntelClusterTemplateSpec struct {
	Template IntelClusterTemplateResource `json:"template"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:path=intelclustertemplates,scope=Namespaced,categories=cluster-api
// +kubebuilder:metadata:labels="cluster.x-k8s.io/v1beta1=v1alpha1"

// IntelClusterTemplate is the Schema for the intelclustertemplates API
type IntelClusterTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec IntelClusterTemplateSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// IntelClusterTemplateList contains a list of IntelClusterTemplate
type IntelClusterTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IntelClusterTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IntelClusterTemplate{}, &IntelClusterTemplateList{})
}

type IntelClusterTemplateResource struct {
	// Standard object's metadata
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta clusterv1.ObjectMeta `json:"metadata,omitempty"`
	Spec       IntelClusterSpec     `json:"spec"`
}
