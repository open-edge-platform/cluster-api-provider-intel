// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"fmt"
	"time"

	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	infrastructurev1alpha1 "github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha1"
	inventory "github.com/open-edge-platform/cluster-api-provider-intel/pkg/inventory"
	"github.com/pkg/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/conditions"
	clog "sigs.k8s.io/cluster-api/util/log"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/cluster-api/util/paused"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// defaultRequeueAfter is used as a fallback if no other duration should be used.
	defaultRequeueAfter = 10 * time.Second

	// intelMachineBindingKey is used to index IntelMachineBinding objects.
	intelMachineBindingKey = ".metadata.intelMachineBindingKey"
)

var (
	ErrNoCluster = fmt.Errorf("no %q label present", clusterv1.ClusterNameLabel)
)

// IntelMachineReconciler reconciles a IntelMachine object
type IntelMachineReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	InventoryClient inventory.InfrastructureProvider
}

type IntelMachineReconcilerContext struct {
	log          logr.Logger
	ctx          context.Context
	machine      *clusterv1.Machine
	cluster      *clusterv1.Cluster
	intelMachine *infrastructurev1alpha1.IntelMachine
	intelCluster *infrastructurev1alpha1.IntelCluster
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=intelmachines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=intelmachines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=intelmachines/finalizers,verbs=update
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=intelmachinebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=intelmachinebindings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=intelclusters;intelclusters/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=intelmachinetemplates,verbs=get;list;watch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines;machines/status,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *IntelMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, rerr error) {
	rc := IntelMachineReconcilerContext{
		log: log.FromContext(ctx),
		ctx: ctx,
	}

	// Fetch the IntelMachine instance
	intelMachine := &infrastructurev1alpha1.IntelMachine{}
	if err := r.Get(ctx, req.NamespacedName, intelMachine); err != nil {
		if k8serr.IsNotFound(err) {
			rc.log.Info("IntelMachine not found")
		} else {
			rc.log.Error(err, "error fetching IntelMachine")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	rc.intelMachine = intelMachine

	// AddOwners adds the owners of IntelMachine as k/v pairs to the logger.
	ctx, log, err := clog.AddOwners(ctx, r.Client, rc.intelMachine)
	if err != nil {
		return ctrl.Result{}, err
	}
	rc.ctx = ctx
	rc.log = log

	// Fetch the Machine.
	machine, err := util.GetOwnerMachine(ctx, r.Client, intelMachine.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if machine == nil {
		rc.log.Info("Waiting for Machine Controller to set OwnerRef on IntelMachine")
		return ctrl.Result{}, nil
	}
	rc.machine = machine

	// Fetch the Cluster.
	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, rc.machine.ObjectMeta)
	if err != nil {
		rc.log.Info("IntelMachine owner Machine is missing cluster label or cluster does not exist")
		return ctrl.Result{}, err
	}
	rc.cluster = cluster

	if isPaused, conditionChanged, err := paused.EnsurePausedCondition(ctx, r.Client, rc.cluster, rc.intelMachine); err != nil || isPaused || conditionChanged {
		rc.log.Info(fmt.Sprintf("IsPaused or condition changed: %v %v %v", isPaused, conditionChanged, err))
		return ctrl.Result{}, err
	}

	if rc.cluster.Spec.InfrastructureRef == nil {
		rc.log.Info("Cluster infrastructureRef is not available yet")
		return ctrl.Result{}, nil
	}

	// Fetch the Intel Cluster.
	intelCluster := &infrastructurev1alpha1.IntelCluster{}
	key := client.ObjectKey{
		Namespace: intelMachine.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Client.Get(ctx, key, intelCluster); err != nil {
		rc.log.Info("IntelCluster is not available yet")
		return ctrl.Result{}, nil
	}
	rc.intelCluster = intelCluster

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(rc.intelMachine, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Always attempt to Patch the IntelMachine object and status after each reconciliation.
	defer func() {
		if err := patchIntelMachine(ctx, patchHelper, rc.intelMachine); err != nil {
			rc.log.Error(err, "Failed to patch IntelMachine")
			if rerr == nil {
				rerr = err
			}
		}
	}()

	// Handle deleted machines
	if !rc.intelMachine.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, r.reconcileDelete(rc)
	}

	// Add finalizers to the IntelMachine if they are not already present.
	if !controllerutil.ContainsFinalizer(rc.intelMachine, infrastructurev1alpha1.FreeInstanceFinalizer) ||
		!controllerutil.ContainsFinalizer(rc.intelMachine, infrastructurev1alpha1.DeauthHostFinalizer) {
		rc.log.Info("Adding finalizers to IntelMachine", "finalizers", []string{
			infrastructurev1alpha1.FreeInstanceFinalizer,
			infrastructurev1alpha1.DeauthHostFinalizer,
		})
		// FreeInstanceFinalizer is used to remove the instance from the Workload in Inventory.
		controllerutil.AddFinalizer(rc.intelMachine, infrastructurev1alpha1.FreeInstanceFinalizer)
		// DeauthHostFinalizer is used to deauthorize the host in Inventory.
		controllerutil.AddFinalizer(rc.intelMachine, infrastructurev1alpha1.DeauthHostFinalizer)
	}

	// Handle non-deleted machines
	if requeue := r.reconcileNormal(rc); requeue {
		return ctrl.Result{RequeueAfter: defaultRequeueAfter}, nil
	}

	return ctrl.Result{}, nil
}

// Combine cluster name and machine template name into a single IndexField due to cache limitations.
func intelMachineBindingIdxFunc(rawObj client.Object) []string {
	imb := rawObj.(*infrastructurev1alpha1.IntelMachineBinding)
	return []string{getIntelMachineBindingKey(imb.Spec.ClusterName, imb.Spec.IntelMachineTemplateName)}
}

// SetupWithManager sets up the controller with the Manager.
func (r *IntelMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Add field indexer with metadata.intelMachineBindingKey as field name and the IntelMachineBinding key as value.
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&infrastructurev1alpha1.IntelMachineBinding{},
		intelMachineBindingKey,
		intelMachineBindingIdxFunc,
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1alpha1.IntelMachine{}).
		Named("intelmachine").
		Complete(r)
}

func (r *IntelMachineReconciler) reconcileDelete(rc IntelMachineReconcilerContext) error {
	rc.log.Info("Handling deleted IntelMachine")

	// Set the HostProvisionedCondition reporting to indicate delete is started, and issue a patch in order to make
	// this visible to the users.
	patchHelper, err := patch.NewHelper(rc.intelMachine, r.Client)
	if err != nil {
		return err
	}
	conditions.MarkFalse(rc.intelMachine, infrastructurev1alpha1.HostProvisionedCondition, clusterv1.DeletingReason, clusterv1.ConditionSeverityInfo, "")
	if err := patchIntelMachine(rc.ctx, patchHelper, rc.intelMachine); err != nil {
		return errors.Wrap(err, "failed to patch IntelMachine")
	}

	if controllerutil.ContainsFinalizer(rc.intelMachine, infrastructurev1alpha1.HostCleanupFinalizer) {
		// upgrade scenario: remove the obsolete HostCleanupFinalizer
		controllerutil.RemoveFinalizer(rc.intelMachine, infrastructurev1alpha1.HostCleanupFinalizer)
	}

	if controllerutil.ContainsFinalizer(rc.intelMachine, infrastructurev1alpha1.FreeInstanceFinalizer) {
		// Remove the instance from the workload in Inventory
		req := inventory.DeleteInstanceFromWorkloadInput{
			TenantId:   rc.intelCluster.Namespace,
			WorkloadId: rc.intelCluster.Spec.ProviderId,
			InstanceId: *rc.intelMachine.Spec.ProviderID,
		}
		res := r.InventoryClient.DeleteInstanceFromWorkload(req)
		if res.Err != nil && !errors.Is(res.Err, inventory.ErrInvalidWorkloadMembers) {
			rc.log.Error(res.Err, "Failed to delete instance from workload in Inventory")
			return res.Err
		}
		controllerutil.RemoveFinalizer(rc.intelMachine, infrastructurev1alpha1.FreeInstanceFinalizer)
	}

	if controllerutil.ContainsFinalizer(rc.intelMachine, infrastructurev1alpha1.DeauthHostFinalizer) {
		// Deauthorize the host in Inventory
		req := inventory.DeauthorizeHostInput{
			TenantId: rc.intelCluster.Namespace,
			HostUUID: rc.intelMachine.Spec.NodeGUID,
		}
		res := r.InventoryClient.DeauthorizeHost(req)
		if res.Err != nil {
			rc.log.Error(res.Err, "Failed to deauthorize host in Inventory")
			return res.Err
		}
		controllerutil.RemoveFinalizer(rc.intelMachine, infrastructurev1alpha1.DeauthHostFinalizer)
	}

	return nil
}

// reconcileNormal reconciles the IntelMachine in its normal state.  It returns true if the
// IntelMachine should be requeued because it is waiting for an external event to happen.
func (r *IntelMachineReconciler) reconcileNormal(rc IntelMachineReconcilerContext) bool {
	rc.log.Info("Reconciling IntelMachine")

	// Check if the infrastructure is ready, otherwise return and wait for the cluster object to be updated
	if !rc.cluster.Status.InfrastructureReady {
		rc.log.Info("Waiting for IntelCluster Controller to create cluster infrastructure")
		conditions.MarkFalse(rc.intelMachine, infrastructurev1alpha1.HostProvisionedCondition, infrastructurev1alpha1.WaitingForClusterInfrastructureReason, clusterv1.ConditionSeverityInfo, "")
		return true
	}

	// if the machine is already provisioned, check host status and return
	if rc.intelMachine.Spec.ProviderID != nil {
		conditions.MarkTrue(rc.intelMachine, infrastructurev1alpha1.HostProvisionedCondition)

		// SB handler will add the host state as reported by the Cluster Agent to the IntelMachine as an annotation.
		hostState, ok := rc.intelMachine.Annotations[infrastructurev1alpha1.HostStateAnnotation]
		if !ok {
			conditions.MarkFalse(rc.intelMachine, infrastructurev1alpha1.BootstrapExecSucceededCondition, infrastructurev1alpha1.BootstrappingReason, clusterv1.ConditionSeverityInfo, "waiting for SB handler to report host state")
			rc.log.Info("Waiting on SB Handler to report host state")

			// Adding Annotation will trigger a requeue, so we don't need to requeue.
			return false
		}

		switch hostState {
		case infrastructurev1alpha1.HostStateActive:
			rc.intelMachine.Status.Ready = true
			conditions.MarkTrue(rc.intelMachine, infrastructurev1alpha1.BootstrapExecSucceededCondition)
		case infrastructurev1alpha1.HostStateError:
			conditions.MarkFalse(rc.intelMachine, infrastructurev1alpha1.BootstrapExecSucceededCondition, infrastructurev1alpha1.BootstrapFailedReason, clusterv1.ConditionSeverityWarning, "")
		case infrastructurev1alpha1.HostStateInProgress:
			conditions.MarkFalse(rc.intelMachine, infrastructurev1alpha1.BootstrapExecSucceededCondition, infrastructurev1alpha1.BootstrappingReason, clusterv1.ConditionSeverityInfo, "")
		case infrastructurev1alpha1.HostStateInactive:
			conditions.MarkFalse(rc.intelMachine, infrastructurev1alpha1.BootstrapExecSucceededCondition, infrastructurev1alpha1.BootstrapWaitingReason, clusterv1.ConditionSeverityInfo, "")
		default:
			rc.log.Info(fmt.Sprintf("Unexpected host state %q reported by SB Handler", hostState))
		}

		// Updating Annotation will trigger a requeue, so we don't need to requeue.
		return false
	}

	dataSecretName := rc.machine.Spec.Bootstrap.DataSecretName

	// Make sure bootstrap data is available and populated.
	if dataSecretName == nil {
		if !util.IsControlPlaneMachine(rc.machine) && !conditions.IsTrue(rc.cluster, clusterv1.ControlPlaneInitializedCondition) {
			rc.log.Info("Waiting for the control plane to be initialized")
			conditions.MarkFalse(rc.intelMachine, infrastructurev1alpha1.HostProvisionedCondition, clusterv1.WaitingForControlPlaneAvailableReason, clusterv1.ConditionSeverityInfo, "")
			return true
		}

		rc.log.Info("Waiting for the Bootstrap provider controller to set bootstrap data")
		conditions.MarkFalse(rc.intelMachine, infrastructurev1alpha1.HostProvisionedCondition, infrastructurev1alpha1.WaitingForBootstrapDataReason, clusterv1.ConditionSeverityInfo, "")
		return true
	}

	// Get the NodeGUID for the host to reserve in Inventory.
	if rc.intelMachine.Spec.NodeGUID == "" {
		err := r.allocateNodeGUID(rc)
		if err != nil {
			conditions.MarkFalse(rc.intelMachine, infrastructurev1alpha1.HostProvisionedCondition, infrastructurev1alpha1.WaitingForMachineBindingReason, clusterv1.ConditionSeverityWarning, "%v", err)
			rc.log.Info("Error allocating NodeGUID", "error", err)
			return true
		}
		rc.log.Info("Allocated NodeGUID to IntelMachine", "NodeGUID", rc.intelMachine.Spec.NodeGUID)
	}

	// Reserve the host in Inventory
	gmReq := inventory.GetInstanceByMachineIdInput{
		TenantId:  rc.intelCluster.Namespace,
		MachineId: rc.intelMachine.Spec.NodeGUID,
	}
	gmRes := r.InventoryClient.GetInstanceByMachineId(gmReq)
	if gmRes.Err != nil {
		conditions.MarkFalse(rc.intelMachine, infrastructurev1alpha1.HostProvisionedCondition, infrastructurev1alpha1.HostProvisioningFailedReason, clusterv1.ConditionSeverityWarning, "%v", gmRes)
		return true
	}

	aiReq := inventory.AddInstanceToWorkloadInput{
		TenantId:   rc.intelCluster.Namespace,
		WorkloadId: rc.intelCluster.Spec.ProviderId,
		InstanceId: gmRes.Instance.Id,
	}

	if aiRes := r.InventoryClient.AddInstanceToWorkload(aiReq); aiRes.Err != nil {
		conditions.MarkFalse(rc.intelMachine, infrastructurev1alpha1.HostProvisionedCondition, infrastructurev1alpha1.HostProvisioningFailedReason, clusterv1.ConditionSeverityWarning, "%v", aiRes.Err)
		return true
	}

	// Set ProviderID so the Cluster API Machine Controller can pull it
	rc.intelMachine.Spec.ProviderID = &gmRes.Instance.Id
	rc.intelMachine.Annotations[infrastructurev1alpha1.HostIdAnnotation] = gmRes.Host.Id
	conditions.MarkTrue(rc.intelMachine, infrastructurev1alpha1.HostProvisionedCondition)
	return false
}

func (r *IntelMachineReconciler) getTemplateName(rc IntelMachineReconcilerContext) (string, error) {
	templateName, ok := rc.intelMachine.Annotations[clusterv1.TemplateClonedFromNameAnnotation]
	if !ok {
		return "", errors.New("IntelMachine is missing machine template name annotation")
	}

	// Get the IntelMachineTemplate object
	template := &infrastructurev1alpha1.IntelMachineTemplate{}
	templateKey := client.ObjectKey{
		Namespace: rc.intelMachine.Namespace,
		Name:      templateName,
	}
	if err := r.Client.Get(rc.ctx, templateKey, template); err != nil {
		rc.log.Info("Error fetching IntelMachineTemplate, using template name annotation", "templateName", templateName, "error", err)
		return templateName, nil
	}

	// If this template is a clone of another template, return the original template name
	baseTemplateName, ok := template.Annotations[clusterv1.TemplateClonedFromNameAnnotation]
	if ok {
		return baseTemplateName, nil
	}

	return templateName, nil
}

// allocateNodeGUID matches the cluster name and machine template name against unallocated IntelMachineBindings to find
// a free NodeGUID. It marks the chosen IntelMachineBinding as allocated and sets its owner reference to the IntelMachine.
// Finally it adds the NodeGUID to the IntelMachine.
func (r *IntelMachineReconciler) allocateNodeGUID(rc IntelMachineReconcilerContext) error {
	// Fetch the IntelMachineBindings matching the cluster name and machine template name.
	// Cluster API core will add clusterv1.TemplateClonedFromNameAnnotation on intelmachine with the machine template name.
	intelMachineBindingList := &infrastructurev1alpha1.IntelMachineBindingList{}
	templateName, err := r.getTemplateName(rc)
	if err != nil {
		return err
	}
	key := getIntelMachineBindingKey(rc.cluster.Name, templateName)
	if err := r.Client.List(rc.ctx, intelMachineBindingList, client.MatchingFields{intelMachineBindingKey: key}); err != nil {
		return err
	}

	intelmachinebinding := selectIntelMachineBinding(rc, intelMachineBindingList)
	if intelmachinebinding == nil {
		return fmt.Errorf("no IntelMachineBinding found for cluster %q and machine template %q", rc.cluster.Name, templateName)
	}

	patchHelper, err := patch.NewHelper(intelmachinebinding, r.Client)
	if err != nil {
		return err
	}

	// Update the IntelMachineBinding to mark it as allocated and set the owner reference.
	intelmachinebinding.Status.Allocated = true
	if err := ctrl.SetControllerReference(rc.intelMachine, intelmachinebinding, r.Scheme); err != nil {
		return err
	}
	if err := patchHelper.Patch(rc.ctx, intelmachinebinding); err != nil {
		return err
	}

	rc.intelMachine.Spec.NodeGUID = intelmachinebinding.Spec.NodeGUID
	rc.intelMachine.ObjectMeta.Labels[infrastructurev1alpha1.NodeGUIDKey] = intelmachinebinding.Spec.NodeGUID
	return nil
}

func selectIntelMachineBinding(rc IntelMachineReconcilerContext, intelMachineBindingList *infrastructurev1alpha1.IntelMachineBindingList) *infrastructurev1alpha1.IntelMachineBinding {
	log := log.FromContext(rc.ctx)

	ref := metav1.OwnerReference{
		APIVersion: rc.intelMachine.APIVersion,
		Kind:       rc.intelMachine.Kind,
		Name:       rc.intelMachine.Name,
		UID:        rc.intelMachine.UID,
	}

	for i := range intelMachineBindingList.Items {
		imb := &intelMachineBindingList.Items[i]
		// If an IntelMachineBinding is already owned by the reference, return it
		if util.HasOwnerRef(imb.OwnerReferences, ref) {
			log.Info("IntelMachineBinding already allocated to IntelMachine", "NodeGUID", imb.Spec.NodeGUID)
			return imb
		}
	}

	// Find an unallocated IntelMachineBinding.
	for i := range intelMachineBindingList.Items {
		imb := &intelMachineBindingList.Items[i]
		if !imb.Status.Allocated {
			return imb
		}
	}

	return nil
}

func getIntelMachineBindingKey(clusterName, intelMachineTemplateName string) string {
	return fmt.Sprintf("%s/%s", clusterName, intelMachineTemplateName)
}

func patchIntelMachine(ctx context.Context, patchHelper *patch.Helper, intelMachine *infrastructurev1alpha1.IntelMachine) error {
	// Always update the readyCondition by summarizing the state of other conditions.
	// A step counter is added to represent progress during the provisioning process (instead we are hiding the step counter during the deletion process).
	conditions.SetSummary(intelMachine,
		conditions.WithConditions(
			infrastructurev1alpha1.HostProvisionedCondition,
			infrastructurev1alpha1.BootstrapExecSucceededCondition,
		),
		conditions.WithStepCounterIf(intelMachine.ObjectMeta.DeletionTimestamp.IsZero() && intelMachine.Spec.ProviderID == nil),
	)

	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	return patchHelper.Patch(
		ctx,
		intelMachine,
		patch.WithOwnedConditions{Conditions: []clusterv1.ConditionType{
			clusterv1.ReadyCondition,
			infrastructurev1alpha1.HostProvisionedCondition,
			infrastructurev1alpha1.BootstrapExecSucceededCondition,
		}},
	)
}
