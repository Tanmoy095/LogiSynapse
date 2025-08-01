// proto/shipment.proto

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: proto/shipment.proto

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
	ShipmentService_GetShipments_FullMethodName   = "/shipment.ShipmentService/GetShipments"
	ShipmentService_CreateShipment_FullMethodName = "/shipment.ShipmentService/CreateShipment"
)

// ShipmentServiceClient is the client API for ShipmentService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ShipmentServiceClient interface {
	GetShipments(ctx context.Context, in *GetShipmentsRequest, opts ...grpc.CallOption) (*GetShipmentsResponse, error)
	CreateShipment(ctx context.Context, in *CreateShipmentRequest, opts ...grpc.CallOption) (*CreateShipmentResponse, error)
}

type shipmentServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewShipmentServiceClient(cc grpc.ClientConnInterface) ShipmentServiceClient {
	return &shipmentServiceClient{cc}
}

func (c *shipmentServiceClient) GetShipments(ctx context.Context, in *GetShipmentsRequest, opts ...grpc.CallOption) (*GetShipmentsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetShipmentsResponse)
	err := c.cc.Invoke(ctx, ShipmentService_GetShipments_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shipmentServiceClient) CreateShipment(ctx context.Context, in *CreateShipmentRequest, opts ...grpc.CallOption) (*CreateShipmentResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateShipmentResponse)
	err := c.cc.Invoke(ctx, ShipmentService_CreateShipment_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ShipmentServiceServer is the server API for ShipmentService service.
// All implementations must embed UnimplementedShipmentServiceServer
// for forward compatibility.
type ShipmentServiceServer interface {
	GetShipments(context.Context, *GetShipmentsRequest) (*GetShipmentsResponse, error)
	CreateShipment(context.Context, *CreateShipmentRequest) (*CreateShipmentResponse, error)
	mustEmbedUnimplementedShipmentServiceServer()
}

// UnimplementedShipmentServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedShipmentServiceServer struct{}

func (UnimplementedShipmentServiceServer) GetShipments(context.Context, *GetShipmentsRequest) (*GetShipmentsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetShipments not implemented")
}
func (UnimplementedShipmentServiceServer) CreateShipment(context.Context, *CreateShipmentRequest) (*CreateShipmentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateShipment not implemented")
}
func (UnimplementedShipmentServiceServer) mustEmbedUnimplementedShipmentServiceServer() {}
func (UnimplementedShipmentServiceServer) testEmbeddedByValue()                         {}

// UnsafeShipmentServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ShipmentServiceServer will
// result in compilation errors.
type UnsafeShipmentServiceServer interface {
	mustEmbedUnimplementedShipmentServiceServer()
}

func RegisterShipmentServiceServer(s grpc.ServiceRegistrar, srv ShipmentServiceServer) {
	// If the following call pancis, it indicates UnimplementedShipmentServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ShipmentService_ServiceDesc, srv)
}

func _ShipmentService_GetShipments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetShipmentsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShipmentServiceServer).GetShipments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShipmentService_GetShipments_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShipmentServiceServer).GetShipments(ctx, req.(*GetShipmentsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShipmentService_CreateShipment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateShipmentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShipmentServiceServer).CreateShipment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShipmentService_CreateShipment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShipmentServiceServer).CreateShipment(ctx, req.(*CreateShipmentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ShipmentService_ServiceDesc is the grpc.ServiceDesc for ShipmentService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ShipmentService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "shipment.ShipmentService",
	HandlerType: (*ShipmentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetShipments",
			Handler:    _ShipmentService_GetShipments_Handler,
		},
		{
			MethodName: "CreateShipment",
			Handler:    _ShipmentService_CreateShipment_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/shipment.proto",
}
