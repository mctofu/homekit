package characteristic

//go:generate go run ../../cmd/gen/main.go -characteristics

import (
	"encoding/hex"
	"fmt"
	"strconv"
)

// RawCharacteristic captures information related to a particular characteristic
// of a service.
type RawCharacteristic struct {
	ID    uint64 `json:"iid"`
	Value Value  `json:"value"`

	Type        string   `json:"type"`
	Events      *bool    `json:"ev,omitempty"`
	Permissions []string `json:"perms,omitempty"`

	Format string `json:"format"`
	Unit   string `json:"unit"`

	MaxLen    *int  `json:"maxLen,omitempty"`
	MaxValue  Value `json:"maxValue,omitempty"`
	MinValue  Value `json:"minValue,omitempty"`
	StepValue Value `json:"minStep,omitempty"`
}

// UndefinedValue represents a value with an unknown format
type UndefinedValue struct{}

func (u UndefinedValue) String() string {
	return "<Undefined>"
}

// UnknownType represents a value from an unknown type
type UnknownType struct{}

func (u UnknownType) String() string {
	return "<Unknown Type>"
}

// TypeMetadata captures standard known metadata that applies to all
// characteristics with the given type.
type TypeMetadata struct {
	Type   string
	Name   string
	Format string
}

var typeMetadataByType map[string]*TypeMetadata

func init() {
	typeMetadataByType = make(map[string]*TypeMetadata)
	for _, tm := range typeMetadatas {
		typeMetadataByType[tm.Type] = tm
	}
}

// RegisterType allows registering a custom type
func RegisterType(tm TypeMetadata) {
	typeMetadataByType[tm.Type] = &tm
}

// NameForType returns a human friendly name for the type or "<Unknown>"
// if the type is not registered.
func NameForType(t string) string {
	tm := typeMetadataByType[t]
	if tm == nil {
		return "<Unknown>"
	}
	return tm.Name
}

// ValueForFormat will parse v to the type matching the format. The value
// must match the format or this method will panic. If the format is unknown
// then an instance of UndefinedValue is returned.
func ValueForFormat(f string, v Value) interface{} {
	switch f {
	case "bool":
		return v.MustBool()
	case "uint8":
		return v.MustByte()
	case "uint16":
		return v.MustUint16()
	case "uint32":
		return v.MustUint32()
	case "uint64":
		return v.MustUint64()
	case "int":
		return v.MustInt32()
	case "float":
		return v.MustFloat64()
	case "string":
		return v.MustString()
	case "tlv8", "data": // TODO: look at tlv8/data examples
		return v.MustBytes()
	default:
		return UndefinedValue{}
	}
}

// ValueForType will parse v based on the format registered in the type's
// metadata.
func ValueForType(t string, v Value) interface{} {
	tm := typeMetadataByType[t]
	if tm == nil {
		return UnknownType{}
	}
	return ValueForFormat(tm.Format, v)
}

// ParseValueForType parses a string value to go type defined by the type
// metadata's format.
func ParseValueForType(t string, v string) (interface{}, error) {
	tm := typeMetadataByType[t]
	if tm == nil {
		return nil, fmt.Errorf("unknown type: %s", t)
	}
	switch tm.Format {
	case "bool":
		return strconv.ParseBool(v)
	case "uint8":
		val, err := strconv.ParseUint(v, 10, 8)
		return byte(val), err
	case "uint16":
		val, err := strconv.ParseUint(v, 10, 16)
		return uint16(val), err
	case "uint32":
		val, err := strconv.ParseUint(v, 10, 32)
		return uint32(val), err
	case "uint64":
		return strconv.ParseUint(v, 10, 64)
	case "int":
		val, err := strconv.ParseInt(v, 10, 32)
		return int32(val), err
	case "float":
		return strconv.ParseFloat(v, 64)
	case "string":
		return v, nil
	case "tlv8", "data":
		return hex.DecodeString(v)
	default:
		return nil, fmt.Errorf("unhandled value type for: %s", t)
	}
}
