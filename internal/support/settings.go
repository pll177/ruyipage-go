package support

// SettingsValues 表示全局默认行为与超时基线。
type SettingsValues struct {
	RaiseWhenEleNotFound  bool
	RaiseWhenClickFailed  bool
	RaiseWhenWaitFailed   bool
	SingletonTabObj       bool
	BiDiTimeout           float64
	BrowserConnectTimeout float64
	ElementFindTimeout    float64
	PageLoadTimeout       float64
	ScriptTimeout         float64
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
