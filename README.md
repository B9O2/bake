# bake

_基于配置文件的go项目打包工具。_支持替换依赖包名、docker编译等功能。

## 安装
`go install github.com/B9O2/bake`
## 流程

1. 在项目文件夹下执行 `bake init` 初始化项目（生成*RECIPE.toml文件*）
2. 在命令行中使用 `bake`进行编译或 `bake [recipes]`顺序运行配置文件中的多个配置

## RECIPE.toml

*RECIPE.toml*是指导 `bake`如何编译的配置文件。为了使每份配置尽量简短，选择了面向行的*toml*格式。

最简单的*RECIPE.toml*只需要两行

```toml
[recipes.default]
entrance="./"
```

一份完整的*RECIPE.toml*看起来像这样:

```toml
# bake
[recipes.default] #默认编译配置，会在不提供配置名时默认执行
# 必要参数
entrance="./" #编译入口，此处指定为当前目录
# 非必要参数
output="./bin" #编译输出目录
all_platform.all_arch.replace.text."google" = "apple" #替换代码文件内容
all_platform.all_arch.replace.dependency."gitlab.google.com" = "gitlab.apple.com" #替换依赖名
#|___平台___||__架构__||__操作__||_操作子项_|

[recipes.my_recipe]
entrance="./"
desc="演示配置"
output="./my_bin"
pairs = ["darwin/arm64","windows/386"] #编译darwin/arm64和windows/386
```

其中*default*配置看起来稍显复杂，设置了以编译目标为单位对原始项目的内容进行替换。

例如：*darwin平台的arm64架构替换文本"google"为"apple"*写成配置就是 `darwin.arm64.replace.text."google" = "apple"`，而示例中的配置则以*all_platform*与*all_arch*指代了全部内置平台与架构。

⚠️*如果您的目标平台架构未被内置在bake，则您需要额外对其配置。*

## 命令

- `bake` 寻找当前目录下的RECIPE.toml，运行其中的default配置
- `bake [recipes]` 寻找当前目录下的RECIPE.toml，运行指定配置。例如 `bake my_recipe`执行*my_recipe*，而 `bake default my_recipe`则会按顺序执行default与my_recipe两个配置

⚠️*如果编译过程被中断，需要您手动清除**临时目录***

## 更多配置选项

### 编译选项

```toml
[recipes.builder_test]
entrance="./"
output="./build_bin"
all_platform.all_arch.builder.path="go"#使用环境变量中的go
all_platform.all_arch.builder.args=["-trimpath","-ldflags","-w -s"]#bake默认参数
all_platform.all_arch.builder.env.ENV_NAME="ENV_ALUE"#设置环境变量
```

### Docker编译

bake可以远程连接Docker进行编译。

```toml
#指定容器编译
[recipes.docker_test]
entrance="./"
output="./build_by_docker_bin"
all_platform.all_arch.docker.host="local" #使用本地docker
all_platform.all_arch.docker.container="2a7c6546eea74b" #容器ID或容器名
```

```toml
#指定镜像编译
[recipes.docker_test]
entrance="./"
output="./build_by_docker_bin"
all_platform.all_arch.docker.host="local" #使用本地docker
all_platform.all_arch.docker.image="golang" #镜像名
```

⚠️*镜像不存在bake会自动下载镜像并启动临时容器编译，编译后**会移除**容器，**不会移除**镜像*

### SSH编译

bake可以连接远程服务器进行编译。

```toml
#SSH密钥认证
[recipes.ssh_test]
entrance="./"
output="./build_by_ssh_bin"
all_platform.all_arch.ssh.host="192.168.1.100" #远程服务器地址
all_platform.all_arch.ssh.user="builder" #SSH用户名
all_platform.all_arch.ssh.private_key_path="~/.ssh/id_rsa" #私钥路径
```

```toml
#SSH密码认证
[recipes.ssh_test]
entrance="./"
output="./build_by_ssh_bin"
all_platform.all_arch.ssh.host="build-server.com" #远程服务器地址
all_platform.all_arch.ssh.user="builder" #SSH用户名
all_platform.all_arch.ssh.password="your_password" #SSH密码
```

```toml
#SSH Agent认证
[recipes.ssh_test]
entrance="./"
output="./build_by_ssh_bin"
all_platform.all_arch.ssh.host="build-server.com" #远程服务器地址
all_platform.all_arch.ssh.user="builder" #SSH用户名
#不需要指定密码或私钥，自动使用SSH Agent
```

⚠️*认证方式自动检测：提供私钥路径时使用私钥认证，提供密码时使用密码认证，否则使用SSH Agent。远程临时目录会在编译完成后自动清理*

### 输出

详细配置输出。

```toml
[recipes.output_test]
entrance="./"
output="./test" #输出目录的根目录
darwin.arm64.output.path="hello/world/CoolApp_dawin_arm64" #设定输出目录下的相对路径
```

💡*如果想要将每个文件都以绝对路径输出在不同位置，请设置 `output="/"`*

#### ZIP压缩输出

bake支持将编译结果自动打包成ZIP文件。

```toml
#基本ZIP压缩
[recipes.zip_test]
entrance="./"
output="./build"
all_platform.all_arch.output.zip.source="myapp" #要压缩的文件或目录
all_platform.all_arch.output.zip.dest="myapp.zip" #ZIP文件名
```

```toml
#带密码的ZIP压缩
[recipes.zip_password]
entrance="./"
output="./build"
all_platform.all_arch.output.zip.source="myapp" #要压缩的文件
all_platform.all_arch.output.zip.dest="myapp_secure.zip" #ZIP文件名
all_platform.all_arch.output.zip.password="secret123" #ZIP密码
```

```toml
#不同平台不同ZIP配置
[recipes.zip_platform]
entrance="./"
output="./release"
linux.amd64.output.zip.source="myapp" #Linux文件
linux.amd64.output.zip.dest="myapp_linux_amd64.zip"
windows.amd64.output.zip.source="myapp.exe" #Windows文件
windows.amd64.output.zip.dest="myapp_windows_amd64.zip"
```

⚠️*ZIP压缩在编译完成后自动执行，原始文件会保留。如果设置了密码，请妥善保管*

#### SFTP上传

bake支持将编译结果自动上传到远程服务器。

```toml
#SFTP密钥认证上传
[recipes.sftp_key]
entrance="./"
output="./build"
all_platform.all_arch.output.ssh.host="upload-server.com" #上传服务器地址
all_platform.all_arch.output.ssh.user="uploader" #SSH用户名
all_platform.all_arch.output.ssh.private_key_path="~/.ssh/upload_key" #私钥路径
all_platform.all_arch.output.ssh.source="myapp" #本地文件
all_platform.all_arch.output.ssh.dest="/opt/apps/myapp_latest" #远程路径
```

```toml
#SFTP密码认证上传
[recipes.sftp_password]
entrance="./"
output="./build"
all_platform.all_arch.output.ssh.host="192.168.1.200" #上传服务器地址
all_platform.all_arch.output.ssh.user="deploy" #SSH用户名
all_platform.all_arch.output.ssh.password="deploy123" #SSH密码
all_platform.all_arch.output.ssh.source="myapp" #本地文件
all_platform.all_arch.output.ssh.dest="/var/www/releases/" #远程目录
```

```toml
#不同平台上传到不同位置
[recipes.sftp_platform]
entrance="./"
output="./release"
linux.amd64.output.ssh.host="linux-server.com"
linux.amd64.output.ssh.user="deploy"
linux.amd64.output.ssh.source="myapp"
linux.amd64.output.ssh.dest="/opt/linux/myapp"
windows.amd64.output.ssh.host="windows-server.com"
windows.amd64.output.ssh.user="deploy"
windows.amd64.output.ssh.source="myapp.exe"
windows.amd64.output.ssh.dest="/opt/windows/myapp.exe"
```

⚠️*SFTP上传在编译完成及ZIP打包后自动执行，支持密钥、密码与Agent认证。上传失败不会影响编译结果*

## Todo
- [ ] 更实用的命令行参数
- [ ] Remote Vendor
- [ ] 输出
  - [x] ZIP压缩(可设置密码)
  - [x] SFTP
  - [ ] S3
- [ ] 直接执行命令(?)
- [ ] Cel表达式(?)
- [X] docker编译
  - [X] 指定容器编译
  - [X] 指定镜像，自动下载启动编译
- [x] ssh编译
- [X] 更好的异常处理
- [X] 指定go二进制程序(可替换支持混淆的编译工具)
