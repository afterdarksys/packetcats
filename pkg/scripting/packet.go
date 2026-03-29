package scripting

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// PacketModule returns the "packet" Starlark module
func PacketModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "packet",
		Members: starlark.StringDict{
			"new_eth":       starlark.NewBuiltin("new_eth", packetNewEth),
			"new_ipv4":      starlark.NewBuiltin("new_ipv4", packetNewIPv4),
			"new_tcp":       starlark.NewBuiltin("new_tcp", packetNewTCP),
			"new_udp":       starlark.NewBuiltin("new_udp", packetNewUDP),
			"new_icmp_echo": starlark.NewBuiltin("new_icmp_echo", packetNewICMPEcho),
			"new_payload":   starlark.NewBuiltin("new_payload", packetNewPayload),
			"assemble":      starlark.NewBuiltin("assemble", packetAssemble),
			"send":          starlark.NewBuiltin("send", packetSend),
		},
	}
}

// LayerWrapper implements starlark.Value so we can pass it around in Starlark.
type LayerWrapper struct {
	Layer gopacket.SerializableLayer
	Name  string
}

func (l *LayerWrapper) String() string        { return fmt.Sprintf("<layer %s>", l.Name) }
func (l *LayerWrapper) Type() string          { return "layer" }
func (l *LayerWrapper) Freeze()               {}
func (l *LayerWrapper) Truth() starlark.Bool  { return starlark.True }
func (l *LayerWrapper) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: layer") }

func packetNewEth(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var srcMAC, dstMAC string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "src", &srcMAC, "dst", &dstMAC); err != nil {
		return nil, err
	}

	src, err := net.ParseMAC(srcMAC)
	if err != nil {
		return nil, fmt.Errorf("invalid src mac %s: %v", srcMAC, err)
	}
	dst, err := net.ParseMAC(dstMAC)
	if err != nil {
		return nil, fmt.Errorf("invalid dst mac %s: %v", dstMAC, err)
	}

	eth := &layers.Ethernet{
		SrcMAC:       src,
		DstMAC:       dst,
		EthernetType: layers.EthernetTypeIPv4,
	}

	return &LayerWrapper{Layer: eth, Name: "Ethernet"}, nil
}

func packetNewIPv4(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var srcIP, dstIP string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "src", &srcIP, "dst", &dstIP); err != nil {
		return nil, err
	}

	src := net.ParseIP(srcIP)
	dst := net.ParseIP(dstIP)
	if src == nil || dst == nil {
		return nil, fmt.Errorf("invalid IP addresses")
	}

	ip := &layers.IPv4{
		SrcIP:    src.To4(),
		DstIP:    dst.To4(),
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolTCP, // Will be overridden in assemble if followed by UDP/ICMP
	}

	return &LayerWrapper{Layer: ip, Name: "IPv4"}, nil
}

func packetNewTCP(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var srcPort, dstPort int
	var flags *starlark.Dict
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "src_port", &srcPort, "dst_port", &dstPort, "flags?", &flags); err != nil {
		return nil, err
	}

	tcp := &layers.TCP{
		SrcPort: layers.TCPPort(srcPort),
		DstPort: layers.TCPPort(dstPort),
		Window:  14600,
	}

	if flags != nil {
		for _, k := range flags.Keys() {
			val, _, _ := flags.Get(k)
			keyStr, _ := k.(starlark.String)
			isSet := bool(val.(starlark.Bool))
			switch string(keyStr) {
			case "syn":
				tcp.SYN = isSet
			case "ack":
				tcp.ACK = isSet
			case "fin":
				tcp.FIN = isSet
			case "rst":
				tcp.RST = isSet
			case "psh":
				tcp.PSH = isSet
			}
		}
	}

	return &LayerWrapper{Layer: tcp, Name: "TCP"}, nil
}

func packetNewUDP(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var srcPort, dstPort int
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "src_port", &srcPort, "dst_port", &dstPort); err != nil {
		return nil, err
	}
	udp := &layers.UDP{
		SrcPort: layers.UDPPort(srcPort),
		DstPort: layers.UDPPort(dstPort),
	}
	return &LayerWrapper{Layer: udp, Name: "UDP"}, nil
}

func packetNewICMPEcho(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var id, seq int
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "id", &id, "seq", &seq); err != nil {
		return nil, err
	}
	icmp := &layers.ICMPv4{
		TypeCode: layers.ICMPv4TypeEchoRequest,
		Id:       uint16(id),
		Seq:      uint16(seq),
	}
	return &LayerWrapper{Layer: icmp, Name: "ICMPv4"}, nil
}

func packetNewPayload(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var data string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "data", &data); err != nil {
		return nil, err
	}
	payload := gopacket.Payload([]byte(data))
	return &LayerWrapper{Layer: &payload, Name: "Payload"}, nil
}

func packetAssemble(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var layerList *starlark.List
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "layers", &layerList); err != nil {
		return nil, err
	}

	var serializableLayers []gopacket.SerializableLayer
	var networkLayer gopacket.NetworkLayer

	for i := 0; i < layerList.Len(); i++ {
		val := layerList.Index(i)
		wrapper, ok := val.(*LayerWrapper)
		if !ok {
			return nil, fmt.Errorf("list item %d is not a valid layer object", i)
		}
		
		// If we see IPv4, remember it to link to TCP/UDP for checksumming
		if ip, ok := wrapper.Layer.(*layers.IPv4); ok {
			networkLayer = ip
		}
		
		// Map up Protocol on IPv4 if we see Transport layers
		if networkLayer != nil {
			if ip, ok := networkLayer.(*layers.IPv4); ok {
				if _, ok := wrapper.Layer.(*layers.TCP); ok {
					ip.Protocol = layers.IPProtocolTCP
					wrapper.Layer.(*layers.TCP).SetNetworkLayerForChecksum(ip)
				} else if _, ok := wrapper.Layer.(*layers.UDP); ok {
					ip.Protocol = layers.IPProtocolUDP
					wrapper.Layer.(*layers.UDP).SetNetworkLayerForChecksum(ip)
				} else if _, ok := wrapper.Layer.(*layers.ICMPv4); ok {
					ip.Protocol = layers.IPProtocolICMPv4
				}
			}
		}

		serializableLayers = append(serializableLayers, wrapper.Layer)
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	err := gopacket.SerializeLayers(buf, opts, serializableLayers...)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize packet: %v", err)
	}

	return starlark.Bytes(buf.Bytes()), nil
}

func packetSend(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var iface string
	var raw starlark.Bytes
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "interface", &iface, "raw", &raw); err != nil {
		return nil, err
	}

	handle, err := pcap.OpenLive(iface, 65536, false, pcap.BlockForever)
	if err != nil {
		return nil, fmt.Errorf("error opening pcap on interface %s: %v", iface, err)
	}
	defer handle.Close()

	if err := handle.WritePacketData([]byte(raw)); err != nil {
		return nil, fmt.Errorf("error writing packet: %v", err)
	}

	return starlark.True, nil
}
