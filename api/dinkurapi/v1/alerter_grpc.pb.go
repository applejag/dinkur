// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package v1

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

// AlerterClient is the client API for Alerter service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AlerterClient interface {
	StreamAlert(ctx context.Context, in *StreamAlertRequest, opts ...grpc.CallOption) (Alerter_StreamAlertClient, error)
	GetAlertList(ctx context.Context, in *GetAlertListRequest, opts ...grpc.CallOption) (*GetAlertListResponse, error)
	DeleteAlert(ctx context.Context, in *DeleteAlertRequest, opts ...grpc.CallOption) (*DeleteAlertResponse, error)
}

type alerterClient struct {
	cc grpc.ClientConnInterface
}

func NewAlerterClient(cc grpc.ClientConnInterface) AlerterClient {
	return &alerterClient{cc}
}

func (c *alerterClient) StreamAlert(ctx context.Context, in *StreamAlertRequest, opts ...grpc.CallOption) (Alerter_StreamAlertClient, error) {
	stream, err := c.cc.NewStream(ctx, &Alerter_ServiceDesc.Streams[0], "/dinkurapi.v1.Alerter/StreamAlert", opts...)
	if err != nil {
		return nil, err
	}
	x := &alerterStreamAlertClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Alerter_StreamAlertClient interface {
	Recv() (*StreamAlertResponse, error)
	grpc.ClientStream
}

type alerterStreamAlertClient struct {
	grpc.ClientStream
}

func (x *alerterStreamAlertClient) Recv() (*StreamAlertResponse, error) {
	m := new(StreamAlertResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *alerterClient) GetAlertList(ctx context.Context, in *GetAlertListRequest, opts ...grpc.CallOption) (*GetAlertListResponse, error) {
	out := new(GetAlertListResponse)
	err := c.cc.Invoke(ctx, "/dinkurapi.v1.Alerter/GetAlertList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *alerterClient) DeleteAlert(ctx context.Context, in *DeleteAlertRequest, opts ...grpc.CallOption) (*DeleteAlertResponse, error) {
	out := new(DeleteAlertResponse)
	err := c.cc.Invoke(ctx, "/dinkurapi.v1.Alerter/DeleteAlert", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AlerterServer is the server API for Alerter service.
// All implementations must embed UnimplementedAlerterServer
// for forward compatibility
type AlerterServer interface {
	StreamAlert(*StreamAlertRequest, Alerter_StreamAlertServer) error
	GetAlertList(context.Context, *GetAlertListRequest) (*GetAlertListResponse, error)
	DeleteAlert(context.Context, *DeleteAlertRequest) (*DeleteAlertResponse, error)
	mustEmbedUnimplementedAlerterServer()
}

// UnimplementedAlerterServer must be embedded to have forward compatible implementations.
type UnimplementedAlerterServer struct {
}

func (UnimplementedAlerterServer) StreamAlert(*StreamAlertRequest, Alerter_StreamAlertServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamAlert not implemented")
}
func (UnimplementedAlerterServer) GetAlertList(context.Context, *GetAlertListRequest) (*GetAlertListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAlertList not implemented")
}
func (UnimplementedAlerterServer) DeleteAlert(context.Context, *DeleteAlertRequest) (*DeleteAlertResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteAlert not implemented")
}
func (UnimplementedAlerterServer) mustEmbedUnimplementedAlerterServer() {}

// UnsafeAlerterServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AlerterServer will
// result in compilation errors.
type UnsafeAlerterServer interface {
	mustEmbedUnimplementedAlerterServer()
}

func RegisterAlerterServer(s grpc.ServiceRegistrar, srv AlerterServer) {
	s.RegisterService(&Alerter_ServiceDesc, srv)
}

func _Alerter_StreamAlert_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamAlertRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(AlerterServer).StreamAlert(m, &alerterStreamAlertServer{stream})
}

type Alerter_StreamAlertServer interface {
	Send(*StreamAlertResponse) error
	grpc.ServerStream
}

type alerterStreamAlertServer struct {
	grpc.ServerStream
}

func (x *alerterStreamAlertServer) Send(m *StreamAlertResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Alerter_GetAlertList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAlertListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AlerterServer).GetAlertList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/dinkurapi.v1.Alerter/GetAlertList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AlerterServer).GetAlertList(ctx, req.(*GetAlertListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Alerter_DeleteAlert_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteAlertRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AlerterServer).DeleteAlert(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/dinkurapi.v1.Alerter/DeleteAlert",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AlerterServer).DeleteAlert(ctx, req.(*DeleteAlertRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Alerter_ServiceDesc is the grpc.ServiceDesc for Alerter service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Alerter_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "dinkurapi.v1.Alerter",
	HandlerType: (*AlerterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAlertList",
			Handler:    _Alerter_GetAlertList_Handler,
		},
		{
			MethodName: "DeleteAlert",
			Handler:    _Alerter_DeleteAlert_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamAlert",
			Handler:       _Alerter_StreamAlert_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "api/dinkurapi/v1/alerter.proto",
}