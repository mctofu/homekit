package client

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/brutella/hc/hap"
)

// AccessoryConnectionConfig captures information needed to communicate with
// a paired accessory.
type AccessoryConnectionConfig struct {
	DeviceID         string
	PublicKey        []byte
	IPConnectionInfo IPConnectionInfo
}

// ControllerIdentity captures required identifying details for a controller.
type ControllerIdentity struct {
	DeviceID   string
	PublicKey  []byte
	PrivateKey []byte
}

// AccessoryClient allows interaction with a HomeKit accessory
type AccessoryClient struct {
	transport        IPTransport
	ipConnectionInfo IPConnectionInfo
	closeFn          func() error
}

// NewAccessoryClient returns a new AccessoryClient using IP transport. The client uses the
// provided ControllerIdentity to connect to the accessory specified in AccessoryConnectionConfig.
//
// Before using AccessoryClient you should first pair with the accessory using SetupClient.
//
// This client is not thread safe.
func NewAccessoryClient(dialer IPDialer, c *ControllerIdentity, a *AccessoryConnectionConfig) *AccessoryClient {
	homekitDialer := NewHomeKitSecureDialer(dialer, c, a)

	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
			DialContext:        homekitDialer.Dial,
		},
	}

	return &AccessoryClient{
		transport:        httpClient,
		ipConnectionInfo: a.IPConnectionInfo,
		closeFn: func() error {
			httpClient.CloseIdleConnections()
			return homekitDialer.Close()
		},
	}
}

func (a *AccessoryClient) endpoint(name string) string {
	return fmt.Sprintf("http://%s:%d/%s", a.ipConnectionInfo.IPAddress, a.ipConnectionInfo.Port, name)
}

func (a *AccessoryClient) sendTLV8(ctx context.Context, endpoint string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", hap.HTTPContentTypePairingTLV8)

	resp, err := a.transport.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code %v", resp.StatusCode)
	}

	return respBody, nil
}

// Close releases any resources used by the client.
func (a *AccessoryClient) Close() error {
	if a.closeFn != nil {
		return a.closeFn()
	}
	return nil
}
