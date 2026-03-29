package workspace

import (
	"path/filepath"
	"time"

	"github.com/Yoak3n/aimin/blood/config"
	"github.com/Yoak3n/aimin/blood/pkg/util"
)

type FileSpec struct {
	// 文件名
	Name     string
	RelPath  string
	Required bool
}

type ContextChoice int8

const (
	Normal ContextChoice = iota
	Remote
)

func makeFileSpecMap(choose ContextChoice) map[string]FileSpec {
	ret := make(map[string]FileSpec)
	// 从工作目录中读取所有文件
	ret["BOOTSTRAP.md"] = FileSpec{
		Name:     "BOOTSTRAP.md",
		RelPath:  "/BOOTSTRAP.md",
		Required: false,
	}
	ret["AGENTS.md"] = FileSpec{
		Name:     "AGENTS.md",
		RelPath:  "/AGENTS.md",
		Required: false,
	}
	ret["SOUL.md"] = FileSpec{
		Name:     "SOUL.md",
		RelPath:  "/SOUL.md",
		Required: false,
	}
	ret["USER.md"] = FileSpec{
		Name:     "USER.md",
		RelPath:  "/USER.md",
		Required: false,
	}

	ret["TASKS.md"] = FileSpec{
		Name:     "TASKS.md",
		RelPath:  "/TASKS.md",
		Required: false,
	}

	ret["HEARTBEAT.md"] = FileSpec{
		Name:     "HEARTBEAT.md",
		RelPath:  "/HEARTBEAT.md",
		Required: false,
	}

	ret["BOOT.md"] = FileSpec{
		Name:     "BOOT.md",
		RelPath:  "/BOOT.md",
		Required: false,
	}

	ret["TOOLS.md"] = FileSpec{
		Name:     "TOOLS.md",
		RelPath:  "/TOOLS.md",
		Required: false,
	}

	ret["MEMORY.md"] = FileSpec{
		Name:     "MEMORY.md",
		RelPath:  "/MEMORY.md",
		Required: false,
	}

	switch choose {
	case Normal:
		setRequired(ret, "AGENTS.md")
		setRequired(ret, "SOUL.md")
		setRequired(ret, "USER.md")
		setRequired(ret, "TASKS.md")
		setRequired(ret, "HEARTBEAT.md")
		setRequired(ret, "BOOT.md")
		setRequired(ret, "TOOLS.md")
		setRequired(ret, "MEMORY.md")
		// 从今天开始往前算的天数
		for _, spec := range diaryFiles() {
			ret[spec.Name] = spec
		}
	case Remote:
		setRequired(ret, "AGENTS.md")
		setRequired(ret, "SOUL.md")
		setRequired(ret, "USER.md")
		setRequired(ret, "TASKS.md")
		setRequired(ret, "BOOT.md")
		setRequired(ret, "TOOLS.md")
		// 从今天开始往前算的天数
		for _, spec := range diaryFiles() {
			ret[spec.Name] = spec
		}
	}
	return ret
}

func setRequired(files map[string]FileSpec, key string) map[string]FileSpec {
	spec, ok := files[key]
	if ok {
		spec.Required = true
		files[key] = spec
	}
	return files
}

func diaryFiles() map[string]FileSpec {
	ret := make(map[string]FileSpec)
	workspacePath := config.GlobalConfiguration().Workspace.Path
	memoryDays := config.GlobalConfiguration().Workspace.MemoryDays
	// 从今天开始往前算的天数
	for i := range memoryDays {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		relPath := "/memory/" + date + ".md"
		absPath := filepath.Join(workspacePath, relPath)
		if util.FileExists(absPath) {
			ret[date+".md"] = FileSpec{
				Name:     date + ".md",
				RelPath:  relPath,
				Required: true,
			}
		}
	}
	return ret
}
