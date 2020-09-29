package cli

import (
	"context"
	"fmt"

	"github.com/mctofu/homekit/client"
	"github.com/mctofu/homekit/cmd/homekit/cli/config"
	"github.com/spf13/cobra"
)

func createControllerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "createController",
		Short: "Initialize a controller to pair with accessories",
	}

	cmd.RunE = configCommandRunner(cmd, createController)

	return cmd
}

func createController(ctx context.Context, configPath, controllerName string) error {
	controllerCfg, err := client.NewRandomControllerConfig()
	if err != nil {
		return fmt.Errorf("NewRandomControllerConfig: %v", err)
	}
	controllerPairings := &config.ControllerConfig{
		Name:       controllerName,
		DeviceID:   controllerCfg.DeviceID,
		PublicKey:  controllerCfg.PublicKey,
		PrivateKey: controllerCfg.PrivateKey,
	}

	if err := config.SaveControllerConfig(configPath, controllerPairings, false); err != nil {
		return err
	}

	fmt.Printf("Created controller %s\n", controllerName)

	return nil
}
