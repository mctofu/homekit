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

	"github.com/brutella/dnssd"
)

const (
	homekitService = "_hap._tcp.local."
)

// FeatureFlags captures Bonjour TXT feature flags
type FeatureFlags byte

func (f FeatureFlags) String() string {
	switch f {
	case 0:
		return "No support for HAP Pairing"
	case 1:
		return "Supports HAP Pairing"
	default:
		return "Unknown"
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

	addFn := func(srv dnssd.Service) {
		onDevice(ctx, &AccessoryDevice{
			Name:         srv.Name,
			ID:           srv.Text["id"],
			Model:        srv.Text["md"],
			IPs:          srv.IPs,
			Port:         srv.Port,
			FeatureFlags: FeatureFlags(parseFlag(srv.Text["ff"])),
			StatusFlags:  StatusFlags(parseFlag(srv.Text["sf"])),
		})
	}

	rmvFn := func(srv dnssd.Service) {
	}

	if err := dnssd.LookupType(ctx, homekitService, addFn, rmvFn); err != nil && err != context.DeadlineExceeded {
		return err
	}

	return nil
}

func parseFlag(v string) byte {
	if v == "" {
		return 0
	}
	f, err := strconv.Atoi(v)
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
