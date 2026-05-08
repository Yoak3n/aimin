package tool

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Yoak3n/aimin/blood/config"
)

const ttsEndpoint = "/chat/completions"
const ttsTimeout = 2 * time.Minute

func TTS(ctx *Context) string {
	raw := strings.TrimSpace(ctx.GetPayload())
	if raw == "" {
		return "ERROR: args is empty"
	}
	args := parseArgs(raw)

	text := strings.TrimSpace(firstNonEmpty(args["text"], args["_0"]))
	if text == "" {
		return "ERROR: missing text"
	}

	instruction := strings.TrimSpace(firstNonEmpty(args["instruction"], args["style"], args["_1"]))

	cfg := config.GlobalConfiguration().TTS
	if cfg == nil {
		cfg = config.DefaultTTSConfig()
	}

	apiKey := strings.TrimSpace(firstNonEmpty(args["api_key"], cfg.APIKey))
	if apiKey == "" {
		return "ERROR: TTS api_key is not configured"
	}

	apiURL := strings.TrimSpace(firstNonEmpty(args["api_url"], cfg.APIUrl))
	if apiURL == "" {
		apiURL = "https://api.xiaomimimo.com/v1"
	}

	model := strings.TrimSpace(firstNonEmpty(args["model"], cfg.Model))
	if model == "" {
		model = "mimo-v2.5-tts"
	}

	voice := strings.TrimSpace(firstNonEmpty(args["voice"], cfg.Voice))
	if voice == "" {
		voice = "mimo_default"
	}

	format := strings.TrimSpace(firstNonEmpty(args["format"]))
	if format == "" {
		format = "wav"
	}

	messages := buildTTSMessages(text, instruction)
	body := map[string]any{
		"model":    model,
		"messages": messages,
		"audio": map[string]string{
			"format": format,
			"voice":  voice,
		},
		"stream": false,
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return "ERROR: marshal request failed: " + err.Error()
	}

	endpoint := strings.TrimRight(apiURL, "/") + ttsEndpoint
	req, err := http.NewRequestWithContext(ctx.Ctx, http.MethodPost, endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return "ERROR: create request failed: " + err.Error()
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", apiKey)

	client := &http.Client{Timeout: ttsTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "ERROR: request failed: " + err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Sprintf("ERROR: TTS API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "ERROR: read response failed: " + err.Error()
	}

	var result struct {
		Choices []struct {
			Message struct {
				Audio struct {
					Data string `json:"data"`
				} `json:"audio"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "ERROR: parse response failed: " + err.Error()
	}
	if len(result.Choices) == 0 {
		return "ERROR: TTS response has no choices"
	}

	audioData := result.Choices[0].Message.Audio.Data
	if audioData == "" {
		return "ERROR: TTS response has no audio data"
	}

	decoded, err := base64.StdEncoding.DecodeString(audioData)
	if err != nil {
		return "ERROR: decode audio data failed: " + err.Error()
	}

	if ctx.OnAudio != nil {
		ctx.OnAudio(format, voice, base64.StdEncoding.EncodeToString(decoded), len(decoded))
	}

	out := map[string]any{
		"status":      "ok",
		"format":      format,
		"voice":       voice,
		"model":       model,
		"text":        text,
		"audio_bytes": len(decoded),
	}
	resultJSON, _ := json.Marshal(out)
	return string(resultJSON)
}

func buildTTSMessages(text, instruction string) []map[string]string {
	messages := make([]map[string]string, 0, 2)
	if instruction != "" {
		messages = append(messages, map[string]string{
			"role":    "user",
			"content": instruction,
		})
	}
	messages = append(messages, map[string]string{
		"role":    "assistant",
		"content": text,
	})
	return messages
}
