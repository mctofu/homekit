package cli

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/mctofu/homekit/client"
	"github.com/mctofu/homekit/cmd/homekit/cli/config"
	"github.com/spf13/cobra"
)

func pairCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pair",
		Short: "Pair an accessory to a controller",
	}

	deviceID := cmd.Flags().String("id", "", "Device id of accessory to pair (from discovery)")
	markFlagRequired(cmd, "id")
	pin := cmd.Flags().String("pin", "", "Accessory PIN for pairing (XXX-XX-XXX)")
	markFlagRequired(cmd, "pin")
	name := cmd.Flags().StringP("name", "n", "", "Alias to reference accessory by")
	markFlagRequired(cmd, "name")

	cmd.RunE = configCommandRunner(cmd,
		func(ctx context.Context, configPath, controllerName string) error {
			return pair(ctx, configPath, controllerName, *deviceID, *pin, *name)
		},
	)

	return cmd
}

func pair(
	ctx context.Context,
	configPath,
	controllerName string,
	deviceID,
	pin,
	name string) error {

	cfg, err := config.ReadControllerConfig(configPath, controllerName)
	if err != nil {
		return err
	}

	for _, pair := range cfg.AccessoryPairings {
		if pair.DeviceID == deviceID {
			return fmt.Errorf("%s is already paired as %s", pair.DeviceID, pair.Name)
		}
		if pair.Name == name {
			return fmt.Errorf("%s is already aliased to %s", pair.Name, pair.DeviceID)
		}
	}

	pairDevice, err := client.DeviceByID(ctx, deviceID, 10*time.Second)
	if err != nil {
		return fmt.Errorf("deviceByID: %v", err)
	}

	pairClient := client.NewSetupClient(&http.Client{})
	accConn, err := pairClient.Pair(
		ctx,
		&client.AccessoryPairingConfig{
			PIN:      pin,
			DeviceID: deviceID,
			IPConnectionInfo: client.IPConnectionInfo{
				IPAddress: pairDevice.IPs[0].String(),
				Port:      pairDevice.Port,
			},
		},
		&client.ControllerIdentity{
			DeviceID:   cfg.DeviceID,
			PublicKey:  cfg.PublicKey,
			PrivateKey: cfg.PrivateKey,
		},
	)
	if err != nil {
		return fmt.Errorf("pair: %v", err)
	}

	cfg.AccessoryPairings = append(cfg.AccessoryPairings,
		&config.AccessoryPairing{
			Name:             name,
			DeviceID:         accConn.DeviceID,
			PublicKey:        accConn.PublicKey,
			IPConnectionInfo: accConn.IPConnectionInfo,
		})

	if err := config.SaveControllerConfig(configPath, cfg, true); err != nil {
		// The accessory may only allow a single pairing so we'll be locked out without
		// this information.
		fmt.Printf("Manual pairing import information:\n")
		fmt.Printf("Device ID: %s\n", accConn.DeviceID)
		fmt.Printf("Public Key: %s\n", hex.EncodeToString(accConn.PublicKey))

		return fmt.Errorf("could not save pairing - review manual pairing info: %v", err)
	}

	fmt.Println("Accessory paired successfully!")

	return nil
}
