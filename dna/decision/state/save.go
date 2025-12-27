package state

type SaveData interface {
	Save() error
	Load() SaveData
	Type() Flag
}
