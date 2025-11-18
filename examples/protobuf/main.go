package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/jeremyhahn/go-codec"
)

func main() {
	person := &Person{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	fmt.Println("=== Protocol Buffers Codec Demo ===\n")

	// Create a Protocol Buffers codec
	// Note: Must use codec.NewProtoBuf instead of codec.New for protobuf types
	protobufCodec := codec.NewProtoBuf[*Person]()

	// Marshal to protobuf bytes
	data, err := protobufCodec.Marshal(person)
	if err != nil {
		log.Fatalf("Failed to marshal: %v", err)
	}

	fmt.Printf("Encoded (%d bytes): %v\n", len(data), data)

	// Unmarshal from protobuf bytes
	decoded := &Person{}
	err = protobufCodec.Unmarshal(data, &decoded)
	if err != nil {
		log.Fatalf("Failed to unmarshal: %v", err)
	}

	fmt.Printf("Decoded: Name=%s, Age=%d, Email=%s\n", decoded.Name, decoded.Age, decoded.Email)

	// Stream encoding/decoding
	var buf bytes.Buffer
	err = protobufCodec.Encode(&buf, person)
	if err != nil {
		log.Fatalf("Failed to encode: %v", err)
	}

	streamDecoded := &Person{}
	err = protobufCodec.Decode(&buf, &streamDecoded)
	if err != nil {
		log.Fatalf("Failed to decode: %v", err)
	}

	fmt.Printf("Stream decoded: Name=%s, Age=%d, Email=%s\n",
		streamDecoded.Name, streamDecoded.Age, streamDecoded.Email)
}
