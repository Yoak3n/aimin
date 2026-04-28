package adapter

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Yoak3n/aimin/blood/pkg/logger"

	"github.com/Yoak3n/aimin/blood/config"
	"github.com/Yoak3n/aimin/blood/schema"
)

type LLMAdapterHub struct {
	Adapters map[string]LLMAdapter
	payload  map[string]float64
	disabled map[string]time.Time
	mutex    sync.RWMutex
}

func (h *LLMAdapterHub) PinAdapter(llmType config.LLMType) (LLMAdapter, string, error) {
	active := h.chatModelByType(llmType)
	return h.getAdapterByTypeWithKey(llmType, active)
}

func (h *LLMAdapterHub) chatModelByType(llmType config.LLMType) string {
	cfg := config.GlobalConfiguration()
	if cfg == nil {
		return ""
	}
	switch llmType {
	case config.LLMTypeChat:
		return cfg.ActiveLLM.ChatModel
	case config.LLMTypeEmbedding:
		return cfg.ActiveLLM.EmbeddingModel
	default:
		return cfg.ActiveLLM.ChatModel
	}
}

func NewLLMAdapterHub() *LLMAdapterHub {
	h := &LLMAdapterHub{
		Adapters: make(map[string]LLMAdapter),
		payload:  make(map[string]float64),
		disabled: make(map[string]time.Time),
		mutex:    sync.RWMutex{},
	}
	h.loadDisabledFromConfig()
	return h
}

func (h *LLMAdapterHub) registerAdapter(llmKey string, adapter LLMAdapter) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.Adapters[llmKey] = adapter
	h.payload[llmKey] = 0
}

func (h *LLMAdapterHub) RegisterAdapter(config *config.LLMConfig) {
	adapter := NewLLMAdapter(config)
	h.registerAdapter(config.Provider+"_"+config.Model, adapter)
	logger.Logger.Infof("注册了适配器: %s", config.Provider+"_"+config.Model)
}

// UnregisterAdapter 移除适配器
func (h *LLMAdapterHub) UnregisterAdapter(key string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	delete(h.Adapters, key)
	delete(h.payload, key)
	delete(h.disabled, key)

	cfg := config.GlobalConfiguration()
	if cfg != nil && cfg.DisabledLLM != nil {
		if _, ok := cfg.DisabledLLM[key]; ok {
			delete(cfg.DisabledLLM, key)
			_ = cfg.Save()
		}
	}
}

func (h *LLMAdapterHub) selectByLoadBalance(candidates []string) string {
	if len(candidates) == 1 {
		return candidates[0]
	}

	// 加权轮询算法：负载越低，权重越高
	var totalWeight float64
	weights := make(map[string]float64)

	for _, key := range candidates {
		load := h.payload[key]
		// 权重 = 1 / (负载 + 0.1)，避免除零
		weight := 1.0 / (load + 0.1)
		weights[key] = weight
		totalWeight += weight
	}

	randomValue := rand.Float64() * totalWeight
	var currentWeight float64
	for _, key := range candidates {
		currentWeight += weights[key]
		if randomValue <= currentWeight {
			return key
		}
	}

	// 兜底返回第一个
	return candidates[0]
}

type disabledAdapterError struct {
	Model string
	Key   string
}

func (e disabledAdapterError) Error() string {
	return fmt.Sprintf("adapter disabled: model=%s key=%s", e.Model, e.Key)
}

func (h *LLMAdapterHub) isDisabledLocked(key string, now time.Time) bool {
	until, ok := h.disabled[key]
	return ok && now.Before(until)
}

func (h *LLMAdapterHub) loadDisabledFromConfig() {
	cfg := config.GlobalConfiguration()
	if cfg == nil || len(cfg.DisabledLLM) == 0 {
		return
	}
	now := time.Now().Unix()
	h.mutex.Lock()
	for key, untilUnix := range cfg.DisabledLLM {
		if untilUnix <= now {
			continue
		}
		h.disabled[key] = time.Unix(untilUnix, 0)
	}
	h.mutex.Unlock()
}

func (h *LLMAdapterHub) cleanupExpiredDisabled(now time.Time) {
	cfg := config.GlobalConfiguration()
	changed := false

	h.mutex.Lock()
	for key, until := range h.disabled {
		if now.Before(until) {
			continue
		}
		delete(h.disabled, key)
		if cfg != nil && cfg.DisabledLLM != nil {
			if _, ok := cfg.DisabledLLM[key]; ok {
				delete(cfg.DisabledLLM, key)
				changed = true
			}
		}
	}
	h.mutex.Unlock()

	if changed && cfg != nil {
		_ = cfg.Save()
	}
}

func (h *LLMAdapterHub) getAdapterByTypeWithKey(llmType config.LLMType, activeModel string) (LLMAdapter, string, error) {
	now := time.Now()
	h.cleanupExpiredDisabled(now)
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	matchType := func(actual config.LLMType) bool {
		if llmType == config.LLMTypeChat {
			return actual == config.LLMTypeChat || actual == config.LLMTypeThink
		}
		return actual == llmType
	}

	if strings.TrimSpace(activeModel) != "" {
		for key, adapter := range h.Adapters {
			if !matchType(config.LLMType(adapter.GetConfig().Type)) {
				continue
			}
			if adapter.GetConfig().Model != activeModel {
				continue
			}
			if h.isDisabledLocked(key, now) {
				return nil, "", disabledAdapterError{Model: activeModel, Key: key}
			}
			return adapter, key, nil
		}
	}

	candidates := make([]string, 0, len(h.Adapters))
	for key, adapter := range h.Adapters {
		if !matchType(config.LLMType(adapter.GetConfig().Type)) {
			continue
		}
		if h.isDisabledLocked(key, now) {
			continue
		}
		candidates = append(candidates, key)
	}

	if len(candidates) == 0 {
		return nil, "", fmt.Errorf("没有找到类型为 %s 的可用适配器", llmType)
	}

	selectedKey := h.selectByLoadBalance(candidates)
	logger.Logger.Infof("选择的适配器: %s", selectedKey)
	return h.Adapters[selectedKey], selectedKey, nil
}

func (h *LLMAdapterHub) disableAdapter(key string, d time.Duration) {
	if strings.TrimSpace(key) == "" {
		return
	}
	if d <= 0 {
		d = 10 * time.Minute
	}
	until := time.Now().Add(d)
	h.mutex.Lock()
	h.disabled[key] = until
	h.mutex.Unlock()

	cfg := config.GlobalConfiguration()
	if cfg != nil {
		if cfg.DisabledLLM == nil {
			cfg.DisabledLLM = map[string]int64{}
		}
		cfg.DisabledLLM[key] = until.Unix()
		_ = cfg.Save()
	}
}

func (h *LLMAdapterHub) clearActiveModel(llmType config.LLMType, model string) {
	model = strings.TrimSpace(model)
	if model == "" {
		return
	}
	cfg := config.GlobalConfiguration()
	changed := false
	switch llmType {
	case config.LLMTypeChat:
		if cfg.ActiveLLM.ChatModel == model {
			cfg.ActiveLLM.ChatModel = ""
			changed = true
		}
	case config.LLMTypeEmbedding:
		if cfg.ActiveLLM.EmbeddingModel == model {
			cfg.ActiveLLM.EmbeddingModel = ""
			changed = true
		}
	}
	if changed {
		_ = cfg.Save()
	}
}

var statusCodeRe = regexp.MustCompile(`状态码:\s*(\d+)`)

func classifyLLMFailure(err error) (bool, time.Duration) {
	if err == nil {
		return false, 0
	}
	logger.Logger.Errorf("LLM error: %v", err)
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "context_length") || strings.Contains(msg, "max tokens") || strings.Contains(msg, "maximum context") {
		return false, 0
	}

	status := 0
	if m := statusCodeRe.FindStringSubmatch(msg); len(m) == 2 {
		if n, convErr := strconv.Atoi(m[1]); convErr == nil {
			status = n
		}
	}

	isQuota := strings.Contains(msg, "insufficient_quota") || strings.Contains(msg, "quota") || strings.Contains(msg, "额度") || strings.Contains(msg, "余额")
	isRate := strings.Contains(msg, "rate limit") || strings.Contains(msg, "too many requests") || status == 429
	isAuth := status == 401 || status == 403
	isPay := status == 402

	if isQuota || isAuth || isPay {
		return true, 24 * time.Hour
	}
	if isRate {
		return true, 10 * time.Minute
	}
	return false, 0
}

func (h *LLMAdapterHub) Chat(userMessages []schema.OpenAIMessage, systemPrompt string) (string, error) {
	for range 2 {
		active := h.chatModelByType(config.LLMTypeChat)
		adapter, key, err := h.getAdapterByTypeWithKey(config.LLMTypeChat, active)
		if err != nil {
			var de disabledAdapterError
			if errors.As(err, &de) {
				h.clearActiveModel(config.LLMTypeChat, de.Model)
				continue
			}
			return "", err
		}

		resp, callErr := adapter.Chat(userMessages, systemPrompt)
		if disable, d := classifyLLMFailure(callErr); disable {
			h.disableAdapter(key, d)
			if adapter != nil {
				h.clearActiveModel(config.LLMTypeChat, adapter.GetConfig().Model)
			}
			logger.Logger.Warnf("LLM adapter disabled: %s", key)
			continue
		}
		return resp, callErr
	}
	return "", fmt.Errorf("LLM 调用失败：没有可用的 chat 适配器")
}

func (h *LLMAdapterHub) ChatStream(userMessages []schema.OpenAIMessage, onDelta func(string) error, systemPrompt ...string) (string, error) {
	sp := ""
	if len(systemPrompt) > 0 {
		// TODO 这里应该有一个默认的系统提示词
		sp = systemPrompt[0]
	}
	for attempt := 0; attempt < 2; attempt++ {
		active := h.chatModelByType(config.LLMTypeChat)
		adapter, key, err := h.getAdapterByTypeWithKey(config.LLMTypeChat, active)
		if err != nil {
			var de disabledAdapterError
			if errors.As(err, &de) {
				h.clearActiveModel(config.LLMTypeChat, de.Model)
				continue
			}
			return "", err
		}

		resp, callErr := adapter.ChatStream(userMessages, nil, onDelta, sp)
		if disable, d := classifyLLMFailure(callErr); disable {
			h.disableAdapter(key, d)
			if adapter != nil {
				h.clearActiveModel(config.LLMTypeChat, adapter.GetConfig().Model)
			}
			logger.Logger.Warnf("LLM adapter disabled: %s", key)
			continue
		}
		return resp.Content, callErr
	}
	return "", fmt.Errorf("LLM 调用失败：没有可用的 chat 适配器")
}

func (h *LLMAdapterHub) ChatStreamWithTools(userMessages []schema.OpenAIMessage, tools []schema.OpenAITool, onDelta func(string) error, systemPrompt ...string) (schema.OpenAIMessage, error) {
	sp := ""
	if len(systemPrompt) > 0 {
		sp = systemPrompt[0]
	}
	for range 2 {
		active := h.chatModelByType(config.LLMTypeChat)
		adapter, key, err := h.getAdapterByTypeWithKey(config.LLMTypeChat, active)
		if err != nil {
			var de disabledAdapterError
			if errors.As(err, &de) {
				h.clearActiveModel(config.LLMTypeChat, de.Model)
				continue
			}
			return schema.OpenAIMessage{}, err
		}

		resp, callErr := adapter.ChatStream(userMessages, tools, onDelta, sp)
		if disable, d := classifyLLMFailure(callErr); disable {
			h.disableAdapter(key, d)
			if adapter != nil {
				h.clearActiveModel(config.LLMTypeChat, adapter.GetConfig().Model)
			}
			logger.Logger.Warnf("LLM adapter disabled: %s", key)
			continue
		}
		return resp, callErr
	}
	return schema.OpenAIMessage{}, fmt.Errorf("LLM 调用失败：没有可用的 chat 适配器")
}

func (h *LLMAdapterHub) Embedding(text []string) ([][]float32, error) {
	for range 2 {
		active := h.chatModelByType(config.LLMTypeEmbedding)
		adapter, key, err := h.getAdapterByTypeWithKey(config.LLMTypeEmbedding, active)
		if err != nil {
			var de disabledAdapterError
			if errors.As(err, &de) {
				h.clearActiveModel(config.LLMTypeEmbedding, de.Model)
				continue
			}
			return nil, err
		}

		resp, callErr := adapter.Embedding(text)
		if disable, d := classifyLLMFailure(callErr); disable {
			h.disableAdapter(key, d)
			if adapter != nil {
				h.clearActiveModel(config.LLMTypeEmbedding, adapter.GetConfig().Model)
			}
			logger.Logger.Warnf("LLM adapter disabled: %s", key)
			continue
		}
		return resp, callErr
	}
	return nil, fmt.Errorf("Embedding 调用失败：没有可用的 embedding 适配器")
}
