package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/mctofu/homekit/client/characteristic"
)

// CharacteristicsReadRequest identifies a list of characteristics to read
// as well as options to control what fields are returned for each characteristic.
type CharacteristicsReadRequest struct {
	Characteristics []CharacteristicReadRequest
	Metadata        bool
	Permissions     bool
	Type            bool
	Events          bool
}

// CharacteristicReadRequest identifies a single characteristic to be read
type CharacteristicReadRequest struct {
	AccessoryID      uint64
	CharacteristicID uint64
}

// CharacteristicReadResponse is the result of reading a single characteristic. Some
// fields are optional depending on the accessory implementation and the options
// specified in the CharacteristicsReadRequest.
type CharacteristicReadResponse struct {
	AccessoryID      uint64               `json:"aid"`
	CharacteristicID uint64               `json:"iid"`
	Value            characteristic.Value `json:"value"`

	Type        *string  `json:"type,omitempty"`
	Status      *int     `json:"status,omitempty"`
	Events      *int     `json:"ev,omitempty"`
	Permissions []string `json:"perms,omitempty"`

	Format *string `json:"format,omitempty"`
	Unit   *string `json:"unit,omitempty"`

	MaxLen    *int                 `json:"maxLen,omitempty"`
	MaxValue  characteristic.Value `json:"maxValue,omitempty"`
	MinValue  characteristic.Value `json:"minValue,omitempty"`
	StepValue characteristic.Value `json:"minStep,omitempty"`
}

// Characteristics returns the values of the characteristics specified in readReq.
func (a *AccessoryClient) Characteristics(
	ctx context.Context,
	readReq *CharacteristicsReadRequest,
) ([]*CharacteristicReadResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.endpointCharacteristics(), nil)
	if err != nil {
		return nil, err
	}

	// velux homekit accessory doesn't seem to like url encoded commas so
	// we need to use raw commas.
	// TODO: confirm if other devices behave similarly
	var query strings.Builder
	query.WriteString(fmt.Sprintf("id=%s", encodeIDs(readReq.Characteristics)))
	if readReq.Metadata {
		query.WriteString("&meta=1")
	}
	if readReq.Permissions {
		query.WriteString("&perms=1")
	}
	if readReq.Type {
		query.WriteString("&type=1")
	}
	if readReq.Events {
		query.WriteString("&ev=1")
	}
	req.URL.RawQuery = query.String()

	resp, err := a.transport.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusMultiStatus {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	respData := struct {
		Characteristics []*CharacteristicReadResponse `json:"characteristics"`
	}{}
	if err := json.Unmarshal(body, &respData); err != nil {
		return nil, fmt.Errorf("unmarshal: %v", err)
	}

	return respData.Characteristics, nil
}

// CharacteristicsWriteRequest specifies multiple characteristics to write values to.
type CharacteristicsWriteRequest struct {
	Characteristics []CharacteristicWriteRequest `json:"characteristics"`
	PrepareID       *uint64                      `json:"pid,omitempty"`
}

// buildDefaultResponse creates stub responses for each characteristic when the accessory
// does not return its own response.
func (c *CharacteristicsWriteRequest) buildDefaultResponse() []*CharacteristicWriteResponse {
	result := make([]*CharacteristicWriteResponse, 0, len(c.Characteristics))
	for _, req := range c.Characteristics {
		result = append(result, &CharacteristicWriteResponse{
			AccessoryID:      req.AccessoryID,
			CharacteristicID: req.CharacteristicID,
		})
	}
	return result
}

// CharacteristicWriteRequest specifies a single characteristic to write a value to
// as well as other optional settings.
type CharacteristicWriteRequest struct {
	AccessoryID      uint64      `json:"aid"`
	CharacteristicID uint64      `json:"iid"`
	Value            interface{} `json:"value,omitempty"`
	Events           bool        `json:"ev,omitempty"`
	AuthData         string      `json:"authdata,omitempty"`
	Remote           bool        `json:"remote,omitempty"`
	Response         bool        `json:"r,omitempty"`
}

// CharacteristicWriteResponse is the result of a write to a single characteristic.
type CharacteristicWriteResponse struct {
	AccessoryID      uint64               `json:"aid"`
	CharacteristicID uint64               `json:"iid"`
	Status           *int                 `json:"status,omitempty"`
	Value            characteristic.Value `json:"value,omitempty"`
}

// SetCharacteristics updates values and settings of characteristics contained in writeReq.
func (a *AccessoryClient) SetCharacteristics(ctx context.Context, writeReq *CharacteristicsWriteRequest) ([]*CharacteristicWriteResponse, error) {
	reqBody, err := json.Marshal(writeReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, a.endpointCharacteristics(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	resp, err := a.transport.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusNoContent:
		return writeReq.buildDefaultResponse(), nil
	case http.StatusOK, http.StatusMultiStatus:
		respData := struct {
			Characteristics []*CharacteristicWriteResponse `json:"characteristics"`
		}{}
		if err := json.Unmarshal(body, &respData); err != nil {
			return nil, fmt.Errorf("unmarshal: %v", err)
		}
		return respData.Characteristics, nil
	default:
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}
}

func encodeIDs(ids []CharacteristicReadRequest) string {
	stringIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		stringIDs = append(stringIDs, fmt.Sprintf("%d.%d", id.AccessoryID, id.CharacteristicID))
	}

	return strings.Join(stringIDs, ",")
}

func (a *AccessoryClient) endpointCharacteristics() string {
	return a.endpoint("characteristics")
}
