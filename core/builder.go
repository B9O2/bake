package core

import (
	"bake/core/recipe"
	"bake/utils"
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

type GoBuilder struct {
	dev                     bool
	executorPath            string
	exec                    *Executor.Manager
	projectPath, shadowPath string
	hashTag                 string
}

// BuildProject 在影子目录中构建
func (gb *GoBuilder) BuildProject(entrance, output string, pair recipe.BuildPair) (string, error) {
	shadowOutput := filepath.Join("./shadow_bin", pair.Name())
	cmd := gb.executorPath
	args := []string{ //不指定一定使用build
		"-trimpath",
		"-ldflags",
		"-w -s",
		"-o",
		shadowOutput,
		entrance,
	}
	err := pair.Remote.Connect()
	if err != nil {
		return "", err
	}

	if !gb.dev {
		defer pair.Remote.Close()
	}

	err = pair.Remote.CopyShadowProjectTo(gb.shadowPath)
	if err != nil {
		return "", err
	}
	_, stderr, err := pair.Remote.BuildExec(cmd, args)
	if err != nil {
		return "", err
	}
	if len(stderr) > 0 {
		if bytes.Contains(stderr, []byte("no Go files in")) {
			err = errors.New("entrance error")
		} else {
			err = errors.New(string(stderr))
		}
		return "", err
	}
	output = filepath.Join(output, pair.Name())
	err = pair.Remote.CopyFileBack(shadowOutput, output)
	return output, err
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
		Insp.Print(Text(string(stdout), decorators.Cyan))
	}
	if len(stderr) > 0 {
		Insp.Print(LEVEL_WARNING, Text(string(stderr), decorators.Yellow))
		if strings.Contains(string(stderr), "go.mod file not found") {
			return errors.New("bake: It seems not a go project")
		}
	}

	for oldDependency, newDependency := range replacement {
		err = os.Rename(filepath.Join(gb.shadowPath, "vendor", oldDependency), filepath.Join(gb.shadowPath, "vendor", newDependency))
		if err != nil {
			return err
		}
	}

	return nil
}

// duplicate 复制当前项目至
func (gb *GoBuilder) duplicate(dest string) error {
	return utils.CopyDirectory(gb.projectPath, dest)
}

func (gb *GoBuilder) ShadowPath() string {
	return gb.shadowPath
}
func (gb *GoBuilder) ProjectPath() string {
	return gb.projectPath
}

func (gb *GoBuilder) Close() error {
	if !gb.dev {
		if gb.shadowPath != "" {
			return os.RemoveAll(filepath.Join(gb.shadowPath, ".."))
		}
	} else {
		return errors.New("DevMode")
	}
	return nil
}

// NewGoProjectBuilder Go项目构建器，初始化构建器后会复制项目至影子目录（默认临时目录）
func NewGoProjectBuilder(projectPath, executorPath string, dev bool) (*GoBuilder, error) {
	projectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, err
	}
	b := &GoBuilder{
		dev:          dev,
		executorPath: executorPath,
		exec:         Executor.NewManager("exec"),
		projectPath:  projectPath,
	}
	b.hashTag = utils.RandStr(12)
	dest := filepath.Join(
		os.TempDir(), "BAKE_TMP", b.hashTag, "SHADOW_PROJECT")

	err = b.duplicate(dest)
	if err != nil {
		return nil, err
	}
	f, _ := filepath.Abs(dest)
	b.shadowPath = f
	return b, nil
}
