package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/jeremyhahn/go-codec"
)

type Person struct {
	Name  string `json:"name" yaml:"name" toml:"name" msgpack:"name"`
	Age   int    `json:"age" yaml:"age" toml:"age" msgpack:"age"`
	Email string `json:"email" yaml:"email" toml:"email" msgpack:"email"`
}

func main() {
	person := Person{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	fmt.Println("=== Codec Demo ===")
	fmt.Println()

	// JSON
	fmt.Println("--- JSON ---")
	demoCodec(codec.JSON, person)

	// YAML
	fmt.Println("\n--- YAML ---")
	demoCodec(codec.YAML, person)

	// TOML
	fmt.Println("\n--- TOML ---")
	demoCodec(codec.TOML, person)

	// MessagePack
	fmt.Println("\n--- MessagePack ---")
	demoCodec(codec.MsgPack, person)
}

func demoCodec(codecType codec.Type, person Person) {
	// Create codec
	c, err := codec.New[Person](codecType)
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
