// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	"context"
	"fmt"

	"github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	validationutils "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:docs-gen:collapse=Imports

var imlog = logf.Log.WithName("intelmachine-resource")

// SetupIntelMachineWebhookWithManager registers the webhook for IntelMachine in the manager.
func SetupIntelMachineWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&v1alpha2.IntelMachine{}).
		WithValidator(&IntelMachineCustomValidator{}).
		WithDefaulter(&IntelMachineCustomDefaulter{
			DefaultProviderId: "default-provider-id", // TODO? do we want default values of fail if empty? (valdiator)
			DefaultHostId:     "default-node-guid",
		}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-intelmachine-v1alpha2,mutating=true,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=intelmachines,verbs=create;update,versions=v1alpha2,name=mutate-intelmachine-v1alpha2.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

// IntelMachineCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind IntelMachine when those are created or updated.
type IntelMachineCustomDefaulter struct {
	// Default values for various IntelMachine fields
	DefaultProviderId string
	DefaultHostId     string
}

var _ webhook.CustomDefaulter = &IntelMachineCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind IntelMachine.
func (d *IntelMachineCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	machine, ok := obj.(*v1alpha2.IntelMachine)
	if !ok {
		return fmt.Errorf("expected an IntelMachine object but got %T", obj)
	}

	imlog.Info("applying default values for IntelMachine", "name", machine.GetName())
	d.applyDefaults(machine)

	return nil
}

// applyDefaults applies default values to IntelMachine fields.
func (d *IntelMachineCustomDefaulter) applyDefaults(machine *v1alpha2.IntelMachine) {
	if *machine.Spec.ProviderID == "" {
		machine.Spec.ProviderID = &d.DefaultProviderId
	}
	if machine.Spec.HostId == "" {
		machine.Spec.HostId = d.DefaultHostId
	}
}

// +kubebuilder:webhook:path=/validate-intelmachine-v1alpha2,mutating=false,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=intelmachines,verbs=create;update,versions=v1alpha2,name=validate-intelmachine-v1alpha2.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

// IntelMachineCustomValidator struct is responsible for validating the IntelMachine resource
// when it is created, updated, or deleted.
type IntelMachineCustomValidator struct{} // +kubebuilder:docs-gen:collapse=Remaining Webhook Code

var _ webhook.CustomValidator = &IntelMachineCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type IntelMachine.
func (v *IntelMachineCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	machine, ok := obj.(*v1alpha2.IntelMachine)
	if !ok {
		return nil, fmt.Errorf("expected a IntelMachine object but got %T", obj)
	}

	imlog.Info("validating IntelMachine creation", "name", machine.GetName())

	return nil, validateIntelMachine(machine)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type IntelMachine.
func (v *IntelMachineCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	machine, ok := newObj.(*v1alpha2.IntelMachine)
	if !ok {
		return nil, fmt.Errorf("expected a IntelMachine object for the newObj but got %T", newObj)
	}

	imlog.Info("validating IntelMachine update", "name", machine.GetName())

	return nil, validateIntelMachine(machine)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type IntelMachine.
func (v *IntelMachineCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	machine, ok := obj.(*v1alpha2.IntelMachine)
	if !ok {
		return nil, fmt.Errorf("expected a IntelMachine object but got %T", obj)
	}

	imlog.Info("validating IntelMachine deletion", "name", machine.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}

// validateIntelMachine validates the fields of a IntelMachine object.
func validateIntelMachine(machine *v1alpha2.IntelMachine) error {
	var allErrs field.ErrorList
	if err := validateIntelMachineName(machine); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := validateIntelMachineSpec(machine); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "infrastructure.cluster.x-k8s.io", Kind: "IntelMachine"},
		machine.Name, allErrs)
}

func validateIntelMachineName(machine *v1alpha2.IntelMachine) *field.Error {
	if len(machine.Name) > validationutils.DNS1035LabelMaxLength-11 {
		return field.Invalid(field.NewPath("metadata").Child("name"), machine.Name, "must be no more than 52 characters")
	}
	return nil
}

func validateIntelMachineSpec(machine *v1alpha2.IntelMachine) *field.Error {
	// TODO! add nodeguid validation
	return nil
}
