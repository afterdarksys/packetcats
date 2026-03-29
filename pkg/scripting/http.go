package scripting

import (
	"bytes"
	"io"
	"net/http"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// HTTPModule returns the "http" Starlark module
func HTTPModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "http",
		Members: starlark.StringDict{
			"get":  starlark.NewBuiltin("get", httpGet),
			"post": starlark.NewBuiltin("post", httpPost),
		},
	}
}

func doReq(req *http.Request) (starlark.Value, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Build headers dict
	headersDict := starlark.NewDict(len(resp.Header))
	for k, v := range resp.Header {
		if len(v) > 0 {
			headersDict.SetKey(starlark.String(k), starlark.String(v[0]))
		}
	}

	return starlarkstruct.FromStringDict(starlark.String("response"), starlark.StringDict{
		"status":      starlark.MakeInt(resp.StatusCode),
		"body":        starlark.String(string(body)),
		"body_bytes":  starlark.Bytes(body),
		"headers":     headersDict,
	}), nil
}

func parseHeaders(hdrs *starlark.Dict, req *http.Request) error {
	if hdrs == nil {
		return nil
	}
	for _, k := range hdrs.Keys() {
		v, _, _ := hdrs.Get(k)
		keyStr, ok1 := k.(starlark.String)
		valStr, ok2 := v.(starlark.String)
		if ok1 && ok2 {
			req.Header.Set(string(keyStr), string(valStr))
		}
	}
	return nil
}

func httpGet(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var url string
	var headers *starlark.Dict
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "url", &url, "headers?", &headers); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	parseHeaders(headers, req)

	return doReq(req)
}

func httpPost(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var url string
	var data string
	var headers *starlark.Dict
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "url", &url, "data", &data, "headers?", &headers); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, err
	}
	parseHeaders(headers, req)

	return doReq(req)
}
