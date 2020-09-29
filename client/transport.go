package client

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/brutella/hc/db"
	"github.com/brutella/hc/hap"
	"github.com/google/uuid"
)

// IPTransport provides http client capabilities.
// TODO: Look at what BluetoothTransport would take
type IPTransport interface {
	Do(req *http.Request) (*http.Response, error)
}

// IPConnectionInfo identifies a HomeKit accessory on an ip network.
type IPConnectionInfo struct {
	IPAddress string
	Port      int
}

// NewRandomControllerConfig returns a new config with random keys and a
// random id.
func NewRandomControllerConfig() (*ControllerIdentity, error) {
	ent, err := db.NewRandomEntityWithName("ignored")
	if err != nil {
		return nil, fmt.Errorf("generate key pairs: %v", err)
	}

	return &ControllerIdentity{
		DeviceID:   uuid.New().String(),
		PublicKey:  ent.PublicKey,
		PrivateKey: ent.PrivateKey,
	}, nil
}

// IPDialer establishes a network connection
type IPDialer func(ctx context.Context, network, address string) (net.Conn, error)

// NewIPDialer returns a default IPDialer that returns a basic connection
func NewIPDialer() IPDialer {
	return (&net.Dialer{}).DialContext
}

// HomeKitSecureDialer negotiates a secure connection with an accessory using the
// pair-verify procedure.
type HomeKitSecureDialer struct {
	dialer     IPDialer
	accessory  *AccessoryConnectionConfig
	controller *ControllerIdentity
	conn       net.Conn
	connMux    sync.Mutex
}

// NewHomeKitSecureDialer returns a new HomeKitSecureDialer suitable for use between controller c and
// accessory a.
func NewHomeKitSecureDialer(dialer IPDialer, c *ControllerIdentity, a *AccessoryConnectionConfig) *HomeKitSecureDialer {
	return &HomeKitSecureDialer{
		dialer:     dialer,
		accessory:  a,
		controller: c,
	}
}

// Dial will create a connection and negotiate secure communication using the pair-verify procedure
// prior to returning it. Further communication on the connection will be transparently encrypted.
// If a connection has already been established then the existing connection is returned.
func (h *HomeKitSecureDialer) Dial(ctx context.Context, network, addr string) (net.Conn, error) {
	h.connMux.Lock()
	defer h.connMux.Unlock()

	if h.conn == nil {
		conn, err := h.establishConnection(ctx, network, addr)
		if err != nil {
			return nil, fmt.Errorf("establish connection: %v", err)
		}
		h.conn = conn
	}

	return h.conn, nil
}

func (h *HomeKitSecureDialer) establishConnection(ctx context.Context, network, addr string) (net.Conn, error) {
	conn, err := h.dialer(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	detachableConn := &detachableConnection{conn: conn}
	defer detachableConn.Detach()

	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
			DialContext: func(ctx context.Context, dialNetwork, dialAddr string) (net.Conn, error) {
				if dialNetwork != network || dialAddr != addr {
					return nil, fmt.Errorf("no connection found for %s:%s", dialNetwork, dialAddr)
				}
				return detachableConn, nil
			},
		},
	}
	defer httpClient.CloseIdleConnections()

	verifyClient := NewVerifyClient(httpClient, h.accessory, h.controller)

	cryptographer, err := verifyClient.Verify(ctx)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("pair verify: %v", err)
	}

	// TODO: more direct way to setup encryption session on connection
	dev, err := hap.NewSecuredDevice("unused", "11122333", &memoryDB{})
	if err != nil {
		return nil, fmt.Errorf("init device: %v", err)
	}
	hapCtx := hap.NewContextForSecuredDevice(dev)

	hapConn, err := hap.NewConnection(conn, hapCtx), nil
	if err != nil {
		return nil, fmt.Errorf("hap.NewConnection: %v", err)
	}

	sess := hapCtx.GetSessionForConnection(hapConn)
	sess.SetCryptographer(cryptographer)
	// hack: trigger Encrypter setup
	_ = sess.Decrypter()

	return hapConn, nil
}

// Close any underlying connections.
func (h *HomeKitSecureDialer) Close() error {
	h.connMux.Lock()
	defer h.connMux.Unlock()

	if h.conn != nil {
		return h.conn.Close()
	}
	return nil
}

type detachableConnection struct {
	conn     net.Conn
	closeMux sync.Mutex
	readWG   sync.WaitGroup
	closed   bool
}

// Read reads from the underlying connection. When detaching the connection
// a timeout error is triggered which is ignored here.
func (i *detachableConnection) Read(b []byte) (n int, err error) {
	i.readWG.Add(1)
	defer i.readWG.Done()

	n, err = i.conn.Read(b)
	i.closeMux.Lock()
	defer i.closeMux.Unlock()
	if err, ok := err.(net.Error); ok && i.closed && err.Timeout() {
		// return and signal EOF to unblock readers
		// TODO: what is the "correct" error to return here to indicate
		// the connection is "closed"
		return n, io.EOF
	}

	return
}

// Write writes data to the underlying connection.
func (i *detachableConnection) Write(b []byte) (n int, err error) {
	return i.conn.Write(b)
}

// Close interrupts any blocked Read or Write operations by setting an immediate
// deadline. Afterwards call Detach to reset the deadline after reads have finished.
// It does not close the underlying connection.
func (i *detachableConnection) Close() error {
	i.closeMux.Lock()
	defer i.closeMux.Unlock()
	i.closed = true
	// unblock any active reads/writes
	return i.conn.SetDeadline(time.Now())
}

// Detach waits until reads have finished and then resets the connection deadline.
// It should be called after Close() to detach the underlying connection
// from any active readers. After it completes it should be safe to reuse the
// connection.
func (i *detachableConnection) Detach() {
	i.readWG.Wait()
	// ignore for now. we'll likely find out later if something is wrong anyway.
	_ = i.conn.SetDeadline(time.Time{})
	i.conn = nil
}

// LocalAddr returns the local network address fom the underlying connection.
func (i *detachableConnection) LocalAddr() net.Addr {
	return i.conn.LocalAddr()
}

// RemoteAddr returns the remote network address from the underlying connection.
func (i *detachableConnection) RemoteAddr() net.Addr {
	return i.conn.RemoteAddr()
}

// SetDeadline delegates to the underlying connection.
func (i *detachableConnection) SetDeadline(t time.Time) error {
	return i.conn.SetDeadline(t)
}

// SetReadDeadline delegates to the underlying connection.
func (i *detachableConnection) SetReadDeadline(t time.Time) error {
	return i.conn.SetReadDeadline(t)
}

// SetWriteDeadline delegates to the underlying connection.
func (i *detachableConnection) SetWriteDeadline(t time.Time) error {
	return i.conn.SetWriteDeadline(t)
}
