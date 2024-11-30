package vcas

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type packet struct {
	Topic  string  `vcas:"name,n"`
	Method Method  `vcas:"method,meth,m"`
	Value  any     `vcas:"val,value,v"`
	Stamp  Time    `vcas:"time,t"`
	Float  float64 `vcas:"float"`
	Int    int64   `vcas:"int"`
	Extra  int64
	close  int64 `vcas:"close"`
}

func TestMethodMarshalPub(t *testing.T) {
	m := PUB
	b, err := m.MarshalText()

	assert.Nil(t, err)
	assert.Equal(t, string(b), "set")
}

func TestMethodMarshalSub(t *testing.T) {
	m := SUB
	b, err := m.MarshalText()

	assert.Nil(t, err)
	assert.Equal(t, string(b), "subscribe")
}

func TestMethodMarshalUsb(t *testing.T) {
	m := USB
	b, err := m.MarshalText()

	assert.Nil(t, err)
	assert.Equal(t, string(b), "release")
}

func TestMethodMarshalGet(t *testing.T) {
	m := GET
	b, err := m.MarshalText()

	assert.Nil(t, err)
	assert.Equal(t, string(b), "get")
}

func TestMethodMarshalUnknownError(t *testing.T) {
	m := Method(-1)
	_, err := m.MarshalText()

	assert.NotNil(t, err)
}

func TestMethodUnmarshalPub(t *testing.T) {
	var m Method

	assert.Nil(t, m.UnmarshalText([]byte("s")))
	assert.Equal(t, PUB, m)

	assert.Nil(t, m.UnmarshalText([]byte("set")))
	assert.Equal(t, PUB, m)
}

func TestMethodUnmarshalSub(t *testing.T) {
	var m Method

	assert.Nil(t, m.UnmarshalText([]byte("sb")))
	assert.Equal(t, SUB, m)

	assert.Nil(t, m.UnmarshalText([]byte("subscr")))
	assert.Equal(t, SUB, m)

	assert.Nil(t, m.UnmarshalText([]byte("subscribe")))
	assert.Equal(t, SUB, m)
}

func TestMethodUnmarshalUsb(t *testing.T) {
	var m Method

	assert.Nil(t, m.UnmarshalText([]byte("rel")))
	assert.Equal(t, USB, m)

	assert.Nil(t, m.UnmarshalText([]byte("release")))
	assert.Equal(t, USB, m)
}

func TestMethodUnmarshalGet(t *testing.T) {
	var m Method

	assert.Nil(t, m.UnmarshalText([]byte("g")))
	assert.Equal(t, GET, m)

	assert.Nil(t, m.UnmarshalText([]byte("gf")))
	assert.Equal(t, GET, m)

	assert.Nil(t, m.UnmarshalText([]byte("get")))
	assert.Equal(t, GET, m)

	assert.Nil(t, m.UnmarshalText([]byte("getfull")))
	assert.Equal(t, GET, m)
}

func TestMethodUnmarshalUnknownError(t *testing.T) {
	var m Method

	assert.NotNil(t, m.UnmarshalText([]byte("unknown")))
}

func TestTimeMarshal(t *testing.T) {
	s := Time{time.UnixMilli(1118509199999)}
	b, err := s.MarshalText()

	assert.Nil(t, err)
	assert.Equal(t, "11.06.2005 23_59_59.999", string(b))
}

func TestTimeUnmarshal(t *testing.T) {
	var s Time

	assert.Nil(t, s.UnmarshalText([]byte("11.06.2005 23_59_59.999")))
	assert.Equal(t, int64(1118509199999), s.UnixMilli())
}

func TestTimeUnmarshalMalformedError(t *testing.T) {
	var s Time

	assert.NotNil(t, s.UnmarshalText([]byte("11.06.2005 23:59:59.999")))
}

func TestTimeMarshalJson(t *testing.T) {
	s := Time{time.UnixMilli(1118509199999)}
	b, err := s.MarshalJSON()

	assert.Nil(t, err)
	assert.Equal(t, `1118509199999`, string(b))
}

func TestTimeUnmarshalJson(t *testing.T) {
	var s Time

	assert.Nil(t, s.UnmarshalJSON([]byte("1118509199999")))
	assert.Equal(t, int64(1118509199999), s.UnixMilli())
}

func TestTimeUnmarshalJsonMalformedError(t *testing.T) {
	var s Time

	assert.NotNil(t, s.UnmarshalJSON([]byte("11.06.2005 23:59:59.999")))
}

func TestMarshal(t *testing.T) {
	p := packet{
		Topic:  "test",
		Method: PUB,
		Value:  11.06,
		Stamp:  Time{time.UnixMilli(1118509199999)},
	}

	b, err := Marshal(&p)

	assert.Nil(t, err)
	assert.Equal(t, "descr:-|type:rw|units:-|name:test|method:set|val:11.06|time:11.06.2005 23_59_59.999|float:0|int:0", string(b))
}

func TestMarshalUnknownMethodError(t *testing.T) {
	p := packet{
		Topic:  "test",
		Method: Method(-1),
		Value:  11.06,
		Stamp:  Time{time.UnixMilli(1118509199999)},
	}

	_, err := Marshal(&p)

	assert.NotNil(t, err)
}

func TestUnmarshal(t *testing.T) {
	var p packet

	assert.Nil(t, Unmarshal([]byte("time:11.06.2005 23_59_59.999|name:test|m:get|value:11.06|float:0.5|int:10"), &p))
	assert.Equal(t, "test", p.Topic)
	assert.Equal(t, GET, p.Method)
	assert.Equal(t, "11.06", p.Value)
	assert.Equal(t, int64(1118509199999), p.Stamp.UnixMilli())
	assert.Equal(t, float64(0.5), p.Float)
	assert.Equal(t, int64(10), p.Int)
	assert.Equal(t, int64(0), p.Extra)
}

func TestUnmarshalWrongDestinationError(t *testing.T) {
	var p int

	assert.NotNil(t, Unmarshal([]byte("time:11.06.2005 23_59_59.999|name:test|m:get|value:11.06"), &p))
}

func TestUnmarshalMalformedToken(t *testing.T) {
	var p packet

	assert.Nil(t, Unmarshal([]byte("time:11.06.2005 23_59_59.999|name:test|extra|m:get|value:11.06"), &p))
	assert.Equal(t, "test", p.Topic)
	assert.Equal(t, GET, p.Method)
	assert.Equal(t, "11.06", p.Value)
	assert.Equal(t, int64(1118509199999), p.Stamp.UnixMilli())
	assert.Equal(t, float64(0), p.Float)
	assert.Equal(t, int64(0), p.Int)
	assert.Equal(t, int64(0), p.Extra)
}

func TestUnmarshalMalformedTimeError(t *testing.T) {
	var p packet

	assert.NotNil(t, Unmarshal([]byte("time:11.06.2005 23:59:59.999|name:test|extra|m:get|value:11.06"), &p))
}

func TestUnmarshalMalformedFloat(t *testing.T) {
	var p packet

	assert.Nil(t, Unmarshal([]byte("time:11.06.2005 23_59_59.999|name:test|m:get|value:11.06|float:error|int:10"), &p))
	assert.Equal(t, "test", p.Topic)
	assert.Equal(t, GET, p.Method)
	assert.Equal(t, "11.06", p.Value)
	assert.Equal(t, int64(1118509199999), p.Stamp.UnixMilli())
	assert.Equal(t, float64(0), p.Float)
	assert.Equal(t, int64(10), p.Int)
	assert.Equal(t, int64(0), p.Extra)
}

func TestUnmarshalMalformedInt(t *testing.T) {
	var p packet

	assert.Nil(t, Unmarshal([]byte("time:11.06.2005 23_59_59.999|name:test|m:get|value:11.06|float:0.5|int:error"), &p))
	assert.Equal(t, "test", p.Topic)
	assert.Equal(t, GET, p.Method)
	assert.Equal(t, "11.06", p.Value)
	assert.Equal(t, int64(1118509199999), p.Stamp.UnixMilli())
	assert.Equal(t, float64(0.5), p.Float)
	assert.Equal(t, int64(0), p.Int)
	assert.Equal(t, int64(0), p.Extra)
}

type unsupported struct {
	Packet packet `vcas:"packet"`
}

func TestUnmarshalUnsupportedType(t *testing.T) {
	var u unsupported

	assert.NotNil(t, Unmarshal([]byte("packet:value"), &u))
}
