package retrieval

import (
	"fmt"
	"strings"

	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/schema"
)

type Embedder interface {
	Embedding([]string) ([][]float32, error)
}

type ConversationStore interface {
	GetRelevantConversationRecords(vec []float32, limit int) ([]schema.ConversationRecord, error)
}

type defaultEmbedder struct{}

func (defaultEmbedder) Embedding(input []string) ([][]float32, error) {
	return helper.UseLLM().Embedding(input)
}

type defaultConversationStore struct{}

func (defaultConversationStore) GetRelevantConversationRecords(vec []float32, limit int) ([]schema.ConversationRecord, error) {
	return helper.UseDB().GetReleventConversationRecords(vec, limit)
}

func VectorSearchConversations(query string, limit int) ([]schema.ConversationRecord, error) {
	return VectorSearchConversationsWith(defaultEmbedder{}, defaultConversationStore{}, query, limit)
}

func VectorSearchConversationsWith(embedder Embedder, store ConversationStore, query string, limit int) ([]schema.ConversationRecord, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is empty")
	}
	if limit <= 0 {
		limit = 5
	}

	if embedder == nil {
		return nil, fmt.Errorf("embedder is nil")
	}
	if store == nil {
		return nil, fmt.Errorf("store is nil")
	}

	emb, err := embedder.Embedding([]string{query})
	if err != nil {
		return nil, err
	}
	if len(emb) == 0 {
		return nil, fmt.Errorf("embedding is empty")
	}
	return store.GetRelevantConversationRecords(emb[0], limit)
}
