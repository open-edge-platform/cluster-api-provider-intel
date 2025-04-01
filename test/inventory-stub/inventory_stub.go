// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package inventory_stub

import (
	"context"
	"errors"
	"fmt"
	"os"

	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"sigs.k8s.io/controller-runtime/pkg/log"

	computev1 "github.com/intel/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/intel/infra-core/inventory/v2/pkg/api/inventory/v1"
	osv1 "github.com/intel/infra-core/inventory/v2/pkg/api/os/v1"
	"github.com/intel/infra-core/inventory/v2/pkg/client"
	"github.com/intel/infra-core/inventory/v2/pkg/client/cache"
)

const (
	defaultTenantID         = "53cd37b9-66b2-4cc8-b080-3722ed7af64a"
	defaultHostID           = "12345678-1234-1234-1234-123456789012"
	defaultWorkloadID       = "workload-12345678"
	defaultWorkloadMemberID = "workloadmember-12345678"
	defaultInstanceID       = "inst-12345678"
	defaultClusterName      = "cluster-12345678"
	duplicateKeyError       = "constraint failed: ERROR: duplicate key value violates unique constraint \"workloadresource_external_id_tenant_id\" (SQLSTATE 23505)" // nolint:lll
)

var (
	workloadAllocated = false
	trace             = true
)

// StubResponse holds the response and error for a stubbed method call
type StubResponse struct {
	Response interface{}
	Error    error
}

// StubTenantAwareInventoryClient is a mock implementation of TenantAwareInventoryClient
type StubTenantAwareInventoryClient struct {
	StubbedResponses map[string]StubResponse
}

// Close is a stub implementation of the Close method
func (c *StubTenantAwareInventoryClient) Close() error {
	return nil
}

// GetHostByUUID returns a stubbed HostResource based on tenantID and uuid
func (c *StubTenantAwareInventoryClient) GetHostByUUID(
	ctx context.Context, tenantID, uuid string,
) (*computev1.HostResource, error) {
	log := log.FromContext(ctx)
	if trace {
		log.Info(fmt.Sprintf("inventory stub GetHostByUUID: UUID %s", uuid))
	}
	key := fmt.Sprintf("GetHostByUUID:%s:%s", tenantID, uuid)
	if stub, ok := c.StubbedResponses[key]; ok {
		return stub.Response.(*computev1.HostResource), stub.Error
	}
	return nil, fmt.Errorf("no stubbed response for key: %s", key)
}

// Get returns a stubbed GetResourceResponse based on tenantID and id
func (c *StubTenantAwareInventoryClient) Get(
	ctx context.Context, tenantID, id string,
) (*inv_v1.GetResourceResponse, error) {
	log := log.FromContext(ctx)
	if trace {
		log.Info(fmt.Sprintf("inventory stub Get: id %s", id))
	}
	key := fmt.Sprintf("Get:%s:%s", tenantID, id)
	if stub, ok := c.StubbedResponses[key]; ok {
		return stub.Response.(*inv_v1.GetResourceResponse), stub.Error
	}
	return nil, fmt.Errorf("no stubbed response for key: %s", key)
}

// Create returns a stubbed Resource based on tenantID and resource
func (c *StubTenantAwareInventoryClient) Create(
	ctx context.Context, tenantID string, res *inv_v1.Resource,
) (*inv_v1.Resource, error) {
	log := log.FromContext(ctx)
	if trace {
		log.Info(fmt.Sprintf("inventory stub Create: res %+v", res))
	}
	key := fmt.Sprintf("Create:%s:", tenantID)

	// Are we creating a workload?
	if workload := res.GetWorkload(); workload != nil {
		key += "workload"
		if stub, ok := c.StubbedResponses[key]; ok {
			if workloadAllocated {
				if trace {
					log.Info("inventory stub Create: duplicate key")
				}
				return nil, errors.New(duplicateKeyError)
			}

			// Mark workload as allocated.
			workloadAllocated = true
			if trace {
				log.Info("inventory stub Create: successfully created workload")
			}

			return stub.Response.(*inv_v1.Resource), stub.Error
		}
	}

	if workloadMember := res.GetWorkloadMember(); workloadMember != nil {
		// Creating a WorkloadMember. May want to revisit this.
		if trace {
			log.Info("inventory stub Create: successfully created workload_member")
		}
		key += "workloadMember"
		if stub, ok := c.StubbedResponses[key]; ok {
			return stub.Response.(*inv_v1.Resource), stub.Error
		}
	}

	return nil, fmt.Errorf("no stubbed response for key: %s", key)
}

// Delete returns a stubbed DeleteResourceResponse based on tenantID and id
func (c *StubTenantAwareInventoryClient) Delete(
	ctx context.Context, tenantID, id string,
) (*inv_v1.DeleteResourceResponse, error) {
	log := log.FromContext(ctx)
	if trace {
		log.Info(fmt.Sprintf("inventory stub Delete: id %s", id))
	}
	key := fmt.Sprintf("Delete:%s:%s", tenantID, id)
	if stub, ok := c.StubbedResponses[key]; ok {
		if id == defaultWorkloadID {
			workloadAllocated = false
		}
		return stub.Response.(*inv_v1.DeleteResourceResponse), stub.Error
	}
	return nil, fmt.Errorf("no stubbed response for key: %s", key)
}

// Update returns a stubbed Resource based on tenantID, id, field mask, and resource
func (c *StubTenantAwareInventoryClient) Update(
	ctx context.Context, tenantID, id string, fm *fieldmaskpb.FieldMask, res *inv_v1.Resource,
) (*inv_v1.Resource, error) {
	log := log.FromContext(ctx)
	if trace {
		log.Info(fmt.Sprintf("inventory stub Update: id %s", id))
	}
	key := fmt.Sprintf("Update:%s:%s", tenantID, id)
	if stub, ok := c.StubbedResponses[key]; ok {
		return stub.Response.(*inv_v1.Resource), stub.Error
	}
	return nil, fmt.Errorf("no stubbed response for key: %s", key)
}

// List returns a stubbed ListResourcesResponse based on the filter
func (c *StubTenantAwareInventoryClient) List(
	ctx context.Context, filter *inv_v1.ResourceFilter,
) (*inv_v1.ListResourcesResponse, error) {
	log := log.FromContext(ctx)
	if trace {
		log.Info("inventory stub List")
	}
	key := "List"
	if stub, ok := c.StubbedResponses[key]; ok {
		return stub.Response.(*inv_v1.ListResourcesResponse), stub.Error
	}
	return nil, fmt.Errorf("no stubbed response for key: %s", key)
}

// ListAll returns a stubbed list of Resources based on the filter
func (c *StubTenantAwareInventoryClient) ListAll(
	ctx context.Context, filter *inv_v1.ResourceFilter,
) ([]*inv_v1.Resource, error) {
	log := log.FromContext(ctx)
	if trace {
		log.Info("inventory stub ListAll")
	}
	key := "ListAll"
	if stub, ok := c.StubbedResponses[key]; ok {
		return stub.Response.([]*inv_v1.Resource), stub.Error
	}
	return nil, fmt.Errorf("no stubbed response for key: %s", key)
}

// Find returns a stubbed FindResourcesResponse based on the filter
func (c *StubTenantAwareInventoryClient) Find(
	ctx context.Context, filter *inv_v1.ResourceFilter,
) (*inv_v1.FindResourcesResponse, error) {
	log := log.FromContext(ctx)
	if trace {
		log.Info("inventory stub Find")
	}
	key := "Find"
	if stub, ok := c.StubbedResponses[key]; ok {
		return stub.Response.(*inv_v1.FindResourcesResponse), stub.Error
	}
	return nil, fmt.Errorf("no stubbed response for key: %s", key)
}

// FindAll returns a stubbed list of ResourceTenantIDCarrier based on the filter
func (c *StubTenantAwareInventoryClient) FindAll(
	ctx context.Context, filter *inv_v1.ResourceFilter,
) ([]*client.ResourceTenantIDCarrier, error) {
	log := log.FromContext(ctx)
	if trace {
		log.Info("inventory stub FindAll")
	}
	key := "FindAll"
	if stub, ok := c.StubbedResponses[key]; ok {
		return stub.Response.([]*client.ResourceTenantIDCarrier), stub.Error
	}
	return nil, fmt.Errorf("no stubbed response for key: %s", key)
}

// UpdateSubscriptions is a stub implementation of the UpdateSubscriptions method
func (c *StubTenantAwareInventoryClient) UpdateSubscriptions(
	ctx context.Context, tenantID string, kinds []inv_v1.ResourceKind,
) error {
	log := log.FromContext(ctx)
	if trace {
		log.Info("inventory stub UpdateSubscriptions")
	}
	return nil
}

// ListInheritedTelemetryProfiles returns a stubbed ListInheritedTelemetryProfilesResponse
func (c *StubTenantAwareInventoryClient) ListInheritedTelemetryProfiles(
	ctx context.Context, tenantID string, inheritBy *inv_v1.ListInheritedTelemetryProfilesRequest_InheritBy,
	filter, orderBy string, limit, offset uint32,
) (*inv_v1.ListInheritedTelemetryProfilesResponse, error) {
	log := log.FromContext(ctx)
	if trace {
		log.Info("inventory stub LisInheritedTelemetryProfiles")
	}
	key := "ListInheritedTelemetryProfiles"
	if stub, ok := c.StubbedResponses[key]; ok {
		return stub.Response.(*inv_v1.ListInheritedTelemetryProfilesResponse), stub.Error
	}
	return nil, fmt.Errorf("no stubbed response for key: %s", key)
}

// GetTreeHierarchy returns a stubbed list of TreeNodes based on the request
func (c *StubTenantAwareInventoryClient) GetTreeHierarchy(
	ctx context.Context, request *inv_v1.GetTreeHierarchyRequest,
) ([]*inv_v1.GetTreeHierarchyResponse_TreeNode, error) {
	log := log.FromContext(ctx)
	if trace {
		log.Info("inventory stub GetTreeHierarchy")
	}
	key := "GetTreeHierarchy"
	if stub, ok := c.StubbedResponses[key]; ok {
		return stub.Response.([]*inv_v1.GetTreeHierarchyResponse_TreeNode), stub.Error
	}
	return nil, fmt.Errorf("no stubbed response for key: %s", key)
}

// GetSitesPerRegion returns a stubbed GetSitesPerRegionResponse based on the request
func (c *StubTenantAwareInventoryClient) GetSitesPerRegion(
	ctx context.Context, request *inv_v1.GetSitesPerRegionRequest,
) (*inv_v1.GetSitesPerRegionResponse, error) {
	log := log.FromContext(ctx)
	if trace {
		log.Info("inventory stub GetSitesPerRegion")
	}
	key := "GetSitesPerRegion"
	if stub, ok := c.StubbedResponses[key]; ok {
		return stub.Response.(*inv_v1.GetSitesPerRegionResponse), stub.Error
	}
	return nil, fmt.Errorf("no stubbed response for key: %s", key)
}

// TestingOnlySetClient is a stub implementation of the TestingOnlySetClient method
func (c *StubTenantAwareInventoryClient) TestingOnlySetClient(
	client inv_v1.InventoryServiceClient,
) {
	// Stub implementation
}

// TestGetClientCache is a stub implementation of the TestGetClientCache method
func (c *StubTenantAwareInventoryClient) TestGetClientCache() *cache.InventoryCache {
	return nil
}

// TestGetClientCacheUUID is a stub implementation of the TestGetClientCacheUUID method
func (c *StubTenantAwareInventoryClient) TestGetClientCacheUUID() *cache.InventoryCache {
	return nil
}

// DeleteAllResources is a stub implementation of the DeleteAllResources method
func (c *StubTenantAwareInventoryClient) DeleteAllResources(
	ctx context.Context, tenantID string, kind inv_v1.ResourceKind, enforce bool,
) error {
	log := log.FromContext(ctx)
	if trace {
		log.Info("inventory stub DeleteAllResources")
	}
	return nil
}

// GetStubClient initializes the stub client with default responses and returns it
func GetStubClient() client.TenantAwareInventoryClient {
	tenantID := os.Getenv("TENANT_ID")
	if tenantID == "" {
		tenantID = defaultTenantID
	}
	hostID := os.Getenv("HOST_ID")
	if hostID == "" {
		hostID = defaultHostID
	}

	stubClient := &StubTenantAwareInventoryClient{
		StubbedResponses: make(map[string]StubResponse),
	}

	// Populate stubbed responses for GetHostByUUID
	stubClient.StubbedResponses[fmt.Sprintf("GetHostByUUID:%s:%s", tenantID, hostID)] = StubResponse{
		Response: &computev1.HostResource{
			Instance: &computev1.InstanceResource{
				ResourceId: defaultInstanceID,
				CurrentOs: &osv1.OperatingSystemResource{
					Name: "Linux",
				},
			},
			SerialNumber: "SN123456",
		},
		Error: nil,
	}

	// Populate stubbed responses for Get
	stubClient.StubbedResponses[fmt.Sprintf("Get:%s:%s", tenantID, defaultInstanceID)] = StubResponse{
		Response: &inv_v1.GetResourceResponse{
			Resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Instance{
					Instance: &computev1.InstanceResource{
						ResourceId: defaultInstanceID,
						CurrentOs: &osv1.OperatingSystemResource{
							Name: "Linux",
						},
						WorkloadMembers: []*computev1.WorkloadMember{
							{
								ResourceId: defaultWorkloadMemberID,
								Kind:       computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
								Workload: &computev1.WorkloadResource{
									ResourceId: defaultWorkloadID,
									Kind:       computev1.WorkloadKind_WORKLOAD_KIND_CLUSTER,
								},
							},
						},
					},
				},
			},
		},
		Error: nil,
	}

	// Populate stubbed responses for Create
	stubClient.StubbedResponses[fmt.Sprintf("Create:%s:workload", tenantID)] = StubResponse{
		Response: &inv_v1.Resource{
			Resource: &inv_v1.Resource_Workload{
				Workload: &computev1.WorkloadResource{
					ResourceId: defaultWorkloadID,
				},
			},
		},
		Error: nil,
	}

	stubClient.StubbedResponses[fmt.Sprintf("Create:%s:workloadMember", tenantID)] = StubResponse{
		Response: &inv_v1.Resource{
			Resource: &inv_v1.Resource_WorkloadMember{
				WorkloadMember: &computev1.WorkloadMember{
					ResourceId: defaultWorkloadMemberID,
					Workload: &computev1.WorkloadResource{
						ResourceId: defaultWorkloadID,
					},
				},
			},
		},
		Error: nil,
	}

	// Populate stubbed responses for Delete
	stubClient.StubbedResponses[fmt.Sprintf("Delete:%s:%s", tenantID, defaultWorkloadID)] = StubResponse{
		Response: &inv_v1.DeleteResourceResponse{},
		Error:    nil,
	}
	stubClient.StubbedResponses[fmt.Sprintf("Delete:%s:%s", tenantID, defaultWorkloadMemberID)] = StubResponse{
		Response: &inv_v1.DeleteResourceResponse{},
		Error:    nil,
	}
	return stubClient
}
