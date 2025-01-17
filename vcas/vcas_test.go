package vcas

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	cases := map[string]struct {
		inp Packet
		exp struct {
			err bool
			res string
		}
	}{
		`publish`: {
			inp: Packet{
				Method: PUB,
				Topic:  "test",
				Stamp:  Time{time.UnixMilli(1118509199999)},
				Value:  "11.06",
			},
			exp: struct {
				err bool
				res string
			}{
				err: false,
				res: "time:11.06.2005 23_59_59.999|method:set|name:test|val:11.06|descr:none|type:rw|units:none\n",
			},
		},
		`subscribe`: {
			inp: Packet{
				Method: SUB,
				Topic:  "test",
				Stamp:  Time{time.UnixMilli(1118509199999)},
				Value:  "11.06",
			},
			exp: struct {
				err bool
				res string
			}{
				err: false,
				res: "time:11.06.2005 23_59_59.999|method:subscribe|name:test|val:11.06|descr:none|type:rw|units:none\n",
			},
		},
		`unsubscribe`: {
			inp: Packet{
				Method: USB,
				Topic:  "test",
				Stamp:  Time{time.UnixMilli(1118509199999)},
				Value:  "11.06",
			},
			exp: struct {
				err bool
				res string
			}{
				err: false,
				res: "time:11.06.2005 23_59_59.999|method:release|name:test|val:11.06|descr:none|type:rw|units:none\n",
			},
		},
		`get`: {
			inp: Packet{
				Method: GET,
				Topic:  "test",
				Stamp:  Time{time.UnixMilli(1118509199999)},
				Value:  "11.06",
			},
			exp: struct {
				err bool
				res string
			}{
				err: false,
				res: "time:11.06.2005 23_59_59.999|method:get|name:test|val:11.06|descr:none|type:rw|units:none\n",
			},
		},
		`without value`: {
			inp: Packet{
				Method: PUB,
				Topic:  "test",
				Stamp:  Time{time.UnixMilli(1118509199999)},
			},
			exp: struct {
				err bool
				res string
			}{
				err: false,
				res: "time:11.06.2005 23_59_59.999|method:set|name:test|val:none|descr:none|type:rw|units:none\n",
			},
		},
		`with unknown method`: {
			inp: Packet{
				Topic: "test",
				Stamp: Time{time.UnixMilli(1118509199999)},
			},
			exp: struct {
				err bool
				res string
			}{
				err: true,
			},
		},
		`without name`: {
			inp: Packet{
				Method: PUB,
				Stamp:  Time{time.UnixMilli(1118509199999)},
			},
			exp: struct {
				err bool
				res string
			}{
				err: true,
			},
		},
	}

	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			res := make([]byte, 0)
			res, err := data.inp.Marshal(res)

			if !data.exp.err {
				assert.Nil(t, err)
				assert.Equal(t, data.exp.res, string(res))
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func BenchmarkMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		pay := make([]byte, 0)
		pkt := Packet{
			Method: PUB,
			Topic:  "VEPP/CCD/1M1L/sigma_x",
			Stamp:  Time{time.UnixMilli(1118509199999)},
			Value:  strconv.FormatFloat(rand.Float64()*100, 'f', 7, 64),
		}

		b.StartTimer()
		pkt.Marshal(pay)
	}
}

func TestUnmarshal(t *testing.T) {
	cases := map[string]struct {
		inp string
		exp struct {
			err bool
			res Packet
		}
	}{
		`pub(set)`: {
			inp: "time:11.06.2005 23_59_59.999|method:set|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: PUB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`pub(s)`: {
			inp: "time:11.06.2005 23_59_59.999|method:s|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: PUB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`sub(subscribe)`: {
			inp: "time:11.06.2005 23_59_59.999|method:subscribe|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: SUB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`sub(subscr)`: {
			inp: "time:11.06.2005 23_59_59.999|method:subscr|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: SUB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`sub(sb)`: {
			inp: "time:11.06.2005 23_59_59.999|method:sb|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: SUB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`usub(release)`: {
			inp: "time:11.06.2005 23_59_59.999|method:release|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: USB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`usub(rel)`: {
			inp: "time:11.06.2005 23_59_59.999|method:rel|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: USB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`get(getfull)`: {
			inp: "time:11.06.2005 23_59_59.999|method:getfull|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: GET,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`get(get)`: {
			inp: "time:11.06.2005 23_59_59.999|method:get|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: GET,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`get(gf)`: {
			inp: "time:11.06.2005 23_59_59.999|method:gf|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: GET,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`get(g)`: {
			inp: "time:11.06.2005 23_59_59.999|method:g|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: GET,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`time(t)`: {
			inp: "t:11.06.2005 23_59_59.999|method:set|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: PUB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`method(meth)`: {
			inp: "time:11.06.2005 23_59_59.999|meth:set|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: PUB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`method(m)`: {
			inp: "time:11.06.2005 23_59_59.999|m:set|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: PUB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`name(n)`: {
			inp: "time:11.06.2005 23_59_59.999|method:set|n:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: PUB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`value(value)`: {
			inp: "time:11.06.2005 23_59_59.999|method:set|name:test|value:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: PUB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`value(v)`: {
			inp: "time:11.06.2005 23_59_59.999|method:set|name:test|v:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: PUB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`with leading escapes`: {
			inp: "\n time:11.06.2005 23_59_59.999|method:set|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: PUB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`with trailing escapes`: {
			inp: "time:11.06.2005 23_59_59.999|method:set|name:test|val:11.06|descr:none|type:rw|units:none\n\t ",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: PUB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`with inner escapes`: {
			inp: "time:11.06.2005 23_59_59.999|\nmethod:set|name :test|val\t:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: PUB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`with extra token`: {
			inp: "time:11.06.2005 23_59_59.999|extra:none|method:set|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: PUB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`with empty value`: {
			inp: "time:11.06.2005 23_59_59.999|method:set|name:test|val:|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: PUB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
				},
			},
		},
		`with none value`: {
			inp: "time:11.06.2005 23_59_59.999|method:set|name:test|val:none|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: PUB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
				},
			},
		},
		`without value`: {
			inp: "time:11.06.2005 23_59_59.999|method:set|name:test|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: PUB,
					Topic:  "test",
					Stamp:  Time{time.UnixMilli(1118509199999)},
				},
			},
		},
		`with malformed time`: {
			inp: "time:11.06.2005 23:59:59.999|method:set|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: true,
			},
		},
		`without time`: {
			inp: "method:set|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: PUB,
					Topic:  "test",
					Value:  "11.06",
				},
			},
		},
		`without name`: {
			inp: "time:11.06.2005 23_59_59.999|method:set|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Method: PUB,
					Stamp:  Time{time.UnixMilli(1118509199999)},
					Value:  "11.06",
				},
			},
		},
		`without method`: {
			inp: "time:11.06.2005 23_59_59.999|name:test|val:11.06|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: false,
				res: Packet{
					Stamp: Time{time.UnixMilli(1118509199999)},
					Topic: "test",
					Value: "11.06",
				},
			},
		},
		`with unknown method`: {
			inp: "time:11.06.2005 23_59_59.999|method:extra|name:test|descr:none|type:rw|units:none",
			exp: struct {
				err bool
				res Packet
			}{
				err: true,
			},
		},
	}

	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			pkt := Packet{}
			err := pkt.Unmarshal([]byte(data.inp))

			if !data.exp.err {
				assert.Nil(t, err)
				assert.Equal(t, data.exp.res, pkt)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()

    pkt := Packet{}
		pay := []byte(fmt.Sprintf(
      "time:11.06.2005 23_59_59.999|method:subscribe|name:VEPP/CCD/1M1L/sigma_x|val:%s|descr:none|type:rw|units:none", 
      strconv.FormatFloat(rand.Float64()*100, 'f', 7, 64),
    ))

    b.StartTimer()
    pkt.Unmarshal(pay)
	}
}
