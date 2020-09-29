package cli

import (
	"context"
	"fmt"

	"github.com/mctofu/homekit/client"
	"github.com/mctofu/homekit/client/characteristic"
	"github.com/spf13/cobra"
)

func getCharacteristicsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "getCharacteristics",
		Aliases: []string{"getC"},
		Short:   "Read values of accessory characteristics",
	}

	characteristicIDs := cmd.Flags().StringArrayP("c", "c", nil, "Characteristic ID (ex: 1.4)")
	markFlagRequired(cmd, "c")

	cmd.RunE = clientCommandRunner(cmd,
		func(ctx context.Context, clientCtx *clientContext, accClient *client.AccessoryClient) error {
			return getCharacteristics(ctx, accClient, *characteristicIDs)
		},
	)

	return cmd
}

func getCharacteristics(ctx context.Context, accClient *client.AccessoryClient, characteristicIDs []string) error {
	var cReqs []client.CharacteristicReadRequest

	for _, cID := range characteristicIDs {
		accID, chID, err := parseCharacteristicIDs(cID)
		if err != nil {
			return err
		}

		cReqs = append(cReqs, client.CharacteristicReadRequest{
			AccessoryID:      accID,
			CharacteristicID: chID,
		})
	}

	req := &client.CharacteristicsReadRequest{
		Characteristics: cReqs,
		Metadata:        true,
		Type:            true,
	}

	resps, err := accClient.Characteristics(ctx, req)
	if err != nil {
		return err
	}

	for _, resp := range resps {
		fmt.Printf("%d.%d: %s\n", resp.AccessoryID, resp.CharacteristicID, characteristic.NameForType(*resp.Type))
		// Velux is not returning format as part of the metadata so we rely on known types
		// to determine the value format. We can consider getting the format from the
		// accClient.Accessories response instead.
		fmt.Printf("Value: %v\n", characteristic.ValueForType(*resp.Type, resp.Value))
	}

	return nil
}
