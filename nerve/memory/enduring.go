package memory

type Enduring struct {
	*Memory
	EntityId []uint `json:"entity_id"`
}
