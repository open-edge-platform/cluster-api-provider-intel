// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/paused"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1 "github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha1"
	inventory "github.com/open-edge-platform/cluster-api-provider-intel/pkg/inventory"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/scope"
	ccgv1 "github.com/open-edge-platform/cluster-connect-gateway/api/v1alpha1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/finalizers"
)

var (
	requeueAfter                       = 5 * time.Second
	ErrClusterConnectionReady          = errors.New("controlplane endpoint is empty, but ClusterConnection is ready")
	ErrInvalidControlPlaneEndpoint     = errors.New("invalid format of controlplane endpoint")
	ErrInvalidControlPlaneEndpointHost = errors.New("invalid host in controlplane endpoint")
	ErrInvalidControlPlaneEndpointPort = errors.New("invalid port in controlplane endpoint")
	ErrInvalidProviderId               = errors.New("invalid provider id")
	// Predicate to trigger reconciliation only on status.Ready changes in the ClusterConnect resource
	ccUpdatePred = predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldObj := e.ObjectOld.(*ccgv1.ClusterConnect)
			newObj := e.ObjectNew.(*ccgv1.ClusterConnect)

			// ClusterConnect uses []metav1.Condition, not clusterv1.Conditions,
			// so we can't use helpers like conditions.Get().
			var oldCond *metav1.Condition
			for i := range oldObj.Status.Conditions {
				if oldObj.Status.Conditions[i].Type == ccgv1.ConnectionProbeCondition {
					oldCond = &oldObj.Status.Conditions[i]
					break
				}
			}

			var newCond *metav1.Condition
			for i := range newObj.Status.Conditions {
				if newObj.Status.Conditions[i].Type == ccgv1.ConnectionProbeCondition {
					newCond = &newObj.Status.Conditions[i]
					break
				}
			}
			return oldCond.Status != newCond.Status
		},
		CreateFunc:  func(e event.CreateEvent) bool { return false },
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}
)

// IntelClusterReconciler reconciles a IntelCluster object
type IntelClusterReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	InventoryClient inventory.InfrastructureProvider
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=intelclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=intelclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=intelclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=edge-orchestrator.intel.com,resources=clusterconnects,verbs=get;list;watch;create;update;patch;delete

func (r *IntelClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	// logger adds by default the controller name, group &  the resource kind, namespace and name it's reconciling
	log := ctrl.LoggerFrom(ctx)

	intelCluster := &infrav1.IntelCluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, intelCluster); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Info("failed to read intelcluster resource")
		return ctrl.Result{}, err
	}

	// add finalizer first if not set to avoid the race condition between init and delete.
	if finalizerAdded, err := finalizers.EnsureFinalizer(ctx, r.Client, intelCluster, infrav1.ClusterFinalizer); err != nil || finalizerAdded {
		return ctrl.Result{}, err
	}

	cluster, err := util.GetOwnerCluster(ctx, r.Client, intelCluster.ObjectMeta)
	if err != nil {
		log.Info("failed to find cluster owner for intelcluster resource")
		return ctrl.Result{}, err
	}

	if cluster == nil {
		log.Info("cluster owner empty for intelcluster resource")
		return ctrl.Result{}, nil
	}

	if isPaused, conditionChanged, err := paused.EnsurePausedCondition(ctx, r.Client, cluster, intelCluster); err != nil || isPaused || conditionChanged {
		return ctrl.Result{}, err
	}

	ctx = ctrl.LoggerInto(ctx, log)

	clusterScope, err := scope.NewClusterReconcileScopeBuilder().
		WithClient(r.Client).
		WithContext(ctx).
		WithLog(&log).
		WithCluster(cluster).
		WithIntelCluster(intelCluster).
		Build()
	if err != nil {
		log.Error(err, "failed to create cluster reconciliation scope")
		return ctrl.Result{}, err
	}

	if clusterScope == nil {
		return ctrl.Result{}, nil
	}

	defer func() {
		if err := clusterScope.Close(); err != nil && reterr == nil {
			log.Error(err, "failed to patch intelcluster")
			reterr = err
		}
	}()

	if !intelCluster.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, r.reconcileDelete(clusterScope)
	}

	return r.reconcileNormal(clusterScope), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IntelClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.IntelCluster{}).
		Named("intelcluster").
		Owns(&ccgv1.ClusterConnect{}).
		Watches(&ccgv1.ClusterConnect{},
			handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, obj client.Object) []reconcile.Request {
					clusterConnect := obj.(*ccgv1.ClusterConnect)
					log := ctrl.LoggerFrom(ctx)

					// Map ClusterConnect to IntelCluster
					// get namespace from clusterRef
					if clusterConnect.Spec.ClusterRef == nil || clusterConnect.Spec.ClusterRef.Namespace == "" {
						log.Info("ClusterRef is empty or invalid in ClusterConnect resource", "ClusterConnect", clusterConnect.Name)
						return nil
					}
					namespace := clusterConnect.Spec.ClusterRef.Namespace

					// get the IntelCluster name from the ClusterConnect name by trimming the namespace prefix
					name := strings.TrimPrefix(clusterConnect.GetName(), namespace+"-")

					return []reconcile.Request{{
						NamespacedName: types.NamespacedName{
							Name:      name,
							Namespace: namespace,
						},
					}}
				}),
			builder.WithPredicates(ccUpdatePred)).
		Complete(r)
}

func (r *IntelClusterReconciler) reconcileWorkloadCreate(clusterScope *scope.ClusterReconcileScope) bool {
	intelCluster := clusterScope.IntelCluster
	cluster := clusterScope.Cluster
	if intelCluster.Spec.ProviderId != "" {
		return false
	}

	req := inventory.CreateWorkloadInput{TenantId: cluster.Namespace, ClusterName: cluster.Name}
	res := r.InventoryClient.CreateWorkload(req)
	if res.Err != nil {
		// all inventory errors (4xx, 5xx types) are handled generically under just one CR condition. this can be made more granular if needed
		conditions.MarkFalse(intelCluster, infrav1.WorkloadCreatedReadyCondition, infrav1.WaitingForWorkloadToBeProvisonedReason, clusterv1.ConditionSeverityWarning, "%v", res.Err)
		return true
	}

	workloadId := res.WorkloadId
	if workloadId == "" {
		conditions.MarkFalse(intelCluster, infrav1.WorkloadCreatedReadyCondition, infrav1.InvalidWorkloadReason, clusterv1.ConditionSeverityError, "%v", ErrInvalidProviderId)
		return true
	}

	intelCluster.Spec.ProviderId = workloadId
	conditions.MarkTrue(intelCluster, infrav1.WorkloadCreatedReadyCondition)
	return false
}

func (r *IntelClusterReconciler) reconcileControlPlaneEndpoint(scope *scope.ClusterReconcileScope) bool {
	intelCluster := scope.IntelCluster
	if intelCluster.Spec.ControlPlaneEndpoint.IsValid() {
		return false
	}

	clusterConnect := &ccgv1.ClusterConnect{}
	if err := r.Client.Get(scope.Ctx, client.ObjectKey{
		Name:      fmt.Sprintf("%s-%s", scope.IntelCluster.Namespace, scope.IntelCluster.Name),
		Namespace: scope.IntelCluster.Namespace,
	}, clusterConnect); err != nil {
		if !apierrors.IsNotFound(err) {
			scope.Log.Info("failed to read cluster connection resource")
			conditions.MarkFalse(intelCluster, infrav1.ControlPlaneEndpointReadyCondition, infrav1.WaitingForControlPlaneEndpointReason, clusterv1.ConditionSeverityWarning, "%v", err)
			return true
		}

		clusterConnectionItem := getClusterConnectionManifest(scope.Cluster, scope.IntelCluster)
		if err := controllerutil.SetControllerReference(scope.IntelCluster, clusterConnectionItem, r.Scheme); err != nil {
			scope.Log.Info("failed to set owner reference")
			conditions.MarkFalse(intelCluster, infrav1.ControlPlaneEndpointReadyCondition, infrav1.WaitingForControlPlaneEndpointReason, clusterv1.ConditionSeverityWarning, "%v", err)
			return true
		}

		if err := r.Client.Create(scope.Ctx, clusterConnectionItem); err != nil {
			scope.Log.Info("failed to create cluster connection resource")
			conditions.MarkFalse(intelCluster, infrav1.ControlPlaneEndpointReadyCondition, infrav1.WaitingForControlPlaneEndpointReason, clusterv1.ConditionSeverityWarning, "%v", err)
			return true

		}
	}

	if clusterConnect.Status.Ready {
		controlPlaneEndpoint := clusterConnect.Status.ControlPlaneEndpoint

		if controlPlaneEndpoint.IsValid() {
			intelCluster.Spec.ControlPlaneEndpoint = controlPlaneEndpoint
			conditions.MarkTrue(intelCluster, infrav1.ControlPlaneEndpointReadyCondition)
			return false
		}

		scope.Log.Info("invalid control plane endpoint value in clusterconnect resource")
		conditions.MarkFalse(intelCluster, infrav1.ControlPlaneEndpointReadyCondition, infrav1.InvalidControlPlaneEndpointReason, clusterv1.ConditionSeverityError, "%v", ErrInvalidControlPlaneEndpoint)
		return true
	}

	return true
}

func (r *IntelClusterReconciler) reconcileClusterConnectConnection(scope *scope.ClusterReconcileScope) bool {
	intelCluster := scope.IntelCluster
	conditions.MarkUnknown(intelCluster, infrav1.SecureTunnelEstablishedCondition, infrav1.SecureTunnelUnknownReason, "Checking connection to cluster")

	// if provisioning is not complete, do not check the connection
	if !scope.Cluster.Status.ControlPlaneReady {
		return false
	}

	clusterConnect := &ccgv1.ClusterConnect{}
	if err := r.Client.Get(scope.Ctx, client.ObjectKey{
		Name:      fmt.Sprintf("%s-%s", scope.IntelCluster.Namespace, scope.IntelCluster.Name),
		Namespace: scope.IntelCluster.Namespace,
	}, clusterConnect); err != nil {
		if !apierrors.IsNotFound(err) {
			scope.Log.Info("failed to read cluster connection resource")
			return true
		}
	}

	// get the connection probe condition from the clusterconnect resource
	connectionProbeCondition := metav1.Condition{}
	ccConditions := clusterConnect.GetConditions()
	for _, condition := range ccConditions {
		if condition.Type == ccgv1.ConnectionProbeCondition {
			connectionProbeCondition = condition
		}
	}

	switch connectionProbeCondition.Status {
	case metav1.ConditionTrue:
		scope.Log.Info("connection probe condition met in clusterconnect resource")
		conditions.MarkTrue(intelCluster, infrav1.SecureTunnelEstablishedCondition)
	case metav1.ConditionFalse:
		scope.Log.Info("connection probe condition not met in clusterconnect resource")
		conditions.MarkFalse(intelCluster, infrav1.SecureTunnelEstablishedCondition, infrav1.SecureTunnelNotEstablishedReason, clusterv1.ConditionSeverityError, "No connection to cluster, waiting for connection probe condition to be true")
		// do not requeue here, as the clusterconnect object status update event
		// will cause intelCluster reconcile and update the condition when the connection is alive
		return false

	}

	return false
}

func (r *IntelClusterReconciler) reconcileNormal(clusterScope *scope.ClusterReconcileScope) reconcile.Result {
	clusterScope.Log.Info("running intelcluster reconciliation normal")

	if shouldRequeue := r.reconcileControlPlaneEndpoint(clusterScope); shouldRequeue {
		return reconcile.Result{RequeueAfter: requeueAfter}
	}

	if shouldRequeue := r.reconcileWorkloadCreate(clusterScope); shouldRequeue {
		return reconcile.Result{RequeueAfter: requeueAfter}
	}

	// after succesfull provisioning, the status.Ready is always set to true.
	clusterScope.IntelCluster.Status.Ready = true

	if shouldRequeue := r.reconcileClusterConnectConnection(clusterScope); shouldRequeue {
		return reconcile.Result{RequeueAfter: requeueAfter}
	}

	return reconcile.Result{}
}

func (r *IntelClusterReconciler) reconcileWorkloadDelete(clusterScope *scope.ClusterReconcileScope) error {
	providerId := clusterScope.IntelCluster.Spec.ProviderId

	if providerId == "" {
		return nil
	}

	req := inventory.DeleteWorkloadInput{TenantId: clusterScope.IntelCluster.Namespace, WorkloadId: providerId}
	if res := r.InventoryClient.DeleteWorkload(req); res.Err != nil {
		return res.Err
	}

	return nil
}

func (r *IntelClusterReconciler) reconcileClusterConnectDelete(clusterScope *scope.ClusterReconcileScope) error {
	clusterConnect := &ccgv1.ClusterConnect{}
	clusterConnectName := fmt.Sprintf("%s-%s", clusterScope.IntelCluster.Namespace, clusterScope.IntelCluster.Name)

	if err := r.Client.Get(clusterScope.Ctx, client.ObjectKey{
		Name:      clusterConnectName,
		Namespace: clusterScope.Cluster.Namespace,
	}, clusterConnect); err != nil {
		if !apierrors.IsNotFound(err) {
			clusterScope.Log.Info("clusterconnect not found during intel-cluster delete", "name", clusterConnectName)
		}
		return err
	}

	if err := r.Client.Delete(clusterScope.Ctx, clusterConnect); err != nil {
		clusterScope.Log.Info("failed to delete clusterconnect during intel-cluster delete", "name", clusterConnectName)
		return err
	}

	return nil
}

func (r *IntelClusterReconciler) reconcileDelete(clusterScope *scope.ClusterReconcileScope) error {
	clusterScope.Log.Info("running intelcluster reconciliation delete")

	if !controllerutil.ContainsFinalizer(clusterScope.IntelCluster, infrav1.ClusterFinalizer) {
		clusterScope.Log.Info("no finalizer on intelcluster, skipping deletion reconciliation")
		return nil
	}

	if err := r.reconcileWorkloadDelete(clusterScope); err != nil {
		clusterScope.Log.Error(err, "failed to run delete workload reconcile logic")
		return err
	}

	if err := r.reconcileClusterConnectDelete(clusterScope); err != nil {
		clusterScope.Log.Error(err, "failed to run delete clusterconnect reconcile logic")
		return err
	}

	controllerutil.RemoveFinalizer(clusterScope.IntelCluster, infrav1.ClusterFinalizer)
	return nil
}

func getClusterConnectionManifest(cluster *clusterv1.Cluster, intelCluster *infrav1.IntelCluster) *ccgv1.ClusterConnect {
	return &ccgv1.ClusterConnect{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha1",
			Kind:       "ClusterConnection",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", intelCluster.Namespace, intelCluster.Name),
			Namespace: cluster.Namespace,
		},
		Spec: ccgv1.ClusterConnectSpec{
			ServerCertRef: &corev1.ObjectReference{
				Name:      cluster.Name + "-ca",
				Namespace: cluster.Namespace,
			},
			ClientCertRef: &corev1.ObjectReference{
				Name:      cluster.Name + "-cca",
				Namespace: cluster.Namespace,
			},
			ClusterRef: &corev1.ObjectReference{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
			},
		},
	}
}
