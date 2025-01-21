package gate

import (
	"context"
	"errors"
	"testing"
	"time"

	gate "github.com/blabtm/emqx-gate/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func now() time.Time {
	return time.UnixMilli(1118509199999)
}

func TestOnReceivedBytes(t *testing.T) {
	cases := map[string]struct {
		before func(*client)
		req    []byte
		pub    *gate.PublishRequest
		sub    *gate.SubscribeRequest
		usub   *gate.UnsubscribeRequest
		send   *gate.SendBytesRequest
		err    error
	}{
		`publish with time`: {
			req: []byte("time:11.06.2005 23_59_59.999|name:test|method:set|val:11.06\n"),
			pub: &gate.PublishRequest{
				Conn:    "test",
				Topic:   "test",
				Qos:     0,
				Payload: []byte(`{"stamp":1118509199999,"value":"11.06"}`),
			},
		},
		`publish without time`: {
			req: []byte("name:test|method:set|val:11.06\n"),
			pub: &gate.PublishRequest{
				Conn:    "test",
				Topic:   "test",
				Qos:     0,
				Payload: []byte(`{"stamp":1118509199999,"value":"11.06"}`),
			},
		},
		`subscribe`: {
			req: []byte("name:test|method:subscr\n"),
			sub: &gate.SubscribeRequest{
				Conn:  "test",
				Topic: "test",
				Qos:   2,
			},
		},
		`unsubscribe`: {
			req: []byte("name:test|method:release\n"),
			usub: &gate.UnsubscribeRequest{
				Conn:  "test",
				Topic: "test",
			},
		},
		`get with message`: {
			req: []byte("name:test|method:get\n"),
			sub: &gate.SubscribeRequest{
				Conn:  "test",
				Topic: "test",
				Qos:   2,
			},
			usub: &gate.UnsubscribeRequest{
				Conn:  "test",
				Topic: "test",
			},
			send: &gate.SendBytesRequest{
				Conn:  "test",
				Bytes: []byte("time:11.06.2005 23_59_59.999|method:set|name:test|val:11.06|descr:none|type:rw|units:none\n"),
			},
			before: func(cli *client) {
				cli.OnReceivedMessage(context.Background(), &gate.Message{
					Topic:   "test",
					Qos:     0,
					Payload: []byte(`{"stamp":1118509199999,"value":"11.06"}`),
				})
			},
		},
		`get without message`: {
			req: []byte("name:test|method:get\n"),
			sub: &gate.SubscribeRequest{
				Conn:  "test",
				Topic: "test",
				Qos:   2,
			},
			usub: &gate.UnsubscribeRequest{
				Conn:  "test",
				Topic: "test",
			},
			send: &gate.SendBytesRequest{
				Conn:  "test",
				Bytes: []byte("time:11.06.2005 23_59_59.999|method:set|name:test|val:none|descr:none|type:rw|units:none\n"),
			},
			before: func(cli *client) {
				time.Sleep(6 * time.Second)
			},
		},
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			apr := &adapterMock{}

			apr.On("Publish", mock.Anything, c.pub, mock.Anything).
				Return(&gate.CodeResponse{Code: gate.ResultCode_SUCCESS}, nil)
			apr.On("Subscribe", mock.Anything, c.sub, mock.Anything).
				Return(&gate.CodeResponse{Code: gate.ResultCode_SUCCESS}, nil)
			apr.On("Unsubscribe", mock.Anything, c.usub, mock.Anything).
				Return(&gate.CodeResponse{Code: gate.ResultCode_SUCCESS}, nil)
			apr.On("Send", mock.Anything, c.send, mock.Anything).
				Return(&gate.CodeResponse{Code: gate.ResultCode_SUCCESS}, nil)

			cli := newClient("test", apr)
			cli.now = now

			err := cli.OnReceivedBytes(context.Background(), c.req)

			if c.err != nil {
				assert.True(t, errors.Is(err, c.err))
			}

			if c.before != nil {
				c.before(cli)
			}

			if c.pub != nil {
				apr.AssertCalled(t, "Publish", mock.Anything, c.pub, mock.Anything)
			} else {
				apr.AssertNotCalled(t, "Publish", mock.Anything, c.pub, mock.Anything)
			}

			if c.sub != nil {
				apr.AssertCalled(t, "Subscribe", mock.Anything, c.sub, mock.Anything)
			} else {
				apr.AssertNotCalled(t, "Subscribe", mock.Anything, c.sub, mock.Anything)
			}

			if c.usub != nil {
				apr.AssertCalled(t, "Unsubscribe", mock.Anything, c.usub, mock.Anything)
			} else {
				apr.AssertNotCalled(t, "Unsubscribe", mock.Anything, c.usub, mock.Anything)
			}

			if c.send != nil {
				apr.AssertCalled(t, "Send", mock.Anything, c.send, mock.Anything)
			} else {
				apr.AssertNotCalled(t, "Send", mock.Anything, c.send, mock.Anything)
			}
		})
	}
}

func TestOnReceivedMessage(t *testing.T) {
	cases := map[string]struct {
		before func(*client)
		req    *gate.Message
		pub    *gate.PublishRequest
		sub    *gate.SubscribeRequest
		usub   *gate.UnsubscribeRequest
		send   *gate.SendBytesRequest
		err    error
	}{
		`publish`: {
			req: &gate.Message{
				Topic:   "test",
				Qos:     0,
				Payload: []byte(`{"stamp":1118509199999,"value":"11.06"}`),
			},
			send: &gate.SendBytesRequest{
				Conn:  "test",
				Bytes: []byte("time:11.06.2005 23_59_59.999|method:set|name:test|val:11.06|descr:none|type:rw|units:none\n"),
			},
		},
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			apr := &adapterMock{}

			apr.On("Publish", mock.Anything, c.pub, mock.Anything).
				Return(&gate.CodeResponse{Code: gate.ResultCode_SUCCESS}, nil)
			apr.On("Subscribe", mock.Anything, c.sub, mock.Anything).
				Return(&gate.CodeResponse{Code: gate.ResultCode_SUCCESS}, nil)
			apr.On("Unsubscribe", mock.Anything, c.usub, mock.Anything).
				Return(&gate.CodeResponse{Code: gate.ResultCode_SUCCESS}, nil)
			apr.On("Send", mock.Anything, c.send, mock.Anything).
				Return(&gate.CodeResponse{Code: gate.ResultCode_SUCCESS}, nil)

			cli := newClient("test", apr)
			cli.now = now

			err := cli.OnReceivedMessage(context.Background(), c.req)

			if c.err != nil {
				assert.True(t, errors.Is(err, c.err))
			}

			if c.before != nil {
				c.before(cli)
			}

			if c.pub != nil {
				apr.AssertCalled(t, "Publish", mock.Anything, c.pub, mock.Anything)
			} else {
				apr.AssertNotCalled(t, "Publish", mock.Anything, c.pub, mock.Anything)
			}

			if c.sub != nil {
				apr.AssertCalled(t, "Subscribe", mock.Anything, c.sub, mock.Anything)
			} else {
				apr.AssertNotCalled(t, "Subscribe", mock.Anything, c.sub, mock.Anything)
			}

			if c.usub != nil {
				apr.AssertCalled(t, "Unsubscribe", mock.Anything, c.usub, mock.Anything)
			} else {
				apr.AssertNotCalled(t, "Unsubscribe", mock.Anything, c.usub, mock.Anything)
			}

			if c.send != nil {
				apr.AssertCalled(t, "Send", mock.Anything, c.send, mock.Anything)
			} else {
				apr.AssertNotCalled(t, "Send", mock.Anything, c.send, mock.Anything)
			}
		})
	}
}
