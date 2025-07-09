package targets

type Target interface {
	Info() string
	//InitAndConnect 连接远程编译目标
	InitAndConnect(hashTag string) error
	// CopyShadowProjectTo 复制影子项目路径到远程目标
	CopyShadowProjectTo(src string) error //返回错误
	// BuildExec 执行编译命令
	BuildExec(cmd string, args []string, env map[string]string) ([]byte, []byte, error)
	// CopyFileBack 复制文件到本地指定输出目录
	CopyFileBack(src, dest string) error
	// Close 清理远程目标
	Close() error
}

type BaseTarget struct {
	platform, arch string
	shadowPath     string
}

func NewBaseTarget(platform, arch string) *BaseTarget {
	return &BaseTarget{
		platform: platform,
		arch:     arch,
	}
}
