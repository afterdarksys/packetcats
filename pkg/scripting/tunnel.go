package scripting

import (
	"encoding/base32"
	"fmt"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// TunnelModule returns the "tunnel" Starlark module
func TunnelModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "tunnel",
		Members: starlark.StringDict{
			"icmp_encode":    starlark.NewBuiltin("icmp_encode", tunnelIcmpEncode),
			"dns_txt_encode": starlark.NewBuiltin("dns_txt_encode", tunnelDnsEncode),
		},
	}
}

func tunnelIcmpEncode(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var payload string
	var chunkSize int = 64
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "data", &payload, "chunk_size?", &chunkSize); err != nil {
		return nil, err
	}

	res := starlark.NewList(nil)
	data := []byte(payload)
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		res.Append(starlark.Bytes(data[i:end]))
	}

	return res, nil
}

func tunnelDnsEncode(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var domain, payload string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "domain", &domain, "data", &payload); err != nil {
		return nil, err
	}

	// Base32 encode the payload (Base32 is case-insensitive, safe for DNS labels)
	b32 := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(payload))
	b32 = strings.ToLower(b32)

	// DNS labels are max 63 characters
	labelSize := 60
	res := starlark.NewList(nil)

	for i := 0; i < len(b32); i += labelSize {
		end := i + labelSize
		if end > len(b32) {
			end = len(b32)
		}
		
		fqdn := fmt.Sprintf("%s.%s", b32[i:end], domain)
		res.Append(starlark.String(fqdn))
	}

	return res, nil
}
