package core

import "github.com/spf13/cobra"

type BenchCmd interface {
	NewBenchCmd() *cobra.Command
	NewConfigCmd() *cobra.Command
}
