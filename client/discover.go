package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
)

const (
	homekitService = "_hap._tcp"
	homekitDomain  = ".local"
)

// FeatureFlags captures Bonjour TXT feature flags
type FeatureFlags byte

func (f FeatureFlags) String() string {
	var pairingTypes []string
	switch {
	case f&1 > 0:
		pairingTypes = append(pairingTypes, "Hardware")
	case f&2 > 0:
		pairingTypes = append(pairingTypes, "Software")
	}

	if len(pairingTypes) == 0 {
		return "No support for HAP Pairing or uncertified"
	}
	return fmt.Sprintf("Supports HAP Pairing with %s authentication", strings.Join(pairingTypes, " and "))
}

func (f FeatureFlags) PairingMethod() PairingMethod {
	switch {
	case f&1 > 0:
		return PairingMethodPairSetupWithAuth
	default:
		return PairingMethodPairSetup
	}
}

// StatusFlags captures Bonjour TXT status flags
type StatusFlags byte

func (s StatusFlags) String() string {
	var notes []string
	if s&1 == 1 {
		notes = append(notes, "is not paired")
	} else {
		notes = append(notes, "is paired")
	}

	if s&2 == 2 {
		notes = append(notes, "is not configured for wifi")
	}

	if s&4 == 4 {
		notes = append(notes, "has a problem")
	}

	return "Accessory " + strings.Join(notes, "/")
}

// AccessoryDevice is information about a discovered HomeKit accessory.
type AccessoryDevice struct {
	Name         string
	Model        string
	ID           string
	IPs          []net.IP
	Port         int
	FeatureFlags FeatureFlags
	StatusFlags  StatusFlags
}

// Discover searches for HomeKit accessory devices on the network for up to searchDuration. When a device is found
// the onDevice callback is triggered with information about the discovered device.
func Discover(ctx context.Context, onDevice func(context.Context, *AccessoryDevice), searchDuration time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, searchDuration)
	defer cancel()

	resolver, err := zeroconf.NewResolver()
	if err != nil {
		return fmt.Errorf("failed to initialize resolver: %v", err)
	}

	var devicesWG sync.WaitGroup
	devicesWG.Add(1)
	devicesCh := make(chan *zeroconf.ServiceEntry)
	go func() {
		for dev := range devicesCh {
			txt := parseTXT(dev.Text)
			onDevice(ctx, &AccessoryDevice{
				Name:         dev.Instance,
				ID:           txt["id"],
				Model:        txt["md"],
				IPs:          append(dev.AddrIPv4, dev.AddrIPv6...),
				Port:         dev.Port,
				FeatureFlags: FeatureFlags(parseFlag(txt["ff"])),
				StatusFlags:  StatusFlags(parseFlag(txt["sf"])),
			})
		}
		devicesWG.Done()
	}()

	if err := resolver.Browse(ctx, homekitService, homekitDomain, devicesCh); err != nil {
		return fmt.Errorf("browse: %v", err)
	}

	devicesWG.Wait()
	<-ctx.Done()
	if err := ctx.Err(); err != nil && err != context.DeadlineExceeded {
		return err
	}

	return nil
}

func parseTXT(txts []string) map[string]string {
	mapped := make(map[string]string)

	for _, txt := range txts {
		parts := strings.SplitN(txt, "=", 2)
		if len(parts) == 2 {
			key := strings.ToLower(parts[0])
			value := parts[1]

			mapped[key] = value
		}
	}

	return mapped
}

func parseFlag(v string) byte {
	if v == "" {
		return 0
	}
	f, err := strconv.ParseUint(v, 10, 8)
	if err != nil {
		panic(err)
	}

	return byte(f)
}

// DeviceByID searches for a device with the provided deviceID and returns it if found.
// If there hasn't been a match within searchDuration then an error is returned.
func DeviceByID(ctx context.Context, deviceID string, searchDuration time.Duration) (*AccessoryDevice, error) {
	var device *AccessoryDevice

	ctx, cancel := context.WithCancel(ctx)

	onDevice := func(ctx context.Context, d *AccessoryDevice) {
		if d.ID == deviceID {
			device = d
			cancel()
		}
	}

	var err error

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		err = Discover(ctx, onDevice, searchDuration)
		wg.Done()
	}()

	wg.Wait()

	if device != nil {
		return device, nil
	}

	if err != nil {
		return nil, fmt.Errorf("device not found: %v", err)
	}

	return nil, errors.New("device not found")
}
