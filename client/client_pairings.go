package client

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/brutella/hc/hap/pair"
	"github.com/brutella/hc/util"
)

const (
	tagSeparator      = 255
	pairingMethodList = 5
)

// ListPairingResponse describes a controller pairing returned from the ListPairings call.
type ListPairingResponse struct {
	ControllerID string
	PublicKey    []byte
	Admin        bool
}

// ListPairings queries for a list of controllers that have been paired with the accessory.
func (a *AccessoryClient) ListPairings(ctx context.Context) ([]*ListPairingResponse, error) {
	out := util.NewTLV8Container()
	out.SetByte(pair.TagSequence, 1)
	out.SetByte(pair.TagPairingMethod, pairingMethodList)

	resp, err := a.sendTLV8(ctx, a.endpointPairing(), out.BytesBuffer().Bytes())
	if err != nil {
		return nil, fmt.Errorf("list request: %v", err)
	}

	splits, err := splitTLV8(resp)
	if err != nil {
		return nil, fmt.Errorf("splitTLV8: %v", err)
	}

	var result []*ListPairingResponse

	for i, split := range splits {
		in, err := util.NewTLV8ContainerFromReader(bytes.NewReader(split))
		if err != nil {
			return nil, fmt.Errorf("parse tlv8 response: %v", err)
		}

		if i == 0 {
			if seq := in.GetByte(pair.TagSequence); seq != 2 {
				return nil, fmt.Errorf("unexpected response sequence: %d", seq)
			}

			if errCode := in.GetByte(pair.TagErrCode); errCode != 0 {
				return nil, fmt.Errorf("error code: %d", errCode)
			}
		}

		result = append(result, &ListPairingResponse{
			ControllerID: in.GetString(pair.TagUsername),
			PublicKey:    in.GetBytes(pair.TagPublicKey),
			Admin:        in.GetByte(pair.TagPermission) == 1,
		})
	}

	return result, nil
}

// splitTLV8 splits a TLV8 response where separator items are detected.
func splitTLV8(v []byte) ([][]byte, error) {
	var result [][]byte

	var tag, length uint8

	var start int
	r := bytes.NewReader(v)

	for {
		if err := binary.Read(r, binary.LittleEndian, &tag); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if tag == tagSeparator {
			end := len(v) - r.Len()
			result = append(result, v[start:end-1])
			start = end
			continue
		}

		if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
			return nil, err
		}

		if _, err := r.Seek(int64(length), io.SeekCurrent); err != nil {
			return nil, err
		}
	}

	result = append(result, v[start:])

	return result, nil
}

// AddPairingRequest specifies an additional controller that should be added to an
// accessory's pairings.
type AddPairingRequest struct {
	DeviceID    string
	PublicKey   []byte
	Permissions byte
}

// AddPairing adds access by an additional controller to the accessory.
func (a *AccessoryClient) AddPairing(ctx context.Context, req *AddPairingRequest) error {
	out := util.NewTLV8Container()
	out.SetByte(pair.TagSequence, 1)
	out.SetByte(pair.TagPairingMethod, pair.PairingMethodAdd.Byte())
	out.SetString(pair.TagUsername, req.DeviceID)
	out.SetBytes(pair.TagPublicKey, req.PublicKey)
	out.SetByte(pair.TagPermission, req.Permissions)

	resp, err := a.sendTLV8(ctx, a.endpointPairing(), out.BytesBuffer().Bytes())
	if err != nil {
		return fmt.Errorf("add pairing request: %v", err)
	}

	in, err := util.NewTLV8ContainerFromReader(bytes.NewReader(resp))
	if err != nil {
		return fmt.Errorf("parse tlv8 response: %v", err)
	}

	if seq := in.GetByte(pair.TagSequence); seq != 2 {
		return fmt.Errorf("unexpected response sequence: %d", seq)
	}

	if errCode := in.GetByte(pair.TagErrCode); errCode != 0 {
		return fmt.Errorf("error code: %d", errCode)
	}

	return nil
}

// RemovePairing removes a pairing for the specified controller
func (a *AccessoryClient) RemovePairing(ctx context.Context, controllerDeviceID string) error {
	out := util.NewTLV8Container()
	out.SetByte(pair.TagSequence, 1)
	out.SetByte(pair.TagPairingMethod, pair.PairingMethodDelete.Byte())
	out.SetString(pair.TagUsername, controllerDeviceID)

	resp, err := a.sendTLV8(ctx, a.endpointPairing(), out.BytesBuffer().Bytes())
	if err != nil {
		return fmt.Errorf("remove pairing request: %v", err)
	}

	in, err := util.NewTLV8ContainerFromReader(bytes.NewReader(resp))
	if err != nil {
		return fmt.Errorf("parse tlv8 response: %v", err)
	}

	if seq := in.GetByte(pair.TagSequence); seq != 2 {
		return fmt.Errorf("unexpected response sequence: %d", seq)
	}

	if errCode := in.GetByte(pair.TagErrCode); errCode != 0 {
		return fmt.Errorf("error code: %d", errCode)
	}

	return nil
}

func (a *AccessoryClient) endpointPairing() string {
	return a.endpoint("pairings")
}
