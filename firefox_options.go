package ruyipage

import "github.com/pll177/ruyipage-go/internal/config"

type (
	// FirefoxLoadMode 表示页面加载等待模式。
	FirefoxLoadMode = config.FirefoxLoadMode
	// FirefoxTimeouts 表示 FirefoxOptions 的超时配置。
	FirefoxTimeouts = config.FirefoxTimeouts
	// FirefoxQuickStartOptions 表示 QuickStart 的快捷预设输入。
	FirefoxQuickStartOptions = config.FirefoxQuickStartOptions
)

const (
	// LoadModeNormal 表示等待完整加载。
	LoadModeNormal FirefoxLoadMode = config.LoadModeNormal
	// LoadModeEager 表示等待 DOMContentLoaded。
	LoadModeEager FirefoxLoadMode = config.LoadModeEager
	// LoadModeNone 表示不等待页面加载。
	LoadModeNone FirefoxLoadMode = config.LoadModeNone
)

// FirefoxOptions 是浏览器启动、attach 与 examples 共用的公开配置对象。
type FirefoxOptions struct {
	cfg *config.FirefoxOptions
}

// NewFirefoxOptions 返回带默认值的 FirefoxOptions。
func NewFirefoxOptions() *FirefoxOptions {
	return &FirefoxOptions{cfg: config.NewFirefoxOptions()}
}

// DefaultFirefoxQuickStartOptions 返回与 Python quick_start 对齐的默认预设。
func DefaultFirefoxQuickStartOptions() FirefoxQuickStartOptions {
	return config.DefaultFirefoxQuickStartOptions()
}

// Clone 返回当前配置的深拷贝。
func (o *FirefoxOptions) Clone() *FirefoxOptions {
	return &FirefoxOptions{cfg: o.raw().Clone()}
}

// BrowserPath 返回浏览器可执行文件路径。
func (o *FirefoxOptions) BrowserPath() string {
	return o.raw().BrowserPath()
}

// Address 返回 host:port 形式的地址。
func (o *FirefoxOptions) Address() string {
	return o.raw().Address()
}

// Host 返回调试主机。
func (o *FirefoxOptions) Host() string {
	return o.raw().Host()
}

// Port 返回调试端口。
func (o *FirefoxOptions) Port() int {
	return o.raw().Port()
}

// ProfilePath 返回 profile 路径。
func (o *FirefoxOptions) ProfilePath() string {
	return o.raw().ProfilePath()
}

// UserDir 返回 user_dir/profile 路径。
func (o *FirefoxOptions) UserDir() string {
	return o.raw().UserDir()
}

// Arguments 返回启动参数副本。
func (o *FirefoxOptions) Arguments() []string {
	return o.raw().Arguments()
}

// Preferences 返回首选项副本。
func (o *FirefoxOptions) Preferences() map[string]any {
	return o.raw().Preferences()
}

// IsHeadless 返回是否启用无头模式。
func (o *FirefoxOptions) IsHeadless() bool {
	return o.raw().IsHeadless()
}

// DownloadPath 返回下载目录。
func (o *FirefoxOptions) DownloadPath() string {
	return o.raw().DownloadPath()
}

// LoadMode 返回加载模式。
func (o *FirefoxOptions) LoadMode() FirefoxLoadMode {
	return o.raw().LoadMode()
}

// Timeouts 返回超时配置副本。
func (o *FirefoxOptions) Timeouts() FirefoxTimeouts {
	return o.raw().Timeouts()
}

// IsExistingOnly 返回是否只连接已有浏览器。
func (o *FirefoxOptions) IsExistingOnly() bool {
	return o.raw().IsExistingOnly()
}

// RetryTimes 返回连接重试次数。
func (o *FirefoxOptions) RetryTimes() int {
	return o.raw().RetryTimes()
}

// RetryInterval 返回连接重试间隔。
func (o *FirefoxOptions) RetryInterval() float64 {
	return o.raw().RetryInterval()
}

// Proxy 返回代理地址。
func (o *FirefoxOptions) Proxy() string {
	return o.raw().Proxy()
}

// IsAutoPortEnabled 返回是否启用自动端口。
func (o *FirefoxOptions) IsAutoPortEnabled() bool {
	return o.raw().IsAutoPortEnabled()
}

// AutoPortStart 返回自动端口起始值。
func (o *FirefoxOptions) AutoPortStart() int {
	return o.raw().AutoPortStart()
}

// UserContext 返回默认 user context。
func (o *FirefoxOptions) UserContext() string {
	return o.raw().UserContext()
}

// FPFile 返回指纹配置文件路径。
func (o *FirefoxOptions) FPFile() string {
	return o.raw().FPFile()
}

// IsPrivateMode 返回是否启用私密模式。
func (o *FirefoxOptions) IsPrivateMode() bool {
	return o.raw().IsPrivateMode()
}

// UserPromptHandler 返回用户提示框处理策略副本。
func (o *FirefoxOptions) UserPromptHandler() map[string]string {
	return o.raw().UserPromptHandler()
}

// IsXPathPickerEnabled 返回是否启用 XPath picker。
func (o *FirefoxOptions) IsXPathPickerEnabled() bool {
	return o.raw().IsXPathPickerEnabled()
}

// IsActionVisualEnabled 返回是否启用鼠标行为可视化调试模式。
func (o *FirefoxOptions) IsActionVisualEnabled() bool {
	return o.raw().IsActionVisualEnabled()
}

// IsCloseBrowserOnExitEnabled 返回是否在当前 Go 进程退出时自动关闭由本进程启动的浏览器。
func (o *FirefoxOptions) IsCloseBrowserOnExitEnabled() bool {
	return o.raw().IsCloseBrowserOnExitEnabled()
}

// WithBrowserPath 设置浏览器可执行文件路径。
func (o *FirefoxOptions) WithBrowserPath(path string) *FirefoxOptions {
	o.raw().WithBrowserPath(path)
	return o
}

// WithAddress 设置调试地址。
func (o *FirefoxOptions) WithAddress(address string) *FirefoxOptions {
	o.raw().WithAddress(address)
	return o
}

// WithPort 设置调试端口。
func (o *FirefoxOptions) WithPort(port int) *FirefoxOptions {
	o.raw().WithPort(port)
	return o
}

// WithProfile 设置 Firefox profile 目录。
func (o *FirefoxOptions) WithProfile(path string) *FirefoxOptions {
	o.raw().WithProfile(path)
	return o
}

// WithUserDir 设置用户目录。
func (o *FirefoxOptions) WithUserDir(path string) *FirefoxOptions {
	o.raw().WithUserDir(path)
	return o
}

// WithArgument 添加启动参数。
func (o *FirefoxOptions) WithArgument(arg string, value ...string) *FirefoxOptions {
	o.raw().WithArgument(arg, value...)
	return o
}

// WithoutArgument 移除启动参数。
func (o *FirefoxOptions) WithoutArgument(arg string) *FirefoxOptions {
	o.raw().WithoutArgument(arg)
	return o
}

// WithPreference 设置 Firefox 首选项。
func (o *FirefoxOptions) WithPreference(key string, value any) *FirefoxOptions {
	o.raw().WithPreference(key, value)
	return o
}

// WithUserPromptHandler 设置 session 级默认提示框处理策略。
func (o *FirefoxOptions) WithUserPromptHandler(handler map[string]string) *FirefoxOptions {
	o.raw().WithUserPromptHandler(handler)
	return o
}

// Headless 设置无头模式。
func (o *FirefoxOptions) Headless(on bool) *FirefoxOptions {
	o.raw().Headless(on)
	return o
}

// WithProxy 设置代理地址。
func (o *FirefoxOptions) WithProxy(proxy string) *FirefoxOptions {
	o.raw().WithProxy(proxy)
	return o
}

// WithDownloadPath 设置下载目录。
func (o *FirefoxOptions) WithDownloadPath(path string) *FirefoxOptions {
	o.raw().WithDownloadPath(path)
	return o
}

// WithLoadMode 设置加载模式。
func (o *FirefoxOptions) WithLoadMode(mode FirefoxLoadMode) *FirefoxOptions {
	o.raw().WithLoadMode(mode)
	return o
}

// WithTimeouts 一次性覆盖三类超时。
func (o *FirefoxOptions) WithTimeouts(base, pageLoad, script float64) *FirefoxOptions {
	o.raw().WithTimeouts(base, pageLoad, script)
	return o
}

// WithBaseTimeout 设置基础超时。
func (o *FirefoxOptions) WithBaseTimeout(seconds float64) *FirefoxOptions {
	o.raw().WithBaseTimeout(seconds)
	return o
}

// WithPageLoadTimeout 设置页面加载超时。
func (o *FirefoxOptions) WithPageLoadTimeout(seconds float64) *FirefoxOptions {
	o.raw().WithPageLoadTimeout(seconds)
	return o
}

// WithScriptTimeout 设置脚本执行超时。
func (o *FirefoxOptions) WithScriptTimeout(seconds float64) *FirefoxOptions {
	o.raw().WithScriptTimeout(seconds)
	return o
}

// ExistingOnly 设置是否只连接已有浏览器。
func (o *FirefoxOptions) ExistingOnly(on bool) *FirefoxOptions {
	o.raw().ExistingOnly(on)
	return o
}

// AutoPortEnabled 设置是否启用自动端口。
func (o *FirefoxOptions) AutoPortEnabled(on bool) *FirefoxOptions {
	o.raw().AutoPortEnabled(on)
	return o
}

// WithAutoPortStart 设置自动端口起始值。
func (o *FirefoxOptions) WithAutoPortStart(start int) *FirefoxOptions {
	o.raw().WithAutoPortStart(start)
	return o
}

// WithRetry 一次性覆盖重试次数与间隔。
func (o *FirefoxOptions) WithRetry(times int, interval float64) *FirefoxOptions {
	o.raw().WithRetry(times, interval)
	return o
}

// WithRetryTimes 设置重试次数。
func (o *FirefoxOptions) WithRetryTimes(times int) *FirefoxOptions {
	o.raw().WithRetryTimes(times)
	return o
}

// WithRetryInterval 设置重试间隔。
func (o *FirefoxOptions) WithRetryInterval(interval float64) *FirefoxOptions {
	o.raw().WithRetryInterval(interval)
	return o
}

// WithUserContext 设置默认 user context。
func (o *FirefoxOptions) WithUserContext(userContext string) *FirefoxOptions {
	o.raw().WithUserContext(userContext)
	return o
}

// WithFPFile 设置指纹配置文件路径。
func (o *FirefoxOptions) WithFPFile(path string) *FirefoxOptions {
	o.raw().WithFPFile(path)
	return o
}

// PrivateMode 设置 Firefox 私密模式。
func (o *FirefoxOptions) PrivateMode(on bool) *FirefoxOptions {
	o.raw().PrivateMode(on)
	return o
}

// XPathPickerEnabled 设置是否启用 XPath picker。
func (o *FirefoxOptions) XPathPickerEnabled(on bool) *FirefoxOptions {
	o.raw().XPathPickerEnabled(on)
	return o
}

// ActionVisualEnabled 设置是否启用鼠标行为可视化调试模式。
func (o *FirefoxOptions) ActionVisualEnabled(on bool) *FirefoxOptions {
	o.raw().ActionVisualEnabled(on)
	return o
}

// CloseBrowserOnExitEnabled 设置是否在当前 Go 进程退出时自动关闭由本进程启动的浏览器。
func (o *FirefoxOptions) CloseBrowserOnExitEnabled(on bool) *FirefoxOptions {
	o.raw().CloseBrowserOnExitEnabled(on)
	return o
}

// WithWindowSize 通过启动参数设置窗口大小。
func (o *FirefoxOptions) WithWindowSize(width, height int) *FirefoxOptions {
	o.raw().WithWindowSize(width, height)
	return o
}

// QuickStart 应用新手友好预设。
func (o *FirefoxOptions) QuickStart(options FirefoxQuickStartOptions) *FirefoxOptions {
	o.raw().QuickStart(options)
	return o
}

// Validate 校验后续启动或 attach 需要的基础字段。
func (o *FirefoxOptions) Validate() error {
	return o.raw().Validate()
}

// BuildCommand 构建 Firefox 启动命令。
func (o *FirefoxOptions) BuildCommand() ([]string, error) {
	return o.raw().BuildCommand()
}

// WritePrefsToProfile 将首选项和代理设置写入 profile 的 user.js。
func (o *FirefoxOptions) WritePrefsToProfile() error {
	return o.raw().WritePrefsToProfile()
}

func (o *FirefoxOptions) raw() *config.FirefoxOptions {
	if o == nil {
		panic("ruyipage: FirefoxOptions receiver is nil")
	}
	if o.cfg == nil {
		o.cfg = config.NewFirefoxOptions()
	}
	return o.cfg
}
