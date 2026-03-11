package configkratos_test

import (
	"testing"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/stretchr/testify/require"
	"github.com/yylego/kratos-config/configkratos"
	"github.com/yylego/must"
	"github.com/yylego/neatjson/neatjsonm"
	"github.com/yylego/neatjson/neatjsons"
	"github.com/yylego/rese"
	"gopkg.in/yaml.v3"
)

// TestNewYamlStatic tests YAML format static config source
// Checks that YAML data can be loaded and scanned without watch mechanism
// This tests the lightweight static config usage
//
// TestNewYamlStatic 测试 YAML 格式的静态配置源
// 验证 YAML 数据可以正确加载和扫描，无需监听机制
// 测试轻量级静态配置的使用场景
func TestNewYamlStatic(t *testing.T) {
	// Define config structure with yaml tags
	// 定义带 yaml 标签的配置结构
	type ConfigType struct {
		Username string `yaml:"username"`
		Nickname string `yaml:"nickname"`
	}

	// Create YAML static source by marshaling struct to YAML bytes
	// 通过将结构体序列化为 YAML 字节创建静态配置源
	yamlSource := configkratos.NewYamlStatic(rese.A1(yaml.Marshal(&ConfigType{
		Username: "abc",
		Nickname: "123",
	})))

	// Create Kratos config instance with YAML static source
	// 使用 YAML 静态配置源创建 Kratos 配置实例
	c := config.New(
		config.WithSource(
			yamlSource, // Load from YAML static source // 从 YAML 静态配置源加载
		),
	)
	defer rese.F0(c.Close)

	// Load static config
	// 加载静态配置
	must.Done(c.Load())

	// Scan and verify YAML config values
	// 扫描并验证 YAML 配置值
	account := &ConfigType{}
	must.Done(c.Scan(account))
	t.Log("account:", string(rese.A1(yaml.Marshal(account))))
	require.Equal(t, "abc", account.Username)
	require.Equal(t, "123", account.Nickname)
}

// TestNewJsonStatic tests JSON format static config source
// Checks that JSON data can be loaded and scanned without watch mechanism
// Compares with TestNewJsonSource to show difference between static and dynamic config
//
// TestNewJsonStatic 测试 JSON 格式的静态配置源
// 验证 JSON 数据可以正确加载和扫描，无需监听机制
// 与 TestNewJsonSource 对比，展示静态配置和动态配置的区别
func TestNewJsonStatic(t *testing.T) {
	// Define config structure with json tags
	// 定义带 json 标签的配置结构
	type Account struct {
		Username string `json:"username"`
		Nickname string `json:"nickname"`
	}

	// Create JSON static source from struct data
	// 从结构体数据创建 JSON 静态配置源
	jsonSource := configkratos.NewJsonStatic(neatjsonm.B(&Account{
		Username: "abc",
		Nickname: "123",
	}))

	// Create Kratos config instance with JSON static source
	// 使用 JSON 静态配置源创建 Kratos 配置实例
	c := config.New(
		config.WithSource(
			jsonSource, // Load from JSON static source // 从 JSON 静态配置源加载
		),
	)
	defer rese.F0(c.Close)

	// Load static config
	// 加载静态配置
	must.Done(c.Load())

	// Scan and verify JSON config values
	// 扫描并验证 JSON 配置值
	account := &Account{}
	must.Done(c.Scan(account))
	t.Log("account:", neatjsons.S(account))
	require.Equal(t, "abc", account.Username)
	require.Equal(t, "123", account.Nickname)
}
