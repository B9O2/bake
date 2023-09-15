package core

import (
	"bytes"
	"errors"
	"fmt"
	"gitlab.huaun.com/lr/filefinder"
	Executor "gitlab.huaun.com/lr/utils/ExecManager"
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

type Builder struct {
	executorPath            string
	exec                    *Executor.Manager
	projectPath, shadowPath string
}

func (b *Builder) BuildProject(entrance, platform, arch, dest string) error {
	os.Setenv("CGO_ENABLED", "0")
	os.Setenv("GOOS", platform)
	os.Setenv("GOARCH", arch)
	pid, err := b.exec.NewProcess(b.executorPath, []string{
		"build",
		"-trimpath",
		"-ldflags",
		"-w -s",
		"-o",
		dest,
		entrance,
	}, b.shadowPath)
	if err != nil {
		return err
	}

	allStderr := ""
	for {
		stdout, stderr, err := b.exec.FetchAll(pid)
		if err != nil {
			break
		}
		fmt.Print(string(stdout))
		allStderr += string(stderr)
	}
	fmt.Println("BUILD DETAIL ERR:", allStderr)
	err = nil
	if strings.Contains(allStderr, "no Go files in") {
		err = errors.New("entrance error")
	} else if len(allStderr) > 0 {
		err = errors.New("unknown build error")
	}
	return err
}

func (b *Builder) GoVendor(replaceRule ReplaceRule) error {
	fmt.Println("bake: Run go mod vendor @ ", b.shadowPath)
	pid, err := b.exec.NewProcess("go", []string{"mod", "vendor"}, b.shadowPath)
	if err != nil {
		return err
	}
	stdout, stderr, err := b.exec.WaitOutput(pid)
	if err != nil {
		return err
	}
	fmt.Println("OUT:", string(stdout))
	fmt.Println("ERR:", string(stderr))
	if len(stderr) > 0 {
		if strings.Contains(string(stderr), "go.mod file not found") {
			return errors.New("bake: It seems not a go project")
		} else {
			return errors.New("bake: vendor error")
		}
	}

	for oldDependency, newDependency := range replaceRule.DependencyReplace {
		err := os.Rename(filepath.Join(b.shadowPath, "vendor", oldDependency), filepath.Join(b.shadowPath, "vendor", newDependency))
		if err != nil {
			fmt.Println(">> rename dependency error ", err)
		}
	}

	//Replace Range
	var files []string
	if replaceRule.Range != nil {
		db, err := filefinder.NewFileDB(b.shadowPath)
		if err != nil {
			return err
		}
		files = db.Search([]filefinder.SearchRule{*replaceRule.Range})["OvO"]
		for _, filePath := range files {
			content, _ := os.ReadFile(filePath)
			for oldWord, newWord := range replaceRule.ReplacementWords {
				content = bytes.Replace(content, []byte(oldWord), []byte(newWord), -1)
			}
			_ = os.WriteFile(filePath, content, 0666)
		}
	} else {
		err := filepath.WalkDir(b.shadowPath, func(filePath string, d fs.DirEntry, err error) error {
			content, _ := os.ReadFile(filePath)
			for oldWord, newWord := range replaceRule.ReplacementWords {
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

func (b *Builder) duplicate(dest string) error {
	fmt.Println("bake: Copy project")
	return CopyDirectory(b.projectPath, dest)
}

func (b *Builder) ShadowPath() string {
	return b.shadowPath
}
func (b *Builder) ProjectPath() string {
	return b.projectPath
}

func (b *Builder) Close() {
	if b.shadowPath != "" {
		os.RemoveAll(b.shadowPath)
	}
}

func NewBuilder(projectPath, executorPath string) (*Builder, error) {
	projectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, err
	}
	b := &Builder{
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
