// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrastructurev1alpha1 "github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha1"
	inventory "github.com/open-edge-platform/cluster-api-provider-intel/pkg/inventory"
	scopepkg "github.com/open-edge-platform/cluster-api-provider-intel/pkg/scope"
	utils "github.com/open-edge-platform/cluster-api-provider-intel/test/utils"
	ccgv1 "github.com/open-edge-platform/cluster-connect-gateway/api/v1alpha1"
)

var _ = Describe("IntelCluster Controller", func() {
	Context("When reconciling a resource", func() {

		const (
			namespaceName   = "275ecb36-5aa8-4c2a-9c47-000000000000"
			clusterName     = "cluster1"
			clusterUID      = "275ecb36-5aa8-4c2a-9c47-d8bb681b9a12"
			intelClusterUID = "275ecb36-5aa8-4c2a-9c47-d8bb681b9a13"
			workloadId      = "w1"
			hostUUID        = "88888888-5aa8-4c2a-9c47-d8bb681b9a14"
			host            = "http://edge-connect-gateway.infra.test"
			port            = int32(3000)

			timeout  = time.Second * 10
			duration = time.Second * 10
			interval = time.Millisecond * 250
		)

		var (
			currentNamespaceName    string
			currentClusterName      string
			intelNamespace          *corev1.Namespace
			intelCluster            *infrastructurev1alpha1.IntelCluster
			cluster                 *clusterv1.Cluster
			clusterConnection       *ccgv1.ClusterConnect
			clusterConnectionFilter client.ObjectKey
			defaultResourceFilter   client.ObjectKey
			endpointUrl             = clusterv1.APIEndpoint{
				Host: host,
				Port: port,
			}
		)

		ctx := context.Background()

		BeforeEach(func() {
			currentNamespaceName = uniqueTestName("275ecb36-5aa8-4c2a-9c47-000000000000")
			currentClusterName = uniqueTestName("cluster1")
			clusterConnectionFilter = client.ObjectKey{
				Name:      fmt.Sprintf("%s-%s", currentNamespaceName, currentClusterName),
				Namespace: currentNamespaceName,
			}
			defaultResourceFilter = client.ObjectKey{
				Name:      currentClusterName,
				Namespace: currentNamespaceName,
			}

			By("creating the custom resource for the Kind IntelCluster")
			intelNamespace = utils.NewNamespace(currentNamespaceName)
			Expect(k8sClient.Create(ctx, intelNamespace)).To(Succeed())
			cluster = utils.NewCluster(currentNamespaceName, currentClusterName)
			Expect(k8sClient.Create(ctx, cluster)).To(Succeed())

			intelCluster = utils.NewIntelClusterNoSpec(cluster)
			Expect(k8sClient.Create(ctx, intelCluster)).To(Succeed())

			cluster.Spec.InfrastructureRef = clusterv1.ContractVersionedObjectReference{
				Kind:     "IntelCluster",
				Name:     intelCluster.Name,
				APIGroup: infrastructurev1alpha1.GroupVersion.Group,
			}
			Expect(k8sClient.Update(ctx, cluster)).To(Succeed())
			cluster.Spec.ControlPlaneRef = clusterv1.ContractVersionedObjectReference{
				Kind: "KubeadmControlPlane",
				Name: cluster.Name + "-controlplane",
			}
			Expect(k8sClient.Update(ctx, cluster)).To(Succeed())

			inventoryClient.On("CreateWorkload",
				inventory.CreateWorkloadInput{TenantId: currentNamespaceName, ClusterName: currentClusterName}).
				Return(inventory.CreateWorkloadOutput{WorkloadId: workloadId, Err: nil}).
				Once()
			inventoryClient.On("DeleteWorkload",
				inventory.DeleteWorkloadInput{TenantId: currentNamespaceName, WorkloadId: workloadId}).
				Return(inventory.DeleteWorkloadOutput{Err: nil}).
				Once()
			inventoryClient.On("DeauthorizeHost",
				inventory.DeauthorizeHostInput{TenantId: currentNamespaceName, HostUUID: hostUUID}).
				Return(inventory.DeauthorizeHostOutput{Err: nil}).
				Once()

			Expect(kerrors.IsNotFound(k8sClient.Get(ctx, clusterConnectionFilter, &ccgv1.ClusterConnect{}))).To(BeTrue())
			clusterConnection = getClusterConnectionManifest(cluster, intelCluster)
			Expect(k8sClient.Create(ctx, clusterConnection)).To(Succeed())
		})

		AfterEach(func() {
			By("delete the IntelCluster and cluster connection")
			Expect(k8sClient.Delete(ctx, intelCluster)).To(Succeed())

			By("the IntelCluster and cluster connection have been deleted")
			Eventually(func(g Gomega) {
				g.Expect(kerrors.IsNotFound(k8sClient.Get(ctx, defaultResourceFilter, &infrastructurev1alpha1.IntelCluster{}))).To(BeTrue())
				g.Expect(kerrors.IsNotFound(k8sClient.Get(ctx, clusterConnectionFilter, &ccgv1.ClusterConnect{}))).To(BeTrue())

				clusterResource := &clusterv1.Cluster{}
				Expect(k8sClient.Get(ctx, defaultResourceFilter, clusterResource)).To(Succeed())
				g.Expect(clusterResource.DeletionTimestamp.IsZero()).To(BeTrue())

			}, timeout, interval).Should(Succeed())

			By("delete the cluster and namespace")
			Expect(k8sClient.Delete(ctx, cluster)).To(Succeed())
			Expect(k8sClient.Delete(ctx, intelNamespace)).To(Succeed())

			By("cluster and namespace have been deleted")
			Eventually(func(g Gomega) {
				g.Expect(kerrors.IsNotFound(k8sClient.Get(ctx, defaultResourceFilter, &clusterv1.Cluster{}))).To(BeTrue())
				g.Expect(kerrors.IsNotFound(k8sClient.Get(ctx, defaultResourceFilter, &corev1.Namespace{}))).To(BeTrue())
			}, timeout, interval).Should(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			resource := &infrastructurev1alpha1.IntelCluster{}
			Expect(k8sClient.Get(ctx, defaultResourceFilter, resource)).To(Succeed())

			By("check initial status of intelcluster")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, defaultResourceFilter, resource)).To(Succeed())
				g.Expect(resource.Spec.ProviderId).To(Equal(""))
				g.Expect(resource.Status.Ready).To(BeFalse())
			}, timeout, interval).Should(Succeed())

			By("updating its status in accordance to ClusterConnection")
			currentClusterConnection := &ccgv1.ClusterConnect{}
			Expect(kerrors.IsNotFound(k8sClient.Get(ctx, clusterConnectionFilter, currentClusterConnection))).To(BeFalse())
			currentClusterConnection.Status.ControlPlaneEndpoint = endpointUrl
			currentClusterConnection.Status.Ready = true
			Expect(k8sClient.Status().Update(ctx, currentClusterConnection)).To(Succeed())

			By("intelcluster status is ready: true")
			Eventually(func(g Gomega) {
				currentClusterConnection := &ccgv1.ClusterConnect{}
				Expect(kerrors.IsNotFound(k8sClient.Get(ctx, clusterConnectionFilter, currentClusterConnection))).To(BeFalse())
				g.Expect(currentClusterConnection.Spec.ClusterRef.Name).To(Equal(cluster.Name))
				g.Expect(currentClusterConnection.Spec.ClusterRef.Namespace).To(Equal(cluster.Namespace))
				resource := &infrastructurev1alpha1.IntelCluster{}
				g.Expect(k8sClient.Get(ctx, defaultResourceFilter, resource)).To(Succeed())

				g.Expect(resource.Spec.ControlPlaneEndpoint.Host).To(Equal(host))
				g.Expect(resource.Spec.ControlPlaneEndpoint.Port).To(Equal(port))
				g.Expect(resource.Spec.ProviderId).To(Equal(workloadId))
				g.Expect(resource.Status.Ready).To(BeTrue())
				g.Expect(resource.Status.Initialization.Provisioned).NotTo(BeNil())
				g.Expect(*resource.Status.Initialization.Provisioned).To(BeTrue())
			}, timeout, interval).Should(Succeed())

		})
	})

	Context("When inventory creation fails", func() {

		const (
			namespaceName   = "275ecb36-5aa8-4c2a-9c47-000000000001"
			clusterName     = "cluster1"
			clusterUID      = "275ecb36-5aa8-4c2a-9c47-d8bb681b9a12"
			intelClusterUID = "275ecb36-5aa8-4c2a-9c47-d8bb681b9a13"
			workloadId      = "w1"
			host            = "http://edge-connect-gateway.infra.test"
			port            = int32(3000)

			timeout  = time.Second * 10
			duration = time.Second * 10
			interval = time.Millisecond * 250
		)

		var (
			currentNamespaceName    string
			currentClusterName      string
			intelNamespace          *corev1.Namespace
			intelCluster            *infrastructurev1alpha1.IntelCluster
			cluster                 *clusterv1.Cluster
			clusterConnection       *ccgv1.ClusterConnect
			clusterConnectionFilter client.ObjectKey
			defaultResourceFilter   client.ObjectKey
			endpointUrl             = clusterv1.APIEndpoint{
				Host: host,
				Port: port,
			}
		)

		ctx := context.Background()

		BeforeEach(func() {
			currentNamespaceName = uniqueTestName("275ecb36-5aa8-4c2a-9c47-000000000001")
			currentClusterName = uniqueTestName("cluster1")
			clusterConnectionFilter = client.ObjectKey{
				Name:      fmt.Sprintf("%s-%s", currentNamespaceName, currentClusterName),
				Namespace: currentNamespaceName,
			}
			defaultResourceFilter = client.ObjectKey{
				Name:      currentClusterName,
				Namespace: currentNamespaceName,
			}

			By("create the necessary resources")

			intelNamespace = utils.NewNamespace(currentNamespaceName)
			Expect(k8sClient.Create(ctx, intelNamespace)).To(Succeed())

			cluster = utils.NewCluster(currentNamespaceName, currentClusterName)
			Expect(k8sClient.Create(ctx, cluster)).To(Succeed())

			intelCluster = utils.NewIntelClusterNoSpec(cluster)
			Expect(k8sClient.Create(ctx, intelCluster)).To(Succeed())

			cluster.Spec.InfrastructureRef = clusterv1.ContractVersionedObjectReference{
				Kind:     "IntelCluster",
				Name:     intelCluster.Name,
				APIGroup: infrastructurev1alpha1.GroupVersion.Group,
			}
			Expect(k8sClient.Update(ctx, cluster)).To(Succeed())

			cluster.Spec.ControlPlaneRef = clusterv1.ContractVersionedObjectReference{
				Kind: "KubeadmControlPlane",
				Name: cluster.Name + "-controlplane",
			}
			Expect(k8sClient.Update(ctx, cluster)).To(Succeed())

			inventoryClient.On("CreateWorkload", inventory.CreateWorkloadInput{TenantId: currentNamespaceName, ClusterName: currentClusterName}).
				Return(inventory.CreateWorkloadOutput{WorkloadId: "", Err: errors.New("test")})

			clusterConnection = getClusterConnectionManifest(cluster, intelCluster)
			Expect(k8sClient.Create(ctx, clusterConnection)).To(Succeed())
		})

		AfterEach(func() {
			By("delete the IntelCluster and cluster connection")
			Expect(k8sClient.Delete(ctx, intelCluster)).To(Succeed())

			By("the IntelCluster and cluster connection have been deleted")
			Eventually(func(g Gomega) {
				g.Expect(kerrors.IsNotFound(k8sClient.Get(ctx, defaultResourceFilter, &infrastructurev1alpha1.IntelCluster{}))).To(BeTrue())
				g.Expect(kerrors.IsNotFound(k8sClient.Get(ctx, clusterConnectionFilter, &ccgv1.ClusterConnect{}))).To(BeTrue())

				clusterResource := &clusterv1.Cluster{}
				Expect(k8sClient.Get(ctx, defaultResourceFilter, clusterResource)).To(Succeed())
				g.Expect(clusterResource.DeletionTimestamp.IsZero()).To(BeTrue())

			}, timeout, interval).Should(Succeed())

			By("delete the cluster and namespace")
			Expect(k8sClient.Delete(ctx, cluster)).To(Succeed())
			Expect(k8sClient.Delete(ctx, intelNamespace)).To(Succeed())

			By("cluster and namespace have been deleted")
			Eventually(func(g Gomega) {
				g.Expect(kerrors.IsNotFound(k8sClient.Get(ctx, defaultResourceFilter, &clusterv1.Cluster{}))).To(BeTrue())
				g.Expect(kerrors.IsNotFound(k8sClient.Get(ctx, defaultResourceFilter, &corev1.Namespace{}))).To(BeTrue())
			}, timeout, interval).Should(Succeed())
		})
		It("should have empty providerId", func() {
			resource := &infrastructurev1alpha1.IntelCluster{}
			Expect(k8sClient.Get(ctx, defaultResourceFilter, resource)).To(Succeed())

			By("check initial status of intelcluster")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, defaultResourceFilter, resource)).To(Succeed())
				g.Expect(resource.Spec.ProviderId).To(Equal(""))
				g.Expect(resource.Status.Ready).To(BeFalse())
			}, timeout, interval).Should(Succeed())

			By("updating its status in accordance to ClusterConnection")
			currentClusterConnection := &ccgv1.ClusterConnect{}
			Expect(kerrors.IsNotFound(k8sClient.Get(ctx, clusterConnectionFilter, currentClusterConnection))).To(BeFalse())
			currentClusterConnection.Status.ControlPlaneEndpoint = endpointUrl
			currentClusterConnection.Status.Ready = true
			Expect(k8sClient.Status().Update(ctx, currentClusterConnection)).To(Succeed())

			By("intelcluster status is ready: false")
			Eventually(func(g Gomega) {
				currentClusterConnection := &ccgv1.ClusterConnect{}
				Expect(kerrors.IsNotFound(k8sClient.Get(ctx, clusterConnectionFilter, currentClusterConnection))).To(BeFalse())
				resource := &infrastructurev1alpha1.IntelCluster{}
				g.Expect(k8sClient.Get(ctx, defaultResourceFilter, resource)).To(Succeed())

				g.Expect(resource.Spec.ControlPlaneEndpoint.Host).To(Equal(host))
				g.Expect(resource.Spec.ControlPlaneEndpoint.Port).To(Equal(port))
				g.Expect(resource.Spec.ProviderId).To(Equal(""))
				g.Expect(resource.Status.Ready).To(BeFalse())
			}, timeout, interval).Should(Succeed())

		})
	})
})

func uniqueTestName(base string) string {
	return fmt.Sprintf("%s-%06d", base, time.Now().UnixNano()%1000000)
}

var _ = Describe("Reconcile loop errors", func() {
	Context("When reconciling a resource", func() {
		const (
			clusterName      = "test-cluster"
			intelClusterName = "test-intelcluster"
			namespace        = "default"
		)

		var (
			intelcluster *infrastructurev1alpha1.IntelCluster
			cluster      *clusterv1.Cluster
			reconciler   *IntelClusterReconciler
			fakeClient   client.Client
		)

		ctx := context.Background()

		BeforeEach(func() {
			scheme := runtime.NewScheme()
			Expect(infrastructurev1alpha1.AddToScheme(scheme)).To(Succeed())
			Expect(clusterv1.AddToScheme(scheme)).To(Succeed())
			Expect(ccgv1.AddToScheme(scheme)).To(Succeed()) // Register ClusterConnect type

			cluster = utils.NewCluster(namespace, clusterName)
			conditions.Set(cluster, metav1.Condition{
				Type:   string(clusterv1.InfrastructureReadyCondition),
				Status: metav1.ConditionTrue,
				Reason: "Ready",
			})
			intelcluster = utils.NewIntelCluster(namespace, intelClusterName, "provider-id", cluster)

			fakeClient = fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(cluster, intelcluster).
				Build()

			reconciler = &IntelClusterReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}
		})

		It("should return error if IntelCluster's owner not found", func() {
			key := types.NamespacedName{Name: intelClusterName, Namespace: namespace}
			req := reconcile.Request{
				NamespacedName: key,
			}

			ic := &infrastructurev1alpha1.IntelCluster{}
			Expect(fakeClient.Get(ctx, key, ic)).To(Succeed())
			ic.OwnerReferences[0].Name = "bad-owner"
			Expect(fakeClient.Update(ctx, ic)).To(Succeed())

			// Finalizer added on the first Reconcile() call
			res, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))

			res, err = reconciler.Reconcile(ctx, req)
			Expect(err).To(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))
		})

		It("should not return error if IntelCluster has no owner", func() {
			key := types.NamespacedName{Name: intelClusterName, Namespace: namespace}
			req := reconcile.Request{
				NamespacedName: key,
			}

			ic := &infrastructurev1alpha1.IntelCluster{}
			Expect(fakeClient.Get(ctx, key, ic)).To(Succeed())
			ic.OwnerReferences = nil
			Expect(fakeClient.Update(ctx, ic)).To(Succeed())

			// Finalizer added on the first Reconcile() call
			res, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))

			res, err = reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))
		})

		It("should requeue when connection probe condition is unknown", func() {
			scheme := runtime.NewScheme()
			Expect(infrastructurev1alpha1.AddToScheme(scheme)).To(Succeed())
			Expect(clusterv1.AddToScheme(scheme)).To(Succeed())
			Expect(ccgv1.AddToScheme(scheme)).To(Succeed())

			cluster = utils.NewCluster(namespace, clusterName)
			conditions.Set(cluster, metav1.Condition{
				Type:   string(clusterv1.ClusterControlPlaneAvailableCondition),
				Status: metav1.ConditionTrue,
				Reason: "Available",
			})

			intelcluster = utils.NewIntelCluster(namespace, intelClusterName, "provider-id", cluster)

			clusterConnect := getClusterConnectionManifest(cluster, intelcluster)
			clusterConnect.Status.Ready = true
			clusterConnect.Status.ControlPlaneEndpoint = clusterv1.APIEndpoint{
				Host: "api.example.invalid",
				Port: 6443,
			}
			clusterConnect.Status.Conditions = []metav1.Condition{{
				Type:   ccgv1.ConnectionProbeCondition,
				Status: metav1.ConditionUnknown,
				Reason: "Pending",
			}}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(cluster, intelcluster, clusterConnect).
				Build()

			logger := logr.Discard()
			clusterScope, err := scopepkg.NewClusterReconcileScopeBuilder().
				WithContext(ctx).
				WithLog(&logger).
				WithClient(fakeClient).
				WithCluster(cluster).
				WithIntelCluster(intelcluster).
				Build()
			Expect(err).NotTo(HaveOccurred())

			reconciler = &IntelClusterReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			result := reconciler.reconcileNormal(clusterScope)

			Expect(result.RequeueAfter).To(Equal(requeueAfter))
			Expect(clusterScope.IntelCluster.Spec.ControlPlaneEndpoint.Host).To(Equal("api.example.invalid"))
			Expect(clusterScope.IntelCluster.Status.Ready).To(BeTrue())
			Expect(clusterScope.IntelCluster.Status.Initialization.Provisioned).NotTo(BeNil())
			Expect(*clusterScope.IntelCluster.Status.Initialization.Provisioned).To(BeTrue())
			Expect(conditions.IsFalse(clusterScope.IntelCluster, string(infrastructurev1alpha1.SecureTunnelEstablishedCondition))).To(BeTrue())
			Expect(conditions.GetReason(clusterScope.IntelCluster, string(infrastructurev1alpha1.SecureTunnelEstablishedCondition))).To(Equal(infrastructurev1alpha1.SecureTunnelUnknownReason))
		})
	})
})
