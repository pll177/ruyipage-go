package support

// SettingsValues 表示全局默认行为与超时基线。
type SettingsValues struct {
	// RaiseWhenEleNotFound 控制查找元素失败时是否直接返回错误。
	// false 表示返回 NoneElement 风格的空元素，后续操作再按需判断。
	RaiseWhenEleNotFound bool
	// RaiseWhenClickFailed 控制点击失败时是否直接返回错误。
	// false 表示尽量兼容执行，不把点击失败立刻升级为强错误。
	RaiseWhenClickFailed bool
	// RaiseWhenWaitFailed 控制等待条件超时时是否直接返回错误。
	// false 表示等待失败按普通结果处理，由调用方决定是否中断流程。
	RaiseWhenWaitFailed bool
	// SingletonTabObj 控制同一个浏览上下文是否复用同一个 Tab/Page 对象。
	// true 可以减少重复包装对象，保持同一 tab 的状态引用稳定。
	SingletonTabObj bool
	// BiDiTimeout 是单次 WebDriver BiDi 命令的默认超时秒数。
	// 例如执行脚本、读取属性、网络命令等底层协议调用会使用该基线。
	BiDiTimeout float64
	// BrowserConnectTimeout 是启动或接管 Firefox 时连接 BiDi 端口的超时秒数。
	BrowserConnectTimeout float64
	// ElementFindTimeout 是查找元素的默认等待秒数。
	// 影响 Ele、Eles、Locator 等元素定位类能力的默认等待时长。
	ElementFindTimeout float64
	// PageLoadTimeout 是页面导航加载的默认超时秒数。
	// 影响 Get、Navigate 等页面跳转类能力。
	PageLoadTimeout float64
	// ScriptTimeout 是执行 JavaScript 的默认超时秒数。
	// 影响 RunJS、RunAsyncJS 等脚本执行类能力。
	ScriptTimeout float64
	// InterceptCompleteGraceTimeout 是 Intercept 开启时 Navigate(..., "complete")
	// 在 interactive 后继续等待 document.readyState=complete 的秒数。
	// 小于等于 0 表示不额外等待。
	InterceptCompleteGraceTimeout float64
	// InterceptCompleteStopLoading 控制上述等待超时后是否调用 window.stop()
	// 收掉持续加载请求，避免浏览器加载指示器一直转。
	InterceptCompleteStopLoading bool
}

// DefaultSettingsValues 返回与 Python 版对齐的默认设置快照。
func DefaultSettingsValues() SettingsValues {
	return SettingsValues{
		RaiseWhenEleNotFound:          false,
		RaiseWhenClickFailed:          false,
		RaiseWhenWaitFailed:           false,
		SingletonTabObj:               true,
		BiDiTimeout:                   DefaultBiDiTimeoutSeconds,
		BrowserConnectTimeout:         DefaultBrowserConnectTimeoutSeconds,
		ElementFindTimeout:            DefaultElementFindTimeoutSeconds,
		PageLoadTimeout:               DefaultPageLoadTimeoutSeconds,
		ScriptTimeout:                 DefaultScriptTimeoutSeconds,
		InterceptCompleteGraceTimeout: 3,
		InterceptCompleteStopLoading:  true,
	}
}

// Settings 是全局可变设置基线。
var Settings = func() *SettingsValues {
	defaults := DefaultSettingsValues()
	return &defaults
}()

// Snapshot 返回当前设置副本。
func (s *SettingsValues) Snapshot() SettingsValues {
	if s == nil {
		return DefaultSettingsValues()
	}
	return *s
}

// Restore 用给定快照恢复设置。
func (s *SettingsValues) Restore(snapshot SettingsValues) {
	if s == nil {
		return
	}
	*s = snapshot
}

// ResetSettings 重置为默认设置。
func ResetSettings() {
	Settings.Restore(DefaultSettingsValues())
}
