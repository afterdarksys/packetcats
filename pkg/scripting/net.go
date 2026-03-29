package scripting

import (
	"fmt"
	"net"
	"time"

	"github.com/google/gopacket/routing"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// NetModule returns the "net" Starlark module
func NetModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "net",
		Members: starlark.StringDict{
			"resolve":     starlark.NewBuiltin("resolve", netResolve),
			"my_ip":       starlark.NewBuiltin("my_ip", netMyIP),
			"my_mac":      starlark.NewBuiltin("my_mac", netMyMAC),
			"gateway_ip":  starlark.NewBuiltin("gateway_ip", netGatewayIP),
		},
	}
}

func netResolve(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var hostname string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "hostname", &hostname); err != nil {
		return nil, err
	}

	ips, err := net.LookupIP(hostname)
	if err != nil || len(ips) == 0 {
		return starlark.None, fmt.Errorf("could not resolve %s", hostname)
	}

	// Standardize on IPv4 for now unless specific
	for _, ip := range ips {
		if ip.To4() != nil {
			return starlark.String(ip.String()), nil
		}
	}

	return starlark.String(ips[0].String()), nil
}

// netMyIP attempts to find the system's preferred outbound IP
func netMyIP(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	conn, err := net.DialTimeout("udp", "8.8.8.8:80", 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("unable to determine local IP: %v", err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return starlark.String(localAddr.IP.String()), nil
}

func netMyMAC(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	conn, err := net.DialTimeout("udp", "8.8.8.8:80", 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("unable to determine local interface: %v", err)
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	conn.Close()

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip.Equal(localAddr.IP) {
				return starlark.String(iface.HardwareAddr.String()), nil
			}
		}
	}
	return starlark.None, fmt.Errorf("mac not found for local IP %s", localAddr.IP.String())
}

func netGatewayIP(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	router, err := routing.New()
	if err != nil {
		return nil, fmt.Errorf("failed to init routing: %v", err)
	}
	_, _, gw, err := router.Route(net.ParseIP("8.8.8.8"))
	if err != nil {
		return nil, fmt.Errorf("failed to find gateway route: %v", err)
	}

	if gw == nil {
		return starlark.None, fmt.Errorf("gateway not found")
	}

	return starlark.String(gw.String()), nil
}
