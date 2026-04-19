package controller

import (
	"fmt"
	"log"
	"time"

	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/schema"
)

const (
	ChunkSize   = 80
	OverlapSize = 10
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
	text := fmt.Sprintf("%s\n%s", question, answer)
	// // 文本分割为多个段落
	// chunks := splitWithOverlap(text, ChunkSize, OverlapSize)

	embeding, err := helper.UseLLM().Embedding([]string{text})
	if err != nil {
		log.Fatal(err)
	}
	helper.UseDB().CreateConversationRecord(c, embeding[0])
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
		end := start + segSize
		if end > length {
			end = length
		}
		segments = append(segments, string(runes[start:end]))

		// 如果最后一段不足 segSize 且已经到达末尾，退出
		if end == length {
			break
		}
	}

	return segments
}
