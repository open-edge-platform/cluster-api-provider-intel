// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package inventory_stub

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestGetHostByUUID(t *testing.T) {
	client := GetStubClient().(*StubTenantAwareInventoryClient)
	host, err := client.GetHostByUUID(context.Background(), defaultTenantID, defaultHostID)
	require.NoError(t, err)
	require.NotNil(t, host)
	assert.Equal(t, defaultInstanceID, host.Instance.ResourceId)
	assert.Equal(t, "Linux", host.Instance.Os.Name)
	assert.Equal(t, "SN123456", host.SerialNumber)
}

func TestGet(t *testing.T) {
	client := GetStubClient().(*StubTenantAwareInventoryClient)
	resp, err := client.Get(context.Background(), defaultTenantID, defaultInstanceID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, defaultInstanceID, resp.Resource.GetInstance().ResourceId)
	assert.Equal(t, "Linux", resp.Resource.GetInstance().Os.Name)
}

func TestCreate(t *testing.T) {
	client := GetStubClient().(*StubTenantAwareInventoryClient)
	resource := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Workload{
			Workload: &computev1.WorkloadResource{
				Kind:       computev1.WorkloadKind_WORKLOAD_KIND_CLUSTER,
				Name:       defaultClusterName,
				ExternalId: defaultClusterName,
			},
		},
	}
	resp, err := client.Create(context.Background(), defaultTenantID, resource)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, defaultWorkloadID, resp.GetWorkload().ResourceId)
}

func TestDelete(t *testing.T) {
	client := GetStubClient().(*StubTenantAwareInventoryClient)

	// Delete Workload
	resp, err := client.Delete(context.Background(), defaultTenantID, defaultWorkloadID)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Delete WorkloadMember
	resp, err = client.Delete(context.Background(), defaultTenantID, defaultWorkloadMemberID)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestCreateWorkloadSemantics(t *testing.T) {
	client := GetStubClient().(*StubTenantAwareInventoryClient)
	resource := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Workload{
			Workload: &computev1.WorkloadResource{
				Kind:       computev1.WorkloadKind_WORKLOAD_KIND_CLUSTER,
				Name:       defaultClusterName,
				ExternalId: defaultClusterName,
			},
		},
	}
	resp, err := client.Create(context.Background(), defaultTenantID, resource)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, defaultWorkloadID, resp.GetWorkload().ResourceId)

	// Calling a second time should fail with "duplicate key" error
	resp, err = client.Create(context.Background(), defaultTenantID, resource)
	require.Error(t, err)
	require.Nil(t, resp)

	// Delete Workload
	resp2, err := client.Delete(context.Background(), defaultTenantID, defaultWorkloadID)
	require.NoError(t, err)
	require.NotNil(t, resp2)

	// Now create should succeed again
	resp, err = client.Create(context.Background(), defaultTenantID, resource)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, defaultWorkloadID, resp.GetWorkload().ResourceId)
}

func TestUpdate(t *testing.T) {
	client := GetStubClient().(*StubTenantAwareInventoryClient)
	resource := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Workload{
			Workload: &computev1.WorkloadResource{
				ResourceId: defaultWorkloadID,
			},
		},
	}
	fieldMask := &fieldmaskpb.FieldMask{}
	resp, err := client.Update(context.Background(), defaultTenantID, defaultWorkloadID, fieldMask, resource)

	// Not implemented
	require.Error(t, err)
	require.Nil(t, resp)

	// require.NoError(t, err)
	// require.NotNil(t, resp)
	// assert.Equal(t, defaultWorkloadID, resp.GetWorkload().ResourceId)
}

func TestList(t *testing.T) {
	client := GetStubClient().(*StubTenantAwareInventoryClient)
	resp, err := client.List(context.Background(), &inv_v1.ResourceFilter{})

	// Not implemented
	require.Error(t, err)
	require.Nil(t, resp)

	// require.NoError(t, err)
	// require.NotNil(t, resp)
}

func TestListAll(t *testing.T) {
	client := GetStubClient().(*StubTenantAwareInventoryClient)
	resp, err := client.ListAll(context.Background(), &inv_v1.ResourceFilter{})

	// Not implemented
	require.Error(t, err)
	require.Nil(t, resp)

	// require.NoError(t, err)
	// require.NotNil(t, resp)
}

func TestFind(t *testing.T) {
	client := GetStubClient().(*StubTenantAwareInventoryClient)
	resp, err := client.Find(context.Background(), &inv_v1.ResourceFilter{})

	// Not implemented
	require.Error(t, err)
	require.Nil(t, resp)

	// require.NoError(t, err)
	// require.NotNil(t, resp)
}

func TestFindAll(t *testing.T) {
	client := GetStubClient().(*StubTenantAwareInventoryClient)
	resp, err := client.FindAll(context.Background(), &inv_v1.ResourceFilter{})

	// Not implemented
	require.Error(t, err)
	require.Nil(t, resp)

	// require.NoError(t, err)
	// require.NotNil(t, resp)
}

func TestListInheritedTelemetryProfiles(t *testing.T) {
	client := GetStubClient().(*StubTenantAwareInventoryClient)
	resp, err := client.ListInheritedTelemetryProfiles(context.Background(), defaultTenantID, nil, "", "", 0, 0)

	// Not implemented
	require.Error(t, err)
	require.Nil(t, resp)

	// require.NoError(t, err)
	// require.NotNil(t, resp)
}

func TestGetTreeHierarchy(t *testing.T) {
	client := GetStubClient().(*StubTenantAwareInventoryClient)
	resp, err := client.GetTreeHierarchy(context.Background(), &inv_v1.GetTreeHierarchyRequest{})

	// Not implemented
	require.Error(t, err)
	require.Nil(t, resp)

	// require.NoError(t, err)
	// require.NotNil(t, resp)
}

func TestGetSitesPerRegion(t *testing.T) {
	client := GetStubClient().(*StubTenantAwareInventoryClient)
	resp, err := client.GetSitesPerRegion(context.Background(), &inv_v1.GetSitesPerRegionRequest{})

	// Not implemented
	require.Error(t, err)
	require.Nil(t, resp)

	// require.NoError(t, err)
	// require.NotNil(t, resp)
}

func TestDeleteAllResources(t *testing.T) {
	client := GetStubClient().(*StubTenantAwareInventoryClient)
	err := client.DeleteAllResources(
		context.Background(),
		defaultTenantID,
		inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD,
		true)
	require.NoError(t, err)
}
