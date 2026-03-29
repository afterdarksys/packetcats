package scripting

import (
	"crypto/rand"
	"math/big"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// FuzzModule returns the "fuzz" Starlark module
func FuzzModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "fuzz",
		Members: starlark.StringDict{
			"mutate": starlark.NewBuiltin("mutate", fuzzMutate),
		},
	}
}

func fuzzMutate(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var raw starlark.Bytes
	var intensity float64 = 0.1
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "raw", &raw, "intensity?", &intensity); err != nil {
		return nil, err
	}

	data := []byte(string(raw))
	if len(data) == 0 {
		return raw, nil
	}

	for i := 0; i < len(data); i++ {
		// Roll a random float between 0.0 and 1.0
		num, _ := rand.Int(rand.Reader, big.NewInt(1000))
		roll := float64(num.Int64()) / 1000.0
		
		if roll < intensity {
			// Flip a random bit
			bit, _ := rand.Int(rand.Reader, big.NewInt(8))
			data[i] ^= (1 << bit.Int64())
		}
	}

	return starlark.Bytes(data), nil
}
