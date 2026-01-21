// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	"context"
	"fmt"

	"github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha2"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	validationutils "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:docs-gen:collapse=Imports

var imblog = logf.Log.WithName("intelmachinebinding-resource")

// SetupIntelMachineBindingWebhookWithManager registers the webhook for IntelMachineBinding in the manager.
func SetupIntelMachineBindingWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&v1alpha2.IntelMachineBinding{}).
		WithValidator(&IntelMachineBindingCustomValidator{}).
		WithDefaulter(&IntelMachineBindingCustomDefaulter{
			DefaultHostId:                   "default-node-guid", // TODO? do we want default values of fail if empty? (valdiator)
			DefaultClusterName:              "default-cluster-name",
			DefaultIntelMachineTemplateName: "default-intel-machine-template-name",
		}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-intelmachinebinding-v1alpha2,mutating=true,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=intelmachinebindings,verbs=create;update,versions=v1alpha2,name=mutate-intelmachinebinding-v1alpha2.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

// IntelMachineBindingCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind IntelMachineBinding when those are created or updated.
type IntelMachineBindingCustomDefaulter struct {
	// Default values for various IntelMachineBinding fields
	DefaultHostId                   string
	DefaultClusterName              string
	DefaultIntelMachineTemplateName string
}

var _ webhook.CustomDefaulter = &IntelMachineBindingCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind IntelMachineBinding.
func (d *IntelMachineBindingCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	binding, ok := obj.(*v1alpha2.IntelMachineBinding)
	if !ok {
		return fmt.Errorf("expected an IntelMachineBinding object but got %T", obj)
	}

	imblog.Info("applying default values for IntelMachineBinding", "name", binding.GetName())
	d.applyDefaults(binding)

	return nil
}

// applyDefaults applies default values to IntelMachineBinding fields.
func (d *IntelMachineBindingCustomDefaulter) applyDefaults(binding *v1alpha2.IntelMachineBinding) {
	if binding.Spec.HostId == "" {
		binding.Spec.HostId = d.DefaultHostId
	}
	if binding.Spec.ClusterName == "" {
		binding.Spec.ClusterName = d.DefaultClusterName
	}
	if binding.Spec.IntelMachineTemplateName == "" {
		binding.Spec.IntelMachineTemplateName = d.DefaultIntelMachineTemplateName
	}
}

// +kubebuilder:webhook:path=/validate-intelmachinebinding-v1alpha2,mutating=false,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=intelmachinebindings,verbs=create;update,versions=v1alpha2,name=validate-intelmachinebinding-v1alpha2.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

// IntelMachineBindingCustomValidator struct is responsible for validating the IntelMachineBinding resource
// when it is created, updated, or deleted.
type IntelMachineBindingCustomValidator struct{} // +kubebuilder:docs-gen:collapse=Remaining Webhook Code

var _ webhook.CustomValidator = &IntelMachineBindingCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type IntelMachineBinding.
func (v *IntelMachineBindingCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	binding, ok := obj.(*v1alpha2.IntelMachineBinding)
	if !ok {
		return nil, fmt.Errorf("expected a IntelMachineBinding object but got %T", obj)
	}

	imblog.Info("validating IntelMachineBinding creation", "name", binding.GetName())

	return nil, validateIntelMachineBinding(binding)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type IntelMachineBinding.
func (v *IntelMachineBindingCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	binding, ok := newObj.(*v1alpha2.IntelMachineBinding)
	if !ok {
		return nil, fmt.Errorf("expected a IntelMachineBinding object for the newObj but got %T", newObj)
	}

	imblog.Info("validating IntelMachineBinding update", "name", binding.GetName())

	return nil, validateIntelMachineBinding(binding)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type IntelMachineBinding.
func (v *IntelMachineBindingCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	binding, ok := obj.(*v1alpha2.IntelMachineBinding)
	if !ok {
		return nil, fmt.Errorf("expected a IntelMachineBinding object but got %T", obj)
	}

	imblog.Info("validating IntelMachineBinding deletion", "name", binding.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}

// validateIntelMachineBinding validates the fields of a IntelMachineBinding object.
func validateIntelMachineBinding(binding *v1alpha2.IntelMachineBinding) error {
	var allErrs field.ErrorList
	if err := validateIntelMachineBindingName(binding); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := validateIntelMachineBindingSpec(binding); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "infrastructure.cluster.x-k8s.io", Kind: "IntelMachineBinding"},
		binding.Name, allErrs)
}

func validateIntelMachineBindingName(binding *v1alpha2.IntelMachineBinding) *field.Error {
	if len(binding.Name) > validationutils.DNS1035LabelMaxLength-11 {
		return field.Invalid(field.NewPath("metadata").Child("name"), binding.Name, "must be no more than 52 characters")
	}
	return nil
}

func validateIntelMachineBindingSpec(binding *v1alpha2.IntelMachineBinding) *field.Error {
	// TODO! add hostid validation
	return nil
}
