// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.0
// source: forge.proto

package subgraph

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	SubgraphService_CreateSubgraph_FullMethodName      = "/proto.SubgraphService/CreateSubgraph"
	SubgraphService_DeleteSubgraph_FullMethodName      = "/proto.SubgraphService/DeleteSubgraph"
	SubgraphService_CreateSubgraphBatch_FullMethodName = "/proto.SubgraphService/CreateSubgraphBatch"
)

// SubgraphServiceClient is the client API for SubgraphService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SubgraphServiceClient interface {
	CreateSubgraph(ctx context.Context, in *CreateSubgraphRequest, opts ...grpc.CallOption) (*CreateSubgraphResponse, error)
	DeleteSubgraph(ctx context.Context, in *DeleteSubgraphRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	CreateSubgraphBatch(ctx context.Context, in *CreateSubgraphBatchRequest, opts ...grpc.CallOption) (*CreateSubgraphBatchResponse, error)
}

type subgraphServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSubgraphServiceClient(cc grpc.ClientConnInterface) SubgraphServiceClient {
	return &subgraphServiceClient{cc}
}

func (c *subgraphServiceClient) CreateSubgraph(ctx context.Context, in *CreateSubgraphRequest, opts ...grpc.CallOption) (*CreateSubgraphResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateSubgraphResponse)
	err := c.cc.Invoke(ctx, SubgraphService_CreateSubgraph_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *subgraphServiceClient) DeleteSubgraph(ctx context.Context, in *DeleteSubgraphRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, SubgraphService_DeleteSubgraph_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *subgraphServiceClient) CreateSubgraphBatch(ctx context.Context, in *CreateSubgraphBatchRequest, opts ...grpc.CallOption) (*CreateSubgraphBatchResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateSubgraphBatchResponse)
	err := c.cc.Invoke(ctx, SubgraphService_CreateSubgraphBatch_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SubgraphServiceServer is the server API for SubgraphService service.
// All implementations must embed UnimplementedSubgraphServiceServer
// for forward compatibility.
type SubgraphServiceServer interface {
	CreateSubgraph(context.Context, *CreateSubgraphRequest) (*CreateSubgraphResponse, error)
	DeleteSubgraph(context.Context, *DeleteSubgraphRequest) (*emptypb.Empty, error)
	CreateSubgraphBatch(context.Context, *CreateSubgraphBatchRequest) (*CreateSubgraphBatchResponse, error)
	mustEmbedUnimplementedSubgraphServiceServer()
}

// UnimplementedSubgraphServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedSubgraphServiceServer struct{}

func (UnimplementedSubgraphServiceServer) CreateSubgraph(context.Context, *CreateSubgraphRequest) (*CreateSubgraphResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSubgraph not implemented")
}
func (UnimplementedSubgraphServiceServer) DeleteSubgraph(context.Context, *DeleteSubgraphRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteSubgraph not implemented")
}
func (UnimplementedSubgraphServiceServer) CreateSubgraphBatch(context.Context, *CreateSubgraphBatchRequest) (*CreateSubgraphBatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSubgraphBatch not implemented")
}
func (UnimplementedSubgraphServiceServer) mustEmbedUnimplementedSubgraphServiceServer() {}
func (UnimplementedSubgraphServiceServer) testEmbeddedByValue()                         {}

// UnsafeSubgraphServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SubgraphServiceServer will
// result in compilation errors.
type UnsafeSubgraphServiceServer interface {
	mustEmbedUnimplementedSubgraphServiceServer()
}

func RegisterSubgraphServiceServer(s grpc.ServiceRegistrar, srv SubgraphServiceServer) {
	// If the following call pancis, it indicates UnimplementedSubgraphServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&SubgraphService_ServiceDesc, srv)
}

func _SubgraphService_CreateSubgraph_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateSubgraphRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SubgraphServiceServer).CreateSubgraph(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SubgraphService_CreateSubgraph_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SubgraphServiceServer).CreateSubgraph(ctx, req.(*CreateSubgraphRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SubgraphService_DeleteSubgraph_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteSubgraphRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SubgraphServiceServer).DeleteSubgraph(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SubgraphService_DeleteSubgraph_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SubgraphServiceServer).DeleteSubgraph(ctx, req.(*DeleteSubgraphRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SubgraphService_CreateSubgraphBatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateSubgraphBatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SubgraphServiceServer).CreateSubgraphBatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SubgraphService_CreateSubgraphBatch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SubgraphServiceServer).CreateSubgraphBatch(ctx, req.(*CreateSubgraphBatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// SubgraphService_ServiceDesc is the grpc.ServiceDesc for SubgraphService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SubgraphService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.SubgraphService",
	HandlerType: (*SubgraphServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateSubgraph",
			Handler:    _SubgraphService_CreateSubgraph_Handler,
		},
		{
			MethodName: "DeleteSubgraph",
			Handler:    _SubgraphService_DeleteSubgraph_Handler,
		},
		{
			MethodName: "CreateSubgraphBatch",
			Handler:    _SubgraphService_CreateSubgraphBatch_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "forge.proto",
}
