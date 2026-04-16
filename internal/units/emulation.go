package units

import (
	"time"

	"github.com/pll177/ruyipage-go/internal/base"
	"github.com/pll177/ruyipage-go/internal/bidi"
)

type emulationOwner interface {
	ContextID() string
	BrowserDriver() *base.BrowserBiDiDriver
	BaseTimeout() time.Duration
	ApplyUserAgentOverride(userAgent string) error
	DefaultUserContext() string
}

// EmulationManager 提供设备、环境与浏览器行为仿真能力。
type EmulationManager struct {
	owner emulationOwner
}

// NewEmulationManager 创建仿真管理器。
func NewEmulationManager(owner emulationOwner) *EmulationManager {
	return &EmulationManager{owner: owner}
}

// SetGeolocation 设置地理位置；accuracy<=0 时默认使用 100 米。
func (m *EmulationManager) SetGeolocation(latitude float64, longitude float64, accuracy float64) error {
	if m == nil || m.owner == nil {
		return nil
	}
	if accuracy <= 0 {
		accuracy = 100
	}
	_, err := bidi.SetGeolocationOverride(
		m.owner.BrowserDriver(),
		&latitude,
		&longitude,
		&accuracy,
		m.currentContexts(),
		m.owner.BaseTimeout(),
	)
	return err
}

// ClearGeolocation 清除地理位置覆盖。
func (m *EmulationManager) ClearGeolocation() error {
	if m == nil || m.owner == nil {
		return nil
	}
	_, err := bidi.SetGeolocationOverride(
		m.owner.BrowserDriver(),
		nil,
		nil,
		nil,
		m.currentContexts(),
		m.owner.BaseTimeout(),
	)
	return err
}

// SetTimezone 设置时区。
func (m *EmulationManager) SetTimezone(timezoneID string) error {
	if m == nil || m.owner == nil {
		return nil
	}
	_, err := bidi.SetTimezoneOverride(
		m.owner.BrowserDriver(),
		timezoneID,
		m.currentContexts(),
		m.owner.BaseTimeout(),
	)
	return err
}

// SetLocale 设置 locale；支持 string 或 []string。
func (m *EmulationManager) SetLocale(locales any) error {
	if m == nil || m.owner == nil {
		return nil
	}
	_, err := bidi.SetLocaleOverride(
		m.owner.BrowserDriver(),
		locales,
		m.currentContexts(),
		m.owner.BaseTimeout(),
	)
	return err
}

// SetScreenOrientation 设置屏幕方向。
func (m *EmulationManager) SetScreenOrientation(orientationType string, angle int) error {
	if m == nil || m.owner == nil {
		return nil
	}
	_, err := bidi.SetScreenOrientationOverride(
		m.owner.BrowserDriver(),
		orientationType,
		angle,
		m.currentContexts(),
		m.owner.BaseTimeout(),
	)
	return err
}

// SetScreenSize 设置屏幕宽高与 DPR。
func (m *EmulationManager) SetScreenSize(width int, height int, devicePixelRatio *float64) error {
	if m == nil || m.owner == nil {
		return nil
	}
	_, err := bidi.SetScreenSettingsOverride(
		m.owner.BrowserDriver(),
		&width,
		&height,
		devicePixelRatio,
		m.currentContexts(),
		m.owner.BaseTimeout(),
	)
	return err
}

// SetUserAgent 设置 UA；标准命令未实现时回退到 preload script 注入。
func (m *EmulationManager) SetUserAgent(userAgent string, platform string) error {
	if m == nil || m.owner == nil {
		return nil
	}
	result, err := bidi.SetUserAgentOverride(
		m.owner.BrowserDriver(),
		userAgent,
		platform,
		m.currentContexts(),
		m.owner.BaseTimeout(),
	)
	if err != nil {
		return err
	}
	if result != nil {
		return nil
	}
	return m.owner.ApplyUserAgentOverride(userAgent)
}

// SetNetworkOffline 切换离线/在线状态；当前浏览器未实现时返回 supported=false。
func (m *EmulationManager) SetNetworkOffline(enabled bool) (bool, error) {
	if m == nil || m.owner == nil {
		return false, nil
	}
	result, err := bidi.SetNetworkConditions(
		m.owner.BrowserDriver(),
		enabled,
		m.currentContexts(),
		m.owner.BaseTimeout(),
	)
	return result != nil, err
}

// SetTouchEnabled 切换触摸模拟；scope 支持 context/global/user_context。
func (m *EmulationManager) SetTouchEnabled(enabled bool, maxTouchPoints int, scope string) (bool, error) {
	if m == nil || m.owner == nil {
		return false, nil
	}
	if maxTouchPoints <= 0 {
		maxTouchPoints = 1
	}

	var value *int
	if enabled {
		value = &maxTouchPoints
	}

	var (
		result map[string]any
		err    error
	)
	switch scope {
	case "", "context":
		result, err = bidi.SetTouchOverride(
			m.owner.BrowserDriver(),
			value,
			m.currentContexts(),
			nil,
			m.owner.BaseTimeout(),
		)
	case "global":
		result, err = bidi.SetTouchOverride(
			m.owner.BrowserDriver(),
			value,
			nil,
			nil,
			m.owner.BaseTimeout(),
		)
	case "user_context":
		userContext := m.owner.DefaultUserContext()
		var target any
		if userContext != "" {
			target = []string{userContext}
		}
		result, err = bidi.SetTouchOverride(
			m.owner.BrowserDriver(),
			value,
			nil,
			target,
			m.owner.BaseTimeout(),
		)
	default:
		result, err = bidi.SetTouchOverride(
			m.owner.BrowserDriver(),
			value,
			m.currentContexts(),
			nil,
			m.owner.BaseTimeout(),
		)
	}
	return result != nil, err
}

// SetJavaScriptEnabled 切换 JavaScript 执行开关。
func (m *EmulationManager) SetJavaScriptEnabled(enabled bool) (bool, error) {
	if m == nil || m.owner == nil {
		return false, nil
	}
	result, err := bidi.SetScriptingEnabled(
		m.owner.BrowserDriver(),
		enabled,
		m.currentContexts(),
		m.owner.BaseTimeout(),
	)
	return result != nil, err
}

// SetScrollbarType 设置滚动条类型。
func (m *EmulationManager) SetScrollbarType(scrollbarType string) (bool, error) {
	if m == nil || m.owner == nil {
		return false, nil
	}
	result, err := bidi.SetScrollbarTypeOverride(
		m.owner.BrowserDriver(),
		scrollbarType,
		m.currentContexts(),
		m.owner.BaseTimeout(),
	)
	return result != nil, err
}

// SetForcedColorsMode 设置强制颜色模式主题。
func (m *EmulationManager) SetForcedColorsMode(mode string) (bool, error) {
	if m == nil || m.owner == nil {
		return false, nil
	}
	result, err := bidi.SetForcedColorsModeThemeOverride(
		m.owner.BrowserDriver(),
		mode,
		m.currentContexts(),
		m.owner.BaseTimeout(),
	)
	return result != nil, err
}

// SetBypassCSP 设置 CSP 绕过；当前浏览器未实现时返回 supported=false。
func (m *EmulationManager) SetBypassCSP(enabled bool) (bool, error) {
	if m == nil || m.owner == nil {
		return false, nil
	}
	result, err := bidi.SetBypassCSPOverride(
		m.owner.BrowserDriver(),
		enabled,
		m.currentContexts(),
		m.owner.BaseTimeout(),
	)
	return result != nil, err
}

// SetFocusEmulation 切换焦点模拟；当前浏览器未实现时返回 supported=false。
func (m *EmulationManager) SetFocusEmulation(enabled bool) (bool, error) {
	if m == nil || m.owner == nil {
		return false, nil
	}
	result, err := bidi.SetFocusEmulation(
		m.owner.BrowserDriver(),
		enabled,
		m.currentContexts(),
		m.owner.BaseTimeout(),
	)
	return result != nil, err
}

// ApplyMobilePreset 一次性应用常见移动端环境参数。
func (m *EmulationManager) ApplyMobilePreset(
	userAgent string,
	width int,
	height int,
	devicePixelRatio float64,
	orientationType string,
	angle int,
	locale string,
	timezoneID string,
	touch *bool,
) map[string]any {
	support := map[string]any{
		"user_agent":  true,
		"screen":      true,
		"orientation": true,
		"touch":       nil,
		"locale":      nil,
		"timezone":    nil,
	}
	if m == nil || m.owner == nil {
		return support
	}

	if width <= 0 {
		width = 390
	}
	if height <= 0 {
		height = 844
	}
	if devicePixelRatio <= 0 {
		devicePixelRatio = 3.0
	}
	if orientationType == "" {
		orientationType = "portrait-primary"
	}
	if touch == nil {
		defaultTouch := true
		touch = &defaultTouch
	}

	if err := m.SetUserAgent(userAgent, ""); err != nil {
		support["user_agent"] = false
	}

	if err := m.SetScreenSize(width, height, &devicePixelRatio); err != nil {
		support["screen"] = false
	}
	if _, err := bidi.SetViewport(
		m.owner.BrowserDriver(),
		m.owner.ContextID(),
		&width,
		&height,
		&devicePixelRatio,
		m.owner.BaseTimeout(),
	); err != nil {
		support["screen"] = false
	}

	if err := m.SetScreenOrientation(orientationType, angle); err != nil {
		support["orientation"] = false
	}

	if touch != nil {
		supported, err := m.SetTouchEnabled(*touch, 1, "context")
		if err != nil {
			support["touch"] = false
		} else {
			support["touch"] = supported
		}
	}

	if locale != "" {
		if err := m.SetLocale(locale); err != nil {
			support["locale"] = false
		} else {
			support["locale"] = true
		}
	}

	if timezoneID != "" {
		if err := m.SetTimezone(timezoneID); err != nil {
			support["timezone"] = false
		} else {
			support["timezone"] = true
		}
	}

	return support
}

func (m *EmulationManager) currentContexts() any {
	if m == nil || m.owner == nil {
		return nil
	}
	contextID := m.owner.ContextID()
	if contextID == "" {
		return nil
	}
	return []string{contextID}
}
