package proc

import (
	context "context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/paraskun/extd/api/proc"
	"github.com/paraskun/extd/pkg/vcas"
	"go.uber.org/zap"
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

type observer = string

type Client struct {
	Conn string
	Log  *zap.SugaredLogger

	obs observer
	pkt packet
	buf []byte
	mut sync.Mutex

	adapter proc.ConnectionAdapterClient
}

func NewClient(conn string, adapter proc.ConnectionAdapterClient, log *zap.SugaredLogger) *Client {
	return &Client{
		Conn: conn,
		Log:  log,

		buf: make([]byte, 0, 0xff),

		adapter: adapter,
	}
}

func (c *Client) OnReceivedBytes(ctx context.Context, msg []byte) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	for _, b := range msg {
		if b != 10 {
			c.buf = append(c.buf, b)
			continue
		}

		if err := vcas.Unmarshal(c.buf, &c.pkt); err != nil {
			return fmt.Errorf("unmarshal: %v", err)
		}

		if err := c.handlePacket(ctx, &c.pkt); err != nil {
			return fmt.Errorf("handle packet: %v", err)
		}

		if cap(c.buf) > 0xff {
			c.buf = make([]byte, 0, 0xff)
		} else {
			c.buf = c.buf[:0]
		}
	}

	return nil
}

func (c *Client) handlePacket(ctx context.Context, pkt *packet) error {
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
	if pkt.Stamp.Time.IsZero() {
		pkt.Stamp.Time = time.Now()
	}

	pay, err := json.Marshal(pkt)

	if err != nil {
		return fmt.Errorf("marshal: %v", err)
	}

	res, err := c.adapter.Publish(ctx, &proc.PublishRequest{
		Conn:    c.Conn,
		Topic:   pkt.Topic,
		Qos:     0,
		Payload: pay,
	})

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	if res.GetCode() != proc.ResultCode_SUCCESS {
		return fmt.Errorf(res.GetMessage())
	}

	return nil
}

func (c *Client) subscribe(ctx context.Context, top string) error {
	res, err := c.adapter.Subscribe(ctx, &proc.SubscribeRequest{
		Conn:  c.Conn,
		Topic: top,
		Qos:   2,
	})

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	if res.GetCode() != proc.ResultCode_SUCCESS {
		return fmt.Errorf(res.GetMessage())
	}

	return nil
}

func (c *Client) unsubscribe(ctx context.Context, top string) error {
	res, err := c.adapter.Unsubscribe(ctx, &proc.UnsubscribeRequest{
		Conn:  c.Conn,
		Topic: top,
	})

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	if res.GetCode() != proc.ResultCode_SUCCESS {
		return fmt.Errorf(res.GetMessage())
	}

	return nil
}

func (c *Client) get(ctx context.Context, top string) error {
	err := c.subscribe(ctx, top)

	if err != nil {
		return fmt.Errorf("subscribe: %v", err)
	}

	c.obs = top

	c.Log.Debug("callback registered for %s", top)

	time.AfterFunc(5*time.Second, func() {
		c.mut.Lock()
		defer c.mut.Unlock()

		c.Log.Debugf("callback activated for %s", top)

		if c.obs != "" {
			c.Log.Debugf("no message found, sending placeholder for %s", top)

			p := &packet{Topic: c.obs}

			c.unsubscribe(context.Background(), c.obs)
			c.send(context.Background(), p)

			c.obs = ""
		}
	})

	return nil
}

func (c *Client) OnReceivedMessage(ctx context.Context, msg *proc.Message) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	if c.obs != "" {
		if c.obs != msg.GetTopic() {
			c.Log.Debugf("message ignored for %s, waiting for %s", msg.GetTopic(), c.obs)
			return nil
		}

		c.Log.Debugf("message found for %s", c.obs)
		c.obs = ""

		if err := c.unsubscribe(ctx, msg.GetTopic()); err != nil {
			return fmt.Errorf("unsubscribe: %v", err)
		}
	}

	pkt := &packet{Topic: msg.GetTopic()}
	err := json.Unmarshal(msg.GetPayload(), pkt)

	if err != nil {
		return fmt.Errorf("parse: %v", err)
	}

	if err := c.send(ctx, pkt); err != nil {
		return fmt.Errorf("send: %v", err)
	}

	return nil
}

func (c *Client) send(ctx context.Context, pkt *packet) error {
	if pkt.Value == nil {
		pkt.Value = "none"
	}

	pkt.Method = vcas.PUB
	pkt.Descr = "none"
	pkt.Units = "none"
	pkt.Type = "rw"

	txt, err := vcas.Marshal(pkt)

	if err != nil {
		return fmt.Errorf("marshal: %v", err)
	}

	res, err := c.adapter.Send(ctx, &proc.SendBytesRequest{
		Conn:  c.Conn,
		Bytes: append(txt, 10),
	})

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	if res.GetCode() != proc.ResultCode_SUCCESS {
		return fmt.Errorf(res.GetMessage())
	}

	return nil
}
