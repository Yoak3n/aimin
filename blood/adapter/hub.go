package adapter

import (
	"blood/pkg/logger"
	"fmt"
	"math/rand/v2"
	"sync"

	"blood/config"
	"blood/schema"
)

type LLMAdapterHub struct {
	Adapters map[string]LLMAdapter
	payload  map[string]float64
	mutex    sync.RWMutex
}

func NewLLMAdapterHub() *LLMAdapterHub {
	return &LLMAdapterHub{
		Adapters: make(map[string]LLMAdapter),
		payload:  make(map[string]float64),
		mutex:    sync.RWMutex{},
	}
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
	logger.Logger.Infof("注册了适配器: %s\n", config.Provider+"_"+config.Model)
}

// UnregisterAdapter 移除适配器
func (h *LLMAdapterHub) UnregisterAdapter(key string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	delete(h.Adapters, key)
	delete(h.payload, key)
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

func (h *LLMAdapterHub) getAdapterByType(llmType config.LLMType) (LLMAdapter, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var candidates []string
	for key, adapter := range h.Adapters {
		if config.LLMType(adapter.GetConfig().Type) == llmType {
			candidates = append(candidates, key)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("没有找到类型为 %s 的适配器", llmType)
	}

	// 负载均衡选择
	selectedKey := h.selectByLoadBalance(candidates)
	return h.Adapters[selectedKey], nil
}

func (h *LLMAdapterHub) Chat(userMessages []schema.OpenAIMessage, systemPrompt string) (string, error) {
	adapter, err := h.getAdapterByType(config.LLMTypeChat)
	if err != nil {
		return "", err
	}
	return adapter.Chat(userMessages, systemPrompt)
}

func (h *LLMAdapterHub) Embedding(text []string) ([][]float32, error) {
	adapter, err := h.getAdapterByType(config.LLMTypeEmbedding)
	if err != nil {
		return nil, err
	}
	return adapter.Embedding(text)
}
