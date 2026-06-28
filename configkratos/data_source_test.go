package configkratos_test

import (
	"testing"
	"time"

	"github.com/go-kratos/kratos/v3/config"
	"github.com/stretchr/testify/require"
	"github.com/yylego/kratos-config/configkratos"
	"github.com/yylego/must"
	"github.com/yylego/neatjson/neatjsonm"
	"github.com/yylego/neatjson/neatjsons"
	"github.com/yylego/rese"
)

// TestNewJsonSource tests basic JSON config source loading and scanning
// Checks that config data can get loaded from JSON byte slice and scanned to struct
//
// TestNewJsonSource 测试基础 JSON 配置源的加载和扫描
// 验证可以从 JSON 字节数据加载配置并扫描到结构体
func TestNewJsonSource(t *testing.T) {
	// Define test config structure
	// 定义测试配置结构
	type Account struct {
		Username string `json:"username"`
		Nickname string `json:"nickname"`
	}

	// Create JSON source from struct data
	// 从结构体数据创建 JSON 配置源
	jsonSource := configkratos.NewJsonSource(neatjsonm.B(&Account{
		Username: "abc",
		Nickname: "123",
	}))

	// Create Kratos config instance with JSON source
	// 使用 JSON 配置源创建 Kratos 配置实例
	c := config.New(
		config.WithSource(
			jsonSource, // Load from JSON config source // 从 JSON 配置源加载
		),
	)
	defer rese.F0(c.Close)

	// Load config data from source
	// 从配置源加载数据
	must.Done(c.Load())

	// Scan config to struct and check the values
	// 将配置扫描到结构体并验证值
	account := &Account{}
	must.Done(c.Scan(account))
	t.Log("account:", neatjsons.S(account))
	require.Equal(t, "abc", account.Username)
	require.Equal(t, "123", account.Nickname)
}

// TestNewJsonSource_Update tests dynamic config updates with watch mechanism
// Checks that config changes get detected through watchers and incremental updates work as expected
// This test demonstrates the main feature of runtime config updates
//
// TestNewJsonSource_Update 测试动态配置更新和监听机制
// 验证监听器可以检测配置变化，并且部分更新能正确工作
// 此测试演示了运行时配置更新的核心特性
func TestNewJsonSource_Update(t *testing.T) {
	// Define nested config structures to test complex config scenarios
	// 定义嵌套配置结构以测试复杂配置场景
	type DatabaseConfig struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	type AuthConfig struct {
		APIKey  string `json:"api_key"`
		Timeout int    `json:"timeout"`
	}

	type ServerConfig struct {
		AppName  string         `json:"app_name"`
		Version  string         `json:"version"`
		Database DatabaseConfig `json:"database"`
		Auth     AuthConfig     `json:"auth"`
		Port     int            `json:"port"`
	}

	// Structures to test field-specific updates
	// 测试字段级别更新的结构
	type ConfigPortV2 struct {
		Port int `json:"port"`
	}

	type ConfigDatabaseV2 struct {
		Database DatabaseConfig `json:"database"`
	}

	type ConfigAuthV2 struct {
		Auth AuthConfig `json:"auth"`
	}

	// Create JSON source with initial complex config
	// 使用初始复杂配置创建 JSON 配置源
	jsonSource := configkratos.NewJsonSource(neatjsonm.B(&ServerConfig{
		AppName: "my-app",
		Version: "1.0.0",
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     3306,
			Username: "abc",
			Password: "xyz",
		},
		Auth: AuthConfig{
			APIKey:  "123",
			Timeout: 30,
		},
		Port: 8080,
	}))

	// Create Kratos config instance
	// 创建 Kratos 配置实例
	c := config.New(
		config.WithSource(
			jsonSource, // Load from JSON source // 从 JSON 配置源加载
		),
	)
	defer rese.F0(c.Close)

	// Load initial config
	// 加载初始配置
	must.Done(c.Load())

	// Setup watchers on various config keys to observe changes
	// Watch both top and nested keys to ensure watch mechanism functions at each depth
	//
	// 在各种配置键上设置监听器以观察变化
	// 监听顶级和嵌套键以验证监听机制在所有层级都能正常工作
	must.Done(c.Watch("app_name", func(s string, value config.Value) {
		t.Log("watch-changes:", s, value)
	}))
	must.Done(c.Watch("version", func(s string, value config.Value) {
		t.Log("watch-changes:", s, value)
	}))
	must.Done(c.Watch("port", func(s string, value config.Value) {
		t.Log("watch-changes:", s, value)
	}))
	must.Done(c.Watch("database", func(s string, value config.Value) {
		t.Log("watch-changes:", s, value)
	}))
	must.Done(c.Watch("database.port", func(s string, value config.Value) {
		t.Log("watch-changes:", s, value)
	}))
	must.Done(c.Watch("auth", func(s string, value config.Value) {
		t.Log("watch-changes:", s, value)
	}))
	must.Done(c.Watch("auth.timeout", func(s string, value config.Value) {
		t.Log("watch-changes:", s, value)
	}))

	// Verify initial config values
	// 验证初始配置值
	{
		res := ServerConfig{}
		must.Done(c.Scan(&res))
		t.Log("config:", neatjsons.S(res))
		require.Equal(t, "my-app", res.AppName)
		require.Equal(t, "1.0.0", res.Version)
		require.Equal(t, 8080, res.Port)
		require.Equal(t, "localhost", res.Database.Host)
		require.Equal(t, 3306, res.Database.Port)
		require.Equal(t, "abc", res.Database.Username)
		require.Equal(t, "xyz", res.Database.Password)
		require.Equal(t, "123", res.Auth.APIKey)
		require.Equal(t, 30, res.Auth.Timeout)
	}

	// Test incremental update: modify just port field
	// Ensure that other fields keep unchanged after this update
	//
	// 测试增量更新：仅修改 port 字段
	// 确保此次更新后其他字段保持不变
	must.Done(jsonSource.Update(neatjsonm.B(&ConfigPortV2{
		Port: 8081,
	})))
	time.Sleep(time.Millisecond * 100) // Wait to update propagate // 等待更新传播

	// Verify port changed while other fields unchanged
	// 验证端口已变化而其他字段未变
	{
		res := ServerConfig{}
		must.Done(c.Scan(&res)) // Must re-scan to get new data after update // 更新后必须重新扫描才能获取新数据
		t.Log("config:", neatjsons.S(res))
		require.Equal(t, "my-app", res.AppName)
		require.Equal(t, "1.0.0", res.Version)
		require.Equal(t, 8081, res.Port) // Port updated // 端口已更新
		require.Equal(t, "localhost", res.Database.Host)
		require.Equal(t, 3306, res.Database.Port)
		require.Equal(t, "abc", res.Database.Username)
		require.Equal(t, "xyz", res.Database.Password)
		require.Equal(t, "123", res.Auth.APIKey)
		require.Equal(t, 30, res.Auth.Timeout)
	}

	// Test nested config update: update database config
	// Verify that nested config can be updated while other top-level fields unchanged
	//
	// 测试嵌套配置更新：更新数据库配置
	// 验证嵌套配置可以更新而其他顶级字段保持不变
	must.Done(jsonSource.Update(neatjsonm.B(&ConfigDatabaseV2{
		Database: DatabaseConfig{
			Host:     "127.0.0.1",
			Port:     3307,
			Username: "aaa",
			Password: "xxx",
		},
	})))
	time.Sleep(time.Millisecond * 100) // Wait to update propagate // 等待更新传播

	// Verify database config updated while other fields unchanged
	// 验证数据库配置已更新而其他字段未变
	{
		res := ServerConfig{}
		must.Done(c.Scan(&res)) // Must re-scan to get new data // 必须重新扫描才能获取新数据
		t.Log("config:", neatjsons.S(res))
		require.Equal(t, "my-app", res.AppName)
		require.Equal(t, "1.0.0", res.Version)
		require.Equal(t, 8081, res.Port)                 // Port remains from previous update // 端口保持上次更新的值
		require.Equal(t, "127.0.0.1", res.Database.Host) // Database updated // 数据库已更新
		require.Equal(t, 3307, res.Database.Port)        // Database updated // 数据库已更新
		require.Equal(t, "aaa", res.Database.Username)   // Database updated // 数据库已更新
		require.Equal(t, "xxx", res.Database.Password)   // Database updated // 数据库已更新
		require.Equal(t, "123", res.Auth.APIKey)
		require.Equal(t, 30, res.Auth.Timeout)
	}

	// Test another nested config update: update auth config
	// Ensure that multiple incremental updates can be applied in sequence
	//
	// 测试另一个嵌套配置更新：更新认证配置
	// 确保多个增量更新可以顺序应用
	must.Done(jsonSource.Update(neatjsonm.B(&ConfigAuthV2{
		Auth: AuthConfig{
			APIKey:  "uvw",
			Timeout: 36000,
		},
	})))
	time.Sleep(time.Millisecond * 100) // Wait to update propagate // 等待更新传播

	// Verify auth config updated while preserving all previous updates
	// 验证认证配置已更新同时保留所有先前的更新
	{
		res := ServerConfig{}
		must.Done(c.Scan(&res)) // Must re-scan to get new data // 必须重新扫描才能获取新数据
		t.Log("config:", neatjsons.S(res))
		require.Equal(t, "my-app", res.AppName)
		require.Equal(t, "1.0.0", res.Version)
		require.Equal(t, 8081, res.Port)                 // Port remains from first update // 端口保持第一次更新的值
		require.Equal(t, "127.0.0.1", res.Database.Host) // Database remains from second update // 数据库保持第二次更新的值
		require.Equal(t, 3307, res.Database.Port)
		require.Equal(t, "aaa", res.Database.Username)
		require.Equal(t, "xxx", res.Database.Password)
		require.Equal(t, "uvw", res.Auth.APIKey)  // Auth updated // 认证已更新
		require.Equal(t, 36000, res.Auth.Timeout) // Auth updated // 认证已更新
	}
}
