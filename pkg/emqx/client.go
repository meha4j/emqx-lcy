package emqx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

type protoGatewayRequest struct {
	Handler struct {
		Address string
	}

	Server struct {
		Bind string
	}

	Enable     bool
	Statistics bool   `json:"enable_stats"`
	Timeout    string `json:"idle_timeout"`
	Mountpoint string

	Listeners []struct {
		Name string
		Type string
		Bind string
	}
}

type Client struct {
	BaseUrl string

	conn *http.Client
	addr *net.TCPAddr
	name string
	pass string
}

func NewClient(url, name, pass string) *Client {
	return &Client{
		BaseUrl: url,

		conn: &http.Client{},
		addr: getOutbound(),
		name: name,
		pass: pass,
	}
}

func (c *Client) UpdateProtoGateway() error {
	pay := protoGatewayRequest{
		Handler: struct{ Address string }{
			Address: c.addr.String(),
		},

		Server: struct{ Bind string }{
			Bind: "9110",
		},

		Enable:     false,
		Statistics: true,
		Timeout:    "30s",
		Mountpoint: "vcas/",

		Listeners: []struct {
			Name string
			Type string
			Bind string
		}{
			{
				Name: "default",
				Type: "tcp",
				Bind: "20041",
			},
		},
	}

	bin, err := json.Marshal(pay)

	if err != nil {
		return fmt.Errorf("marshal body: %v", err)
	}

	url := c.BaseUrl + "/gateways/exproto"
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(bin))

	if err != nil {
		return fmt.Errorf("new request: %v", err)
	}

	req.SetBasicAuth(c.name, c.pass)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.conn.Do(req)

	if err != nil {
		return fmt.Errorf("client: %v", err)
	}

	defer res.Body.Close()

	if res.StatusCode != 204 {
		var buf bytes.Buffer

		_, err := buf.ReadFrom(res.Body)

		if err != nil {
			return fmt.Errorf("response: %v", err)
		}

		return fmt.Errorf(buf.String())
	}

	return nil
}

func getOutbound() *net.TCPAddr {
	con, err := net.Dial("tcp", "8.8.8.8:80")

	if err != nil {
		panic("could not access base endpoint")
	}

	defer con.Close()
	return con.LocalAddr().(*net.TCPAddr)
}
