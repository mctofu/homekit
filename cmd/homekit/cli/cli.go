package cli

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/mctofu/homekit/client"
	"github.com/mctofu/homekit/cmd/homekit/cli/config"
	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use: "homekit",
}

func init() {
	rootCommand.AddCommand(createControllerCmd())
	rootCommand.AddCommand(discoverCmd())
	rootCommand.AddCommand(listPairedAccessoriesCmd())
	rootCommand.AddCommand(listPairingsCmd())
	rootCommand.AddCommand(pairCmd())
	rootCommand.AddCommand(unpairCmd())
	rootCommand.AddCommand(listCharacteristicsCmd())
	rootCommand.AddCommand(getCharacteristicsCmd())
	rootCommand.AddCommand(setCharacteristicsCmd())
	rootCommand.AddCommand(addPairingCmd())
	rootCommand.AddCommand(importPairingCmd())
}

// Execute the command line interface
func Execute() error {
	return rootCommand.ExecuteContext(context.Background())
}

func dump(v interface{}) {
	j, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal: %v\n", err)
		return
	}

	fmt.Printf("%s\n", string(j))
}

func markFlagRequired(cmd *cobra.Command, name string) {
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}

// runner is cobra.Command.RunE
type runner func(cmd *cobra.Command, args []string) error

type configCommand func(ctx context.Context, configPath, controllerName string) error

func configCommandRunner(cmd *cobra.Command, cfgCmd configCommand) runner {
	configDir, configDirErr := os.UserConfigDir()
	defaultConfigPath := path.Join(configDir, "mctofu", "homekit")

	configPath := cmd.Flags().String("configPath", defaultConfigPath, "Directory to store controller configs.")
	controllerName := cmd.Flags().String("controller", "default", "Name of controller profile to use.")

	return func(cmd *cobra.Command, args []string) error {
		if *configPath == "" && configDirErr != nil {
			return fmt.Errorf("resolve config dir: %v", configDirErr)
		}

		return cfgCmd(cmd.Context(), *configPath, *controllerName)
	}
}

type clientContext struct {
	AccessoryName string
	Config        *config.ControllerConfig
	ConfigPath    string
}

type clientCommand func(ctx context.Context, clientCtx *clientContext, accClient *client.AccessoryClient) error

func clientCommandRunner(cmd *cobra.Command, clientCmd clientCommand) runner {
	name := cmd.Flags().StringP("name", "n", "", "Name of accessory to act on")
	markFlagRequired(cmd, "name")

	cfgCmd := func(ctx context.Context, configPath, controllerName string) error {
		cfg, err := config.ReadControllerConfig(configPath, controllerName)
		if err != nil {
			return fmt.Errorf("read controller config: %v", err)
		}

		accClient, err := accessoryClient(cfg, *name)
		if err != nil {
			return err
		}

		clientCtx := clientContext{
			AccessoryName: *name,
			Config:        cfg,
			ConfigPath:    configPath,
		}

		return clientCmd(cmd.Context(), &clientCtx, accClient)
	}

	return configCommandRunner(cmd, cfgCmd)
}

func accessoryClient(cfg *config.ControllerConfig, name string) (*client.AccessoryClient, error) {
	var accPairing *config.AccessoryPairing

	for _, pair := range cfg.AccessoryPairings {
		if pair.Name == name {
			accPairing = pair
			break
		}
	}

	if accPairing == nil {
		return nil, fmt.Errorf("accessory %s not found", name)
	}

	return client.NewAccessoryClient(
		client.NewIPDialer(),
		&client.ControllerIdentity{
			DeviceID:   cfg.DeviceID,
			PrivateKey: cfg.PrivateKey,
			PublicKey:  cfg.PublicKey,
		},
		&client.AccessoryConnectionConfig{
			DeviceID:         accPairing.DeviceID,
			PublicKey:        accPairing.PublicKey,
			IPConnectionInfo: accPairing.IPConnectionInfo,
		},
	), nil
}

func parseCharacteristicIDs(charateristicIDParam string) (accID, chID uint64, err error) {
	parts := strings.Split(charateristicIDParam, ".")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid characteristicID format %s, expect x.x", charateristicIDParam)
	}
	accID, err = strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("could not parse accessoryID %s: %v", parts[0], err)
	}
	chID, err = strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("could not parse charactersticID %s: %v", parts[1], err)
	}
	return accID, chID, nil
}

func parsePublicKey(key string) ([]byte, error) {
	if len(key) == 64 {
		return hex.DecodeString(key)
	}

	return base64.StdEncoding.DecodeString(key)
}
