// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IntelMachineTemplateSpecTemplate struct {
	Spec IntelMachineSpec `json:"spec,omitempty"`
}

// IntelMachineTemplateSpec defines the desired state of IntelMachineTemplate.
// The Spec.Template field must be present in order to satisfy cAPI.
type IntelMachineTemplateSpec struct {
	Template IntelMachineTemplateSpecTemplate `json:"template"`
}

// IntelMachineTemplateStatus defines the observed state of IntelMachineTemplate.
type IntelMachineTemplateStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cluster.x-k8s.io/v1beta1=v1alpha1"

// IntelMachineTemplate is the Schema for the intelmachinetemplates API.
type IntelMachineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IntelMachineTemplateSpec   `json:"spec,omitempty"`
	Status IntelMachineTemplateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IntelMachineTemplateList contains a list of IntelMachineTemplate.
type IntelMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IntelMachineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IntelMachineTemplate{}, &IntelMachineTemplateList{})
}
