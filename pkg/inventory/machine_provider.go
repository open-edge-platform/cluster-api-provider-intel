// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package inventory

import (
	"context"
	"log/slog"
	"sync"

	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
)

type MachineProvider struct {
	client InventoryClient
}

func NewMachineProvider(wg *sync.WaitGroup, inventoryAddress string,
	enableTracing, enableMetrics bool, useStub bool) (*MachineProvider, error) {
	options := NewOptionsBuilder().
		WithInventoryAddress(inventoryAddress).
		WithWaitGroup(wg).WithTracing(enableTracing).
		WithMetrics(enableMetrics).
		WithStub(useStub).
		Build()

	client, err := NewInventoryClientWithOptions(options)

	if err != nil {
		return nil, err
	}

	return &MachineProvider{client: *client}, nil
}

func (n *MachineProvider) CreateWorkload(in CreateWorkloadInput) CreateWorkloadOutput {
	if in.ClusterName == "" {
		return CreateWorkloadOutput{"", ErrInvalidClusterNameInput}
	}

	if in.TenantId == "" {
		return CreateWorkloadOutput{"", ErrInvalidTenantIdInput}
	}

	workloadId, err := n.client.createWorkload(context.Background(), in.TenantId, in.ClusterName)
	return CreateWorkloadOutput{workloadId, err}
}

func (n *MachineProvider) DeleteWorkload(in DeleteWorkloadInput) DeleteWorkloadOutput {
	if in.WorkloadId == "" {
		return DeleteWorkloadOutput{ErrInvalidWorkloadIdInput}
	}

	if in.TenantId == "" {
		return DeleteWorkloadOutput{ErrInvalidTenantIdInput}
	}

	err := n.client.deleteWorkload(context.Background(), in.TenantId, in.WorkloadId)
	return DeleteWorkloadOutput{Err: err}
}

func (n *MachineProvider) GetWorkload(in GetWorkloadInput) GetWorkloadOutput {
	if in.WorkloadId == "" {
		return GetWorkloadOutput{nil, ErrInvalidWorkloadIdInput}
	}

	if in.TenantId == "" {
		return GetWorkloadOutput{nil, ErrInvalidTenantIdInput}
	}

	// workload resource and workload id are validated in the client's getWorkload; any
	// invalid values would return an error
	workloadResource, err := n.client.getWorkload(context.Background(), in.TenantId, in.WorkloadId)
	if err != nil {
		return GetWorkloadOutput{nil, err}
	}

	return GetWorkloadOutput{Workload: &Workload{Id: workloadResource.GetResourceId()}, Err: nil}
}

func (n *MachineProvider) GetInstanceByMachineId(in GetInstanceByMachineIdInput) GetInstanceByMachineIdOutput {
	if in.MachineId == "" {
		return GetInstanceByMachineIdOutput{nil, nil, ErrInvalidMachineIdInput}
	}

	if in.TenantId == "" {
		return GetInstanceByMachineIdOutput{nil, nil, ErrInvalidTenantIdInput}
	}

	// checks for host being nil are made in the client's getHost
	// any invalid values would return an error
	host, err := n.client.getHost(context.Background(), in.TenantId, in.MachineId)
	if err != nil {
		return GetInstanceByMachineIdOutput{nil, nil, err}
	}

	instance := host.GetInstance()
	if instance == nil || instance.GetResourceId() == "" {
		slog.Warn("invalid instance associated to host resource", "error:", err,
			"tenant id", in.TenantId, "machine id", in.MachineId)
		return GetInstanceByMachineIdOutput{Err: ErrInvalidInstance}
	}

	return GetInstanceByMachineIdOutput{
		&Host{
			Id: host.GetResourceId(),
		},
		&Instance{
			Id:        instance.GetResourceId(),
			SerialNo:  host.GetSerialNumber(),
			Os:        instance.GetCurrentOs().GetName(),
			MachineId: in.MachineId},
		nil}
}

func (n *MachineProvider) AddInstanceToWorkload(in AddInstanceToWorkloadInput) AddInstanceToWorkloadOutput {
	if in.WorkloadId == "" {
		return AddInstanceToWorkloadOutput{ErrInvalidWorkloadIdInput}
	}

	if in.InstanceId == "" {
		return AddInstanceToWorkloadOutput{ErrInvalidInstanceIdInput}
	}

	if in.TenantId == "" {
		return AddInstanceToWorkloadOutput{ErrInvalidTenantIdInput}
	}

	if err := n.client.createWorkloadMember(context.Background(), in.TenantId, in.WorkloadId, in.InstanceId); err != nil {
		return AddInstanceToWorkloadOutput{err}
	}

	return AddInstanceToWorkloadOutput{}
}

func (n *MachineProvider) DeleteInstanceFromWorkload(
	in DeleteInstanceFromWorkloadInput) DeleteInstanceFromWorkloadOutput {
	if in.WorkloadId == "" {
		return DeleteInstanceFromWorkloadOutput{ErrInvalidWorkloadIdInput}
	}

	if in.InstanceId == "" {
		return DeleteInstanceFromWorkloadOutput{ErrInvalidInstanceIdInput}
	}

	if in.TenantId == "" {
		return DeleteInstanceFromWorkloadOutput{ErrInvalidTenantIdInput}
	}

	ctx := context.Background()
	instance, err := n.client.getInstance(ctx, in.TenantId, in.InstanceId)
	if err != nil {
		slog.Warn("invalid instance found", "error:", err,
			"tenant id", in.TenantId, "workload id", in.WorkloadId, "instance id", in.InstanceId)
		return DeleteInstanceFromWorkloadOutput{err}
	}

	if instance.GetWorkloadMembers() == nil {
		slog.Warn("empty workload members associated with instaince", "error:", err,
			"tenant id", in.TenantId, "workload id", in.WorkloadId, "instance id", in.InstanceId)
		return DeleteInstanceFromWorkloadOutput{ErrInvalidWorkloadMembers}
	}

	var associatedWorkloadId string
	// one instance is associated to just one workload member but instance.GetWorkloadMembers()
	// returns an array, so more straight forward to iterate through instead of working with indices
	// no notice from inventory yet regarding the refactor of the method name/signature
	for _, workloadMember := range instance.GetWorkloadMembers() {
		if workloadMember.GetKind() == computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE {
			workload := workloadMember.GetWorkload()
			if workload != nil && workload.GetKind() == computev1.WorkloadKind_WORKLOAD_KIND_CLUSTER {
				associatedWorkloadId = workload.GetResourceId()
				// only delete if we found the right workloadMember ID - workload ID associated to the instance
				if associatedWorkloadId != "" && associatedWorkloadId == in.WorkloadId && workloadMember.GetResourceId() != "" {
					if err = n.client.deleteWorkloadMember(ctx, in.TenantId, workloadMember.GetResourceId()); err != nil {
						return DeleteInstanceFromWorkloadOutput{err}
					}
				}
			}
		}
	}

	return DeleteInstanceFromWorkloadOutput{}
}

func (n *MachineProvider) DeauthorizeHost(in DeauthorizeHostInput) DeauthorizeHostOutput {
	if in.HostUUID == "" {
		return DeauthorizeHostOutput{ErrInvalidHostUUIDInput}
	}

	if in.TenantId == "" {
		return DeauthorizeHostOutput{ErrInvalidTenantIdInput}
	}

	if err := n.client.deauthorizeHost(context.Background(), in.TenantId, in.HostUUID); err != nil {
		return DeauthorizeHostOutput{err}
	}

	return DeauthorizeHostOutput{}
}

func (n *MachineProvider) Close() {
	n.client.close()
}
