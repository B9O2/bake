package targets

import (
	"github.com/B9O2/bake/utils"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/B9O2/Inspector/decorators"
	. "github.com/B9O2/Inspector/templates/simple"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
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
		err := dt.dc.ContainerRemove(dt.ctx, dt.containerID, types.ContainerRemoveOptions{
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
		return nil
	}
	if dt.imageID == "" {
		return errors.New("image not set")
	}
	//拉取镜像
	Insp.Print(Text("Pull Image", decorators.Yellow), Text(dt.imageID))
	out, err := dt.dc.ImagePull(dt.ctx, dt.imageID, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()

	var jsonRaw []byte
	for {
		line := make([]byte, 1024)
		n, err := out.Read(line)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		jsonRaw = append(jsonRaw, line[:n]...)
		jsons := bytes.Split(jsonRaw, []byte{'\n'})
		s := map[string]interface{}{}
		if len(jsons) > 0 {
			for _, j := range jsons {
				err := json.Unmarshal(j, &s)
				if err != nil { //无论是未闭合串或空串都应当中断，继续读取
					jsonRaw = j
					break
				}
				jsonRaw = []byte{} //以防末尾无换行的单行json
				if _, ok := s["status"]; ok {
					Insp.JustPrint(Text(s["status"], decorators.Cyan), Text("\n"))
					delete(s, "status")
					for k, v := range s {
						Insp.JustPrint(Text(" \\__ "+k+":", decorators.Blue), Text(fmt.Sprint(v)+"\n"))
					}
				} else {
					Insp.Print(Text("Docker Response", decorators.Cyan), Text(fmt.Sprint(s)))
				}
			}

		}
	}
	//启动容器
	resp, err := dt.dc.ContainerCreate(dt.ctx, &container.Config{
		Image: dt.imageID,
		Cmd:   []string{"tail", "-f", "/dev/null"},
	}, nil, nil, nil, "")
	if err != nil {
		return err
	}

	if err = dt.dc.ContainerStart(dt.ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
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

	err = dt.dc.CopyToContainer(dt.ctx, dt.containerID, dt.temp, f, types.CopyToContainerOptions{
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
