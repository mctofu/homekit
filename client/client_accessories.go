package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mctofu/homekit/client/service"
)

// RawAccessory captures information related to an accessory of a
// HomeKit accessory.
type RawAccessory struct {
	ID       uint64                `json:"aid"`
	Services []*service.RawService `json:"services"`
}

// Info returns the characteristic data from the accessory's info service.Info
// This service is required by the spec to be present for all accessories but
// if it's not found then nil is returned.
func (r *RawAccessory) Info() *service.AccessoryInfo {
	infoSvc := r.ServiceByType(service.TypeAccessoryInformation)
	if infoSvc == nil {
		return nil
	}

	return service.ReadAccessoryInfo(infoSvc.Characteristics)
}

// ServiceByType returns the service of the accessory with a matching type if
// present. Nil otherwise.
func (r *RawAccessory) ServiceByType(t string) *service.RawService {
	for _, svc := range r.Services {
		if svc.Type == t {
			return svc
		}
	}

	return nil
}

// Accessories queries the HomeKit accessory to retrieve its Accessory Attribute Database. This
// returns all hap accessories, services and characteristics available.
func (a *AccessoryClient) Accessories(ctx context.Context) ([]*RawAccessory, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.endpoint("accessories"), nil)
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
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	respData := struct {
		Accessories []*RawAccessory `json:"accessories"`
	}{}
	if err := json.Unmarshal(body, &respData); err != nil {
		return nil, fmt.Errorf("unmarshal: %v\n%s", err, string(body))
	}

	return respData.Accessories, nil
}
