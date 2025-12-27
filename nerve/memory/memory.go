package memory

import "time"

type Memory struct {
	Id            string    `json:"id"`
	Topic         string    `json:"topic"`
	Content       string    `json:"content"`
	Level         float64   `json:"level"`
	Create        time.Time `json:"create"`
	LastSimulated time.Time `json:"last_simulated"`
}
