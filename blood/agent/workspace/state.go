package workspace

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/Yoak3n/aimin/blood/config"
	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/util"
)

type WorkspaceState struct {
	BootstrapSeededAt string `json:"bootstrapSeededAt,omitempty"`
	SetupCompletedAt  string `json:"setupCompletedAt,omitempty"`
}

// 确保工作目录存在并初始化
// 返回值：是否是第一次创建工作目录
func EnsureWorkspace() bool {
	// 标记是否是第一次创建工作目录
	flag := false
	workspacePath := config.GlobalConfiguration().Workspace.Path
	if workspacePath == "" {
		// 默认工作目录
		workspacePath = "./default_workspace"
		config.GlobalConfiguration().Workspace.Path, _ = filepath.Abs(workspacePath)
		config.GlobalConfiguration().Save()
		flag = true
	}

	// 如果目录不存在，则创建
	if !util.FileExists(workspacePath) {
		err := os.MkdirAll(workspacePath, 0755)
		if err != nil {
			panic(err)
		}
		flag = true
	}
	brandNew := isBrandNewWorkspace(workspacePath)

	// 从模板文件中复制到工作目录
	templateDir := "docs/templates"
	// 确保模板目录存在
	fp, err := os.Stat(templateDir)
	if err != nil || !fp.IsDir() {
		panic(errors.New("template directory does not exist"))
	}
	entries, err := os.ReadDir(templateDir)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Name() == "BOOTSTRAP.md" {
			continue
		}
		srcPath := filepath.Join(templateDir, entry.Name())
		dstPath := filepath.Join(workspacePath, entry.Name())
		err = util.CopyFile(srcPath, dstPath)
		if err != nil {
			panic(err)
		}
	}

	// 创建必要的子目录
	_ = os.Mkdir(filepath.Join(workspacePath, "skills"), 0755)
	_ = os.Mkdir(filepath.Join(workspacePath, "memory"), 0755)

	state, _ := LoadWorkspaceState(workspacePath)
	bootstrapPath := filepath.Join(workspacePath, "BOOTSTRAP.md")
	bootstrapExists := util.FileExists(bootstrapPath)

	if state.BootstrapSeededAt == "" && bootstrapExists {
		state.BootstrapSeededAt = nowIso()
		_ = SaveWorkspaceState(workspacePath, state)
	}

	if state.SetupCompletedAt == "" && state.BootstrapSeededAt != "" && !bootstrapExists {
		state.SetupCompletedAt = nowIso()
		_ = SaveWorkspaceState(workspacePath, state)
		return flag || brandNew
	}

	if state.BootstrapSeededAt == "" && !bootstrapExists {
		legacySetupCompleted := detectLegacySetupCompleted(workspacePath, templateDir)
		if legacySetupCompleted {
			state.SetupCompletedAt = nowIso()
			_ = SaveWorkspaceState(workspacePath, state)
			return flag || brandNew
		}

		bootstrapTemplatePath := filepath.Join(templateDir, "BOOTSTRAP.md")
		err = util.CopyFile(bootstrapTemplatePath, bootstrapPath)
		if err != nil {
			panic(err)
		}
		state.BootstrapSeededAt = nowIso()
		_ = SaveWorkspaceState(workspacePath, state)
	}

	return flag || brandNew
}

func LoadWorkspaceState(workspacePath string) (WorkspaceState, error) {
	statePath := workspaceStatePath(workspacePath)
	buf, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return WorkspaceState{}, nil
		}
		return WorkspaceState{}, err
	}
	if len(buf) == 0 {
		return WorkspaceState{}, nil
	}
	var state WorkspaceState
	if err := json.Unmarshal(buf, &state); err != nil {
		return WorkspaceState{}, err
	}
	return state, nil
}

func SaveWorkspaceState(workspacePath string, state WorkspaceState) error {
	statePath := workspaceStatePath(workspacePath)
	if err := os.MkdirAll(filepath.Dir(statePath), 0755); err != nil {
		return err
	}
	buf, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(statePath, buf, 0644)
}

func workspaceStatePath(workspacePath string) string {
	return filepath.Join(workspacePath, "workspace-state.json")
}

func nowIso() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func isBrandNewWorkspace(workspacePath string) bool {
	requiredPaths := []string{
		filepath.Join(workspacePath, "AGENTS.md"),
		filepath.Join(workspacePath, "SOUL.md"),
		filepath.Join(workspacePath, "TOOLS.md"),
		filepath.Join(workspacePath, "IDENTITY.md"),
		filepath.Join(workspacePath, "USER.md"),
		filepath.Join(workspacePath, "memory"),
		filepath.Join(workspacePath, ".git"),
	}
	return !slices.ContainsFunc(requiredPaths, util.FileExists)
}

func detectLegacySetupCompleted(workspacePath string, templateDir string) bool {
	identityTemplate := readStrippedTemplate(filepath.Join(templateDir, "IDENTITY.md"))
	userTemplate := readStrippedTemplate(filepath.Join(templateDir, "USER.md"))

	identityPath := filepath.Join(workspacePath, "IDENTITY.md")
	userPath := filepath.Join(workspacePath, "USER.md")

	identityContent := readStrippedFile(identityPath)
	userContent := readStrippedFile(userPath)

	identityModified := identityContent != "" && identityTemplate != "" && identityContent != identityTemplate
	userModified := userContent != "" && userTemplate != "" && userContent != userTemplate

	if identityModified || userModified {
		return true
	}

	if util.FileExists(filepath.Join(workspacePath, ".git")) {
		return true
	}

	if hasAnyFile(filepath.Join(workspacePath, "memory")) {
		return true
	}

	if util.FileExists(filepath.Join(workspacePath, "MEMORY.md")) {
		return true
	}

	return false
}

func readStrippedTemplate(path string) string {
	buf, err := os.ReadFile(path)
	if err != nil || len(buf) == 0 {
		return ""
	}
	return helper.StripFrontMatter(string(buf))
}

func readStrippedFile(path string) string {
	buf, err := os.ReadFile(path)
	if err != nil || len(buf) == 0 {
		return ""
	}
	return helper.StripFrontMatter(string(buf))
}

func hasAnyFile(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for range entries {
		return true
	}
	return false
}
