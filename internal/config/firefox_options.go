package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pll177/ruyipage-go/internal/support"
)

const (
	defaultFirefoxBrowserPath   = `C:\Program Files\Mozilla Firefox\firefox.exe`
	defaultRetryTimes           = 10
	defaultRetryIntervalSeconds = 2.0
	defaultQuickStartWidth      = 1280
	defaultQuickStartHeight     = 800
)

// FirefoxLoadMode 表示页面加载等待模式。
type FirefoxLoadMode string

const (
	// LoadModeNormal 表示等待页面完整加载。
	LoadModeNormal FirefoxLoadMode = "normal"
	// LoadModeEager 表示等待 DOMContentLoaded。
	LoadModeEager FirefoxLoadMode = "eager"
	// LoadModeNone 表示不等待页面加载。
	LoadModeNone FirefoxLoadMode = "none"
)

// FirefoxTimeouts 表示 FirefoxOptions 中的超时配置，单位为秒。
type FirefoxTimeouts struct {
	Base     float64
	PageLoad float64
	Script   float64
}

// FirefoxQuickStartOptions 表示 QuickStart 的快捷预设输入。
type FirefoxQuickStartOptions struct {
	Port            int
	BrowserPath     string
	UserDir         string
	Private         bool
	Headless        bool
	XPathPicker     bool
	ActionVisual    bool
	WindowWidth     int
	WindowHeight    int
	TimeoutBase     float64
	TimeoutPageLoad float64
	TimeoutScript   float64
}

// ProxyAuthCredentials 表示从 fpfile 中解析出的代理认证信息。
type ProxyAuthCredentials struct {
	Username string
	Password string
}

// FirefoxOptions 承接 Firefox 启动、attach 与 examples 共用的配置状态。
type FirefoxOptions struct {
	initialized bool

	browserPath         string
	host                string
	port                int
	profilePath         string
	arguments           []string
	preferences         map[string]any
	headless            bool
	downloadPath        string
	loadMode            FirefoxLoadMode
	timeouts            FirefoxTimeouts
	existingOnly        bool
	retryTimes          int
	retryInterval       float64
	proxy               string
	autoPortEnabled     bool
	autoPortStart       int
	userContext         string
	fpfile              string
	privateMode         bool
	userPromptHandler   map[string]string
	xpathPickerEnabled  bool
	actionVisualEnabled bool
	closeBrowserOnExit  bool
}

// NewFirefoxOptions 返回带 Python 对齐默认值的配置对象。
func NewFirefoxOptions() *FirefoxOptions {
	opts := &FirefoxOptions{}
	opts.ensureDefaults()
	return opts
}

// DefaultFirefoxQuickStartOptions 返回与 Python quick_start 对齐的默认预设。
func DefaultFirefoxQuickStartOptions() FirefoxQuickStartOptions {
	return FirefoxQuickStartOptions{
		Port:            support.DefaultPort,
		WindowWidth:     defaultQuickStartWidth,
		WindowHeight:    defaultQuickStartHeight,
		TimeoutBase:     float64(support.DefaultBaseTimeoutSeconds),
		TimeoutPageLoad: float64(support.DefaultPageLoadTimeoutSeconds),
		TimeoutScript:   float64(support.DefaultScriptTimeoutSeconds),
	}
}

// Clone 返回当前配置的深拷贝。
func (o *FirefoxOptions) Clone() *FirefoxOptions {
	o.ensureDefaults()

	clone := *o
	clone.arguments = cloneStringSlice(o.arguments)
	clone.preferences = clonePreferences(o.preferences)
	clone.userPromptHandler = clonePromptHandler(o.userPromptHandler)

	return &clone
}

// BrowserPath 返回浏览器可执行文件路径。
func (o *FirefoxOptions) BrowserPath() string {
	o.ensureDefaults()
	return o.browserPath
}

// Address 返回 host:port 形式的地址。
func (o *FirefoxOptions) Address() string {
	o.ensureDefaults()
	return fmt.Sprintf("%s:%d", o.host, o.port)
}

// Host 返回调试主机。
func (o *FirefoxOptions) Host() string {
	o.ensureDefaults()
	return o.host
}

// Port 返回调试端口。
func (o *FirefoxOptions) Port() int {
	o.ensureDefaults()
	return o.port
}

// ProfilePath 返回 profile 路径。
func (o *FirefoxOptions) ProfilePath() string {
	o.ensureDefaults()
	return o.profilePath
}

// UserDir 返回 user_dir/profile 路径。
func (o *FirefoxOptions) UserDir() string {
	o.ensureDefaults()
	return o.profilePath
}

// Arguments 返回启动参数副本。
func (o *FirefoxOptions) Arguments() []string {
	o.ensureDefaults()
	return cloneStringSlice(o.arguments)
}

// Preferences 返回首选项副本。
func (o *FirefoxOptions) Preferences() map[string]any {
	o.ensureDefaults()
	return clonePreferences(o.preferences)
}

// IsHeadless 返回是否启用无头模式。
func (o *FirefoxOptions) IsHeadless() bool {
	o.ensureDefaults()
	return o.headless
}

// DownloadPath 返回下载目录。
func (o *FirefoxOptions) DownloadPath() string {
	o.ensureDefaults()
	return o.downloadPath
}

// LoadMode 返回加载模式。
func (o *FirefoxOptions) LoadMode() FirefoxLoadMode {
	o.ensureDefaults()
	return o.loadMode
}

// Timeouts 返回超时配置副本。
func (o *FirefoxOptions) Timeouts() FirefoxTimeouts {
	o.ensureDefaults()
	return o.timeouts
}

// IsExistingOnly 返回是否只连接已有浏览器。
func (o *FirefoxOptions) IsExistingOnly() bool {
	o.ensureDefaults()
	return o.existingOnly
}

// RetryTimes 返回连接重试次数。
func (o *FirefoxOptions) RetryTimes() int {
	o.ensureDefaults()
	return o.retryTimes
}

// RetryInterval 返回连接重试间隔，单位秒。
func (o *FirefoxOptions) RetryInterval() float64 {
	o.ensureDefaults()
	return o.retryInterval
}

// Proxy 返回代理地址。
func (o *FirefoxOptions) Proxy() string {
	o.ensureDefaults()
	return o.proxy
}

// IsAutoPortEnabled 返回是否启用自动端口。
func (o *FirefoxOptions) IsAutoPortEnabled() bool {
	o.ensureDefaults()
	return o.autoPortEnabled
}

// AutoPortStart 返回自动端口搜索起始值；0 表示沿用当前端口。
func (o *FirefoxOptions) AutoPortStart() int {
	o.ensureDefaults()
	return o.autoPortStart
}

// UserContext 返回默认 user context。
func (o *FirefoxOptions) UserContext() string {
	o.ensureDefaults()
	return o.userContext
}

// FPFile 返回指纹配置文件路径。
func (o *FirefoxOptions) FPFile() string {
	o.ensureDefaults()
	return o.fpfile
}

// IsPrivateMode 返回是否启用私密模式。
func (o *FirefoxOptions) IsPrivateMode() bool {
	o.ensureDefaults()
	return o.privateMode
}

// UserPromptHandler 返回用户提示框处理策略副本。
func (o *FirefoxOptions) UserPromptHandler() map[string]string {
	o.ensureDefaults()
	return clonePromptHandler(o.userPromptHandler)
}

// IsXPathPickerEnabled 返回是否启用 XPath picker。
func (o *FirefoxOptions) IsXPathPickerEnabled() bool {
	o.ensureDefaults()
	return o.xpathPickerEnabled
}

// IsActionVisualEnabled 返回是否启用鼠标行为可视化调试模式。
func (o *FirefoxOptions) IsActionVisualEnabled() bool {
	o.ensureDefaults()
	return o.actionVisualEnabled
}

// IsCloseBrowserOnExitEnabled 返回是否在当前 Go 进程退出时自动关闭由本进程启动的浏览器。
func (o *FirefoxOptions) IsCloseBrowserOnExitEnabled() bool {
	o.ensureDefaults()
	return o.closeBrowserOnExit
}

// WithBrowserPath 设置浏览器可执行文件路径。
func (o *FirefoxOptions) WithBrowserPath(path string) *FirefoxOptions {
	o.ensureDefaults()
	o.browserPath = path
	return o
}

// WithAddress 设置调试地址，支持 host:port 或仅 host。
func (o *FirefoxOptions) WithAddress(address string) *FirefoxOptions {
	o.ensureDefaults()

	if idx := strings.LastIndex(address, ":"); idx >= 0 {
		o.host = address[:idx]
		port, err := strconv.Atoi(address[idx+1:])
		if err != nil {
			o.port = 0
			return o
		}
		o.port = port
		return o
	}

	o.host = address
	return o
}

// WithPort 设置调试端口。
func (o *FirefoxOptions) WithPort(port int) *FirefoxOptions {
	o.ensureDefaults()
	o.port = port
	return o
}

// WithProfile 设置 Firefox profile 目录。
func (o *FirefoxOptions) WithProfile(path string) *FirefoxOptions {
	o.ensureDefaults()
	o.profilePath = path
	return o
}

// WithUserDir 设置用户目录，是 WithProfile 的新手友好别名。
func (o *FirefoxOptions) WithUserDir(path string) *FirefoxOptions {
	return o.WithProfile(path)
}

// WithArgument 添加启动参数；传 value 时会拼成 arg=value。
func (o *FirefoxOptions) WithArgument(arg string, value ...string) *FirefoxOptions {
	o.ensureDefaults()

	if len(value) > 0 {
		o.arguments = append(o.arguments, fmt.Sprintf("%s=%s", arg, value[0]))
		return o
	}

	for _, existing := range o.arguments {
		if existing == arg {
			return o
		}
	}

	o.arguments = append(o.arguments, arg)
	return o
}

// WithoutArgument 移除指定参数及其 arg=value 形式。
func (o *FirefoxOptions) WithoutArgument(arg string) *FirefoxOptions {
	o.ensureDefaults()

	filtered := make([]string, 0, len(o.arguments))
	for _, existing := range o.arguments {
		if existing == arg || strings.HasPrefix(existing, arg+"=") {
			continue
		}
		filtered = append(filtered, existing)
	}
	o.arguments = filtered
	return o
}

// WithPreference 设置 Firefox 首选项。
func (o *FirefoxOptions) WithPreference(key string, value any) *FirefoxOptions {
	o.ensureDefaults()
	o.preferences[key] = value
	return o
}

// WithUserPromptHandler 设置 session 级默认提示框处理策略。
func (o *FirefoxOptions) WithUserPromptHandler(handler map[string]string) *FirefoxOptions {
	o.ensureDefaults()
	if len(handler) == 0 {
		o.userPromptHandler = nil
		return o
	}

	o.userPromptHandler = clonePromptHandler(handler)
	return o
}

// Headless 设置无头模式。
func (o *FirefoxOptions) Headless(on bool) *FirefoxOptions {
	o.ensureDefaults()
	o.headless = on
	return o
}

// WithProxy 设置代理地址。
func (o *FirefoxOptions) WithProxy(proxy string) *FirefoxOptions {
	o.ensureDefaults()
	o.proxy = proxy
	return o
}

// WithDownloadPath 设置下载目录；与 Python 一致，显式设置时会转绝对路径。
func (o *FirefoxOptions) WithDownloadPath(path string) *FirefoxOptions {
	o.ensureDefaults()
	absPath, err := filepath.Abs(path)
	if err != nil {
		o.downloadPath = path
		return o
	}

	o.downloadPath = absPath
	return o
}

// WithLoadMode 设置加载模式；最终合法性由 Validate 统一校验。
func (o *FirefoxOptions) WithLoadMode(mode FirefoxLoadMode) *FirefoxOptions {
	o.ensureDefaults()
	o.loadMode = mode
	return o
}

// WithTimeouts 一次性覆盖三类超时，单位为秒。
func (o *FirefoxOptions) WithTimeouts(base, pageLoad, script float64) *FirefoxOptions {
	o.ensureDefaults()
	o.timeouts = FirefoxTimeouts{
		Base:     base,
		PageLoad: pageLoad,
		Script:   script,
	}
	return o
}

// WithBaseTimeout 设置基础超时。
func (o *FirefoxOptions) WithBaseTimeout(seconds float64) *FirefoxOptions {
	o.ensureDefaults()
	o.timeouts.Base = seconds
	return o
}

// WithPageLoadTimeout 设置页面加载超时。
func (o *FirefoxOptions) WithPageLoadTimeout(seconds float64) *FirefoxOptions {
	o.ensureDefaults()
	o.timeouts.PageLoad = seconds
	return o
}

// WithScriptTimeout 设置脚本执行超时。
func (o *FirefoxOptions) WithScriptTimeout(seconds float64) *FirefoxOptions {
	o.ensureDefaults()
	o.timeouts.Script = seconds
	return o
}

// ExistingOnly 设置是否只连接已有浏览器。
func (o *FirefoxOptions) ExistingOnly(on bool) *FirefoxOptions {
	o.ensureDefaults()
	o.existingOnly = on
	return o
}

// AutoPortEnabled 设置是否启用自动端口；关闭时会清空起始端口。
func (o *FirefoxOptions) AutoPortEnabled(on bool) *FirefoxOptions {
	o.ensureDefaults()
	o.autoPortEnabled = on
	if !on {
		o.autoPortStart = 0
	}
	return o
}

// WithAutoPortStart 设置自动端口搜索起始值，同时启用自动端口。
func (o *FirefoxOptions) WithAutoPortStart(start int) *FirefoxOptions {
	o.ensureDefaults()
	o.autoPortEnabled = true
	o.autoPortStart = start
	return o
}

// WithRetry 一次性覆盖重试次数与间隔。
func (o *FirefoxOptions) WithRetry(times int, interval float64) *FirefoxOptions {
	o.ensureDefaults()
	o.retryTimes = times
	o.retryInterval = interval
	return o
}

// WithRetryTimes 设置重试次数。
func (o *FirefoxOptions) WithRetryTimes(times int) *FirefoxOptions {
	o.ensureDefaults()
	o.retryTimes = times
	return o
}

// WithRetryInterval 设置重试间隔。
func (o *FirefoxOptions) WithRetryInterval(interval float64) *FirefoxOptions {
	o.ensureDefaults()
	o.retryInterval = interval
	return o
}

// WithUserContext 设置默认 user context。
func (o *FirefoxOptions) WithUserContext(userContext string) *FirefoxOptions {
	o.ensureDefaults()
	o.userContext = userContext
	return o
}

// WithFPFile 设置指纹配置文件路径。
func (o *FirefoxOptions) WithFPFile(path string) *FirefoxOptions {
	o.ensureDefaults()
	o.fpfile = path
	return o
}

// PrivateMode 设置 Firefox 私密模式。
func (o *FirefoxOptions) PrivateMode(on bool) *FirefoxOptions {
	o.ensureDefaults()
	o.privateMode = on
	return o
}

// XPathPickerEnabled 设置是否启用 XPath picker。
func (o *FirefoxOptions) XPathPickerEnabled(on bool) *FirefoxOptions {
	o.ensureDefaults()
	o.xpathPickerEnabled = on
	return o
}

// ActionVisualEnabled 设置是否启用鼠标行为可视化调试模式。
func (o *FirefoxOptions) ActionVisualEnabled(on bool) *FirefoxOptions {
	o.ensureDefaults()
	o.actionVisualEnabled = on
	return o
}

// CloseBrowserOnExitEnabled 设置是否在当前 Go 进程退出时自动关闭由本进程启动的浏览器。
func (o *FirefoxOptions) CloseBrowserOnExitEnabled(on bool) *FirefoxOptions {
	o.ensureDefaults()
	o.closeBrowserOnExit = on
	return o
}

// WithWindowSize 通过启动参数设置窗口大小，并覆盖已有 width/height 参数。
func (o *FirefoxOptions) WithWindowSize(width, height int) *FirefoxOptions {
	o.ensureDefaults()

	filtered := make([]string, 0, len(o.arguments))
	for _, existing := range o.arguments {
		if strings.HasPrefix(existing, "--width=") || strings.HasPrefix(existing, "--height=") {
			continue
		}
		filtered = append(filtered, existing)
	}

	o.arguments = append(filtered,
		fmt.Sprintf("--width=%d", width),
		fmt.Sprintf("--height=%d", height),
	)
	return o
}

// QuickStart 按 Python quick_start 语义快速套用新手友好预设。
func (o *FirefoxOptions) QuickStart(options FirefoxQuickStartOptions) *FirefoxOptions {
	o.ensureDefaults()

	if options.Port > 0 {
		o.WithPort(options.Port)
	}
	if options.BrowserPath != "" {
		o.WithBrowserPath(options.BrowserPath)
	}
	if options.UserDir != "" {
		o.WithUserDir(options.UserDir)
	}

	o.PrivateMode(options.Private)
	o.Headless(options.Headless)
	o.XPathPickerEnabled(options.XPathPicker)
	o.ActionVisualEnabled(options.ActionVisual)

	width := options.WindowWidth
	height := options.WindowHeight
	if width == 0 && height == 0 {
		width = defaultQuickStartWidth
		height = defaultQuickStartHeight
	}
	if width > 0 && height > 0 {
		o.WithWindowSize(width, height)
	}

	base := options.TimeoutBase
	if base == 0 {
		base = float64(support.DefaultBaseTimeoutSeconds)
	}
	pageLoad := options.TimeoutPageLoad
	if pageLoad == 0 {
		pageLoad = float64(support.DefaultPageLoadTimeoutSeconds)
	}
	script := options.TimeoutScript
	if script == 0 {
		script = float64(support.DefaultScriptTimeoutSeconds)
	}

	o.WithTimeouts(base, pageLoad, script)
	return o
}

// Validate 校验后续启动或 attach 需要的基础字段。
func (o *FirefoxOptions) Validate() error {
	if o == nil {
		return errors.New("FirefoxOptions 不能为空")
	}

	o.ensureDefaults()

	if o.host == "" {
		return errors.New("FirefoxOptions.Host 不能为空")
	}
	if o.port < 1 || o.port > 65535 {
		return fmt.Errorf("FirefoxOptions.Port 必须在 1-65535 之间，当前为 %d", o.port)
	}
	if !o.loadMode.Valid() {
		return fmt.Errorf(
			"FirefoxOptions.LoadMode 必须是 %q、%q 或 %q，当前为 %q",
			LoadModeNormal,
			LoadModeEager,
			LoadModeNone,
			o.loadMode,
		)
	}

	return nil
}

// BuildCommand 构建 Firefox 启动命令。
func (o *FirefoxOptions) BuildCommand() ([]string, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}

	cmd := []string{
		o.browserPath,
		fmt.Sprintf("--remote-debugging-port=%d", o.port),
		"--no-remote",
		"--marionette",
	}

	if o.profilePath != "" {
		cmd = append(cmd, "--profile", o.profilePath)
	}
	if o.headless {
		cmd = append(cmd, "--headless")
	}
	if o.privateMode {
		cmd = append(cmd, "-private")
	}
	if o.fpfile != "" {
		cmd = append(cmd, fmt.Sprintf("--fpfile=%s", o.fpfile))
	}

	cmd = append(cmd, cloneStringSlice(o.arguments)...)
	return cmd, nil
}

// WritePrefsToProfile 将首选项和代理设置写入 profile 的 user.js。
func (o *FirefoxOptions) WritePrefsToProfile() error {
	o.ensureDefaults()

	if o.profilePath == "" {
		return nil
	}

	prefs := clonePreferences(o.preferences)
	setDefaultPreference(prefs, "remote.prefs.recommended", true)
	setDefaultPreference(prefs, "datareporting.policy.dataSubmissionEnabled", false)
	setDefaultPreference(prefs, "toolkit.telemetry.reportingpolicy.firstRun", false)
	setDefaultPreference(prefs, "browser.shell.checkDefaultBrowser", false)
	setDefaultPreference(prefs, "browser.startup.homepage_override.mstone", "ignore")
	setDefaultPreference(prefs, "browser.tabs.warnOnClose", false)
	setDefaultPreference(prefs, "browser.warnOnQuit", false)
	setDefaultPreference(prefs, "marionette.enabled", true)

	if o.downloadPath != "" {
		prefs["browser.download.dir"] = o.downloadPath
		prefs["browser.download.folderList"] = 2
		prefs["browser.download.useDownloadDir"] = true
	}

	if o.proxy != "" {
		applyProxyPreferences(prefs, o.proxy)
	}

	if len(prefs) == 0 {
		return nil
	}

	if err := os.MkdirAll(o.profilePath, 0o755); err != nil {
		return err
	}
	userJSPath := filepath.Join(o.profilePath, "user.js")
	return support.NewJSPrefsFile(userJSPath).RewriteAll(prefs)
}

// ProxyAuthCredentials 从 fpfile 中提取代理认证信息。
func (o *FirefoxOptions) ProxyAuthCredentials() (*ProxyAuthCredentials, error) {
	o.ensureDefaults()

	auth, err := readHTTPAuthFromFPFile(o.fpfile)
	if err != nil {
		return nil, err
	}
	if len(auth) == 0 {
		return nil, nil
	}

	username, hasUsername := auth["username"]
	password, hasPassword := auth["password"]
	if !hasUsername && !hasPassword {
		return nil, nil
	}

	return &ProxyAuthCredentials{
		Username: username,
		Password: password,
	}, nil
}

// Valid 返回加载模式是否合法。
func (m FirefoxLoadMode) Valid() bool {
	switch m {
	case LoadModeNormal, LoadModeEager, LoadModeNone:
		return true
	default:
		return false
	}
}

func (o *FirefoxOptions) ensureDefaults() {
	if o.initialized {
		return
	}

	o.browserPath = defaultFirefoxBrowserPath
	o.host = support.DefaultHost
	o.port = support.DefaultPort
	o.arguments = []string{}
	o.preferences = map[string]any{}
	o.downloadPath = "."
	o.loadMode = LoadModeNormal
	o.timeouts = FirefoxTimeouts{
		Base:     float64(support.DefaultBaseTimeoutSeconds),
		PageLoad: float64(support.DefaultPageLoadTimeoutSeconds),
		Script:   float64(support.DefaultScriptTimeoutSeconds),
	}
	o.retryTimes = defaultRetryTimes
	o.retryInterval = defaultRetryIntervalSeconds
	o.closeBrowserOnExit = true
	o.initialized = true
}

func cloneStringSlice(src []string) []string {
	if len(src) == 0 {
		return []string{}
	}

	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

func clonePreferences(src map[string]any) map[string]any {
	if len(src) == 0 {
		return map[string]any{}
	}

	dst := make(map[string]any, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func clonePromptHandler(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}

	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func setDefaultPreference(prefs map[string]any, key string, value any) {
	if _, exists := prefs[key]; exists {
		return
	}
	prefs[key] = value
}

func applyProxyPreferences(prefs map[string]any, proxy string) {
	scheme := "http"
	address := proxy

	if idx := strings.Index(proxy, "://"); idx >= 0 {
		scheme = proxy[:idx]
		address = proxy[idx+3:]
	}

	host := address
	port := "8080"
	if idx := strings.LastIndex(address, ":"); idx >= 0 {
		host = address[:idx]
		port = address[idx+1:]
	}

	proxyPort, err := strconv.Atoi(port)
	if err != nil {
		proxyPort = 0
	}

	if strings.HasPrefix(scheme, "socks") {
		prefs["network.proxy.type"] = 1
		prefs["network.proxy.socks"] = host
		prefs["network.proxy.socks_port"] = proxyPort
		if strings.Contains(scheme, "5") {
			prefs["network.proxy.socks_version"] = 5
		} else {
			prefs["network.proxy.socks_version"] = 4
		}
		return
	}

	prefs["network.proxy.type"] = 1
	prefs["network.proxy.http"] = host
	prefs["network.proxy.http_port"] = proxyPort
	prefs["network.proxy.ssl"] = host
	prefs["network.proxy.ssl_port"] = proxyPort
	setDefaultPreference(prefs, "signon.autologin.proxy", true)
	setDefaultPreference(prefs, "network.auth.subresource-http-auth-allow", 2)
}

func readHTTPAuthFromFPFile(path string) (map[string]string, error) {
	if path == "" {
		return map[string]string{}, nil
	}

	fpfilePath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(fpfilePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("fpfile 不存在: %s", fpfilePath)
		}
		return nil, err
	}

	content, err := os.ReadFile(fpfilePath)
	if err != nil {
		return nil, err
	}

	result := map[string]string{}
	pattern := regexp.MustCompile(`^\s*(httpauth\.(?:username|password))\s*[:=]\s*(.*?)\s*$`)

	for _, rawLine := range strings.Split(string(content), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		match := pattern.FindStringSubmatch(line)
		if len(match) != 3 {
			continue
		}

		key := match[1]
		value := match[2]
		switch key {
		case "httpauth.username":
			result["username"] = value
		case "httpauth.password":
			result["password"] = value
		}
	}

	return result, nil
}
