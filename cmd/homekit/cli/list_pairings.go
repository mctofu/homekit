package cli

import (
	"context"
	"fmt"

	"github.com/mctofu/homekit/client"
	"github.com/spf13/cobra"
)

func listPairingsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "listPairings",
		Short: "List controllers paired to an accessory",
	}

	cmd.RunE = clientCommandRunner(cmd, listPairings)

	return cmd
}

func listPairings(ctx context.Context, clientCtx *clientContext, accClient *client.AccessoryClient) error {
	pairs, err := accClient.ListPairings(ctx)
	if err != nil {
		return fmt.Errorf("listPairings: %v", err)
	}

	dump(pairs)

	return nil
}
