# bake
_基于配置文件的go项目打包工具。_支持替换依赖包名、docker编译等功能。

## 流程
1. 在项目目录下放置*RECIPE.toml*文件
2. 在命令行中使用`bake`进行编译或`bake [recipes]`顺序运行配置文件中的多个配置

## RECIPE.toml

*RECIPE.toml*是指导`bake`如何编译的配置文件。为了使每份配置尽量简短，选择了面向行的*toml*格式。

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
output="./my_bin"
pairs = ["darwin/arm64","windows/386"] #编译darwin/arm64和windows/386
```

其中*default*配置看起来稍显复杂，Line3是以编译目标为单位对原始项目的内容进行替换。

例如：*darwin平台的arm64架构替换文本"google"为"apple"*写成配置就是`darwin.arm64.replace.text."google" = "apple"`，而示例中的配置则以*all_platform*与*all_arch*指代了全部内置平台与架构。

⚠️*如果您的目标平台架构未被内置在bake，则您需要额外对其配置。*

## 命令

- `bake` 寻找当前目录下的RECIPE.toml，运行其中的default配置
- `bake [recipes]` 寻找当前目录下的RECIPE.toml，运行指定配置。例如`bake my_recipe`执行*my_recipe*，而`bake default my_recipe`则会按顺序执行default与my_recipe两个配置

⚠️*如果编译过程被中断，需要您手动清除**临时目录***

## 更多配置选项

### 编译选项

```toml
[recipes.builder_test]
entrance="./"
output="./build_bin"
builder.path="go"#使用环境变量中的go
builder.args=["-trimpath","-ldflags","-w -s"]#bake默认参数
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

### 输出

详细配置输出。

```toml
[recipes.output_test]
entrance="./"
output="./test" #输出目录的根目录
darwin.arm64.output.path="hello/world/CoolApp_dawin_arm64" #设定输出目录下的相对路径
```

💡*如果想要将每个文件都以绝对路径输出在不同位置，请设置`output="/"`*

## Todo

- [ ] 输出
  - [ ] 压缩
- [ ] 直接执行命令
- [ ] docker编译
  - [x] 指定容器编译
  - [x] 指定镜像，自动下载启动编译
- [ ] ssh编译
- [x] 更好的异常处理
- [x] 指定go二进制程序(可替换支持混淆的编译工具)
