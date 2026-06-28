[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/kratos-config/release.yml?branch=main&label=BUILD)](https://github.com/yylego/kratos-config/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/kratos-config)](https://pkg.go.dev/github.com/yylego/kratos-config)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/kratos-config/main.svg)](https://coveralls.io/github/yylego/kratos-config?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.25%2B-lightgrey.svg)](https://github.com/yylego/kratos-config)
[![GitHub Release](https://img.shields.io/github/release/yylego/kratos-config.svg)](https://github.com/yylego/kratos-config/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/kratos-config)](https://goreportcard.com/report/github.com/yylego/kratos-config)

# kratos-config

In-memory config source implementation to use with [Kratos](https://github.com/go-kratos/kratos) framework.

---

<!-- TEMPLATE (EN) BEGIN: LANGUAGE NAVIGATION -->

## CHINESE README

[中文说明](README.zh.md)

<!-- TEMPLATE (EN) END: LANGUAGE NAVIGATION -->

## Features

- **Byte Slice Source**: Load config from byte data instead of files
- **Multiple Formats**: Support JSON, YAML, XML, Proto plus additional formats
- **Dynamic Updates**: Runtime config updates with watch mechanism (DataSource)
- **Static Mode**: Lightweight static config without watch overhead (DataStatic)
- **Type Safe**: Complete support to Kratos config.Source interface
- **Thread Safe**: Concurrent-safe config updates

## Installation

```bash
go get github.com/yylego/kratos-config/configkratos
```

## Quick Start

### Static Config (Recommended when config is immutable)

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

	fmt.Printf("App: %s, Version: %s\n", app.Name, app.Version)
}
```

⬆️ **Source:** [Source](internal/demos/demo1x/main.go)

### Dynamic Config (Support runtime updates)

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
	fmt.Printf("Start with port: %d\n", server.Port)

	must.Done(cfg.Watch("port", func(key string, value config.Value) {
		var port int
		must.Done(value.Scan(&port))
		fmt.Printf("Port update: %d\n", port)
	}))

	time.Sleep(time.Second)
	newData := []byte(`{"port":9090}`)
	must.Done(source.Update(newData))

	time.Sleep(time.Second)
}
```

⬆️ **Source:** [Source](internal/demos/demo2x/main.go)

### YAML Config (Static mode with YAML format)

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

⬆️ **Source:** [Source](internal/demos/demo3x/main.go)

## API Reference

### DataSource (Dynamic Config)

**Create Source**

```go
// Generic function with format option
source := configkratos.NewDataSource(data, "json")

// Convenient functions to specific formats
jsonSource := configkratos.NewJsonSource(data)
yamlSource := configkratos.NewYamlSource(data)
```

**Update Config**

```go
newData := []byte(`{"key":"new-value"}`)
err := source.Update(newData)
```

### DataStatic (Static Config)

**Create Static Source**

```go
// Generic function with format parameter
source := configkratos.NewDataStatic(data, "yaml")

// Convenient functions to specific formats
jsonSource := configkratos.NewJsonStatic(data)
yamlSource := configkratos.NewYamlStatic(data)
```

## Supported Formats

- `json` - JSON format
- `yaml` - YAML format
- `xml` - XML format
- `proto` - Protocol Buffers
- `form` - URL encoded form
- Other formats registered in Kratos encoding

## Use Cases

- **Unit Testing**: Mock config data without creating temp files
- **Embedded Config**: Package config data into binary
- **Dynamic Config**: Runtime config updates from remote sources
- **Config Aggregation**: Combine multiple config sources

## Design Philosophy

**DataSource vs DataStatic**

- **DataSource**: Full-featured with watch mechanism, suitable to scenarios requiring runtime updates
- **DataStatic**: Lightweight without watch overhead, suitable to scenarios with immutable config

The dual implementation gives flexibility in choosing the best approach to different scenarios.

<!-- TEMPLATE (EN) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-09-26 07:39:27.188023 +0000 UTC -->

## 📄 License

MIT License. See [LICENSE](LICENSE).

---

## 🤝 Contributing

Contributions are welcome! Report bugs, suggest features, and contribute code:

- 🐛 **Found a mistake?** Open an issue on GitHub with reproduction steps
- 💡 **Have a feature idea?** Create an issue to discuss the suggestion
- 📖 **Documentation confusing?** Report it so we can improve
- 🚀 **Need new features?** Share the use cases to help us understand requirements
- ⚡ **Performance issue?** Help us optimize through reporting slow operations
- 🔧 **Configuration problem?** Ask questions about complex setups
- 📢 **Follow project progress?** Watch the repo to get new releases and features
- 🌟 **Success stories?** Share how this package improved the workflow
- 💬 **Feedback?** We welcome suggestions and comments

---

## 🔧 Development

New code contributions, follow this process:

1. **Fork**: Fork the repo on GitHub (using the webpage UI).
2. **Clone**: Clone the forked project (`git clone https://github.com/yourname/kratos-config.git`).
3. **Navigate**: Navigate to the cloned project (`cd kratos-config`)
4. **Branch**: Create a feature branch (`git checkout -b feature/xxx`).
5. **Code**: Implement the changes with comprehensive tests
6. **Testing**: (Golang project) Ensure tests pass (`go test ./...`) and follow Go code style conventions
7. **Documentation**: Update documentation to support client-facing changes and use significant commit messages
8. **Stage**: Stage changes (`git add .`)
9. **Commit**: Commit changes (`git commit -m "Add feature xxx"`) ensuring backward compatible code
10. **Push**: Push to the branch (`git push origin feature/xxx`).
11. **PR**: Open a merge request on GitHub (on the GitHub webpage) with detailed description.

Please ensure tests pass and include relevant documentation updates.

---

## 🌟 Support

Welcome to contribute to this project via submitting merge requests and reporting issues.

**Project Support:**

- ⭐ **Give GitHub stars** if this project helps you
- 🤝 **Share with teammates** and (golang) programming friends
- 📝 **Write tech blogs** about development tools and workflows - we provide content writing support
- 🌟 **Join the ecosystem** - committed to supporting open source and the (golang) development scene

**Have Fun Coding with this package!** 🎉🎉🎉

<!-- TEMPLATE (EN) END: STANDARD PROJECT FOOTER -->

---

## GitHub Stars

[![Stargazers](https://starchart.cc/yylego/kratos-config.svg?variant=adaptive)](https://starchart.cc/yylego/kratos-config)
