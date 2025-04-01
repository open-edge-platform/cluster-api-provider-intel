// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package southboundhandler

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	cloudinit "sigs.k8s.io/cluster-api/test/infrastructure/docker/cloudinit"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	infrastructurev1alpha1 "github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha1"
	pb "github.com/open-edge-platform/cluster-api-provider-intel/pkg/api/proto"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/tenant"
	utils "github.com/open-edge-platform/cluster-api-provider-intel/test/utils"
)

const (
	projectId        = "731ee171-c626-43ab-8422-38d14579f020"
	intelMachineName = "test-intelmachine"
	machineName      = "test-machine"
	clusterName      = "test-cluster"
	secretName       = "test-secret"
	secretFormat     = "cloud-config"
	nodeGUID         = "test-nodeGUID"
	ownerRefKind     = "Machine"
	bootstrapKind    = "RKE2Config"
)

var (
	providerID = "test-providerID"
)

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
			namespace:     projectId,
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
			namespace:     projectId,
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
			namespace:     projectId,
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
			namespace:     projectId,
			bootstrapKind: bootstrapKind,
			secretName:    secretName,
			secretFormat:  secretFormat,
			secretValueEn: true,
			err:           true,
		}, {
			name:          "No Machine",
			nodeGUID:      nodeGUID,
			nodeGUIDLabel: nodeGUID,
			providerID:    &providerID,
			ownerRefKind:  ownerRefKind,
			namespace:     "x",
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
			namespace:     projectId,
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
			namespace:     projectId,
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
			namespace:     projectId,
			bootstrapKind: bootstrapKind,
			secretName:    "x",
			secretFormat:  secretFormat,
			secretValueEn: true,
			err:           true,
		}, {
			name:          "No Secret Format",
			nodeGUID:      nodeGUID,
			nodeGUIDLabel: nodeGUID,
			providerID:    &providerID,
			ownerRefKind:  ownerRefKind,
			namespace:     projectId,
			bootstrapKind: bootstrapKind,
			secretName:    secretName,
			secretFormat:  "",
			secretValueEn: true,
			err:           true,
		}, {
			name:          "No Secret Format",
			nodeGUID:      nodeGUID,
			nodeGUIDLabel: nodeGUID,
			providerID:    &providerID,
			ownerRefKind:  ownerRefKind,
			namespace:     projectId,
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
			namespace:     projectId,
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
			namespace:     projectId,
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
			machine := utils.NewMachine(projectId, clusterName, machineName, bootstrapKind)
			if tc.secretName == "" {
				machine.Spec.Bootstrap.DataSecretName = nil
			} else {
				machine.Spec.Bootstrap.DataSecretName = &tc.secretName
			}
			machine.Namespace = tc.namespace
			machine.Spec.Bootstrap.ConfigRef.Kind = tc.bootstrapKind
			machineUnstructured, err := getUnstructured(machine)
			assert.NoError(t, err)

			// Create IntelMachine
			intelmachine := utils.NewIntelMachine(projectId, intelMachineName, machine)
			intelmachine.Spec.NodeGUID = tc.nodeGUID
			intelmachine.Spec.ProviderID = tc.providerID
			intelmachine.ObjectMeta.Labels[infrastructurev1alpha1.NodeGUIDKey] = tc.nodeGUIDLabel
			intelmachine.OwnerReferences[0].Kind = tc.ownerRefKind
			intelMachineUnstructured, err := getUnstructured(intelmachine)
			assert.NoError(t, err)

			// Create Secret
			secret := utils.NewBootstrapSecret(projectId, secretName)
			if tc.secretFormat == "" {
				delete(secret.Data, "format")
			} else {
				secret.Data["format"] = []byte(tc.secretFormat)
			}
			if !tc.secretValueEn {
				delete(secret.Data, "value")
			}
			secretUnstructured, err := getUnstructured(secret)
			assert.NoError(t, err)

			// Set up fake dynamic client
			testHandler := &Handler{
				client: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(),
					machineUnstructured,
					intelMachineUnstructured,
					secretUnstructured),
			}

			// Add Project ID to context
			ctx := tenant.AddActiveProjectIdToContext(context.Background(), projectId)

			if !tc.err {
				installCmd, uninstallCmd, resp, err := testHandler.Register(ctx, nodeGUID)
				assert.NoError(t, err)
				assert.Equal(t, pb.RegisterClusterResponse_SUCCESS, resp)
				assert.NotEmpty(t, installCmd)
				assert.NotEmpty(t, uninstallCmd)
			} else {
				_, _, _, err := testHandler.Register(ctx, nodeGUID)
				assert.Error(t, err)
			}
		})
	}
}

func FuzzHandlerRegister(f *testing.F) {
	f.Add("abc")
	f.Fuzz(func(t *testing.T, nodeGUID string) {
		// Create Machine
		machine := utils.NewMachine(projectId, clusterName, machineName, bootstrapKind)
		secretName := secretName
		machine.Spec.Bootstrap.DataSecretName = &secretName
		machineUnstructured, err := getUnstructured(machine)
		assert.NoError(t, err)

		// Create IntelMachine
		intelmachine := utils.NewIntelMachine(projectId, intelMachineName, machine)
		intelmachine.Spec.NodeGUID = nodeGUID
		intelmachine.Spec.ProviderID = &providerID
		intelmachine.ObjectMeta.Labels[infrastructurev1alpha1.NodeGUIDKey] = nodeGUID
		intelMachineUnstructured, err := getUnstructured(intelmachine)
		assert.NoError(t, err)

		// Create Secret
		secret := utils.NewBootstrapSecret(projectId, secretName)
		secretUnstructured, err := getUnstructured(secret)
		assert.NoError(t, err)

		// Set up fake dynamic client
		testHandler := &Handler{
			client: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(),
				machineUnstructured,
				intelMachineUnstructured,
				secretUnstructured),
		}

		// Add Project ID to context
		ctx := tenant.AddActiveProjectIdToContext(context.Background(), projectId)

		_, _, _, _ = testHandler.Register(ctx, nodeGUID)
	})
}

func TestHandler_UpdateStatus_MachineReady(t *testing.T) {
	// Create Machine
	machine := utils.NewMachine(projectId, clusterName, machineName, bootstrapKind)
	secretName := secretName
	machine.Spec.Bootstrap.DataSecretName = &secretName
	machineUnstructured, err := getUnstructured(machine)
	assert.NoError(t, err)

	// Create IntelMachine
	intelmachine := utils.NewIntelMachine(projectId, intelMachineName, machine)
	intelmachine.Spec.NodeGUID = nodeGUID
	intelmachine.ObjectMeta.Labels[infrastructurev1alpha1.NodeGUIDKey] = nodeGUID
	intelmachine.Spec.ProviderID = &providerID
	intelMachineUnstructured, err := getUnstructured(intelmachine)
	assert.NoError(t, err)

	// Set up fake dynamic client
	testHandler := &Handler{
		client: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(),
			machineUnstructured,
			intelMachineUnstructured),
	}

	// Add Project ID to context
	ctx := tenant.AddActiveProjectIdToContext(context.Background(), projectId)

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
			im, err := getIntelMachine(ctx, testHandler.client, projectId, nodeGUID)
			assert.NoError(t, err)
			hostStatus, ok := im.Annotations[infrastructurev1alpha1.HostStateAnnotation]
			assert.True(t, ok)
			assert.Equal(t, tc.expectedHostState, hostStatus)
		})
	}
}

func TestHandler_UpdateStatus_MachineDeleted(t *testing.T) {
	// Create Machine
	machine := utils.NewMachine(projectId, clusterName, machineName, bootstrapKind)
	secretName := secretName
	machine.Spec.Bootstrap.DataSecretName = &secretName
	machineUnstructured, err := getUnstructured(machine)
	assert.NoError(t, err)

	// Create IntelMachine
	intelmachine := utils.NewIntelMachine(projectId, intelMachineName, machine)
	assert.True(t, controllerutil.AddFinalizer(intelmachine, infrastructurev1alpha1.HostCleanupFinalizer))
	assert.True(t, controllerutil.ContainsFinalizer(intelmachine, infrastructurev1alpha1.HostCleanupFinalizer))
	intelmachine.Spec.NodeGUID = nodeGUID
	intelmachine.ObjectMeta.Labels[infrastructurev1alpha1.NodeGUIDKey] = nodeGUID
	intelmachine.DeletionTimestamp = &v1.Time{Time: time.Now()}
	intelmachine.Status.Ready = false
	intelMachineUnstructured, err := getUnstructured(intelmachine)
	assert.NoError(t, err)

	// Set up fake dynamic client
	testHandler := &Handler{
		client: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(),
			machineUnstructured,
			intelMachineUnstructured),
	}

	// Add Project ID to context
	ctx := tenant.AddActiveProjectIdToContext(context.Background(), projectId)

	cases := []struct {
		name               string
		status             pb.UpdateClusterStatusRequest_Code
		expectedAction     pb.UpdateClusterStatusResponse_ActionRequest
		expectedHostState  string
		expectedFinalizers []string
	}{
		{
			name:               "Deregister host when intelmachine is being deleted",
			status:             pb.UpdateClusterStatusRequest_ACTIVE,
			expectedAction:     pb.UpdateClusterStatusResponse_DEREGISTER,
			expectedHostState:  infrastructurev1alpha1.HostStateActive,
			expectedFinalizers: []string{infrastructurev1alpha1.HostCleanupFinalizer},
		},
		{
			name:               "Remove finalizer after host is deregistered",
			status:             pb.UpdateClusterStatusRequest_INACTIVE,
			expectedAction:     pb.UpdateClusterStatusResponse_NONE,
			expectedHostState:  infrastructurev1alpha1.HostStateInactive,
			expectedFinalizers: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actionReq, err := testHandler.UpdateStatus(ctx, nodeGUID, tc.status)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedAction, actionReq)

			// Check that IntelMachine has been updated with the correct host state
			im, err := getIntelMachine(ctx, testHandler.client, projectId, nodeGUID)
			assert.NoError(t, err)
			hostStatus, ok := im.Annotations[infrastructurev1alpha1.HostStateAnnotation]
			assert.True(t, ok)
			assert.Equal(t, tc.expectedHostState, hostStatus)
			assert.Equal(t, tc.expectedFinalizers, im.Finalizers)
		})
	}
}

func TestHandler_UpdateStatus_Error(t *testing.T) {
	cases := []struct {
		name          string
		nodeGUID      string
		nodeGUIDLabel string
		expectError   bool
	}{
		{
			name:          "Wrong NodeGUID",
			nodeGUID:      "x",
			nodeGUIDLabel: nodeGUID,
			expectError:   true,
		}, {
			name:          "Wrong NodeGUID label",
			nodeGUID:      nodeGUID,
			nodeGUIDLabel: "x",
			expectError:   false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Create Machine
			machine := utils.NewMachine(projectId, clusterName, machineName, bootstrapKind)
			secretName := secretName
			machine.Spec.Bootstrap.DataSecretName = &secretName
			machineUnstructured, err := getUnstructured(machine)
			assert.NoError(t, err)

			// Create IntelMachine
			intelmachine := utils.NewIntelMachine(projectId, intelMachineName, machine)
			intelmachine.Spec.NodeGUID = tc.nodeGUID
			intelmachine.ObjectMeta.Labels[infrastructurev1alpha1.NodeGUIDKey] = tc.nodeGUIDLabel
			intelmachine.Spec.ProviderID = &providerID
			intelMachineUnstructured, err := getUnstructured(intelmachine)
			assert.NoError(t, err)

			// Set up fake dynamic client
			testHandler := &Handler{
				client: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(),
					machineUnstructured,
					intelMachineUnstructured),
			}

			// Add Project ID to context
			ctx := tenant.AddActiveProjectIdToContext(context.Background(), projectId)
			actionReq, err := testHandler.UpdateStatus(ctx, nodeGUID, pb.UpdateClusterStatusRequest_INACTIVE)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, pb.UpdateClusterStatusResponse_NONE, actionReq)
		})
	}
}

func FuzzHandlerUpdateStatus(f *testing.F) {
	f.Add("abc", int32(0))
	f.Fuzz(func(t *testing.T, nodeGUID string, code int32) {
		// Create Machine
		machine := utils.NewMachine(projectId, clusterName, machineName, bootstrapKind)
		secretName := secretName
		machine.Spec.Bootstrap.DataSecretName = &secretName
		machineUnstructured, err := getUnstructured(machine)
		assert.NoError(t, err)

		// Create IntelMachine
		intelmachine := utils.NewIntelMachine(projectId, intelMachineName, machine)
		intelmachine.Spec.NodeGUID = nodeGUID
		intelmachine.Spec.ProviderID = &providerID
		intelmachine.ObjectMeta.Labels[infrastructurev1alpha1.NodeGUIDKey] = nodeGUID
		intelMachineUnstructured, err := getUnstructured(intelmachine)
		assert.NoError(t, err)

		// Set up fake dynamic client
		testHandler := &Handler{
			client: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(),
				machineUnstructured,
				intelMachineUnstructured),
		}

		// Add Project ID to context
		ctx := tenant.AddActiveProjectIdToContext(context.Background(), projectId)

		_, _ = testHandler.UpdateStatus(ctx, nodeGUID, pb.UpdateClusterStatusRequest_Code(code))
	})
}

// Currently this function just tests that the commands observed when parsing the
// RKE2 Bootstrap cloud-init script are translated correctly to bash.
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
			name: "Test file write command -- convert from cat to echo",
			inputCommand: cloudinit.Cmd{
				Cmd:   "/bin/sh",
				Args:  []string{"-c", "cat > /etc/test/test.yaml /dev/stdin"},
				Stdin: "file contents",
			},
			expectedCommand: "echo 'file contents' > /etc/test/test.yaml",
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

func Test_ExtractBootstrapScript(t *testing.T) {
	// Create Secret
	secret := utils.NewBootstrapSecret(projectId, secretName)

	bs, err := extractBootstrapScript(secret, "RKE2Config", "test-provider-id")
	assert.NoError(t, err)
	assert.NotEmpty(t, bs)

	// Save the result to a file for further inspection / testing
	assert.NoError(t, os.WriteFile("/tmp/bootstrap.sh", []byte(bs), 0644))
}

func getUnstructured(obj runtime.Object) (*unstructured.Unstructured, error) {
	result, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	unstructuredObj := &unstructured.Unstructured{}
	unstructuredObj.SetUnstructuredContent(result)
	return unstructuredObj, nil
}
