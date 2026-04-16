package base

import (
	"sync"
	"time"

	"github.com/pll177/ruyipage-go/internal/support"
)

const defaultDispatchTimeout = time.Duration(support.DefaultBiDiTimeoutSeconds) * time.Second

type commandRunner interface {
	Run(method string, params map[string]any, timeout time.Duration) (map[string]any, error)
}

// CommandDispatcher 为 driver 提供统一的同步命令入口。
type CommandDispatcher struct {
	mu             sync.RWMutex
	transport      commandRunner
	defaultTimeout time.Duration
}

// NewCommandDispatcher 创建命令派发器。
func NewCommandDispatcher(transport commandRunner) *CommandDispatcher {
	return &CommandDispatcher{
		transport:      transport,
		defaultTimeout: defaultDispatchTimeout,
	}
}

// SetDefaultTimeout 设置默认命令超时；非法值回退到固定默认值。
func (d *CommandDispatcher) SetDefaultTimeout(timeout time.Duration) {
	if d == nil {
		return
	}
	if timeout <= 0 {
		timeout = defaultDispatchTimeout
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	d.defaultTimeout = timeout
}

// Dispatch 通过底层 transport 发送命令并等待响应。
func (d *CommandDispatcher) Dispatch(method string, params map[string]any, timeout time.Duration) (map[string]any, error) {
	if params == nil {
		params = map[string]any{}
	}

	transport, resolvedTimeout := d.resolveState(timeout)
	if transport == nil {
		return nil, support.NewPageDisconnectedError("命令传输未初始化", nil)
	}

	return transport.Run(method, params, resolvedTimeout)
}

func (d *CommandDispatcher) resolveState(timeout time.Duration) (commandRunner, time.Duration) {
	if d == nil {
		if timeout <= 0 {
			timeout = defaultDispatchTimeout
		}
		return nil, timeout
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	if timeout <= 0 {
		timeout = d.defaultTimeout
	}
	if timeout <= 0 {
		timeout = defaultDispatchTimeout
	}

	return d.transport, timeout
}
