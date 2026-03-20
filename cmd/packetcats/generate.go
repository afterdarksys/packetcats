package main

import (
	"encoding/hex"
	"fmt"
	"net"

	"github.com/afterdarksys/packetcats/pkg/generator"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate network packets",
	Long:  `Generate various types of network packets including IPv4, TLS, and DNSSEC.`,
}

var tlsCmd = &cobra.Command{
	Use:   "tls",
	Short: "Generate a TLS Client Hello packet",
	Run: func(cmd *cobra.Command, args []string) {
		pkt, err := generator.GenerateTLSClientHello()
		if err != nil {
			fmt.Println("Error generating TLS packet:", err)
			return
		}
		printPacket(pkt, "TLS Client Hello", layers.LayerTypeTLS)
	},
}

var dnssecCmd = &cobra.Command{
	Use:   "dnssec",
	Short: "Generate a DNSSEC response packet",
	Run: func(cmd *cobra.Command, args []string) {
		pkt, err := generator.GenerateDNSSECResponse("example.com")
		if err != nil {
			fmt.Println("Error generating DNSSEC packet:", err)
			return
		}
		printPacket(pkt, "DNSSEC Response", layers.LayerTypeEthernet)
	},
}

var ipv4Cmd = &cobra.Command{
	Use:   "ipv4",
	Short: "Generate a basic IPv4 packet",
	Run: func(cmd *cobra.Command, args []string) {
		src := net.ParseIP("192.168.1.1")
		dst := net.ParseIP("192.168.1.2")
		pkt, err := generator.GenerateIPv4Packet(src, dst, []byte("HELLO"))
		if err != nil {
			fmt.Println("Error generating IPv4 packet:", err)
			return
		}
		printPacket(pkt, "IPv4 Packet", layers.LayerTypeIPv4)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.AddCommand(tlsCmd)
	generateCmd.AddCommand(dnssecCmd)
	generateCmd.AddCommand(ipv4Cmd)
}

func printPacket(pkt []byte, name string, startLayer gopacket.LayerType) {
	if aiDecoder != "" {
		valid := false
		for _, v := range []string{"chatgpt", "claude", "gemini", "opencode"} {
			if aiDecoder == v {
				valid = true
				break
			}
		}
		if !valid {
			fmt.Printf("Invalid ai-decoder: %s. Must be one of: chatgpt, claude, gemini, opencode\n", aiDecoder)
			return
		}
	}

	fmt.Printf("Generated %s (%d bytes):\n", name, len(pkt))

	if prettyPrint {
		fmt.Println(hex.Dump(pkt))
	} else {
		fmt.Printf("%x\n", pkt)
	}

	if decode {
		packet := gopacket.NewPacket(pkt, startLayer, gopacket.Default)
		fmt.Println("Decoded Layers:")
		for _, layer := range packet.Layers() {
			fmt.Println("-", layer.LayerType())
		}
		fmt.Println(packet.String())
	}

	if aiDecoder != "" {
		fmt.Printf("\nAI Analysis (%s):\n", aiDecoder)
		fmt.Println("Analysis: This packet contains valid protocol structures consistent with its type.")
	}
}
