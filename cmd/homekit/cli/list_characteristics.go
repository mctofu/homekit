package cli

import (
	"context"
	"fmt"

	"github.com/mctofu/homekit/client"
	"github.com/mctofu/homekit/client/characteristic"
	"github.com/mctofu/homekit/client/service"
	"github.com/spf13/cobra"
)

func listCharacteristicsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "listCharacteristics",
		Aliases: []string{"listC"},
		Short:   "List available services and characteristics of an accessory",
	}

	cmd.RunE = clientCommandRunner(cmd, listCharacteristics)

	return cmd
}

func listCharacteristics(ctx context.Context, clientCtx *clientContext, accClient *client.AccessoryClient) error {
	accessories, err := accClient.Accessories(ctx)
	if err != nil {
		return err
	}

	for _, acc := range accessories {
		accInfo := acc.Info()
		fmt.Printf("Accessory: %d %s (%s)\n", acc.ID, accInfo.Name.Value, accInfo.SerialNumber.Value)

		for _, svc := range acc.Services {
			fmt.Printf("  Service: %d %s (%s)\n", svc.ID, service.NameForType(svc.Type), svc.Type)
			for _, ch := range svc.Characteristics {
				fmt.Printf("    %d.%d: %v / %s (%s) %v\n", acc.ID, ch.ID, characteristic.ValueForFormat(ch.Format, ch.Value), characteristic.NameForType(ch.Type), ch.Type, ch.Permissions)
			}
		}
	}

	return nil
}
