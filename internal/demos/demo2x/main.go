// demo2x: Dynamic config with runtime updates example
// Shows how to use Watch mechanism and Update config at runtime
//
// demo2x: 动态配置运行时更新示例
// 展示如何使用 Watch 机制和运行时更新配置
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
