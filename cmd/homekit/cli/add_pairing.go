package cli

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/mctofu/homekit/client"
	"github.com/spf13/cobra"
)

func addPairingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addPairing",
		Short: "Allow an additional controller to access a paired accessory",
	}

	deviceID := cmd.Flags().String("id", "", "Device id of controller to add")
	markFlagRequired(cmd, "id")
	key := cmd.Flags().String("key", "", "Controller public key in base64 or hex")
	markFlagRequired(cmd, "key")
	admin := cmd.Flags().Bool("admin", false, "Allow admin access")

	cmd.RunE = clientCommandRunner(cmd,
		func(ctx context.Context, clientCtx *clientContext, accClient *client.AccessoryClient) error {
			return addPairing(ctx, clientCtx, accClient, *deviceID, *key, *admin)
		},
	)
	return cmd
}

func addPairing(
	ctx context.Context,
	clientCtx *clientContext,
	accClient *client.AccessoryClient,
	deviceID string,
	key string,
	admin bool,
) error {
	publicKey, err := parsePublicKey(key)
	if err != nil {
		return fmt.Errorf("invalid key: %v", err)
	}

	addReq := &client.AddPairingRequest{
		DeviceID:  deviceID,
		PublicKey: publicKey,
	}
	if admin {
		addReq.Permissions = 1
	}

	if err := accClient.AddPairing(ctx, addReq); err != nil {
		return fmt.Errorf("addPairing: %v", err)
	}

	fmt.Println("Add pairing successful. Import these accessory settings on the added controller:")
	for _, pair := range clientCtx.Config.AccessoryPairings {
		if pair.Name == clientCtx.AccessoryName {
			fmt.Printf("ID: %s\n", pair.DeviceID)
			fmt.Printf("Key: %s\n", base64.StdEncoding.EncodeToString(pair.PublicKey))
			break
		}
	}

	return nil
}
