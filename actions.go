package ruyipage

import (
	"time"

	internalunits "github.com/pll177/ruyipage-go/internal/units"
)

// Actions 是页面级鼠标、键盘、滚轮动作链。
type Actions = internalunits.Actions

// TouchActions 是页面级触摸动作链。
type TouchActions = internalunits.TouchActions

// WindowManager 是当前页面对应窗口的状态管理器。
type WindowManager = internalunits.WindowManager

// NavigationEvent 是导航事件快照。
type NavigationEvent = internalunits.NavigationEvent

// NavigationTracker 是导航事件跟踪器。
type NavigationTracker = internalunits.NavigationTracker

// Actions 返回页面级动作链管理器。
func (p *FirefoxBase) Actions() *Actions {
	if p == nil || p.inner == nil {
		return nil
	}
	p.managersMu.Lock()
	defer p.managersMu.Unlock()
	if p.actions == nil {
		p.actions = internalunits.NewActions(p.inner)
	}
	return p.actions
}

// Touch 返回页面级触摸动作链管理器。
func (p *FirefoxBase) Touch() *TouchActions {
	if p == nil || p.inner == nil {
		return nil
	}
	p.managersMu.Lock()
	defer p.managersMu.Unlock()
	if p.touch == nil {
		p.touch = internalunits.NewTouchActions(p.inner)
	}
	return p.touch
}

// Window 返回窗口管理器。
func (p *FirefoxBase) Window() *WindowManager {
	if p == nil || p.inner == nil {
		return nil
	}
	p.managersMu.Lock()
	defer p.managersMu.Unlock()
	if p.window == nil {
		p.window = internalunits.NewWindowManager(p.inner)
	}
	return p.window
}

// Navigation 返回导航事件跟踪器。
func (p *FirefoxBase) Navigation() *NavigationTracker {
	if p == nil || p.inner == nil {
		return nil
	}
	p.managersMu.Lock()
	defer p.managersMu.Unlock()
	if p.navigation == nil {
		p.navigation = internalunits.NewNavigationTracker(
			func() internalunits.NavigationCommandDriver { return p.inner.BrowserDriver() },
			func() internalunits.NavigationCallbackDriver { return p.inner.Driver() },
			func() string { return p.inner.ContextID() },
			func() time.Duration { return p.inner.BaseTimeout() },
		)
	}
	return p.navigation
}
