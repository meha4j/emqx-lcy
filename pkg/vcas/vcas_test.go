package vcas

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type person struct {
	Name string `vcas:"name"`
	Age  int    `vcas:"age"`
}

type student struct {
	person

	Score float64 `vcas:"score"`
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

func TestUnmarshalInt(t *testing.T) {
	var s int

	assert.Nil(t, Unmarshal([]byte("10"), &s))
	assert.Equal(t, 10, s)
}

func TestUnmarshalInt8(t *testing.T) {
	var s int8

	assert.Nil(t, Unmarshal([]byte("10"), &s))
	assert.Equal(t, int8(10), s)
}

func TestUnmarshalInt16(t *testing.T) {
	var s int16

	assert.Nil(t, Unmarshal([]byte("10"), &s))
	assert.Equal(t, int16(10), s)
}

func TestUnmarshalInt32(t *testing.T) {
	var s int32

	assert.Nil(t, Unmarshal([]byte("10"), &s))
	assert.Equal(t, int32(10), s)
}

func TestUnmarshalInt64(t *testing.T) {
	var s int64

	assert.Nil(t, Unmarshal([]byte("10"), &s))
	assert.Equal(t, int64(10), s)
}

func TestUnmarshalUint(t *testing.T) {
	var s uint

	assert.Nil(t, Unmarshal([]byte("10"), &s))
	assert.Equal(t, uint(10), s)
}

func TestUnmarshalUint8(t *testing.T) {
	var s uint8

	assert.Nil(t, Unmarshal([]byte("10"), &s))
	assert.Equal(t, uint8(10), s)
}

func TestUnmarshalUint16(t *testing.T) {
	var s uint16

	assert.Nil(t, Unmarshal([]byte("10"), &s))
	assert.Equal(t, uint16(10), s)
}

func TestUnmarshalUint32(t *testing.T) {
	var s uint32

	assert.Nil(t, Unmarshal([]byte("10"), &s))
	assert.Equal(t, uint32(10), s)
}

func TestUnmarshalUint64(t *testing.T) {
	var s uint64

	assert.Nil(t, Unmarshal([]byte("10"), &s))
	assert.Equal(t, uint64(10), s)
}

func TestUnmarshalFloat32(t *testing.T) {
	var s float32

	assert.Nil(t, Unmarshal([]byte("11.06"), &s))
	assert.Equal(t, float32(11.06), s)
}

func TestUnmarshalFloat64(t *testing.T) {
	var s float64

	assert.Nil(t, Unmarshal([]byte("11.06"), &s))
	assert.Equal(t, float64(11.06), s)
}

func TestUnmarshalString(t *testing.T) {
	var s string

	assert.Nil(t, Unmarshal([]byte("11.06"), &s))
	assert.Equal(t, "11.06", s)
}

func TestUnmarshalInterface(t *testing.T) {
	var s any

	assert.Nil(t, Unmarshal([]byte("11.06"), &s))
	assert.Equal(t, "11.06", s)
}

func TestUnmarshalBool(t *testing.T) {
	var s bool = false

	assert.Nil(t, Unmarshal([]byte("true"), &s))
	assert.Equal(t, true, s)
}

func TestUnmarshalStruct(t *testing.T) {
	var p person

	assert.Nil(t, Unmarshal([]byte("name:Dasha|age:18"), &p))
	assert.Equal(t, "Dasha", p.Name)
	assert.Equal(t, 18, p.Age)
}

func TestUnmarshalEmbeddedStruct(t *testing.T) {
	var s student

	assert.Nil(t, Unmarshal([]byte("name:Dasha|age:18|score:5.0"), &s))
	assert.Equal(t, "Dasha", s.Name)
	assert.Equal(t, 18, s.Age)
	assert.Equal(t, float64(5.0), s.Score)
}

func TestUnmarshalMap(t *testing.T) {
	var m map[string]string = make(map[string]string, 2)

	assert.Nil(t, Unmarshal([]byte("name:Dasha|age:18"), &m))
	assert.Contains(t, m, "name")
	assert.Contains(t, m, "age")
}

func TestMarshalInt(t *testing.T) {
	m, err := Marshal(int(10))

	assert.Nil(t, err)
	assert.Equal(t, "10", string(m))
}

func TestMarshalInt8(t *testing.T) {
	m, err := Marshal(int8(10))

	assert.Nil(t, err)
	assert.Equal(t, "10", string(m))
}

func TestMarshalInt16(t *testing.T) {
	m, err := Marshal(int16(10))

	assert.Nil(t, err)
	assert.Equal(t, "10", string(m))
}

func TestMarshalInt32(t *testing.T) {
	m, err := Marshal(int32(10))

	assert.Nil(t, err)
	assert.Equal(t, "10", string(m))
}

func TestMarshalInt64(t *testing.T) {
	m, err := Marshal(int64(10))

	assert.Nil(t, err)
	assert.Equal(t, "10", string(m))
}

func TestMarshalUint(t *testing.T) {
	m, err := Marshal(uint(10))

	assert.Nil(t, err)
	assert.Equal(t, "10", string(m))
}

func TestMarshalUint8(t *testing.T) {
	m, err := Marshal(uint8(10))

	assert.Nil(t, err)
	assert.Equal(t, "10", string(m))
}

func TestMarshalUint16(t *testing.T) {
	m, err := Marshal(uint16(10))

	assert.Nil(t, err)
	assert.Equal(t, "10", string(m))
}

func TestMarshalUint32(t *testing.T) {
	m, err := Marshal(uint32(10))

	assert.Nil(t, err)
	assert.Equal(t, "10", string(m))
}

func TestMarshalUint64(t *testing.T) {
	m, err := Marshal(uint64(10))

	assert.Nil(t, err)
	assert.Equal(t, "10", string(m))
}
func TestMarshalFloat32(t *testing.T) {
	m, err := Marshal(float32(11.06))

	assert.Nil(t, err)
	assert.Contains(t, string(m), "11.06")
}

func TestMarshalFloat64(t *testing.T) {
	m, err := Marshal(float64(11.06))

	assert.Nil(t, err)
	assert.Contains(t, string(m), "11.06")
}

func TestMarshalString(t *testing.T) {
	m, err := Marshal("11.06")

	assert.Nil(t, err)
	assert.Equal(t, "11.06", string(m))
}

func TestMarshalInterface(t *testing.T) {
	var s any = "11.06"

	m, err := Marshal(s)

	assert.Nil(t, err)
	assert.Equal(t, "11.06", string(m))
}

func TestMarshalBool(t *testing.T) {
	m, err := Marshal(true)

	assert.Nil(t, err)
	assert.Equal(t, "true", string(m))

	m, err = Marshal(false)

	assert.Nil(t, err)
	assert.Equal(t, "false", string(m))
}

func TestMarshalStruct(t *testing.T) {
	p := person{
		Name: "Dasha",
		Age:  18,
	}

	m, err := Marshal(&p)

	assert.Nil(t, err)
	assert.Equal(t, "name:Dasha|age:18", string(m))
}

func TestMarshalEmbeddedStruct(t *testing.T) {
	s := student{
		person: person{
			Name: "Dasha",
			Age:  18,
		},
		Score: 5.5,
	}

	m, err := Marshal(&s)

	assert.Nil(t, err)
	assert.Equal(t, "name:Dasha|age:18|score:5.5", string(m))
}

func TestMarshalMap(t *testing.T) {
	m := map[string]string{
		"name": "Dasha",
		"age":  "18",
	}

	r, err := Marshal(m)

	assert.Nil(t, err)
	assert.Contains(t, string(r), "name:Dasha")
	assert.Contains(t, string(r), "age:18")
}
