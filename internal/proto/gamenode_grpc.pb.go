// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.4
// source: internal/proto/gamenode.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	GameNodeGRPCService_Register_FullMethodName             = "/gamenode.GameNodeGRPCService/Register"
	GameNodeGRPCService_Heartbeat_FullMethodName            = "/gamenode.GameNodeGRPCService/Heartbeat"
	GameNodeGRPCService_ReportMetrics_FullMethodName        = "/gamenode.GameNodeGRPCService/ReportMetrics"
	GameNodeGRPCService_UpdateResourceInfo_FullMethodName   = "/gamenode.GameNodeGRPCService/UpdateResourceInfo"
	GameNodeGRPCService_ExecutePipeline_FullMethodName      = "/gamenode.GameNodeGRPCService/ExecutePipeline"
	GameNodeGRPCService_UpdatePipelineStatus_FullMethodName = "/gamenode.GameNodeGRPCService/UpdatePipelineStatus"
	GameNodeGRPCService_UpdateStepStatus_FullMethodName     = "/gamenode.GameNodeGRPCService/UpdateStepStatus"
	GameNodeGRPCService_CancelPipeline_FullMethodName       = "/gamenode.GameNodeGRPCService/CancelPipeline"
	GameNodeGRPCService_StreamLogs_FullMethodName           = "/gamenode.GameNodeGRPCService/StreamLogs"
)

// GameNodeGRPCServiceClient is the client API for GameNodeGRPCService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// GameNodeGRPCService 定义节点Agent的gRPC服务
type GameNodeGRPCServiceClient interface {
	// 节点管理
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
	Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error)
	ReportMetrics(ctx context.Context, in *MetricsReport, opts ...grpc.CallOption) (*ReportResponse, error)
	UpdateResourceInfo(ctx context.Context, in *ResourceInfo, opts ...grpc.CallOption) (*UpdateResponse, error)
	// Pipeline管理
	ExecutePipeline(ctx context.Context, in *ExecutePipelineRequest, opts ...grpc.CallOption) (*ExecutePipelineResponse, error)
	UpdatePipelineStatus(ctx context.Context, in *PipelineStatusUpdate, opts ...grpc.CallOption) (*UpdateResponse, error)
	UpdateStepStatus(ctx context.Context, in *StepStatusUpdate, opts ...grpc.CallOption) (*UpdateResponse, error)
	CancelPipeline(ctx context.Context, in *PipelineCancelRequest, opts ...grpc.CallOption) (*CancelResponse, error)
	// 日志流
	StreamLogs(ctx context.Context, in *LogRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[LogEntry], error)
}

type gameNodeGRPCServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGameNodeGRPCServiceClient(cc grpc.ClientConnInterface) GameNodeGRPCServiceClient {
	return &gameNodeGRPCServiceClient{cc}
}

func (c *gameNodeGRPCServiceClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RegisterResponse)
	err := c.cc.Invoke(ctx, GameNodeGRPCService_Register_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gameNodeGRPCServiceClient) Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(HeartbeatResponse)
	err := c.cc.Invoke(ctx, GameNodeGRPCService_Heartbeat_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gameNodeGRPCServiceClient) ReportMetrics(ctx context.Context, in *MetricsReport, opts ...grpc.CallOption) (*ReportResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ReportResponse)
	err := c.cc.Invoke(ctx, GameNodeGRPCService_ReportMetrics_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gameNodeGRPCServiceClient) UpdateResourceInfo(ctx context.Context, in *ResourceInfo, opts ...grpc.CallOption) (*UpdateResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateResponse)
	err := c.cc.Invoke(ctx, GameNodeGRPCService_UpdateResourceInfo_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gameNodeGRPCServiceClient) ExecutePipeline(ctx context.Context, in *ExecutePipelineRequest, opts ...grpc.CallOption) (*ExecutePipelineResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ExecutePipelineResponse)
	err := c.cc.Invoke(ctx, GameNodeGRPCService_ExecutePipeline_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gameNodeGRPCServiceClient) UpdatePipelineStatus(ctx context.Context, in *PipelineStatusUpdate, opts ...grpc.CallOption) (*UpdateResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateResponse)
	err := c.cc.Invoke(ctx, GameNodeGRPCService_UpdatePipelineStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gameNodeGRPCServiceClient) UpdateStepStatus(ctx context.Context, in *StepStatusUpdate, opts ...grpc.CallOption) (*UpdateResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateResponse)
	err := c.cc.Invoke(ctx, GameNodeGRPCService_UpdateStepStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gameNodeGRPCServiceClient) CancelPipeline(ctx context.Context, in *PipelineCancelRequest, opts ...grpc.CallOption) (*CancelResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CancelResponse)
	err := c.cc.Invoke(ctx, GameNodeGRPCService_CancelPipeline_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gameNodeGRPCServiceClient) StreamLogs(ctx context.Context, in *LogRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[LogEntry], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &GameNodeGRPCService_ServiceDesc.Streams[0], GameNodeGRPCService_StreamLogs_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[LogRequest, LogEntry]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type GameNodeGRPCService_StreamLogsClient = grpc.ServerStreamingClient[LogEntry]

// GameNodeGRPCServiceServer is the server API for GameNodeGRPCService service.
// All implementations must embed UnimplementedGameNodeGRPCServiceServer
// for forward compatibility.
//
// GameNodeGRPCService 定义节点Agent的gRPC服务
type GameNodeGRPCServiceServer interface {
	// 节点管理
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
	Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error)
	ReportMetrics(context.Context, *MetricsReport) (*ReportResponse, error)
	UpdateResourceInfo(context.Context, *ResourceInfo) (*UpdateResponse, error)
	// Pipeline管理
	ExecutePipeline(context.Context, *ExecutePipelineRequest) (*ExecutePipelineResponse, error)
	UpdatePipelineStatus(context.Context, *PipelineStatusUpdate) (*UpdateResponse, error)
	UpdateStepStatus(context.Context, *StepStatusUpdate) (*UpdateResponse, error)
	CancelPipeline(context.Context, *PipelineCancelRequest) (*CancelResponse, error)
	// 日志流
	StreamLogs(*LogRequest, grpc.ServerStreamingServer[LogEntry]) error
	mustEmbedUnimplementedGameNodeGRPCServiceServer()
}

// UnimplementedGameNodeGRPCServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedGameNodeGRPCServiceServer struct{}

func (UnimplementedGameNodeGRPCServiceServer) Register(context.Context, *RegisterRequest) (*RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedGameNodeGRPCServiceServer) Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Heartbeat not implemented")
}
func (UnimplementedGameNodeGRPCServiceServer) ReportMetrics(context.Context, *MetricsReport) (*ReportResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReportMetrics not implemented")
}
func (UnimplementedGameNodeGRPCServiceServer) UpdateResourceInfo(context.Context, *ResourceInfo) (*UpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateResourceInfo not implemented")
}
func (UnimplementedGameNodeGRPCServiceServer) ExecutePipeline(context.Context, *ExecutePipelineRequest) (*ExecutePipelineResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecutePipeline not implemented")
}
func (UnimplementedGameNodeGRPCServiceServer) UpdatePipelineStatus(context.Context, *PipelineStatusUpdate) (*UpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdatePipelineStatus not implemented")
}
func (UnimplementedGameNodeGRPCServiceServer) UpdateStepStatus(context.Context, *StepStatusUpdate) (*UpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateStepStatus not implemented")
}
func (UnimplementedGameNodeGRPCServiceServer) CancelPipeline(context.Context, *PipelineCancelRequest) (*CancelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelPipeline not implemented")
}
func (UnimplementedGameNodeGRPCServiceServer) StreamLogs(*LogRequest, grpc.ServerStreamingServer[LogEntry]) error {
	return status.Errorf(codes.Unimplemented, "method StreamLogs not implemented")
}
func (UnimplementedGameNodeGRPCServiceServer) mustEmbedUnimplementedGameNodeGRPCServiceServer() {}
func (UnimplementedGameNodeGRPCServiceServer) testEmbeddedByValue()                             {}

// UnsafeGameNodeGRPCServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GameNodeGRPCServiceServer will
// result in compilation errors.
type UnsafeGameNodeGRPCServiceServer interface {
	mustEmbedUnimplementedGameNodeGRPCServiceServer()
}

func RegisterGameNodeGRPCServiceServer(s grpc.ServiceRegistrar, srv GameNodeGRPCServiceServer) {
	// If the following call pancis, it indicates UnimplementedGameNodeGRPCServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&GameNodeGRPCService_ServiceDesc, srv)
}

func _GameNodeGRPCService_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GameNodeGRPCServiceServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GameNodeGRPCService_Register_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GameNodeGRPCServiceServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GameNodeGRPCService_Heartbeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeartbeatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GameNodeGRPCServiceServer).Heartbeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GameNodeGRPCService_Heartbeat_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GameNodeGRPCServiceServer).Heartbeat(ctx, req.(*HeartbeatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GameNodeGRPCService_ReportMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetricsReport)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GameNodeGRPCServiceServer).ReportMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GameNodeGRPCService_ReportMetrics_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GameNodeGRPCServiceServer).ReportMetrics(ctx, req.(*MetricsReport))
	}
	return interceptor(ctx, in, info, handler)
}

func _GameNodeGRPCService_UpdateResourceInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResourceInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GameNodeGRPCServiceServer).UpdateResourceInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GameNodeGRPCService_UpdateResourceInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GameNodeGRPCServiceServer).UpdateResourceInfo(ctx, req.(*ResourceInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _GameNodeGRPCService_ExecutePipeline_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExecutePipelineRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GameNodeGRPCServiceServer).ExecutePipeline(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GameNodeGRPCService_ExecutePipeline_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GameNodeGRPCServiceServer).ExecutePipeline(ctx, req.(*ExecutePipelineRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GameNodeGRPCService_UpdatePipelineStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PipelineStatusUpdate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GameNodeGRPCServiceServer).UpdatePipelineStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GameNodeGRPCService_UpdatePipelineStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GameNodeGRPCServiceServer).UpdatePipelineStatus(ctx, req.(*PipelineStatusUpdate))
	}
	return interceptor(ctx, in, info, handler)
}

func _GameNodeGRPCService_UpdateStepStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StepStatusUpdate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GameNodeGRPCServiceServer).UpdateStepStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GameNodeGRPCService_UpdateStepStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GameNodeGRPCServiceServer).UpdateStepStatus(ctx, req.(*StepStatusUpdate))
	}
	return interceptor(ctx, in, info, handler)
}

func _GameNodeGRPCService_CancelPipeline_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PipelineCancelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GameNodeGRPCServiceServer).CancelPipeline(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GameNodeGRPCService_CancelPipeline_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GameNodeGRPCServiceServer).CancelPipeline(ctx, req.(*PipelineCancelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GameNodeGRPCService_StreamLogs_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(LogRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GameNodeGRPCServiceServer).StreamLogs(m, &grpc.GenericServerStream[LogRequest, LogEntry]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type GameNodeGRPCService_StreamLogsServer = grpc.ServerStreamingServer[LogEntry]

// GameNodeGRPCService_ServiceDesc is the grpc.ServiceDesc for GameNodeGRPCService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GameNodeGRPCService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gamenode.GameNodeGRPCService",
	HandlerType: (*GameNodeGRPCServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _GameNodeGRPCService_Register_Handler,
		},
		{
			MethodName: "Heartbeat",
			Handler:    _GameNodeGRPCService_Heartbeat_Handler,
		},
		{
			MethodName: "ReportMetrics",
			Handler:    _GameNodeGRPCService_ReportMetrics_Handler,
		},
		{
			MethodName: "UpdateResourceInfo",
			Handler:    _GameNodeGRPCService_UpdateResourceInfo_Handler,
		},
		{
			MethodName: "ExecutePipeline",
			Handler:    _GameNodeGRPCService_ExecutePipeline_Handler,
		},
		{
			MethodName: "UpdatePipelineStatus",
			Handler:    _GameNodeGRPCService_UpdatePipelineStatus_Handler,
		},
		{
			MethodName: "UpdateStepStatus",
			Handler:    _GameNodeGRPCService_UpdateStepStatus_Handler,
		},
		{
			MethodName: "CancelPipeline",
			Handler:    _GameNodeGRPCService_CancelPipeline_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamLogs",
			Handler:       _GameNodeGRPCService_StreamLogs_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "internal/proto/gamenode.proto",
}
