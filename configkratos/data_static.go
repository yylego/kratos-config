package configkratos

import (
	"context"

	"github.com/go-kratos/kratos/v3/config"
	"github.com/yylego/erero"
	"github.com/yylego/must"
)

// DataStatic is static data source without dynamic update support
// It fits best when loading configurations that remain fixed at runtime
// More lightweight than DataSource since it skips watch mechanism
//
// DataStatic 是不支持动态更新的静态数据源
// 适合加载在运行时保持不变的配置
// 比 DataSource 更轻量，因为不需要监听机制
type DataStatic struct {
	data   []byte // Raw config data in specified format // 指定格式的原始配置数据
	format string // Data format: json, yaml, xml, proto, etc. // 数据格式：json、yaml、xml、proto 等
}

// NewDataStatic creates a static config source from byte slice data with specified format
// Supports format: "json", "yaml", "form", "xml", "proto" and extra formats in Kratos encoding
// The created source is lightweight without watch mechanism, suitable when config remains immutable
//
// Usage Example:
//
//	data := []byte(`{"app":"demo","version":"1.0"}`)
//	source := configkratos.NewDataStatic(data, "json")
//	cfg := config.New(config.WithSource(source))
//	cfg.Load()
//
// NewDataStatic 从字节数据创建静态配置源，需要指定数据格式
// 支持格式："json"、"yaml"、"form"、"xml"、"proto" 以及在 Kratos encoding 中注册的额外格式
// 创建的配置源是轻量级的，不包含监听机制，适合配置保持不变的场景
//
// 使用示例：
//
//	data := []byte(`{"app":"demo","version":"1.0"}`)
//	source := configkratos.NewDataStatic(data, "json")
//	cfg := config.New(config.WithSource(source))
//	cfg.Load()
func NewDataStatic(data []byte, format string) *DataStatic {
	return &DataStatic{
		data:   data,
		format: must.Nice(format),
	}
}

// NewYamlStatic creates a YAML format static config source
//
// NewYamlStatic 创建 YAML 格式的静态配置源
func NewYamlStatic(data []byte) *DataStatic {
	return NewDataStatic(data, "yaml")
}

// NewJsonStatic creates a JSON format static config source
//
// NewJsonStatic 创建 JSON 格式的静态配置源
func NewJsonStatic(data []byte) *DataStatic {
	return NewDataStatic(data, "json")
}

// Load implements config.Source interface, returns static config data
//
// Load 实现 config.Source 接口，返回静态配置数据
func (p *DataStatic) Load() ([]*config.KeyValue, error) {
	cfgItem := &config.KeyValue{
		Key:    "config",
		Value:  p.data,
		Format: p.format,
	}
	return []*config.KeyValue{cfgItem}, nil
}

// Watch implements config.Source interface, returns a no-op watch mechanism without updates
// The watch blocks awaiting Stop() invocation
//
// Watch 实现 config.Source 接口，返回一个永不更新的虚拟监听器
// 监听器会阻塞直到调用 Stop()
func (p *DataStatic) Watch() (config.Watcher, error) {
	w, err := NewStaticWatcher()
	if err != nil {
		return nil, erero.Wro(err)
	}
	return w, nil
}

// StaticWatcher is a no-op watch mechanism that won't emit config updates
// It satisfies config.Watcher interface but provides passive watch pattern
// Reference: https://github.com/go-kratos/kratos/blob/dbd7664eff951d1e13b291d95a226f1e9101c8e1/config/env/watcher.go#L11
//
// StaticWatcher 是一个永不发出配置更新的虚拟监听器
// 满足 config.Watcher 接口但提供空操作的监听行为
// 参考：https://github.com/go-kratos/kratos/blob/dbd7664eff951d1e13b291d95a226f1e9101c8e1/config/env/watcher.go#L11
type StaticWatcher struct {
	ctx context.Context    // Context to manage watch lifecycle // 上下文，控制监听器生命周期
	can context.CancelFunc // Stop function to halt the watch // 取消函数，停止监听器
}

// NewStaticWatcher builds a new static watch mechanism that won't emit updates
//
// NewStaticWatcher 创建一个永不发出更新的静态监听器
func NewStaticWatcher() (config.Watcher, error) {
	ctx, can := context.WithCancel(context.Background())
	return &StaticWatcher{
		ctx: ctx,
		can: can,
	}, nil
}

// Next implements config.Watcher interface, blocks awaiting Stop() invocation
// Returns context issue when unblocked
//
// Next 实现 config.Watcher 接口，阻塞直到调用 Stop()
// 解除阻塞时总是返回上下文错误
func (w *StaticWatcher) Next() ([]*config.KeyValue, error) {
	<-w.ctx.Done()
	return nil, w.ctx.Err()
}

// Stop implements config.Watcher interface, cancels context to unblock Next()
//
// Stop 实现 config.Watcher 接口，取消上下文以解除 Next() 的阻塞
func (w *StaticWatcher) Stop() error {
	w.can()
	return nil
}
