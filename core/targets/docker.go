package targets

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/B9O2/Inspector/decorators"
	. "github.com/B9O2/Inspector/templates/simple"
	"github.com/B9O2/bake/utils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"
)

// DockerTarget todo docker远程目标
type DockerTarget struct {
	*BaseTarget
	host                 string
	temp                 string
	dc                   *client.Client
	ctx                  context.Context
	containerID, imageID string
	removeContainer      bool
	stopContainer        bool
}

func (dt *DockerTarget) InitAndConnect(string) error {
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

func (dt *DockerTarget) Close() error {
	if dt.removeContainer {
		err := dt.dc.ContainerRemove(dt.ctx, dt.containerID, container.RemoveOptions{
			RemoveVolumes: true,
			RemoveLinks:   false,
			Force:         true,
		})
		if err != nil {
			Insp.Print(Text("Docker Remove", decorators.Red), Error(err))
		} else {
			Insp.Print(Text("Docker Remove", decorators.Green), Text("container '"+dt.containerID+"'("+dt.imageID+") removed"))
		}
	} else {
		_, _ = dt.ExecCommand("/", nil, "rm", "-rf", dt.temp)
		if dt.stopContainer {
			timeout := 5
			if err := dt.dc.ContainerStop(dt.ctx, dt.containerID, container.StopOptions{
				Signal:  "SIGTERM",
				Timeout: &timeout, // 5秒超时
			}); err != nil {
				Insp.Print(Text("Docker Stop", decorators.Red), Error(err))
			} else {
				Insp.Print(Text("Docker Stop", decorators.Green), Text(dt.containerID, decorators.Cyan))
			}
		}
	}
	return dt.dc.Close()
}

func (dt *DockerTarget) Info() string {
	return "Docker Build"
}

func (dt *DockerTarget) CheckContainer() error {
	stats, err := dt.dc.ContainerInspect(dt.ctx, dt.containerID)
	if err == nil {
		Insp.Print(Text("Container Find", decorators.Green), Text(fmt.Sprintf("%s(%s)", stats.Name, stats.ID), decorators.Cyan))
		if !stats.State.Running {
			Insp.Print(Text("Container is not running, restarting...", decorators.Yellow))
			if err = dt.dc.ContainerStart(dt.ctx, dt.containerID, container.StartOptions{}); err != nil {
				return err
			}
			Insp.Print(Text("Container restarted", decorators.Green), Text(dt.containerID, decorators.Cyan))
			dt.removeContainer = false //不删除主动重启的容器
			dt.stopContainer = true
			return nil
		}
		return nil
	}

	if dt.imageID == "" {
		return errors.New("image not set")
	}
	//拉取镜像
	Insp.Print(Text("Pulling Image", decorators.Yellow), Text(dt.imageID, decorators.Cyan))
	out, err := dt.dc.ImagePull(dt.ctx, dt.imageID, image.PullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()

	// 使用官方推荐的 jsonmessage 来美化输出
	termFd, isTerm := term.GetFdInfo(os.Stderr)
	err = jsonmessage.DisplayJSONMessagesStream(out, os.Stderr, termFd, isTerm, nil)
	if err != nil {
		return err
	}

	Insp.Print(Text("Image pulled successfully", decorators.Green), Text(dt.imageID, decorators.Cyan))
	//启动容器
	resp, err := dt.dc.ContainerCreate(dt.ctx, &container.Config{
		Image: dt.imageID,
		Cmd:   []string{"tail", "-f", "/dev/null"},
	}, nil, nil, nil, "")
	if err != nil {
		return err
	}

	if err = dt.dc.ContainerStart(dt.ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}

	dt.containerID = resp.ID

	Insp.Print(Text("Container Started", decorators.Green), Text(resp.ID))

	dt.removeContainer = true
	return nil
}

func (dt *DockerTarget) CopyShadowProjectTo(src string) error {
	dt.shadowPath = src
	if err := dt.CheckContainer(); err != nil {
		return err
	}

	tarPath := filepath.Join(src, "../shadow_tar")
	err := utils.MakeTar(src, tarPath)
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

	err = dt.dc.CopyToContainer(dt.ctx, dt.containerID, dt.temp, f, container.CopyToContainerOptions{
		AllowOverwriteDirWithFile: false,
		CopyUIDGID:                false,
	})
	if err != nil {
		return err
	}

	return nil
}

func (dt *DockerTarget) BuildExec(executor string, args []string, env map[string]string) ([]byte, []byte, error) {
	enviorments := []string{
		"CGO_ENABLED=0",
		"GOOS=" + dt.platform,
		"GOARCH=" + dt.arch,
	}

	for k, v := range env {
		enviorments = append(enviorments, k+"="+v)
	}

	Insp.Print(Text("Command: "+executor, decorators.Cyan), Text("Args: "+strings.Join(args, " "), decorators.Cyan))
	output, err := dt.ExecCommand(dt.temp, enviorments, executor, append([]string{"build", "-buildvcs=false"}, args...)...)
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

	tarPath := filepath.Join(dt.shadowPath, "../docker_return_tar")

	err = utils.SaveFile(tarPath, tarData, true)
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

	return utils.CopyFile(filepath.ToSlash(filepath.Join(tarUnpackPath, src)), dest, stat.Mode)
}

func (dt *DockerTarget) ExecCommand(dir string, env []string, cmd string, args ...string) ([]byte, error) {
	if dir == "" {
		dir = "/"
	}
	cmds := append([]string{cmd}, args...)

	// 创建一个容器执行请求
	createResp, err := dt.dc.ContainerExecCreate(dt.ctx, dt.containerID, container.ExecOptions{
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
	resp, err := dt.dc.ContainerExecAttach(dt.ctx, createResp.ID, container.ExecAttachOptions{})
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

func NewDockerTarget(host, container, image, temp, platform, arch string) *DockerTarget {
	dt := &DockerTarget{
		BaseTarget:  NewBaseTarget(platform, arch),
		host:        host,
		containerID: container,
		imageID:     image,
		temp:        "/BAKE_DOCKER_TMP",
	}
	if temp != "" {
		dt.temp = temp
	}
	return dt
}
