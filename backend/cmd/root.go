package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fasim",
	Short: "Factory Automation Simulator",
	Long: `Factory Automation Simulator (Fasim) is a tool for simulating and
optimizing manufacturing processes in factory automation systems.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
