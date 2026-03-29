package main

import (
	"fmt"
	"os"

	"github.com/afterdarksys/packetcats/pkg/scripting"
	"github.com/spf13/cobra"
)

var scriptFlag string

var decodeCmd = &cobra.Command{
	Use:   "decode",
	Short: "Decode a pcap stream from STDIN",
	Long:  `Reads a PCAP stream from STDIN and streams each packet into a Starlark script for decoding/filtering using packet_hook.`,
	Run: func(cmd *cobra.Command, args []string) {
		if scriptFlag == "" {
			fmt.Println("Error: --script is required for decoding via starlark")
			os.Exit(1)
		}

		err := scripting.RunStream(scriptFlag)
		if err != nil {
			fmt.Printf("Decode failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	decodeCmd.Flags().StringVar(&scriptFlag, "script", "", "Starlark script file (must contain packet_hook)")
	rootCmd.AddCommand(decodeCmd)
}
