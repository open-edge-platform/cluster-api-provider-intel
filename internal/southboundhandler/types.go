// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package southboundhandler

import (
	"context"

	pb "github.com/open-edge-platform/cluster-api-provider-intel/pkg/api/proto"
)

type SouthboundHandler interface {
	Register(ctx context.Context, nodeGUID string) (*pb.ShellScriptCommand, *pb.ShellScriptCommand, pb.RegisterClusterResponse_Result, error)
	UpdateStatus(ctx context.Context, nodeGUID string, status pb.UpdateClusterStatusRequest_Code) (pb.UpdateClusterStatusResponse_ActionRequest, error)
}
