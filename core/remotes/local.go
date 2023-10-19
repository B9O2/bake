package remotes

import (
	"bake/core/utils"
	Executor "github.com/B9O2/ExecManager"
	"os"
	"path/filepath"
)

type LocalTarget struct {
	exec           *Executor.Manager
	platform, arch string
	shadowPath     string
}

func (lt *LocalTarget) Connect() error {
	lt.exec = Executor.NewManager("LocalTargetExec")
	return nil
}

func (lt *LocalTarget) CopyShadowProjectTo(src string) error {
	lt.shadowPath = src
	return nil
}

func (lt *LocalTarget) BuildExec(cmd string, args []string) ([]byte, []byte, error) {
	os.Setenv("CGO_ENABLED", "0")
	os.Setenv("GOOS", lt.platform)
	os.Setenv("GOARCH", lt.arch)
	pid, err := lt.exec.NewProcess(cmd, append([]string{"build"}, args...), lt.shadowPath)
	if err != nil {
		return nil, nil, err
	}

	var allStderr, allStdout []byte
	for {
		stdout, stderr, err := lt.exec.FetchAll(pid)
		if err != nil {
			break
		}
		allStdout = append(allStdout, stdout...)
		allStderr = append(allStderr, stderr...)
	}
	return allStdout, allStderr, nil
}

func (lt *LocalTarget) CopyFileBack(src, dest string) error {
	_, filename := filepath.Split(src)
	return utils.CopyFile(filepath.Join(lt.shadowPath, src), filepath.Join(dest, filename), 0660)
}

func (lt *LocalTarget) Info() string {
	return "Local Build"
}
func (lt *LocalTarget) Close() {}

func NewLocalTarget(platform, arch string) *LocalTarget {
	return &LocalTarget{
		platform: platform,
		arch:     arch,
	}
}
