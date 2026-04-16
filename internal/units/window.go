package units

import (
	"github.com/pll177/ruyipage-go/internal/base"
	"github.com/pll177/ruyipage-go/internal/bidi"
	"github.com/pll177/ruyipage-go/internal/support"
	"time"
)

type windowOwner interface {
	BrowserDriver() *base.BrowserBiDiDriver
	BaseTimeout() time.Duration
	RunJS(script string, args ...any) (any, error)
	RunJSExpr(expression string, args ...any) (any, error)
	WindowSize() map[string]int
}

// WindowManager 提供当前页面所在浏览器窗口的常用操作。
type WindowManager struct {
	owner windowOwner
}

// NewWindowManager 创建窗口管理器。
func NewWindowManager(owner windowOwner) *WindowManager {
	return &WindowManager{owner: owner}
}

// Maximize 最大化窗口。
func (w *WindowManager) Maximize() error {
	if w == nil || w.owner == nil {
		return nil
	}
	windowID, _ := w.currentWindowID()
	if windowID != "" {
		if _, err := bidi.SetClientWindowState(w.owner.BrowserDriver(), windowID, "maximized", nil, nil, nil, nil, w.owner.BaseTimeout()); err == nil {
			return nil
		}
	}
	_, err := w.owner.RunJS("window.moveTo(0, 0); window.resizeTo(screen.width, screen.height)")
	return err
}

// Minimize 最小化窗口；BiDi 不支持时静默返回。
func (w *WindowManager) Minimize() error {
	if w == nil || w.owner == nil {
		return nil
	}
	windowID, _ := w.currentWindowID()
	if windowID == "" {
		return nil
	}
	_, _ = bidi.SetClientWindowState(w.owner.BrowserDriver(), windowID, "minimized", nil, nil, nil, nil, w.owner.BaseTimeout())
	return nil
}

// Fullscreen 切换到全屏。
func (w *WindowManager) Fullscreen() error {
	if w == nil || w.owner == nil {
		return nil
	}
	windowID, _ := w.currentWindowID()
	if windowID != "" {
		if _, err := bidi.SetClientWindowState(w.owner.BrowserDriver(), windowID, "fullscreen", nil, nil, nil, nil, w.owner.BaseTimeout()); err == nil {
			return nil
		}
	}
	_, err := w.owner.RunJS("if (document.documentElement.requestFullscreen) { return document.documentElement.requestFullscreen(); }")
	return err
}

// Normal 恢复普通窗口状态。
func (w *WindowManager) Normal() error {
	if w == nil || w.owner == nil {
		return nil
	}
	windowID, _ := w.currentWindowID()
	if windowID != "" {
		if _, err := bidi.SetClientWindowState(w.owner.BrowserDriver(), windowID, "normal", nil, nil, nil, nil, w.owner.BaseTimeout()); err == nil {
			return nil
		}
	}
	_, err := w.owner.RunJS("window.resizeTo(1280, 800)")
	return err
}

// SetSize 设置窗口尺寸。
func (w *WindowManager) SetSize(width int, height int) error {
	if w == nil || w.owner == nil {
		return nil
	}
	windowID, _ := w.currentWindowID()
	if windowID != "" {
		if _, err := bidi.SetClientWindowState(w.owner.BrowserDriver(), windowID, "normal", &width, &height, nil, nil, w.owner.BaseTimeout()); err == nil {
			return nil
		}
	}
	_, err := w.owner.RunJS("window.resizeTo(arguments[0], arguments[1])", width, height)
	return err
}

// SetPosition 设置窗口位置。
func (w *WindowManager) SetPosition(x int, y int) error {
	if w == nil || w.owner == nil {
		return nil
	}
	windowID, _ := w.currentWindowID()
	if windowID != "" {
		if _, err := bidi.SetClientWindowState(w.owner.BrowserDriver(), windowID, "normal", nil, nil, &x, &y, w.owner.BaseTimeout()); err == nil {
			return nil
		}
	}
	_, err := w.owner.RunJS("window.moveTo(arguments[0], arguments[1])", x, y)
	return err
}

// Center 将窗口居中；仅当宽高都大于 0 时同时调整尺寸。
func (w *WindowManager) Center(width int, height int) error {
	if w == nil || w.owner == nil {
		return nil
	}
	if width > 0 && height > 0 {
		if err := w.SetSize(width, height); err != nil {
			return err
		}
	}

	screen, err := w.owner.RunJSExpr("([window.screen.availWidth, window.screen.availHeight])")
	if err != nil {
		return err
	}
	values, ok := screen.([]any)
	if !ok || len(values) < 2 {
		return support.NewRuyiPageError("读取屏幕尺寸失败", nil)
	}
	windowSize := w.owner.WindowSize()
	targetX := (intFromActionAny(values[0]) - windowSize["width"]) / 2
	targetY := (intFromActionAny(values[1]) - windowSize["height"]) / 2
	if targetX < 0 {
		targetX = 0
	}
	if targetY < 0 {
		targetY = 0
	}
	return w.SetPosition(targetX, targetY)
}

// Info 返回当前窗口信息快照。
func (w *WindowManager) Info() map[string]any {
	if w == nil || w.owner == nil {
		return map[string]any{}
	}
	_, info := w.currentWindowID()
	return cloneActionRow(info)
}

func (w *WindowManager) currentWindowID() (string, map[string]any) {
	if w == nil || w.owner == nil {
		return "", nil
	}
	result, err := bidi.GetClientWindows(w.owner.BrowserDriver(), w.owner.BaseTimeout())
	if err != nil {
		return "", nil
	}
	windows, _ := result["clientWindows"].([]map[string]any)
	if len(windows) == 0 {
		values, _ := result["clientWindows"].([]any)
		for _, value := range values {
			if mapped, ok := value.(map[string]any); ok {
				windows = append(windows, cloneActionRow(mapped))
			}
		}
	}
	if len(windows) == 0 {
		return "", nil
	}
	info := cloneActionRow(windows[0])
	return stringifyActionValue(info["clientWindow"]), info
}
