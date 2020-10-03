package config

import (
	"testing"

	"github.com/mctofu/homekit/client"
	"github.com/stretchr/testify/require"
)

func TestConfigSaveAndRead(t *testing.T) {
	cfgPath := t.TempDir()

	// create
	require.NoError(t, SaveControllerConfig(cfgPath, controllerConfig("t"), false))
	readCfg, err := ReadControllerConfig(cfgPath, "test-controller")
	require.NoError(t, err)
	require.Equal(t, controllerConfig("t"), readCfg)

	// update
	require.NoError(t, SaveControllerConfig(cfgPath, controllerConfig("a"), true))
	updatedCfg, err := ReadControllerConfig(cfgPath, "test-controller")
	require.NoError(t, err)
	require.Equal(t, controllerConfig("a"), updatedCfg)

	// reject create if already existing
	require.Error(t, SaveControllerConfig(cfgPath, controllerConfig("a"), false))
}

func controllerConfig(model string) *ControllerConfig {
	return &ControllerConfig{
		DeviceID:   "dID",
		Name:       "test-controller",
		PublicKey:  []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		PrivateKey: []byte{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		AccessoryPairings: []*AccessoryPairing{
			{
				DeviceID:   "AA:BB",
				DeviceName: "Test",
				Name:       "alias",
				IPConnectionInfo: client.IPConnectionInfo{
					IPAddress: "127.0.0.1",
					Port:      5001,
				},
				Model:     model,
				PublicKey: []byte{1, 2, 3, 4, 5},
			},
		},
	}
}
