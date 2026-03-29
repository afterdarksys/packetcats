package scripting

import (
	"context"
	"net"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// DNSModule returns the "dns" Starlark module
func DNSModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "dns",
		Members: starlark.StringDict{
			"query": starlark.NewBuiltin("query", dnsQuery),
		},
	}
}

func dnsQuery(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var domain string
	var qtype string = "A"
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "domain", &domain, "type?", &qtype); err != nil {
		return nil, err
	}

	resolver := net.DefaultResolver

	answers := starlark.NewList(nil)
	
	switch qtype {
	case "A", "AAAA":
		ips, err := resolver.LookupIPAddr(context.Background(), domain)
		if err == nil {
			for _, ip := range ips {
				answers.Append(starlark.String(ip.IP.String()))
			}
		}
	case "TXT":
		txts, err := resolver.LookupTXT(context.Background(), domain)
		if err == nil {
			for _, t := range txts {
				answers.Append(starlark.String(t))
			}
		}
	case "MX":
		mxs, err := resolver.LookupMX(context.Background(), domain)
		if err == nil {
			for _, m := range mxs {
				answers.Append(starlark.String(m.Host))
			}
		}
	case "CNAME":
		c, err := resolver.LookupCNAME(context.Background(), domain)
		if err == nil {
			answers.Append(starlark.String(c))
		}
	}

	return answers, nil
}
