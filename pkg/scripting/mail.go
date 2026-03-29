package scripting

import (
	"fmt"
	"net/smtp"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// MailModule returns the "smtp" Starlark module
func MailModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "smtp",
		Members: starlark.StringDict{
			"send": starlark.NewBuiltin("send", smtpSend),
		},
	}
}

// simple smtp implementation to test capabilities
func smtpSend(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var host, port, user, pass, to, from, subject, body string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, 
		"host", &host, "port", &port, "user", &user, "pass", &pass, 
		"to", &to, "from", &from, "subject", &subject, "body", &body); err != nil {
		return nil, err
	}

	auth := smtp.PlainAuth("", user, pass, host)
	addr := fmt.Sprintf("%s:%s", host, port)

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", to, from, subject, body))

	err := smtp.SendMail(addr, auth, from, strings.Split(to, ","), msg)
	if err != nil {
		return nil, err
	}

	return starlark.True, nil
}
