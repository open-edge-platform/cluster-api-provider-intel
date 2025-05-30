// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.3
// 	protoc        v5.27.1
// source: cluster_orchestrator_southbound.proto

package cluster_orchestrator_southbound

import (
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type RegisterClusterResponse_Result int32

const (
	RegisterClusterResponse_SUCCESS RegisterClusterResponse_Result = 0
	RegisterClusterResponse_ERROR   RegisterClusterResponse_Result = 1
)

// Enum value maps for RegisterClusterResponse_Result.
var (
	RegisterClusterResponse_Result_name = map[int32]string{
		0: "SUCCESS",
		1: "ERROR",
	}
	RegisterClusterResponse_Result_value = map[string]int32{
		"SUCCESS": 0,
		"ERROR":   1,
	}
)

func (x RegisterClusterResponse_Result) Enum() *RegisterClusterResponse_Result {
	p := new(RegisterClusterResponse_Result)
	*p = x
	return p
}

func (x RegisterClusterResponse_Result) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (RegisterClusterResponse_Result) Descriptor() protoreflect.EnumDescriptor {
	return file_cluster_orchestrator_southbound_proto_enumTypes[0].Descriptor()
}

func (RegisterClusterResponse_Result) Type() protoreflect.EnumType {
	return &file_cluster_orchestrator_southbound_proto_enumTypes[0]
}

func (x RegisterClusterResponse_Result) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use RegisterClusterResponse_Result.Descriptor instead.
func (RegisterClusterResponse_Result) EnumDescriptor() ([]byte, []int) {
	return file_cluster_orchestrator_southbound_proto_rawDescGZIP(), []int{1, 0}
}

type UpdateClusterStatusRequest_Code int32

const (
	UpdateClusterStatusRequest_INACTIVE              UpdateClusterStatusRequest_Code = 0
	UpdateClusterStatusRequest_REGISTERING           UpdateClusterStatusRequest_Code = 1
	UpdateClusterStatusRequest_INSTALL_IN_PROGRESS   UpdateClusterStatusRequest_Code = 2
	UpdateClusterStatusRequest_ACTIVE                UpdateClusterStatusRequest_Code = 3
	UpdateClusterStatusRequest_DEREGISTERING         UpdateClusterStatusRequest_Code = 4
	UpdateClusterStatusRequest_UNINSTALL_IN_PROGRESS UpdateClusterStatusRequest_Code = 5
	UpdateClusterStatusRequest_ERROR                 UpdateClusterStatusRequest_Code = 6
)

// Enum value maps for UpdateClusterStatusRequest_Code.
var (
	UpdateClusterStatusRequest_Code_name = map[int32]string{
		0: "INACTIVE",
		1: "REGISTERING",
		2: "INSTALL_IN_PROGRESS",
		3: "ACTIVE",
		4: "DEREGISTERING",
		5: "UNINSTALL_IN_PROGRESS",
		6: "ERROR",
	}
	UpdateClusterStatusRequest_Code_value = map[string]int32{
		"INACTIVE":              0,
		"REGISTERING":           1,
		"INSTALL_IN_PROGRESS":   2,
		"ACTIVE":                3,
		"DEREGISTERING":         4,
		"UNINSTALL_IN_PROGRESS": 5,
		"ERROR":                 6,
	}
)

func (x UpdateClusterStatusRequest_Code) Enum() *UpdateClusterStatusRequest_Code {
	p := new(UpdateClusterStatusRequest_Code)
	*p = x
	return p
}

func (x UpdateClusterStatusRequest_Code) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (UpdateClusterStatusRequest_Code) Descriptor() protoreflect.EnumDescriptor {
	return file_cluster_orchestrator_southbound_proto_enumTypes[1].Descriptor()
}

func (UpdateClusterStatusRequest_Code) Type() protoreflect.EnumType {
	return &file_cluster_orchestrator_southbound_proto_enumTypes[1]
}

func (x UpdateClusterStatusRequest_Code) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use UpdateClusterStatusRequest_Code.Descriptor instead.
func (UpdateClusterStatusRequest_Code) EnumDescriptor() ([]byte, []int) {
	return file_cluster_orchestrator_southbound_proto_rawDescGZIP(), []int{3, 0}
}

type UpdateClusterStatusResponse_ActionRequest int32

const (
	UpdateClusterStatusResponse_NONE       UpdateClusterStatusResponse_ActionRequest = 0
	UpdateClusterStatusResponse_REGISTER   UpdateClusterStatusResponse_ActionRequest = 1
	UpdateClusterStatusResponse_DEREGISTER UpdateClusterStatusResponse_ActionRequest = 2
)

// Enum value maps for UpdateClusterStatusResponse_ActionRequest.
var (
	UpdateClusterStatusResponse_ActionRequest_name = map[int32]string{
		0: "NONE",
		1: "REGISTER",
		2: "DEREGISTER",
	}
	UpdateClusterStatusResponse_ActionRequest_value = map[string]int32{
		"NONE":       0,
		"REGISTER":   1,
		"DEREGISTER": 2,
	}
)

func (x UpdateClusterStatusResponse_ActionRequest) Enum() *UpdateClusterStatusResponse_ActionRequest {
	p := new(UpdateClusterStatusResponse_ActionRequest)
	*p = x
	return p
}

func (x UpdateClusterStatusResponse_ActionRequest) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (UpdateClusterStatusResponse_ActionRequest) Descriptor() protoreflect.EnumDescriptor {
	return file_cluster_orchestrator_southbound_proto_enumTypes[2].Descriptor()
}

func (UpdateClusterStatusResponse_ActionRequest) Type() protoreflect.EnumType {
	return &file_cluster_orchestrator_southbound_proto_enumTypes[2]
}

func (x UpdateClusterStatusResponse_ActionRequest) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use UpdateClusterStatusResponse_ActionRequest.Descriptor instead.
func (UpdateClusterStatusResponse_ActionRequest) EnumDescriptor() ([]byte, []int) {
	return file_cluster_orchestrator_southbound_proto_rawDescGZIP(), []int{4, 0}
}

type GetClusterNumByTemplateIdentifierResponse_Result int32

const (
	GetClusterNumByTemplateIdentifierResponse_SUCCESS GetClusterNumByTemplateIdentifierResponse_Result = 0
	GetClusterNumByTemplateIdentifierResponse_ERROR   GetClusterNumByTemplateIdentifierResponse_Result = 1
)

// Enum value maps for GetClusterNumByTemplateIdentifierResponse_Result.
var (
	GetClusterNumByTemplateIdentifierResponse_Result_name = map[int32]string{
		0: "SUCCESS",
		1: "ERROR",
	}
	GetClusterNumByTemplateIdentifierResponse_Result_value = map[string]int32{
		"SUCCESS": 0,
		"ERROR":   1,
	}
)

func (x GetClusterNumByTemplateIdentifierResponse_Result) Enum() *GetClusterNumByTemplateIdentifierResponse_Result {
	p := new(GetClusterNumByTemplateIdentifierResponse_Result)
	*p = x
	return p
}

func (x GetClusterNumByTemplateIdentifierResponse_Result) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (GetClusterNumByTemplateIdentifierResponse_Result) Descriptor() protoreflect.EnumDescriptor {
	return file_cluster_orchestrator_southbound_proto_enumTypes[3].Descriptor()
}

func (GetClusterNumByTemplateIdentifierResponse_Result) Type() protoreflect.EnumType {
	return &file_cluster_orchestrator_southbound_proto_enumTypes[3]
}

func (x GetClusterNumByTemplateIdentifierResponse_Result) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use GetClusterNumByTemplateIdentifierResponse_Result.Descriptor instead.
func (GetClusterNumByTemplateIdentifierResponse_Result) EnumDescriptor() ([]byte, []int) {
	return file_cluster_orchestrator_southbound_proto_rawDescGZIP(), []int{6, 0}
}

// RegisterClusterRequest contains Edge Node identity assigned by Inventory
type RegisterClusterRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	NodeGuid      string                 `protobuf:"bytes,1,opt,name=node_guid,json=nodeGuid,proto3" json:"node_guid,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RegisterClusterRequest) Reset() {
	*x = RegisterClusterRequest{}
	mi := &file_cluster_orchestrator_southbound_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RegisterClusterRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterClusterRequest) ProtoMessage() {}

func (x *RegisterClusterRequest) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_orchestrator_southbound_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterClusterRequest.ProtoReflect.Descriptor instead.
func (*RegisterClusterRequest) Descriptor() ([]byte, []int) {
	return file_cluster_orchestrator_southbound_proto_rawDescGZIP(), []int{0}
}

func (x *RegisterClusterRequest) GetNodeGuid() string {
	if x != nil {
		return x.NodeGuid
	}
	return ""
}

// RegisterClusterResponse contains shell script to be executed by Cluster Agent to install cluster
type RegisterClusterResponse struct {
	state         protoimpl.MessageState         `protogen:"open.v1"`
	InstallCmd    *ShellScriptCommand            `protobuf:"bytes,1,opt,name=install_cmd,json=installCmd,proto3" json:"install_cmd,omitempty"`
	UninstallCmd  *ShellScriptCommand            `protobuf:"bytes,2,opt,name=uninstall_cmd,json=uninstallCmd,proto3" json:"uninstall_cmd,omitempty"`
	Res           RegisterClusterResponse_Result `protobuf:"varint,3,opt,name=res,proto3,enum=cluster_orchestrator_southbound_proto.RegisterClusterResponse_Result" json:"res,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RegisterClusterResponse) Reset() {
	*x = RegisterClusterResponse{}
	mi := &file_cluster_orchestrator_southbound_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RegisterClusterResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterClusterResponse) ProtoMessage() {}

func (x *RegisterClusterResponse) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_orchestrator_southbound_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterClusterResponse.ProtoReflect.Descriptor instead.
func (*RegisterClusterResponse) Descriptor() ([]byte, []int) {
	return file_cluster_orchestrator_southbound_proto_rawDescGZIP(), []int{1}
}

func (x *RegisterClusterResponse) GetInstallCmd() *ShellScriptCommand {
	if x != nil {
		return x.InstallCmd
	}
	return nil
}

func (x *RegisterClusterResponse) GetUninstallCmd() *ShellScriptCommand {
	if x != nil {
		return x.UninstallCmd
	}
	return nil
}

func (x *RegisterClusterResponse) GetRes() RegisterClusterResponse_Result {
	if x != nil {
		return x.Res
	}
	return RegisterClusterResponse_SUCCESS
}

// ShellScriptCommand is a command to be executed by Cluster Agent to install/uninstall LKPE.
// command is to be executed in shell, like this `sh -c command`
type ShellScriptCommand struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// example1: "curl -fL https://DOMAIN.NAME/system-agent-install.sh | sudo  sh -s - --server https://DOMAIN.NAME --label 'cattle.io/os=linux' --token 86f9cqfnvvlmwmvvmsptmr5wqj9d6bqpxkmxbvjw2txklhbglcdtff --ca-checksum b50da8bfa2cbcc13e209b9ffbab4b39c699e0aa2b3fe50f44ec4477c54725ea3 --etcd --controlplane --worker"
	// example2: "/usr/local/bin/rancher-system-agent-uninstall.sh; /usr/local/bin/rke2-uninstall.sh"
	Command       string `protobuf:"bytes,1,opt,name=command,proto3" json:"command,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ShellScriptCommand) Reset() {
	*x = ShellScriptCommand{}
	mi := &file_cluster_orchestrator_southbound_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ShellScriptCommand) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShellScriptCommand) ProtoMessage() {}

func (x *ShellScriptCommand) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_orchestrator_southbound_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShellScriptCommand.ProtoReflect.Descriptor instead.
func (*ShellScriptCommand) Descriptor() ([]byte, []int) {
	return file_cluster_orchestrator_southbound_proto_rawDescGZIP(), []int{2}
}

func (x *ShellScriptCommand) GetCommand() string {
	if x != nil {
		return x.Command
	}
	return ""
}

// UpdateClusterStatusRequest is used by Cluster Agent to represent its internal state machine
type UpdateClusterStatusRequest struct {
	state         protoimpl.MessageState          `protogen:"open.v1"`
	Code          UpdateClusterStatusRequest_Code `protobuf:"varint,1,opt,name=code,proto3,enum=cluster_orchestrator_southbound_proto.UpdateClusterStatusRequest_Code" json:"code,omitempty"`
	NodeGuid      string                          `protobuf:"bytes,2,opt,name=node_guid,json=nodeGuid,proto3" json:"node_guid,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateClusterStatusRequest) Reset() {
	*x = UpdateClusterStatusRequest{}
	mi := &file_cluster_orchestrator_southbound_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateClusterStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateClusterStatusRequest) ProtoMessage() {}

func (x *UpdateClusterStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_orchestrator_southbound_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateClusterStatusRequest.ProtoReflect.Descriptor instead.
func (*UpdateClusterStatusRequest) Descriptor() ([]byte, []int) {
	return file_cluster_orchestrator_southbound_proto_rawDescGZIP(), []int{3}
}

func (x *UpdateClusterStatusRequest) GetCode() UpdateClusterStatusRequest_Code {
	if x != nil {
		return x.Code
	}
	return UpdateClusterStatusRequest_INACTIVE
}

func (x *UpdateClusterStatusRequest) GetNodeGuid() string {
	if x != nil {
		return x.NodeGuid
	}
	return ""
}

// UpdateClusterStatusResponse is used to request Cluster Agent to transition to new internal state
type UpdateClusterStatusResponse struct {
	state         protoimpl.MessageState                    `protogen:"open.v1"`
	ActionRequest UpdateClusterStatusResponse_ActionRequest `protobuf:"varint,1,opt,name=action_request,json=actionRequest,proto3,enum=cluster_orchestrator_southbound_proto.UpdateClusterStatusResponse_ActionRequest" json:"action_request,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateClusterStatusResponse) Reset() {
	*x = UpdateClusterStatusResponse{}
	mi := &file_cluster_orchestrator_southbound_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateClusterStatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateClusterStatusResponse) ProtoMessage() {}

func (x *UpdateClusterStatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_orchestrator_southbound_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateClusterStatusResponse.ProtoReflect.Descriptor instead.
func (*UpdateClusterStatusResponse) Descriptor() ([]byte, []int) {
	return file_cluster_orchestrator_southbound_proto_rawDescGZIP(), []int{4}
}

func (x *UpdateClusterStatusResponse) GetActionRequest() UpdateClusterStatusResponse_ActionRequest {
	if x != nil {
		return x.ActionRequest
	}
	return UpdateClusterStatusResponse_NONE
}

type GetClusterNumByTemplateIdentifierRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// it's format is "template name" + "-" + "template version"
	TemplateIdentifier string `protobuf:"bytes,1,opt,name=templateIdentifier,proto3" json:"templateIdentifier,omitempty"`
	unknownFields      protoimpl.UnknownFields
	sizeCache          protoimpl.SizeCache
}

func (x *GetClusterNumByTemplateIdentifierRequest) Reset() {
	*x = GetClusterNumByTemplateIdentifierRequest{}
	mi := &file_cluster_orchestrator_southbound_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetClusterNumByTemplateIdentifierRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetClusterNumByTemplateIdentifierRequest) ProtoMessage() {}

func (x *GetClusterNumByTemplateIdentifierRequest) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_orchestrator_southbound_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetClusterNumByTemplateIdentifierRequest.ProtoReflect.Descriptor instead.
func (*GetClusterNumByTemplateIdentifierRequest) Descriptor() ([]byte, []int) {
	return file_cluster_orchestrator_southbound_proto_rawDescGZIP(), []int{5}
}

func (x *GetClusterNumByTemplateIdentifierRequest) GetTemplateIdentifier() string {
	if x != nil {
		return x.TemplateIdentifier
	}
	return ""
}

type GetClusterNumByTemplateIdentifierResponse struct {
	state         protoimpl.MessageState                           `protogen:"open.v1"`
	Res           GetClusterNumByTemplateIdentifierResponse_Result `protobuf:"varint,1,opt,name=res,proto3,enum=cluster_orchestrator_southbound_proto.GetClusterNumByTemplateIdentifierResponse_Result" json:"res,omitempty"`
	ClusterNum    int32                                            `protobuf:"varint,2,opt,name=clusterNum,proto3" json:"clusterNum,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetClusterNumByTemplateIdentifierResponse) Reset() {
	*x = GetClusterNumByTemplateIdentifierResponse{}
	mi := &file_cluster_orchestrator_southbound_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetClusterNumByTemplateIdentifierResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetClusterNumByTemplateIdentifierResponse) ProtoMessage() {}

func (x *GetClusterNumByTemplateIdentifierResponse) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_orchestrator_southbound_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetClusterNumByTemplateIdentifierResponse.ProtoReflect.Descriptor instead.
func (*GetClusterNumByTemplateIdentifierResponse) Descriptor() ([]byte, []int) {
	return file_cluster_orchestrator_southbound_proto_rawDescGZIP(), []int{6}
}

func (x *GetClusterNumByTemplateIdentifierResponse) GetRes() GetClusterNumByTemplateIdentifierResponse_Result {
	if x != nil {
		return x.Res
	}
	return GetClusterNumByTemplateIdentifierResponse_SUCCESS
}

func (x *GetClusterNumByTemplateIdentifierResponse) GetClusterNum() int32 {
	if x != nil {
		return x.ClusterNum
	}
	return 0
}

var File_cluster_orchestrator_southbound_proto protoreflect.FileDescriptor

var file_cluster_orchestrator_southbound_proto_rawDesc = []byte{
	0x0a, 0x25, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x5f, 0x6f, 0x72, 0x63, 0x68, 0x65, 0x73,
	0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x73, 0x6f, 0x75, 0x74, 0x68, 0x62, 0x6f, 0x75, 0x6e,
	0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x25, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72,
	0x5f, 0x6f, 0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x73, 0x6f,
	0x75, 0x74, 0x68, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17,
	0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x7a, 0x0a, 0x16, 0x52, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x65, 0x72, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x60, 0x0a, 0x09, 0x6e, 0x6f, 0x64, 0x65, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x43, 0xfa, 0x42, 0x40, 0x72, 0x3e, 0x32, 0x3c, 0x5e, 0x5b, 0x7b,
	0x5d, 0x3f, 0x5b, 0x30, 0x2d, 0x39, 0x61, 0x2d, 0x66, 0x41, 0x2d, 0x46, 0x5d, 0x7b, 0x38, 0x7d,
	0x2d, 0x28, 0x5b, 0x30, 0x2d, 0x39, 0x61, 0x2d, 0x66, 0x41, 0x2d, 0x46, 0x5d, 0x7b, 0x34, 0x7d,
	0x2d, 0x29, 0x7b, 0x33, 0x7d, 0x5b, 0x30, 0x2d, 0x39, 0x61, 0x2d, 0x66, 0x41, 0x2d, 0x46, 0x5d,
	0x7b, 0x31, 0x32, 0x7d, 0x5b, 0x7d, 0x5d, 0x3f, 0x24, 0x52, 0x08, 0x6e, 0x6f, 0x64, 0x65, 0x47,
	0x75, 0x69, 0x64, 0x22, 0xd0, 0x02, 0x0a, 0x17, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72,
	0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x5a, 0x0a, 0x0b, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x5f, 0x63, 0x6d, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x39, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x5f, 0x6f,
	0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x73, 0x6f, 0x75, 0x74,
	0x68, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x68, 0x65,
	0x6c, 0x6c, 0x53, 0x63, 0x72, 0x69, 0x70, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x52,
	0x0a, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x43, 0x6d, 0x64, 0x12, 0x5e, 0x0a, 0x0d, 0x75,
	0x6e, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x5f, 0x63, 0x6d, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x39, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x5f, 0x6f, 0x72, 0x63,
	0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x73, 0x6f, 0x75, 0x74, 0x68, 0x62,
	0x6f, 0x75, 0x6e, 0x64, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x68, 0x65, 0x6c, 0x6c,
	0x53, 0x63, 0x72, 0x69, 0x70, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x52, 0x0c, 0x75,
	0x6e, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x43, 0x6d, 0x64, 0x12, 0x57, 0x0a, 0x03, 0x72,
	0x65, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x45, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74,
	0x65, 0x72, 0x5f, 0x6f, 0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f,
	0x73, 0x6f, 0x75, 0x74, 0x68, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52,
	0x03, 0x72, 0x65, 0x73, 0x22, 0x20, 0x0a, 0x06, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x0b,
	0x0a, 0x07, 0x53, 0x55, 0x43, 0x43, 0x45, 0x53, 0x53, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x45,
	0x52, 0x52, 0x4f, 0x52, 0x10, 0x01, 0x22, 0x2e, 0x0a, 0x12, 0x53, 0x68, 0x65, 0x6c, 0x6c, 0x53,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x18, 0x0a, 0x07,
	0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63,
	0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x22, 0xe0, 0x02, 0x0a, 0x1a, 0x55, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x5a, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x46, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x5f, 0x6f, 0x72,
	0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x73, 0x6f, 0x75, 0x74, 0x68,
	0x62, 0x6f, 0x75, 0x6e, 0x64, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x04, 0x63, 0x6f, 0x64,
	0x65, 0x12, 0x60, 0x0a, 0x09, 0x6e, 0x6f, 0x64, 0x65, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x43, 0xfa, 0x42, 0x40, 0x72, 0x3e, 0x32, 0x3c, 0x5e, 0x5b, 0x7b,
	0x5d, 0x3f, 0x5b, 0x30, 0x2d, 0x39, 0x61, 0x2d, 0x66, 0x41, 0x2d, 0x46, 0x5d, 0x7b, 0x38, 0x7d,
	0x2d, 0x28, 0x5b, 0x30, 0x2d, 0x39, 0x61, 0x2d, 0x66, 0x41, 0x2d, 0x46, 0x5d, 0x7b, 0x34, 0x7d,
	0x2d, 0x29, 0x7b, 0x33, 0x7d, 0x5b, 0x30, 0x2d, 0x39, 0x61, 0x2d, 0x66, 0x41, 0x2d, 0x46, 0x5d,
	0x7b, 0x31, 0x32, 0x7d, 0x5b, 0x7d, 0x5d, 0x3f, 0x24, 0x52, 0x08, 0x6e, 0x6f, 0x64, 0x65, 0x47,
	0x75, 0x69, 0x64, 0x22, 0x83, 0x01, 0x0a, 0x04, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x0c, 0x0a, 0x08,
	0x49, 0x4e, 0x41, 0x43, 0x54, 0x49, 0x56, 0x45, 0x10, 0x00, 0x12, 0x0f, 0x0a, 0x0b, 0x52, 0x45,
	0x47, 0x49, 0x53, 0x54, 0x45, 0x52, 0x49, 0x4e, 0x47, 0x10, 0x01, 0x12, 0x17, 0x0a, 0x13, 0x49,
	0x4e, 0x53, 0x54, 0x41, 0x4c, 0x4c, 0x5f, 0x49, 0x4e, 0x5f, 0x50, 0x52, 0x4f, 0x47, 0x52, 0x45,
	0x53, 0x53, 0x10, 0x02, 0x12, 0x0a, 0x0a, 0x06, 0x41, 0x43, 0x54, 0x49, 0x56, 0x45, 0x10, 0x03,
	0x12, 0x11, 0x0a, 0x0d, 0x44, 0x45, 0x52, 0x45, 0x47, 0x49, 0x53, 0x54, 0x45, 0x52, 0x49, 0x4e,
	0x47, 0x10, 0x04, 0x12, 0x19, 0x0a, 0x15, 0x55, 0x4e, 0x49, 0x4e, 0x53, 0x54, 0x41, 0x4c, 0x4c,
	0x5f, 0x49, 0x4e, 0x5f, 0x50, 0x52, 0x4f, 0x47, 0x52, 0x45, 0x53, 0x53, 0x10, 0x05, 0x12, 0x09,
	0x0a, 0x05, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x10, 0x06, 0x22, 0xcf, 0x01, 0x0a, 0x1b, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x77, 0x0a, 0x0e, 0x61, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x50, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x5f, 0x6f, 0x72, 0x63, 0x68,
	0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x73, 0x6f, 0x75, 0x74, 0x68, 0x62, 0x6f,
	0x75, 0x6e, 0x64, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x52, 0x0d, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x22, 0x37, 0x0a, 0x0d, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x08, 0x0a, 0x04, 0x4e, 0x4f, 0x4e, 0x45, 0x10, 0x00, 0x12, 0x0c, 0x0a,
	0x08, 0x52, 0x45, 0x47, 0x49, 0x53, 0x54, 0x45, 0x52, 0x10, 0x01, 0x12, 0x0e, 0x0a, 0x0a, 0x44,
	0x45, 0x52, 0x45, 0x47, 0x49, 0x53, 0x54, 0x45, 0x52, 0x10, 0x02, 0x22, 0x5a, 0x0a, 0x28, 0x47,
	0x65, 0x74, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4e, 0x75, 0x6d, 0x42, 0x79, 0x54, 0x65,
	0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2e, 0x0a, 0x12, 0x74, 0x65, 0x6d, 0x70, 0x6c,
	0x61, 0x74, 0x65, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x12, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x49, 0x64, 0x65,
	0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x22, 0xd8, 0x01, 0x0a, 0x29, 0x47, 0x65, 0x74, 0x43,
	0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4e, 0x75, 0x6d, 0x42, 0x79, 0x54, 0x65, 0x6d, 0x70, 0x6c,
	0x61, 0x74, 0x65, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x69, 0x0a, 0x03, 0x72, 0x65, 0x73, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x57, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x5f, 0x6f, 0x72, 0x63,
	0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x73, 0x6f, 0x75, 0x74, 0x68, 0x62,
	0x6f, 0x75, 0x6e, 0x64, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x6c,
	0x75, 0x73, 0x74, 0x65, 0x72, 0x4e, 0x75, 0x6d, 0x42, 0x79, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61,
	0x74, 0x65, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x03, 0x72, 0x65, 0x73,
	0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4e, 0x75, 0x6d, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4e, 0x75, 0x6d,
	0x22, 0x20, 0x0a, 0x06, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x0b, 0x0a, 0x07, 0x53, 0x55,
	0x43, 0x43, 0x45, 0x53, 0x53, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x45, 0x52, 0x52, 0x4f, 0x52,
	0x10, 0x01, 0x32, 0xa0, 0x04, 0x0a, 0x1d, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4f, 0x72,
	0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x53, 0x6f, 0x75, 0x74, 0x68, 0x62,
	0x6f, 0x75, 0x6e, 0x64, 0x12, 0x92, 0x01, 0x0a, 0x0f, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65,
	0x72, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x12, 0x3d, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74,
	0x65, 0x72, 0x5f, 0x6f, 0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f,
	0x73, 0x6f, 0x75, 0x74, 0x68, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x3e, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65,
	0x72, 0x5f, 0x6f, 0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x73,
	0x6f, 0x75, 0x74, 0x68, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x9e, 0x01, 0x0a, 0x13, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x12, 0x41, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x5f, 0x6f, 0x72, 0x63, 0x68,
	0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x73, 0x6f, 0x75, 0x74, 0x68, 0x62, 0x6f,
	0x75, 0x6e, 0x64, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x42, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x5f, 0x6f,
	0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x73, 0x6f, 0x75, 0x74,
	0x68, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0xc8, 0x01, 0x0a, 0x21, 0x47,
	0x65, 0x74, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4e, 0x75, 0x6d, 0x42, 0x79, 0x54, 0x65,
	0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72,
	0x12, 0x4f, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x5f, 0x6f, 0x72, 0x63, 0x68, 0x65,
	0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x73, 0x6f, 0x75, 0x74, 0x68, 0x62, 0x6f, 0x75,
	0x6e, 0x64, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x6c, 0x75, 0x73,
	0x74, 0x65, 0x72, 0x4e, 0x75, 0x6d, 0x42, 0x79, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65,
	0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x50, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x5f, 0x6f, 0x72, 0x63, 0x68,
	0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x73, 0x6f, 0x75, 0x74, 0x68, 0x62, 0x6f,
	0x75, 0x6e, 0x64, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x6c, 0x75,
	0x73, 0x74, 0x65, 0x72, 0x4e, 0x75, 0x6d, 0x42, 0x79, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74,
	0x65, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x23, 0x5a, 0x21, 0x2e, 0x3b, 0x63, 0x6c, 0x75, 0x73, 0x74,
	0x65, 0x72, 0x5f, 0x6f, 0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f,
	0x73, 0x6f, 0x75, 0x74, 0x68, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_cluster_orchestrator_southbound_proto_rawDescOnce sync.Once
	file_cluster_orchestrator_southbound_proto_rawDescData = file_cluster_orchestrator_southbound_proto_rawDesc
)

func file_cluster_orchestrator_southbound_proto_rawDescGZIP() []byte {
	file_cluster_orchestrator_southbound_proto_rawDescOnce.Do(func() {
		file_cluster_orchestrator_southbound_proto_rawDescData = protoimpl.X.CompressGZIP(file_cluster_orchestrator_southbound_proto_rawDescData)
	})
	return file_cluster_orchestrator_southbound_proto_rawDescData
}

var file_cluster_orchestrator_southbound_proto_enumTypes = make([]protoimpl.EnumInfo, 4)
var file_cluster_orchestrator_southbound_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_cluster_orchestrator_southbound_proto_goTypes = []any{
	(RegisterClusterResponse_Result)(0),                   // 0: cluster_orchestrator_southbound_proto.RegisterClusterResponse.Result
	(UpdateClusterStatusRequest_Code)(0),                  // 1: cluster_orchestrator_southbound_proto.UpdateClusterStatusRequest.Code
	(UpdateClusterStatusResponse_ActionRequest)(0),        // 2: cluster_orchestrator_southbound_proto.UpdateClusterStatusResponse.ActionRequest
	(GetClusterNumByTemplateIdentifierResponse_Result)(0), // 3: cluster_orchestrator_southbound_proto.GetClusterNumByTemplateIdentifierResponse.Result
	(*RegisterClusterRequest)(nil),                        // 4: cluster_orchestrator_southbound_proto.RegisterClusterRequest
	(*RegisterClusterResponse)(nil),                       // 5: cluster_orchestrator_southbound_proto.RegisterClusterResponse
	(*ShellScriptCommand)(nil),                            // 6: cluster_orchestrator_southbound_proto.ShellScriptCommand
	(*UpdateClusterStatusRequest)(nil),                    // 7: cluster_orchestrator_southbound_proto.UpdateClusterStatusRequest
	(*UpdateClusterStatusResponse)(nil),                   // 8: cluster_orchestrator_southbound_proto.UpdateClusterStatusResponse
	(*GetClusterNumByTemplateIdentifierRequest)(nil),      // 9: cluster_orchestrator_southbound_proto.GetClusterNumByTemplateIdentifierRequest
	(*GetClusterNumByTemplateIdentifierResponse)(nil),     // 10: cluster_orchestrator_southbound_proto.GetClusterNumByTemplateIdentifierResponse
}
var file_cluster_orchestrator_southbound_proto_depIdxs = []int32{
	6,  // 0: cluster_orchestrator_southbound_proto.RegisterClusterResponse.install_cmd:type_name -> cluster_orchestrator_southbound_proto.ShellScriptCommand
	6,  // 1: cluster_orchestrator_southbound_proto.RegisterClusterResponse.uninstall_cmd:type_name -> cluster_orchestrator_southbound_proto.ShellScriptCommand
	0,  // 2: cluster_orchestrator_southbound_proto.RegisterClusterResponse.res:type_name -> cluster_orchestrator_southbound_proto.RegisterClusterResponse.Result
	1,  // 3: cluster_orchestrator_southbound_proto.UpdateClusterStatusRequest.code:type_name -> cluster_orchestrator_southbound_proto.UpdateClusterStatusRequest.Code
	2,  // 4: cluster_orchestrator_southbound_proto.UpdateClusterStatusResponse.action_request:type_name -> cluster_orchestrator_southbound_proto.UpdateClusterStatusResponse.ActionRequest
	3,  // 5: cluster_orchestrator_southbound_proto.GetClusterNumByTemplateIdentifierResponse.res:type_name -> cluster_orchestrator_southbound_proto.GetClusterNumByTemplateIdentifierResponse.Result
	4,  // 6: cluster_orchestrator_southbound_proto.ClusterOrchestratorSouthbound.RegisterCluster:input_type -> cluster_orchestrator_southbound_proto.RegisterClusterRequest
	7,  // 7: cluster_orchestrator_southbound_proto.ClusterOrchestratorSouthbound.UpdateClusterStatus:input_type -> cluster_orchestrator_southbound_proto.UpdateClusterStatusRequest
	9,  // 8: cluster_orchestrator_southbound_proto.ClusterOrchestratorSouthbound.GetClusterNumByTemplateIdentifier:input_type -> cluster_orchestrator_southbound_proto.GetClusterNumByTemplateIdentifierRequest
	5,  // 9: cluster_orchestrator_southbound_proto.ClusterOrchestratorSouthbound.RegisterCluster:output_type -> cluster_orchestrator_southbound_proto.RegisterClusterResponse
	8,  // 10: cluster_orchestrator_southbound_proto.ClusterOrchestratorSouthbound.UpdateClusterStatus:output_type -> cluster_orchestrator_southbound_proto.UpdateClusterStatusResponse
	10, // 11: cluster_orchestrator_southbound_proto.ClusterOrchestratorSouthbound.GetClusterNumByTemplateIdentifier:output_type -> cluster_orchestrator_southbound_proto.GetClusterNumByTemplateIdentifierResponse
	9,  // [9:12] is the sub-list for method output_type
	6,  // [6:9] is the sub-list for method input_type
	6,  // [6:6] is the sub-list for extension type_name
	6,  // [6:6] is the sub-list for extension extendee
	0,  // [0:6] is the sub-list for field type_name
}

func init() { file_cluster_orchestrator_southbound_proto_init() }
func file_cluster_orchestrator_southbound_proto_init() {
	if File_cluster_orchestrator_southbound_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_cluster_orchestrator_southbound_proto_rawDesc,
			NumEnums:      4,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_cluster_orchestrator_southbound_proto_goTypes,
		DependencyIndexes: file_cluster_orchestrator_southbound_proto_depIdxs,
		EnumInfos:         file_cluster_orchestrator_southbound_proto_enumTypes,
		MessageInfos:      file_cluster_orchestrator_southbound_proto_msgTypes,
	}.Build()
	File_cluster_orchestrator_southbound_proto = out.File
	file_cluster_orchestrator_southbound_proto_rawDesc = nil
	file_cluster_orchestrator_southbound_proto_goTypes = nil
	file_cluster_orchestrator_southbound_proto_depIdxs = nil
}
