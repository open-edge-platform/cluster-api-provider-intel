// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package inventory

import "errors"

type InfrastructureProvider interface {
	CreateWorkload(in CreateWorkloadInput) CreateWorkloadOutput
	DeleteWorkload(in DeleteWorkloadInput) DeleteWorkloadOutput
	GetWorkload(in GetWorkloadInput) GetWorkloadOutput
	GetInstanceByMachineId(in GetInstanceByMachineIdInput) GetInstanceByMachineIdOutput
	AddInstanceToWorkload(in AddInstanceToWorkloadInput) AddInstanceToWorkloadOutput
	DeleteInstanceFromWorkload(in DeleteInstanceFromWorkloadInput) DeleteInstanceFromWorkloadOutput
}

type Workload struct {
	Id string
}

type Instance struct {
	Id        string
	SerialNo  string
	Os        string
	MachineId string
}

type Host struct {
	Id string
}

type CreateWorkloadInput struct {
	TenantId    string
	ClusterName string
}

type CreateWorkloadOutput struct {
	WorkloadId string
	Err        error
}

type DeleteWorkloadInput struct {
	TenantId   string
	WorkloadId string
}

type DeleteWorkloadOutput struct {
	Err error
}

type GetWorkloadInput struct {
	TenantId   string
	WorkloadId string
}

type GetWorkloadOutput struct {
	Workload *Workload
	Err      error
}

type GetInstanceByMachineIdInput struct {
	TenantId  string
	MachineId string
}

type GetInstanceByMachineIdOutput struct {
	Host     *Host
	Instance *Instance
	Err      error
}

type AddInstanceToWorkloadInput struct {
	TenantId   string
	WorkloadId string
	InstanceId string
}

type AddInstanceToWorkloadOutput struct {
	Err error
}

type DeleteInstanceFromWorkloadInput struct {
	TenantId   string
	WorkloadId string
	InstanceId string
}

type DeleteInstanceFromWorkloadOutput struct {
	Err error
}

type DeauthorizeHostInput struct {
	TenantId string
	HostUUID string
}

type DeauthorizeHostOutput struct {
	Err error
}

var (
	ErrInvalidHostUUIDInput          = errors.New("invalid host UUID value")
	ErrInvalidInstanceIdInput        = errors.New("invalid instance id value")
	ErrInvalidWorkloadIdInput        = errors.New("invalid workload id value")
	ErrInvalidMachineIdInput         = errors.New("invalid machine id value")
	ErrInvalidTenantIdInput          = errors.New("invalid tenant id value")
	ErrInvalidClusterNameInput       = errors.New("invalid cluster name value")
	ErrInvalidInventoryResource      = errors.New("invalid inventory resource")
	ErrInvalidInstance               = errors.New("invalid instance")
	ErrInvalidWorkload               = errors.New("invalid workload")
	ErrInvalidWorkloadInput          = errors.New("invalid workload input values")
	ErrInvalidWorkloadMembers        = errors.New("invalid workload members")
	ErrInvalidWorkloadMembersInput   = errors.New("invalid workload members input values")
	ErrInvalidHost                   = errors.New("invalid host")
	ErrFailedInventoryGetHostByUuid  = errors.New("failed inventory getHostByUUID call")
	ErrFailedInventoryGetResource    = errors.New("failed inventory get resource call")
	ErrFailedInventoryCreateResource = errors.New("failed inventory create resource call")
	ErrFailedInventoryDeleteResource = errors.New("failed inventory delete resource call")
	ErrFailedInventoryGetResponse    = errors.New("failed inventory get resource call")
	ErrFailedInventoryGetHost        = errors.New("failed inventory resource getHost call")
)
