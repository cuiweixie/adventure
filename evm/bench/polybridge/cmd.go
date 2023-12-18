package polybridge

import "github.com/spf13/cobra"

var (
	// used for flags
	configPath string
)

func PolyBridgeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "poly-bridge",
		Short: "send bridgeAsset to L1 and monitor claimAsset on L2",
		Run:   polybridge,
	}
	cmd.Flags().StringVar(&configPath, "f", "", "the location of transfer config file")
	return cmd
}
