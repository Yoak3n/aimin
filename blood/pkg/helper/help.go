package helper

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/Yoak3n/aimin/blood/adapter"
	"github.com/Yoak3n/aimin/blood/config"
	"github.com/Yoak3n/aimin/blood/dao/controller"
	"github.com/Yoak3n/aimin/blood/dao/implements"
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

func UseDB() *implements.Database {
	return controller.GetDB()
}
