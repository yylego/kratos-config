// demo1x: Static JSON config basic usage example
// Shows how to load config from JSON bytes with NewJsonStatic
//
// demo1x: 静态 JSON 配置基础用法示例
// 展示如何使用 NewJsonStatic 从 JSON 字节数据加载配置
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
