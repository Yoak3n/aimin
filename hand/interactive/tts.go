package interactive

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Yoak3n/aimin/blood/config"
)

const defaultTTSTimeout = 2 * time.Minute

type TTSClient struct {
	apiKey string
	apiURL string
	model  string
	voice  string
	client *http.Client
}

type TTSOption func(*TTSClient)

func WithTTSModel(model string) TTSOption {
	return func(c *TTSClient) { c.model = model }
}

func WithTTSVoice(voice string) TTSOption {
	return func(c *TTSClient) { c.voice = voice }
}

func WithTTSAPIKey(key string) TTSOption {
	return func(c *TTSClient) { c.apiKey = key }
}

func WithTTSAPIURL(url string) TTSOption {
	return func(c *TTSClient) { c.apiURL = url }
}

func NewTTSClient(opts ...TTSOption) *TTSClient {
	cfg := config.GlobalConfiguration().TTS
	if cfg == nil {
		cfg = config.DefaultTTSConfig()
	}
	c := &TTSClient{
		apiKey: cfg.APIKey,
		apiURL: cfg.APIUrl,
		model:  cfg.Model,
		voice:  cfg.Voice,
		client: &http.Client{Timeout: defaultTTSTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.apiURL == "" {
		c.apiURL = "https://api.xiaomimimo.com/v1"
	}
	return c
}

type TTSRequest struct {
	Text        string
	Instruction string
	Model       string
	Voice       string
	Format      string
}

func (c *TTSClient) Synthesize(ctx context.Context, req TTSRequest) ([]byte, error) {
	if req.Text == "" {
		return nil, fmt.Errorf("TTS文本不能为空")
	}
	model := req.Model
	if model == "" {
		model = c.model
	}
	voice := req.Voice
	if voice == "" {
		voice = c.voice
	}
	format := req.Format
	if format == "" {
		format = "wav"
	}

	messages := c.buildMessages(req.Text, req.Instruction)
	body := map[string]any{
		"model":    model,
		"messages": messages,
		"audio": map[string]string{
			"format": format,
			"voice":  voice,
		},
		"stream": false,
	}

	return c.doRequest(ctx, body)
}

type TTSStreamChunk struct {
	AudioData []byte
	Error     error
}

func (c *TTSClient) SynthesizeStream(ctx context.Context, req TTSRequest) (<-chan TTSStreamChunk, error) {
	if req.Text == "" {
		return nil, fmt.Errorf("TTS文本不能为空")
	}
	model := req.Model
	if model == "" {
		model = c.model
	}
	voice := req.Voice
	if voice == "" {
		voice = c.voice
	}

	messages := c.buildMessages(req.Text, req.Instruction)
	body := map[string]any{
		"model":    model,
		"messages": messages,
		"audio": map[string]string{
			"format": "pcm16",
			"voice":  voice,
		},
		"stream": true,
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	apiURL := strings.TrimRight(c.apiURL, "/") + "/chat/completions"
	reqHTTP, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Header.Set("api-key", c.apiKey)
	reqHTTP.Header.Set("Accept", "text/event-stream")

	resp, err := c.client.Do(reqHTTP)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("TTS API 请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	ch := make(chan TTSStreamChunk, 64)
	go func() {
		defer resp.Body.Close()
		defer close(ch)

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			line = strings.TrimSpace(line)
			if line == "" {
				if err != nil {
					if err != io.EOF {
						ch <- TTSStreamChunk{Error: err}
					}
					return
				}
				continue
			}

			if after, ok := strings.CutPrefix(line, "data:"); ok {
				payload := strings.TrimSpace(after)
				if payload == "[DONE]" {
					return
				}

				var chunk struct {
					Choices []struct {
						Delta struct {
							Audio map[string]string `json:"audio"`
						} `json:"delta"`
					} `json:"choices"`
				}
				if jsonErr := json.Unmarshal([]byte(payload), &chunk); jsonErr != nil {
					continue
				}
				if len(chunk.Choices) == 0 {
					continue
				}
				audio := chunk.Choices[0].Delta.Audio
				if audio == nil {
					continue
				}
				data, ok := audio["data"]
				if !ok || data == "" {
					continue
				}
				decoded, decErr := base64.StdEncoding.DecodeString(data)
				if decErr != nil {
					ch <- TTSStreamChunk{Error: fmt.Errorf("解码音频数据失败: %w", decErr)}
					continue
				}
				ch <- TTSStreamChunk{AudioData: decoded}
			}

			if err != nil {
				if err != io.EOF {
					ch <- TTSStreamChunk{Error: err}
				}
				return
			}
		}
	}()

	return ch, nil
}

func (c *TTSClient) buildMessages(text, instruction string) []map[string]string {
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

func (c *TTSClient) doRequest(ctx context.Context, body any) ([]byte, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	apiURL := strings.TrimRight(c.apiURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TTS API 请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
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
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("TTS响应中没有数据")
	}

	audioData := result.Choices[0].Message.Audio.Data
	if audioData == "" {
		return nil, fmt.Errorf("TTS响应中没有音频数据")
	}

	decoded, err := base64.StdEncoding.DecodeString(audioData)
	if err != nil {
		return nil, fmt.Errorf("解码音频数据失败: %w", err)
	}

	return decoded, nil
}
