package client

import (
	"context"
	"fmt"
	"net"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/crypto"
	"github.com/brutella/hc/event"
	"github.com/brutella/hc/hap"
	hchttp "github.com/brutella/hc/hap/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupClient(t *testing.T) {
	testServer, err := deviceServer()
	require.NoError(t, err, "deviceServer")
	defer testServer.Close()

	controller, err := NewRandomControllerConfig()
	require.NoError(t, err, "controller setup")

	ctx := context.Background()
	connectionConfig, err := setupDeviceServer(ctx, testServer, controller)
	require.NoError(t, err, "pair")

	assert.Equal(t, "5F-7A-CA-6A-83-92", connectionConfig.DeviceID)
}

func deviceServer() (*httptest.Server, error) {
	switchAcc := accessory.NewSwitch(
		accessory.Info{
			Name: "Test",
		},
	)

	container := accessory.NewContainer()
	if err := container.AddAccessory(switchAcc.Accessory); err != nil {
		return nil, err
	}

	serverDB := &memoryDB{}

	switchDev, err := hap.NewSecuredDevice("5F-7A-CA-6A-83-92", "12344321", serverDB)
	if err != nil {
		return nil, fmt.Errorf("NewSecuredDevice: %v", err)
	}

	switchCtx := &wrapperContext{hap.NewContextForSecuredDevice(switchDev)}

	hcServer := hchttp.NewServer(hchttp.Config{
		Context:   switchCtx,
		Container: container,
		Device:    switchDev,
		Database:  serverDB,
		Mutex:     &sync.Mutex{},
		Emitter:   event.NewEmitter(),
	})

	testServer := httptest.NewUnstartedServer(hcServer.Mux)
	testServer.Listener = hcServer
	testServer.Start()

	return testServer, nil
}

func setupDeviceServer(ctx context.Context, testServer *httptest.Server, controller *ControllerIdentity) (*AccessoryConnectionConfig, error) {
	setupClient := NewSetupClient(testServer.Client())
	baseURL, err := url.Parse(testServer.URL)
	if err != nil {
		return nil, fmt.Errorf("parseURL: %v", err)
	}
	port, err := strconv.Atoi(baseURL.Port())
	if err != nil {
		return nil, fmt.Errorf("parse URL port: %v", err)
	}

	return setupClient.Pair(ctx,
		&AccessoryPairingConfig{
			IPConnectionInfo: IPConnectionInfo{
				IPAddress: baseURL.Hostname(),
				Port:      port,
			},
			PIN:      "12344321",
			DeviceID: "5F-7A-CA-6A-83-92",
		},
		controller,
	)
}

// wrapperContext is a temporary workaround for a race reported during test
// due to session cryptographer access.
type wrapperContext struct {
	hap.Context
}

func (w *wrapperContext) SetSessionForConnection(s hap.Session, c net.Conn) {
	w.Context.SetSessionForConnection(&syncSession{session: s}, c)
}

// syncSession works around a race condition accessing crypto in tests
type syncSession struct {
	session hap.Session
	mutex   sync.Mutex
}

// Decrypter returns decrypter for incoming data, may be nil
func (s *syncSession) Decrypter() crypto.Decrypter {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.session.Decrypter()
}

// Encrypter returns encrypter for outgoing data, may be nil
func (s *syncSession) Encrypter() crypto.Encrypter {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.session.Encrypter()
}

// SetCryptographer sets the new cryptographer used for en-/decryption
func (s *syncSession) SetCryptographer(c crypto.Cryptographer) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.session.SetCryptographer(c)
}

// PairSetupHandler returns the pairing setup handler
func (s *syncSession) PairSetupHandler() hap.ContainerHandler {
	return s.session.PairSetupHandler()
}

// SetPairSetupHandler sets the handler for pairing setup
func (s *syncSession) SetPairSetupHandler(c hap.ContainerHandler) {
	s.session.SetPairSetupHandler(c)
}

// PairVerifyHandler returns the pairing verify handler
func (s *syncSession) PairVerifyHandler() hap.PairVerifyHandler {
	return s.session.PairVerifyHandler()
}

// SetPairVerifyHandler sets the handler for pairing verify
func (s *syncSession) SetPairVerifyHandler(c hap.PairVerifyHandler) {
	s.session.SetPairVerifyHandler(c)
}

// Connection returns the associated connection
func (s *syncSession) Connection() net.Conn {
	return s.session.Connection()
}

func (s *syncSession) IsSubscribedTo(ch *characteristic.Characteristic) bool {
	return s.session.IsSubscribedTo(ch)
}

func (s *syncSession) Subscribe(ch *characteristic.Characteristic) {
	s.session.Subscribe(ch)
}

func (s *syncSession) Unsubscribe(ch *characteristic.Characteristic) {
	s.session.Unsubscribe(ch)
}
