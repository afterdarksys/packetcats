package scripting

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// TLSModule returns the "tls" Starlark module
func TLSModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "tls",
		Members: starlark.StringDict{
			"hello": starlark.NewBuiltin("hello", tlsHello),
		},
	}
}

func tlsHello(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var host string
	var port int = 443
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "host", &host, "port?", &port); err != nil {
		return nil, err
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", addr, &tls.Config{
		InsecureSkipVerify: true, // test mode
	})

	if err != nil {
		return nil, err
	}
	defer conn.Close()

	state := conn.ConnectionState()
	
	certs := starlark.NewList(nil)
	for _, cert := range state.PeerCertificates {
		certInfo := starlarkstruct.FromStringDict(starlark.String("certificate"), starlark.StringDict{
			"subject": starlark.String(cert.Subject.String()),
			"issuer":  starlark.String(cert.Issuer.String()),
			"dns_names": convertStringList(cert.DNSNames),
		})
		certs.Append(certInfo)
	}

	return starlarkstruct.FromStringDict(starlark.String("tls_session"), starlark.StringDict{
		"version":      starlark.MakeInt(int(state.Version)),
		"cipher_suite": starlark.MakeInt(int(state.CipherSuite)),
		"certificates": certs,
	}), nil
}

func convertStringList(strs []string) *starlark.List {
	l := starlark.NewList(nil)
	for _, s := range strs {
		l.Append(starlark.String(s))
	}
	return l
}
