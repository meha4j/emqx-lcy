package emqx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

type Server struct {
	Bind string `json:"bind"`
}

type Handler struct {
	Addr string `json:"address"`
}

type Listener struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Bind string `json:"bind"`
}

type ExProtoGatewayUpdateRequest struct {
	Name       string     `json:"name"`
	Timeout    string     `json:"idle_timeout"`
	Mountpoint string     `json:"mountpoint"`
	Enable     bool       `json:"enable"`
	Statistics bool       `json:"enable_stats"`
	Server     Server     `json:"server"`
	Handler    Handler    `json:"handler"`
	Listeners  []Listener `json:"listeners"`
}

type ExHookServerUpdateRequest struct {
	Name      string `json:"name"`
	Enable    bool   `json:"enable"`
	Addr      string `json:"url"`
	Timeout   string `json:"request_timeout,omitempty"`
	Action    string `json:"failed_action,omitempty"`
	Reconnect string `json:"auto_reconnect,omitempty"`
	PoolSize  int    `json:"pool_size,omitempty"`
}

type options struct {
	addr *string
	user *string
	pass *string
	tout *time.Duration
	rmax *int
}

type Option func(*options) error

func WithAddr(addr string) Option {
	return func(o *options) error {
		o.addr = &addr
		return nil
	}
}

func WithUser(user string) Option {
	return func(o *options) error {
		o.user = &user
		return nil
	}
}

func WithPass(pass string) Option {
	return func(o *options) error {
		o.pass = &pass
		return nil
	}
}

func WithTimeout(tout string) Option {
	return func(o *options) error {
		res, err := time.ParseDuration(tout)

		if err != nil {
			return fmt.Errorf("parse: %v", err)
		}

		o.tout = &res
		return nil
	}
}

func WithRetries(rmax int) Option {
	return func(o *options) error {
		if rmax < 0 {
			return fmt.Errorf("negative retries count")
		}

		o.rmax = &rmax
		return nil
	}
}

type Client struct {
	Base string
	Addr string

	conn *http.Client
	user string
	pass string
	tout time.Duration
	rmax int
}

func NewClient(base string, opts ...Option) (*Client, error) {
	var opt options

	for _, exe := range opts {
		if err := exe(&opt); err != nil {
			return nil, fmt.Errorf("opt: %v", err)
		}
	}

	cli := &Client{
		Base: base,
		conn: &http.Client{},
	}

	if opt.addr != nil {
		cli.Addr = *opt.addr
	}

	if opt.user != nil {
		cli.user = *opt.user
	}

	if opt.pass != nil {
		cli.pass = *opt.pass
	}

	if opt.tout != nil {
		cli.tout = *opt.tout
	}

	if opt.rmax != nil {
		cli.rmax = *opt.rmax
	}

	return cli, nil
}

func (c *Client) LookupAddress(host string) error {
	addr, err := net.LookupIP(host)

	if err != nil {
		return fmt.Errorf("req: %v", err)
	}

	if len(addr) == 0 {
		return fmt.Errorf("no record")
	}

	c.Addr = addr[0].String()

	return nil
}

func (c *Client) Do(req *http.Request) (res *http.Response, err error) {
	req.SetBasicAuth(c.user, c.pass)

	for r := 1; r <= c.rmax+1; r++ {
		res, err = c.conn.Do(req)

		if err == nil {
			break
		}

		log.Printf("request failed [%d/%d]: %v\n", r, c.rmax+1, err)
		time.Sleep(c.tout)
	}

	return
}

func (c *Client) UpdateExProtoGateway(pay *ExProtoGatewayUpdateRequest) error {
	bin, err := json.Marshal(pay)

	if err != nil {
		return fmt.Errorf("pay: %v", err)
	}

	url := c.Base + "/gateways/exproto"
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(bin))

	if err != nil {
		return fmt.Errorf("req: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := c.Do(req)

	if err != nil {
		return fmt.Errorf("req: %v", err)
	}

	defer res.Body.Close()

	if res.StatusCode != 204 {
		var buf bytes.Buffer

		_, err := buf.ReadFrom(res.Body)

		if err != nil {
			return fmt.Errorf("res: %v", err)
		}

		return fmt.Errorf("%v", buf.String())
	}

	return nil
}

func (c *Client) UpdateExHookServer(pay *ExHookServerUpdateRequest) error {
	ok, err := c.CheckExHookServer(pay.Name)

	if err != nil {
		return fmt.Errorf("check: %v", err)
	}

	bin, err := json.Marshal(pay)

	if err != nil {
		return fmt.Errorf("pay: %v", err)
	}

	url := c.Base + "/exhooks"
	mod := http.MethodPost

	if ok {
		url += "/" + pay.Name
		mod = http.MethodPut
	}

	req, err := http.NewRequest(mod, url, bytes.NewReader(bin))

	if err != nil {
		return fmt.Errorf("req: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := c.Do(req)

	if err != nil {
		return fmt.Errorf("req: %v", err)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		var buf bytes.Buffer

		_, err := buf.ReadFrom(res.Body)

		if err != nil {
			return fmt.Errorf("res: %v", err)
		}

		return fmt.Errorf("%v", buf.String())
	}

	return nil
}

func (c *Client) CheckExHookServer(name string) (bool, error) {
	url := c.Base + "/exhooks/" + name
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return false, fmt.Errorf("req: %v", err)
	}

	res, err := c.Do(req)

	if err != nil {
		return false, fmt.Errorf("req: %v", err)
	}

	defer res.Body.Close()

	if res.StatusCode == 200 {
		return true, nil
	}

	return false, nil
}
