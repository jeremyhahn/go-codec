//go:build codec_msgpack

package msgpack

import (
	"bytes"
	"testing"
)

type BenchStruct struct {
	Name   string  `msgpack:"name"`
	Age    int     `msgpack:"age"`
	Email  string  `msgpack:"email"`
	Score  float64 `msgpack:"score"`
	Active bool    `msgpack:"active"`
}

var benchData = BenchStruct{
	Name:   "John Doe",
	Age:    30,
	Email:  "john@example.com",
	Score:  95.5,
	Active: true,
}

func BenchmarkCodec_Marshal(b *testing.B) {
	codec := New[BenchStruct]()
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
	codec := New[BenchStruct]()
	data, _ := codec.Marshal(benchData)
	var result BenchStruct

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
	codec := New[BenchStruct]()

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
	codec := New[BenchStruct]()
	var buf bytes.Buffer
	if err := codec.Encode(&buf, benchData); err != nil {
		b.Fatal(err)
	}
	data := buf.Bytes()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var result BenchStruct
		reader := bytes.NewReader(data)
		err := codec.Decode(reader, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCodec_MarshalLarge(b *testing.B) {
	type LargeStruct struct {
		Items []BenchStruct `msgpack:"items"`
	}

	items := make([]BenchStruct, 100)
	for i := range items {
		items[i] = benchData
	}

	codec := New[LargeStruct]()
	data := LargeStruct{Items: items}

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
	type LargeStruct struct {
		Items []BenchStruct `msgpack:"items"`
	}

	items := make([]BenchStruct, 100)
	for i := range items {
		items[i] = benchData
	}

	codec := New[LargeStruct]()
	data := LargeStruct{Items: items}
	marshaled, _ := codec.Marshal(data)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var result LargeStruct
		err := codec.Unmarshal(marshaled, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}
