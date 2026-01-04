package helper

import (
	"blood/adapter"
	"blood/config"
	"encoding/json"
	"os"
	"sync"
)

var once sync.Once
var hub *adapter.LLMAdapterHub

func UseLLM() *adapter.LLMAdapterHub {
	once.Do(func() {
		buf, err := os.ReadFile("config.json")
		if err != nil {
			panic(err)
		}
		var conf config.Configuration
		if err = json.Unmarshal(buf, &conf); err != nil {
			panic(err)
		}
		hub = adapter.NewLLMAdapterHub()
		for _, llm := range conf.LLMs {
			hub.RegisterAdapter(&llm)
		}
	})
	return hub
}
