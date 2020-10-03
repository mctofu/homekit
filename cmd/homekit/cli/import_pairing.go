package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/mctofu/homekit/client"
	"github.com/mctofu/homekit/cmd/homekit/cli/config"
	"github.com/spf13/cobra"
)

func importPairingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "importPairing",
		Short: "Import a known pairing",
	}

	deviceID := cmd.Flags().String("id", "", "Device id of accessory to import")
	markFlagRequired(cmd, "id")
	key := cmd.Flags().String("key", "", "Accessory public key in base64 or hex")
	markFlagRequired(cmd, "key")
	name := cmd.Flags().StringP("name", "n", "", "Alias to reference accessory by")
	markFlagRequired(cmd, "name")

	cmd.RunE = configCommandRunner(cmd,
		func(ctx context.Context, configPath, controllerName string) error {
			return importPairing(ctx, configPath, controllerName, *deviceID, *key, *name)
		},
	)

	return cmd
}

func importPairing(ctx context.Context, configPath, controllerName string, deviceID, key, name string) error {
	cfg, err := config.ReadControllerConfig(configPath, controllerName)
	if err != nil {
		return fmt.Errorf("read controller config: %v", err)
	}

	for _, pair := range cfg.AccessoryPairings {
		if pair.DeviceID == deviceID {
			return fmt.Errorf("%s is already paired as %s", pair.DeviceID, pair.Name)
		}
		if pair.Name == name {
			return fmt.Errorf("%s is already aliased to %s", pair.Name, pair.DeviceID)
		}
	}

	publicKey, err := parsePublicKey(key)
	if err != nil {
		return fmt.Errorf("invalid key: %v", err)
	}

	pairDevice, err := client.DeviceByID(ctx, deviceID, 10*time.Second)
	if err != nil {
		return fmt.Errorf("deviceByID: %v", err)
	}

	cfg.AccessoryPairings = append(cfg.AccessoryPairings,
		&config.AccessoryPairing{
			Name:       name,
			DeviceName: pairDevice.Name,
			Model:      pairDevice.Model,
			DeviceID:   deviceID,
			PublicKey:  publicKey,
			IPConnectionInfo: client.IPConnectionInfo{
				IPAddress: pairDevice.IPs[0].String(),
				Port:      pairDevice.Port,
			},
		})

	if err := config.SaveControllerConfig(configPath, cfg, true); err != nil {
		return fmt.Errorf("saveControllerConfig: %v", err)
	}

	fmt.Println("Pairing imported")

	return nil
}
