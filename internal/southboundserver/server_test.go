// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package southboundserver

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/bnkamalesh/errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/undefinedlabs/go-mpatch"
	"google.golang.org/grpc"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	config "github.com/open-edge-platform/cluster-api-provider-intel/internal/southboundconfig"
	"github.com/open-edge-platform/cluster-api-provider-intel/internal/southboundhandler"
	"github.com/open-edge-platform/cluster-api-provider-intel/mocks/m_southboundhandler"
	pb "github.com/open-edge-platform/cluster-api-provider-intel/pkg/api/proto"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/rbac"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/tenant"
	testingutils "github.com/open-edge-platform/cluster-api-provider-intel/pkg/testing"
)

const (
	regoFilePath = "../rego/authz.rego"
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunGrpcServer(t *testing.T) {
	type args struct {
		cfg config.Config
	}

	TestRunGrpcServerNoErr := func(ctrl *gomock.Controller) []*mpatch.Patch {
		patchNetListen, patchErr := mpatch.PatchMethod(net.Listen, func(network string, address string) (net.Listener, error) {
			return nil, nil
		})
		if patchErr != nil {
			t.Errorf("patch error: %v", patchErr)
		}

		patchServe, patchErr := mpatch.PatchInstanceMethodByName(reflect.TypeOf(grpc.NewServer()), "Serve", func(this *grpc.Server, lis net.Listener) error {
			return nil
		})
		if patchErr != nil {
			t.Errorf("patch error: %v", patchErr)
		}

		patchContext, patchErr := mpatch.PatchMethod(signals.SetupSignalHandler, func() context.Context {
			return t.Context()
		})
		if patchErr != nil {
			t.Errorf("patch error: %v", patchErr)
		}

		patchHandler, patchErr := mpatch.PatchMethod(southboundhandler.NewHandler, func(context.Context, *rest.Config) (*southboundhandler.Handler, error) {
			return nil, nil
		})
		if patchErr != nil {
			t.Errorf("patch error: %v", patchErr)
		}

		return []*mpatch.Patch{patchNetListen, patchServe, patchContext, patchHandler}
	}

	tests := []struct {
		name           string
		args           args
		funcBeforeTest func(*gomock.Controller) []*mpatch.Patch
		wantErr        bool
	}{
		{
			name: "start grpc server",
			args: args{config.Config{
				GrpcAddr: "localhost",
				GrpcPort: "25000",
			}},
			funcBeforeTest: TestRunGrpcServerNoErr,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			if tt.funcBeforeTest != nil {
				plist := tt.funcBeforeTest(ctrl)
				for _, p := range plist {
					defer unpatch(t, p)
				}
			}
			gRPC, lis := NewGrpcServer(tt.args.cfg, regoFilePath)
			if gRPC == nil {
				t.Errorf("failed to create gRPC server")
			}
			if err := RunGrpcServer(gRPC, lis); (err != nil) != tt.wantErr {
				t.Errorf("RunGrpcServer() error = %v, expectedError %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewGrpcServerFail(t *testing.T) {
	type args struct {
		cfg config.Config
	}

	TestNewGrpcServerErr := func(ctrl *gomock.Controller) []*mpatch.Patch {
		patchNetListen, patchErr := mpatch.PatchMethod(net.Listen, func(network string, address string) (net.Listener, error) {
			return nil, fmt.Errorf("new grpc server error")
		})
		if patchErr != nil {
			t.Errorf("patch error: %v", patchErr)
		}

		patchContext, patchErr := mpatch.PatchMethod(signals.SetupSignalHandler, func() context.Context {
			return t.Context()
		})
		if patchErr != nil {
			t.Errorf("patch error: %v", patchErr)
		}

		patchHandler, patchErr := mpatch.PatchMethod(southboundhandler.NewHandler, func(context.Context, *rest.Config) (*southboundhandler.Handler, error) {
			return nil, nil
		})
		if patchErr != nil {
			t.Errorf("patch error: %v", patchErr)
		}

		return []*mpatch.Patch{patchNetListen, patchContext, patchHandler}
	}

	tests := []struct {
		name           string
		args           args
		funcBeforeTest func(*gomock.Controller) []*mpatch.Patch
		wantErr        bool
	}{
		{
			name: "net.Listen error",
			args: args{config.Config{
				GrpcAddr: "localhost",
				GrpcPort: "25000",
			}},
			funcBeforeTest: TestNewGrpcServerErr,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			if tt.funcBeforeTest != nil {
				plist := tt.funcBeforeTest(ctrl)
				for _, p := range plist {
					defer unpatch(t, p)
				}
			}
			srvr, listener := NewGrpcServer(tt.args.cfg, regoFilePath)
			if srvr != nil || listener != nil {
				t.Errorf("new grpc server expected to fail, but passed")
			}
		})
	}
}

func TestRunGrpcServerFail(t *testing.T) {
	type args struct {
		cfg config.Config
	}

	TestRunGrpcServerErr := func(ctrl *gomock.Controller) []*mpatch.Patch {
		patchNetListen, patchErr := mpatch.PatchMethod(net.Listen, func(network string, address string) (net.Listener, error) {
			return nil, nil
		})
		if patchErr != nil {
			t.Errorf("patch error: %v", patchErr)
		}

		patchServe, patchErr := mpatch.PatchInstanceMethodByName(reflect.TypeOf(grpc.NewServer()), "Serve", func(this *grpc.Server, lis net.Listener) error {
			return fmt.Errorf("start grpc server failed")
		})
		if patchErr != nil {
			t.Errorf("patch error: %v", patchErr)
		}

		patchContext, patchErr := mpatch.PatchMethod(signals.SetupSignalHandler, func() context.Context {
			return t.Context()
		})
		if patchErr != nil {
			t.Errorf("patch error: %v", patchErr)
		}

		patchHandler, patchErr := mpatch.PatchMethod(southboundhandler.NewHandler, func(context.Context, *rest.Config) (*southboundhandler.Handler, error) {
			return nil, nil
		})
		if patchErr != nil {
			t.Errorf("patch error: %v", patchErr)
		}

		return []*mpatch.Patch{patchNetListen, patchServe, patchContext, patchHandler}
	}
	tests := []struct {
		name           string
		args           args
		funcBeforeTest func(*gomock.Controller) []*mpatch.Patch
		wantErr        bool
	}{
		{
			name: "start grpc server",
			args: args{cfg: config.Config{
				GrpcAddr: "localhost",
				GrpcPort: "25000",
			}},
			funcBeforeTest: TestRunGrpcServerErr,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			if tt.funcBeforeTest != nil {
				plist := tt.funcBeforeTest(ctrl)
				for _, p := range plist {
					defer unpatch(t, p)
				}
			}
			gRPC, lis := NewGrpcServer(tt.args.cfg, regoFilePath)
			if gRPC == nil {
				t.Errorf("failed to create gRPC server")
			}
			if err := RunGrpcServer(gRPC, lis); (err != nil) != tt.wantErr {
				t.Errorf("RunGrpcServer() error = %v, expectedError %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegisterCluster(t *testing.T) {
	ctx, cancel := testingutils.CreateIncomingContextWithJWT(t)
	defer cancel()

	type args struct {
		ctx context.Context
		in  *pb.RegisterClusterRequest
	}

	opaPolicy, err := rbac.New(regoFilePath)
	require.Nil(t, err)

	mockHandler := m_southboundhandler.NewMockSouthboundHandler(t)

	testServer := &server{
		listen:  "0.0.0.0:50020",
		rbac:    opaPolicy,
		handler: mockHandler,
	}

	cases := []struct {
		name          string
		args          args
		expectedResp  pb.RegisterClusterResponse_Result
		expectedError string
		mocks         func() []*mock.Call
	}{
		{
			name:          "invalid guid",
			args:          args{ctx: ctx, in: &pb.RegisterClusterRequest{NodeGuid: "50bcc9a2-d1f-11ed-afa1-0242ac120002"}},
			expectedError: "invalid RegisterClusterRequest.NodeGuid: value does not match regex pattern \"^[{]?[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}[}]?$\"",
		},
		{
			name:         "HandleRegisterCluster failed",
			args:         args{ctx: ctx, in: &pb.RegisterClusterRequest{NodeGuid: "50bcc9a2-d1f2-11ed-afa1-0242ac120002"}},
			expectedResp: pb.RegisterClusterResponse_ERROR,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockHandler.On("Register", ctx, "50bcc9a2-d1f2-11ed-afa1-0242ac120002").
						Return(nil, nil, pb.RegisterClusterResponse_ERROR, errors.New("error message")).Once(),
				}
			},
		},
		{
			name:         "HandleRegisterCluster success",
			args:         args{ctx: ctx, in: &pb.RegisterClusterRequest{NodeGuid: "50bcc9a2-d1f2-11ed-afa1-0242ac120002"}},
			expectedResp: pb.RegisterClusterResponse_SUCCESS,
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockHandler.On("Register", ctx, "50bcc9a2-d1f2-11ed-afa1-0242ac120002").
						Return(&pb.ShellScriptCommand{}, &pb.ShellScriptCommand{}, pb.RegisterClusterResponse_SUCCESS, nil).Once(),
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			res, err := testServer.RegisterCluster(tc.args.ctx, tc.args.in)
			if res != nil {
				assert.Equal(t, tc.expectedResp, res.Res)
			}

			if tc.expectedError != "" {
				assert.ErrorContains(t, err, tc.expectedError)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestUpdateClusterStatus(t *testing.T) {
	ctx, cancel := testingutils.CreateIncomingContextWithJWT(t)
	ctx = tenant.AddActiveProjectIdToContext(ctx, tenant.DefaultProjectId)
	defer cancel()

	type args struct {
		ctx context.Context
		in  *pb.UpdateClusterStatusRequest
	}

	opaPolicy, err := rbac.New(regoFilePath)
	require.Nil(t, err)

	mockHandler := m_southboundhandler.NewMockSouthboundHandler(t)

	testServer := &server{
		listen:  "0.0.0.0:50020",
		rbac:    opaPolicy,
		handler: mockHandler,
	}

	cases := []struct {
		name          string
		args          args
		expectedError string
		mocks         func() []*mock.Call
	}{
		{
			name: "invalid guid",
			args: args{ctx: ctx, in: &pb.UpdateClusterStatusRequest{
				Code:     pb.UpdateClusterStatusRequest_INACTIVE,
				NodeGuid: "50bcc9a2-d1f-11ed-afa1-0242ac120002"},
			},
			expectedError: "invalid UpdateClusterStatusRequest.NodeGuid: value does not match regex pattern \"^[{]?[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}[}]?$\"",
		},
		{
			name: "HandleUpdateClusterStatus failure",
			args: args{ctx: ctx, in: &pb.UpdateClusterStatusRequest{
				Code:     pb.UpdateClusterStatusRequest_INACTIVE,
				NodeGuid: "50bcc9a2-d1f2-11ed-afa1-0242ac120002"},
			},
			expectedError: "HandleUpdateClusterStatus error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockHandler.On("UpdateStatus",
						ctx, "50bcc9a2-d1f2-11ed-afa1-0242ac120002", pb.UpdateClusterStatusRequest_INACTIVE).
						Return(pb.UpdateClusterStatusResponse_NONE, errors.NotFoundf("HandleUpdateClusterStatus error")).Once(),
				}
			},
		},
		{
			name: "HandleUpdateClusterStatus success",
			args: args{ctx: ctx, in: &pb.UpdateClusterStatusRequest{
				Code:     pb.UpdateClusterStatusRequest_ACTIVE,
				NodeGuid: "50bcc9a2-d1f2-11ed-afa1-0242ac120002"},
			},
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockHandler.On("UpdateStatus",
						ctx, "50bcc9a2-d1f2-11ed-afa1-0242ac120002", pb.UpdateClusterStatusRequest_ACTIVE).
						Return(pb.UpdateClusterStatusResponse_NONE, nil).Once(),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			// TODO: validate response - low priority
			_, err := testServer.UpdateClusterStatus(tc.args.ctx, tc.args.in)

			if tc.expectedError != "" {
				assert.ErrorContains(t, err, tc.expectedError)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
