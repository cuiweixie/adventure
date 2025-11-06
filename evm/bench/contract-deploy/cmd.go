package contractdeploy

import "github.com/spf13/cobra"

const (
	FlagContractBytecode = "bytecode"
	FlagBytecodeFile     = "bytecode-file"
	FlagGasLimit         = "gas-limit"
)

func ContractDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract-deploy",
		Short: "deploy contract for performance testing",
		Run:   contractDeploy,
	}
	cmd.Flags().StringVar(&configPath, "f", "", "the location of transfer config file")
	cmd.Flags().StringVar(&contractBytecode, FlagContractBytecode, "", "contract bytecode (hex string)")
	cmd.Flags().StringVar(&bytecodeFile, FlagBytecodeFile, "", "path to file containing contract bytecode (hex string)")
	cmd.Flags().Uint64Var(&gasLimit, FlagGasLimit, 1000000, "gas limit for contract deployment")
	return cmd
}

var (
	configPath       string
	contractBytecode string
	bytecodeFile     string
	gasLimit         uint64
)

