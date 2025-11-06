// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package southboundhandler

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	cloudinit "sigs.k8s.io/cluster-api/test/infrastructure/docker/cloudinit"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	infrastructurev1alpha1 "github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha1"
	pb "github.com/open-edge-platform/cluster-api-provider-intel/pkg/api/proto"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/tenant"
	utils "github.com/open-edge-platform/cluster-api-provider-intel/test/utils"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	intelMachineName = "test-intelmachine"
	machineName      = "test-machine"
	clusterName      = "test-cluster"
	secretName       = "test-secret"
	secretFormat     = cloudConfigFormat
	nodeGUID         = "test-nodeGUID"
	ownerRefKind     = "Machine"
	bootstrapKind    = configTypeRKE2
)

var (
	providerID = "test-providerID"
)

var testEnv *envtest.Environment
var k8sClient client.Client

func TestMain(m *testing.M) {
	// Set up the test environment
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			"../../config/crd/bases",
			"../../config/crd/deps",
		},
	}

	cfg, err := testEnv.Start()
	if err != nil {
		log.Fatal().Msgf("Failed to start test environment: %v", err)
	}

	// Add your schemes here
	err = scheme.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatal().Msgf("Failed to add scheme: %v", err)
	}
	err = infrastructurev1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatal().Msgf("Failed to add scheme: %v", err)
	}
	err = clusterv1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatal().Msgf("Failed to add scheme: %v", err)
	}

	// Create the controller-runtime client
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		log.Fatal().Msgf("Failed to create client: %v", err)
	}

	code := m.Run()

	// Tear down the test environment
	err = testEnv.Stop()
	if err != nil {
		log.Fatal().Msgf("Failed to stop test environment: %v", err)
	}

	os.Exit(code)
}

func TestHandler_Register(t *testing.T) {
	cases := []struct {
		name          string
		nodeGUID      string
		nodeGUIDLabel string
		providerID    *string
		ownerRefKind  string
		namespace     string
		bootstrapKind string
		secretName    string
		secretFormat  string
		secretValueEn bool
		err           bool
	}{
		{
			name:          "Success",
			nodeGUID:      nodeGUID,
			nodeGUIDLabel: nodeGUID,
			providerID:    &providerID,
			ownerRefKind:  ownerRefKind,
			namespace:     "00000000-0000-0000-0000-000000000001",
			bootstrapKind: bootstrapKind,
			secretName:    secretName,
			secretFormat:  secretFormat,
			secretValueEn: true,
			err:           false,
		}, {
			name:          "No IntelMachine - wrong NodeGUID",
			nodeGUID:      "x",
			nodeGUIDLabel: nodeGUID,
			providerID:    &providerID,
			ownerRefKind:  ownerRefKind,
			namespace:     "00000000-0000-0000-0000-000000000002",
			bootstrapKind: bootstrapKind,
			secretName:    secretName,
			secretFormat:  secretFormat,
			secretValueEn: true,
			err:           true,
		}, {
			name:          "No IntelMachine - wrong NodeGUID label",
			nodeGUID:      nodeGUID,
			nodeGUIDLabel: "x",
			providerID:    &providerID,
			ownerRefKind:  ownerRefKind,
			namespace:     "00000000-0000-0000-0000-000000000003",
			bootstrapKind: bootstrapKind,
			secretName:    secretName,
			secretFormat:  secretFormat,
			secretValueEn: true,
			err:           true,
		}, {
			name:          "No Owner",
			nodeGUID:      nodeGUID,
			nodeGUIDLabel: nodeGUID,
			providerID:    &providerID,
			ownerRefKind:  "x",
			namespace:     "00000000-0000-0000-0000-000000000004",
			bootstrapKind: bootstrapKind,
			secretName:    secretName,
			secretFormat:  secretFormat,
			secretValueEn: true,
			err:           true,
		}, {
			name:          "No ProviderID",
			nodeGUID:      nodeGUID,
			nodeGUIDLabel: nodeGUID,
			providerID:    nil,
			ownerRefKind:  ownerRefKind,
			namespace:     "00000000-0000-0000-0000-000000000005",
			bootstrapKind: bootstrapKind,
			secretName:    secretName,
			secretFormat:  secretFormat,
			secretValueEn: true,
			err:           true,
		}, {
			name:          "No Data Secret",
			nodeGUID:      nodeGUID,
			nodeGUIDLabel: nodeGUID,
			providerID:    &providerID,
			ownerRefKind:  ownerRefKind,
			namespace:     "00000000-0000-0000-0000-000000000006",
			bootstrapKind: bootstrapKind,
			secretName:    "",
			secretFormat:  secretFormat,
			secretValueEn: true,
			err:           true,
		}, {
			name:          "No Bootstrap Secret",
			nodeGUID:      nodeGUID,
			nodeGUIDLabel: nodeGUID,
			providerID:    &providerID,
			ownerRefKind:  ownerRefKind,
			namespace:     "00000000-0000-0000-0000-000000000007",
			bootstrapKind: bootstrapKind,
			secretName:    "x",
			secretFormat:  secretFormat,
			secretValueEn: true,
			err:           true,
		}, {
			// This is a special case where the secret format is not specified
			// and the secret value is not empty. The handler should assume
			// that the secret format is cloud-config.
			name:          "No Secret Format",
			nodeGUID:      nodeGUID,
			nodeGUIDLabel: nodeGUID,
			providerID:    &providerID,
			ownerRefKind:  ownerRefKind,
			namespace:     "00000000-0000-0000-0000-000000000008",
			bootstrapKind: bootstrapKind,
			secretName:    secretName,
			secretFormat:  "",
			secretValueEn: true,
			err:           false,
		}, {
			name:          "Unknown Secret Format",
			nodeGUID:      nodeGUID,
			nodeGUIDLabel: nodeGUID,
			providerID:    &providerID,
			ownerRefKind:  ownerRefKind,
			namespace:     "00000000-0000-0000-0000-000000000009",
			bootstrapKind: bootstrapKind,
			secretName:    secretName,
			secretFormat:  "x",
			secretValueEn: true,
			err:           true,
		}, {
			name:          "No Secret Value",
			nodeGUID:      nodeGUID,
			nodeGUIDLabel: nodeGUID,
			providerID:    &providerID,
			ownerRefKind:  ownerRefKind,
			namespace:     "00000000-0000-0000-0000-000000000010",
			bootstrapKind: bootstrapKind,
			secretName:    secretName,
			secretFormat:  secretFormat,
			secretValueEn: false,
			err:           true,
		}, {
			name:          "Invalid Bootstrap Kind",
			nodeGUID:      nodeGUID,
			nodeGUIDLabel: nodeGUID,
			providerID:    &providerID,
			ownerRefKind:  ownerRefKind,
			namespace:     "00000000-0000-0000-0000-000000000011",
			bootstrapKind: "x",
			secretName:    secretName,
			secretFormat:  secretFormat,
			secretValueEn: true,
			err:           true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Create Machine
			machine := utils.NewMachine(tc.namespace, clusterName, machineName, bootstrapKind)
			if tc.secretName == "" {
				machine.Spec.Bootstrap.DataSecretName = nil
			} else {
				machine.Spec.Bootstrap.DataSecretName = &tc.secretName
			}
			machine.Namespace = tc.namespace
			machine.Spec.Bootstrap.ConfigRef.Kind = tc.bootstrapKind

			// Create IntelMachine
			intelmachine := utils.NewIntelMachine(tc.namespace, intelMachineName, machine)
			intelmachine.Spec.NodeGUID = tc.nodeGUID
			intelmachine.Spec.ProviderID = tc.providerID
			intelmachine.ObjectMeta.Labels[infrastructurev1alpha1.NodeGUIDKey] = tc.nodeGUIDLabel
			intelmachine.OwnerReferences[0].Kind = tc.ownerRefKind

			// Create Secret
			secret := utils.NewRKE2BootstrapSecret(tc.namespace, secretName)
			if tc.secretFormat == "" {
				delete(secret.Data, "format")
			} else {
				secret.Data["format"] = []byte(tc.secretFormat)
			}
			if !tc.secretValueEn {
				delete(secret.Data, "value")
			}

			testHandler := &Handler{
				client: k8sClient,
			}

			// Add Project ID to context
			ctx := tenant.AddActiveProjectIdToContext(context.Background(), tc.namespace)

			// Create the namespace
			ns := &corev1.Namespace{
				ObjectMeta: v1.ObjectMeta{
					Name: tc.namespace,
				},
			}
			err := k8sClient.Create(ctx, ns)
			assert.NoError(t, err)

			err = k8sClient.Create(ctx, machine)
			assert.NoError(t, err)

			err = k8sClient.Create(ctx, intelmachine)
			assert.NoError(t, err)

			err = k8sClient.Create(ctx, secret)
			assert.NoError(t, err)

			if !tc.err {
				installCmd, uninstallCmd, resp, err := testHandler.Register(ctx, tc.nodeGUID)
				assert.NoError(t, err)
				assert.Equal(t, pb.RegisterClusterResponse_SUCCESS, resp)
				assert.NotEmpty(t, installCmd)
				assert.NotEmpty(t, uninstallCmd)
			} else {
				_, _, _, err := testHandler.Register(ctx, tc.nodeGUID)
				assert.Error(t, err)
			}
		})
	}
}

func FuzzHandlerRegister(f *testing.F) {
	projectId := "00000000-0000-0000-0000-000000000100"
	f.Add("abc")
	f.Fuzz(func(t *testing.T, nodeGUID string) {

		// Create Machine
		machine := utils.NewMachine(projectId, clusterName, machineName, bootstrapKind)
		secretName := secretName
		machine.Spec.Bootstrap.DataSecretName = &secretName

		// Create IntelMachine
		intelmachine := utils.NewIntelMachine(projectId, intelMachineName, machine)
		intelmachine.Spec.NodeGUID = nodeGUID
		intelmachine.Spec.ProviderID = &providerID
		intelmachine.ObjectMeta.Labels[infrastructurev1alpha1.NodeGUIDKey] = nodeGUID

		// Create Secret
		secret := utils.NewRKE2BootstrapSecret(projectId, secretName)

		testHandler := &Handler{
			client: k8sClient,
		}

		// Add Project ID to context
		ctx := tenant.AddActiveProjectIdToContext(context.Background(), projectId)

		// Create the namespace
		ns := &corev1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name: projectId,
			},
		}
		err := k8sClient.Create(ctx, ns)
		assert.NoError(t, err)

		err = k8sClient.Create(ctx, machine)
		assert.NoError(t, err)

		err = k8sClient.Create(ctx, intelmachine)
		assert.NoError(t, err)

		err = k8sClient.Create(ctx, secret)
		assert.NoError(t, err)

		_, _, _, _ = testHandler.Register(ctx, nodeGUID)
	})
}

func TestHandler_UpdateStatus_MachineReady(t *testing.T) {
	projectId := "00000000-0000-0000-0000-000000000200"

	// Create Machine
	machine := utils.NewMachine(projectId, clusterName, machineName, bootstrapKind)
	secretName := secretName
	machine.Spec.Bootstrap.DataSecretName = &secretName

	// Create IntelMachine
	intelmachine := utils.NewIntelMachine(projectId, intelMachineName, machine)
	intelmachine.Spec.NodeGUID = nodeGUID
	intelmachine.ObjectMeta.Labels[infrastructurev1alpha1.NodeGUIDKey] = nodeGUID
	intelmachine.Spec.ProviderID = &providerID

	// Set up fake dynamic client
	testHandler := &Handler{
		client: k8sClient,
	}

	// Add Project ID to context
	ctx := tenant.AddActiveProjectIdToContext(context.Background(), projectId)

	// Create the namespace
	ns := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: projectId,
		},
	}
	err := k8sClient.Create(ctx, ns)
	assert.NoError(t, err)

	err = k8sClient.Create(ctx, machine)
	assert.NoError(t, err)

	err = k8sClient.Create(ctx, intelmachine)
	assert.NoError(t, err)

	cases := []struct {
		name              string
		status            pb.UpdateClusterStatusRequest_Code
		expectedAction    pb.UpdateClusterStatusResponse_ActionRequest
		expectedHostState string
	}{
		{
			name:              "Test INACTIVE status",
			status:            pb.UpdateClusterStatusRequest_INACTIVE,
			expectedAction:    pb.UpdateClusterStatusResponse_REGISTER,
			expectedHostState: infrastructurev1alpha1.HostStateInactive,
		},
		{
			name:              "Test REGISTERING status",
			status:            pb.UpdateClusterStatusRequest_REGISTERING,
			expectedAction:    pb.UpdateClusterStatusResponse_NONE,
			expectedHostState: infrastructurev1alpha1.HostStateInProgress,
		},
		{
			name:              "Test INSTALL_IN_PROGRESS status",
			status:            pb.UpdateClusterStatusRequest_INSTALL_IN_PROGRESS,
			expectedAction:    pb.UpdateClusterStatusResponse_NONE,
			expectedHostState: infrastructurev1alpha1.HostStateInProgress,
		},
		{
			name:              "Test ACTIVE status",
			status:            pb.UpdateClusterStatusRequest_ACTIVE,
			expectedAction:    pb.UpdateClusterStatusResponse_NONE,
			expectedHostState: infrastructurev1alpha1.HostStateActive,
		},
		{
			name:              "Test DEREGISTERING status",
			status:            pb.UpdateClusterStatusRequest_DEREGISTERING,
			expectedAction:    pb.UpdateClusterStatusResponse_NONE,
			expectedHostState: infrastructurev1alpha1.HostStateInProgress,
		},
		{
			name:              "Test UNINSTALL_IN_PROGRESS",
			status:            pb.UpdateClusterStatusRequest_UNINSTALL_IN_PROGRESS,
			expectedAction:    pb.UpdateClusterStatusResponse_NONE,
			expectedHostState: infrastructurev1alpha1.HostStateInProgress,
		},
		{
			name:              "Test ERROR status",
			status:            pb.UpdateClusterStatusRequest_ERROR,
			expectedAction:    pb.UpdateClusterStatusResponse_NONE,
			expectedHostState: infrastructurev1alpha1.HostStateError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actionReq, err := testHandler.UpdateStatus(ctx, nodeGUID, tc.status)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedAction, actionReq)

			// Check that IntelMachine has been updated with the correct host state
			im, err := testHandler.getIntelMachine(ctx, testHandler.client, projectId, nodeGUID)
			assert.NoError(t, err)
			hostStatus, ok := im.Annotations[infrastructurev1alpha1.HostStateAnnotation]
			assert.True(t, ok)
			assert.Equal(t, tc.expectedHostState, hostStatus)

			updatedIntelMachine := &infrastructurev1alpha1.IntelMachine{}
			err = k8sClient.Get(ctx, client.ObjectKey{
				Namespace: projectId,
				Name:      intelMachineName,
			}, updatedIntelMachine)
			assert.NoError(t, err)

			hostStatus, ok = updatedIntelMachine.Annotations[infrastructurev1alpha1.HostStateAnnotation]
			assert.True(t, ok)
			assert.Equal(t, tc.expectedHostState, hostStatus)
		})
	}
}

func TestHandler_UpdateStatus_MachineDeleted(t *testing.T) {
	projectId := "00000000-0000-0000-0000-000000000300"

	// Create Machine
	machine := utils.NewMachine(projectId, clusterName, machineName, bootstrapKind)
	secretName := secretName
	machine.Spec.Bootstrap.DataSecretName = &secretName

	// Create IntelMachine
	intelmachine := utils.NewIntelMachine(projectId, intelMachineName, machine)
	assert.True(t, controllerutil.AddFinalizer(intelmachine, infrastructurev1alpha1.HostCleanupFinalizer))
	assert.True(t, controllerutil.ContainsFinalizer(intelmachine, infrastructurev1alpha1.HostCleanupFinalizer))
	intelmachine.Spec.NodeGUID = nodeGUID
	intelmachine.ObjectMeta.Labels[infrastructurev1alpha1.NodeGUIDKey] = nodeGUID
	intelmachine.Status.Ready = false

	testHandler := &Handler{
		client: k8sClient,
	}

	// Add Project ID to context
	ctx := tenant.AddActiveProjectIdToContext(context.Background(), projectId)

	// Create the namespace
	ns := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: projectId,
		},
	}
	err := k8sClient.Create(ctx, ns)
	assert.NoError(t, err)

	err = k8sClient.Create(ctx, machine)
	assert.NoError(t, err)

	err = k8sClient.Create(ctx, intelmachine)
	assert.NoError(t, err)

	// Delete the IntelMachine
	err = k8sClient.Delete(ctx, intelmachine)
	assert.NoError(t, err)

	cases := []struct {
		name               string
		status             pb.UpdateClusterStatusRequest_Code
		expectedAction     pb.UpdateClusterStatusResponse_ActionRequest
		expectedHostState  string
		expectedFinalizers []string
		stillExists        bool
	}{
		{
			name:               "Deregister host when intelmachine is being deleted",
			status:             pb.UpdateClusterStatusRequest_ACTIVE,
			expectedAction:     pb.UpdateClusterStatusResponse_DEREGISTER,
			expectedHostState:  infrastructurev1alpha1.HostStateActive,
			expectedFinalizers: []string{infrastructurev1alpha1.HostCleanupFinalizer},
			stillExists:        true,
		},
		{
			name:               "Remove finalizer after host is deregistered",
			status:             pb.UpdateClusterStatusRequest_INACTIVE,
			expectedAction:     pb.UpdateClusterStatusResponse_NONE,
			expectedHostState:  "",
			expectedFinalizers: nil,
			stillExists:        false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actionReq, err := testHandler.UpdateStatus(ctx, nodeGUID, tc.status)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedAction, actionReq)

			// Check that IntelMachine has been updated with the correct host state
			im, err := testHandler.getIntelMachine(ctx, testHandler.client, projectId, nodeGUID)
			assert.NoError(t, err)
			if tc.stillExists {
				hostStatus, ok := im.Annotations[infrastructurev1alpha1.HostStateAnnotation]
				assert.True(t, ok)
				assert.Equal(t, tc.expectedHostState, hostStatus)
				assert.Equal(t, tc.expectedFinalizers, im.Finalizers)
			} else {
				assert.Nil(t, im)
			}
		})
	}
}

func TestHandler_UpdateStatus_Error(t *testing.T) {
	cases := []struct {
		name          string
		namespace     string
		nodeGUID      string
		nodeGUIDLabel string
		expectError   bool
	}{
		{
			name:          "Wrong NodeGUID",
			namespace:     "00000000-0000-0000-0000-000000000400",
			nodeGUID:      "x",
			nodeGUIDLabel: nodeGUID,
			expectError:   true,
		}, {
			name:          "Wrong NodeGUID label",
			namespace:     "00000000-0000-0000-0000-000000000401",
			nodeGUID:      nodeGUID,
			nodeGUIDLabel: "x",
			expectError:   true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Create Machine
			machine := utils.NewMachine(tc.namespace, clusterName, machineName, bootstrapKind)
			secretName := secretName
			machine.Spec.Bootstrap.DataSecretName = &secretName

			// Create IntelMachine
			intelmachine := utils.NewIntelMachine(tc.namespace, intelMachineName, machine)
			intelmachine.Spec.NodeGUID = tc.nodeGUID
			intelmachine.ObjectMeta.Labels[infrastructurev1alpha1.NodeGUIDKey] = tc.nodeGUIDLabel
			intelmachine.Spec.ProviderID = &providerID

			// Set up fake dynamic client
			testHandler := &Handler{
				client: k8sClient,
			}

			// Add Project ID to context
			ctx := tenant.AddActiveProjectIdToContext(context.Background(), tc.namespace)

			// Create the namespace
			ns := &corev1.Namespace{
				ObjectMeta: v1.ObjectMeta{
					Name: tc.namespace,
				},
			}
			err := k8sClient.Create(ctx, ns)
			assert.NoError(t, err)

			err = k8sClient.Create(ctx, machine)
			assert.NoError(t, err)

			err = k8sClient.Create(ctx, intelmachine)
			assert.NoError(t, err)

			actionReq, err := testHandler.UpdateStatus(ctx, tc.nodeGUID, pb.UpdateClusterStatusRequest_INACTIVE)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, pb.UpdateClusterStatusResponse_NONE, actionReq)
		})
	}
}

func TestHandler_UpdateStatus_UnpauseCluster(t *testing.T) {
	projectId := "00000000-0000-0000-0000-000000000500"

	// Define test Cluster
	cluster := utils.NewCluster(projectId, clusterName)
	cluster.Spec.Paused = true

	// Define test IntelMachineBinding
	machineBinding := utils.NewIntelMachineBinding(projectId, clusterName, nodeGUID, clusterName, "test-template")

	// Setup test handler with the controller-manager's client
	err := os.Setenv("INVENTORY_ADDRESS", "")
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.TODO())
	testHandler, err := NewHandler(ctx, testEnv.Config)
	assert.NoError(t, err)
	defer cancel()

	// Create test project and resources
	ctx = tenant.AddActiveProjectIdToContext(ctx, projectId)
	assert.NoError(t, testHandler.client.Create(ctx, &corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: projectId}}))
	assert.NoError(t, testHandler.client.Create(ctx, cluster))
	assert.NoError(t, testHandler.client.Create(ctx, machineBinding))

	t.Run("Unpause cluster upon first host update request", func(t *testing.T) {
		actionReq, err := testHandler.UpdateStatus(ctx, nodeGUID, pb.UpdateClusterStatusRequest_INACTIVE)
		assert.NoError(t, err)
		assert.Equal(t, pb.UpdateClusterStatusResponse_NONE, actionReq)

		// Check that Cluster Pause flag has been updated
		updatedCluster := clusterv1.Cluster{}
		assert.Eventually(t, func() bool {
			err = testHandler.client.Get(ctx, client.ObjectKey{Name: clusterName, Namespace: projectId},
				&updatedCluster)
			assert.NoError(t, err)
			return updatedCluster.Spec.Paused == false
		}, 3*time.Second, 10*time.Millisecond)
	})
}

func FuzzHandlerUpdateStatus(f *testing.F) {
	projectId := "00000000-0000-0000-0000-000000000600"

	f.Add("abc", int32(0))
	f.Fuzz(func(t *testing.T, nodeGUID string, code int32) {
		// Create Machine
		machine := utils.NewMachine(projectId, clusterName, machineName, bootstrapKind)
		secretName := secretName
		machine.Spec.Bootstrap.DataSecretName = &secretName

		// Create IntelMachine
		intelmachine := utils.NewIntelMachine(projectId, intelMachineName, machine)
		intelmachine.Spec.NodeGUID = nodeGUID
		intelmachine.Spec.ProviderID = &providerID
		intelmachine.ObjectMeta.Labels[infrastructurev1alpha1.NodeGUIDKey] = nodeGUID

		// Set up fake dynamic client
		testHandler := &Handler{
			client: k8sClient,
		}

		// Add Project ID to context
		ctx := tenant.AddActiveProjectIdToContext(context.Background(), projectId)

		// Create the namespace
		ns := &corev1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name: projectId,
			},
		}
		err := k8sClient.Create(ctx, ns)
		assert.NoError(t, err)

		err = k8sClient.Create(ctx, machine)
		assert.NoError(t, err)

		err = k8sClient.Create(ctx, intelmachine)
		assert.NoError(t, err)

		_, _ = testHandler.UpdateStatus(ctx, nodeGUID, pb.UpdateClusterStatusRequest_Code(code))
	})
}

// Currently this function just tests that the commands observed when parsing the
// RKE2 / K3S Bootstrap cloud-init script are translated correctly to bash.
func TestHandler_GetCommand(t *testing.T) {
	cases := []struct {
		name            string
		inputCommand    cloudinit.Cmd
		expectedCommand string
	}{
		{
			name: "Test mkdir command -- concatenate command and args",
			inputCommand: cloudinit.Cmd{
				Cmd:  "mkdir",
				Args: []string{"-p", "/etc/test"},
			},
			expectedCommand: "mkdir -p /etc/test",
		},
		{
			name: "Test chmod command -- concatenate command and args",
			inputCommand: cloudinit.Cmd{
				Cmd:  "chmod",
				Args: []string{"0640", "/etc/test/test.yaml"},
			},
			expectedCommand: "chmod 0640 /etc/test/test.yaml",
		},
		{
			name: "Test shell run command -- strip off /bin/sh -c",
			inputCommand: cloudinit.Cmd{
				Cmd:  "/bin/sh",
				Args: []string{"-c", "curl -sfL https://get.rke2.io | INSTALL_RKE2_VERSION=v1.30.6+rke2r1 sh -s - server"},
			},
			expectedCommand: "curl -sfL https://get.rke2.io | INSTALL_RKE2_VERSION=v1.30.6+rke2r1 sh -s - server",
		},
		{
			name: "Test file write command -- convert from cat to echo of base64-encoded string",
			inputCommand: cloudinit.Cmd{
				Cmd:   "/bin/sh",
				Args:  []string{"-c", "cat > /etc/test/test.yaml /dev/stdin"},
				Stdin: "file contents",
			},
			expectedCommand: "echo ZmlsZSBjb250ZW50cw== | base64 -d > /etc/test/test.yaml",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			command, err := getCommand(tc.inputCommand)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedCommand, command)
		})
	}

}

func Test_RKE2ExtractBootstrapScript(t *testing.T) {
	projectId := "00000000-0000-0000-0000-00000000600"
	// Create Secret
	secret := utils.NewRKE2BootstrapSecret(projectId, secretName)

	bs, err := extractBootstrapScript(secret, configTypeRKE2, "test-provider-id")
	assert.NoError(t, err)
	assert.NotEmpty(t, bs)

	// Save the result to a file for further inspection / testing
	assert.NoError(t, os.WriteFile("/tmp/rke2bootstrap.sh", []byte(bs), 0644))
}

func Test_K3SExtractBootstrapScript(t *testing.T) {
	projectId := "00000000-0000-0000-0000-00000000700"
	// Create Secret
	secret := utils.NewK3SBootstrapSecret(projectId, secretName)

	bs, err := extractBootstrapScript(secret, configTypeKThrees, "test-provider-id")
	assert.NoError(t, err)
	assert.NotEmpty(t, bs)

	// Save the result to a file for further inspection / testing
	assert.NoError(t, os.WriteFile("/tmp/k3sbootstrap.sh", []byte(bs), 0644))
}

func Test_EncodeContents(t *testing.T) {
	// Test with a simple string
	input := "Hello, World!"
	expectedOutput := "SGVsbG8sIFdvcmxkIQ=="
	output := encodeContents("/tmp/test.txt", input)
	assert.Equal(t, expectedOutput, output)

	// Extra escapes should be removed from config.toml.tmpl files as part of the encoding process.
	// Verify that \" in the string is replaced with ".
	input = "Hello, \\\"World!\\\""
	expectedOutput = "SGVsbG8sICJXb3JsZCEi" // Hello, "World!"
	output = encodeContents("/tmp/config.toml.tmpl", input)
	assert.Equal(t, expectedOutput, output)

	// config.toml.tmpl without extra escapes should be unchanged.
	input = "Hello, \"World!\""
	expectedOutput = "SGVsbG8sICJXb3JsZCEi" // Hello, "World!"
	output = encodeContents("/tmp/config.toml.tmpl", input)
	assert.Equal(t, expectedOutput, output)
}
