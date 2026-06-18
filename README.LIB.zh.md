# yogo - JSON数组操作 Go 类库文档
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/ymc-github/yogo)](https://goreportcard.com/report/github.com/ymc-github/yogo)

`ipkg/yogo` 是底层纯 Go 工具类库，无第三方重型依赖，提供 JSON 文件读写、嵌套路径解析、数组批量增删、命令行参数解析完整能力。上层 CLI 程序 `cmd/yogo` 完全基于该库封装。

## ✨ 类库特性
- 无副作用纯函数设计，输入输出隔离，便于单元测试
- 完整支持多层嵌套 JSON 路径解析，兼容数组下标、对象键特殊语法
- 内置数组标准化处理：去重、排序、批量新增、批量删除
- 文件路径智能区分绝对/相对路径，规避标准库 `filepath.Join` 根路径丢失问题
- 支持文件缺失自动初始化、缺失节点自动创建
- 内置调试打印、格式化 JSON 输出工具函数
- 参数解析模块分离，可嵌入任意 Go 程序，不强制依赖 CLI

## 📦 引入类库
```bash
go get github.com/ymc-github/yogo/ipkg/yogo
```
```go
import "github.com/ymc-github/yogo/ipkg/yogo"
```

## 📚 模块结构说明
1. `config.go`：命令行参数结构体、参数解析、帮助/版本/特性打印
2. `jsonutil.go`：JSON 文件读写、字符串分割、数组核心处理逻辑
3. `pathparse.go`：嵌套路径解析、路径遍历、定位目标数组与父对象

## 1. config.go API
### Config 配置结构体
```go
type Config struct {
	Version       bool
	Help          bool
	PrintFeature  bool
	Workspace     string
	File          string
	DryRun        bool
	Debug         bool
	DefaultText   string
	Name          string
	Include       string
	Exclude       string
	Sep           string
	Ns            string
	NsSep         string
	CreateMissing bool
	Api           string
}
```

#### ParseFlags
解析全部命令行入参，绑定长短参数别名，返回完整配置实例
```go
func ParseFlags() Config
```

#### CheckAPI
校验子命令合法性，仅允许 `edit-json-arr`，非法值返回错误
```go
func (cfg Config) CheckAPI() error
```

#### CheckRequired
校验必填参数 `--name`，缺失返回报错
```go
func (cfg Config) CheckRequired() error
```

#### PrintHelp
输出完整 CLI 使用文档、路径语法、参数对照表、业务示例
```go
func PrintHelp()
```

#### PrintFeature
打印全部功能特性清单，用于 `--print-feature`
```go
func PrintFeature()
```

#### PrintDebugConfig
格式化输出配置 JSON，调试模式使用
```go
func PrintDebugConfig(cfg Config)
```

## 2. jsonutil.go API
### LoadJSON
读取 JSON 文件；文件不存在时使用 defaultText 初始化空对象
```go
func LoadJSON(path string, defaultText string, out *map[string]interface{}) error
```

### SaveJSON
带缩进格式化写入 JSON 文件，文件权限 0644
```go
func SaveJSON(path string, root map[string]interface{}) error
```

### PrintJSON
格式化打印 map 类型 JSON，dryrun 预览专用
```go
func PrintJSON(root map[string]interface{})
```

### SplitBySep
按分隔符切割字符串，自动去除首尾空白、过滤空片段
```go
func SplitBySep(s, sep string) []string
```

### ProcessArray
数组核心处理函数，实现增删、去重、排序一体化逻辑
```go
func ProcessArray(arr []interface{}, include, exclude []string) []interface{}
```
执行流程：
1. 遍历原始数组存入集合自动去重
2. 删除所有 `exclude` 匹配的字符串元素
3. 追加全部 `include` 新增元素
4. 字符串升序排序，转回 `[]interface{}` 返回

## 3. pathparse.go API
### PathSegment 路径单元结构体
存储单一层级路径信息，区分普通键、数组下标、对象自定义键
```go
type PathSegment struct {
	Key    string
	ArrIdx *int    // [number] 数字数组下标
	ObjKey *string // [text] 对象自定义键名
}
```

### ParseNsSegments
按分隔符拆分嵌套路径，解析 `[数字]` / `[字符串]` 特殊语法
```go
func ParseNsSegments(nsPath, nsSep string) ([]PathSegment, error)
```

### GetTargetArray
逐层遍历嵌套路径，定位目标数组；返回**父对象 Map、原始数组、错误**
```go
func GetTargetArray(root map[string]interface{}, nsPath, nsSep, arrName string, createMissing bool) (map[string]interface{}, []interface{}, error)
```
核心逻辑：
1. 逐层解析 ns 路径，节点不存在时根据 createMissing 创建空对象
2. 支持数组下标取值、对象自定义键取值
3. 定位最终父对象，读取目标 `--name` 数组
4. 数组不存在且开启自动创建时，生成空数组写入父对象

## 🚀 类库代码调用示例
### 示例1：修改根层级数组 daemon.json insecure-registries
```go
package main

import (
	"fmt"
	"github.com/ymc-github/yogo/ipkg/yogo"
)

func main() {
	filePath := "/etc/docker/daemon.json"
	var root map[string]interface{}
	// 加载JSON，文件不存在默认空对象
	err := yogo.LoadJSON(filePath, "{}", &root)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}

	// 定位根层级数组，ns传空字符串
	parentObj, rawArr, err := yogo.GetTargetArray(root, "", ".", "insecure-registries", true)
	if err != nil {
		fmt.Printf("定位数组失败: %v\n", err)
		return
	}

	// 分割新增、删除列表
	incList := yogo.SplitBySep("mirror.test.com,docker.ustc.edu.cn", ",")
	excList := yogo.SplitBySep("mirrors.sohu.com", ",")

	// 处理数组
	newArr := yogo.ProcessArray(rawArr, incList, excList)
	// 写回父对象，更新JSON数据
	parentObj["insecure-registries"] = newArr

	// 持久化写入文件
	err = yogo.SaveJSON(filePath, root)
	if err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
		return
	}
	fmt.Println("配置修改完成")
}
```

### 示例2：操作多层嵌套 JSON 数组
```go
package main

import (
	"fmt"
	"github.com/ymc-github/yogo/ipkg/yogo"
)

func main() {
	var root map[string]interface{}
	_ = yogo.LoadJSON("app.json", "{}", &root)

	// 嵌套路径 config.docker[0].mirror，目标数组 urls
	parent, arr, err := yogo.GetTargetArray(root, "config.docker[0].mirror", ".", "urls", true)
	if err != nil {
		fmt.Println("路径解析失败：", err)
		return
	}

	newArr := yogo.ProcessArray(arr, []string{"new.mirror.com"}, []string{})
	parent["urls"] = newArr

	_ = yogo.SaveJSON("app.json", root)
	fmt.Println("嵌套数组修改完成")
}
```

## 🧪 单元测试
```bash
# 执行类库完整单元测试
go test -v ./ipkg/yogo
# 输出测试覆盖率
go test -cover ./ipkg/yogo
```

## 设计规范
1. 所有底层函数纯输入输出，无全局变量、无副作用
2. 路径解析与文件读写完全解耦，可单独复用
3. 数组处理逻辑与 CLI 参数隔离，可嵌入任意业务程序
4. 错误分层返回，便于上层自定义提示文案

## 📄 开源协议
MIT License or Apache 2.0 License

## 📮 仓库地址
https://github.com/ymc-github/yogo