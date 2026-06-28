[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/kratos-config/release.yml?branch=main&label=BUILD)](https://github.com/yylego/kratos-config/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/kratos-config)](https://pkg.go.dev/github.com/yylego/kratos-config)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/kratos-config/main.svg)](https://coveralls.io/github/yylego/kratos-config?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.25%2B-lightgrey.svg)](https://github.com/yylego/kratos-config)
[![GitHub Release](https://img.shields.io/github/release/yylego/kratos-config.svg)](https://github.com/yylego/kratos-config/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/kratos-config)](https://goreportcard.com/report/github.com/yylego/kratos-config)

# kratos-config

为 [Kratos](https://github.com/go-kratos/kratos) 框架提供的内存配置源实现。

---

<!-- TEMPLATE (ZH) BEGIN: LANGUAGE NAVIGATION -->

## 英文文档

[ENGLISH README](README.md)

<!-- TEMPLATE (ZH) END: LANGUAGE NAVIGATION -->

## 特性

- **字节数据源**: 从字节数据加载配置而非文件
- **多格式支持**: 支持 JSON、YAML、XML、Proto 等格式
- **动态更新**: 通过 watch 机制实现实时配置更新（DataSource）
- **静态模式**: 无 watch 开销的轻量静态配置（DataStatic）
- **类型安全**: 完整支持 Kratos config.Source 接口
- **线程安全**: 并发安全的配置更新

## 安装

```bash
go get github.com/yylego/kratos-config/configkratos
```

## 快速开始

### 静态配置（简单场景推荐）

```go
package main

import (
	"fmt"

	"github.com/go-kratos/kratos/v3/config"
	"github.com/yylego/kratos-config/configkratos"
	"github.com/yylego/must"
	"github.com/yylego/rese"
)

type AppConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func main() {
	jsonData := []byte(`{"name":"my-app","version":"1.0.0"}`)
	source := configkratos.NewJsonStatic(jsonData)

	cfg := config.New(config.WithSource(source))
	defer rese.F0(cfg.Close)

	must.Done(cfg.Load())

	var app AppConfig
	must.Done(cfg.Scan(&app))

	fmt.Printf("应用: %s, 版本: %s\n", app.Name, app.Version)
}
```

⬆️ **源码:** [源码](internal/demos/demo1x/main.go)

### 动态配置（支持运行时更新）

```go
package main

import (
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v3/config"
	"github.com/yylego/kratos-config/configkratos"
	"github.com/yylego/must"
	"github.com/yylego/rese"
)

type ServerConfig struct {
	Port int `json:"port"`
}

func main() {
	jsonData := []byte(`{"port":8080}`)
	source := configkratos.NewJsonSource(jsonData)

	cfg := config.New(config.WithSource(source))
	defer rese.F0(cfg.Close)

	must.Done(cfg.Load())

	var server ServerConfig
	must.Done(cfg.Scan(&server))
	fmt.Printf("启动端口: %d\n", server.Port)

	must.Done(cfg.Watch("port", func(key string, value config.Value) {
		var port int
		must.Done(value.Scan(&port))
		fmt.Printf("端口更新: %d\n", port)
	}))

	time.Sleep(time.Second)
	newData := []byte(`{"port":9090}`)
	must.Done(source.Update(newData))

	time.Sleep(time.Second)
}
```

⬆️ **源码:** [源码](internal/demos/demo2x/main.go)

### YAML 配置（YAML 格式的静态模式）

```go
package main

import (
	"fmt"

	"github.com/go-kratos/kratos/v3/config"
	"github.com/yylego/kratos-config/configkratos"
	"github.com/yylego/must"
	"github.com/yylego/rese"
)

type DatabaseConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func main() {
	yamlData := []byte(`host: localhost
port: 5432`)
	source := configkratos.NewYamlStatic(yamlData)

	cfg := config.New(config.WithSource(source))
	defer rese.F0(cfg.Close)

	must.Done(cfg.Load())

	var res DatabaseConfig
	must.Done(cfg.Scan(&res))

	fmt.Printf("Database: %s:%d\n", res.Host, res.Port)
}
```

⬆️ **源码:** [源码](internal/demos/demo3x/main.go)

## API 参考

### DataSource（动态配置）

**创建数据源**

```go
// 通用函数，指定格式参数
source := configkratos.NewDataSource(data, "json")

// 特定格式的便捷函数
jsonSource := configkratos.NewJsonSource(data)
yamlSource := configkratos.NewYamlSource(data)
```

**更新配置**

```go
newData := []byte(`{"key":"new-value"}`)
err := source.Update(newData)
```

### DataStatic（静态配置）

**创建静态源**

```go
// 通用函数，指定格式参数
source := configkratos.NewDataStatic(data, "yaml")

// 特定格式的便捷函数
jsonSource := configkratos.NewJsonStatic(data)
yamlSource := configkratos.NewYamlStatic(data)
```

## 支持的格式

- `json` - JSON 格式
- `yaml` - YAML 格式
- `xml` - XML 格式
- `proto` - Protocol Buffers
- `form` - URL 编码表单
- 其他在 Kratos encoding 中注册的格式

## 使用场景

- **单元测试**: 模拟配置数据而不创建临时文件
- **嵌入式配置**: 将配置数据打包到二进制文件
- **动态配置**: 从远程源实现运行时配置更新
- **配置聚合**: 组合多个配置源

## 设计理念

**DataSource vs DataStatic**

- **DataSource**: 功能完整，带 watch 机制，适合需要运行时更新的场景
- **DataStatic**: 轻量级，无 watch 开销，适合配置不变的场景

双重实现提供了灵活性，可以根据不同场景选择最佳方案。

<!-- TEMPLATE (ZH) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-09-26 07:39:27.188023 +0000 UTC -->

## 📄 许可证类型

MIT 许可证。详见 [LICENSE](LICENSE)。

---

## 🤝 项目贡献

非常欢迎贡献代码！报告 BUG、建议功能、贡献代码：

- 🐛 **发现问题？** 在 GitHub 上提交问题并附上重现步骤
- 💡 **功能建议？** 创建 issue 讨论您的想法
- 📖 **文档疑惑？** 报告问题，帮助我们改进文档
- 🚀 **需要功能？** 分享使用场景，帮助理解需求
- ⚡ **性能瓶颈？** 报告慢操作，帮助我们优化性能
- 🔧 **配置困扰？** 询问复杂设置的相关问题
- 📢 **关注进展？** 关注仓库以获取新版本和功能
- 🌟 **成功案例？** 分享这个包如何改善工作流程
- 💬 **反馈意见？** 欢迎提出建议和意见

---

## 🔧 代码贡献

新代码贡献，请遵循此流程：

1. **Fork**：在 GitHub 上 Fork 仓库（使用网页界面）
2. **克隆**：克隆 Fork 的项目（`git clone https://github.com/yourname/kratos-config.git`）
3. **导航**：进入克隆的项目（`cd kratos-config`）
4. **分支**：创建功能分支（`git checkout -b feature/xxx`）
5. **编码**：实现您的更改并编写全面的测试
6. **测试**：（Golang 项目）确保测试通过（`go test ./...`）并遵循 Go 代码风格约定
7. **文档**：为面向用户的更改更新文档，并使用有意义的提交消息
8. **暂存**：暂存更改（`git add .`）
9. **提交**：提交更改（`git commit -m "Add feature xxx"`）确保向后兼容的代码
10. **推送**：推送到分支（`git push origin feature/xxx`）
11. **PR**：在 GitHub 上打开 Merge Request（在 GitHub 网页上）并提供详细描述

请确保测试通过并包含相关的文档更新。

---

## 🌟 项目支持

非常欢迎通过提交 Merge Request 和报告问题来为此项目做出贡献。

**项目支持：**

- ⭐ **给予星标**如果项目对您有帮助
- 🤝 **分享项目**给团队成员和（golang）编程朋友
- 📝 **撰写博客**关于开发工具和工作流程 - 我们提供写作支持
- 🌟 **加入生态** - 致力于支持开源和（golang）开发场景

**祝你用这个包编程愉快！** 🎉🎉🎉

<!-- TEMPLATE (ZH) END: STANDARD PROJECT FOOTER -->

---

## GitHub 标星点赞

[![标星点赞](https://starchart.cc/yylego/kratos-config.svg?variant=adaptive)](https://starchart.cc/yylego/kratos-config)
