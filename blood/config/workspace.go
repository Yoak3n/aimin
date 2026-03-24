package config

type Workspace struct {
	Path            string `json:"path"`
	MemoryDays      int    `json:"memory_days"`
	ContextSize     uint   `json:"context_size"`
	FileContentSize uint   `json:"file_content_size"`
}

func DefaultWorkspace() Workspace {
	return Workspace{
		Path:            "",
		MemoryDays:      7,
		ContextSize:     8000,
		FileContentSize: 2000,
	}
}
