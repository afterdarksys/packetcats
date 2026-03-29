package scripting

import (
	"fmt"
	"os"

	"go.starlark.net/starlark"
)

// printBuiltin allows scripts to use `print()`
func printBuiltin(thread *starlark.Thread, msg string) {
	fmt.Println(msg)
}

// RunScript executes a given Starlark script with the packetcats environment
func RunScript(scriptPath string) error {
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script not found: %s", scriptPath)
	}

	thread := &starlark.Thread{
		Name:  "packetcats",
		Print: printBuiltin,
	}

	// Predeclared globals
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

	_, err := starlark.ExecFile(thread, scriptPath, nil, predeclared)
	if err != nil {
		if evalErr, ok := err.(*starlark.EvalError); ok {
			return fmt.Errorf("script execution error:\n%s", evalErr.Backtrace())
		}
		return fmt.Errorf("script error: %w", err)
	}
	return nil
}
