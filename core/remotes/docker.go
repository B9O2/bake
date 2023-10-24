package remotes

import (
	"bake/utils"
	"context"
	"errors"
	"fmt"
	"github.com/B9O2/Inspector/decorators"
	. "github.com/B9O2/Inspector/templates/simple"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// DockerTarget todo docker远程目标
type DockerTarget struct {
	host           string
	temp           string
	platform, arch string
	dc             *client.Client
	ctx            context.Context
	shadowPath     string
	containerID    string
}

func (dt *DockerTarget) Connect() error {
	var options []client.Opt
	var err error
	if dt.host == "" {
		return errors.New("docker host is empty (use 'local' ?)")
	}

	if dt.host != "local" {
		options = append(options, client.WithHost(dt.host))
	}

	dt.dc, err = client.NewClientWithOpts(options...)
	if err != nil {
		return err
	}

	dt.ctx = context.Background()

	info, err := dt.dc.Info(dt.ctx)
	if err != nil {
		return err
	}
	Insp.Print(Text("Docker Connected", decorators.Green), Text(info.Name, decorators.Cyan))
	return nil
}

func (dt *DockerTarget) Close() {
	_, _ = dt.ExecCommand("/", nil, "rm", "-rf", dt.temp)
	_ = dt.dc.Close()
}

func (dt *DockerTarget) Info() string {
	return "Docker Build"
}

func (dt *DockerTarget) CopyShadowProjectTo(src string) error {
	dt.shadowPath = src
	if dt.containerID == "" {
		return errors.New("container not set")
	}
	stats, err := dt.dc.ContainerInspect(dt.ctx, dt.containerID)
	if err != nil {
		return err
	}
	Insp.Print(Text("Container Find", decorators.Green), Text(fmt.Sprintf("%s(%s)", stats.Name, stats.ID), decorators.Cyan))
	tarPath := filepath.Join(src, "../shadow_tar")
	err = utils.MakeTar(src, tarPath)
	if err != nil {
		return err
	}
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}

	_, err = dt.ExecCommand("", nil, "mkdir", dt.temp)
	if err != nil {
		return err
	}

	err = dt.dc.CopyToContainer(dt.ctx, dt.containerID, dt.temp, f, types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: false,
		CopyUIDGID:                false,
	})
	if err != nil {
		return err
	}

	return nil
}

func (dt *DockerTarget) BuildExec(executor string, args []string) ([]byte, []byte, error) {
	output, err := dt.ExecCommand(dt.temp, []string{
		"CGO_ENABLED=0",
		"GOOS=" + dt.platform,
		"GOARCH=" + dt.arch,
	}, executor, append([]string{"build", "-buildvcs=false"}, args...)...)
	if err != nil {
		return nil, nil, err
	}
	out := string(output)
	if len(out) > 0 {
		Insp.Print(Text("Container <"+dt.containerID+"> Build", decorators.Cyan), Text(out))
	}

	if strings.Contains(out, "failed") {
		return nil, nil, errors.New(out)
	}
	return output, nil, nil
}

func (dt *DockerTarget) CopyFileBack(src, dest string) error {
	tarData, stat, err := dt.dc.CopyFromContainer(dt.ctx, dt.containerID, filepath.Join(dt.temp, "shadow_bin"))
	if err != nil {
		return err
	}

	all, err := io.ReadAll(tarData)
	if err != nil {
		return err
	}

	tarPath := filepath.Join(dt.shadowPath, "../docker_return_tar")

	err = utils.SaveFile(tarPath, all, true)
	if err != nil {
		return err
	}

	tarUnpackPath := filepath.Join(dt.shadowPath, "../docker_return")
	if err = os.Mkdir(tarUnpackPath, 0750); err != nil {
		return err
	}

	err = utils.UnpackTar(tarPath, tarUnpackPath, false)
	if err != nil {
		return err
	}

	return utils.CopyFile(filepath.Join(tarUnpackPath, src), dest, stat.Mode)
}

func (dt *DockerTarget) ExecCommand(dir string, env []string, cmd string, args ...string) ([]byte, error) {
	if dir == "" {
		dir = "/"
	}
	cmds := append([]string{cmd}, args...)

	// 创建一个容器执行请求
	createResp, err := dt.dc.ContainerExecCreate(dt.ctx, dt.containerID, types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmds,
		Env:          env,
		WorkingDir:   dir,
	})
	if err != nil {
		return nil, err
	}

	// 执行命令并获取输出
	resp, err := dt.dc.ContainerExecAttach(dt.ctx, createResp.ID, types.ExecStartCheck{})
	if err != nil {
		return nil, err
	}
	defer resp.Close()
	// 读取命令输出
	output, err := io.ReadAll(resp.Reader)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func NewDockerTarget(host, container, temp, platform, arch string) *DockerTarget {
	dt := &DockerTarget{
		host:        host,
		containerID: container,
		temp:        "/BAKE_DOCKER_TMP",
		platform:    platform,
		arch:        arch,
	}
	if temp != "" {
		dt.temp = temp
	}
	return dt
}
