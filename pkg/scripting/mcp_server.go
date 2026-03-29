package scripting

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// StartMCPServer runs an infinite loop reading JSON-RPC calls over Stdin
func StartMCPServer() {
	scanner := bufio.NewScanner(os.Stdin)
	// We might receive massive scripts, ensure scanner buffer is large
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var req rpcRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			sendError(nil, -32700, "Parse error")
			continue
		}

		handleRequest(&req)
	}
}

func handleRequest(req *rpcRequest) {
	switch req.Method {
	case "initialize":
		sendResponse(req.ID, map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "SuperPacketCat",
				"version": "1.0.0",
			},
		})
	case "notifications/initialized":
		// Do nothing, just ack state
	case "tools/list":
		sendResponse(req.ID, map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "execute_script",
					"description": "Executes a SuperPacketCat Starlark script and returns standard output.",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"code": map[string]interface{}{
								"type":        "string",
								"description": "The raw python-like Starlark code to run natively against the PacketCats engine.",
							},
						},
						"required": []string{"code"},
					},
				},
			},
		})
	case "tools/call":
		var params struct {
			Name      string `json:"name"`
			Arguments struct {
				Code string `json:"code"`
			} `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			sendError(req.ID, -32602, "Invalid params")
			return
		}

		if params.Name == "execute_script" {
			out, err := runCodeAndCapture(params.Arguments.Code)

			resText := out
			isError := false
			if err != nil {
				resText += fmt.Sprintf("\n[Execution Error]: %v\n", err)
				isError = true
			}

			sendResponse(req.ID, map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": resText,
					},
				},
				"isError": isError,
			})
		} else {
			sendError(req.ID, -32601, "Tool not found")
		}
	default:
		// Unsupported method, just ignore or return standard error if an ID is present
		if req.ID != nil {
			sendError(req.ID, -32601, "Method not found")
		}
	}
}

func runCodeAndCapture(code string) (string, error) {
	// Write to temp file
	tmpFile, err := os.CreateTemp("", "mcp_script_*.star")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(code); err != nil {
		return "", err
	}
	tmpFile.Close()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run
	err = RunScript(tmpFile.Name())

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String(), err
}

func sendResponse(id interface{}, result interface{}) {
	res := rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	b, _ := json.Marshal(res)
	fmt.Fprintln(os.Stdout, string(b))
}

func sendError(id interface{}, code int, message string) {
	if id == nil {
		return
	}
	res := rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	b, _ := json.Marshal(res)
	fmt.Fprintln(os.Stdout, string(b))
}
