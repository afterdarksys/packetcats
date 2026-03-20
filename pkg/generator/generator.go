package generator

import (
	"encoding/hex"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/miekg/dns"
)

// GenerateIPv4Packet generates a basic IPv4 packet
func GenerateIPv4Packet(srcIP, dstIP net.IP, payload []byte) ([]byte, error) {
	ip := &layers.IPv4{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolTCP, // Default to TCP
	}
	
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	
	err := gopacket.SerializeLayers(buf, opts, ip, gopacket.Payload(payload))
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// GenerateTLSClientHello generates a TCP payload containing a mock TLS Client Hello
// Note: gopacket layers for TLS are mostly for decoding. We construct a simple ClientHello payload here.
func GenerateTLSClientHello() ([]byte, error) {
	// Simple TLS 1.2 Client Hello payload (truncated/simplified for example)
	// Handshake Type: Client Hello (1)
	// Length: ...
	// Version: TLS 1.2 (0x0303)
	// Random: ...
	// Session ID Length: 0
	// Cipher Suites: ...
	// Compression Methods: ...
	
	// Hardcoding a minimal ClientHello for demonstration
	// This is a raw byte construction.
	clientHelloStr := "160303002f0100002b0303" + // Record Header + Handshake Header + Version
		"0000000000000000000000000000000000000000000000000000000000000000" + // Random (32 bytes)
		"00" + // Session ID Length
		"0002002f" + // Cipher Suites Length + Cipher Suite (TLS_RSA_WITH_AES_128_CBC_SHA)
		"0100" // Compression Methods Length + Compression Method (Null)
	
	return hex.DecodeString(clientHelloStr)
}

// GenerateDNSSECResponse generates a DNS response packet with DNSSEC records
func GenerateDNSSECResponse(questionName string) ([]byte, error) {
	// Create DNS message using miekg/dns
	m := new(dns.Msg)
	m.SetReply(&dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id: 0x1234,
			Opcode: dns.OpcodeQuery,
		},
		Question: []dns.Question{{Name: dns.Fqdn(questionName), Qtype: dns.TypeA, Qclass: dns.ClassINET}},
	})
	
	m.Answer = append(m.Answer, &dns.A{
		Hdr: dns.RR_Header{Name: dns.Fqdn(questionName), Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300},
		A:   net.ParseIP("1.2.3.4"),
	})
	
	// Add DNSKEY
	m.Answer = append(m.Answer, &dns.DNSKEY{
		Hdr: dns.RR_Header{Name: dns.Fqdn(questionName), Rrtype: dns.TypeDNSKEY, Class: dns.ClassINET, Ttl: 3600},
		Flags: 256,
		Protocol: 3,
		Algorithm: 5,
		PublicKey: "AwEAAQ==", // Invalid key, just for structure
	})
	
	// Add RRSIG
	m.Answer = append(m.Answer, &dns.RRSIG{
		Hdr: dns.RR_Header{Name: dns.Fqdn(questionName), Rrtype: dns.TypeRRSIG, Class: dns.ClassINET, Ttl: 3600},
		TypeCovered: dns.TypeA,
		Algorithm: 5,
		Labels: 3,
		OrigTtl: 300,
		Expiration: 1234567890,
		Inception: 1234567890,
		KeyTag: 12345,
		SignerName: dns.Fqdn(questionName),
		Signature: "MwEAAQ==", // Invalid signature
	})

	dnsBytes, err := m.Pack()
	if err != nil {
		return nil, err
	}

	eth := &layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
		DstMAC:       net.HardwareAddr{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF},
		EthernetType: layers.EthernetTypeIPv4,
	}
	ip := &layers.IPv4{
		SrcIP:    net.IP{192, 168, 1, 1},
		DstIP:    net.IP{192, 168, 1, 2},
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolUDP,
	}
	udp := &layers.UDP{
		SrcPort: 53,
		DstPort: 12345,
	}
	udp.SetNetworkLayerForChecksum(ip)
	
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	
	err = gopacket.SerializeLayers(buf, opts, eth, ip, udp, gopacket.Payload(dnsBytes))
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}
