// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: connector.proto

package connector

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// NodeAPIClient is the client API for NodeAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NodeAPIClient interface {
	Get(ctx context.Context, in *Request, opts ...grpc.CallOption) (*NodeResponse, error)
	Set(ctx context.Context, in *Request, opts ...grpc.CallOption) (*NodeResponse, error)
	Del(ctx context.Context, in *Request, opts ...grpc.CallOption) (*NodeResponse, error)
}

type nodeAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewNodeAPIClient(cc grpc.ClientConnInterface) NodeAPIClient {
	return &nodeAPIClient{cc}
}

func (c *nodeAPIClient) Get(ctx context.Context, in *Request, opts ...grpc.CallOption) (*NodeResponse, error) {
	out := new(NodeResponse)
	err := c.cc.Invoke(ctx, "/NodeAPI/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeAPIClient) Set(ctx context.Context, in *Request, opts ...grpc.CallOption) (*NodeResponse, error) {
	out := new(NodeResponse)
	err := c.cc.Invoke(ctx, "/NodeAPI/Set", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeAPIClient) Del(ctx context.Context, in *Request, opts ...grpc.CallOption) (*NodeResponse, error) {
	out := new(NodeResponse)
	err := c.cc.Invoke(ctx, "/NodeAPI/Del", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NodeAPIServer is the server API for NodeAPI service.
// All implementations must embed UnimplementedNodeAPIServer
// for forward compatibility
type NodeAPIServer interface {
	Get(context.Context, *Request) (*NodeResponse, error)
	Set(context.Context, *Request) (*NodeResponse, error)
	Del(context.Context, *Request) (*NodeResponse, error)
	mustEmbedUnimplementedNodeAPIServer()
}

// UnimplementedNodeAPIServer must be embedded to have forward compatible implementations.
type UnimplementedNodeAPIServer struct {
}

func (UnimplementedNodeAPIServer) Get(context.Context, *Request) (*NodeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedNodeAPIServer) Set(context.Context, *Request) (*NodeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Set not implemented")
}
func (UnimplementedNodeAPIServer) Del(context.Context, *Request) (*NodeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Del not implemented")
}
func (UnimplementedNodeAPIServer) mustEmbedUnimplementedNodeAPIServer() {}

// UnsafeNodeAPIServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NodeAPIServer will
// result in compilation errors.
type UnsafeNodeAPIServer interface {
	mustEmbedUnimplementedNodeAPIServer()
}

func RegisterNodeAPIServer(s grpc.ServiceRegistrar, srv NodeAPIServer) {
	s.RegisterService(&NodeAPI_ServiceDesc, srv)
}

func _NodeAPI_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeAPIServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/NodeAPI/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeAPIServer).Get(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeAPI_Set_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeAPIServer).Set(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/NodeAPI/Set",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeAPIServer).Set(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeAPI_Del_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeAPIServer).Del(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/NodeAPI/Del",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeAPIServer).Del(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

// NodeAPI_ServiceDesc is the grpc.ServiceDesc for NodeAPI service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var NodeAPI_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "NodeAPI",
	HandlerType: (*NodeAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _NodeAPI_Get_Handler,
		},
		{
			MethodName: "Set",
			Handler:    _NodeAPI_Set_Handler,
		},
		{
			MethodName: "Del",
			Handler:    _NodeAPI_Del_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "connector.proto",
}
