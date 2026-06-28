// Package configkratos provides in-memory config source implementations to use with Kratos framework
// Implements both dynamic (DataSource) and static (DataStatic) config sources from byte slice data
// Supports runtime config updates through watch mechanism and multiple encoding formats
// Designed as lightweight alternative to file-based config sources in testing and embedded scenarios
//
// configkratos 为 Kratos 框架提供内存配置源实现
// 从字节切片数据实现动态(DataSource)和静态(DataStatic)配置源
// 通过监听机制支持运行时配置更新，支持多种编码格式
// 设计为测试和嵌入式场景中基于文件的配置源的轻量级替代方案
package configkratos

import (
	"context"
	"errors"
	"sync"

	"github.com/go-kratos/kratos/v3/config"
	"github.com/yylego/erero"
	"github.com/yylego/must"
)

var (
	// ErrWatchRepeat indicates Watch() was invoked multiple times on same source
	// ErrWatchRepeat 表示在同一个 source 上多次调用 Watch()
	ErrWatchRepeat = errors.New("configkratos: Watch() invoked multiple times")

	// ErrWatchNotSet indicates Update() was invoked before Watch()
	// ErrWatchNotSet 表示在 Watch() 之前调用了 Update()
	ErrWatchNotSet = errors.New("configkratos: Update() invoked before Watch()")

	// ErrWatchClosed indicates watcher has been closed
	// ErrWatchClosed 表示监听器已经关闭
	ErrWatchClosed = errors.New("configkratos: config watcher has been closed")
)

// DataSource implements config.Source interface to support dynamic config updates from byte slice data
// It provides Watch mechanism to listen config changes and update them at runtime
//
// DataSource 实现 config.Source 接口，支持从字节数据动态更新配置
// 提供 Watch 机制，可以监听配置变化并在运行时更新
type DataSource struct {
	data    []byte         // Raw config data in specified format // 指定格式的原始配置数据
	format  string         // Data format: json, yaml, xml, proto, etc. // 数据格式：json、yaml、xml、proto 等
	watcher *ConfigWatcher // Watcher instance, deferred init on first Watch() invocation // 监听器实例，首次调用 Watch() 时延迟初始化
}

// NewDataSource creates a config source from byte slice data with specified format
// Supports format: "json", "yaml", "form", "xml", "proto" and extra formats in Kratos encoding
// The created source supports runtime config updates through Update() method and watch mechanism
// Reference: https://github.com/go-kratos/kratos/blob/main/encoding/encoding.go
//
// Usage Example:
//
//	data := []byte(`{"port":8080}`)
//	source := configkratos.NewDataSource(data, "json")
//	cfg := config.New(config.WithSource(source))
//	cfg.Load()
//	// Update config at runtime
//	source.Update([]byte(`{"port":9090}`))
//
// NewDataSource 从字节数据创建配置源，需要指定数据格式
// 支持格式："json"、"yaml"、"form"、"xml"、"proto" 以及在 Kratos encoding 中注册的额外格式
// 创建的配置源支持通过 Update() 方法和监听机制进行运行时配置更新
// 参考：https://github.com/go-kratos/kratos/blob/main/encoding/encoding.go
//
// 使用示例：
//
//	data := []byte(`{"port":8080}`)
//	source := configkratos.NewDataSource(data, "json")
//	cfg := config.New(config.WithSource(source))
//	cfg.Load()
//	// 运行时更新配置
//	source.Update([]byte(`{"port":9090}`))
func NewDataSource(data []byte, format string) *DataSource {
	return &DataSource{
		data:    data,
		format:  must.Nice(format),
		watcher: nil, // Deferred init on Watch() to prevent resource waste // 在 Watch() 时延迟初始化，避免资源浪费
	}
}

// NewYamlSource creates a YAML format config source
//
// NewYamlSource 创建 YAML 格式的配置源
func NewYamlSource(data []byte) *DataSource {
	return NewDataSource(data, "yaml")
}

// NewJsonSource creates a JSON format config source
//
// NewJsonSource 创建 JSON 格式的配置源
func NewJsonSource(data []byte) *DataSource {
	return NewDataSource(data, "json")
}

// Load implements config.Source interface, returns the starting config data
//
// Load 实现 config.Source 接口，返回初始配置数据
func (p *DataSource) Load() ([]*config.KeyValue, error) {
	cfgItem := &config.KeyValue{
		Key:    "config",
		Value:  p.data,
		Format: p.format,
	}
	return []*config.KeyValue{cfgItem}, nil
}

// Watch implements config.Source interface, creates the watch mechanism to track config changes
// Returns issue when Watch() gets invoked multiple times on the same source
//
// Watch 实现 config.Source 接口，创建监听器以监控配置变化
// 当在同一个 source 上多次调用 Watch() 时会返回错误
func (p *DataSource) Watch() (config.Watcher, error) {
	if p.watcher != nil {
		return nil, erero.WithMessage(ErrWatchRepeat, "source binding conflict detected") // Watch() gets invoked once, issue happens when same source passes to config.WithSource() twice // 正常情况下只会调用一次，当同一个 source 两次传给 config.WithSource() 时会出现此错误
	}
	watcher, err := NewConfigWatcher(p.format)
	if err != nil {
		return nil, erero.Wro(err)
	}
	p.watcher = watcher
	return watcher, nil
}

// Update pushes new config data to the watch mechanism, triggers config reload
// Must invoke Watch() before Update(), otherwise returns NOT-WATCHING issue
// The update is thread-safe and notifies watchers about config changes
//
// Usage Example:
//
//	source := configkratos.NewJsonSource([]byte(`{"port":8080}`))
//	cfg := config.New(config.WithSource(source))
//	cfg.Load()
//	// Watch config changes
//	cfg.Watch("port", func(key string, value config.Value) {
//	    fmt.Println("port changed")
//	})
//	// Update config at runtime
//	source.Update([]byte(`{"port":9090}`))
//
// Update 推送新的配置数据到监听器，触发配置重新加载
// 必须先调用 Watch() 再调用 Update()，否则返回 NOT-WATCHING 错误
// 更新操作是线程安全的，会通知所有监听器配置已变化
//
// 使用示例：
//
//	source := configkratos.NewJsonSource([]byte(`{"port":8080}`))
//	cfg := config.New(config.WithSource(source))
//	cfg.Load()
//	// 监听配置变化
//	cfg.Watch("port", func(key string, value config.Value) {
//	    fmt.Println("端口变化")
//	})
//	// 运行时更新配置
//	source.Update([]byte(`{"port":9090}`))
func (p *DataSource) Update(data []byte) error {
	if p.watcher == nil {
		return erero.WithMessage(ErrWatchNotSet, "must invoke Watch() before Update()")
	}
	if err := p.watcher.update(data); err != nil {
		return erero.Wro(err)
	}
	return nil
}

// ConfigWatcher implements config.Watcher interface to watch config changes
// Uses buffered chan to receive config updates without blocking
//
// ConfigWatcher 实现 config.Watcher 接口以监听配置变化
// 使用缓冲通道接收配置更新而不阻塞
type ConfigWatcher struct {
	dataChan   chan []byte        // Buffered chan to receive config data updates // 缓冲通道，接收配置数据更新
	format     string             // Data format when encoding/decoding // 数据格式，用于编解码
	ctx        context.Context    // Context to manage watch lifecycle // 上下文，控制监听器生命周期
	can        context.CancelFunc // Stop function to halt the watch // 取消函数，停止监听器
	mutex      *sync.Mutex        // Sync lock to protect concurrent access // 互斥锁，保护并发访问
	isWatching bool               // Status flag, true when watch is active // 标志位，表示监听器是否活跃
}

// NewConfigWatcher builds a new config watch mechanism with the given format
//
// NewConfigWatcher 创建指定格式的配置监听器
func NewConfigWatcher(format string) (*ConfigWatcher, error) {
	ctx, can := context.WithCancel(context.Background())
	res := &ConfigWatcher{
		dataChan:   make(chan []byte, 1), // Buffered chan to avoid blocking // 缓冲通道，避免阻塞
		format:     must.Nice(format),
		ctx:        ctx,
		can:        can,
		mutex:      &sync.Mutex{},
		isWatching: true,
	}
	return res, nil
}

// update pushes new config data to the chan, invoked from DataSource.Update()
// Thread-safe with sync lock protection
//
// update 推送新的配置数据到通道，由 DataSource.Update() 调用
// 使用互斥锁保证线程安全
func (w *ConfigWatcher) update(data []byte) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.isWatching {
		return erero.WithMessage(ErrWatchClosed, "cannot update data when watcher closed")
	}

	w.dataChan <- data
	return nil
}

// Next implements config.Watcher interface, blocks awaiting next config update (unless context cancels)
// Returns context.Canceled when the watch mechanism gets stopped
//
// Next 实现 config.Watcher 接口，阻塞直到下一次配置更新或上下文取消
// 当监听器停止时返回 context.Canceled
func (w *ConfigWatcher) Next() ([]*config.KeyValue, error) {
	select {
	case <-w.ctx.Done():
		return nil, w.ctx.Err() // Get context.Canceled when context ends // 上下文取消时返回 context.Canceled
	case data, ok := <-w.dataChan:
		if !ok {
			return nil, context.Canceled // Get context.Canceled when dataChan gets closed // dataChan 关闭时返回 context.Canceled
		}
		cfgItem := &config.KeyValue{
			Key:    "config", // The static config name as mandated in Kratos // Kratos 配置系统要求的固定键名
			Value:  data,
			Format: w.format,
		}
		return []*config.KeyValue{cfgItem}, nil
	}
}

// Stop implements config.Watcher interface, halts the watch mechanism and cleans up
// Thread-safe with sync lock protection
//
// Stop 实现 config.Watcher 接口，停止监听器并释放资源
// 使用互斥锁保证线程安全
func (w *ConfigWatcher) Stop() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.isWatching {
		return erero.WithMessage(ErrWatchClosed, "Stop() func invoked, cannot close again")
	}
	w.isWatching = false

	w.can()           // Activate ctx.Done() to unblock Next() // 触发 ctx.Done() 以解除 Next() 的阻塞
	close(w.dataChan) // Shut down the chan to cleanup resources // 关闭通道，确保资源清理
	return nil
}
