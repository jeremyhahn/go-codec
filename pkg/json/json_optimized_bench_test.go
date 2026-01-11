//go:build codec_json

package json

import (
	"testing"
)

// Optimized benchmarks using the new zero-allocation APIs

func BenchmarkOptimizedCodec_MarshalTo(b *testing.B) {
	codec := NewPool[BenchStruct]()
	buf := make([]byte, 0, 256)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := codec.MarshalTo(buf, benchData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOptimizedCodec_AppendMarshal(b *testing.B) {
	codec := NewPool[BenchStruct]()
	buf := make([]byte, 0, 256)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf = buf[:0] // Reset length
		_, err := codec.AppendMarshal(buf, benchData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOptimizedCodec_UnmarshalFrom(b *testing.B) {
	codec := NewPool[BenchStruct]()
	data, _ := codec.Marshal(benchData)
	var result BenchStruct
	scratch := make([]byte, 256)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := codec.UnmarshalFrom(data, &result, scratch)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOptimizedCodec_MarshalLarge(b *testing.B) {
	type LargeStruct struct {
		Items []BenchStruct `json:"items"`
	}

	items := make([]BenchStruct, 100)
	for i := range items {
		items[i] = benchData
	}

	codec := NewPool[LargeStruct]()
	data := LargeStruct{Items: items}
	buf := make([]byte, 0, 16*1024)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := codec.MarshalTo(buf, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOptimizedCodec_UnmarshalLarge(b *testing.B) {
	type LargeStruct struct {
		Items []BenchStruct `json:"items"`
	}

	items := make([]BenchStruct, 100)
	for i := range items {
		items[i] = benchData
	}

	codec := NewPool[LargeStruct]()
	data := LargeStruct{Items: items}
	marshaled, _ := codec.Marshal(data)
	scratch := make([]byte, 16*1024)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var result LargeStruct
		err := codec.UnmarshalFrom(marshaled, &result, scratch)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Comparison benchmarks: reuse buffer vs allocate each time

func BenchmarkComparison_ReuseBuffer(b *testing.B) {
	codec := NewPool[BenchStruct]()
	buf := make([]byte, 0, 256)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf = buf[:0]
		_, err := codec.MarshalTo(buf, benchData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkComparison_AllocateEachTime(b *testing.B) {
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
