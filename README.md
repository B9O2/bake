# bake
_基于配置文件的go项目打包工具_

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

⚠️*如果编译过程被中断，可能需要您手动清除**临时目录***

## Todo

- [ ] 打包压缩文件
- [ ] 直接执行命令
- [ ] 更好的异常处理
- [ ] 远程编译
