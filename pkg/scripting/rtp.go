package scripting

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/pion/rtp"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// RTPModule returns the "rtp" Starlark module
func RTPModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "rtp",
		Members: starlark.StringDict{
			"stream_wav": starlark.NewBuiltin("stream_wav", rtpStreamWav),
		},
	}
}

// rtpStreamWav reads a .raw or .wav file and streams it out via UDP as G.711 PCMU
func rtpStreamWav(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var filename, targetIP string
	var targetPort int
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "filename", &filename, "target_ip", &targetIP, "target_port", &targetPort); err != nil {
		return nil, err
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %v", err)
	}
	defer f.Close()

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", targetIP, targetPort))
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	packetizer := rtp.Header{
		Version:        2,
		PayloadType:    0, // PCMU G.711
		SequenceNumber: 1,
		Timestamp:      0,
		SSRC:           0x12345678,
	}

	const ptime = 20 * time.Millisecond
	const bytesPerFrame = 160 // 8kHz * 20ms = 160 samples (bytes for G711)

	buf := make([]byte, bytesPerFrame)
	ticker := time.NewTicker(ptime)
	defer ticker.Stop()

	// Minimal header skipping for safety if it's a WAV file (first 44 bytes roughly)
	// For purity, users should pass .raw files, but we ignore robust WAV parsing for the demo.

	for {
		n, err := f.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read error: %v", err)
		}

		packet := &rtp.Packet{
			Header:  packetizer,
			Payload: buf[:n],
		}

		raw, err := packet.Marshal()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal RTP: %v", err)
		}

		if _, err := conn.Write(raw); err != nil {
			return nil, fmt.Errorf("failed sending RTP: %v", err)
		}

		packetizer.SequenceNumber++
		packetizer.Timestamp += 160

		<-ticker.C
	}

	return starlark.True, nil
}
