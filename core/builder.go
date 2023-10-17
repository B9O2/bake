package core

import (
	"bytes"
	"errors"
	Executor "github.com/B9O2/ExecManager"
	"github.com/B9O2/Inspector/decorators"
	. "github.com/B9O2/Inspector/templates/simple"
	"github.com/B9O2/filefinder"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type ReplaceRule struct {
	DependencyReplace map[string]string
	ReplacementWords  map[string]string
	Range             *filefinder.SearchRule
}

type GoBuilder struct {
	executorPath            string
	exec                    *Executor.Manager
	projectPath, shadowPath string
}

// BuildProject 在影子目录中构建
func (gb *GoBuilder) BuildProject(entrance, platform, arch, dest string) error {
	os.Setenv("CGO_ENABLED", "0")
	os.Setenv("GOOS", platform)
	os.Setenv("GOARCH", arch)
	pid, err := gb.exec.NewProcess(gb.executorPath, []string{
		"build",
		"-trimpath",
		"-ldflags",
		"-w -s",
		"-o",
		dest,
		entrance,
	}, gb.shadowPath)
	if err != nil {
		return err
	}

	allStderr := ""
	for {
		stdout, stderr, err := gb.exec.FetchAll(pid)
		if err != nil {
			break
		}
		Insp.Print(Text(string(stdout)))
		allStderr += string(stderr)
	}
	if len(allStderr) > 0 {
		Insp.Print(Text("BUILD DETAIL ERR"), Text(allStderr, decorators.Red))
	}
	err = nil
	if strings.Contains(allStderr, "no Go files in") {
		err = errors.New("entrance error")
	} else if len(allStderr) > 0 {
		err = errors.New("unknown build error")
	}
	return err
}

// FileReplace 对影子目录中的文件内容进行替换
func (gb *GoBuilder) FileReplace(replacement map[string]string, replaceRange *filefinder.SearchRule) error {
	//Replace Range
	var files []string
	if replaceRange != nil {
		db, err := filefinder.NewFileDB(gb.shadowPath)
		if err != nil {
			return err
		}
		files = db.Search([]filefinder.SearchRule{*replaceRange})["OvO"]
		for _, filePath := range files {
			content, _ := os.ReadFile(filePath)
			for oldWord, newWord := range replacement {
				content = bytes.Replace(content, []byte(oldWord), []byte(newWord), -1)
			}
			_ = os.WriteFile(filePath, content, 0666)
		}
	} else {
		err := filepath.WalkDir(gb.shadowPath, func(filePath string, d fs.DirEntry, err error) error {
			content, _ := os.ReadFile(filePath)
			for oldWord, newWord := range replacement {
				content = bytes.Replace(content, []byte(oldWord), []byte(newWord), -1)
			}
			_ = os.WriteFile(filePath, content, 0666)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// GoVendor 对影子项目进行本地化依赖处理，在此过程中可以对依赖进行修改
func (gb *GoBuilder) GoVendor(replacement map[string]string) error {
	pid, err := gb.exec.NewProcess(gb.executorPath, []string{"mod", "vendor"}, gb.shadowPath)
	if err != nil {
		return err
	}
	stdout, stderr, err := gb.exec.WaitOutput(pid)
	if err != nil {
		return err
	}
	if len(stdout) > 0 {
		Insp.Print(Text(stdout, decorators.Cyan))
	}
	if len(stderr) > 0 {
		Insp.Print(Text(stderr, decorators.Red))
	}

	if len(stderr) > 0 {
		if strings.Contains(string(stderr), "go.mod file not found") {
			return errors.New("bake: It seems not a go project")
		} else {
			return errors.New("bake: vendor error")
		}
	}

	for oldDependency, newDependency := range replacement {
		err = os.Rename(filepath.Join(gb.shadowPath, "vendor", oldDependency), filepath.Join(gb.shadowPath, "vendor", newDependency))
		if err != nil {

		}
	}

	return nil
}

// duplicate 复制当前项目至
func (gb *GoBuilder) duplicate(dest string) error {
	return CopyDirectory(gb.projectPath, dest)
}

func (gb *GoBuilder) ShadowPath() string {
	return gb.shadowPath
}
func (gb *GoBuilder) ProjectPath() string {
	return gb.projectPath
}

func (gb *GoBuilder) Close() {
	if gb.shadowPath != "" {
		os.RemoveAll(gb.shadowPath)
	}
}

// NewGoProjectBuilder Go项目构建器，初始化构建器后会复制项目至影子目录（默认临时目录）
func NewGoProjectBuilder(projectPath, executorPath string) (*GoBuilder, error) {
	projectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, err
	}
	b := &GoBuilder{
		executorPath: executorPath,
		exec:         Executor.NewManager("exec"),
		projectPath:  projectPath,
		shadowPath:   "",
	}

	dest := filepath.Join(
		os.TempDir(), "BAKE_TMP", "PROJECT_TMP")

	err = b.duplicate(dest)
	if err != nil {
		return nil, err
	}
	f, _ := filepath.Abs(dest)
	b.shadowPath = f
	return b, nil
}
