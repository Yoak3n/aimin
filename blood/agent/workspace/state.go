package workspace

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/Yoak3n/aimin/blood/config"
	"github.com/Yoak3n/aimin/blood/pkg/util"
)

// 确保工作目录存在并初始化
// 返回值：是否是第一次创建工作目录
func EnsureWorkspace() bool {
	// 标记是否是第一次创建工作目录
	flag := false
	path := config.GlobalConfiguration().Workspace.Path
	if path == "" {
		// 默认工作目录
		path = "./default_workspace"
		config.GlobalConfiguration().Workspace.Path, _ = filepath.Abs(path)
		config.GlobalConfiguration().Save()
		flag = true
	}

	// 如果目录不存在，则创建
	if !util.FileExists(path) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			panic(err)
		}
		flag = true
	} // 从模板文件中复制到工作目录
	templateDir := "docs/templates"
	// 确保模板目录存在
	fp, err := os.Stat(templateDir)
	if err != nil && !fp.IsDir() {
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
		srcPath := filepath.Join(templateDir, entry.Name())
		dstPath := filepath.Join(path, entry.Name())
		err = util.CopyFile(srcPath, dstPath)
		if err != nil {
			panic(err)
		}
	}

	// 创建必要的子目录
	os.Mkdir(filepath.Join(path, "skills"), 0755)
	os.Mkdir(filepath.Join(path, "memory"), 0755)
	return flag
}
