// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package inventory

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/cluster-api-provider-intel/mocks/m_client"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventoryv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
)

var (
	nodeUUId                      = "3c7ef083-0e9e-4d05-b5a6-5bcd72261000"
	hostResourceId                = "host-3c7ef083"
	hostSerialNumber              = "host-serial-number-1"
	invalidHostResourceId         = "3c7ef083-0e9e-4d05-b5a6-5bcd72261001"
	tenantId                      = "d88fead9-abf3-47e5-aeb4-cc5de5b143dc"
	instanceResourceId            = "inst-3c7ef083"
	invalidInstanceResourceId     = "3c7ef083"
	instanceOsName                = "os name"
	workloadResourceId            = "workload-ac124511"
	workloadMemberResourceId      = "workloadmember-ac124511"
	invalidWorkloadResourceId     = "workload-tt124511"
	clusterName                   = "test-cluster-1"
	invalidClusterNameForWorkload = "45676767676767676767676767676767676767676767676767676767676767676767767676"
	errFailedGet                  = errors.New("failed get error")
)

func newMockedInventoryTestClient() *m_client.MockTenantAwareInventoryClient {
	return &m_client.MockTenantAwareInventoryClient{}
}

func TestGetHostAndInstanceByHostUUID(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	inventoryClient := InventoryClient{Client: mockedClient}

	cases := []struct {
		name         string
		tenantId     string
		nodeUUId     string
		expectedHost *computev1.HostResource
		mocks        func() []*mock.Call
	}{
		{
			name:     "successful with valid result",
			nodeUUId: nodeUUId,
			tenantId: tenantId,
			expectedHost: &computev1.HostResource{
				ResourceId:   hostResourceId,
				SerialNumber: hostSerialNumber,
				Instance: &computev1.InstanceResource{
					ResourceId: instanceResourceId,
					CurrentOs: &osv1.OperatingSystemResource{
						Name: instanceOsName,
					},
				},
			},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("GetHostByUUID", mock.Anything, tenantId, nodeUUId).
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
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			host, err := inventoryClient.getHost(context.TODO(), tc.tenantId, tc.nodeUUId)
			require.Nil(t, err)
			require.NotNil(t, host)
			assert.Equal(t, tc.expectedHost.GetResourceId(), host.GetResourceId())
			assert.Equal(t, tc.expectedHost.GetSerialNumber(), host.GetSerialNumber())
			require.NotNil(t, host.GetInstance())
			assert.Equal(t, tc.expectedHost.GetInstance().GetResourceId(), host.GetInstance().GetResourceId())
			assert.Equal(t, tc.expectedHost.GetInstance().GetCurrentOs().GetName(), host.GetInstance().GetCurrentOs().GetName())
		})
	}
}

func TestGetHostAndInstanceByHostUUIDWithErrors(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	inventoryClient := InventoryClient{Client: mockedClient}

	cases := []struct {
		name               string
		hasClientError     bool
		expectedError      error
		hasValidationError bool
		mocks              func() []*mock.Call
	}{
		{name: "successful with invalid result",
			expectedError: ErrInvalidHost,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("GetHostByUUID", mock.Anything, mock.Anything, mock.Anything).
						Return(&computev1.HostResource{ResourceId: invalidHostResourceId}, nil).Once(),
				}
			},
		},
		{name: "successful with nil result",
			expectedError: ErrInvalidHost,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("GetHostByUUID", mock.Anything, mock.Anything, mock.Anything).
						Return(nil, nil).Once(),
				}
			},
		},
		{name: "failed get request to inventory",
			expectedError: ErrFailedInventoryGetHostByUuid,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("GetHostByUUID", mock.Anything, mock.Anything, mock.Anything).
						Return(nil, errFailedGet).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			host, err := inventoryClient.getHost(context.TODO(), tenantId, nodeUUId)
			require.Nil(t, host)
			require.NotNil(t, err)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestGetInstance(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	inventoryClient := InventoryClient{Client: mockedClient}

	cases := []struct {
		name             string
		tenantId         string
		instanceId       string
		expectedInstance *computev1.InstanceResource
		mocks            func() []*mock.Call
	}{
		{
			name:       "successful with valid result",
			instanceId: instanceResourceId,
			tenantId:   tenantId,
			expectedInstance: &computev1.InstanceResource{
				ResourceId: instanceResourceId,
				CurrentOs: &osv1.OperatingSystemResource{
					Name: instanceOsName,
				},
			},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, tenantId, instanceResourceId).
						Return(&inventoryv1.GetResourceResponse{
							Resource: &inventoryv1.Resource{
								Resource: &inventoryv1.Resource_Instance{
									Instance: &computev1.InstanceResource{
										ResourceId: instanceResourceId,
										CurrentOs: &osv1.OperatingSystemResource{
											Name: instanceOsName,
										},
									},
								},
							},
						}, nil).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			out, err := inventoryClient.getInstance(context.TODO(), tc.tenantId, tc.instanceId)
			require.Nil(t, err)
			require.NotNil(t, out)
			assert.Equal(t, out.GetResourceId(), tc.instanceId)
			assert.Equal(t, tc.expectedInstance.GetCurrentOs().GetName(), out.GetCurrentOs().GetName())
		})
	}
}

func TestGetInstanceWithErrors(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	inventoryClient := InventoryClient{Client: mockedClient}

	cases := []struct {
		name          string
		expectedError error
		mocks         func() []*mock.Call
	}{
		{name: "successful get with invalid instance id",
			expectedError: ErrInvalidInstance,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, mock.Anything, mock.Anything).
						Return(&inventoryv1.GetResourceResponse{
							Resource: &inventoryv1.Resource{
								Resource: &inventoryv1.Resource_Instance{
									Instance: &computev1.InstanceResource{
										ResourceId: invalidInstanceResourceId,
									},
								},
							},
						}, nil).Once(),
				}
			},
		},
		{name: "successful get with empty instance",
			expectedError: ErrInvalidInstance,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, mock.Anything, mock.Anything).
						Return(&inventoryv1.GetResourceResponse{
							Resource: &inventoryv1.Resource{
								Resource: &inventoryv1.Resource_Instance{},
							},
						}, nil).Once(),
				}
			},
		},
		{name: "successful get with empty resource",
			expectedError: ErrInvalidInstance,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, mock.Anything, mock.Anything).
						Return(&inventoryv1.GetResourceResponse{}, nil).Once(),
				}
			},
		},
		{name: "successful get with empty get response",
			expectedError: ErrInvalidInstance,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, mock.Anything, mock.Anything).
						Return(nil, nil).Once(),
				}
			},
		},
		{name: "failed get request to inventory",
			expectedError: ErrFailedInventoryGetResource,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, mock.Anything, mock.Anything).
						Return(nil, errFailedGet).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			instance, err := inventoryClient.getInstance(context.TODO(), tenantId, instanceResourceId)
			require.Nil(t, instance)
			require.NotNil(t, err)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestCreateWorkload(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	inventoryClient := InventoryClient{Client: mockedClient}

	cases := []struct {
		name               string
		tenantId           string
		workloadId         string
		expectedWorkloadId string
		clusterName        string
		mocks              func() []*mock.Call
	}{
		{
			name:               "successful with valid result",
			workloadId:         workloadResourceId,
			tenantId:           tenantId,
			expectedWorkloadId: workloadResourceId,
			clusterName:        clusterName,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, tenantId, mock.Anything).
						Return(&inventoryv1.Resource{
							Resource: &inventoryv1.Resource_Workload{
								Workload: &computev1.WorkloadResource{
									ResourceId:   workloadResourceId,
									Name:         clusterName,
									ExternalId:   clusterName,
									Kind:         computev1.WorkloadKind_WORKLOAD_KIND_CLUSTER,
									DesiredState: computev1.WorkloadState_WORKLOAD_STATE_PROVISIONED,
								},
							},
						}, nil).Once(),
				}
			},
		},
		{
			name:               "successful with empty clusterName",
			workloadId:         workloadResourceId,
			tenantId:           tenantId,
			expectedWorkloadId: workloadResourceId,
			clusterName:        "",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, tenantId, mock.Anything).
						Return(&inventoryv1.Resource{
							Resource: &inventoryv1.Resource_Workload{
								Workload: &computev1.WorkloadResource{
									ResourceId:   workloadResourceId,
									Name:         clusterName,
									ExternalId:   clusterName,
									Kind:         computev1.WorkloadKind_WORKLOAD_KIND_CLUSTER,
									DesiredState: computev1.WorkloadState_WORKLOAD_STATE_PROVISIONED,
								},
							},
						}, nil).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			out, err := inventoryClient.createWorkload(context.TODO(), tc.tenantId, tc.clusterName)
			require.Nil(t, err)
			require.NotNil(t, out)
			assert.Equal(t, out, tc.expectedWorkloadId)
		})
	}
}

func TestCreateWorkloadWithErrors(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	inventoryClient := InventoryClient{Client: mockedClient}

	cases := []struct {
		name          string
		tenantId      string
		clusterName   string
		expectedError error
		mocks         func() []*mock.Call
	}{
		{name: "successful create with invalid workload id",
			tenantId:      tenantId,
			clusterName:   clusterName,
			expectedError: ErrInvalidWorkload,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, tenantId, mock.Anything).
						Return(&inventoryv1.Resource{
							Resource: &inventoryv1.Resource_Workload{
								Workload: &computev1.WorkloadResource{
									ResourceId: invalidWorkloadResourceId,
									Name:       clusterName,
									ExternalId: clusterName,
								},
							},
						}, nil).Once(),
				}
			},
		},
		{name: "successful create with empty workload id",
			tenantId:      tenantId,
			clusterName:   clusterName,
			expectedError: ErrInvalidWorkload,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, tenantId, mock.Anything).
						Return(&inventoryv1.Resource{
							Resource: &inventoryv1.Resource_Workload{
								Workload: &computev1.WorkloadResource{
									ResourceId: "",
									Name:       clusterName,
									ExternalId: clusterName,
								},
							},
						}, nil).Once(),
				}
			},
		},
		{name: "successful create with empty workload",
			tenantId:      tenantId,
			clusterName:   clusterName,
			expectedError: ErrInvalidWorkload,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, tenantId, mock.Anything).
						Return(&inventoryv1.Resource{
							Resource: &inventoryv1.Resource_Workload{},
						}, nil).Once(),
				}
			},
		},
		{name: "successful create with empty resource",
			tenantId:      tenantId,
			clusterName:   clusterName,
			expectedError: ErrInvalidWorkload,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, tenantId, mock.Anything).
						Return(nil, nil).Once(),
				}
			},
		},
		{name: "faild with invalid externalId",
			tenantId:      tenantId,
			clusterName:   invalidClusterNameForWorkload,
			expectedError: ErrInvalidWorkloadInput,
		},
		{name: "failed create request to inventory",
			tenantId:      tenantId,
			clusterName:   clusterName,
			expectedError: ErrFailedInventoryCreateResource,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, tenantId, mock.Anything).
						Return(nil, errFailedGet).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			out, err := inventoryClient.createWorkload(context.TODO(), tc.tenantId, tc.clusterName)
			require.NotNil(t, err)
			assert.Equal(t, out, "")
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestDeleteWorkload(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	inventoryClient := InventoryClient{Client: mockedClient}

	cases := []struct {
		name       string
		tenantId   string
		workloadId string
		mocks      func() []*mock.Call
	}{
		{
			name:       "successful with valid result",
			workloadId: workloadResourceId,
			tenantId:   tenantId,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, tenantId, workloadResourceId).
						Return(&inventoryv1.DeleteResourceResponse{}, nil).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			err := inventoryClient.deleteWorkload(context.TODO(), tc.tenantId, tc.workloadId)
			require.Nil(t, err)
		})
	}
}

func TestDeleteWorkloadWithErrors(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	inventoryClient := InventoryClient{Client: mockedClient}

	cases := []struct {
		name                string
		tenantId            string
		workloadId          string
		hasClientError      bool
		expectedClientError error
		mocks               func() []*mock.Call
	}{
		{name: "failed get request to inventory",
			tenantId:            tenantId,
			workloadId:          workloadResourceId,
			hasClientError:      true,
			expectedClientError: ErrFailedInventoryDeleteResource,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, tenantId, workloadResourceId).
						Return(nil, errFailedGet).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			err := inventoryClient.deleteWorkload(context.TODO(), tc.tenantId, tc.workloadId)
			require.NotNil(t, err)

			assert.Equal(t, tc.expectedClientError, err)
		})
	}
}

func TestGetWorkload(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	inventoryClient := InventoryClient{Client: mockedClient}

	cases := []struct {
		name       string
		tenantId   string
		workloadId string
		mocks      func() []*mock.Call
	}{
		{
			name:       "successful with valid result",
			workloadId: workloadResourceId,
			tenantId:   tenantId,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, tenantId, workloadResourceId).
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
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			out, err := inventoryClient.getWorkload(context.TODO(), tc.tenantId, tc.workloadId)
			require.Nil(t, err)
			require.NotNil(t, out)
			assert.Equal(t, out.GetResourceId(), tc.workloadId)
		})
	}
}

func TestGetWorkloadWithErrors(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	inventoryClient := InventoryClient{Client: mockedClient}

	cases := []struct {
		name          string
		tenantId      string
		workloadId    string
		expectedError error
		mocks         func() []*mock.Call
	}{
		{name: "successful get with invalid workload id",
			tenantId:      tenantId,
			workloadId:    invalidWorkloadResourceId,
			expectedError: ErrInvalidWorkload,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, tenantId, invalidWorkloadResourceId).
						Return(&inventoryv1.GetResourceResponse{
							Resource: &inventoryv1.Resource{
								Resource: &inventoryv1.Resource_Workload{
									Workload: &computev1.WorkloadResource{
										ResourceId: invalidWorkloadResourceId,
									},
								},
							},
						}, nil).Once(),
				}
			},
		},
		{name: "successful get with empty workload",
			tenantId:      tenantId,
			workloadId:    invalidWorkloadResourceId,
			expectedError: ErrInvalidWorkload,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, tenantId, invalidWorkloadResourceId).
						Return(&inventoryv1.GetResourceResponse{
							Resource: &inventoryv1.Resource{
								Resource: &inventoryv1.Resource_Workload{},
							},
						}, nil).Once(),
				}
			},
		},
		{name: "successful get with empty resource",
			tenantId:      tenantId,
			workloadId:    invalidWorkloadResourceId,
			expectedError: ErrInvalidWorkload,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, tenantId, invalidWorkloadResourceId).
						Return(&inventoryv1.GetResourceResponse{}, nil).Once(),
				}
			},
		},
		{name: "successful get with empty get response",
			tenantId:      tenantId,
			workloadId:    invalidWorkloadResourceId,
			expectedError: ErrInvalidWorkload,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, tenantId, invalidWorkloadResourceId).
						Return(nil, nil).Once(),
				}
			},
		},
		{name: "failed get request to inventory",
			tenantId:      tenantId,
			workloadId:    workloadResourceId,
			expectedError: ErrFailedInventoryGetResource,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, tenantId, workloadResourceId).
						Return(nil, errFailedGet).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			out, err := inventoryClient.getWorkload(context.TODO(), tc.tenantId, tc.workloadId)
			require.NotNil(t, err)
			require.Nil(t, out)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestCreateWorkloadMember(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	inventoryClient := InventoryClient{Client: mockedClient}

	cases := []struct {
		name        string
		tenantId    string
		workloadId  string
		instanceId  string
		clusterName string
		mocks       func() []*mock.Call
	}{
		{
			name:        "successful with valid result",
			workloadId:  workloadResourceId,
			tenantId:    tenantId,
			instanceId:  instanceResourceId,
			clusterName: clusterName,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, tenantId, mock.Anything).
						Return(&inventoryv1.Resource{
							Resource: &inventoryv1.Resource_WorkloadMember{
								WorkloadMember: &computev1.WorkloadMember{
									Workload: &computev1.WorkloadResource{},
								},
							},
						}, nil).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			err := inventoryClient.createWorkloadMember(context.TODO(), tc.tenantId, tc.workloadId, tc.instanceId)
			require.Nil(t, err)
		})
	}
}

func TestCreateWorkloadMemberWithErrors(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	inventoryClient := InventoryClient{Client: mockedClient}

	cases := []struct {
		name          string
		tenantId      string
		workloadId    string
		instanceId    string
		expectedError error
		mocks         func() []*mock.Call
	}{
		{name: "failed with invalid workload id",
			tenantId:      tenantId,
			workloadId:    invalidWorkloadResourceId,
			expectedError: ErrInvalidWorkloadMembersInput,
		},
		{name: "failed with invalid instance id",
			tenantId:      tenantId,
			workloadId:    workloadResourceId,
			instanceId:    invalidInstanceResourceId,
			expectedError: ErrInvalidWorkloadMembersInput,
		},
		{name: "successful create with empty resource",
			tenantId:      tenantId,
			expectedError: ErrInvalidWorkloadMembers,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything, mock.Anything).
						Return(nil, nil).Once(),
				}
			},
		},
		{name: "successful create with empty workload member",
			tenantId:      tenantId,
			expectedError: ErrInvalidWorkloadMembers,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything, mock.Anything).
						Return(&inventoryv1.Resource{
							Resource: &inventoryv1.Resource_WorkloadMember{},
						}, nil).Once(),
				}
			},
		},
		{name: "successful create with invalid workload member",
			tenantId:      tenantId,
			expectedError: ErrInvalidWorkloadMembers,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything, mock.Anything).
						Return(&inventoryv1.Resource{
							Resource: &inventoryv1.Resource_WorkloadMember{
								WorkloadMember: &computev1.WorkloadMember{
									Instance: &computev1.InstanceResource{
										ResourceId: invalidInstanceResourceId,
									},
								},
							},
						}, nil).Once(),
				}
			},
		},
		{name: "failed create request to inventory",
			tenantId:      tenantId,
			workloadId:    workloadResourceId,
			instanceId:    instanceResourceId,
			expectedError: ErrFailedInventoryCreateResource,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything, mock.Anything).
						Return(nil, errFailedGet).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			err := inventoryClient.createWorkloadMember(context.TODO(), tc.tenantId, tc.workloadId, tc.instanceId)
			require.NotNil(t, err)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestDeleteWorkloadMember(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	inventoryClient := InventoryClient{Client: mockedClient}

	cases := []struct {
		name             string
		tenantId         string
		workloadMemberId string
		mocks            func() []*mock.Call
	}{
		{
			name:             "successful with valid result",
			workloadMemberId: workloadMemberResourceId,
			tenantId:         tenantId,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, tenantId, workloadMemberResourceId).
						Return(&inventoryv1.DeleteResourceResponse{}, nil).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			err := inventoryClient.deleteWorkloadMember(context.TODO(), tc.tenantId, tc.workloadMemberId)
			require.Nil(t, err)
		})
	}
}

func TestDeleteWorkloadMemberWithErrors(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	inventoryClient := InventoryClient{Client: mockedClient}

	cases := []struct {
		name                string
		tenantId            string
		workloadMemberId    string
		hasClientError      bool
		expectedClientError error
		mocks               func() []*mock.Call
	}{
		{name: "failed delete request to inventory",
			tenantId:            tenantId,
			workloadMemberId:    workloadMemberResourceId,
			hasClientError:      true,
			expectedClientError: ErrFailedInventoryDeleteResource,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, tenantId, workloadMemberResourceId).
						Return(nil, errFailedGet).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			err := inventoryClient.deleteWorkloadMember(context.TODO(), tc.tenantId, tc.workloadMemberId)
			require.NotNil(t, err)
			assert.Equal(t, tc.expectedClientError, err)
		})
	}
}
