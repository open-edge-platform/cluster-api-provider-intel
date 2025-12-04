// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package southboundhandler

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	cloudinit "sigs.k8s.io/cluster-api/test/infrastructure/docker/cloudinit"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	infrav1alpha2 "github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha2"
	"github.com/open-edge-platform/cluster-api-provider-intel/mocks/m_client"
	pb "github.com/open-edge-platform/cluster-api-provider-intel/pkg/api/proto"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/inventory"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/tenant"
	utils "github.com/open-edge-platform/cluster-api-provider-intel/test/utils"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	intelMachineName1 = "test-intelmachine1"
	intelMachineName2 = "test-intelmachine2"
	machineName       = "test-machine"
	clusterName       = "test-cluster"
	secretName        = "test-secret"
	secretFormat      = cloudConfigFormat
	testHostId        = "test-host-id"
	testHostUuid      = "test-host-uuid"
	ownerRefKind      = "Machine"
	bootstrapKind     = configTypeRKE2
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
	err = infrav1alpha2.AddToScheme(scheme.Scheme)
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
		name             string
		hostUuid         string
		hostId           string
		hostIdLabelValue string
		providerID       *string
		ownerRefKind     string
		namespace        string
		bootstrapKind    string
		secretName       string
		secretFormat     string
		secretValueEn    bool
		err              bool
	}{
		{
			name:             "Success",
			hostUuid:         testHostUuid,
			hostId:           testHostId,
			hostIdLabelValue: testHostId,
			providerID:       &providerID,
			ownerRefKind:     ownerRefKind,
			namespace:        "00000000-0000-0000-0000-000000000001",
			bootstrapKind:    bootstrapKind,
			secretName:       secretName,
			secretFormat:     secretFormat,
			secretValueEn:    true,
			err:              false,
		}, {
			name:             "No IntelMachine - wrong HostID",
			hostUuid:         testHostUuid,
			hostId:           "x",
			hostIdLabelValue: testHostId,
			providerID:       &providerID,
			ownerRefKind:     ownerRefKind,
			namespace:        "00000000-0000-0000-0000-000000000002",
			bootstrapKind:    bootstrapKind,
			secretName:       secretName,
			secretFormat:     secretFormat,
			secretValueEn:    true,
			err:              true,
		}, {
			name:             "No IntelMachine - wrong HostID label value",
			hostUuid:         testHostUuid,
			hostId:           testHostId,
			hostIdLabelValue: "x",
			providerID:       &providerID,
			ownerRefKind:     ownerRefKind,
			namespace:        "00000000-0000-0000-0000-000000000003",
			bootstrapKind:    bootstrapKind,
			secretName:       secretName,
			secretFormat:     secretFormat,
			secretValueEn:    true,
			err:              true,
		}, {
			name:             "No Owner",
			hostUuid:         testHostUuid,
			hostId:           testHostId,
			hostIdLabelValue: testHostId,
			providerID:       &providerID,
			ownerRefKind:     "x",
			namespace:        "00000000-0000-0000-0000-000000000004",
			bootstrapKind:    bootstrapKind,
			secretName:       secretName,
			secretFormat:     secretFormat,
			secretValueEn:    true,
			err:              true,
		}, {
			name:             "No ProviderID",
			hostUuid:         testHostUuid,
			hostId:           testHostId,
			hostIdLabelValue: testHostId,
			providerID:       nil,
			ownerRefKind:     ownerRefKind,
			namespace:        "00000000-0000-0000-0000-000000000005",
			bootstrapKind:    bootstrapKind,
			secretName:       secretName,
			secretFormat:     secretFormat,
			secretValueEn:    true,
			err:              true,
		}, {
			name:             "No Data Secret",
			hostUuid:         testHostUuid,
			hostId:           testHostId,
			hostIdLabelValue: testHostId,
			providerID:       &providerID,
			ownerRefKind:     ownerRefKind,
			namespace:        "00000000-0000-0000-0000-000000000006",
			bootstrapKind:    bootstrapKind,
			secretName:       "",
			secretFormat:     secretFormat,
			secretValueEn:    true,
			err:              true,
		}, {
			name:             "No Bootstrap Secret",
			hostUuid:         testHostUuid,
			hostId:           testHostId,
			hostIdLabelValue: testHostId,
			providerID:       &providerID,
			ownerRefKind:     ownerRefKind,
			namespace:        "00000000-0000-0000-0000-000000000007",
			bootstrapKind:    bootstrapKind,
			secretName:       "x",
			secretFormat:     secretFormat,
			secretValueEn:    true,
			err:              true,
		}, {
			// This is a special case where the secret format is not specified
			// and the secret value is not empty. The handler should assume
			// that the secret format is cloud-config.
			name:             "No Secret Format",
			hostUuid:         testHostUuid,
			hostId:           testHostId,
			hostIdLabelValue: testHostId,
			providerID:       &providerID,
			ownerRefKind:     ownerRefKind,
			namespace:        "00000000-0000-0000-0000-000000000008",
			bootstrapKind:    bootstrapKind,
			secretName:       secretName,
			secretFormat:     "",
			secretValueEn:    true,
			err:              false,
		}, {
			name:             "Unknown Secret Format",
			hostUuid:         testHostUuid,
			hostId:           testHostId,
			hostIdLabelValue: testHostId,
			providerID:       &providerID,
			ownerRefKind:     ownerRefKind,
			namespace:        "00000000-0000-0000-0000-000000000009",
			bootstrapKind:    bootstrapKind,
			secretName:       secretName,
			secretFormat:     "x",
			secretValueEn:    true,
			err:              true,
		}, {
			name:             "No Secret Value",
			hostUuid:         testHostUuid,
			hostId:           testHostId,
			hostIdLabelValue: testHostId,
			providerID:       &providerID,
			ownerRefKind:     ownerRefKind,
			namespace:        "00000000-0000-0000-0000-000000000010",
			bootstrapKind:    bootstrapKind,
			secretName:       secretName,
			secretFormat:     secretFormat,
			secretValueEn:    false,
			err:              true,
		}, {
			name:             "Invalid Bootstrap Kind",
			hostUuid:         testHostUuid,
			hostId:           testHostId,
			hostIdLabelValue: testHostId,
			providerID:       &providerID,
			ownerRefKind:     ownerRefKind,
			namespace:        "00000000-0000-0000-0000-000000000011",
			bootstrapKind:    "x",
			secretName:       secretName,
			secretFormat:     secretFormat,
			secretValueEn:    true,
			err:              true,
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
			intelmachine := utils.NewIntelMachine(tc.namespace, intelMachineName1, machine)
			intelmachine.Spec.HostId = tc.hostId
			intelmachine.Spec.ProviderID = tc.providerID
			intelmachine.ObjectMeta.Labels[infrav1alpha2.HostIdKey] = tc.hostIdLabelValue
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

			mockInventoryClient := &m_client.MockTenantAwareInventoryClient{}
			mockInventoryClient.On("GetHostByUUID", mock.Anything, tc.namespace, tc.hostUuid).Return(&computev1.HostResource{
				ResourceId: tc.hostId,
			}, nil)
			testHandler := &Handler{client: k8sClient, inventoryClient: &inventory.InventoryClient{Client: mockInventoryClient}}

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
				installCmd, uninstallCmd, resp, err := testHandler.Register(ctx, tc.hostUuid)
				assert.NoError(t, err)
				assert.Equal(t, pb.RegisterClusterResponse_SUCCESS, resp)
				assert.NotEmpty(t, installCmd)
				assert.NotEmpty(t, uninstallCmd)
			} else {
				_, _, _, err := testHandler.Register(ctx, tc.hostUuid)
				assert.Error(t, err)
			}
		})
	}
}

func FuzzHandlerRegister(f *testing.F) {
	projectId := "00000000-0000-0000-0000-000000000100"
	f.Add("abc")
	f.Fuzz(func(t *testing.T, hostId string) {

		// Create Machine
		machine := utils.NewMachine(projectId, clusterName, machineName, bootstrapKind)
		secretName := secretName
		machine.Spec.Bootstrap.DataSecretName = &secretName

		// Create IntelMachine
		intelmachine := utils.NewIntelMachine(projectId, intelMachineName1, machine)
		intelmachine.Spec.HostId = hostId
		intelmachine.Spec.ProviderID = &providerID
		intelmachine.ObjectMeta.Labels[infrav1alpha2.HostIdKey] = hostId

		// Create Secret
		secret := utils.NewRKE2BootstrapSecret(projectId, secretName)

		mockInventoryClient := &m_client.MockTenantAwareInventoryClient{}
		mockInventoryClient.On("GetHostByUUID", mock.Anything, projectId, hostId).Return(&computev1.HostResource{
			ResourceId: hostId,
		}, nil)
		testHandler := &Handler{
			client:          k8sClient,
			inventoryClient: &inventory.InventoryClient{Client: mockInventoryClient},
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

		_, _, _, _ = testHandler.Register(ctx, hostId)
	})
}

func TestHandler_UpdateStatus_MachineReady(t *testing.T) {
	projectId := "00000000-0000-0000-0000-000000000200"

	// Create Machine
	machine := utils.NewMachine(projectId, clusterName, machineName, bootstrapKind)
	secretName := secretName
	machine.Spec.Bootstrap.DataSecretName = &secretName

	// Create IntelMachine
	intelmachine := utils.NewIntelMachine(projectId, intelMachineName1, machine)
	intelmachine.Spec.HostId = testHostId
	intelmachine.ObjectMeta.Labels[infrav1alpha2.HostIdKey] = testHostId
	intelmachine.Spec.ProviderID = &providerID

	// Set up fake dynamic client
	mockInventoryClient := &m_client.MockTenantAwareInventoryClient{}
	mockInventoryClient.On("GetHostByUUID", mock.Anything, projectId, testHostUuid).Return(&computev1.HostResource{
		ResourceId: testHostId,
	}, nil)
	testHandler := &Handler{client: k8sClient, inventoryClient: &inventory.InventoryClient{Client: mockInventoryClient}}

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
			expectedHostState: infrav1alpha2.HostStateInactive,
		},
		{
			name:              "Test REGISTERING status",
			status:            pb.UpdateClusterStatusRequest_REGISTERING,
			expectedAction:    pb.UpdateClusterStatusResponse_NONE,
			expectedHostState: infrav1alpha2.HostStateInProgress,
		},
		{
			name:              "Test INSTALL_IN_PROGRESS status",
			status:            pb.UpdateClusterStatusRequest_INSTALL_IN_PROGRESS,
			expectedAction:    pb.UpdateClusterStatusResponse_NONE,
			expectedHostState: infrav1alpha2.HostStateInProgress,
		},
		{
			name:              "Test ACTIVE status",
			status:            pb.UpdateClusterStatusRequest_ACTIVE,
			expectedAction:    pb.UpdateClusterStatusResponse_NONE,
			expectedHostState: infrav1alpha2.HostStateActive,
		},
		{
			name:              "Test DEREGISTERING status",
			status:            pb.UpdateClusterStatusRequest_DEREGISTERING,
			expectedAction:    pb.UpdateClusterStatusResponse_NONE,
			expectedHostState: infrav1alpha2.HostStateInProgress,
		},
		{
			name:              "Test UNINSTALL_IN_PROGRESS",
			status:            pb.UpdateClusterStatusRequest_UNINSTALL_IN_PROGRESS,
			expectedAction:    pb.UpdateClusterStatusResponse_NONE,
			expectedHostState: infrav1alpha2.HostStateInProgress,
		},
		{
			name:              "Test ERROR status",
			status:            pb.UpdateClusterStatusRequest_ERROR,
			expectedAction:    pb.UpdateClusterStatusResponse_NONE,
			expectedHostState: infrav1alpha2.HostStateError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actionReq, err := testHandler.UpdateStatus(ctx, testHostUuid, tc.status)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedAction, actionReq)

			// Check that IntelMachine has been updated with the correct host state
			im, err := testHandler.getIntelMachine(ctx, testHandler.client, projectId, testHostId)
			assert.NoError(t, err)
			hostStatus, ok := im.Annotations[infrav1alpha2.HostStateAnnotation]
			assert.True(t, ok)
			assert.Equal(t, tc.expectedHostState, hostStatus)

			updatedIntelMachine := &infrav1alpha2.IntelMachine{}
			err = k8sClient.Get(ctx, client.ObjectKey{
				Namespace: projectId,
				Name:      intelMachineName1,
			}, updatedIntelMachine)
			assert.NoError(t, err)

			hostStatus, ok = updatedIntelMachine.Annotations[infrav1alpha2.HostStateAnnotation]
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
	intelmachine := utils.NewIntelMachine(projectId, intelMachineName1, machine)
	assert.True(t, controllerutil.AddFinalizer(intelmachine, infrav1alpha2.HostCleanupFinalizer))
	assert.True(t, controllerutil.ContainsFinalizer(intelmachine, infrav1alpha2.HostCleanupFinalizer))
	intelmachine.Spec.HostId = testHostId
	intelmachine.ObjectMeta.Labels[infrav1alpha2.HostIdKey] = testHostId
	intelmachine.Status.Ready = false

	// Set up fake dynamic client
	mockInventoryClient := &m_client.MockTenantAwareInventoryClient{}
	mockInventoryClient.On("GetHostByUUID", mock.Anything, projectId, testHostUuid).Return(&computev1.HostResource{
		ResourceId: testHostId,
	}, nil)
	testHandler := &Handler{client: k8sClient, inventoryClient: &inventory.InventoryClient{Client: mockInventoryClient}}

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
			expectedHostState:  infrav1alpha2.HostStateActive,
			expectedFinalizers: []string{infrav1alpha2.HostCleanupFinalizer},
			stillExists:        true,
		},
		{
			name:               "Remove finalizer after host is deregistered",
			status:             pb.UpdateClusterStatusRequest_INACTIVE,
			expectedAction:     pb.UpdateClusterStatusResponse_NONE,
			expectedHostState:  "",
			expectedFinalizers: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actionReq, err := testHandler.UpdateStatus(ctx, testHostUuid, tc.status)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedAction, actionReq)

			// Check that IntelMachine has been updated with the correct host state
			im, err := testHandler.getIntelMachine(ctx, testHandler.client, projectId, testHostId)
			assert.NoError(t, err)

			if tc.stillExists {
				hostStatus, ok := im.Annotations[infrav1alpha2.HostStateAnnotation]
				assert.True(t, ok)
				assert.Equal(t, tc.expectedHostState, hostStatus)
				assert.Equal(t, tc.expectedFinalizers, im.Finalizers)
			}
		})
	}
}

func TestHandler_UpdateStatus_Error(t *testing.T) {
	secretName := secretName
	testError := fmt.Errorf("test error")

	mockInventoryClient := m_client.NewMockTenantAwareInventoryClient(t)
	testHandler := &Handler{client: k8sClient, inventoryClient: &inventory.InventoryClient{Client: mockInventoryClient}}

	cases := []struct {
		name             string
		namespace        string
		hostId           string
		hostIdLabelValue string
		expectedErr      error
		duplicateMachine bool
		mocks            func() []*mock.Call
	}{
		{
			name:             "failed to get host id",
			namespace:        "00000000-0000-0000-0000-000000000400",
			hostId:           "x",
			hostIdLabelValue: testHostId,
			expectedErr:      testError,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockInventoryClient.On("GetHostByUUID", mock.Anything, mock.Anything, testHostUuid).Return(nil, testError).Once(),
				}
			},
		},
		{
			name:             "invalid host id",
			namespace:        "00000000-0000-0000-0000-000000000401",
			hostId:           testHostId,
			hostIdLabelValue: testHostId,
			expectedErr:      fmt.Errorf("invalid label value for HostID '%s'", "invalid-host-id!"),
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockInventoryClient.On("GetHostByUUID", mock.Anything, mock.Anything, testHostUuid).Return(
						&computev1.HostResource{
							ResourceId: "invalid-host-id!",
						}, nil).Once(),
				}
			},
		},
		{
			name:             "duplicate intel machines",
			namespace:        "00000000-0000-0000-0000-000000000402",
			hostId:           testHostId,
			hostIdLabelValue: testHostId,
			expectedErr:      fmt.Errorf("duplicate IntelMachines with HostID '%s' in project '%s'", testHostId, "00000000-0000-0000-0000-000000000402"),
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockInventoryClient.On("GetHostByUUID", mock.Anything, mock.Anything, testHostUuid).Return(
						&computev1.HostResource{
							ResourceId: testHostId,
						}, nil).Once(),
				}
			},
			duplicateMachine: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			ctx := tenant.AddActiveProjectIdToContext(context.Background(), tc.namespace)
			require.NoError(t, k8sClient.Create(ctx, &corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: tc.namespace}}))

			machine := utils.NewMachine(tc.namespace, clusterName, machineName, bootstrapKind)
			machine.Spec.Bootstrap.DataSecretName = &secretName
			require.NoError(t, k8sClient.Create(ctx, machine))

			intelmachine := utils.NewIntelMachine(tc.namespace, intelMachineName1, machine)
			intelmachine.ObjectMeta.Labels[infrav1alpha2.HostIdKey] = tc.hostIdLabelValue
			intelmachine.Spec.HostId = tc.hostId
			intelmachine.Spec.ProviderID = &providerID
			require.NoError(t, k8sClient.Create(ctx, intelmachine))

			if tc.duplicateMachine {
				intelmachine := utils.NewIntelMachine(tc.namespace, intelMachineName2, machine)
				intelmachine.ObjectMeta.Labels[infrav1alpha2.HostIdKey] = tc.hostIdLabelValue
				intelmachine.Spec.HostId = tc.hostId
				intelmachine.Spec.ProviderID = &providerID
				require.NoError(t, k8sClient.Create(ctx, intelmachine))
			}

			actionReq, err := testHandler.UpdateStatus(ctx, testHostUuid, pb.UpdateClusterStatusRequest_INACTIVE)
			assert.Equal(t, pb.UpdateClusterStatusResponse_NONE, actionReq)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestHandler_UpdateStatus_UnpauseCluster(t *testing.T) {
	projectId := "00000000-0000-0000-0000-000000000500"

	// Define test Cluster
	cluster := utils.NewCluster(projectId, clusterName)
	cluster.Spec.Paused = true

	// Define test IntelMachineBinding
	machineBinding := utils.NewIntelMachineBinding(projectId, clusterName, testHostId, clusterName, "test-template")

	// Setup test handler with the controller-manager's client
	err := os.Setenv("INVENTORY_ADDRESS", "")
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.TODO())

	// Set up fake dynamic client
	mockInventoryClient := &m_client.MockTenantAwareInventoryClient{}
	mockInventoryClient.On("GetHostByUUID", mock.Anything, projectId, testHostUuid).Return(&computev1.HostResource{
		ResourceId: testHostId,
	}, nil)
	testHandler := &Handler{client: k8sClient, inventoryClient: &inventory.InventoryClient{Client: mockInventoryClient}}

	assert.NoError(t, err)
	defer cancel()

	// Create test project and resources
	ctx = tenant.AddActiveProjectIdToContext(ctx, projectId)
	assert.NoError(t, testHandler.client.Create(ctx, &corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: projectId}}))
	assert.NoError(t, testHandler.client.Create(ctx, cluster))
	assert.NoError(t, testHandler.client.Create(ctx, machineBinding))

	t.Run("Unpause cluster upon first host update request", func(t *testing.T) {
		actionReq, err := testHandler.UpdateStatus(ctx, testHostUuid, pb.UpdateClusterStatusRequest_INACTIVE)
		assert.NoError(t, err)
		assert.Equal(t, pb.UpdateClusterStatusResponse_NONE, actionReq)

		// Check that Cluster Pause flag has been updated
		updatedCluster := clusterv1.Cluster{}
		assert.Eventually(t, func() bool {
			assert.NoError(t, testHandler.client.Get(ctx, client.ObjectKey{Name: clusterName, Namespace: projectId}, &updatedCluster))
			return updatedCluster.Spec.Paused == false
		}, 3*time.Second, 10*time.Millisecond)
	})
}

func FuzzHandlerUpdateStatus(f *testing.F) {
	projectId := "00000000-0000-0000-0000-000000000600"

	f.Add("abc", int32(0))
	f.Fuzz(func(t *testing.T, hostId string, code int32) {
		// Create Machine
		machine := utils.NewMachine(projectId, clusterName, machineName, bootstrapKind)
		secretName := secretName
		machine.Spec.Bootstrap.DataSecretName = &secretName

		// Create IntelMachine
		intelmachine := utils.NewIntelMachine(projectId, intelMachineName1, machine)
		intelmachine.Spec.HostId = hostId
		intelmachine.Spec.ProviderID = &providerID
		intelmachine.ObjectMeta.Labels[infrav1alpha2.HostIdKey] = hostId

		// Set up fake dynamic client
		mockInventoryClient := &m_client.MockTenantAwareInventoryClient{}
		mockInventoryClient.On("GetHostByUUID", mock.Anything, projectId, hostId).Return(&computev1.HostResource{
			ResourceId: hostId,
		}, nil)
		testHandler := &Handler{
			client:          k8sClient,
			inventoryClient: &inventory.InventoryClient{Client: mockInventoryClient},
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

		_, _ = testHandler.UpdateStatus(ctx, hostId, pb.UpdateClusterStatusRequest_Code(code))
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
