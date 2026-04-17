package browser

import (
	stderrors "errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/pll177/ruyipage-go/internal/adapter"
	"github.com/pll177/ruyipage-go/internal/base"
	"github.com/pll177/ruyipage-go/internal/bidi"
	"github.com/pll177/ruyipage-go/internal/config"
	"github.com/pll177/ruyipage-go/internal/support"
	"github.com/pll177/ruyipage-go/internal/units"
)

var (
	getFirefoxBiDiWSURL   = adapter.GetBiDiWSURL
	waitForFirefoxReady   = adapter.WaitForFirefox
	launchFirefoxProcess  = adapter.LaunchFirefox
	bindProcessKillOnExit = adapter.BindProcessKillOnParentExit
)

var (
	firefoxRegistryMu sync.Mutex
	firefoxRegistry   = make(map[string]*Firefox)
)

// Firefox 承接浏览器启动、连接、attach、tab 基础管理与退出清理。
type Firefox struct {
	mu     sync.RWMutex
	quitMu sync.Mutex

	options *config.FirefoxOptions
	address string

	driver             *base.BrowserBiDiDriver
	process            *exec.Cmd
	processDone        chan error
	processExitBinding func() error

	sessionID   string
	ownsSession bool

	contextIDs    []string
	contexts      map[string]ProbeContextInfo
	clientWindows []map[string]any
	autoProfile   string
	quitting      bool
}

// NewFirefox 按 FirefoxOptions 启动或连接浏览器。
func NewFirefox(options *config.FirefoxOptions) (*Firefox, error) {
	prepared, err := prepareFirefoxOptions(options)
	if err != nil {
		return nil, err
	}

	address := prepared.Address()

	firefoxRegistryMu.Lock()
	if existing := firefoxRegistry[address]; existing != nil {
		if existing.canReuse() {
			firefoxRegistryMu.Unlock()
			return existing, nil
		}
		delete(firefoxRegistry, address)
	}

	instance := newFirefoxInstance(prepared)
	firefoxRegistry[address] = instance
	firefoxRegistryMu.Unlock()

	if err := instance.connectOrLaunch(); err != nil {
		instance.cleanupAfterFailedInitialization()
		return nil, err
	}

	return instance, nil
}

// ConnectFirefox 连接到现有地址。
func ConnectFirefox(address string) (*Firefox, error) {
	options := config.NewFirefoxOptions().
		WithAddress(address).
		ExistingOnly(true)
	return NewFirefox(options)
}

// CreateFirefoxFromProbeInfo 直接接管 live probe 的 driver 与 session。
func CreateFirefoxFromProbeInfo(info *ProbeInfo) (*Firefox, error) {
	if info == nil || info.Driver == nil {
		return nil, support.NewBrowserConnectError("探测结果中缺少可复用的 BiDi 连接", nil)
	}

	firefoxRegistryMu.Lock()
	if existing := firefoxRegistry[info.Address]; existing != nil {
		if existing.canReuse() {
			firefoxRegistryMu.Unlock()
			_ = CloseProbeInfo(info)
			return existing, nil
		}
		delete(firefoxRegistry, info.Address)
	}

	options := config.NewFirefoxOptions().
		WithAddress(info.Address).
		ExistingOnly(true)
	instance := newFirefoxInstance(options)
	firefoxRegistry[info.Address] = instance
	firefoxRegistryMu.Unlock()

	if err := instance.attachProbeInfo(info); err != nil {
		instance.cleanupAfterFailedInitialization()
		return nil, err
	}

	return instance, nil
}

// Address 返回当前浏览器地址。
func (f *Firefox) Address() string {
	if f == nil {
		return ""
	}

	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.address
}

// Driver 返回当前 BrowserBiDiDriver。
func (f *Firefox) Driver() *base.BrowserBiDiDriver {
	if f == nil {
		return nil
	}

	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.driver
}

// SessionID 返回当前 session id。
func (f *Firefox) SessionID() string {
	if f == nil {
		return ""
	}

	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.sessionID
}

// Options 返回浏览器配置副本。
func (f *Firefox) Options() *config.FirefoxOptions {
	if f == nil {
		return nil
	}

	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.options.Clone()
}

// Process 返回受管 Firefox 进程。
func (f *Firefox) Process() *exec.Cmd {
	if f == nil {
		return nil
	}

	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.process
}

// TabsCount 返回当前 tab 数量。
func (f *Firefox) TabsCount() int {
	if f == nil {
		return 0
	}

	_ = f.refreshTabs()

	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.contextIDs)
}

// TabIDs 返回当前 tab id 列表副本。
func (f *Firefox) TabIDs() []string {
	if f == nil {
		return []string{}
	}

	_ = f.refreshTabs()

	f.mu.RLock()
	defer f.mu.RUnlock()
	return cloneFirefoxStringSlice(f.contextIDs)
}

// LatestTabID 返回最新 tab id。
func (f *Firefox) LatestTabID() string {
	if f == nil {
		return ""
	}

	_ = f.refreshTabs()

	f.mu.RLock()
	defer f.mu.RUnlock()
	if len(f.contextIDs) == 0 {
		return ""
	}
	return f.contextIDs[len(f.contextIDs)-1]
}

// Contexts 返回当前上下文快照副本。
func (f *Firefox) Contexts() []ProbeContextInfo {
	if f == nil {
		return []ProbeContextInfo{}
	}

	_ = f.refreshTabs()

	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make([]ProbeContextInfo, 0, len(f.contextIDs))
	for _, contextID := range f.contextIDs {
		info, ok := f.contexts[contextID]
		if !ok {
			continue
		}
		result = append(result, info)
	}
	return result
}

// WindowHandles 返回客户端窗口信息副本。
func (f *Firefox) WindowHandles() []map[string]any {
	if f == nil {
		return []map[string]any{}
	}

	_ = f.refreshClientWindows()

	f.mu.RLock()
	defer f.mu.RUnlock()
	return cloneMapSlice(f.clientWindows)
}

// Cookies 返回浏览器级可见 Cookie 列表。
func (f *Firefox) Cookies(allInfo bool) ([]units.CookieInfo, error) {
	if f == nil {
		return []units.CookieInfo{}, nil
	}

	driver := f.Driver()
	if driver == nil {
		return nil, support.NewPageDisconnectedError("Firefox 尚未连接", nil)
	}

	result, err := bidi.GetCookies(driver, nil, nil, f.baseTimeout())
	if err != nil {
		return nil, err
	}

	raw := readMapSliceFromAny(result["cookies"])
	cookies := make([]units.CookieInfo, 0, len(raw))
	for _, cookie := range raw {
		view := cloneMap(cookie)
		if !allInfo {
			view = map[string]any{
				"name":   cookie["name"],
				"value":  cookie["value"],
				"domain": cookie["domain"],
				"path":   cookie["path"],
			}
		}
		cookies = append(cookies, units.NewCookieInfo(view))
	}
	return cookies, nil
}

// NewTab 创建新 tab，并在传入 url 时立即导航。
func (f *Firefox) NewTab(url string, background bool) (string, error) {
	driver := f.Driver()
	if driver == nil {
		return "", support.NewPageDisconnectedError("Firefox 尚未连接", nil)
	}

	referenceContext := ""
	tabIDs := f.TabIDs()
	if len(tabIDs) > 0 {
		referenceContext = tabIDs[0]
	}

	result, err := bidi.Create(
		driver,
		"tab",
		referenceContext,
		background,
		f.options.UserContext(),
		f.baseTimeout(),
	)
	if err != nil {
		return "", err
	}

	contextID := readStringValue(result["context"])
	if contextID == "" {
		return "", support.NewContextLostError("新建标签页未返回有效 context id", nil)
	}

	f.onContextCreated(map[string]any{
		"context":     contextID,
		"url":         "",
		"userContext": f.options.UserContext(),
	})

	if url != "" {
		if _, err := bidi.Navigate(driver, contextID, url, navigateWaitMode(f.options.LoadMode()), f.pageLoadTimeout()); err != nil {
			return contextID, err
		}
		f.onNavigationEvent(map[string]any{
			"context": contextID,
			"url":     url,
		})
	}

	return contextID, nil
}

func navigateWaitMode(mode config.FirefoxLoadMode) string {
	switch mode {
	case config.LoadModeEager:
		return "interactive"
	case config.LoadModeNone:
		return "none"
	default:
		return "complete"
	}
}

// ActivateTab 激活指定 tab。
func (f *Firefox) ActivateTab(contextID string) error {
	driver := f.Driver()
	if driver == nil {
		return support.NewPageDisconnectedError("Firefox 尚未连接", nil)
	}
	if contextID == "" {
		return support.NewContextLostError("context id 不能为空", nil)
	}

	_, err := bidi.Activate(driver, contextID, f.baseTimeout())
	return err
}

// CloseTabs 关闭指定 tab；others=true 时表示关闭其他 tab。
func (f *Firefox) CloseTabs(contextIDs []string, others bool) error {
	driver := f.Driver()
	if driver == nil {
		return support.NewPageDisconnectedError("Firefox 尚未连接", nil)
	}

	if len(contextIDs) == 0 && !others {
		return nil
	}

	current := f.TabIDs()
	targetSet := make(map[string]struct{}, len(contextIDs))
	for _, contextID := range contextIDs {
		if contextID == "" {
			continue
		}
		targetSet[contextID] = struct{}{}
	}

	toClose := make([]string, 0, len(current))
	for _, contextID := range current {
		_, selected := targetSet[contextID]
		if (others && !selected) || (!others && selected) {
			toClose = append(toClose, contextID)
		}
	}

	var firstErr error
	for _, contextID := range toClose {
		if _, err := bidi.Close(driver, contextID, false, f.baseTimeout()); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		f.onContextDestroyed(map[string]any{"context": contextID})
	}

	_ = f.refreshTabs()
	return firstErr
}

// Close 使用默认参数关闭浏览器。
func (f *Firefox) Close() error {
	return f.Quit(5*time.Second, false)
}

// Quit 关闭浏览器、结束 session 并清理进程与缓存。
func (f *Firefox) Quit(timeout time.Duration, force bool) error {
	if f == nil {
		return nil
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	f.quitMu.Lock()
	defer f.quitMu.Unlock()

	f.mu.Lock()
	if f.quitting {
		f.mu.Unlock()
		return nil
	}
	f.quitting = true

	driver := f.driver
	process := f.process
	processDone := f.processDone
	processExitBinding := f.processExitBinding
	ownsSession := f.ownsSession

	f.driver = nil
	f.process = nil
	f.processDone = nil
	f.processExitBinding = nil
	f.ownsSession = false
	f.sessionID = ""
	f.contextIDs = []string{}
	f.contexts = make(map[string]ProbeContextInfo)
	f.clientWindows = []map[string]any{}
	autoProfile := f.autoProfile
	f.autoProfile = ""
	f.mu.Unlock()

	var firstErr error
	if driver != nil {
		driver.MarkClosing()
		_, _ = bidi.CloseBrowser(driver, minFirefoxDuration(f.baseTimeout(), 3*time.Second))
		if ownsSession {
			if err := bidi.End(driver, f.baseTimeout()); err != nil && !isFirefoxIgnorableShutdownError(err) && firstErr == nil {
				firstErr = err
			}
		}
		if err := driver.Stop(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if err := waitManagedProcessExit(process, processDone, timeout, force); err != nil && firstErr == nil {
		firstErr = err
	}
	if processExitBinding != nil && managedProcessExited(process) {
		if err := processExitBinding(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if autoProfile != "" {
		if err := os.RemoveAll(autoProfile); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if f.options != nil {
		if err := f.options.CleanupManagedFPFile(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	f.removeFromRegistry()
	return firstErr
}

func newFirefoxInstance(options *config.FirefoxOptions) *Firefox {
	return &Firefox{
		options:       options,
		address:       options.Address(),
		contextIDs:    []string{},
		contexts:      make(map[string]ProbeContextInfo),
		clientWindows: []map[string]any{},
	}
}

func (f *Firefox) canReuse() bool {
	if f == nil {
		return false
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.driver != nil && f.driver.IsRunning() {
		return true
	}
	return f.process != nil && (f.process.ProcessState == nil || !f.process.ProcessState.Exited())
}

func prepareFirefoxOptions(options *config.FirefoxOptions) (*config.FirefoxOptions, error) {
	if options == nil {
		options = config.NewFirefoxOptions()
	}

	prepared := options.Clone()
	if err := prepared.Validate(); err != nil {
		return nil, err
	}

	if prepared.IsAutoPortEnabled() {
		start := prepared.AutoPortStart()
		if start <= 0 {
			start = prepared.Port()
		}

		port, err := findFirefoxFreePort(start)
		if err != nil {
			return nil, err
		}
		prepared.SetResolvedPort(port)
	}

	return prepared, nil
}

func (f *Firefox) connectOrLaunch() error {
	connected, err := f.tryConnect()
	if connected {
		return nil
	}
	if err != nil && f.options.IsExistingOnly() {
		return err
	}
	if f.options.IsExistingOnly() {
		return support.NewBrowserConnectError(
			fmt.Sprintf(
				"无法连接到 %s，请先启动 Firefox：\n  firefox.exe --remote-debugging-port=%d",
				f.Address(),
				f.options.Port(),
			),
			err,
		)
	}

	if err := f.ensureLaunchPortAvailable(); err != nil {
		return err
	}
	if err := f.launchBrowser(); err != nil {
		return err
	}

	host, port, splitErr := splitProbeAddress(f.Address())
	if splitErr != nil {
		return support.NewBrowserConnectError("Firefox 地址无效", splitErr)
	}

	connectTimeout := f.browserConnectTimeout()
	if !waitForFirefoxReady(host, port, connectTimeout) {
		err = support.NewBrowserConnectError(
			fmt.Sprintf("等待 Firefox 远端调试端口 %s 就绪超时", f.Address()),
			nil,
		)
	}

	var lastErr error
	for attempt := 0; attempt <= f.options.RetryTimes(); attempt++ {
		connected, tryErr := f.tryConnect()
		if connected {
			return nil
		}
		if tryErr != nil {
			lastErr = tryErr
		}

		if attempt == f.options.RetryTimes() {
			break
		}
		time.Sleep(f.retryInterval())
	}

	if lastErr == nil {
		lastErr = err
	}
	return support.NewBrowserConnectError(
		fmt.Sprintf("启动后无法连接到 %s，请检查 Firefox 是否正常启动", f.Address()),
		lastErr,
	)
}

func (f *Firefox) attachProbeInfo(info *ProbeInfo) error {
	if info == nil || info.Driver == nil {
		return support.NewBrowserConnectError("探测结果中缺少可复用的 BiDi 连接", nil)
	}
	if strings.TrimSpace(info.SessionID) == "" {
		return support.NewBrowserConnectError("探测结果中缺少有效的 Firefox BiDi session", nil)
	}

	driver := info.Driver
	info.Driver = nil

	f.mu.Lock()
	f.driver = driver
	f.sessionID = info.SessionID
	f.ownsSession = info.SessionOwned
	f.clientWindows = cloneMapSlice(info.ClientWindows)
	f.contextIDs = make([]string, 0, len(info.Contexts))
	f.contexts = make(map[string]ProbeContextInfo, len(info.Contexts))
	for _, context := range info.Contexts {
		f.contextIDs = append(f.contextIDs, context.Context)
		f.contexts[context.Context] = context
	}
	f.mu.Unlock()

	if driver != nil {
		driver.SetSessionID(info.SessionID)
	}

	return f.subscribeLifecycleEvents(driver)
}

func (f *Firefox) tryConnect() (bool, error) {
	host, port, err := splitProbeAddress(f.Address())
	if err != nil {
		return false, support.NewBrowserConnectError("Firefox 地址无效", err)
	}
	if !support.IsPortOpen(host, port, 2*time.Second) {
		return false, nil
	}

	wsURL, err := getFirefoxBiDiWSURL(host, port, f.browserConnectTimeout())
	if err != nil {
		return false, err
	}

	driver := base.NewBrowserBiDiDriver(f.Address())
	if err := driver.Start(wsURL, f.browserConnectTimeout()); err != nil {
		_ = driver.Stop()
		return false, err
	}

	sessionID, ownsSession, err := f.createSession(driver)
	if err != nil {
		_ = driver.Stop()
		return false, err
	}
	driver.SetSessionID(sessionID)

	if err := f.subscribeLifecycleEvents(driver); err != nil {
		if ownsSession {
			_ = bidi.End(driver, f.baseTimeout())
		}
		_ = driver.Stop()
		return false, err
	}

	clientWindows, contexts := f.snapshotBrowserState(driver)

	f.mu.Lock()
	f.driver = driver
	f.sessionID = sessionID
	f.ownsSession = ownsSession
	f.clientWindows = cloneMapSlice(clientWindows)
	f.contextIDs = make([]string, 0, len(contexts))
	f.contexts = make(map[string]ProbeContextInfo, len(contexts))
	for _, context := range contexts {
		f.contextIDs = append(f.contextIDs, context.Context)
		f.contexts[context.Context] = context
	}
	f.mu.Unlock()
	return true, nil
}

func (f *Firefox) createSession(driver *base.BrowserBiDiDriver) (string, bool, error) {
	status, err := bidi.Status(driver, f.baseTimeout())
	if err != nil {
		return "", false, support.NewBrowserConnectError("查询 Firefox BiDi session 状态失败", err)
	}

	createNewSession := func() (string, bool, error) {
		result, err := bidi.New(driver, map[string]any{}, convertPromptHandler(f.options.UserPromptHandler()), f.baseTimeout())
		if err != nil {
			if isMaximumActiveSessionsError(err) {
				return "", false, support.NewBrowserConnectError(
					buildOccupiedFirefoxMessage(status.Message, err.Error()),
					err,
				)
			}
			return "", false, support.NewBrowserConnectError("创建 Firefox BiDi session 失败", err)
		}
		if strings.TrimSpace(result.SessionID) == "" {
			return "", false, support.NewBrowserConnectError("创建 Firefox BiDi session 失败：未返回有效 session id", nil)
		}
		return result.SessionID, true, nil
	}

	if status.Ready {
		return createNewSession()
	}

	_ = bidi.End(driver, f.baseTimeout())
	return createNewSession()
}

func (f *Firefox) subscribeLifecycleEvents(driver *base.BrowserBiDiDriver) error {
	if driver == nil {
		return support.NewBrowserConnectError("BrowserBiDiDriver 未初始化", nil)
	}
	if strings.TrimSpace(driver.SessionID()) == "" {
		return support.NewBrowserConnectError("Firefox BiDi session 未初始化", nil)
	}

	events := []string{
		"browsingContext.contextCreated",
		"browsingContext.contextDestroyed",
		"browsingContext.load",
		"browsingContext.domContentLoaded",
	}
	if _, err := bidi.Subscribe(driver, events, nil, f.baseTimeout()); err != nil {
		return support.NewBrowserConnectError("订阅 Firefox 生命周期事件失败", err)
	}

	if err := driver.SetCallback("browsingContext.contextCreated", f.onContextCreated, "", false); err != nil {
		return err
	}
	if err := driver.SetCallback("browsingContext.contextDestroyed", f.onContextDestroyed, "", false); err != nil {
		return err
	}
	if err := driver.SetCallback("browsingContext.load", f.onNavigationEvent, "", true); err != nil {
		return err
	}
	if err := driver.SetCallback("browsingContext.domContentLoaded", f.onNavigationEvent, "", true); err != nil {
		return err
	}
	return nil
}

func (f *Firefox) snapshotBrowserState(driver *base.BrowserBiDiDriver) ([]map[string]any, []ProbeContextInfo) {
	clientWindows := []map[string]any{}
	if result, err := bidi.GetClientWindows(driver, f.baseTimeout()); err == nil {
		clientWindows = readMapSliceFromAny(result["clientWindows"])
	}

	contexts := []ProbeContextInfo{}
	maxDepth := 0
	if result, err := bidi.GetTree(driver, &maxDepth, "", f.baseTimeout()); err == nil {
		contexts = buildProbeContexts(readMapSliceFromAny(result["contexts"]))
	}

	return clientWindows, contexts
}

func (f *Firefox) refreshTabs() error {
	driver := f.Driver()
	if driver == nil {
		return support.NewPageDisconnectedError("Firefox 尚未连接", nil)
	}

	maxDepth := 0
	result, err := bidi.GetTree(driver, &maxDepth, "", f.baseTimeout())
	if err != nil {
		return err
	}

	contexts := buildProbeContexts(readMapSliceFromAny(result["contexts"]))

	f.mu.Lock()
	defer f.mu.Unlock()

	f.contextIDs = make([]string, 0, len(contexts))
	f.contexts = make(map[string]ProbeContextInfo, len(contexts))
	for _, context := range contexts {
		f.contextIDs = append(f.contextIDs, context.Context)
		f.contexts[context.Context] = context
	}
	return nil
}

func (f *Firefox) refreshClientWindows() error {
	driver := f.Driver()
	if driver == nil {
		return support.NewPageDisconnectedError("Firefox 尚未连接", nil)
	}

	result, err := bidi.GetClientWindows(driver, f.baseTimeout())
	if err != nil {
		return err
	}

	clientWindows := readMapSliceFromAny(result["clientWindows"])

	f.mu.Lock()
	f.clientWindows = cloneMapSlice(clientWindows)
	f.mu.Unlock()
	return nil
}

func (f *Firefox) onContextCreated(params map[string]any) {
	context := ProbeContextInfo{
		Context:        readStringValue(params["context"]),
		URL:            readStringValue(params["url"]),
		UserContext:    resolveContextUserContext(params),
		OriginalOpener: params["originalOpener"],
	}
	if context.Context == "" {
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.contexts[context.Context]; !exists {
		f.contextIDs = append(f.contextIDs, context.Context)
	}
	f.contexts[context.Context] = context
}

func (f *Firefox) onContextDestroyed(params map[string]any) {
	contextID := readStringValue(params["context"])
	if contextID == "" {
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.contexts, contextID)
	filtered := make([]string, 0, len(f.contextIDs))
	for _, existing := range f.contextIDs {
		if existing == contextID {
			continue
		}
		filtered = append(filtered, existing)
	}
	f.contextIDs = filtered
}

func (f *Firefox) onNavigationEvent(params map[string]any) {
	contextID := readStringValue(params["context"])
	if contextID == "" {
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	info, ok := f.contexts[contextID]
	if !ok {
		info = ProbeContextInfo{Context: contextID, UserContext: "default"}
	}
	if url := readStringValue(params["url"]); url != "" {
		info.URL = url
	}
	f.contexts[contextID] = info
}

func (f *Firefox) launchBrowser() error {
	if f.options.ProfilePath() == "" {
		autoProfile, err := os.MkdirTemp("", "ruyipage_")
		if err != nil {
			return support.NewBrowserLaunchError("创建 Firefox 临时 profile 失败", err)
		}
		f.options.SetResolvedProfilePath(autoProfile)
		f.mu.Lock()
		f.autoProfile = autoProfile
		f.mu.Unlock()
	}

	if err := f.options.WritePrefsToProfile(); err != nil {
		return support.NewBrowserLaunchError("写入 Firefox profile 首选项失败", err)
	}

	command, err := f.options.BuildCommand()
	if err != nil {
		return support.NewBrowserLaunchError("构建 Firefox 启动命令失败", err)
	}

	process, err := launchFirefoxProcess(command, nil)
	if err != nil {
		return err
	}
	if process != nil {
		if process.Process != nil {
			f.watchManagedProcess(process)
			if err := f.bindManagedProcessExit(process); err != nil {
				return support.NewBrowserLaunchError("绑定 Firefox 进程到 Go 进程退出联动失败", err)
			}
		} else {
			f.mu.Lock()
			f.process = process
			f.mu.Unlock()
		}
	}
	return nil
}

func (f *Firefox) watchManagedProcess(process *exec.Cmd) {
	done := make(chan error, 1)

	f.mu.Lock()
	f.process = process
	f.processDone = done
	f.mu.Unlock()

	go func() {
		err := process.Wait()
		done <- err
		close(done)
		f.handleManagedProcessExit(process)
	}()
}

func (f *Firefox) bindManagedProcessExit(process *exec.Cmd) error {
	if process == nil || process.Process == nil || !f.options.IsCloseBrowserOnExitEnabled() {
		return nil
	}

	release, err := bindProcessKillOnExit(process)
	if err != nil {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.process != process {
		_ = release()
		return nil
	}
	if previous := f.processExitBinding; previous != nil {
		_ = previous()
	}
	f.processExitBinding = release
	return nil
}

func (f *Firefox) handleManagedProcessExit(process *exec.Cmd) {
	f.mu.Lock()
	if f.process != process {
		f.mu.Unlock()
		return
	}

	quitting := f.quitting
	autoProfile := f.autoProfile
	driver := f.driver
	processExitBinding := f.processExitBinding
	f.process = nil
	f.processDone = nil
	f.processExitBinding = nil
	f.autoProfile = ""
	f.driver = nil
	f.sessionID = ""
	f.ownsSession = false
	f.contextIDs = []string{}
	f.contexts = make(map[string]ProbeContextInfo)
	f.clientWindows = []map[string]any{}
	f.mu.Unlock()

	if quitting {
		return
	}

	if driver != nil {
		_ = driver.Stop()
	}
	if processExitBinding != nil {
		_ = processExitBinding()
	}
	if autoProfile != "" {
		_ = os.RemoveAll(autoProfile)
	}
	if f.options != nil {
		_ = f.options.CleanupManagedFPFile()
	}
	f.removeFromRegistry()
}

func (f *Firefox) ensureLaunchPortAvailable() error {
	if f.options.IsAutoPortEnabled() {
		return nil
	}

	host, port, err := splitProbeAddress(f.Address())
	if err != nil {
		return support.NewBrowserLaunchError("Firefox 地址无效", err)
	}
	if !support.IsPortOpen(host, port, time.Second) {
		return nil
	}

	newPort, err := findFirefoxFreePort(port + 1)
	if err != nil {
		return err
	}

	newAddress := fmt.Sprintf("%s:%d", host, newPort)
	if err := f.rebindAddress(newAddress); err != nil {
		return err
	}
	f.options.SetResolvedPort(newPort)
	return nil
}

func (f *Firefox) rebindAddress(newAddress string) error {
	if newAddress == "" || newAddress == f.Address() {
		return nil
	}

	firefoxRegistryMu.Lock()
	defer firefoxRegistryMu.Unlock()

	if existing := firefoxRegistry[newAddress]; existing != nil && existing != f {
		return support.NewBrowserConnectError(
			fmt.Sprintf("Firefox 地址 %s 已存在活动实例", newAddress),
			nil,
		)
	}

	oldAddress := f.Address()
	if current := firefoxRegistry[oldAddress]; current == f {
		delete(firefoxRegistry, oldAddress)
	}
	firefoxRegistry[newAddress] = f

	f.mu.Lock()
	f.address = newAddress
	f.mu.Unlock()
	return nil
}

func (f *Firefox) removeFromRegistry() {
	firefoxRegistryMu.Lock()
	defer firefoxRegistryMu.Unlock()

	address := f.Address()
	if current := firefoxRegistry[address]; current == f {
		delete(firefoxRegistry, address)
	}
}

func (f *Firefox) cleanupAfterFailedInitialization() {
	_ = f.Quit(time.Second, true)
}

func (f *Firefox) baseTimeout() time.Duration {
	return resolveFirefoxTimeoutSeconds(f.options.Timeouts().Base, support.DefaultBaseTimeoutSeconds)
}

func (f *Firefox) pageLoadTimeout() time.Duration {
	return resolveFirefoxTimeoutSeconds(f.options.Timeouts().PageLoad, support.DefaultPageLoadTimeoutSeconds)
}

func (f *Firefox) browserConnectTimeout() time.Duration {
	return time.Duration(support.DefaultBrowserConnectTimeoutSeconds) * time.Second
}

func (f *Firefox) retryInterval() time.Duration {
	interval := f.options.RetryInterval()
	if interval <= 0 {
		interval = 2
	}
	return time.Duration(interval * float64(time.Second))
}

func waitManagedProcessExit(process *exec.Cmd, done <-chan error, timeout time.Duration, force bool) error {
	if process == nil || process.Process == nil || done == nil {
		return nil
	}

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		if !force {
			return nil
		}
		if err := process.Process.Kill(); err != nil && !stderrors.Is(err, os.ErrProcessDone) {
			return err
		}
		select {
		case <-done:
		case <-time.After(timeout):
		}
		return nil
	}
}

func managedProcessExited(process *exec.Cmd) bool {
	if process == nil || process.Process == nil {
		return true
	}
	return process.ProcessState != nil && process.ProcessState.Exited()
}

func findFirefoxFreePort(start int) (int, error) {
	if start < 1 {
		start = support.DefaultPort
	}

	port, err := support.FindFreePort(start, start+100)
	if err != nil {
		return 0, support.NewBrowserLaunchError(
			fmt.Sprintf("在端口范围 %d-%d 中找不到可用调试端口", start, start+99),
			err,
		)
	}
	return port, nil
}

func convertPromptHandler(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func buildOccupiedFirefoxMessage(statusMessage string, errorMessage string) string {
	switch {
	case statusMessage != "" && errorMessage != "":
		return fmt.Sprintf("Firefox BiDi 会话已被占用：%s (%s)", statusMessage, errorMessage)
	case statusMessage != "":
		return fmt.Sprintf("Firefox BiDi 会话已被占用：%s", statusMessage)
	case errorMessage != "":
		return fmt.Sprintf("Firefox BiDi 会话已被占用：%s", errorMessage)
	default:
		return "Firefox BiDi 会话已被占用"
	}
}

func resolveFirefoxTimeoutSeconds(seconds float64, fallback int) time.Duration {
	if seconds <= 0 {
		seconds = float64(fallback)
	}
	return time.Duration(seconds * float64(time.Second))
}

func minFirefoxDuration(left, right time.Duration) time.Duration {
	if left < right {
		return left
	}
	return right
}

func isFirefoxIgnorableShutdownError(err error) bool {
	if err == nil {
		return false
	}

	var disconnected *support.PageDisconnectedError
	if stderrors.As(err, &disconnected) {
		return true
	}

	var bidiErr *support.BiDiError
	if stderrors.As(err, &bidiErr) {
		message := bidiErr.Error()
		return isMaximumActiveSessionsError(bidiErr) || containsAnyFirefoxShutdownText(message)
	}
	return containsAnyFirefoxShutdownText(err.Error())
}

func containsAnyFirefoxShutdownText(message string) bool {
	lowered := strings.ToLower(message)
	return strings.Contains(lowered, "invalid session id") ||
		strings.Contains(lowered, "websocket 连接未建立") ||
		strings.Contains(lowered, "websocket connection closed") ||
		strings.Contains(lowered, "connection reset")
}

func cloneFirefoxStringSlice(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}
