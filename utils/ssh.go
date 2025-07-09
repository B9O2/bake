package utils

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/B9O2/Inspector/decorators"
	. "github.com/B9O2/Inspector/templates/simple"
	"github.com/pkg/sftp"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type SSHAuthMethod string

func (m SSHAuthMethod) String() string {
	return string(m)
}

const (
	SSHAuthMethodPassword    SSHAuthMethod = "password"
	SSHAuthMethodPrivateKey  SSHAuthMethod = "privatekey"
	SSHAuthMethodAgent       SSHAuthMethod = "agent"
	SSHAuthMethodInteractive SSHAuthMethod = "interactive"
)

// SSHAuthConfig 定义 SSH 认证配置
type SSHAuthConfig struct {
	User               string
	Password           string
	PrivateKeyPath     string
	PrivateKeyPassword string
	HostKeyCallback    ssh.HostKeyCallback
}

type SSHClient struct {
	client     *ssh.Client
	sftpClient *sftp.Client
	host       string
	port       int
	authConfig *SSHAuthConfig
}

// NewSSHClient 创建新的 SSH 客户端
func NewSSHClient(host string, port int, config *SSHAuthConfig) *SSHClient {
	return &SSHClient{
		host:       host,
		port:       port,
		authConfig: config,
	}
}

// Connect 连接到 SSH 服务器
func (sc *SSHClient) Connect() error {
	methods, authMethods, err := sc.buildAuthMethods()
	if err != nil {
		return fmt.Errorf("failed to build auth methods: %w", err)
	}

	Insp.Print(Text("Using auth method", decorators.Cyan), Text(JoinStrings(methods, ","), decorators.Green))

	// 如果没有指定主机密钥回调，则使用默认的忽略主机
	hostKeyCallback := sc.authConfig.HostKeyCallback
	if hostKeyCallback == nil {
		hostKeyCallback = ssh.InsecureIgnoreHostKey() // 默认忽略主机密钥验证
	}

	config := &ssh.ClientConfig{
		User:            sc.authConfig.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sc.host, sc.port), config)
	if err != nil {
		return fmt.Errorf("failed to connect to %s:%d : %w", sc.host, sc.port, err)
	}
	sc.client = client
	Insp.Print(Text("SSH Connected", decorators.Green))

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	sc.sftpClient = sftpClient
	Insp.Print(Text("SFTP Connected", decorators.Green))
	return nil
}

// ExecCommand 执行远程命令
func (sc *SSHClient) ExecCommand(cmd string) ([]byte, []byte, error) {
	session, err := sc.client.NewSession()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	Insp.Print(Text("Executing command", decorators.Cyan), Text(cmd, decorators.Yellow))

	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderrPipe, err := session.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	err = session.Start(cmd)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start command: %w", err)
	}

	var stdout, stderr []byte
	var stdoutErr, stderrErr error

	done := make(chan bool, 2)

	go func() {
		stdout, stdoutErr = io.ReadAll(stdoutPipe)
		done <- true
	}()

	go func() {
		stderr, stderrErr = io.ReadAll(stderrPipe)
		done <- true
	}()

	<-done
	<-done

	if stdoutErr != nil {
		return nil, nil, fmt.Errorf("failed to read stdout: %w", stdoutErr)
	}
	if stderrErr != nil {
		return nil, nil, fmt.Errorf("failed to read stderr: %w", stderrErr)
	}

	err = session.Wait()
	if err != nil {
		return stdout, stderr, fmt.Errorf("command execution failed: %w", err)
	}

	return stdout, stderr, nil
}

// UploadFile 上传单个文件
func (sc *SSHClient) UploadFile(localPath, remotePath string) error {
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file %s: %w", localPath, err)
	}
	defer localFile.Close()

	remoteFile, err := sc.sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	// 读取本地文件内容
	fileData, err := io.ReadAll(localFile)
	if err != nil {
		return fmt.Errorf("failed to read local file %s: %w", localPath, err)
	}

	// 写入到远程文件
	_, err = remoteFile.Write(fileData)
	if err != nil {
		return fmt.Errorf("failed to write remote file %s: %w", remotePath, err)
	}
	return nil
}

// UploadDir 上传整个目录，带进度条
func (sc *SSHClient) UploadDir(localDir, remoteDir string) error {
	Insp.Print(Text("Uploading Directory", decorators.Yellow), Text(localDir, decorators.Magenta), Text("->", decorators.Yellow), Text(remoteDir, decorators.Magenta))

	// 首先统计总文件数量
	totalFiles, err := sc.countFiles(localDir)
	if err != nil {
		return fmt.Errorf("failed to count files: %w", err)
	}

	if totalFiles == 0 {
		Insp.Print(Text("No files to upload", decorators.Yellow))
		return nil
	}

	// 创建进度条
	progressBar := progressbar.NewOptions(totalFiles,
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("files"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionOnCompletion(func() {
			fmt.Println()
			Insp.Print(Text("Upload completed!", decorators.Green))
		}),
	)

	return sc.uploadDirWithProgress(localDir, remoteDir, progressBar)
}

// uploadDirWithProgress 带进度的上传目录
func (sc *SSHClient) uploadDirWithProgress(localDir, remoteDir string, progressBar *progressbar.ProgressBar) error {
	// 创建远程目录
	err := sc.sftpClient.MkdirAll(remoteDir)
	if err != nil {
		return fmt.Errorf("failed to create remote directory %s: %w", remoteDir, err)
	}

	// 获取本地目录中的所有文件和子目录
	files, err := os.ReadDir(localDir)
	if err != nil {
		return fmt.Errorf("failed to read local directory %s: %w", localDir, err)
	}

	// 遍历本地目录中的文件和子目录
	for _, file := range files {
		localPath := filepath.Join(localDir, file.Name())
		remotePath := filepath.Join(remoteDir, file.Name())

		// 如果是目录，递归调用上传
		if file.IsDir() {
			err = sc.uploadDirWithProgress(localPath, remotePath, progressBar)
			if err != nil {
				return err
			}
		} else {
			// 如果是文件，上传文件
			err = sc.UploadFile(localPath, remotePath)
			if err != nil {
				return err
			}

			// 更新进度条
			progressBar.Add(1)
		}
	}
	return nil
}

// countFiles 递归统计目录中的文件数量
func (sc *SSHClient) countFiles(dir string) (int, error) {
	count := 0

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})

	return count, err
}

// DownloadFile 下载文件
func (sc *SSHClient) DownloadFile(remotePath, localPath string) error {
	remoteFile, err := sc.sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	err = SaveFile(localPath, remoteFile, true)
	if err != nil {
		return fmt.Errorf("failed to write to local file %s: %w", localPath, err)
	}

	return nil
}

// DownloadDir 下载目录
func (sc *SSHClient) DownloadDir(remoteDir, localDir string) error {
	// 创建本地目录
	err := os.MkdirAll(localDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create local directory %s: %w", localDir, err)
	}

	// 获取远程目录中的所有文件和子目录
	files, err := sc.sftpClient.ReadDir(remoteDir)
	if err != nil {
		return fmt.Errorf("failed to read remote directory %s: %w", remoteDir, err)
	}

	// 遍历远程目录中的文件和子目录
	for _, file := range files {
		remotePath := filepath.Join(remoteDir, file.Name())
		localPath := filepath.Join(localDir, file.Name())

		// 如果是目录，递归调用下载
		if file.IsDir() {
			err = sc.DownloadDir(remotePath, localPath)
			if err != nil {
				return err
			}
		} else {
			// 如果是文件，下载文件
			err = sc.DownloadFile(remotePath, localPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Close 关闭连接
func (sc *SSHClient) Close() error {
	if sc.sftpClient != nil {
		sc.sftpClient.Close()
	}
	if sc.client != nil {
		sc.client.Close()
	}
	return nil
}

// GetSFTPClient 获取 SFTP 客户端
func (sc *SSHClient) GetSFTPClient() *sftp.Client {
	return sc.sftpClient
}

// GetSSHClient 获取 SSH 客户端
func (sc *SSHClient) GetSSHClient() *ssh.Client {
	return sc.client
}

// buildAuthMethods 根据配置构建认证方式
func (sc *SSHClient) buildAuthMethods() ([]SSHAuthMethod, []ssh.AuthMethod, error) {
	var methods []SSHAuthMethod
	var authMethods []ssh.AuthMethod

	if sc.authConfig.PrivateKeyPath != "" {
		methods = append(methods, SSHAuthMethodPrivateKey)
		if authMethod, err := sc.buildPrivateKeyAuth(); err != nil {
			return nil, nil, err
		} else {
			authMethods = append(authMethods, authMethod)
		}
	}

	if sc.authConfig.Password != "" {
		methods = append(methods, SSHAuthMethodPassword)
		authMethods = append(authMethods, ssh.Password(sc.authConfig.Password))
	}

	if len(methods) == 0 {
		methods = append(methods, SSHAuthMethodAgent)
		if authMethod, err := sc.buildAgentAuth(); err != nil {
			return nil, nil, err
		} else {
			authMethods = append(authMethods, authMethod)
		}
	}

	return methods, authMethods, nil
}

// buildPrivateKeyAuth 构建私钥认证方式
func (sc *SSHClient) buildPrivateKeyAuth() (ssh.AuthMethod, error) {
	key, err := os.ReadFile(sc.authConfig.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file %s: %w", sc.authConfig.PrivateKeyPath, err)
	}

	var signer ssh.Signer
	if sc.authConfig.PrivateKeyPassword != "" {
		// 带密码的私钥
		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(sc.authConfig.PrivateKeyPassword))
	} else {
		// 无密码的私钥
		signer, err = ssh.ParsePrivateKey(key)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return ssh.PublicKeys(signer), nil
}

// buildAgentAuth 构建 SSH Agent 认证方式
func (sc *SSHClient) buildAgentAuth() (ssh.AuthMethod, error) {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil, fmt.Errorf("SSH_AUTH_SOCK environment variable not set")
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH agent: %w", err)
	}

	agentClient := agent.NewClient(conn)
	return ssh.PublicKeysCallback(agentClient.Signers), nil
}
