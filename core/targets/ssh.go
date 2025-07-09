package targets

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/B9O2/Inspector/decorators"
	. "github.com/B9O2/Inspector/templates/simple"
	"github.com/B9O2/bake/utils"
	"github.com/kballard/go-shellquote"
)

type SSHTarget struct {
	*BaseTarget
	sshClient  *utils.SSHClient
	temp       string
	host       string
	hashTag    string
	port       int
	authConfig *utils.SSHAuthConfig
}

func (st *SSHTarget) InitAndConnect(hashTag string) error {
	st.hashTag = hashTag
	st.temp = filepath.Join(st.temp, hashTag, "project")

	// 创建 SSH 客户端并连接
	st.sshClient = utils.NewSSHClient(st.host, st.port, st.authConfig)
	err := st.sshClient.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect SSH: %w", err)
	}

	return nil
}

func (st *SSHTarget) CopyShadowProjectTo(src string) error {
	st.shadowPath = src
	return st.sshClient.UploadDir(src, st.temp)
}

func (st *SSHTarget) BuildExec(cmd string, args []string, env map[string]string) ([]byte, []byte, error) {
	envVars := []string{
		"CGO_ENABLED=0",
		fmt.Sprintf("GOOS=%s", st.platform),
		fmt.Sprintf("GOARCH=%s", st.arch),
	}

	for k, v := range env {
		escapedValue := shellquote.Join(v)
		envVars = append(envVars, fmt.Sprintf("%s=%s", k, escapedValue))
	}

	escapedCmd := shellquote.Join(cmd)
	var escapedArgs []string
	for _, arg := range args {
		escapedArgs = append(escapedArgs, shellquote.Join(arg))
	}

	fullCmd := fmt.Sprintf("cd %s && %s %s build %s",
		shellquote.Join(st.temp),
		strings.Join(envVars, " "),
		escapedCmd,
		strings.Join(escapedArgs, " "))

	return st.sshClient.ExecCommand(fullCmd)
}

func (st *SSHTarget) CopyFileBack(src, dest string) error {
	remotePath := filepath.Join(st.temp, src)
	return st.sshClient.DownloadFile(remotePath, dest)
}

func (st *SSHTarget) Info() string {
	return fmt.Sprintf("SSH Build (%s@%s:%d)", st.authConfig.User, st.host, st.port)
}

func (st *SSHTarget) Close() error {
	if st.temp != "/" && len(st.hashTag) == 12 && strings.Contains(st.temp, st.hashTag) {
		stdout, stderr, err := st.sshClient.ExecCommand("rm -rf " + filepath.Join(st.temp, ".."))
		if err != nil {
			Insp.Print(Text("Failed to clean up temp directory", decorators.Red), Text(string(stderr), decorators.Yellow))
		}
		if len(stdout) > 0 {
			Insp.Print(Text("Cleanup output", decorators.Green), Text(string(stdout), decorators.Yellow))
		}
		if len(stderr) > 0 {
			Insp.Print(Text("Cleanup error", decorators.Red), Text(string(stderr), decorators.Yellow))
		}
		Insp.Print(Text("Remote temp directory cleanup completed", decorators.Green))
	} else {
		Insp.Print(Text("Skipping temp cleanup, temp path is suspicious", decorators.Yellow))
	}

	if st.sshClient != nil {
		return st.sshClient.Close()
	}
	return nil
}

// NewSSHTargetWithConfig 创建 SSH 目标
func NewSSHTargetWithConfig(host string, port int, temp string, platform, arch string, config *utils.SSHAuthConfig) *SSHTarget {
	st := &SSHTarget{
		BaseTarget: NewBaseTarget(platform, arch),
		host:       host,
		port:       port,
		temp:       temp,
		authConfig: config,
	}
	if st.temp == "" {
		st.temp = "/tmp/BAKE_SSH_TMP"
	}
	return st
}
