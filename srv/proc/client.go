package proc

import (
	context "context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	authapi "github.com/paraskun/extd/api/auth"
	procapi "github.com/paraskun/extd/api/proc"

	"github.com/paraskun/extd/pkg/vcas"
	"github.com/paraskun/extd/srv/auth"
)

type packet struct {
	Topic  string      `vcas:"name,n" json:"-"`
	Stamp  vcas.Time   `vcas:"time,t" json:"timestamp"`
	Method vcas.Method `vcas:"method,meth,m" json:"-"`
	Value  any         `vcas:"val,value,v" json:"value,omitempty"`

	Units string `vcas:"units" json:"-"`
	Descr string `vcas:"descr" json:"-"`
	Type  string `vcas:"type" json:"-"`
}

type Client struct {
	Con string

	ctl *auth.ACL
	obs string
	buf []byte
	pkt packet
	mux sync.Mutex

	adapter procapi.ConnectionAdapterClient
}

func NewClient(con string, ctl *auth.ACL, adapter procapi.ConnectionAdapterClient) *Client {
	buf := make([]byte, 0, 0xff)

	return &Client{
		Con: con,

		ctl: ctl,
		buf: buf,
		pkt: packet{
			Units: "none",
			Descr: "none",
			Type:  "none",
		},

		adapter: adapter,
	}
}

func (c *Client) OnReceivedBytes(ctx context.Context, msg []byte) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	for _, b := range msg {
		if b != 10 {
			c.buf = append(c.buf, b)
			continue
		}

		c.pkt.Stamp.Time = time.Now()
		c.pkt.Value = nil

		if err := vcas.Unmarshal(c.buf, &c.pkt); err != nil {
			return fmt.Errorf("unmarshal: %v", err)
		}

		if err := c.handlepacket(ctx, &c.pkt); err != nil {
			return fmt.Errorf("handle: %v", err)
		}

		if cap(c.buf) > 0xff {
			c.buf = make([]byte, 0, 0xff)
		} else {
			c.buf = c.buf[:0]
		}
	}

	return nil
}

func (c *Client) handlepacket(ctx context.Context, pkt *packet) error {
	if c.obs != "" {
		return nil
	}

	if pkt.Topic == "" {
		return fmt.Errorf("unknown topic")
	}

	switch c.pkt.Method {
	case vcas.PUB:
		if err := c.publish(ctx, &c.pkt); err != nil {
			return fmt.Errorf("publish: %v", err)
		}
	case vcas.SUB:
		if err := c.subscribe(ctx, c.pkt.Topic); err != nil {
			return fmt.Errorf("subscribe: %v", err)
		}
	case vcas.USB:
		if err := c.unsubscribe(ctx, c.pkt.Topic); err != nil {
			return fmt.Errorf("unsubscribe: %v", err)
		}
	case vcas.GET:
		if err := c.get(ctx, c.pkt.Topic); err != nil {
			return fmt.Errorf("get: %v", err)
		}
	default:
		return fmt.Errorf("unknown method")
	}

	return nil
}

func (c *Client) publish(ctx context.Context, pkt *packet) error {
	if !c.ctl.Check(pkt.Topic, c.Con, authapi.ClientAuthorizeRequest_PUBLISH) {
		return nil
	}

	pay, err := json.Marshal(pkt)

	if err != nil {
		return fmt.Errorf("marshal: %v", err)
	}

	res, err := c.adapter.Publish(ctx, &procapi.PublishRequest{
		Conn:    c.Con,
		Topic:   pkt.Topic,
		Qos:     0,
		Payload: pay,
	})

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	if res.GetCode() != procapi.ResultCode_SUCCESS {
		return fmt.Errorf("remote: %v", res.GetMessage())
	}

	return nil
}

func (c *Client) subscribe(ctx context.Context, top string) error {
	res, err := c.adapter.Subscribe(ctx, &procapi.SubscribeRequest{
		Conn:  c.Con,
		Topic: top,
		Qos:   2,
	})

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	if res.GetCode() != procapi.ResultCode_SUCCESS {
		return fmt.Errorf("remote: %v", res.GetMessage())
	}

	return nil
}

func (c *Client) unsubscribe(ctx context.Context, top string) error {
	res, err := c.adapter.Unsubscribe(ctx, &procapi.UnsubscribeRequest{
		Conn:  c.Con,
		Topic: top,
	})

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	if res.GetCode() != procapi.ResultCode_SUCCESS {
		return fmt.Errorf("remote: %v", res.GetMessage())
	}

	return nil
}

func (c *Client) get(ctx context.Context, top string) error {
	err := c.subscribe(ctx, top)

	if err != nil {
		return fmt.Errorf("subscribe: %v", err)
	}

	c.obs = top

	time.AfterFunc(5*time.Second, func() {
		c.mux.Lock()
		defer c.mux.Unlock()

		if c.obs != "" {
			c.pkt.Topic = c.obs
			c.pkt.Stamp.Time = time.Now()
			c.pkt.Value = nil

			c.unsubscribe(context.Background(), c.obs)
			c.send(context.Background(), &c.pkt)

			c.obs = ""
		}
	})

	return nil
}

func (c *Client) OnReceivedMessage(ctx context.Context, msg *procapi.Message) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.obs != "" {
		if c.obs != msg.GetTopic() {
			return nil
		}

		c.obs = ""

		if err := c.unsubscribe(ctx, msg.GetTopic()); err != nil {
			return fmt.Errorf("unsubscribe: %v", err)
		}
	}

	c.pkt.Topic = msg.GetTopic()
	c.pkt.Value = nil

	if err := json.Unmarshal(msg.GetPayload(), &c.pkt); err != nil {
		return fmt.Errorf("parse: %v", err)
	}

	if err := c.send(ctx, &c.pkt); err != nil {
		return fmt.Errorf("send: %v", err)
	}

	return nil
}

func (c *Client) send(ctx context.Context, pkt *packet) error {
	if pkt.Value == nil {
		pkt.Value = "none"
	}

	pkt.Method = vcas.PUB
	txt, err := vcas.Marshal(pkt)

	if err != nil {
		return fmt.Errorf("marshal: %v", err)
	}

	res, err := c.adapter.Send(ctx, &procapi.SendBytesRequest{
		Conn:  c.Con,
		Bytes: append(txt, 10),
	})

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	if res.GetCode() != procapi.ResultCode_SUCCESS {
		return fmt.Errorf("remote: %v", res.GetMessage())
	}

	return nil
}
