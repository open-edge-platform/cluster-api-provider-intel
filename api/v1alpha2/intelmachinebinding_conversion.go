// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	"fmt"

	"github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this IntelMachineBinding (v1alpha2) to the Hub version (v1alpha1).
func (src *IntelMachineBinding) ConvertTo(conv conversion.Hub) error {
	dst := conv.(*v1alpha1.IntelMachineBinding)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.NodeGUID = src.Spec.HostId
	dst.Spec.ClusterName = src.Spec.ClusterName
	dst.Spec.IntelMachineTemplateName = src.Spec.IntelMachineTemplateName

	// Status
	dst.Status.Allocated = src.Status.Allocated

	return nil
}

// ConvertTo converts the Hub version (v1alpha1) to this IntelMachineBinding version (v1alpha2).
func (dst *IntelMachineBinding) ConvertFrom(conv conversion.Hub) error {
	src := conv.(*v1alpha1.IntelMachineBinding)

	fmt.Println("Converting from v1alpha1 to v1alpha2")
	fmt.Println("Source:", src)
	fmt.Println("Destination:", dst)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.HostId = src.Spec.NodeGUID
	dst.Spec.ClusterName = src.Spec.ClusterName
	dst.Spec.IntelMachineTemplateName = src.Spec.IntelMachineTemplateName

	// Status
	dst.Status.Allocated = src.Status.Allocated

	return nil
}
