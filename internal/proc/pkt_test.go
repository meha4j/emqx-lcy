package proc

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRecordMarshal(t *testing.T) {
	rec := Record{
		Value:     "11.06",
		Timestamp: Time{time.UnixMilli(1118509199999)},
	}

	res, err := json.Marshal(rec)

	assert.Nil(t, err)
	assert.Equal(t, `{"value":"11.06","timestamp":1118509199999}`, string(res))
}

func TestRecordMarshalValueEmpty(t *testing.T) {
	rec := Record{
		Timestamp: Time{time.UnixMilli(1118509199999)},
	}

	res, err := json.Marshal(rec)

	assert.Nil(t, err)
	assert.Equal(t, `{"timestamp":1118509199999}`, string(res))
}

func TestRecordUnmarshalValueString(t *testing.T) {
	var rec Record

	err := json.Unmarshal([]byte(`
	{
		"value": "11.06",
		"timestamp": 1118509199999
	}
	`), &rec)

	assert.Nil(t, err)
	assert.Equal(t, "11.06", rec.Value)
	assert.Equal(t, int64(1118509199999), rec.Timestamp.UnixMilli())
}

func TestRecordUnmarshalValueNumber(t *testing.T) {
	var rec Record

	err := json.Unmarshal([]byte(`
	{
		"value": 11.06,
		"timestamp": 1118509199999
	}
	`), &rec)

	assert.Nil(t, err)
	assert.Equal(t, 11.06, rec.Value)
	assert.Equal(t, int64(1118509199999), rec.Timestamp.UnixMilli())
}

func TestRecordUnmarshalValueEmpty(t *testing.T) {
	rec := Record{
		Value:     "11.06",
		Timestamp: Time{time.UnixMilli(1118509199999)},
	}

	err := json.Unmarshal([]byte(`
	{
		"timestamp": 1118509199999
	}
	`), &rec)

	assert.Nil(t, err)
	assert.Equal(t, "11.06", rec.Value)
	assert.Equal(t, int64(1118509199999), rec.Timestamp.UnixMilli())
}

func TestRecordUnmarshalTimestampEmpty(t *testing.T) {
	rec := Record{
		Timestamp: Time{time.UnixMilli(1118509199000)},
	}

	err := json.Unmarshal([]byte(`
	{
		"value": "11.06"
	}
	`), &rec)

	assert.Nil(t, err)
	assert.Equal(t, "11.06", rec.Value)
	assert.Equal(t, int64(1118509199000), rec.Timestamp.UnixMilli())
}

func TestRecordUnmarshalTimestampMalformedError(t *testing.T) {
	var rec Record

	err := json.Unmarshal([]byte(`
	{
		"value": 11.06,
		"timestamp": "11.06"
	}
	`), &rec)

	assert.NotNil(t, err)
}

func TestCommandDecode(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:s|v:11.06")))
	assert.Equal(t, "test", cmd.Topic)
	assert.Equal(t, PUB, cmd.Method)
	assert.Equal(t, "11.06", cmd.Value)
	assert.Equal(t, int64(1118509199999), cmd.Timestamp.UnixMilli())

	assert.Nil(t, cmd.Decode([]byte("time:11.06.2005 23_59_59.999|name:test|meth:s|val:11.06")))
	assert.Equal(t, "test", cmd.Topic)
	assert.Equal(t, PUB, cmd.Method)
	assert.Equal(t, "11.06", cmd.Value)
	assert.Equal(t, int64(1118509199999), cmd.Timestamp.UnixMilli())

	assert.Nil(t, cmd.Decode([]byte("time:11.06.2005 23_59_59.999|name:test|method:s|value:11.06")))
	assert.Equal(t, PUB, cmd.Method)
	assert.Equal(t, "11.06", cmd.Value)
}

func TestCommandDecodePub(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:set|v:11.06")))
	assert.Equal(t, PUB, cmd.Method)
}

func TestCommandDecodeSub(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:sb|v:11.06")))
	assert.Equal(t, SUB, cmd.Method)

	cmd.Method = 0

	assert.Nil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:subscr|v:11.06")))
	assert.Equal(t, SUB, cmd.Method)

	cmd.Method = 0

	assert.Nil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:subscribe|v:11.06")))
	assert.Equal(t, SUB, cmd.Method)
}

func TestCommandDecodeUsb(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:rel|v:11.06")))
	assert.Equal(t, USB, cmd.Method)

	cmd.Method = 0

	assert.Nil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:release|v:11.06")))
	assert.Equal(t, USB, cmd.Method)
}

func TestCommandDecodeGet(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:g|v:11.06")))
	assert.Equal(t, GET, cmd.Method)

	cmd.Method = 0

	assert.Nil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:gf|v:11.06")))
	assert.Equal(t, GET, cmd.Method)

	cmd.Method = 0

	assert.Nil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:get|v:11.06")))
	assert.Equal(t, GET, cmd.Method)

	cmd.Method = 0

	assert.Nil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:getfull|v:11.06")))
	assert.Equal(t, GET, cmd.Method)
}

func TestCommandDecodeMethodUnknownError(t *testing.T) {
	var cmd Command

	assert.NotNil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:unknown|v:11.06")))
}

func TestCommandDecodeMethodMissedError(t *testing.T) {
	var cmd Command

	assert.NotNil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|v:11.06")))
}

func TestCommandDecodeTopicMissedError(t *testing.T) {
	var cmd Command

	assert.NotNil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|m:set|v:11.06")))
}

func TestCommandDecodeTimeMissed(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("n:test|m:set|v:11.06")))
	assert.Less(t, time.Since(cmd.Timestamp.Time), 3*time.Second)
}

func TestCommandDecodeExtraToken(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("n:test|extra:token|m:set|v:11.06")))
}

func TestCommandDecodeMalformedToken(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("n:test|malformed|m:set|v:11.06")))
}

func TestCommandDecodeTimeMalformedError(t *testing.T) {
	var cmd Command

	assert.NotNil(t, cmd.Decode([]byte("n:test|m:set|time:11.06.2005 23:59:59.999")))
}

func BenchmarkCommandDecode(b *testing.B) {
	var cmd Command

	for range b.N {
		cmd.Decode([]byte("time:11.06.2005 23_59_59.999|name:test|method:getfull|value:11.06"))
	}
}

func TestCommandEncode(t *testing.T) {
	cmd := Command{
		Topic: "test",
		Record: Record{
			Value:     "11.06",
			Timestamp: Time{time.UnixMilli(1118509199999)},
		},
	}

	assert.Equal(t, "time:11.06.2005 23_59_59.999|name:test|descr:-|units:-|type:rw|val:11.06\n", string(cmd.Encode()))
}

func TestCommandEncodeValueEmpty(t *testing.T) {
	cmd := Command{
		Topic: "test",
		Record: Record{
			Timestamp: Time{time.UnixMilli(1118509199999)},
		},
	}

	assert.Equal(t, "time:11.06.2005 23_59_59.999|name:test|descr:-|units:-|type:rw|val:none\n", string(cmd.Encode()))
}

func BenchmarkCommandEncode(b *testing.B) {
	cmd := Command{
		Topic: "test",
		Record: Record{
			Value:     "11.06",
			Timestamp: Time{time.UnixMilli(1118509199999)},
		},
	}

	for range b.N {
		cmd.Encode()
	}
}
