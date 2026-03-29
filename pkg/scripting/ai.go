package scripting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// AIModule returns the "ai" Starlark module
func AIModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "ai",
		Members: starlark.StringDict{
			"analyze": starlark.NewBuiltin("analyze", aiAnalyze),
		},
	}
}

func aiAnalyze(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var prompt string
	var raw starlark.Bytes
	var provider string = "gemini"

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "prompt", &prompt, "raw", &raw, "provider?", &provider); err != nil {
		return nil, err
	}

	payloadStr := string(raw)
	if !isPrintable(raw) {
		payloadStr = fmt.Sprintf("%x", []byte(raw))
	}
	fullPrompt := fmt.Sprintf("%s\n\nPacket Payload:\n%s", prompt, payloadStr)

	var answer string
	var err error

	switch provider {
	case "gemini":
		answer, err = askGemini(fullPrompt)
	case "openai":
		answer, err = askOpenAI(fullPrompt)
	case "anthropic":
		answer, err = askAnthropic(fullPrompt)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", provider)
	}

	if err != nil {
		return nil, err
	}

	if answer == "" {
		return starlark.String("AI was unable to fulfill this request."), nil
	}

	return starlark.String(answer), nil
}

func askGemini(prompt string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable is missing")
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=%s", apiKey)
	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
	}
	
	resp, err := doJSONPost(url, nil, reqBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if c, ok := result["candidates"].([]interface{}); ok && len(c) > 0 {
		if cand, ok := c[0].(map[string]interface{}); ok {
			if cnt, ok := cand["content"].(map[string]interface{}); ok {
				if parts, ok := cnt["parts"].([]interface{}); ok && len(parts) > 0 {
					if p, ok := parts[0].(map[string]interface{}); ok {
						if t, ok := p["text"].(string); ok {
							return t, nil
						}
					}
				}
			}
		}
	}
	return "", fmt.Errorf("failed to parse gemini response")
}

func askOpenAI(prompt string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable is missing")
	}

	reqBody := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	
	headers := map[string]string{
		"Authorization": "Bearer " + apiKey,
	}

	resp, err := doJSONPost("https://api.openai.com/v1/chat/completions", headers, reqBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if msg, ok := choice["message"].(map[string]interface{}); ok {
				if t, ok := msg["content"].(string); ok {
					return t, nil
				}
			}
		}
	}
	return "", fmt.Errorf("failed to parse openai response")
}

func askAnthropic(prompt string) (string, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY environment variable is missing")
	}

	reqBody := map[string]interface{}{
		"model": "claude-3-5-sonnet-20241022",
		"max_tokens": 1024,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	
	headers := map[string]string{
		"x-api-key": apiKey,
		"anthropic-version": "2023-06-01",
	}

	resp, err := doJSONPost("https://api.anthropic.com/v1/messages", headers, reqBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if content, ok := result["content"].([]interface{}); ok && len(content) > 0 {
		if item, ok := content[0].(map[string]interface{}); ok {
			if t, ok := item["text"].(string); ok {
				return t, nil
			}
		}
	}
	return "", fmt.Errorf("failed to parse anthropic response")
}

func doJSONPost(url string, headers map[string]string, body map[string]interface{}) (*http.Response, error) {
	reqBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(reqBytes))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(b))
	}
	return resp, nil
}

func isPrintable(data starlark.Bytes) bool {
	for _, b := range []byte(data) {
		if b < 32 || b > 126 {
			if b != '\n' && b != '\r' && b != '\t' {
				return false
			}
		}
	}
	return true
}
