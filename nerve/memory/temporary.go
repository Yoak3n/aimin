package memory

import "time"

type Temporary struct {
	*Memory
	Expired time.Time `json:"expired"`
}
