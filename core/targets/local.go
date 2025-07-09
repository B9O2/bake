package targets

import (
	"os"
	"path/filepath"

	"github.com/B9O2/bake/utils"

	Executor "github.com/B9O2/ExecManager"
)

type LocalTarget struct {
	*BaseTarget
	exec *Executor.Manager
}

func (lt *LocalTarget) InitAndConnect(string) error {
	lt.exec = Executor.NewManager("LocalTargetExec")
	return nil
}

func (lt *LocalTarget) CopyShadowProjectTo(src string) error {
	lt.shadowPath = src
	return nil
}

func (lt *LocalTarget) BuildExec(cmd string, args []string, env map[string]string) ([]byte, []byte, error) {
	os.Setenv("CGO_ENABLED", "0")
	os.Setenv("GOOS", lt.platform)
	os.Setenv("GOARCH", lt.arch)

	for k, v := range env {
		os.Setenv(k, v)
	}

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
	return utils.CopyFile(filepath.Join(lt.shadowPath, src), dest, 0660)
}

func (lt *LocalTarget) Info() string {
	return "Local Build"
}
func (lt *LocalTarget) Close() error { return nil }

func NewLocalTarget(platform, arch string) *LocalTarget {
	return &LocalTarget{
		BaseTarget: NewBaseTarget(platform, arch),
	}
}
