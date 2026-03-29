package main

import (
	"fmt"
	"os"

	"github.com/afterdarksys/packetcats/pkg/scripting"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [script.star]",
	Short: "Run a PacketCats Starlark script",
	Long:  `Executes an intelligent packetscript.star to generate, munge, and send network packets.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := scripting.RunScript(args[0])
		if err != nil {
			fmt.Printf("Execution failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
