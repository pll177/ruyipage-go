package ruyipage

import "github.com/pll177/ruyipage-go/internal/support"

type (
	// RuyiPageError 是全部公开错误的根错误类型。
	RuyiPageError = support.RuyiPageError
	// ElementNotFoundError 表示元素未找到。
	ElementNotFoundError = support.ElementNotFoundError
	// ElementLostError 表示元素引用失效。
	ElementLostError = support.ElementLostError
	// ContextLostError 表示浏览上下文已销毁。
	ContextLostError = support.ContextLostError
	// BiDiError 表示 BiDi 协议错误。
	BiDiError = support.BiDiError
	// PageDisconnectedError 表示页面连接断开。
	PageDisconnectedError = support.PageDisconnectedError
	// JavaScriptError 表示 JavaScript 执行错误。
	JavaScriptError = support.JavaScriptError
	// BrowserConnectError 表示浏览器连接失败。
	BrowserConnectError = support.BrowserConnectError
	// BrowserLaunchError 表示浏览器启动失败。
	BrowserLaunchError = support.BrowserLaunchError
	// AlertExistsError 表示存在阻塞操作的对话框。
	AlertExistsError = support.AlertExistsError
	// WaitTimeoutError 表示等待超时。
	WaitTimeoutError = support.WaitTimeoutError
	// NoRectError 表示元素没有可用矩形区域。
	NoRectError = support.NoRectError
	// CanNotClickError 表示元素当前不可点击。
	CanNotClickError = support.CanNotClickError
	// LocatorError 表示定位器语法或格式错误。
	LocatorError = support.LocatorError
)
