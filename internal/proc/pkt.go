package proc

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	Stamp = "11.06.2005 23_59_59.999"
)

const (
	PUB = iota
	SUB = iota
	USB = iota
	GET = iota
)

type Time struct {
	time.Time
}

func (t *Time) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(t.UnixMilli(), 10)), nil
}

func (t *Time) UnmarshalJSON(json []byte) error {
	milli, err := strconv.ParseInt(string(json), 10, 64)

	if err != nil {
		return err
	}

	t.Time = time.UnixMilli(milli)

	return nil
}

type Record struct {
	Value string `json:"value"`
	Stamp Time   `json:"timestamp"`
}

type Event struct {
	Topic  string
	Method int

	Record
}

func (e *Event) Decode(in []byte) error {
	tok := strings.Split(string(in), "|")

	for _, t := range tok {
		tok := strings.Split(t, ":")

		if len(tok) != 2 {
			continue
		}

		switch tok[0] {
		case "n":
		case "name":
			e.Topic = tok[1]
		case "time":
		}
	}

	return nil
}

func (e *Event) Encode() ([]byte, error) {
	msg := fmt.Sprintf("time:%v|name:%v|descr:-|units:-|type:rw|val:%v\n",
		e.Stamp.Format(Stamp),
		e.Topic,
		e.Value,
	)

	return []byte(msg), nil
}
