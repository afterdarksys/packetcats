package scripting

import (
	"fmt"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// SIPModule returns the "sip" Starlark module
func SIPModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "sip",
		Members: starlark.StringDict{
			"invite":   starlark.NewBuiltin("invite", sipInvite),
			"register": starlark.NewBuiltin("register", sipRegister),
		},
	}
}

func sipInvite(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var requestURI, to, from, callID, sdp string
	var cseq int = 1
	var localIP, localPort string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, 
		"request_uri", &requestURI, "to", &to, "from", &from, "call_id", &callID, 
		"sdp?", &sdp, "cseq?", &cseq, "local_ip?", &localIP, "local_port?", &localPort); err != nil {
		return nil, err
	}

	if localIP == "" {
		localIP = "127.0.0.1"
	}
	if localPort == "" {
		localPort = "5060"
	}

	via := fmt.Sprintf("SIP/2.0/UDP %s:%s;branch=z9hG4bK-%s", localIP, localPort, callID)
	contact := fmt.Sprintf("<sip:%s:%s>", localIP, localPort)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("INVITE %s SIP/2.0\r\n", requestURI))
	sb.WriteString(fmt.Sprintf("Via: %s\r\n", via))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", to))
	sb.WriteString(fmt.Sprintf("From: %s;tag=%s\r\n", from, callID)) // using callID as tag for simplicity
	sb.WriteString(fmt.Sprintf("Call-ID: %s\r\n", callID))
	sb.WriteString(fmt.Sprintf("CSeq: %d INVITE\r\n", cseq))
	sb.WriteString(fmt.Sprintf("Contact: %s\r\n", contact))
	sb.WriteString("Max-Forwards: 70\r\n")
	sb.WriteString("User-Agent: PacketCats SIP Engine\r\n")

	if sdp != "" {
		sb.WriteString("Content-Type: application/sdp\r\n")
		sb.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(sdp)))
		sb.WriteString("\r\n")
		sb.WriteString(sdp)
	} else {
		sb.WriteString("Content-Length: 0\r\n\r\n")
	}

	return starlark.String(sb.String()), nil
}

func sipRegister(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var requestURI, to, from, callID string
	var cseq int = 1
	var localIP, localPort string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, 
		"request_uri", &requestURI, "to", &to, "from", &from, "call_id", &callID, 
		"cseq?", &cseq, "local_ip?", &localIP, "local_port?", &localPort); err != nil {
		return nil, err
	}

	if localIP == "" {
		localIP = "127.0.0.1"
	}
	if localPort == "" {
		localPort = "5060"
	}

	via := fmt.Sprintf("SIP/2.0/UDP %s:%s;branch=z9hG4bK-register-%s", localIP, localPort, callID)
	contact := fmt.Sprintf("<sip:%s:%s>", localIP, localPort)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("REGISTER %s SIP/2.0\r\n", requestURI))
	sb.WriteString(fmt.Sprintf("Via: %s\r\n", via))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", to))
	sb.WriteString(fmt.Sprintf("From: %s;tag=%s-reg\r\n", from, callID))
	sb.WriteString(fmt.Sprintf("Call-ID: %s\r\n", callID))
	sb.WriteString(fmt.Sprintf("CSeq: %d REGISTER\r\n", cseq))
	sb.WriteString(fmt.Sprintf("Contact: %s\r\n", contact))
	sb.WriteString("Max-Forwards: 70\r\n")
	sb.WriteString("Expires: 3600\r\n")
	sb.WriteString("User-Agent: PacketCats SIP Engine\r\n")
	sb.WriteString("Content-Length: 0\r\n\r\n")

	return starlark.String(sb.String()), nil
}
