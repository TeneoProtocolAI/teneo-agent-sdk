package modules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ForwardToOpenAI sends prompt to Google Gemini (preferred) or OpenAI (fallback).
// Important fix: always normalize GOOGLE_MODEL by STRIPPING leading "models/" if present,
// then build endpoint: /v1beta/models/{modelName}:generateContent
func ForwardToOpenAI(prompt string) (string, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return "", fmt.Errorf("empty prompt")
	}

	googleKey := strings.TrimSpace(os.Getenv("GOOGLE_API_KEY"))
	openaiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))

	shortKey := func(k string) string {
		if k == "" {
			return ""
		}
		if len(k) <= 8 {
			return k
		}
		return k[:8] + "..."
	}

	// Prefer Google Gemini if key present
	if googleKey != "" {
		modelEnv := strings.TrimSpace(os.Getenv("GOOGLE_MODEL"))
		if modelEnv == "" {
			modelEnv = "gemini-2.5-flash"
		}

		// **NORMALIZE**: strip any leading "models/" if present
		if strings.HasPrefix(modelEnv, "models/") {
			modelEnv = strings.TrimPrefix(modelEnv, "models/")
		}

		// Now modelEnv is plain model name (e.g. "gemini-2.5-flash")
		// Construct endpoint: /v1beta/models/{modelEnv}:generateContent
		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", modelEnv, googleKey)

		// Build request body per Gemini docs
		reqBody := map[string]interface{}{
			"contents": []interface{}{
				map[string]interface{}{
					"parts": []interface{}{
						map[string]interface{}{"text": prompt},
					},
				},
			},
			"generationConfig": map[string]interface{}{
				"maxOutputTokens": 256,
				"temperature":     0.2,
			},
		}
		b, _ := json.Marshal(reqBody)

		log.Printf("ForwardToOpenAI: Google request -> model=%s key_preview=%s prompt_len=%d",
			modelEnv, shortKey(googleKey), len(prompt))

		req, err := http.NewRequest("POST", url, bytes.NewReader(b))
		if err != nil {
			return "", fmt.Errorf("failed build request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-goog-api-key", googleKey)

		client := &http.Client{Timeout: 25 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("google http err: %w", err)
		}
		defer resp.Body.Close()
		respBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 200*1024))

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			log.Printf("ForwardToOpenAI: Google response status=%d body_preview=%s", resp.StatusCode, sanitizeForLog(string(respBytes)))
			return "", fmt.Errorf("google api error: status %d: %s", resp.StatusCode, sanitizeForLog(string(respBytes)))
		}

		// parse response and extract candidate text
		var parsed map[string]interface{}
		if err := json.Unmarshal(respBytes, &parsed); err != nil {
			log.Printf("ForwardToOpenAI: google parse json err: %v", err)
			return strings.TrimSpace(string(respBytes)), nil
		}

		// Typical path: candidates[0].content.parts[0].text
		if cands, ok := parsed["candidates"].([]interface{}); ok && len(cands) > 0 {
			if cand0, ok := cands[0].(map[string]interface{}); ok {
				if content, ok := cand0["content"].(map[string]interface{}); ok {
					if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
						if p0, ok := parts[0].(map[string]interface{}); ok {
							if txt, ok := p0["text"].(string); ok && txt != "" {
								return strings.TrimSpace(txt), nil
							}
						}
					}
				}
				if txt, ok := cand0["text"].(string); ok && txt != "" {
					return strings.TrimSpace(txt), nil
				}
			}
		}

		// fallback path: output.content.parts[0].text
		if out, ok := parsed["output"].(map[string]interface{}); ok {
			if content, ok := out["content"].(map[string]interface{}); ok {
				if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
					if p0, ok := parts[0].(map[string]interface{}); ok {
						if txt, ok := p0["text"].(string); ok && txt != "" {
							return strings.TrimSpace(txt), nil
						}
					}
				}
			}
		}

		// final fallback: first string leaf
		if s := findFirstString(parsed); s != "" {
			return strings.TrimSpace(s), nil
		}
		return strings.TrimSpace(string(respBytes)), nil
	}

	// fallback: OpenAI
	if openaiKey != "" {
		reqURL := "https://api.openai.com/v1/chat/completions"
		reqBodyMap := map[string]interface{}{
			"model": "gpt-4o-mini",
			"messages": []map[string]interface{}{
				{"role": "user", "content": prompt},
			},
			"max_tokens": 256,
			"temperature": 0.2,
		}
		reqB, _ := json.Marshal(reqBodyMap)
		req, _ := http.NewRequest("POST", reqURL, bytes.NewReader(reqB))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+openaiKey)

		client := &http.Client{Timeout: 20 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("openai http err: %w", err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 200*1024))
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			log.Printf("ForwardToOpenAI: OpenAI response status=%d body_preview=%s", resp.StatusCode, sanitizeForLog(string(b)))
			return "", fmt.Errorf("openai api error: status %d: %s", resp.StatusCode, sanitizeForLog(string(b)))
		}
		var parsed map[string]interface{}
		if err := json.Unmarshal(b, &parsed); err != nil {
			return strings.TrimSpace(string(b)), nil
		}
		if choices, ok := parsed["choices"].([]interface{}); ok && len(choices) > 0 {
			if ch0, ok := choices[0].(map[string]interface{}); ok {
				if msg, ok := ch0["message"].(map[string]interface{}); ok {
					if content, ok := msg["content"].(string); ok {
						return strings.TrimSpace(content), nil
					}
				}
				if txt, ok := ch0["text"].(string); ok {
					return strings.TrimSpace(txt), nil
				}
			}
		}
		return strings.TrimSpace(string(b)), nil
	}

	return "", fmt.Errorf("no AI API key configured (set GOOGLE_API_KEY or OPENAI_API_KEY)")
}

// helper functions
func sanitizeForLog(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 800 {
		return s
	}
	return s[:800] + "...[truncated]"
}

func findFirstString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []interface{}:
		for _, e := range t {
			if s := findFirstString(e); s != "" {
				return s
			}
		}
	case map[string]interface{}:
		for _, val := range t {
			if s := findFirstString(val); s != "" {
				return s
			}
		}
	}
	return ""
}
