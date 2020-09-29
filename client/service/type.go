package service

//go:generate go run ../../cmd/gen/main.go -services

import "github.com/mctofu/homekit/client/characteristic"

// RawService captures information related to a service of an accessory.
type RawService struct {
	ID              uint64                              `json:"iid"`
	Type            string                              `json:"type"`
	Characteristics []*characteristic.RawCharacteristic `json:"characteristics"`
	Hidden          *bool                               `json:"hidden,omitempty"`
	Primary         *bool                               `json:"primary,omitempty"`
	Linked          []uint64                            `json:"linked,omitempty"`
}

// CharacteristicByType returns the characteristic of the accessory with a matching type
// if present. Nil otherwise.
func (s *RawService) CharacteristicByType(t string) *characteristic.RawCharacteristic {
	for _, c := range s.Characteristics {
		if c.Type == t {
			return c
		}
	}

	return nil
}

// TypeMetadata captures standard known metadata that applies to all
// services with the given type.
type TypeMetadata struct {
	Type string
	Name string
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
