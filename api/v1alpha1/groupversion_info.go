// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// Package v1alpha1 contains API Schema definitions for the infrastructure v1alpha1 API group.
// +kubebuilder:object:generate=true
// +groupName=infrastructure.cluster.x-k8s.io
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// GroupVersion is group version used to register these objects.
	GroupVersion = schema.GroupVersion{Group: "infrastructure.cluster.x-k8s.io", Version: "v1alpha1"}

	// SchemeBuilder collects functions to add types to a scheme.
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// addKnownTypes registers the types in this package with the given scheme.
func addKnownTypes(s *runtime.Scheme) error {
	s.AddKnownTypes(GroupVersion,
		&IntelCluster{}, &IntelClusterList{},
		&IntelMachine{}, &IntelMachineList{},
		&IntelMachineBinding{}, &IntelMachineBindingList{},
		&IntelMachineTemplate{}, &IntelMachineTemplateList{},
		&IntelClusterTemplate{}, &IntelClusterTemplateList{},
	)
	metav1.AddToGroupVersion(s, GroupVersion)
	return nil
}
