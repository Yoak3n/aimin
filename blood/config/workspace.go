package config

type Workspace struct {
	Path            string `json:"path"`
	MemoryDays      int    `json:"memory_days"`
	ContextSize     uint   `json:"context_size"`
	FileContentSize uint   `json:"file_content_size"`
	AccessMode      string `json:"access_mode"`
	DenyPaths       []string `json:"deny_paths"`
}

func DefaultWorkspace() *Workspace {
	return &Workspace{
		Path:            "",
		MemoryDays:      7,
		ContextSize:     8000,
		FileContentSize: 2000,
		AccessMode:      "blacklist",
		DenyPaths: []string{
			"**/.ssh/**",
			"**/.gnupg/**",
			"**/.aws/**",
			"**/.kube/**",
			"**/.docker/**",
			"**/.git-credentials",
			"**/.npmrc",
			"**/AppData/**",
			"c:/**/Windows/**",
			"c:/**/Program Files/**",
			"c:/**/Program Files (x86)/**",
			"c:/**/ProgramData/**",
		},
	}
}
