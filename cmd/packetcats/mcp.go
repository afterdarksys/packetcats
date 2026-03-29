package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/afterdarksys/packetcats/pkg/scripting"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start a Model Context Protocol (MCP) server for SuperPacketCat",
	Long: `Starts a JSON-RPC MCP server over standard input and output.
This allows AI agents (like Claude Desktop or an MCP-compatible assistant)
to natively invoke PacketCats Starlark execution tools.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Log that we're starting over STDERR so we don't pollute STDOUT JSON-RPC
		fmt.Fprintf(os.Stderr, "SuperPacketCat MCP Server listening on stdio...\n")
		scripting.StartMCPServer()
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
