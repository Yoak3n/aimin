package controller

import (
	"fmt"
	"log"
	"time"

	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/util"
	"github.com/Yoak3n/aimin/blood/schema"
)

const (
	ChunkSize   = 80
	OverlapSize = 20
)

func InsertConversation(cid, system, question, thoughts, answer string) {
	now := time.Now()
	c := &schema.ConversationRecord{
		Id:        cid,
		Question:  question,
		Thoughts:  thoughts,
		Answer:    answer,
		System:    system,
		CreateAt:  now,
		UpdatedAt: now,
	}

	text := fmt.Sprintf("%s\n%s\n%s", question, thoughts, answer)
	chunks := splitWithOverlap(text, ChunkSize, OverlapSize)

	if len(chunks) == 0 {
		chunks = []string{text}
	}

	embeddings, err := helper.UseLLM().Embedding(chunks)
	if err != nil {
		log.Fatal(err)
	}

	if err := helper.UseDB().CreateConversationRecord(c); err != nil {
		log.Fatal(err)
	}

	textChunks := make([]schema.ConversationTextChunk, len(chunks))
	for i, chunk := range chunks {
		textChunks[i] = schema.ConversationTextChunk{
			Id:             util.RandomIdWithPrefix("ctc"),
			ConversationId: cid,
			ChunkIndex:     i,
			Content:        chunk,
			Embedding:      embeddings[i],
			CreatedAt:      now,
			UpdatedAt:      now,
		}
	}

	if err := helper.UseDB().CreateConversationTextChunks(textChunks); err != nil {
		log.Fatal(err)
	}
}

func splitWithOverlap(s string, segSize, overlap int) []string {
	// 转换为 rune 切片，以便正确处理中文
	runes := []rune(s)
	length := len(runes)

	if length == 0 {
		return []string{}
	}

	if segSize <= 0 {
		return []string{string(runes)}
	}

	if overlap >= segSize {
		overlap = segSize - 1 // 确保重叠部分小于段大小
	}

	var segments []string
	step := segSize - overlap // 每次前进的步长

	for start := 0; start < length; start += step {
		end := min(start+segSize, length)
		segments = append(segments, string(runes[start:end]))

		// 如果最后一段不足 segSize 且已经到达末尾，退出
		if end == length {
			break
		}
	}

	return segments
}
