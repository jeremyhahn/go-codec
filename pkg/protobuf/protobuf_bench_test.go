package protobuf

import (
	"bytes"
	"testing"

	"github.com/jeremyhahn/go-codec/pkg/protobuf/testdata"
)

var benchData = &testdata.TestMessage{
	Name:  "John Doe",
	Age:   30,
	Email: "john@example.com",
}

func BenchmarkCodec_Marshal(b *testing.B) {
	codec := New[*testdata.TestMessage]()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := codec.Marshal(benchData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCodec_Unmarshal(b *testing.B) {
	codec := New[*testdata.TestMessage]()
	data, _ := codec.Marshal(benchData)
	result := &testdata.TestMessage{}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := codec.Unmarshal(data, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCodec_Encode(b *testing.B) {
	codec := New[*testdata.TestMessage]()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		err := codec.Encode(&buf, benchData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCodec_Decode(b *testing.B) {
	codec := New[*testdata.TestMessage]()
	var buf bytes.Buffer
	if err := codec.Encode(&buf, benchData); err != nil {
		b.Fatal(err)
	}
	data := buf.Bytes()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := &testdata.TestMessage{}
		reader := bytes.NewReader(data)
		err := codec.Decode(reader, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCodec_MarshalLarge(b *testing.B) {
	items := make([]*testdata.TestMessage, 100)
	for i := range items {
		items[i] = benchData
	}

	codec := New[*testdata.StringList]()
	data := &testdata.StringList{Values: make([]string, 100)}
	for i := range data.Values {
		data.Values[i] = "test string value"
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := codec.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCodec_UnmarshalLarge(b *testing.B) {
	codec := New[*testdata.StringList]()
	data := &testdata.StringList{Values: make([]string, 100)}
	for i := range data.Values {
		data.Values[i] = "test string value"
	}
	marshaled, _ := codec.Marshal(data)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := &testdata.StringList{}
		err := codec.Unmarshal(marshaled, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}
