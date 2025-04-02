// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package inventory

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventoryv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
)

var (
	ErrGenericInventoryClientErrror = errors.New("generic test error")
)

func TestCreateWorkloadInInventory(t *testing.T) {
	mockedInventoryClient := newMockedInventoryTestClient()
	machineProvider := MachineProvider{client: InventoryClient{Client: mockedInventoryClient}}

	cases := []struct {
		name           string
		tenantId       string
		clusterName    string
		expectedOutput CreateWorkloadOutput
		mocks          func() []*mock.Call
	}{
		{
			name:        "failed with invalid cluster name",
			clusterName: "",
			tenantId:    tenantId,
			expectedOutput: CreateWorkloadOutput{
				Err: ErrInvalidClusterNameInput,
			},
		},
		{
			name:        "failed with invalid tenant id",
			clusterName: clusterName,
			tenantId:    "",
			expectedOutput: CreateWorkloadOutput{
				Err: ErrInvalidTenantIdInput,
			},
		},
		{
			name:        "successful, no errors",
			clusterName: clusterName,
			tenantId:    tenantId,
			expectedOutput: CreateWorkloadOutput{
				WorkloadId: workloadResourceId,
				Err:        nil,
			},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedInventoryClient.On("Create", mock.Anything, tenantId, mock.Anything).
						Return(&inventoryv1.Resource{
							Resource: &inventoryv1.Resource_Workload{
								Workload: &computev1.WorkloadResource{
									ResourceId: workloadResourceId,
								},
							},
						}, nil).Once(),
				}
			},
		},
		{
			name:        "failed with client error",
			clusterName: clusterName,
			tenantId:    tenantId,
			expectedOutput: CreateWorkloadOutput{
				Err: ErrFailedInventoryCreateResource,
			},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedInventoryClient.On("Create", mock.Anything, tenantId, mock.Anything).
						Return(nil, ErrGenericInventoryClientErrror).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			out := machineProvider.CreateWorkload(CreateWorkloadInput{
				TenantId:    tc.tenantId,
				ClusterName: tc.clusterName})
			require.NotNil(t, out)

			if tc.expectedOutput.Err == nil {
				require.Nil(t, out.Err)
				assert.Equal(t, tc.expectedOutput.WorkloadId, out.WorkloadId)
				return
			}

			require.NotNil(t, out.Err)
			assert.EqualError(t, tc.expectedOutput.Err, out.Err.Error())
		})
	}
}

func TestDeleteWorkloadInInventory(t *testing.T) {
	mockedInventoryClient := newMockedInventoryTestClient()
	machineProvider := MachineProvider{client: InventoryClient{Client: mockedInventoryClient}}

	cases := []struct {
		name           string
		tenantId       string
		workloadId     string
		expectedOutput DeleteWorkloadOutput
		mocks          func() []*mock.Call
	}{
		{
			name:       "failed with invalid workload id",
			workloadId: "",
			expectedOutput: DeleteWorkloadOutput{
				Err: ErrInvalidWorkloadIdInput,
			},
		},
		{
			name:       "failed with invalid tenant id",
			workloadId: workloadResourceId,
			tenantId:   "",
			expectedOutput: DeleteWorkloadOutput{
				Err: ErrInvalidTenantIdInput,
			},
		},
		{
			name:           "successful, no errors",
			workloadId:     workloadResourceId,
			tenantId:       tenantId,
			expectedOutput: DeleteWorkloadOutput{},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedInventoryClient.On("Delete", mock.Anything, tenantId, workloadResourceId).
						Return(&inventoryv1.DeleteResourceResponse{}, nil).Once(),
				}
			},
		},
		{
			name:       "failed with client error",
			workloadId: workloadResourceId,
			tenantId:   tenantId,
			expectedOutput: DeleteWorkloadOutput{
				Err: ErrFailedInventoryDeleteResource,
			},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedInventoryClient.On("Delete", mock.Anything, tenantId, workloadResourceId).
						Return(nil, ErrGenericInventoryClientErrror).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			out := machineProvider.DeleteWorkload(DeleteWorkloadInput{
				TenantId:   tc.tenantId,
				WorkloadId: tc.workloadId})
			require.NotNil(t, out)

			if tc.expectedOutput.Err != nil {
				require.NotNil(t, out.Err)
				assert.EqualError(t, tc.expectedOutput.Err, out.Err.Error())
				return
			}

			require.Nil(t, out.Err)
		})
	}
}

func TestGetWorkloadFromInventory(t *testing.T) {
	mockedInventoryClient := newMockedInventoryTestClient()
	machineProvider := MachineProvider{client: InventoryClient{Client: mockedInventoryClient}}

	cases := []struct {
		name           string
		tenantId       string
		workloadId     string
		expectedOutput GetWorkloadOutput
		mocks          func() []*mock.Call
	}{
		{
			name:       "failed with invalid workload id",
			workloadId: "",
			expectedOutput: GetWorkloadOutput{
				Workload: &Workload{},
				Err:      ErrInvalidWorkloadIdInput,
			},
		},
		{
			name:       "failed with invalid tenant id",
			workloadId: workloadResourceId,
			tenantId:   "",
			expectedOutput: GetWorkloadOutput{
				Workload: &Workload{},
				Err:      ErrInvalidTenantIdInput,
			},
		},
		{
			name:       "successful, no errors",
			workloadId: workloadResourceId,
			tenantId:   tenantId,
			expectedOutput: GetWorkloadOutput{
				Workload: &Workload{
					Id: workloadResourceId,
				},
				Err: nil,
			},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedInventoryClient.On("Get", mock.Anything, tenantId, workloadResourceId).
						Return(&inventoryv1.GetResourceResponse{
							Resource: &inventoryv1.Resource{
								Resource: &inventoryv1.Resource_Workload{
									Workload: &computev1.WorkloadResource{
										ResourceId: workloadResourceId,
									},
								},
							},
						}, nil).Once(),
				}
			},
		},
		{
			name:       "failed with client error",
			workloadId: workloadResourceId,
			tenantId:   tenantId,
			expectedOutput: GetWorkloadOutput{
				Workload: &Workload{},
				Err:      ErrFailedInventoryGetResource,
			},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedInventoryClient.On("Get", mock.Anything, tenantId, workloadResourceId).
						Return(nil, ErrGenericInventoryClientErrror).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			out := machineProvider.GetWorkload(GetWorkloadInput{
				TenantId:   tc.tenantId,
				WorkloadId: tc.workloadId})
			require.NotNil(t, out)

			if tc.expectedOutput.Err == nil {
				require.Nil(t, out.Err)
				assert.Equal(t, tc.expectedOutput.Workload, out.Workload)
				return
			}

			require.NotNil(t, out.Err)
			assert.EqualError(t, tc.expectedOutput.Err, out.Err.Error())
		})
	}
}

func TestGetInstanceByMachineId(t *testing.T) {
	mockedInventoryClient := newMockedInventoryTestClient()
	machineProvider := MachineProvider{client: InventoryClient{Client: mockedInventoryClient}}

	cases := []struct {
		name           string
		tenantId       string
		nodeUUId       string
		expectedOutput GetInstanceByMachineIdOutput
		mocks          func() []*mock.Call
	}{
		{
			name:     "successful, no errors",
			nodeUUId: nodeUUId,
			tenantId: tenantId,
			expectedOutput: GetInstanceByMachineIdOutput{
				Instance: &Instance{
					Id:        instanceResourceId,
					Os:        instanceOsName,
					SerialNo:  hostSerialNumber,
					MachineId: nodeUUId,
				},
				Err: nil,
			},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedInventoryClient.On("GetHostByUUID", mock.Anything, tenantId, nodeUUId).
						Return(&computev1.HostResource{
							ResourceId:   hostResourceId,
							SerialNumber: hostSerialNumber,
							Instance: &computev1.InstanceResource{
								ResourceId: instanceResourceId,
								CurrentOs: &osv1.OperatingSystemResource{
									Name: instanceOsName,
								},
							},
						}, nil).Once(),
				}
			},
		},
		{
			name:     "successful, empty instance",
			nodeUUId: nodeUUId,
			tenantId: tenantId,
			expectedOutput: GetInstanceByMachineIdOutput{
				Instance: &Instance{},
				Err:      ErrInvalidInstance,
			},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedInventoryClient.On("GetHostByUUID", mock.Anything, tenantId, nodeUUId).
						Return(&computev1.HostResource{
							ResourceId:   hostResourceId,
							SerialNumber: hostSerialNumber,
						}, nil).Once(),
				}
			},
		},
		{
			name:     "failed with invalid machine id",
			nodeUUId: "",
			tenantId: tenantId,
			expectedOutput: GetInstanceByMachineIdOutput{
				Instance: &Instance{},
				Err:      ErrInvalidMachineIdInput,
			},
		},
		{
			name:     "failed with invalid tenant id",
			nodeUUId: nodeUUId,
			tenantId: "",
			expectedOutput: GetInstanceByMachineIdOutput{
				Instance: &Instance{},
				Err:      ErrInvalidTenantIdInput,
			},
		},
		{
			name:     "failed with client error",
			nodeUUId: nodeUUId,
			tenantId: tenantId,
			expectedOutput: GetInstanceByMachineIdOutput{
				Instance: &Instance{},
				Err:      ErrFailedInventoryGetHostByUuid,
			},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedInventoryClient.On("GetHostByUUID", mock.Anything, tenantId, nodeUUId).
						Return(nil, ErrGenericInventoryClientErrror).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			out := machineProvider.GetInstanceByMachineId(GetInstanceByMachineIdInput{
				TenantId:  tc.tenantId,
				MachineId: tc.nodeUUId})
			require.NotNil(t, out)

			if tc.expectedOutput.Err == nil {
				require.Nil(t, out.Err)
				assert.Equal(t, tc.expectedOutput.Instance, out.Instance)
				return
			}

			require.NotNil(t, out.Err)
			assert.EqualError(t, tc.expectedOutput.Err, out.Err.Error())
		})
	}
}

func TestAddWorkloadInInventory(t *testing.T) {
	mockedInventoryClient := newMockedInventoryTestClient()
	machineProvider := MachineProvider{client: InventoryClient{Client: mockedInventoryClient}}

	cases := []struct {
		name           string
		tenantId       string
		workloadId     string
		instanceId     string
		expectedOutput AddInstanceToWorkloadOutput
		mocks          func() []*mock.Call
	}{
		{
			name:       "failed with invalid workload id",
			workloadId: "",
			expectedOutput: AddInstanceToWorkloadOutput{
				Err: ErrInvalidWorkloadIdInput,
			},
		},
		{
			name:       "failed with invalid instance id",
			workloadId: workloadResourceId,
			instanceId: "",
			expectedOutput: AddInstanceToWorkloadOutput{
				Err: ErrInvalidInstanceIdInput,
			},
		},
		{
			name:       "failed with invalid tenant id",
			workloadId: workloadResourceId,
			instanceId: instanceResourceId,
			tenantId:   "",
			expectedOutput: AddInstanceToWorkloadOutput{
				Err: ErrInvalidTenantIdInput,
			},
		},
		{
			name:       "successful, no errors",
			workloadId: workloadResourceId,
			instanceId: instanceResourceId,
			tenantId:   tenantId,
			expectedOutput: AddInstanceToWorkloadOutput{
				Err: nil,
			},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedInventoryClient.On("Create", mock.Anything, tenantId, mock.Anything).
						Return(&inventoryv1.Resource{
							Resource: &inventoryv1.Resource_WorkloadMember{
								WorkloadMember: &computev1.WorkloadMember{
									ResourceId: workloadMemberResourceId,
									Instance: &computev1.InstanceResource{
										ResourceId: instanceResourceId,
									},
								},
							},
						}, nil).Once(),
				}
			},
		},
		{
			name:       "failed with client error",
			workloadId: workloadResourceId,
			instanceId: instanceResourceId,
			tenantId:   tenantId,
			expectedOutput: AddInstanceToWorkloadOutput{
				Err: ErrFailedInventoryCreateResource,
			},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedInventoryClient.On("Create", mock.Anything, tenantId, mock.Anything).
						Return(nil, ErrGenericInventoryClientErrror).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			out := machineProvider.AddInstanceToWorkload(AddInstanceToWorkloadInput{
				TenantId:   tc.tenantId,
				WorkloadId: tc.workloadId,
				InstanceId: tc.instanceId})
			require.NotNil(t, out)

			if tc.expectedOutput.Err != nil {
				require.NotNil(t, out.Err)
				assert.EqualError(t, tc.expectedOutput.Err, out.Err.Error())
				return
			}

			require.Nil(t, out.Err)
		})
	}
}

func TestDeleteWorkloadMemberInInventory(t *testing.T) {
	mockedInventoryClient := newMockedInventoryTestClient()
	machineProvider := MachineProvider{client: InventoryClient{Client: mockedInventoryClient}}

	cases := []struct {
		name           string
		tenantId       string
		workloadId     string
		instanceId     string
		expectedOutput DeleteInstanceFromWorkloadOutput
		mocks          func() []*mock.Call
	}{
		{
			name:       "failed with invalid workload id",
			workloadId: "",
			expectedOutput: DeleteInstanceFromWorkloadOutput{
				Err: ErrInvalidWorkloadIdInput,
			},
		},
		{
			name:       "failed with invalid workload id",
			workloadId: workloadResourceId,
			instanceId: "",
			expectedOutput: DeleteInstanceFromWorkloadOutput{
				Err: ErrInvalidInstanceIdInput,
			},
		},
		{
			name:       "failed with invalid tenant id",
			workloadId: workloadResourceId,
			instanceId: instanceResourceId,
			tenantId:   "",
			expectedOutput: DeleteInstanceFromWorkloadOutput{
				Err: ErrInvalidTenantIdInput,
			},
		},
		{
			name:       "failed with get instance error",
			workloadId: workloadResourceId,
			instanceId: instanceResourceId,
			tenantId:   tenantId,
			expectedOutput: DeleteInstanceFromWorkloadOutput{
				Err: ErrFailedInventoryGetResource,
			},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedInventoryClient.On("Get", mock.Anything, tenantId, instanceResourceId).
						Return(nil, ErrGenericInventoryClientErrror).Once(),
				}
			},
		},
		{
			name:       "failed with no instance",
			workloadId: workloadResourceId,
			instanceId: instanceResourceId,
			tenantId:   tenantId,
			expectedOutput: DeleteInstanceFromWorkloadOutput{
				Err: ErrInvalidInstance,
			},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedInventoryClient.On("Get", mock.Anything, tenantId, instanceResourceId).
						Return(&inventoryv1.GetResourceResponse{
							Resource: &inventoryv1.Resource{
								Resource: &inventoryv1.Resource_Instance{},
							},
						}, nil).Once(),
				}
			},
		},
		{
			name:       "failed with no workload members",
			workloadId: workloadResourceId,
			instanceId: instanceResourceId,
			tenantId:   tenantId,
			expectedOutput: DeleteInstanceFromWorkloadOutput{
				Err: ErrInvalidWorkloadMembers,
			},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedInventoryClient.On("Get", mock.Anything, tenantId, instanceResourceId).
						Return(&inventoryv1.GetResourceResponse{
							Resource: &inventoryv1.Resource{
								Resource: &inventoryv1.Resource_Instance{
									Instance: &computev1.InstanceResource{
										ResourceId: instanceResourceId,
									},
								},
							},
						}, nil).Once(),
				}
			},
		},
		{
			name:           "successful, no errors",
			workloadId:     workloadResourceId,
			instanceId:     instanceResourceId,
			tenantId:       tenantId,
			expectedOutput: DeleteInstanceFromWorkloadOutput{},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedInventoryClient.On("Get", mock.Anything, tenantId, instanceResourceId).
						Return(&inventoryv1.GetResourceResponse{
							Resource: &inventoryv1.Resource{
								Resource: &inventoryv1.Resource_Instance{
									Instance: &computev1.InstanceResource{
										ResourceId: instanceResourceId,
										WorkloadMembers: []*computev1.WorkloadMember{
											{
												ResourceId: workloadMemberResourceId,
												Kind:       computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
												Workload: &computev1.WorkloadResource{
													ResourceId: workloadResourceId,
													Kind:       computev1.WorkloadKind_WORKLOAD_KIND_CLUSTER,
												},
											},
										},
									},
								},
							},
						}, nil).Once(),
					mockedInventoryClient.On("Delete", mock.Anything, tenantId, workloadMemberResourceId).
						Return(&inventoryv1.DeleteResourceResponse{}, nil).Once(),
				}
			},
		},
		{
			name:       "failed with client delete error",
			workloadId: workloadResourceId,
			tenantId:   tenantId,
			instanceId: instanceResourceId,
			expectedOutput: DeleteInstanceFromWorkloadOutput{
				Err: ErrFailedInventoryDeleteResource,
			},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedInventoryClient.On("Get", mock.Anything, tenantId, instanceResourceId).
						Return(&inventoryv1.GetResourceResponse{
							Resource: &inventoryv1.Resource{
								Resource: &inventoryv1.Resource_Instance{
									Instance: &computev1.InstanceResource{
										ResourceId: instanceResourceId,
										WorkloadMembers: []*computev1.WorkloadMember{
											{
												ResourceId: workloadMemberResourceId,
												Kind:       computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
												Workload: &computev1.WorkloadResource{
													ResourceId: workloadResourceId,
													Kind:       computev1.WorkloadKind_WORKLOAD_KIND_CLUSTER,
												},
											},
										},
									},
								},
							},
						}, nil).Once(),
					mockedInventoryClient.On("Delete", mock.Anything, tenantId, workloadMemberResourceId).
						Return(nil, ErrGenericInventoryClientErrror).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			out := machineProvider.DeleteInstanceFromWorkload(DeleteInstanceFromWorkloadInput{
				TenantId:   tc.tenantId,
				WorkloadId: tc.workloadId,
				InstanceId: tc.instanceId})
			require.NotNil(t, out)

			if tc.expectedOutput.Err != nil {
				require.NotNil(t, out.Err)
				assert.EqualError(t, tc.expectedOutput.Err, out.Err.Error())
				return
			}

			require.Nil(t, out.Err)
		})
	}
}
