// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrastructurev1alpha1 "github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha1"
	inventory "github.com/open-edge-platform/cluster-api-provider-intel/pkg/inventory"
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
			hostId          = "host-12345678"
			host            = "http://edge-connect-gateway.infra.test"
			port            = int32(3000)

			timeout  = time.Second * 10
			duration = time.Second * 10
			interval = time.Millisecond * 250
		)

		var (
			intelNamespace          *corev1.Namespace
			intelCluster            *infrastructurev1alpha1.IntelCluster
			cluster                 *clusterv1.Cluster
			clusterConnection       *ccgv1.ClusterConnect
			clusterConnectionFilter = client.ObjectKey{
				Name:      fmt.Sprintf("%s-%s", namespaceName, clusterName),
				Namespace: namespaceName}
			defaultResourceFilter = client.ObjectKey{
				Name:      clusterName,
				Namespace: namespaceName,
			}
			endpointUrl = clusterv1.APIEndpoint{
				Host: host,
				Port: port,
			}
		)

		ctx := context.Background()

		BeforeEach(func() {
			By("creating the custom resource for the Kind IntelCluster")
			intelNamespace = utils.NewNamespace(namespaceName)
			Expect(k8sClient.Create(ctx, intelNamespace)).To(Succeed())
			Expect(kerrors.IsNotFound(k8sClient.Get(ctx, defaultResourceFilter, &clusterv1.Cluster{}))).To(BeTrue())
			cluster = utils.NewCluster(namespaceName, clusterName)
			Expect(k8sClient.Create(ctx, cluster)).To(Succeed())

			Expect(kerrors.IsNotFound(k8sClient.Get(ctx, defaultResourceFilter, &infrastructurev1alpha1.IntelCluster{}))).To(BeTrue())
			intelCluster = utils.NewIntelClusterNoSpec(cluster)
			Expect(k8sClient.Create(ctx, intelCluster)).To(Succeed())

			cluster.Spec.InfrastructureRef = utils.GetObjectRef(&intelCluster.ObjectMeta, "IntelCluster")
			Expect(k8sClient.Update(ctx, cluster)).To(Succeed())
			// TODO: refactor to reuse GetObjectRef
			cluster.Spec.ControlPlaneRef = &corev1.ObjectReference{Name: cluster.Name + "-controlplane"}
			Expect(k8sClient.Update(ctx, cluster)).To(Succeed())

			inventoryClient.On("CreateWorkload",
				inventory.CreateWorkloadInput{TenantId: namespaceName, ClusterName: clusterName}).
				Return(inventory.CreateWorkloadOutput{WorkloadId: workloadId, Err: nil}).
				Once()
			inventoryClient.On("DeleteWorkload",
				inventory.DeleteWorkloadInput{TenantId: namespaceName, WorkloadId: workloadId}).
				Return(inventory.DeleteWorkloadOutput{Err: nil}).
				Once()
			inventoryClient.On("DeauthorizeHost",
				inventory.DeauthorizeHostInput{TenantId: namespaceName, HostId: hostId}).
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
			intelNamespace          *corev1.Namespace
			intelCluster            *infrastructurev1alpha1.IntelCluster
			cluster                 *clusterv1.Cluster
			clusterConnection       *ccgv1.ClusterConnect
			clusterConnectionFilter = client.ObjectKey{
				Name:      fmt.Sprintf("%s-%s", namespaceName, clusterName),
				Namespace: namespaceName}
			defaultResourceFilter = client.ObjectKey{
				Name:      clusterName,
				Namespace: namespaceName,
			}
			endpointUrl = clusterv1.APIEndpoint{
				Host: host,
				Port: port,
			}
		)

		ctx := context.Background()

		BeforeEach(func() {
			By("create the necessary resources")

			Expect(kerrors.IsNotFound(k8sClient.Get(ctx, defaultResourceFilter, &corev1.Namespace{}))).To(BeTrue())
			intelNamespace = utils.NewNamespace(namespaceName)
			Expect(k8sClient.Create(ctx, intelNamespace)).To(Succeed())

			Expect(kerrors.IsNotFound(k8sClient.Get(ctx, defaultResourceFilter, &clusterv1.Cluster{}))).To(BeTrue())
			cluster = utils.NewCluster(namespaceName, clusterName)
			Expect(k8sClient.Create(ctx, cluster)).To(Succeed())

			Expect(kerrors.IsNotFound(k8sClient.Get(ctx, defaultResourceFilter, &infrastructurev1alpha1.IntelCluster{}))).To(BeTrue())
			intelCluster = utils.NewIntelClusterNoSpec(cluster)
			Expect(k8sClient.Create(ctx, intelCluster)).To(Succeed())

			cluster.Spec.InfrastructureRef = utils.GetObjectRef(&intelCluster.ObjectMeta, "IntelCluster")
			Expect(k8sClient.Update(ctx, cluster)).To(Succeed())

			// TODO: refactor to reuse GetObjectRef
			cluster.Spec.ControlPlaneRef = &corev1.ObjectReference{Name: cluster.Name + "-controlplane"}
			Expect(k8sClient.Update(ctx, cluster)).To(Succeed())

			inventoryClient.On("CreateWorkload", inventory.CreateWorkloadInput{TenantId: namespaceName, ClusterName: clusterName}).
				Return(inventory.CreateWorkloadOutput{WorkloadId: "", Err: errors.New("test")})

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
			cluster.Status.InfrastructureReady = true
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
	})
})
