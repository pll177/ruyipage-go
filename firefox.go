package ruyipage

import (
	"os/exec"
	"time"

	internalbrowser "ruyipage-go/internal/browser"
	"ruyipage-go/internal/config"
)

// Firefox 是公开浏览器生命周期对象。
type Firefox struct {
	inner *internalbrowser.Firefox
}

func newFirefoxFromInner(inner *internalbrowser.Firefox) *Firefox {
	if inner == nil {
		return nil
	}
	return &Firefox{inner: inner}
}

// NewFirefox 按 FirefoxOptions 启动或连接浏览器。
func NewFirefox(options *FirefoxOptions) (*Firefox, error) {
	raw := config.NewFirefoxOptions()
	if options != nil {
		raw = options.raw().Clone()
	}

	instance, err := internalbrowser.NewFirefox(raw)
	if err != nil {
		return nil, err
	}
	return newFirefoxFromInner(instance), nil
}

// Address 返回当前浏览器地址。
func (f *Firefox) Address() string {
	if f == nil || f.inner == nil {
		return ""
	}
	return f.inner.Address()
}

// SessionID 返回当前 session id。
func (f *Firefox) SessionID() string {
	if f == nil || f.inner == nil {
		return ""
	}
	return f.inner.SessionID()
}

// Options 返回配置副本。
func (f *Firefox) Options() *FirefoxOptions {
	if f == nil || f.inner == nil {
		return nil
	}
	return &FirefoxOptions{cfg: f.inner.Options()}
}

// Process 返回受管 Firefox 进程。
func (f *Firefox) Process() *exec.Cmd {
	if f == nil || f.inner == nil {
		return nil
	}
	return f.inner.Process()
}

// TabsCount 返回当前 tab 数量。
func (f *Firefox) TabsCount() int {
	if f == nil || f.inner == nil {
		return 0
	}
	return f.inner.TabsCount()
}

// TabIDs 返回当前 tab id 列表副本。
func (f *Firefox) TabIDs() []string {
	if f == nil || f.inner == nil {
		return []string{}
	}
	return f.inner.TabIDs()
}

// WindowHandles 返回窗口信息副本。
func (f *Firefox) WindowHandles() []map[string]any {
	if f == nil || f.inner == nil {
		return []map[string]any{}
	}
	return f.inner.WindowHandles()
}

// Cookies 返回浏览器级可见 Cookie 列表。
func (f *Firefox) Cookies(allInfo bool) ([]CookieInfo, error) {
	if f == nil || f.inner == nil {
		return []CookieInfo{}, nil
	}
	return f.inner.Cookies(allInfo)
}

// Close 使用默认参数关闭浏览器。
func (f *Firefox) Close() error {
	if f == nil || f.inner == nil {
		return nil
	}
	return f.inner.Close()
}

// Quit 关闭浏览器、结束 session 并清理进程。
func (f *Firefox) Quit(timeout time.Duration, force bool) error {
	if f == nil || f.inner == nil {
		return nil
	}
	return f.inner.Quit(timeout, force)
}
