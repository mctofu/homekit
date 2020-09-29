package cli

import (
	"context"
	"fmt"

	"github.com/mctofu/homekit/client"
	"github.com/mctofu/homekit/cmd/homekit/cli/config"
	"github.com/spf13/cobra"
)

func unpairCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unpair",
		Short: "Unpair an accessory from the controller",
	}

	cmd.RunE = clientCommandRunner(cmd, unpair)

	return cmd
}

func unpair(ctx context.Context, clientCtx *clientContext, accClient *client.AccessoryClient) error {
	cfg := *clientCtx.Config

	if err := accClient.RemovePairing(ctx, cfg.DeviceID); err != nil {
		return fmt.Errorf("removePairing: %v", err)
	}

	newPairings := make([]*config.AccessoryPairing, 0, len(cfg.AccessoryPairings)-1)
	for _, pairing := range cfg.AccessoryPairings {
		if pairing.Name != clientCtx.AccessoryName {
			newPairings = append(newPairings, pairing)
		}
	}
	cfg.AccessoryPairings = newPairings

	if err := config.SaveControllerConfig(clientCtx.ConfigPath, &cfg, true); err != nil {
		return err
	}

	fmt.Println("Pairing removed")

	return nil
}
