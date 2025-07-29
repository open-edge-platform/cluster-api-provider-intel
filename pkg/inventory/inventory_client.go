// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package inventory

import (
	"context"
	"log/slog"
	"sync"
	"time"

	invStub "github.com/open-edge-platform/cluster-api-provider-intel/test/inventory-stub"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventoryv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	defaultInventoryTimeout = 5 * time.Second
	ScaleFactor             = 5
	clientName              = "IntelProviderInventoryClient"
)

type Options struct {
	wg               *sync.WaitGroup
	inventoryAddress string
	enableTracing    bool
	enableMetrics    bool
	useStub          bool
}

type optionsBuilder struct {
	options *Options
}

type OptionsBuilder interface {
	WithWaitGroup(wg *sync.WaitGroup) OptionsBuilder
	WithInventoryAddress(address string) OptionsBuilder
	WithTracing(enableTracing bool) OptionsBuilder
	WithMetrics(enableMetrics bool) OptionsBuilder
	WithStub(useStub bool) OptionsBuilder
	Build() Options
}

func NewOptionsBuilder() OptionsBuilder {
	return &optionsBuilder{
		options: &Options{},
	}
}

func (b *optionsBuilder) WithWaitGroup(wg *sync.WaitGroup) OptionsBuilder {
	b.options.wg = wg
	return b
}

func (b *optionsBuilder) WithInventoryAddress(address string) OptionsBuilder {
	b.options.inventoryAddress = address
	return b
}

func (b *optionsBuilder) WithTracing(enableTracing bool) OptionsBuilder {
	b.options.enableTracing = enableTracing
	return b
}

func (b *optionsBuilder) WithMetrics(enableMetrics bool) OptionsBuilder {
	b.options.enableMetrics = enableMetrics
	return b
}

func (b *optionsBuilder) WithStub(useStub bool) OptionsBuilder {
	b.options.useStub = useStub
	return b
}

func (b *optionsBuilder) Build() Options {
	return *b.options
}

type InventoryClient struct {
	Client client.TenantAwareInventoryClient
}

func newInventoryClientWithOptions(opt Options) (*InventoryClient, error) {
	var invClient client.TenantAwareInventoryClient
	var err error

	if opt.useStub {
		invClient = invStub.GetStubClient()
	} else {
		eventsWatcher := make(chan *client.WatchEvents)
		invClient, err = client.NewTenantAwareInventoryClient(context.Background(), client.InventoryClientConfig{
			Name:                      clientName,
			AbortOnUnknownClientError: true,
			Address:                   opt.inventoryAddress,
			EnableTracing:             opt.enableTracing,
			EnableMetrics:             opt.enableMetrics,
			SecurityCfg:               &client.SecurityConfig{Insecure: true},
			Wg:                        opt.wg,
			Events:                    eventsWatcher,
			ClientKind:                inventoryv1.ClientKind_CLIENT_KIND_API,
		})
		if err != nil {
			return nil, err
		}

	}

	slog.Debug("inventory client started")
	return &InventoryClient{Client: invClient}, nil
}

// IMPORTANT: always close the Inventory client in case of errors
// or signals like syscall.SIGTERM, syscall.SIGINT etc.
func (c *InventoryClient) close() {
	if err := c.Client.Close(); err != nil {
		slog.Error("failed to close the inventory client", "error", err)
	}
	slog.Info("inventory client stopped")
}

func (c *InventoryClient) validateHostResource(host *computev1.HostResource) error {
	if host == nil {
		slog.Info("empty host resource")
		return ErrInvalidHost
	}

	// always a good idea to validate the message. we will use inventory's validators
	// for now. the protos used in the validation are in infra-core repo
	if err := validator.ValidateMessage(host); err != nil {
		slog.Info("host resource proto validation failed", "error", err)
		return ErrInvalidHost
	}

	return nil
}

func (c *InventoryClient) getHost(ctx context.Context, tenantId, hostUUID string) (
	*computev1.HostResource, error,
) {
	slog.Debug("getHost", "tenantId", tenantId, "hostUuid", hostUUID)

	childCtx, cancel := context.WithTimeout(ctx, defaultInventoryTimeout)
	defer cancel()

	respHost, err := c.Client.GetHostByUUID(childCtx, tenantId, hostUUID)
	if err != nil {
		slog.Warn("failed to get host from host uuid, attempting to get by resourceid",
			"tenantId", tenantId, "hostId", hostUUID, "error", err)
		response, err := c.Client.Get(ctx, tenantId, hostUUID)
		if err != nil {
			slog.Warn("failed to get host by resourceId", "error", err, "tenantId", tenantId, "hostId", hostUUID)
			return nil, err
		}

		resource := response.GetResource()
		if resource == nil {
			slog.Warn("response resource is nil", "tenantId", tenantId, "hostUuid", hostUUID)
			return nil, ErrFailedInventoryGetResponse
		}
		respHost = resource.GetHost()
		if respHost == nil {
			slog.Warn("host in response resource is nil", "tenantId", tenantId, "hostUuid", hostUUID)
			return nil, ErrFailedInventoryGetHost
		}

		slog.Debug("success in getting resourceId", "tenantId", tenantId, "hostUuid", hostUUID)
	}

	if err := c.validateHostResource(respHost); err != nil {
		slog.Warn("failed to validate host resource", "error", err, "tenantId", tenantId, "hostUuid", hostUUID)
		return nil, err
	}

	return respHost, nil
}

func (c *InventoryClient) validateResource(resource *inventoryv1.Resource) error {
	if resource == nil {
		slog.Info("empty resource")
		return ErrInvalidInventoryResource
	}

	return nil
}

func (c *InventoryClient) validateInstanceResource(instance *computev1.InstanceResource) error {
	if instance == nil {
		slog.Info("empty instance resource")
		return ErrInvalidInstance
	}

	if err := validator.ValidateMessage(instance); err != nil {
		slog.Info("instance resource proto validation failed", "error", err)
		return ErrInvalidInstance
	}

	return nil
}

func (c *InventoryClient) getInstance(ctx context.Context, tenantId, instanceId string) (
	*computev1.InstanceResource, error) {
	slog.Debug("getInstance", "tenant id", tenantId, "instance id", instanceId)

	childCtx, cancel := context.WithTimeout(ctx, defaultInventoryTimeout)
	defer cancel()

	respInstance, err := c.Client.Get(childCtx, tenantId, instanceId)
	if err != nil {
		slog.Warn("failed to get instance by id", "tenant id", tenantId, "instance id",
			instanceId, "error", err)
		return nil, ErrFailedInventoryGetResource
	}

	if respInstance == nil {
		slog.Warn("failed to validate get resource response from inventory", "tenantId", tenantId, "instance id", instanceId)
		return nil, ErrInvalidInstance
	}

	inventoryResource := respInstance.GetResource()
	if err = c.validateResource(inventoryResource); err != nil {
		slog.Warn("failed to validate resource", "error", err, "tenantId", tenantId, "instance id", instanceId)
		return nil, ErrInvalidInstance
	}

	invInstance := inventoryResource.GetInstance()
	if err = c.validateInstanceResource(invInstance); err != nil {
		slog.Warn("failed to validate instance resource", "error", err, "tenantId", tenantId, "instance id", instanceId)
		return nil, err
	}

	return invInstance, nil
}

func (c *InventoryClient) toGrpcWorkloadResource(tenantId, clusterName string) (*computev1.WorkloadResource, error) {
	workload := &computev1.WorkloadResource{
		Kind:         computev1.WorkloadKind_WORKLOAD_KIND_CLUSTER,
		DesiredState: computev1.WorkloadState_WORKLOAD_STATE_PROVISIONED,
		TenantId:     tenantId,
	}
	if clusterName != "" {
		workload.Name = clusterName
		workload.ExternalId = clusterName
	}

	err := validator.ValidateMessage(workload)
	if err != nil {
		slog.Info("failed to serialize to grpc workload", "error", err)
		return nil, ErrInvalidWorkloadInput
	}

	return workload, nil
}

func (c *InventoryClient) validateWorkloadResource(workload *computev1.WorkloadResource) error {
	if workload == nil {
		slog.Info("empty workload resource")
		return ErrInvalidWorkload
	}

	if err := validator.ValidateMessage(workload); err != nil {
		slog.Info("failed to serialize from resource to workload", "error", err)
		return ErrInvalidWorkload
	}

	if workload.GetResourceId() == "" {
		slog.Info("empty workload id found")
		return ErrInvalidWorkload
	}

	return nil
}

func (c *InventoryClient) createWorkload(ctx context.Context, tenantId, clusterName string) (
	workloadId string, err error) {
	slog.Debug("createWorkload", "tenantID", tenantId, "cluster name", clusterName)

	childCtx, cancel := context.WithTimeout(ctx, defaultInventoryTimeout)
	defer cancel()

	grpcWorkload, err := c.toGrpcWorkloadResource(tenantId, clusterName)
	if err != nil {
		slog.Warn("failed to create workload resource type",
			"tenantId", tenantId, "cluster name", clusterName, "error", err)
		return "", err
	}

	reqResource := &inventoryv1.Resource{
		Resource: &inventoryv1.Resource_Workload{
			Workload: grpcWorkload,
		},
	}

	respWorkload, err := c.Client.Create(childCtx, tenantId, reqResource)
	if err != nil {
		slog.Warn("failed to create workload", "error", err, "tenantID", tenantId, "cluster name", clusterName)
		return "", ErrFailedInventoryCreateResource
	}

	if err := c.validateResource(respWorkload); err != nil {
		slog.Warn("failed to validate resource", "tenantId", tenantId, "cluster name", clusterName)
		return "", ErrInvalidWorkload
	}

	invWorkload := respWorkload.GetWorkload()
	if err = c.validateWorkloadResource(invWorkload); err != nil {
		slog.Warn("failed to validate workload resource", "error", err, "tenantId", tenantId, "cluster name", clusterName)
		return "", err
	}

	return invWorkload.GetResourceId(), nil
}

func (c *InventoryClient) deleteWorkload(ctx context.Context, tenantId, workloadId string) error {
	slog.Debug("deleteWorkload", "tenantID", tenantId, "workload id", workloadId)

	childCtx, cancel := context.WithTimeout(ctx, defaultInventoryTimeout)
	defer cancel()

	_, err := c.Client.Delete(childCtx, tenantId, workloadId)
	if err != nil {
		slog.Warn("failed to delete workload", "tenantId", tenantId, "workload id",
			workloadId, "error", err)
		return ErrFailedInventoryDeleteResource
	}

	return nil
}

func (c *InventoryClient) getWorkload(ctx context.Context, tenantId, workloadId string) (
	*computev1.WorkloadResource, error) {
	slog.Debug("getWorkload", "tenantID", tenantId, "workload id", workloadId)

	childCtx, cancel := context.WithTimeout(ctx, defaultInventoryTimeout)
	defer cancel()

	respWorkload, err := c.Client.Get(childCtx, tenantId, workloadId)
	if err != nil {
		slog.Warn("failed to get workload", "tenantId", tenantId, "workload id",
			workloadId, "error", err)
		return nil, ErrFailedInventoryGetResource
	}

	if respWorkload == nil {
		slog.Warn("failed to validate get workload response from inventory", "tenantId", tenantId, "workload id", workloadId)
		return nil, ErrInvalidWorkload
	}

	workloadResource := respWorkload.GetResource()
	if err = c.validateResource(workloadResource); err != nil {
		slog.Warn("failed to validate resource", "error", err, "tenantId", tenantId, "workload id", workloadId)
		return nil, ErrInvalidWorkload
	}

	invWorkload := workloadResource.GetWorkload()
	if err = c.validateWorkloadResource(invWorkload); err != nil {
		slog.Warn("failed to validate workload resource", "error", err, "tenantId", tenantId, "workload id", workloadId)
		return nil, err
	}

	return invWorkload, nil
}

func (c *InventoryClient) toGrpcWorkloadMember(
	tenantId, workloadId, instanceId string) (*computev1.WorkloadMember, error) {
	workloadMember := &computev1.WorkloadMember{
		Kind: computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
		Workload: &computev1.WorkloadResource{
			ResourceId: workloadId,
		},
		TenantId: tenantId,
	}

	if instanceId != "" {
		workloadMember.Instance = &computev1.InstanceResource{
			ResourceId: instanceId,
		}
	}

	if err := validator.ValidateMessage(workloadMember); err != nil {
		slog.Info("failed to serialize to workload member", "error", err)
		return nil, ErrInvalidWorkloadMembersInput
	}

	return workloadMember, nil
}

func (c *InventoryClient) validateWorkloadMemberResource(workloadMember *computev1.WorkloadMember) error {
	if workloadMember == nil {
		slog.Info("empty workload member resource")
		return ErrInvalidWorkloadMembers
	}

	if err := validator.ValidateMessage(workloadMember); err != nil {
		slog.Info("workload member resource proto validation failed", "error", err)
		return ErrInvalidWorkloadMembers
	}

	// the validation above does not check for an empty workload member id
	// at the moment we don't use the id, but consider addition validation once
	// the situation changes
	return nil
}

func (c *InventoryClient) createWorkloadMember(ctx context.Context, tenantId, workloadId, instanceId string) error {
	slog.Debug("createWorkloadMember", "tenantID", tenantId, "workload id", workloadId, "instance id", instanceId)

	childCtx, cancel := context.WithTimeout(ctx, defaultInventoryTimeout)
	defer cancel()

	workloadMemberGrpc, err := c.toGrpcWorkloadMember(tenantId, workloadId, instanceId)
	if err != nil {
		slog.Warn("failed to create workload member resource type", "tenantId", tenantId, "workload id",
			workloadId, "instanceId", instanceId, "error", err)
		return err
	}

	request := &inventoryv1.Resource{
		Resource: &inventoryv1.Resource_WorkloadMember{
			WorkloadMember: workloadMemberGrpc,
		},
	}

	respWorkloadMember, err := c.Client.Create(childCtx, tenantId, request)
	if err != nil {
		slog.Warn("failed to create workload member", "tenantId", tenantId, "workload id",
			workloadId, "instanceId", instanceId, "error", err)
		return ErrFailedInventoryCreateResource
	}

	if err := c.validateResource(respWorkloadMember); err != nil {
		slog.Warn("failed to validate resource", "error", err, "tenantID", tenantId, "workload id", workloadId, "instance id", instanceId)
		return ErrInvalidWorkloadMembers
	}

	invWorkloadMember := respWorkloadMember.GetWorkloadMember()
	if err = c.validateWorkloadMemberResource(invWorkloadMember); err != nil {
		slog.Warn("failed to validate workload member resource", "error", err, "tenantID", tenantId, "workload id", workloadId, "instance id", instanceId)
		return err
	}

	return nil
}

func (c *InventoryClient) deleteWorkloadMember(ctx context.Context, tenantId, workloadMemberId string) error {
	slog.Debug("deleteWorkloadMember", "tenantID", tenantId, "workload member id", workloadMemberId)

	childCtx, cancel := context.WithTimeout(ctx, defaultInventoryTimeout)
	defer cancel()

	if _, err := c.Client.Delete(childCtx, tenantId, workloadMemberId); err != nil {
		slog.Warn("failed to delete workload member", "tenantId", tenantId, "workload member id",
			workloadMemberId, "error", err)
		return ErrFailedInventoryDeleteResource
	}

	return nil
}

func (c *InventoryClient) deauthorizeHost(ctx context.Context, tenantId, hostUUID string) error {
	slog.Debug("deauthorizeHost", "tenantID", tenantId, "host id", hostUUID)

	childCtx, cancel := context.WithTimeout(ctx, defaultInventoryTimeout)
	defer cancel()

	host, err := c.getHost(childCtx, tenantId, hostUUID)
	if err != nil {
		slog.Warn("failed to get host by uuid", "tenantId", tenantId, "host uuid", hostUUID, "error", err)
		return err
	}
	slog.Debug("read host from inventory", "tenantId", tenantId, "host uuid", hostUUID, "host id", host.ResourceId)
	resource := &inventoryv1.Resource{
		Resource: &inventoryv1.Resource_Host{
			Host: &computev1.HostResource{
				ResourceId:   host.ResourceId,
				DesiredState: computev1.HostState_HOST_STATE_UNTRUSTED,
				Note:         "Deauthorized by Cluster API Provider Intel because the cluster was deleted",
			},
		},
	}

	fieldMask := &fieldmaskpb.FieldMask{
		Paths: []string{"desired_state"},
	}

	if _, err = c.Client.Update(childCtx, tenantId, host.ResourceId, fieldMask, resource); err != nil {
		slog.Warn("failed to deauthorize host", "tenantId", tenantId, "host uuid", hostUUID, "error", err)
		return ErrFailedInventoryGetResource
	}

	return nil
}
