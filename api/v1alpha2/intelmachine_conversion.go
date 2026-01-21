// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	"fmt"

	"github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this IntelMachine (v1alpha2) to the Hub version (v1alpha1).
func (src *IntelMachine) ConvertTo(conv conversion.Hub) error {
	dst := conv.(*v1alpha1.IntelMachine)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	if src.Spec.ProviderID != nil {
		providerID := *src.Spec.ProviderID
		dst.Spec.ProviderID = &providerID
	}

	dst.Spec.NodeGUID = src.Spec.HostId

	// Status
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.Ready = src.Status.Ready

	if src.Status.V1Beta2 != nil {
		dst.Status.V1Beta2 = &v1alpha1.IntelMachineV1Beta2Status{
			Conditions: src.Status.V1Beta2.Conditions,
		}
	}

	return nil
}

// ConvertTo converts the Hub version (v1alpha1) to this IntelMachine version (v1alpha2).
func (dst *IntelMachine) ConvertFrom(conv conversion.Hub) error {
	src := conv.(*v1alpha1.IntelMachine)

	fmt.Println("Converting from v1alpha1 to v1alpha2")
	fmt.Println("Source:", src)
	fmt.Println("Destination:", dst)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.ProviderID = src.Spec.ProviderID
	dst.Spec.HostId = src.Spec.NodeGUID

	// Status
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.Ready = src.Status.Ready

	if src.Status.V1Beta2 != nil {
		dst.Status.V1Beta2 = &IntelMachineV1Beta2Status{
			Conditions: src.Status.V1Beta2.Conditions,
		}
	}

	return nil
}
