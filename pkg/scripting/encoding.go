package scripting

import (
	"encoding/base64"
	"encoding/json"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Base64Module returns the "base64" Starlark module
func Base64Module() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "base64",
		Members: starlark.StringDict{
			"encode": starlark.NewBuiltin("encode", base64Encode),
			"decode": starlark.NewBuiltin("decode", base64Decode),
		},
	}
}

// JSONModule returns the "json" Starlark module
func JSONModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "json",
		Members: starlark.StringDict{
			"encode": starlark.NewBuiltin("encode", jsonEncode),
			"decode": starlark.NewBuiltin("decode", jsonDecode),
		},
	}
}

func base64Encode(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var data string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "data", &data); err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(data))
	return starlark.String(encoded), nil
}

func base64Decode(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var encoded string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "data", &encoded); err != nil {
		return nil, err
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	return starlark.String(string(decoded)), nil
}

// Very basic JSON encode (Dict -> JSON String)
func jsonEncode(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var val starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "data", &val); err != nil {
		return nil, err
	}

	converted := starlarkToInterface(val)
	bytes, err := json.Marshal(converted)
	if err != nil {
		return nil, err
	}
	return starlark.String(string(bytes)), nil
}

// Very basic JSON decode (JSON String -> Dict/List)
func jsonDecode(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var data string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "data", &data); err != nil {
		return nil, err
	}

	var parsed interface{}
	if err := json.Unmarshal([]byte(data), &parsed); err != nil {
		return nil, err
	}
	return interfaceToStarlark(parsed), nil
}

func starlarkToInterface(v starlark.Value) interface{} {
	switch v := v.(type) {
	case starlark.String:
		return string(v)
	case starlark.Int:
		i, _ := v.Int64()
		return i
	case starlark.Float:
		return float64(v)
	case starlark.Bool:
		return bool(v)
	case *starlark.List:
		var result []interface{}
		for i := 0; i < v.Len(); i++ {
			result = append(result, starlarkToInterface(v.Index(i)))
		}
		return result
	case *starlark.Dict:
		result := make(map[string]interface{})
		for _, k := range v.Keys() {
			val, _, _ := v.Get(k)
			keyStr, ok := k.(starlark.String)
			if ok {
				result[string(keyStr)] = starlarkToInterface(val)
			}
		}
		return result
	}
	return nil
}

func interfaceToStarlark(v interface{}) starlark.Value {
	switch v := v.(type) {
	case string:
		return starlark.String(v)
	case float64:
		// JSON numbers decode to float64, check if it's an integer
		if v == float64(int64(v)) {
			return starlark.MakeInt64(int64(v))
		}
		return starlark.Float(v)
	case bool:
		return starlark.Bool(v)
	case []interface{}:
		list := starlark.NewList(nil)
		for _, item := range v {
			list.Append(interfaceToStarlark(item))
		}
		return list
	case map[string]interface{}:
		dict := starlark.NewDict(len(v))
		for k, val := range v {
			dict.SetKey(starlark.String(k), interfaceToStarlark(val))
		}
		return dict
	}
	return starlark.None
}
