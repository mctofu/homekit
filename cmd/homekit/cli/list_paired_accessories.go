package cli

import (
	"context"
	"fmt"

	"github.com/mctofu/homekit/cmd/homekit/cli/config"
	"github.com/spf13/cobra"
)

func listPairedAccessoriesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "listPairedAccessories",
		Short: "List accessories paired with controller",
	}

	cmd.RunE = configCommandRunner(cmd, listPairedAccessories)

	return cmd
}

func listPairedAccessories(ctx context.Context, configPath, controllerName string) error {
	cfg, err := config.ReadControllerConfig(configPath, controllerName)
	if err != nil {
		return fmt.Errorf("read controller config: %v", err)
	}

	if len(cfg.AccessoryPairings) == 0 {
		fmt.Println("No paired accessories")
		return nil
	}

	for _, acc := range cfg.AccessoryPairings {
		fmt.Printf("%s (%s)\n", acc.Name, acc.DeviceID)
	}

	return nil
}
