package retrieval

import (
	"testing"

	"github.com/Yoak3n/aimin/blood/schema"
)

type fakeEmbedder struct {
	out [][]float32
	err error
}

func (f fakeEmbedder) Embedding(_ []string) ([][]float32, error) { return f.out, f.err }

type fakeStore struct {
	gotVec   []float32
	gotLimit int
	out      []schema.ConversationRecord
	err      error
}

func (s *fakeStore) GetRelevantConversationRecords(vec []float32, limit int) ([]schema.ConversationRecord, error) {
	s.gotVec = append([]float32(nil), vec...)
	s.gotLimit = limit
	return s.out, s.err
}

func TestVectorSearchConversationsWith_PassesEmbeddingAndLimit(t *testing.T) {
	emb := fakeEmbedder{out: [][]float32{{1, 2, 3}}}
	store := &fakeStore{out: []schema.ConversationRecord{{Id: "con_1"}}}

	records, err := VectorSearchConversationsWith(emb, store, "hello", 7)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(records) != 1 || records[0].Id != "con_1" {
		t.Fatalf("unexpected records: %#v", records)
	}
	if store.gotLimit != 7 {
		t.Fatalf("unexpected limit: %d", store.gotLimit)
	}
	if len(store.gotVec) != 3 || store.gotVec[0] != 1 || store.gotVec[2] != 3 {
		t.Fatalf("unexpected vec: %#v", store.gotVec)
	}
}
