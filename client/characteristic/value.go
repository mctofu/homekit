package characteristic

import (
	"encoding/json"
	"errors"
)

// Value is a clone of json.RawMessage that lets us defer unmarshalling
// of a value until we determine the expected data type. Convenience methods
// are available to unmarshal to a known type.
type Value []byte

// MarshalJSON returns v as the JSON encoding of v.
func (v Value) MarshalJSON() ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	return v, nil
}

// UnmarshalJSON sets *v to a copy of data.
func (v *Value) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("client.Value: UnmarshalJSON on nil pointer")
	}
	*v = append((*v)[0:0], data...)
	return nil
}

// String attempts to unmarshal v to a string
func (v Value) String() (string, error) {
	if v == nil {
		return "", nil
	}

	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return "", err
	}
	return s, nil
}

// MustString attempts to unmarshal v to a string and panics if unsuccessful
func (v Value) MustString() string {
	s, err := v.String()
	if err != nil {
		panic(err)
	}
	return s
}

// Byte attempts to unmarshal v to a byte
func (v Value) Byte() (byte, error) {
	if v == nil {
		return 0, nil
	}

	var b byte
	if err := json.Unmarshal(v, &b); err != nil {
		return 0, err
	}
	return b, nil
}

// MustByte attempts to unmarshal v to a byte and panics if unsuccessful
func (v Value) MustByte() byte {
	b, err := v.Byte()
	if err != nil {
		panic(err)
	}
	return b
}

// Uint16 attempts to unmarshal v to a uint16
func (v Value) Uint16() (uint16, error) {
	if v == nil {
		return 0, nil
	}

	var u uint16
	if err := json.Unmarshal(v, &u); err != nil {
		return 0, err
	}
	return u, nil
}

// MustUint16 attempts to unmarshal v to a uint16 and panics if unsuccessful
func (v Value) MustUint16() uint16 {
	u, err := v.Uint16()
	if err != nil {
		panic(err)
	}
	return u
}

// Uint32 attempts to unmarshal v to a uint32
func (v Value) Uint32() (uint32, error) {
	if v == nil {
		return 0, nil
	}

	var u uint32
	if err := json.Unmarshal(v, &u); err != nil {
		return 0, err
	}
	return u, nil
}

// MustUint32 attempts to unmarshal v to a uint32 and panics if unsuccessful
func (v Value) MustUint32() uint32 {
	u, err := v.Uint32()
	if err != nil {
		panic(err)
	}
	return u
}

// Uint64 attempts to unmarshal v to a uint64
func (v Value) Uint64() (uint64, error) {
	if v == nil {
		return 0, nil
	}

	var u uint64
	if err := json.Unmarshal(v, &u); err != nil {
		return 0, err
	}
	return u, nil
}

// MustUint64 attempts to unmarshal v to a uint64 and panics if unsuccessful
func (v Value) MustUint64() uint64 {
	u, err := v.Uint64()
	if err != nil {
		panic(err)
	}
	return u
}

// Int32 attempts to unmarshal v to a int32
func (v Value) Int32() (int32, error) {
	if v == nil {
		return 0, nil
	}

	var i int32
	if err := json.Unmarshal(v, &i); err != nil {
		return 0, err
	}
	return i, nil
}

// MustInt32 attempts to unmarshal v to a int32 and panics if unsuccessful
func (v Value) MustInt32() int32 {
	i, err := v.Int32()
	if err != nil {
		panic(err)
	}
	return i
}

// Float64 attempts to unmarshal v to a float64
func (v Value) Float64() (float64, error) {
	if v == nil {
		return 0, nil
	}

	var f float64
	if err := json.Unmarshal(v, &f); err != nil {
		return 0, err
	}
	return f, nil
}

// MustFloat64 attempts to unmarshal v to a float64 and panics if unsuccessful
func (v Value) MustFloat64() float64 {
	f, err := v.Float64()
	if err != nil {
		panic(err)
	}
	return f
}

// Bool attempts to unmarshal v to a bool
func (v Value) Bool() (bool, error) {
	if v == nil {
		return false, nil
	}

	var b bool
	if err := json.Unmarshal(v, &b); err != nil {
		return false, err
	}
	return b, nil
}

// MustBool attempts to unmarshal v to a bool and panics if unsuccessful
func (v Value) MustBool() bool {
	b, err := v.Bool()
	if err != nil {
		panic(err)
	}
	return b
}

// Bytes attempts to unmarshal v to a byte slice
func (v Value) Bytes() ([]byte, error) {
	if v == nil {
		return nil, nil
	}

	var b []byte
	if err := json.Unmarshal(v, &b); err != nil {
		return nil, err
	}
	return b, nil
}

// MustBytes attempts to unmarshal v to a byte slice and panics if unsuccessful
func (v Value) MustBytes() []byte {
	b, err := v.Bytes()
	if err != nil {
		panic(err)
	}
	return b
}
