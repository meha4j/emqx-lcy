package proc

import (
	context "context"
	"log"
)

type service struct {
	UnimplementedConnectionUnaryHandlerServer
}

func NewService() ConnectionUnaryHandlerServer {
	return &service{}
}

func (*service) OnSocketCreated(_ context.Context, req *SocketCreatedRequest) (*EmptySuccess, error) {
	log.Println(req)

	return nil, nil
}

func (*service) OnSocketClosed(_ context.Context, req *SocketClosedRequest) (*EmptySuccess, error) {
	log.Println(req)

	return nil, nil
}

func (*service) OnReceivedBytes(_ context.Context, req *ReceivedBytesRequest) (*EmptySuccess, error) {
	log.Println(req)

	return nil, nil
}

func (*service) OnTimerTimeout(_ context.Context, req *TimerTimeoutRequest) (*EmptySuccess, error) {
	log.Println(req)

	return nil, nil
}

func (*service) OnReceivedMessages(_ context.Context, req *ReceivedMessagesRequest) (*EmptySuccess, error) {
	log.Println(req)

	return nil, nil
}
