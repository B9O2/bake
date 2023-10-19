package remotes

type RemoteTarget interface {
	Info() string
	Connect() error
	CopyShadowProjectTo(src string) error //返回项目路径和错误
	BuildExec(cmd string, args []string) ([]byte, []byte, error)
	CopyFileBack(src, dest string) error
	Close()
}
