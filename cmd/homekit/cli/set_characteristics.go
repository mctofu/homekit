package cli

import (
	"context"
	"fmt"

	"github.com/mctofu/homekit/client"
	"github.com/mctofu/homekit/client/characteristic"
	"github.com/spf13/cobra"
)

func setCharacteristicsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "setCharacteristics",
		Aliases: []string{"setC"},
		Short:   "Set values of accessory characteristics",
	}

	characteristicParams := cmd.Flags().StringToStringP("c", "c", nil, "Characteristic ID value pair (ex: 1.4=10)")
	markFlagRequired(cmd, "c")

	cmd.RunE = clientCommandRunner(cmd,
		func(ctx context.Context, clientCtx *clientContext, accClient *client.AccessoryClient) error {
			return setCharacteristics(ctx, accClient, *characteristicParams)
		},
	)

	return cmd
}

func setCharacteristics(ctx context.Context, accClient *client.AccessoryClient, characteristicParams map[string]string) error {
	var reads []client.CharacteristicReadRequest

	for k := range characteristicParams {
		accID, chID, err := parseCharacteristicIDs(k)
		if err != nil {
			return err
		}

		reads = append(reads,
			client.CharacteristicReadRequest{
				AccessoryID:      accID,
				CharacteristicID: chID,
			},
		)
	}

	req := &client.CharacteristicsReadRequest{
		Characteristics: reads,
		Metadata:        true,
		Type:            true,
	}

	resps, err := accClient.Characteristics(ctx, req)
	if err != nil {
		return err
	}

	var writes []client.CharacteristicWriteRequest

	for _, resp := range resps {
		cKey := fmt.Sprintf("%d.%d", resp.AccessoryID, resp.CharacteristicID)
		val, ok := characteristicParams[cKey]
		if !ok {
			return fmt.Errorf("unexpected characteristic returned: %s", cKey)
		}
		writeVal, err := characteristic.ParseValueForType(*resp.Type, val)
		if err != nil {
			return err
		}
		writes = append(writes,
			client.CharacteristicWriteRequest{
				AccessoryID:      resp.AccessoryID,
				CharacteristicID: resp.CharacteristicID,
				Value:            writeVal,
			},
		)
	}

	writeReq := &client.CharacteristicsWriteRequest{
		Characteristics: writes,
	}

	_, err = accClient.SetCharacteristics(ctx, writeReq)
	if err != nil {
		return err
	}

	return nil
}
