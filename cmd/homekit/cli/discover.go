package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/mctofu/homekit/client"
	"github.com/spf13/cobra"
)

func discoverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover available homekit accessories",
	}

	timeout := cmd.Flags().Int("timeout", 10, "number of seconds to wait for devices to respond")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return discover(cmd.Context(), *timeout)
	}

	return cmd
}

func discover(ctx context.Context, timeout int) error {
	var found uint64

	foundFn := func(ctx context.Context, d *client.AccessoryDevice) {
		found++
		fmt.Printf("Detected Device: %s\n", d.Name)
		fmt.Printf("Model: %s\n", d.Model)
		fmt.Printf("ID: %s\n", d.ID)
		fmt.Printf("IPs: %v\n", d.IPs)
		fmt.Printf("Port: %d\n", d.Port)
		fmt.Printf("Feature flags: %s (%d)\n", d.FeatureFlags, d.FeatureFlags)
		fmt.Printf("Status flags: %s (%d)\n", d.StatusFlags, d.StatusFlags)
		fmt.Printf("\n")
	}

	if err := client.Discover(ctx, foundFn, time.Duration(timeout)*time.Second); err != nil {
		return err
	}

	fmt.Printf("Found %d devices\n", found)

	return nil
}
