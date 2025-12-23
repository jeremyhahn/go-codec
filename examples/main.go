package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/jeremyhahn/go-codec"
	"github.com/jeremyhahn/go-codec/pkg/factory"
)

// Person demonstrates struct tags for all supported codecs
type Person struct {
	Name  string `json:"name" yaml:"name" toml:"name" msgpack:"name" bson:"name" cbor:"name" avro:"name"`
	Age   int    `json:"age" yaml:"age" toml:"age" msgpack:"age" bson:"age" cbor:"age" avro:"age"`
	Email string `json:"email" yaml:"email" toml:"email" msgpack:"email" bson:"email" cbor:"email" avro:"email"`
}

func main() {
	person := Person{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	fmt.Println("=== go-codec Demo ===")
	fmt.Println()
	fmt.Println("Demonstrating all supported serialization formats.")
	fmt.Println()

	// Text formats
	fmt.Println("--- JSON ---")
	demoCodec(codec.JSON, person)

	fmt.Println("\n--- YAML ---")
	demoCodec(codec.YAML, person)

	fmt.Println("\n--- TOML ---")
	demoCodec(codec.TOML, person)

	// Binary formats
	fmt.Println("\n--- MessagePack ---")
	demoBinaryCodec(codec.MsgPack, person)

	fmt.Println("\n--- BSON ---")
	demoBinaryCodec(codec.BSON, person)

	fmt.Println("\n--- CBOR ---")
	demoBinaryCodec(codec.CBOR, person)

	fmt.Println("\n--- Avro ---")
	demoBinaryCodec(codec.Avro, person)

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("See examples/protobuf/ for Protocol Buffers example.")
}

// demoCodec demonstrates a text-based codec (output is printable)
func demoCodec(codecType codec.Type, person Person) {
	c, err := factory.New[Person](codecType)
	if err != nil {
		log.Fatalf("Failed to create %s codec: %v", codecType, err)
	}

	// Marshal to bytes
	data, err := c.Marshal(person)
	if err != nil {
		log.Fatalf("Failed to marshal: %v", err)
	}

	fmt.Printf("Encoded (%d bytes):\n%s\n", len(data), string(data))

	// Unmarshal from bytes
	var decoded Person
	err = c.Unmarshal(data, &decoded)
	if err != nil {
		log.Fatalf("Failed to unmarshal: %v", err)
	}

	fmt.Printf("Decoded: %+v\n", decoded)

	// Stream encoding/decoding
	var buf bytes.Buffer
	err = c.Encode(&buf, person)
	if err != nil {
		log.Fatalf("Failed to encode: %v", err)
	}

	var streamDecoded Person
	err = c.Decode(&buf, &streamDecoded)
	if err != nil {
		log.Fatalf("Failed to decode: %v", err)
	}

	fmt.Printf("Stream decoded: %+v\n", streamDecoded)
}

// demoBinaryCodec demonstrates a binary codec (output shown as hex)
func demoBinaryCodec(codecType codec.Type, person Person) {
	c, err := factory.New[Person](codecType)
	if err != nil {
		log.Fatalf("Failed to create %s codec: %v", codecType, err)
	}

	// Marshal to bytes
	data, err := c.Marshal(person)
	if err != nil {
		log.Fatalf("Failed to marshal: %v", err)
	}

	fmt.Printf("Encoded (%d bytes): %x\n", len(data), data)

	// Unmarshal from bytes
	var decoded Person
	err = c.Unmarshal(data, &decoded)
	if err != nil {
		log.Fatalf("Failed to unmarshal: %v", err)
	}

	fmt.Printf("Decoded: %+v\n", decoded)

	// Stream encoding/decoding
	var buf bytes.Buffer
	err = c.Encode(&buf, person)
	if err != nil {
		log.Fatalf("Failed to encode: %v", err)
	}

	var streamDecoded Person
	err = c.Decode(&buf, &streamDecoded)
	if err != nil {
		log.Fatalf("Failed to decode: %v", err)
	}

	fmt.Printf("Stream decoded: %+v\n", streamDecoded)
}
