package client

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/brutella/hc/db"
	"github.com/brutella/hc/hap"
	"github.com/brutella/hc/hap/pair"
)

// AccessoryPairingConfig contains accessory details needed to perform a pairing.
type AccessoryPairingConfig struct {
	DeviceID         string
	PIN              string
	IPConnectionInfo IPConnectionInfo
}

// SetupClient negotiates an initial pairing between a controller and an accessory.
type SetupClient struct {
	ipTransport IPTransport
}

// NewSetupClient returns a new SetupClient for ip accessible accessories.
func NewSetupClient(t IPTransport) *SetupClient {
	return &SetupClient{
		ipTransport: t,
	}
}

// Pair will pair the controller c with the accessory a.
func (s *SetupClient) Pair(ctx context.Context, a *AccessoryPairingConfig, c *ControllerIdentity) (*AccessoryConnectionConfig, error) {
	deviceDB := &memoryDB{}
	if err := deviceDB.SaveEntity(db.NewEntity(c.DeviceID, c.PublicKey, c.PrivateKey)); err != nil {
		return nil, fmt.Errorf("SaveEntity: %v", err)
	}

	clientDevice, err := hap.NewDevice(c.DeviceID, deviceDB)
	if err != nil {
		return nil, err
	}

	controller := pair.NewSetupClientController(a.PIN, clientDevice, deviceDB)
	endpoint := fmt.Sprintf("http://%s:%d/pair-setup", a.IPConnectionInfo.IPAddress, a.IPConnectionInfo.Port)

	pairStartReq, err := ioutil.ReadAll(controller.InitialPairingRequest())
	if err != nil {
		return nil, err
	}

	pairStartResp, err := s.sendTLV8(ctx, endpoint, pairStartReq)
	if err != nil {
		return nil, fmt.Errorf("pairStartRequest: %v", err)
	}

	pairVerifyReqReader, err := pair.HandleReaderForHandler(bytes.NewReader(pairStartResp), controller)
	if err != nil {
		return nil, fmt.Errorf("handle pairStartResponse: %v", err)
	}
	pairVerifyReq, err := ioutil.ReadAll(pairVerifyReqReader)
	if err != nil {
		return nil, err
	}

	pairVerifyResp, err := s.sendTLV8(ctx, endpoint, pairVerifyReq)
	if err != nil {
		return nil, fmt.Errorf("pairVerifyRequest: %v", err)
	}

	pairKeyReqReader, err := pair.HandleReaderForHandler(bytes.NewReader(pairVerifyResp), controller)
	if err != nil {
		return nil, fmt.Errorf("handle pairVerifyResponse: %v", err)
	}
	pairKeyReq, err := ioutil.ReadAll(pairKeyReqReader)
	if err != nil {
		return nil, err
	}

	pairKeyResp, err := s.sendTLV8(ctx, endpoint, pairKeyReq)
	if err != nil {
		return nil, fmt.Errorf("pairKeyRequest: %v", err)
	}

	if _, err := pair.HandleReaderForHandler(bytes.NewReader(pairKeyResp), controller); err != nil {
		return nil, fmt.Errorf("handle pairKeyResponse: %v", err)
	}

	// accessory's public key should now be in the db
	entity, err := deviceDB.EntityWithName(a.DeviceID)
	if err != nil {
		return nil, err
	}

	return &AccessoryConnectionConfig{
		PublicKey:        entity.PublicKey,
		DeviceID:         a.DeviceID,
		IPConnectionInfo: a.IPConnectionInfo,
	}, nil
}

func (s *SetupClient) sendTLV8(ctx context.Context, endpoint string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", hap.HTTPContentTypePairingTLV8)

	resp, err := s.ipTransport.Do(req)
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
