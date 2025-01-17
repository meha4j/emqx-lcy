package vcas

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

const (
	Name    = "VCAS"
	Version = "1.0-SNAPSHOT"

	Stamp = "02.01.2006 15_04_05.000"

	PUB Method = iota + 1
	SUB
	USB
	GET

	OuterSep = '|'
	InnerSep = ':'
)

type Method int

func (m Method) marshal(buf *bytes.Buffer) error {
	switch m {
	case PUB:
		buf.WriteString("set")
	case SUB:
		buf.WriteString("subscribe")
	case USB:
		buf.WriteString("release")
	case GET:
		buf.WriteString("get")
	default:
		return fmt.Errorf("unknown: %v", m)
	}

	return nil
}

func (m *Method) unmarshal(b []byte) error {
	s := string(b)

	switch s {
	case "s", "set":
		*m = PUB
	case "sb", "subscr", "subscribe":
		*m = SUB
	case "rel", "release":
		*m = USB
	case "g", "gf", "get", "getfull":
		*m = GET
	default:
		return fmt.Errorf("unknown: %v", s)
	}

	return nil
}

type Time struct {
	time.Time
}

func (t Time) marshal(buf *bytes.Buffer) error {
	buf.WriteString(t.Format(Stamp))

	return nil
}

func (t *Time) unmarshal(b []byte) error {
	tm, err := time.ParseInLocation(Stamp, string(b), time.Local)

	if err != nil {
		return fmt.Errorf("format: %v", err)
	}

	t.Time = tm

	return nil
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(t.UnixMilli(), 10)), nil
}

func (t *Time) UnmarshalJSON(json []byte) error {
	milli, err := strconv.ParseInt(string(json), 10, 64)

	if err != nil {
		return fmt.Errorf("format: %v", err)
	}

	t.Time = time.UnixMilli(milli)

	return nil
}

type Packet struct {
	Method Method `json:"-"`
	Stamp  Time   `json:"stamp"`
	Topic  string `json:"-"`
	Value  string `json:"value,omitempty"`
}

func (pkt *Packet) Marshal(pay []byte) ([]byte, error) {
  if pkt.Topic == "" {
    return nil, fmt.Errorf("topic: not found")
  }

  if pkt.Value == "" {
    pkt.Value = "none"
  }

	buf := bytes.NewBuffer(pay)

  buf.Grow(63 + len(pkt.Topic) + len(pkt.Value))
	buf.WriteString("time:")

	if err := pkt.Stamp.marshal(buf); err != nil {
		return nil, fmt.Errorf("time: %v", err)
	}

	buf.WriteString("|method:")

	if err := pkt.Method.marshal(buf); err != nil {
		return nil, fmt.Errorf("method: %v", err)
	}

	buf.WriteString("|name:")
	buf.WriteString(pkt.Topic)
	buf.WriteString("|val:")
	buf.WriteString(pkt.Value)
  buf.WriteString("|descr:none|type:rw|units:none\n")

	return buf.Bytes(), nil
}

func (pkt *Packet) Unmarshal(pay []byte) error {
	for _, tok := range bytes.Split(pay, []byte{OuterSep}) {
		tok := bytes.SplitN(tok, []byte{InnerSep}, 2)

		if len(tok) != 2 {
			continue
		}

		k := string(bytes.Trim(tok[0], "\n\t\r "))
		v := tok[1]

		switch k {
		case "method", "meth", "m":
			if err := pkt.Method.unmarshal(v); err != nil {
				return fmt.Errorf("method: %v", err)
			}
		case "time", "t":
			if err := pkt.Stamp.unmarshal(v); err != nil {
				return fmt.Errorf("time: %v", err)
			}
		case "name", "n":
			pkt.Topic = string(v)
		case "value", "val", "v":
			pkt.Value = string(v)
		}
	}

  if pkt.Value == "none" {
    pkt.Value = ""
  }

	return nil
}
