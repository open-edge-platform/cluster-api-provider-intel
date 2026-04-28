// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	infrastructurev1alpha1 "github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha1"
	"github.com/open-edge-platform/cluster-api-provider-intel/mocks/m_inventory"
	ccgv1 "github.com/open-edge-platform/cluster-connect-gateway/api/v1alpha1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc
var inventoryClient = &m_inventory.MockInfrastructureProvider{}
var testCRDDir string

const clusterTestCRDYAML = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: clusters.cluster.x-k8s.io
spec:
  group: cluster.x-k8s.io
  names:
    kind: Cluster
    listKind: ClusterList
    plural: clusters
    singular: cluster
  scope: Namespaced
  versions:
  - name: v1beta2
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          spec:
            type: object
            x-kubernetes-preserve-unknown-fields: true
          status:
            type: object
            x-kubernetes-preserve-unknown-fields: true
    subresources:
      status: {}
`

const machineTestCRDYAML = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: machines.cluster.x-k8s.io
spec:
  group: cluster.x-k8s.io
  names:
    kind: Machine
    listKind: MachineList
    plural: machines
    singular: machine
  scope: Namespaced
  versions:
  - name: v1beta2
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          spec:
            type: object
            x-kubernetes-preserve-unknown-fields: true
          status:
            type: object
            x-kubernetes-preserve-unknown-fields: true
    subresources:
      status: {}
`

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	var err error
	testCRDDir, err = os.MkdirTemp("", "controller-crds-*")
	Expect(err).NotTo(HaveOccurred())

	clusterConnectCRD, err := os.ReadFile(filepath.Join("..", "..", "config", "crd", "deps", "cluster.edge-orchestrator.intel.com_clusterconnects.yaml"))
	Expect(err).NotTo(HaveOccurred())
	Expect(os.WriteFile(filepath.Join(testCRDDir, "clusterconnects.yaml"), clusterConnectCRD, 0o600)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(testCRDDir, "clusters.v1beta2.yaml"), []byte(clusterTestCRDYAML), 0o600)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(testCRDDir, "machines.v1beta2.yaml"), []byte(machineTestCRDYAML), 0o600)).To(Succeed())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "config", "crd", "bases"),
			testCRDDir,
		},
		ErrorIfCRDPathMissing: true,
	}

	// Prefer KUBEBUILDER_ASSETS when provided by make/envtest setup and fall back to the
	// repository-local bin directory for direct test execution.
	binaryAssetsDirectory := os.Getenv("KUBEBUILDER_ASSETS")
	if binaryAssetsDirectory == "" {
		binaryAssetsDirectory = filepath.Join("..", "..", "bin", "k8s",
			fmt.Sprintf("1.31.0-%s-%s", runtime.GOOS, runtime.GOARCH))
	} else if !filepath.IsAbs(binaryAssetsDirectory) {
		binaryAssetsDirectory = filepath.Join("..", "..", binaryAssetsDirectory)
	}
	absoluteBinaryAssetsDirectory, absErr := filepath.Abs(binaryAssetsDirectory)
	Expect(absErr).NotTo(HaveOccurred())
	Expect(os.Setenv("KUBEBUILDER_ASSETS", absoluteBinaryAssetsDirectory)).To(Succeed())
	testEnv.BinaryAssetsDirectory = absoluteBinaryAssetsDirectory

	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = infrastructurev1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = clusterv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = ccgv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&IntelMachineReconciler{
		Client:          k8sManager.GetClient(),
		Scheme:          k8sManager.GetScheme(),
		InventoryClient: inventoryClient,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	err = (&IntelClusterReconciler{
		Client:          k8sManager.GetClient(),
		Scheme:          k8sManager.GetScheme(),
		InventoryClient: inventoryClient,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
	if testCRDDir != "" {
		Expect(os.RemoveAll(testCRDDir)).To(Succeed())
	}
})
