package scripting

import (
	"fmt"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"go.starlark.net/starlark"
)

// RunStream executes a Starlark script and checks for a `packet_hook` function.
// It then reads a pcap format stream from STDIN and calls the hook on each packet.
func RunStream(scriptPath string) error {
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script not found: %s", scriptPath)
	}

	thread := &starlark.Thread{
		Name:  "packetcats-stream",
		Print: printBuiltin,
	}

	predeclared := starlark.StringDict{
		"net":      NetModule(),
		"packet":   PacketModule(),
		"base64":   Base64Module(),
		"json":     JSONModule(),
		"http":     HTTPModule(),
		"dns":      DNSModule(),
		"smtp":     MailModule(),
		"tls":      TLSModule(),
		"sip":      SIPModule(),
		"rtp":      RTPModule(),
		"pcap":     PCAPModule(),
		"fuzz":     FuzzModule(),
		"tunnel":   TunnelModule(),
		"ai":       AIModule(),
		"tcpstack": TCPStackModule(),
	}

	globals, err := starlark.ExecFile(thread, scriptPath, nil, predeclared)
	if err != nil {
		if evalErr, ok := err.(*starlark.EvalError); ok {
			return fmt.Errorf("script execution error:\n%s", evalErr.Backtrace())
		}
		return fmt.Errorf("script error: %w", err)
	}

	hookVal, ok := globals["packet_hook"]
	if !ok {
		return fmt.Errorf("script %s must define a 'packet_hook(raw)' function for decoding", scriptPath)
	}

	hookFn, ok := hookVal.(starlark.Callable)
	if !ok {
		return fmt.Errorf("'packet_hook' must be a function")
	}

	handle, err := pcap.OpenOfflineFile(os.Stdin)
	if err != nil {
		return fmt.Errorf("error opening stdin pcap (are you piping a valid pcap stream?): %v", err)
	}
	defer handle.Close()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		data := packet.Data()
		_, err := starlark.Call(thread, hookFn, starlark.Tuple{starlark.Bytes(data)}, nil)
		if err != nil {
			if evalErr, ok := err.(*starlark.EvalError); ok {
				fmt.Fprintf(os.Stderr, "packet_hook error: %s\n", evalErr.Backtrace())
			} else {
				fmt.Fprintf(os.Stderr, "packet_hook error: %v\n", err)
			}
		}
	}

	return nil
}
