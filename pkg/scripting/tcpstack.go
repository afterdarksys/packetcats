package scripting

import (
	"fmt"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// TCPStackModule returns the "tcpstack" Starlark module
func TCPStackModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "tcpstack",
		Members: starlark.StringDict{
			"listen_syn_ack": starlark.NewBuiltin("listen_syn_ack", tcpStackListen),
		},
	}
}

// listen_syn_ack listens on an interface and a port, automatically responding with SYN-ACK to incoming SYNs
func tcpStackListen(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var iface string
	var listenPort int
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "interface", &iface, "port", &listenPort); err != nil {
		return nil, err
	}

	go func() {
		handle, err := pcap.OpenLive(iface, 65536, true, pcap.BlockForever)
		if err != nil {
			fmt.Printf("tcpstack error opening pcap: %v\n", err)
			return
		}
		defer handle.Close()

		// Filter for TCP SYNs on the specified port
		err = handle.SetBPFFilter(fmt.Sprintf("tcp port %d and tcp[tcpflags] & tcp-syn != 0 and tcp[tcpflags] & tcp-ack == 0", listenPort))
		if err != nil {
			fmt.Printf("tcpstack error setting bpf: %v\n", err)
			return
		}

		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {
			ethLayer := packet.Layer(layers.LayerTypeEthernet)
			ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
			tcpLayer := packet.Layer(layers.LayerTypeTCP)

			if ethLayer == nil || ipv4Layer == nil || tcpLayer == nil {
				continue
			}

			eth := ethLayer.(*layers.Ethernet)
			ip := ipv4Layer.(*layers.IPv4)
			tcp := tcpLayer.(*layers.TCP)

			// Construct SYN-ACK response
			respEth := &layers.Ethernet{
				SrcMAC:       eth.DstMAC,
				DstMAC:       eth.SrcMAC,
				EthernetType: layers.EthernetTypeIPv4,
			}

			respIP := &layers.IPv4{
				SrcIP:    ip.DstIP,
				DstIP:    ip.SrcIP,
				Version:  4,
				TTL:      64,
				Protocol: layers.IPProtocolTCP,
			}

			respTCP := &layers.TCP{
				SrcPort: tcp.DstPort,
				DstPort: tcp.SrcPort,
				Seq:     1000, // random ISN
				Ack:     tcp.Seq + 1,
				SYN:     true,
				ACK:     true,
				Window:  64240,
			}
			respTCP.SetNetworkLayerForChecksum(respIP)

			buf := gopacket.NewSerializeBuffer()
			opts := gopacket.SerializeOptions{
				ComputeChecksums: true,
				FixLengths:       true,
			}

			gopacket.SerializeLayers(buf, opts, respEth, respIP, respTCP)

			time.Sleep(10 * time.Millisecond) // Slight latency
			handle.WritePacketData(buf.Bytes())
			fmt.Printf("[tcpstack] Responded with SYN-ACK to %s\n", ip.SrcIP)
		}
	}()

	return starlark.True, nil
}
