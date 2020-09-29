package client

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/brutella/hc/crypto"
	"github.com/brutella/hc/db"
	"github.com/brutella/hc/hap"
	"github.com/brutella/hc/hap/pair"
	"github.com/mctofu/homekit/client/pairing"
)

// VerifyClient sets up a secure session between a controller and accessory. This authenticates
// that the controller and accessory are trusted and negotiates a shared secret to encrypt
// further communciation.
type VerifyClient struct {
	ipTransport IPTransport
	controller  *pairing.VerifyClientController
	endpoint    string
}

// NewVerifyClient returns a new VerifyClient suitable for performing pair-verify between
// accessory a and controller c.
func NewVerifyClient(t IPTransport, a *AccessoryConnectionConfig, c *ControllerIdentity) *VerifyClient {
	var deviceDB memoryDB
	deviceDB.MustSaveEntity(db.NewEntity(a.DeviceID, a.PublicKey, []byte{}))
	deviceDB.MustSaveEntity(db.NewEntity(c.DeviceID, c.PublicKey, c.PrivateKey))

	clientDevice, err := hap.NewDevice(c.DeviceID, &deviceDB)
	if err != nil {
		// this shouldn't happen
		panic(err)
	}

	return &VerifyClient{
		ipTransport: t,
		controller:  pairing.NewVerifyClientController(clientDevice, &deviceDB),
		endpoint:    fmt.Sprintf("http://%s:%d/pair-verify", a.IPConnectionInfo.IPAddress, a.IPConnectionInfo.Port),
	}
}

// Verify performs the pair-verify authentication and shared secret exchange and
// returns a crypto.Cryptographer that can encrypt future communication with the
// accessory.
func (v *VerifyClient) Verify(ctx context.Context) (crypto.Cryptographer, error) {
	verifyStartReq, err := ioutil.ReadAll(v.controller.InitialKeyVerifyRequest())
	if err != nil {
		return nil, err
	}

	verifyStartResp, err := v.sendTLV8(ctx, verifyStartReq)
	if err != nil {
		return nil, fmt.Errorf("verifyStartRequest: %v", err)
	}

	verifyFinishReqReader, err := pair.HandleReaderForHandler(bytes.NewReader(verifyStartResp), v.controller)
	if err != nil {
		return nil, fmt.Errorf("handle verifyStartResponse: %v", err)
	}
	verifyFinishReq, err := ioutil.ReadAll(verifyFinishReqReader)
	if err != nil {
		return nil, fmt.Errorf("read verifyFinishRequest: %v", err)
	}

	verifyFinishResp, err := v.sendTLV8(ctx, verifyFinishReq)
	if err != nil {
		return nil, fmt.Errorf("verifyFinishRequest: %v", err)
	}

	if _, err := pair.HandleReaderForHandler(bytes.NewReader(verifyFinishResp), v.controller); err != nil {
		return nil, fmt.Errorf("handle verifyFinishResponse: %v", err)
	}

	return v.controller.SessionCryptographer()
}

func (v *VerifyClient) sendTLV8(ctx context.Context, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", hap.HTTPContentTypePairingTLV8)

	resp, err := v.ipTransport.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code %v: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
