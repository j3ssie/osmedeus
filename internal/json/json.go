// Package json provides a high-performance drop-in replacement for encoding/json.
// Uses json-iterator for 2-6x faster JSON operations while maintaining full compatibility.
package json

import (
	"io"

	jsoniter "github.com/json-iterator/go"
)

// ConfigCompatibleWithStandardLibrary is configured to be 100% compatible with standard library.
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Marshal returns the JSON encoding of v.
func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// MarshalIndent is like Marshal but applies Indent to format the output.
func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v.
func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// MarshalToString returns the JSON encoding of v as a string.
func MarshalToString(v any) (string, error) {
	return json.MarshalToString(v)
}

// UnmarshalFromString parses the JSON-encoded string and stores the result in the value pointed to by v.
func UnmarshalFromString(str string, v any) error {
	return json.UnmarshalFromString(str, v)
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *jsoniter.Encoder {
	return json.NewEncoder(w)
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *jsoniter.Decoder {
	return json.NewDecoder(r)
}

// Valid reports whether data is a valid JSON encoding.
func Valid(data []byte) bool {
	return json.Valid(data)
}

// RawMessage is a raw encoded JSON value.
// It implements Marshaler and Unmarshaler and can be used to delay JSON decoding or precompute a JSON encoding.
type RawMessage = jsoniter.RawMessage
