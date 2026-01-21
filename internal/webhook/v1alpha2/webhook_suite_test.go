// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha1"
	"github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha2"
	webhookv1alpha1 "github.com/open-edge-platform/cluster-api-provider-intel/internal/webhook/v1alpha1"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	ctx       context.Context
	cancel    context.CancelFunc
	k8sClient client.Client
	cfg       *rest.Config
	testEnv   *envtest.Environment
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Webhook Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	var err error
	err = v1alpha1.AddToScheme(scheme.Scheme)
	err = v1alpha2.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: false,
	}

	// Retrieve the first found binary directory to allow running tests from IDEs
	envTestBinDir := getFirstFoundEnvTestBinaryDir()
	if envTestBinDir != "" {
		testEnv.BinaryAssetsDirectory = envTestBinDir
	}

	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	// start webhook server using Manager.
	webhookInstallOptions := &testEnv.WebhookInstallOptions
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
		WebhookServer: webhook.NewServer(webhook.Options{
			Host:    webhookInstallOptions.LocalServingHost,
			Port:    webhookInstallOptions.LocalServingPort,
			CertDir: webhookInstallOptions.LocalServingCertDir,
		}),
		LeaderElection: false,
		Metrics:        metricsserver.Options{BindAddress: "0"},
	})
	Expect(err).NotTo(HaveOccurred())

	// Register v1alpha1 webhooks (hub)
	Expect(webhookv1alpha1.SetupIntelMachineWebhookWithManager(mgr)).To(Succeed())
	Expect(webhookv1alpha1.SetupIntelMachineBindingWebhookWithManager(mgr)).To(Succeed())

	// Register v1alpha2 webhooks (spoke/storage version)
	Expect(SetupIntelMachineWebhookWithManager(mgr)).To(Succeed())
	Expect(SetupIntelMachineBindingWebhookWithManager(mgr)).To(Succeed())

	// +kubebuilder:scaffold:webhook

	go func() {
		defer GinkgoRecover()
		Expect(mgr.Start(ctx)).To(Succeed())
	}()

	// wait for the webhook server to get ready.
	dialer := &net.Dialer{Timeout: time.Second}
	addrPort := fmt.Sprintf("%s:%d", webhookInstallOptions.LocalServingHost, webhookInstallOptions.LocalServingPort)
	Eventually(func() error {
		conn, err := tls.DialWithDialer(dialer, "tcp", addrPort, &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return err
		}
		return conn.Close()
	}).Should(Succeed())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	Expect(testEnv.Stop()).To(Succeed())
})

var _ = Describe("IntelMachine API version conversion", func() {
	Context("when creating IntelMachine resources", func() {
		It("should convert v1alpha1 IntelMachine to v1alpha2", func() {
			providerId := "test-provider-id"
			machine := &v1alpha1.IntelMachine{
				TypeMeta:   metav1.TypeMeta{Kind: "IntelMachine", APIVersion: v1alpha1.GroupVersion.String()},
				ObjectMeta: metav1.ObjectMeta{Name: "test-intel-machine-v1", Namespace: "default"},
				Spec:       v1alpha1.IntelMachineSpec{ProviderID: &providerId, NodeGUID: "test-node-guid"},
			}

			logf.Log.Info("Creating v1alpha1 IntelMachine", "name", machine.Name)
			Expect(k8sClient.Create(ctx, machine)).To(Succeed())

			retrievedMachine := &v1alpha2.IntelMachine{}
			Expect(k8sClient.Get(ctx, client.ObjectKey{Name: machine.Name, Namespace: machine.Namespace}, retrievedMachine)).To(Succeed())

			// Set the GVK on the retrieved object using the scheme
			gvks, _, err := scheme.Scheme.ObjectKinds(retrievedMachine)
			Expect(err).NotTo(HaveOccurred())
			Expect(gvks).To(HaveLen(1))
			Expect(gvks[0].Group).To(Equal(v1alpha2.GroupVersion.Group))
			Expect(gvks[0].Version).To(Equal(v1alpha2.GroupVersion.Version))
			Expect(gvks[0].Kind).To(Equal("IntelMachine"))

			// Verify the conversion happened correctly - check that v1alpha1.NodeGUID was converted to v1alpha2.HostID
			Expect(retrievedMachine.Spec.HostId).To(Equal("test-node-guid"))
			Expect(retrievedMachine.Spec.ProviderID).NotTo(BeNil())
			Expect(*retrievedMachine.Spec.ProviderID).To(Equal("test-provider-id"))
		})

		It("should not convert v1alpha2 IntelMachine", func() {
			providerId := "test-provider-id-v2"
			machine := &v1alpha2.IntelMachine{
				TypeMeta:   metav1.TypeMeta{Kind: "IntelMachine", APIVersion: v1alpha2.GroupVersion.String()},
				ObjectMeta: metav1.ObjectMeta{Name: "test-intel-machine-v2", Namespace: "default"},
				Spec:       v1alpha2.IntelMachineSpec{ProviderID: &providerId, HostId: "test-host-id"},
			}

			logf.Log.Info("Creating v1alpha2 IntelMachine", "name", machine.Name)
			Expect(k8sClient.Create(ctx, machine)).To(Succeed())

			retrievedMachine := &v1alpha2.IntelMachine{}
			Expect(k8sClient.Get(ctx, client.ObjectKey{Name: machine.Name, Namespace: machine.Namespace}, retrievedMachine)).To(Succeed())

			// Verify the object remains v1alpha2 with no conversion
			Expect(retrievedMachine.Spec.HostId).To(Equal("test-host-id"))
			Expect(retrievedMachine.Spec.ProviderID).NotTo(BeNil())
			Expect(*retrievedMachine.Spec.ProviderID).To(Equal("test-provider-id-v2"))
		})
	})
})

// getFirstFoundEnvTestBinaryDir locates the first binary in the specified path.
// ENVTEST-based tests depend on specific binaries, usually located in paths set by
// controller-runtime. When running tests directly (e.g., via an IDE) without using
// Makefile targets, the 'BinaryAssetsDirectory' must be explicitly configured.
//
// This function streamlines the process by finding the required binaries, similar to
// setting the 'KUBEBUILDER_ASSETS' environment variable. To ensure the binaries are
// properly set up, run 'make setup-envtest' beforehand.
func getFirstFoundEnvTestBinaryDir() string {
	basePath := filepath.Join("..", "..", "..", "bin", "k8s")
	entries, err := os.ReadDir(basePath)
	if err != nil {
		logf.Log.Error(err, "Failed to read directory", "path", basePath)
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() {
			return filepath.Join(basePath, entry.Name())
		}
	}

	return ""
}
