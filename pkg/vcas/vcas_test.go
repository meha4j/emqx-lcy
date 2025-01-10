package vcas

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type person struct {
	Name      string `vcas:"name"`
	BirthDate Time   `vcas:"birth"`
	Age       uint   `vcas:"age"`
}

type player struct {
	person

	Active   bool       `vcas:"active"`
	Score    []float64  `vcas:"score"`
	Location complex128 `vcas:"loc,pos"`
	Meta     any        `vcas:"meta"`
}

func TestUnmarshal(t *testing.T) {
	tests := map[string]struct {
		inp string
		exp struct {
			res player
			err bool
		}
	}{
		`complete`: {
			inp: `name:Dasha|birth:11.06.2005 23_59_59.999|age:18|active:true|score:9.7,9.8,9.9|loc:(6+8i)|meta:0x2`,
			exp: struct {
				res player
				err bool
			}{
				res: player{
					person: person{
						Name:      "Dasha",
						BirthDate: Time{time.UnixMilli(1118509199999)},
						Age:       18,
					},
					Active:   true,
					Score:    []float64{9.7, 9.8, 9.9},
					Location: complex(6, 8),
					Meta:     "0x2",
				},
				err: false,
			},
		},
		`field missed`: {
			inp: `name:Dasha|birth:11.06.2005 23_59_59.999|age:18|active:true|score:9.7,9.8,9.9|loc:(6+8i)`,
			exp: struct {
				res player
				err bool
			}{
				res: player{
					person: person{
						Name:      "Dasha",
						BirthDate: Time{time.UnixMilli(1118509199999)},
						Age:       18,
					},
					Active:   true,
					Score:    []float64{9.7, 9.8, 9.9},
					Location: complex(6, 8),
					Meta:     nil,
				},
				err: false,
			},
		},
		`muliple tag selection`: {
			inp: `name:Dasha|birth:11.06.2005 23_59_59.999|age:18|active:true|score:9.7,9.8,9.9|pos:(6+8i)|meta:0x2`,
			exp: struct {
				res player
				err bool
			}{
				res: player{
					person: person{
						Name:      "Dasha",
						BirthDate: Time{time.UnixMilli(1118509199999)},
						Age:       18,
					},
					Active:   true,
					Score:    []float64{9.7, 9.8, 9.9},
					Location: complex(6, 8),
					Meta:     "0x2",
				},
				err: false,
			},
		},
		`malformed time`: {
			inp: `name:Dasha|birth:11.06.200 23_59_59.999|age:18|active:true|score:9.7,9.8,9.9|pos:(6+8i)|meta:0x2`,
			exp: struct {
				res player
				err bool
			}{
				err: true,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var p player

			err := Unmarshal([]byte(test.inp), &p)

			if !test.exp.err {
				assert.Nil(t, err)
				assert.Equal(t, &test.exp.res, &p)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	tests := map[string]struct {
		inp player
		exp struct {
			res string
			err bool
		}
	}{
		`complete`: {
			inp: player{
				person: person{
					Name:      "Dasha",
					BirthDate: Time{time.UnixMilli(1118509199999)},
					Age:       18,
				},
				Active:   true,
				Score:    []float64{9.7, 9.8, 9.9},
				Location: complex(6, 8),
				Meta:     "0x2",
			},
			exp: struct {
				res string
				err bool
			}{
				res: `name:Dasha|birth:11.06.2005 23_59_59.999|age:18|active:true|score:9.7,9.8,9.9|loc:(6+8i)|meta:0x2`,
				err: false,
			},
		},
		`field missed`: {
			inp: player{
				person: person{
					Name:      "Dasha",
					BirthDate: Time{time.UnixMilli(1118509199999)},
					Age:       18,
				},
				Active:   true,
				Score:    []float64{9.7, 9.8, 9.9},
				Location: complex(6, 8),
				Meta:     nil,
			},
			exp: struct {
				res string
				err bool
			}{
				res: `name:Dasha|birth:11.06.2005 23_59_59.999|age:18|active:true|score:9.7,9.8,9.9|loc:(6+8i)|meta:none`,
				err: false,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := Marshal(&test.inp)

			if !test.exp.err {
				assert.Nil(t, err)
				assert.Equal(t, test.exp.res, string(res))
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
