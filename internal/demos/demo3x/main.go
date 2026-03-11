// demo3x: Static YAML config basic usage example
// Shows how to load YAML format config with NewYamlStatic
//
// demo3x: 静态 YAML 配置基础用法示例
// 展示如何使用 NewYamlStatic 加载 YAML 格式配置
package main

import (
	"fmt"

	"github.com/go-kratos/kratos/v2/config"
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
