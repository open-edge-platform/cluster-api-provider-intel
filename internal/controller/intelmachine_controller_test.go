// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	cutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrastructurev1alpha1 "github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha1"
	"github.com/open-edge-platform/cluster-api-provider-intel/mocks/m_inventory"
	inventory "github.com/open-edge-platform/cluster-api-provider-intel/pkg/inventory"
	utils "github.com/open-edge-platform/cluster-api-provider-intel/test/utils"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/cluster-api/util/conditions"
)

const (
	namespace     = "default"
	bootstrapKind = "RKE2Config"
)

var _ = Describe("IntelMachine Controller", func() {
	Context("When reconciling a resource", func() {
		const (
			machineName             = "test-machine"
			clusterName             = "test-cluster"
			intelMachineName        = "test-intelmachine"
			nodeGUID                = "1234567890"
			intelClusterName        = "test-intelcluster"
			intelMachineBindingName = "test-intelmachinebinding"
			machineTemplateName     = "test-machinetemplate"
			workloadId              = "test-workload-id"
			instanceId              = "test-instance-id"
			hostId                  = "test-host-id"

			timeout        = time.Second * 10
			cleanupTimeout = time.Second * 30
			interval       = time.Millisecond * 250
		)

		var (
			testNamespace       string
			namespaceObj        *corev1.Namespace
			intelmachine        *infrastructurev1alpha1.IntelMachine
			intelmachinebinding *infrastructurev1alpha1.IntelMachineBinding
			intelcluster        *infrastructurev1alpha1.IntelCluster
			machine             *clusterv1.Machine
			cluster             *clusterv1.Cluster
		)

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{}

		BeforeEach(func() {
			By("creating the custom resources")
			testNamespace = uniqueTestName(namespace)
			namespaceObj = &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testNamespace}}
			Expect(k8sClient.Create(ctx, namespaceObj)).To(Succeed())
			typeNamespacedName = types.NamespacedName{
				Name:      intelMachineName,
				Namespace: testNamespace,
			}

			// Create the cluster.
			cluster = utils.NewCluster(testNamespace, clusterName)
			Expect(k8sClient.Create(ctx, cluster)).To(Succeed())
			conditions.Set(cluster, metav1.Condition{
				Type:   string(clusterv1.InfrastructureReadyCondition),
				Status: metav1.ConditionTrue,
				Reason: "Ready",
			})
			Expect(k8sClient.Status().Update(ctx, cluster)).To(Succeed())

			// Create the intelcluster.
			intelcluster = utils.NewIntelCluster(testNamespace, intelClusterName, workloadId, cluster)
			Expect(k8sClient.Create(ctx, intelcluster)).To(Succeed())

			// Update the cluster's infrastructureRef to point to the intelcluster
			cluster.Spec.InfrastructureRef = clusterv1.ContractVersionedObjectReference{
				Kind:     "IntelCluster",
				Name:     intelcluster.Name,
				APIGroup: infrastructurev1alpha1.GroupVersion.Group,
			}
			Expect(k8sClient.Update(ctx, cluster)).To(Succeed())

			// Create the machine.
			machine = utils.NewMachine(testNamespace, clusterName, machineName, bootstrapKind)
			Expect(k8sClient.Create(ctx, machine)).To(Succeed())

			// Create the intelmachinebinding.
			intelmachinebinding = utils.NewIntelMachineBinding(testNamespace, intelMachineBindingName, nodeGUID, clusterName, machineTemplateName)
			Expect(k8sClient.Create(ctx, intelmachinebinding)).To(Succeed())

			// Add mocks before creating IntelMachine
			host := &inventory.Host{
				Id: hostId,
			}

			instance := &inventory.Instance{
				Id: instanceId,
			}

			inventoryClient.On("GetInstanceByMachineId", inventory.GetInstanceByMachineIdInput{TenantId: testNamespace, MachineId: nodeGUID}).
				Return(inventory.GetInstanceByMachineIdOutput{Host: host, Instance: instance, Err: nil}).
				Once()
			inventoryClient.On("AddInstanceToWorkload",
				inventory.AddInstanceToWorkloadInput{TenantId: testNamespace, WorkloadId: workloadId, InstanceId: instanceId}).
				Return(inventory.AddInstanceToWorkloadOutput{Err: nil}).
				Once()
			inventoryClient.On("DeleteInstanceFromWorkload",
				inventory.DeleteInstanceFromWorkloadInput{TenantId: testNamespace, WorkloadId: workloadId, InstanceId: instanceId}).
				Return(inventory.DeleteInstanceFromWorkloadOutput{Err: nil}).
				Once()
			inventoryClient.On("DeleteWorkload",
				inventory.DeleteWorkloadInput{TenantId: testNamespace, WorkloadId: workloadId}).
				Return(inventory.DeleteWorkloadOutput{Err: nil}).
				Once()
			inventoryClient.On("DeauthorizeHost",
				inventory.DeauthorizeHostInput{TenantId: testNamespace, HostUUID: nodeGUID}).
				Return(inventory.DeauthorizeHostOutput{Err: nil}).
				Once()

			// Create the intelmachine.
			intelmachine = utils.NewIntelMachine(testNamespace, intelMachineName, machine)
			intelmachine.Annotations = map[string]string{
				clusterv1.TemplateClonedFromNameAnnotation: machineTemplateName,
			}
			Expect(k8sClient.Create(ctx, intelmachine)).To(Succeed())

			// Update the machine to point to the intelmachine
			machine.Spec.InfrastructureRef = clusterv1.ContractVersionedObjectReference{
				Kind:     "IntelMachine",
				Name:     intelmachine.Name,
				APIGroup: infrastructurev1alpha1.GroupVersion.Group,
			}
			Expect(k8sClient.Update(ctx, machine)).To(Succeed())
		})

		AfterEach(func() {
			// Apparently cascading deletes based on OwnerReferences are not handled.
			// Delete all the resources in the reverse order they were created.
			By("deleting the IntelMachine")
			Expect(k8sClient.Delete(ctx, intelmachine)).To(Succeed())

			By("Waiting for HostProvisionedCondition to be False")
			Eventually(func(g Gomega) {
				resource := &infrastructurev1alpha1.IntelMachine{}
				g.Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
				g.Expect(resource.DeletionTimestamp.IsZero()).To(BeFalse())
				g.Expect(conditions.IsFalse(resource, string(infrastructurev1alpha1.HostProvisionedCondition))).To(BeTrue())
			}, cleanupTimeout, interval).Should(Succeed())

			By("Removing the IntelMachine's HostCleanupFinalizer")
			Expect(k8sClient.Get(ctx, typeNamespacedName, intelmachine)).To(Succeed())
			cutil.RemoveFinalizer(intelmachine, infrastructurev1alpha1.HostCleanupFinalizer)
			Expect(k8sClient.Update(ctx, intelmachine)).To(Succeed())

			By("Deleting the other custom resources")
			Expect(k8sClient.Delete(ctx, intelmachinebinding)).To(Succeed())
			Expect(k8sClient.Delete(ctx, machine)).To(Succeed())
			Expect(k8sClient.Delete(ctx, intelcluster)).To(Succeed())

			By("Checking the IntelMachine is removed")
			Eventually(func(g Gomega) {
				g.Expect(errors.IsNotFound(k8sClient.Get(ctx, typeNamespacedName, &infrastructurev1alpha1.IntelMachine{}))).To(BeTrue())
			}, timeout, interval).Should(Succeed())

			By("Checking the IntelMachineBinding is removed")
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: intelMachineBindingName, Namespace: testNamespace}
				g.Expect(errors.IsNotFound(k8sClient.Get(ctx, key, &infrastructurev1alpha1.IntelMachineBinding{}))).To(BeTrue())
			}, timeout, interval).Should(Succeed())

			By("Checking the IntelCluster and Machine are removed")
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: intelClusterName, Namespace: testNamespace}
				g.Expect(errors.IsNotFound(k8sClient.Get(ctx, key, &infrastructurev1alpha1.IntelCluster{}))).To(BeTrue())
				key = types.NamespacedName{Name: machineName, Namespace: testNamespace}
				g.Expect(errors.IsNotFound(k8sClient.Get(ctx, key, &clusterv1.Machine{}))).To(BeTrue())
			}, timeout, interval).Should(Succeed())

			By("Delete the cluster and namespace")
			Expect(k8sClient.Delete(ctx, cluster)).To(Succeed())
			Expect(k8sClient.Delete(ctx, namespaceObj)).To(Succeed())

			By("Checking the Cluster and namespace are removed")
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: clusterName, Namespace: testNamespace}
				g.Expect(errors.IsNotFound(k8sClient.Get(ctx, key, &clusterv1.Cluster{}))).To(BeTrue())
				ns := &corev1.Namespace{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: testNamespace}, ns)
				if errors.IsNotFound(err) {
					return
				}
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(ns.DeletionTimestamp).NotTo(BeNil())
			}, timeout, interval).Should(Succeed())
		})

		It("should successfully reconcile the intelmachine", func() {
			resource := &infrastructurev1alpha1.IntelMachine{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Checking that the IntelMachineBinding is allocated")
			key := types.NamespacedName{Name: intelMachineBindingName, Namespace: testNamespace}
			Eventually(func(g Gomega) {
				imb := &infrastructurev1alpha1.IntelMachineBinding{}
				g.Expect(k8sClient.Get(ctx, key, imb)).To(Succeed())
				g.Expect(imb.Status.Allocated).To(BeTrue())
				g.Expect(imb.OwnerReferences).To(HaveLen(1))
			}, timeout, interval).Should(Succeed())

			By("Checking that condition HostProvisionedCondition is True and Provider ID is set")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
				g.Expect(resource.Spec.ProviderID).NotTo(BeNil())
				g.Expect(resource.Spec.NodeGUID).To(Equal(nodeGUID))
				g.Expect(resource.ObjectMeta.Labels[infrastructurev1alpha1.NodeGUIDKey]).To(Equal(nodeGUID))
				g.Expect(resource.Status.Ready).To(BeFalse())
				g.Expect(conditions.IsTrue(resource, string(infrastructurev1alpha1.HostProvisionedCondition))).To(BeTrue())
				g.Expect(conditions.IsFalse(resource, string(infrastructurev1alpha1.BootstrapExecSucceededCondition))).To(BeTrue())
				g.Expect(conditions.GetReason(resource, string(infrastructurev1alpha1.BootstrapExecSucceededCondition))).To(Equal(infrastructurev1alpha1.BootstrappingReason))
				g.Expect(conditions.IsFalse(resource, string(clusterv1.ReadyCondition))).To(BeTrue())
			}, timeout, interval).Should(Succeed())

			By("By checking the IntelMachine has the expected finalizers")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
				g.Expect(cutil.ContainsFinalizer(resource, infrastructurev1alpha1.FreeInstanceFinalizer)).To(BeTrue())
				g.Expect(cutil.ContainsFinalizer(resource, infrastructurev1alpha1.HostCleanupFinalizer)).To(BeTrue())
			}, timeout, interval).Should(Succeed())

			By("Updating the Host State annotation")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(resource.Annotations).NotTo(HaveKey(infrastructurev1alpha1.HostStateAnnotation))
			if resource.Annotations == nil {
				resource.Annotations = make(map[string]string)
			}
			resource.Annotations[infrastructurev1alpha1.HostStateAnnotation] = infrastructurev1alpha1.HostStateActive
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Checking that the IntelMachine is ready")
			Eventually(func(g Gomega) {
				resource := &infrastructurev1alpha1.IntelMachine{}
				g.Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
				g.Expect(resource.Spec.ProviderID).NotTo(BeNil())
				g.Expect(resource.Status.Ready).To(BeTrue())
				g.Expect(resource.Status.Initialization.Provisioned).NotTo(BeNil())
				g.Expect(*resource.Status.Initialization.Provisioned).To(BeTrue())
				g.Expect(resource.Annotations[infrastructurev1alpha1.HostIdAnnotation]).To(Equal(hostId))
				g.Expect(conditions.IsTrue(resource, string(infrastructurev1alpha1.HostProvisionedCondition))).To(BeTrue())
				g.Expect(conditions.IsTrue(resource, string(infrastructurev1alpha1.BootstrapExecSucceededCondition))).To(BeTrue())
				g.Expect(conditions.IsTrue(resource, string(clusterv1.ReadyCondition))).To(BeTrue())
			}, timeout, interval).Should(Succeed())

			By("Checking that the owner Machine has skip-remediation annotation")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: machineName, Namespace: testNamespace}, machine)).To(Succeed())
				g.Expect(machine.Annotations).To(HaveKey(clusterv1.MachineSkipRemediationAnnotation))
			}, timeout, interval).Should(Succeed())
		})
	})
})

var _ = Describe("Reconcile loop errors", func() {

	Context("When reconciling a resource", func() {
		const (
			machineName             = "test-machine"
			clusterName             = "test-cluster"
			intelMachineName        = "test-intelmachine"
			nodeGUID                = "1234567890"
			intelClusterName        = "test-intelcluster"
			intelMachineBindingName = "test-intelmachinebinding"
			machineTemplateName     = "test-machinetemplate"
			workloadId              = "test-workload-id"
			instanceId              = "test-instance-id"
			bootstrapKind           = "BootstrapConfig"
			hostId                  = "test-host-id"
		)

		var (
			intelmachine        *infrastructurev1alpha1.IntelMachine
			intelmachinebinding *infrastructurev1alpha1.IntelMachineBinding
			intelcluster        *infrastructurev1alpha1.IntelCluster
			machine             *clusterv1.Machine
			cluster             *clusterv1.Cluster
			reconciler          *IntelMachineReconciler
			fakeClient          client.Client
			inventoryClient     *m_inventory.MockInfrastructureProvider
		)

		ctx := context.Background()

		BeforeEach(func() {
			By("creating the custom resources")

			scheme := runtime.NewScheme()
			Expect(infrastructurev1alpha1.AddToScheme(scheme)).To(Succeed())
			Expect(clusterv1.AddToScheme(scheme)).To(Succeed())

			cluster = utils.NewCluster(namespace, clusterName)
			conditions.Set(cluster, metav1.Condition{
				Type:   string(clusterv1.InfrastructureReadyCondition),
				Status: metav1.ConditionTrue,
				Reason: "Ready",
			})
			intelcluster = utils.NewIntelCluster(namespace, intelClusterName, workloadId, cluster)
			machine = utils.NewMachine(namespace, clusterName, machineName, bootstrapKind)
			intelmachinebinding = utils.NewIntelMachineBinding(namespace, intelMachineBindingName, nodeGUID, clusterName, machineTemplateName)
			intelmachine = utils.NewIntelMachine(namespace, intelMachineName, machine)
			intelmachine.Annotations = map[string]string{
				clusterv1.TemplateClonedFromNameAnnotation: machineTemplateName,
			}

			fakeClient = fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(cluster, intelcluster, machine, intelmachine, intelmachinebinding).
				WithStatusSubresource(intelmachine, intelmachinebinding).
				WithIndex(
					&infrastructurev1alpha1.IntelMachineBinding{},
					intelMachineBindingKey,
					intelMachineBindingIdxFunc).
				Build()
			inventoryClient = &m_inventory.MockInfrastructureProvider{}

			reconciler = &IntelMachineReconciler{
				Client:          fakeClient,
				Scheme:          scheme,
				InventoryClient: inventoryClient,
			}

			// Update the cluster's infrastructureRef to point to the intelcluster
			cluster.Spec.InfrastructureRef = clusterv1.ContractVersionedObjectReference{
				Kind:     "IntelCluster",
				Name:     intelcluster.Name,
				APIGroup: infrastructurev1alpha1.GroupVersion.Group,
			}
			Expect(fakeClient.Update(ctx, cluster)).To(Succeed())

			// Update the machine to point to the intelmachine
			machine.Spec.InfrastructureRef = clusterv1.ContractVersionedObjectReference{
				Kind:     "IntelMachine",
				Name:     intelmachine.Name,
				APIGroup: infrastructurev1alpha1.GroupVersion.Group,
			}
			Expect(fakeClient.Update(ctx, machine)).To(Succeed())

			// Add mocks for IntelMachine
			host := &inventory.Host{
				Id: hostId,
			}

			instance := &inventory.Instance{
				Id: instanceId,
			}

			inventoryClient.On("GetInstanceByMachineId", inventory.GetInstanceByMachineIdInput{TenantId: namespace, MachineId: nodeGUID}).
				Return(inventory.GetInstanceByMachineIdOutput{Host: host, Instance: instance, Err: nil}).
				Once()
			inventoryClient.On("AddInstanceToWorkload",
				inventory.AddInstanceToWorkloadInput{TenantId: namespace, WorkloadId: workloadId, InstanceId: instanceId}).
				Return(inventory.AddInstanceToWorkloadOutput{Err: nil}).
				Once()
			inventoryClient.On("DeleteInstanceFromWorkload",
				inventory.DeleteInstanceFromWorkloadInput{TenantId: namespace, WorkloadId: workloadId, InstanceId: instanceId}).
				Return(inventory.DeleteInstanceFromWorkloadOutput{Err: nil}).
				Once()
			inventoryClient.On("DeleteWorkload",
				inventory.DeleteWorkloadInput{TenantId: namespace, WorkloadId: workloadId}).
				Return(inventory.DeleteWorkloadOutput{Err: nil}).
				Once()
		})

		It("should reconcile IntelMachine successfully", func() {
			key := types.NamespacedName{Name: intelMachineName, Namespace: namespace}
			req := reconcile.Request{
				NamespacedName: key,
			}

			// Reconcile() returns after paused.EnsurePausedCondition()
			res, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))

			// GUID is allocated to machine
			res, err = reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))

			// Machine is not fully Ready at this point
			im := &infrastructurev1alpha1.IntelMachine{}
			Expect(fakeClient.Get(ctx, key, im)).To(Succeed())
			Expect(im.Spec.ProviderID).NotTo(BeNil())
			Expect(im.Status.Ready).To(BeFalse())
		})

		It("should set initialization provisioned when host becomes active", func() {
			logger := logr.Discard()
			providerID := instanceId
			intelmachine.Spec.ProviderID = &providerID
			intelmachine.Annotations[infrastructurev1alpha1.HostStateAnnotation] = infrastructurev1alpha1.HostStateActive

			rc := IntelMachineReconcilerContext{
				log:          logger,
				ctx:          ctx,
				machine:      machine,
				cluster:      cluster,
				intelMachine: intelmachine,
				intelCluster: intelcluster,
			}

			requeue := reconciler.reconcileNormal(rc)

			Expect(requeue).To(BeFalse())
			Expect(intelmachine.Status.Ready).To(BeTrue())
			Expect(intelmachine.Status.Initialization.Provisioned).NotTo(BeNil())
			Expect(*intelmachine.Status.Initialization.Provisioned).To(BeTrue())
			Expect(conditions.IsTrue(intelmachine, string(infrastructurev1alpha1.BootstrapExecSucceededCondition))).To(BeTrue())
		})

		It("should not return error if IntelMachine is not found", func() {
			key := types.NamespacedName{Name: "unknown-intelmachine", Namespace: namespace}
			req := reconcile.Request{
				NamespacedName: key,
			}
			res, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))
		})

		It("should return error if IntelMachine's owner not found'", func() {
			key := types.NamespacedName{Name: intelMachineName, Namespace: namespace}
			im := &infrastructurev1alpha1.IntelMachine{}
			Expect(fakeClient.Get(ctx, key, im)).To(Succeed())
			im.OwnerReferences[0].Name = "bad-owner"
			Expect(fakeClient.Update(ctx, im)).To(Succeed())

			req := reconcile.Request{
				NamespacedName: key,
			}
			res, err := reconciler.Reconcile(ctx, req)
			Expect(err).To(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))
		})

		It("should not return error if IntelMachine has no owner yet", func() {
			key := types.NamespacedName{Name: intelMachineName, Namespace: namespace}
			im := &infrastructurev1alpha1.IntelMachine{}
			Expect(fakeClient.Get(ctx, key, im)).To(Succeed())
			im.OwnerReferences = nil
			Expect(fakeClient.Update(ctx, im)).To(Succeed())

			req := reconcile.Request{
				NamespacedName: key,
			}
			res, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))
		})

		It("should return error if owner Machine is missing cluster label", func() {
			key := types.NamespacedName{Name: machineName, Namespace: namespace}
			m := &clusterv1.Machine{}
			Expect(fakeClient.Get(ctx, key, m)).To(Succeed())
			Expect(m.Labels[clusterv1.ClusterNameLabel]).NotTo(BeEmpty())
			m.Labels[clusterv1.ClusterNameLabel] = ""
			Expect(fakeClient.Update(ctx, m)).To(Succeed())

			key = types.NamespacedName{Name: intelMachineName, Namespace: namespace}
			req := reconcile.Request{
				NamespacedName: key,
			}
			res, err := reconciler.Reconcile(ctx, req)
			Expect(err).To(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))
		})

		It("should not return error if Cluster has no InfrastructureRef", func() {
			key := types.NamespacedName{Name: clusterName, Namespace: namespace}
			c := &clusterv1.Cluster{}
			Expect(fakeClient.Get(ctx, key, c)).To(Succeed())
			Expect(c.Spec.InfrastructureRef.Name).NotTo(BeEmpty())
			c.Spec.InfrastructureRef = clusterv1.ContractVersionedObjectReference{}
			Expect(fakeClient.Update(ctx, c)).To(Succeed())

			key = types.NamespacedName{Name: intelMachineName, Namespace: namespace}
			req := reconcile.Request{
				NamespacedName: key,
			}

			// Reconcile() returns after paused.EnsurePausedCondition()
			res, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))

			res, err = reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))
		})

		It("should not return error if IntelCluster not available yet", func() {
			key := types.NamespacedName{Name: intelClusterName, Namespace: namespace}
			ic := &infrastructurev1alpha1.IntelCluster{}
			Expect(fakeClient.Get(ctx, key, ic)).To(Succeed())
			Expect(fakeClient.Delete(ctx, ic)).To(Succeed())
			Expect(fakeClient.Get(ctx, key, ic)).ToNot(Succeed())

			key = types.NamespacedName{Name: intelMachineName, Namespace: namespace}
			req := reconcile.Request{
				NamespacedName: key,
			}

			// Reconcile() returns after paused.EnsurePausedCondition()
			res, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))

			res, err = reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))
		})
	})
})
