package gate

import (
	"context"

	gate "github.com/blabtm/emqx-gate/api"

	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type adapterMock struct {
	mock.Mock
}

func (a *adapterMock) Send(ctx context.Context, in *gate.SendBytesRequest, opts ...grpc.CallOption) (*gate.CodeResponse, error) {
	args := a.Called(ctx, in, opts)
	return args.Get(0).(*gate.CodeResponse), args.Error(1)
}

func (a *adapterMock) Close(ctx context.Context, in *gate.CloseSocketRequest, opts ...grpc.CallOption) (*gate.CodeResponse, error) {
	args := a.Called(ctx, in, opts)
	return args.Get(0).(*gate.CodeResponse), args.Error(1)
}

func (a *adapterMock) Authenticate(ctx context.Context, in *gate.AuthenticateRequest, opts ...grpc.CallOption) (*gate.CodeResponse, error) {
	args := a.Called(ctx, in, opts)
	return args.Get(0).(*gate.CodeResponse), args.Error(1)
}

func (a *adapterMock) StartTimer(ctx context.Context, in *gate.TimerRequest, opts ...grpc.CallOption) (*gate.CodeResponse, error) {
	args := a.Called(ctx, in, opts)
	return args.Get(0).(*gate.CodeResponse), args.Error(1)
}

func (a *adapterMock) Publish(ctx context.Context, in *gate.PublishRequest, opts ...grpc.CallOption) (*gate.CodeResponse, error) {
	args := a.Called(ctx, in, opts)
	return args.Get(0).(*gate.CodeResponse), args.Error(1)
}

func (a *adapterMock) Subscribe(ctx context.Context, in *gate.SubscribeRequest, opts ...grpc.CallOption) (*gate.CodeResponse, error) {
	args := a.Called(ctx, in, opts)
	return args.Get(0).(*gate.CodeResponse), args.Error(1)
}

func (a *adapterMock) Unsubscribe(ctx context.Context, in *gate.UnsubscribeRequest, opts ...grpc.CallOption) (*gate.CodeResponse, error) {
	args := a.Called(ctx, in, opts)
	return args.Get(0).(*gate.CodeResponse), args.Error(1)
}

func (a *adapterMock) RawPublish(ctx context.Context, in *gate.RawPublishRequest, opts ...grpc.CallOption) (*gate.CodeResponse, error) {
	args := a.Called(ctx, in, opts)
	return args.Get(0).(*gate.CodeResponse), args.Error(1)
}
