package helper

import (
	"sync"

	"github.com/Yoak3n/aimin/blood/config"
	"github.com/Yoak3n/aimin/gut/dao/controller"
	"github.com/Yoak3n/aimin/gut/dao/implements"
	"github.com/Yoak3n/aimin/lung/adapter"
)

var once sync.Once
var hub *adapter.LLMAdapterHub

func UseLLM() *adapter.LLMAdapterHub {
	once.Do(func() {
		hub = adapter.NewLLMAdapterHub()
		for _, llm := range config.GlobalConfiguration().LLMs {
			hub.RegisterAdapter(&llm)
		}
	})
	return hub
}

func UseDB() *implements.Database {
	return controller.GetDB()
}
