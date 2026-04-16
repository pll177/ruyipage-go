package pages

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"ruyipage-go/internal/base"
	"ruyipage-go/internal/bidi"
	"ruyipage-go/internal/config"
	"ruyipage-go/internal/elements"
	"ruyipage-go/internal/support"
	"ruyipage-go/internal/units"
)

const (
	waitPollInterval    = 100 * time.Millisecond
	elementPollInterval = 200 * time.Millisecond
)

var xpathPickerPreloads sync.Map

type FirefoxBrowser interface {
	Address() string
	SessionID() string
	Driver() *base.BrowserBiDiDriver
	Options() *config.FirefoxOptions
}

type FirefoxBase struct {
	base.BasePage

	mu sync.RWMutex

	browser   FirefoxBrowser
	page      *FirefoxPage
	contextID string
	driver    *base.ContextDriver
	loadMode  config.FirefoxLoadMode

	readyState string
	isLoading  bool

	scroll    *PageScroller
	waiter    *PageWaiter
	rect      *TabRect
	setter    *PageSetter
	states    *PageStates
	listen    *units.Listener
	console   *units.ConsoleListener
	intercept *units.Interceptor
	network   *units.NetworkManager
	downloads *units.DownloadsManager
	events    *units.EventTracker
	realms    *units.RealmTracker
	contexts  *units.ContextManager
	exts      *units.ExtensionManager
	emulate   *units.EmulationManager
	cookies   *units.CookiesSetter
	prefs     *units.PrefsManager
	config    *units.ConfigManager
	local     *units.StorageManager
	session   *units.StorageManager

	lastPromptOpened   map[string]any
	lastPromptClosed   map[string]any
	promptOpen         bool
	promptSubscription string
	uaPreloadScriptID  string
	cspPreloadScriptID string
}

func NewFirefoxBase(browser FirefoxBrowser, contextID string) (*FirefoxBase, error) {
	page := &FirefoxBase{}
	page.BasePage = base.NewBasePage("FirefoxBase", func() string {
		url, err := page.URL()
		if err != nil {
			return ""
		}
		return url
	})
	if err := page.InitContext(browser, contextID); err != nil {
		return nil, err
	}
	return page, nil
}

func (p *FirefoxBase) InitContext(browser FirefoxBrowser, contextID string) error {
	if p == nil {
		return support.NewPageDisconnectedError("FirefoxBase 未初始化", nil)
	}
	if browser == nil || browser.Driver() == nil {
		return support.NewPageDisconnectedError("Firefox browser 未初始化", nil)
	}
	if contextID == "" {
		return support.NewContextLostError("context id 不能为空", nil)
	}

	p.BasePage.SetURLGetter(func() string {
		url, err := p.URL()
		if err != nil {
			return ""
		}
		return url
	})

	p.mu.Lock()
	oldDriver := p.browserDriverLocked()
	oldContextID := p.contextID
	p.browser = browser
	p.contextID = contextID
	p.driver = base.NewContextDriver(browser.Driver(), contextID)
	if options := browser.Options(); options != nil {
		p.loadMode = options.LoadMode()
	} else {
		p.loadMode = config.LoadModeNormal
	}
	p.readyState = ""
	p.isLoading = false
	p.lastPromptOpened = nil
	p.lastPromptClosed = nil
	p.promptOpen = false
	p.promptSubscription = ""
	p.mu.Unlock()

	p.clearPromptTracking(oldDriver, oldContextID)
	if err := p.ensurePromptTracking(); err != nil {
		return err
	}
	_ = p.ensureXPathPickerPreload()
	_ = p.reinjectXPathPickerIfNeeded()
	return nil
}

func (p *FirefoxBase) SetContextID(contextID string) error {
	if p == nil {
		return support.NewPageDisconnectedError("FirefoxBase 未初始化", nil)
	}
	p.mu.RLock()
	browser := p.browser
	p.mu.RUnlock()
	return p.InitContext(browser, contextID)
}

func (p *FirefoxBase) Browser() FirefoxBrowser {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.browser
}

// BrowserDriver 返回浏览器级 driver，供元素对象复用。
func (p *FirefoxBase) BrowserDriver() *base.BrowserBiDiDriver {
	return p.browserDriver()
}

func (p *FirefoxBase) setPageOwner(page *FirefoxPage) {
	if p == nil {
		return
	}
	p.mu.Lock()
	p.page = page
	p.mu.Unlock()
}

func (p *FirefoxBase) pageOwner() *FirefoxPage {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.page
}

func (p *FirefoxBase) ContextID() string {
	if p == nil {
		return ""
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.contextID
}

func (p *FirefoxBase) Driver() *base.ContextDriver {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.driver
}

func (p *FirefoxBase) IsConnected() bool {
	driver := p.browserDriver()
	return driver != nil && driver.IsRunning()
}

func (p *FirefoxBase) Wait() *PageWaiter {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.waiter == nil {
		p.waiter = &PageWaiter{page: p}
	}
	return p.waiter
}

func (p *FirefoxBase) Scroll() *PageScroller {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.scroll == nil {
		p.scroll = &PageScroller{page: p}
	}
	return p.scroll
}

func (p *FirefoxBase) Rect() *TabRect {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.rect == nil {
		p.rect = &TabRect{page: p}
	}
	return p.rect
}

// WindowSize 返回当前窗口尺寸，供高层窗口管理能力复用。
func (p *FirefoxBase) WindowSize() map[string]int {
	if p == nil {
		return map[string]int{"width": 0, "height": 0}
	}
	return p.Rect().WindowSize()
}

// ViewportSize 返回当前视口尺寸，供高层触摸动作能力复用。
func (p *FirefoxBase) ViewportSize() map[string]int {
	if p == nil {
		return map[string]int{"width": 0, "height": 0}
	}
	return p.Rect().ViewportSize()
}

func (p *FirefoxBase) States() *PageStates {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.states == nil {
		p.states = &PageStates{page: p}
	}
	return p.states
}

func (p *FirefoxBase) Set() *PageSetter {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.setter == nil {
		p.setter = &PageSetter{page: p}
	}
	return p.setter
}

func (p *FirefoxBase) CookiesSetter() *units.CookiesSetter {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cookies == nil {
		p.cookies = units.NewCookiesSetter(p)
	}
	return p.cookies
}

func (p *FirefoxBase) Prefs() *units.PrefsManager {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.prefs == nil {
		p.prefs = units.NewPrefsManagerWithResolver(func() string { return p.profilePath() })
	}
	return p.prefs
}

func (p *FirefoxBase) Config() *units.ConfigManager {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.config == nil {
		p.config = units.NewConfigManagerWithResolver(func() string { return p.profilePath() })
		p.config.SetRefresher(p)
	}
	return p.config
}

func (p *FirefoxBase) LocalStorage() *units.StorageManager {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.local == nil {
		p.local = units.NewStorageManager(p, "localStorage")
	}
	return p.local
}

func (p *FirefoxBase) SessionStorage() *units.StorageManager {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.session == nil {
		p.session = units.NewStorageManager(p, "sessionStorage")
	}
	return p.session
}

func (p *FirefoxBase) Listen() *units.Listener {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.listen == nil {
		p.listen = units.NewListener(p)
	}
	return p.listen
}

func (p *FirefoxBase) Console() *units.ConsoleListener {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.console == nil {
		p.console = units.NewConsoleListener(p)
	}
	return p.console
}

func (p *FirefoxBase) Intercept() *units.Interceptor {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.intercept == nil {
		p.intercept = units.NewInterceptor(p)
	}
	return p.intercept
}

func (p *FirefoxBase) Network() *units.NetworkManager {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.network == nil {
		p.network = units.NewNetworkManager(p)
	}
	return p.network
}

func (p *FirefoxBase) Downloads() *units.DownloadsManager {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.downloads == nil {
		p.downloads = units.NewDownloadsManager(p)
	}
	return p.downloads
}

func (p *FirefoxBase) Events() *units.EventTracker {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.events == nil {
		p.events = units.NewEventTracker(p)
	}
	return p.events
}

func (p *FirefoxBase) Realms() *units.RealmTracker {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.realms == nil {
		p.realms = units.NewRealmTracker(p)
	}
	return p.realms
}

func (p *FirefoxBase) Contexts() *units.ContextManager {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.contexts == nil {
		p.contexts = units.NewContextManager(p)
	}
	return p.contexts
}

func (p *FirefoxBase) Extensions() *units.ExtensionManager {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.exts == nil {
		p.exts = units.NewExtensionManager(p)
	}
	return p.exts
}

func (p *FirefoxBase) Emulation() *units.EmulationManager {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.emulate == nil {
		p.emulate = units.NewEmulationManager(p)
	}
	return p.emulate
}

func (p *FirefoxBase) URL() (string, error) {
	value, err := p.RunJSExpr("location.href")
	if err != nil {
		return "", err
	}
	return stringify(value), nil
}

func (p *FirefoxBase) Title() (string, error) {
	value, err := p.RunJSExpr("document.title")
	if err != nil {
		return "", err
	}
	return stringify(value), nil
}

func (p *FirefoxBase) HTML() (string, error) {
	value, err := p.RunJSExpr("document.documentElement ? document.documentElement.outerHTML : ''")
	if err != nil {
		return "", err
	}
	return stringify(value), nil
}

// SEle 在当前页面 HTML 上执行静态解析查找。
func (p *FirefoxBase) SEle(locator any) (elements.StaticNode, error) {
	htmlText, err := p.HTML()
	if err != nil {
		return nil, err
	}
	return elements.MakeStaticElement(htmlText, locator), nil
}

// SEles 在当前页面 HTML 上执行静态解析查找。
func (p *FirefoxBase) SEles(locator any) ([]*elements.StaticElement, error) {
	htmlText, err := p.HTML()
	if err != nil {
		return nil, err
	}
	return elements.MakeStaticElements(htmlText, locator), nil
}

func (p *FirefoxBase) ReadyState() (string, error) {
	value, err := p.RunJSExpr("document.readyState")
	if err != nil {
		return "", err
	}
	state := stringify(value)
	p.mu.Lock()
	p.readyState = state
	p.isLoading = state == "loading"
	p.mu.Unlock()
	return state, nil
}

func (p *FirefoxBase) Navigate(url string, wait string) error {
	if url == "" {
		return support.NewRuyiPageError("URL 不能为空", nil)
	}
	if err := p.ensureInitialized(); err != nil {
		return err
	}
	if wait == "" {
		wait = loadModeWait(p.loadMode)
	}
	p.setLoadingState("loading", true)
	if _, err := bidi.Navigate(p.browserDriver(), p.ContextID(), url, wait, p.pageLoadTimeout()); err != nil {
		return err
	}
	if wait != "none" {
		_, _ = p.ReadyState()
	}
	_ = p.reinjectXPathPickerIfNeeded()
	return nil
}

func (p *FirefoxBase) Get(url string) error {
	return p.Navigate(url, "")
}

func (p *FirefoxBase) Activate() error {
	if err := p.ensureInitialized(); err != nil {
		return err
	}
	_, err := bidi.Activate(p.browserDriver(), p.ContextID(), p.baseTimeout())
	return err
}

func (p *FirefoxBase) Reload(ignoreCache bool, wait string) error {
	if err := p.ensureInitialized(); err != nil {
		return err
	}
	if wait == "" {
		wait = loadModeWait(p.loadMode)
	}
	p.setLoadingState("loading", true)
	if _, err := bidi.Reload(p.browserDriver(), p.ContextID(), ignoreCache, wait, p.pageLoadTimeout()); err != nil {
		return err
	}
	if wait != "none" {
		_, _ = p.ReadyState()
	}
	_ = p.reinjectXPathPickerIfNeeded()
	return nil
}

func (p *FirefoxBase) Refresh() error {
	return p.Reload(false, "")
}

func (p *FirefoxBase) Back() error {
	if err := p.ensureInitialized(); err != nil {
		return err
	}
	p.setLoadingState("loading", true)
	if _, err := bidi.TraverseHistory(p.browserDriver(), p.ContextID(), -1, p.pageLoadTimeout()); err != nil {
		return err
	}
	_ = p.reinjectXPathPickerIfNeeded()
	return nil
}

func (p *FirefoxBase) Forward() error {
	if err := p.ensureInitialized(); err != nil {
		return err
	}
	p.setLoadingState("loading", true)
	if _, err := bidi.TraverseHistory(p.browserDriver(), p.ContextID(), 1, p.pageLoadTimeout()); err != nil {
		return err
	}
	_ = p.reinjectXPathPickerIfNeeded()
	return nil
}

func (p *FirefoxBase) LocateNodes(locator any, maxCount int, startNodes []map[string]any) ([]map[string]any, error) {
	if err := p.ensureInitialized(); err != nil {
		return nil, err
	}
	parsed, err := support.ParseLocator(locator)
	if err != nil {
		return nil, err
	}
	var maxCountPtr *int
	if maxCount > 0 {
		maxCountPtr = &maxCount
	}
	result, err := bidi.LocateNodes(
		p.browserDriver(),
		p.ContextID(),
		parsed,
		maxCountPtr,
		map[string]any{"maxDomDepth": 0, "includeShadowTree": "open"},
		cloneMapSlice(startNodes),
		p.elementFindTimeout(),
	)
	if err != nil {
		if parsed["type"] == "innerText" {
			return p.locateByTextFallback(stringify(parsed["value"]), startNodes)
		}
		return nil, err
	}
	nodes := anyToMapSlice(result["nodes"])
	if len(nodes) == 0 && parsed["type"] == "innerText" {
		return p.locateByTextFallback(stringify(parsed["value"]), startNodes)
	}
	return nodes, nil
}

func (p *FirefoxBase) Ele(locator any, index int, timeout time.Duration) (map[string]any, error) {
	return p.findNode(locator, index, timeout, nil)
}

// Eles 查找所有匹配节点的原始结果。
func (p *FirefoxBase) Eles(locator any, timeout time.Duration) ([]map[string]any, error) {
	return p.findNodes(locator, timeout, nil)
}

// FindElement 查找单个动态元素；支持从指定 startNodes 相对查找。
func (p *FirefoxBase) FindElement(locator any, index int, timeout time.Duration, startNodes []map[string]any) (*elements.FirefoxElement, error) {
	node, err := p.findNode(locator, index, timeout, startNodes)
	if err != nil || node == nil {
		return nil, err
	}
	return elements.FromNode(p, node, buildElementLocatorInfo(locator, startNodes)), nil
}

// FindElements 查找全部动态元素；支持从指定 startNodes 相对查找。
func (p *FirefoxBase) FindElements(locator any, timeout time.Duration, startNodes []map[string]any) ([]*elements.FirefoxElement, error) {
	nodes, err := p.findNodes(locator, timeout, startNodes)
	if err != nil {
		return nil, err
	}
	result := make([]*elements.FirefoxElement, 0, len(nodes))
	locatorInfo := buildElementLocatorInfo(locator, startNodes)
	for _, node := range nodes {
		if element := elements.FromNode(p, node, locatorInfo); element != nil {
			result = append(result, element)
		}
	}
	return result, nil
}

func (p *FirefoxBase) findNode(locator any, index int, timeout time.Duration, startNodes []map[string]any) (map[string]any, error) {
	if timeout <= 0 {
		timeout = p.elementFindTimeout()
	}
	if index == 0 {
		index = 1
	}
	deadline := time.Now().Add(timeout)
	for {
		nodes, err := p.LocateNodes(locator, 0, cloneMapSlice(startNodes))
		if err != nil {
			return nil, err
		}
		if len(nodes) > 0 {
			resolvedIndex := index
			if resolvedIndex > 0 {
				resolvedIndex--
			} else {
				resolvedIndex = len(nodes) + resolvedIndex
			}
			if resolvedIndex >= 0 && resolvedIndex < len(nodes) {
				return cloneMap(nodes[resolvedIndex]), nil
			}
		}
		if time.Now().After(deadline) {
			if support.Settings.RaiseWhenEleNotFound {
				return nil, support.NewElementNotFoundError(fmt.Sprintf("未找到元素: %#v", locator), nil)
			}
			return nil, nil
		}
		time.Sleep(elementPollInterval)
	}
}

func (p *FirefoxBase) findNodes(locator any, timeout time.Duration, startNodes []map[string]any) ([]map[string]any, error) {
	if timeout <= 0 {
		timeout = p.elementFindTimeout()
	}
	deadline := time.Now().Add(timeout)
	for {
		nodes, err := p.LocateNodes(locator, 0, cloneMapSlice(startNodes))
		if err != nil {
			return nil, err
		}
		if len(nodes) > 0 {
			return cloneMapSlice(nodes), nil
		}
		if time.Now().After(deadline) {
			return []map[string]any{}, nil
		}
		time.Sleep(elementPollInterval)
	}
}

func (p *FirefoxBase) RunJS(script string, args ...any) (any, error) {
	return p.runJS(script, nil, "", 0, args...)
}

func (p *FirefoxBase) RunJSInSandbox(script string, sandbox string, args ...any) (any, error) {
	return p.runJS(script, nil, sandbox, 0, args...)
}

func (p *FirefoxBase) RunJSExpr(expression string, args ...any) (any, error) {
	asExpr := true
	return p.runJS(expression, &asExpr, "", 0, args...)
}

func (p *FirefoxBase) RunJSExprInSandbox(expression string, sandbox string, args ...any) (any, error) {
	asExpr := true
	return p.runJS(expression, &asExpr, sandbox, 0, args...)
}

func (p *FirefoxBase) RunJSRaw(script string, args ...any) (units.ScriptResult, error) {
	return p.runJSRaw(script, nil, "", 0, args...)
}

func (p *FirefoxBase) runJS(script string, asExpr *bool, sandbox string, timeout time.Duration, args ...any) (any, error) {
	result, err := p.runJSRaw(script, asExpr, sandbox, timeout, args...)
	if err != nil {
		return nil, err
	}
	return support.ParseBiDiValue(result.Result.Raw), nil
}

func (p *FirefoxBase) runJSRaw(script string, asExpr *bool, sandbox string, timeout time.Duration, args ...any) (units.ScriptResult, error) {
	if err := p.ensureInitialized(); err != nil {
		return units.ScriptResult{}, err
	}
	script = strings.TrimSpace(script)
	if timeout <= 0 {
		timeout = p.scriptTimeout()
	}
	target := map[string]any{"context": p.ContextID()}
	useExpr := shouldUseExpr(script, asExpr, len(args) > 0)
	resolvedSandbox := resolveScriptSandbox(sandbox)
	if useExpr {
		result, err := bidi.Evaluate(p.browserDriver(), script, target, nil, resolvedSandbox, "", nil, false, timeout)
		return units.NewScriptResultFromData(result), err
	}
	result, err := bidi.CallFunction(
		p.browserDriver(),
		normalizeFunction(script),
		target,
		args,
		nil,
		nil,
		resolvedSandbox,
		"",
		nil,
		false,
		timeout,
	)
	return units.NewScriptResultFromData(result), err
}

func (p *FirefoxBase) EvalHandle(expression string, awaitPromise bool) (units.ScriptResult, error) {
	if err := p.ensureInitialized(); err != nil {
		return units.ScriptResult{}, err
	}
	result, err := bidi.Evaluate(
		p.browserDriver(),
		expression,
		map[string]any{"context": p.ContextID()},
		&awaitPromise,
		"",
		"root",
		nil,
		false,
		p.scriptTimeout(),
	)
	return units.NewScriptResultFromData(result), err
}

func (p *FirefoxBase) GetRealms(typeName string) ([]units.RealmInfo, error) {
	if err := p.ensureInitialized(); err != nil {
		return nil, err
	}
	realms, err := bidi.GetRealms(p.browserDriver(), p.ContextID(), typeName, p.baseTimeout())
	if err != nil {
		return nil, err
	}
	return units.NewRealmInfosFromData(realms), nil
}

func (p *FirefoxBase) AddPreloadScript(script string) (units.PreloadScript, error) {
	if err := p.ensureInitialized(); err != nil {
		return units.PreloadScript{}, err
	}
	result, err := bidi.AddPreloadScript(p.browserDriver(), script, nil, []string{p.ContextID()}, "", p.baseTimeout())
	return units.NewPreloadScriptFromData(result), err
}

func resolveScriptSandbox(sandbox string) string {
	sandbox = strings.TrimSpace(sandbox)
	if sandbox == "" {
		return "root"
	}
	return sandbox
}

func (p *FirefoxBase) RemovePreloadScript(scriptID string) error {
	if err := p.ensureInitialized(); err != nil {
		return err
	}
	return bidi.RemovePreloadScript(p.browserDriver(), scriptID, p.baseTimeout())
}

func (p *FirefoxBase) DisownHandles(handles []string) error {
	if err := p.ensureInitialized(); err != nil {
		return err
	}
	return bidi.Disown(p.browserDriver(), cloneStrings(handles), map[string]any{"context": p.ContextID()}, p.baseTimeout())
}

func (p *FirefoxBase) Cookies(allInfo bool) ([]units.CookieInfo, error) {
	if err := p.ensureInitialized(); err != nil {
		return nil, err
	}
	result, err := bidi.GetCookies(p.browserDriver(), nil, map[string]any{"context": p.ContextID()}, p.baseTimeout())
	if err != nil {
		value, jsErr := p.RunJSExpr("document.cookie")
		if jsErr != nil {
			return nil, jsErr
		}
		return units.NewCookieInfos(support.CookieStrToList(stringify(value))), nil
	}
	raw := anyToMapSlice(result["cookies"])
	cookies := make([]units.CookieInfo, 0, len(raw))
	for _, cookie := range raw {
		cookies = append(cookies, units.NewCookieInfo(cookieView(cookie, allInfo)))
	}
	return cookies, nil
}

func (p *FirefoxBase) SetCookie(cookie map[string]any) error {
	if err := p.ensureInitialized(); err != nil {
		return err
	}
	payload := cloneMap(cookie)
	if payload == nil {
		payload = map[string]any{}
	}
	if valueMap, ok := payload["value"].(map[string]any); !ok || valueMap["type"] == nil {
		payload["value"] = support.SerializeBiDiValue(payload["value"])
	}
	_, err := bidi.SetCookie(p.browserDriver(), payload, map[string]any{"context": p.ContextID()}, p.baseTimeout())
	return err
}

func (p *FirefoxBase) SetCookies(cookies any) error {
	switch typed := cookies.(type) {
	case map[string]any:
		return p.SetCookie(typed)
	case units.CookieInfo:
		return p.SetCookie(cookieInfoToMap(typed))
	case *units.CookieInfo:
		if typed == nil {
			return nil
		}
		return p.SetCookie(cookieInfoToMap(*typed))
	case []map[string]any:
		for _, cookie := range typed {
			if err := p.SetCookie(cookie); err != nil {
				return err
			}
		}
		return nil
	case []units.CookieInfo:
		for _, cookie := range typed {
			if err := p.SetCookie(cookieInfoToMap(cookie)); err != nil {
				return err
			}
		}
		return nil
	case []any:
		for _, item := range typed {
			if err := p.SetCookies(item); err != nil {
				return err
			}
		}
		return nil
	default:
		return support.NewRuyiPageError("cookies 参数必须是 map[string]any、CookieInfo 或它们的切片", nil)
	}
}

func (p *FirefoxBase) DeleteCookies(filter map[string]any) error {
	if err := p.ensureInitialized(); err != nil {
		return err
	}
	_, err := bidi.DeleteCookies(p.browserDriver(), cloneMap(filter), map[string]any{"context": p.ContextID()}, p.baseTimeout())
	return err
}

func (p *FirefoxBase) Screenshot(path string, fullPage bool) ([]byte, error) {
	if err := p.ensureInitialized(); err != nil {
		return nil, err
	}
	origin := "viewport"
	if fullPage {
		origin = "document"
	}
	result, err := bidi.CaptureScreenshot(p.browserDriver(), p.ContextID(), origin, nil, nil, p.baseTimeout())
	if err != nil {
		return nil, err
	}
	data, err := decodeBase64(result["data"])
	if err != nil {
		return nil, err
	}
	if path != "" {
		if err := writeBytes(path, data); err != nil {
			return nil, err
		}
	}
	return data, nil
}

func (p *FirefoxBase) PDF(path string, options map[string]any) ([]byte, error) {
	if err := p.ensureInitialized(); err != nil {
		return nil, err
	}
	background, margin, orientation, page, ranges, scale, shrink := parsePDFOptions(options)
	result, err := bidi.Print(p.browserDriver(), p.ContextID(), background, margin, orientation, page, ranges, scale, shrink, p.baseTimeout())
	if err != nil {
		return nil, err
	}
	data, err := decodeBase64(result["data"])
	if err != nil {
		return nil, err
	}
	if path != "" {
		if err := writeBytes(path, data); err != nil {
			return nil, err
		}
	}
	return data, nil
}

func (p *FirefoxBase) SavePage(path string) error {
	if path == "" {
		return support.NewRuyiPageError("保存路径不能为空", nil)
	}
	html, err := p.HTML()
	if err != nil {
		return err
	}
	return writeBytes(path, []byte(html))
}

func (p *FirefoxBase) PromptOpen() bool {
	if p == nil {
		return false
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.promptOpen
}

func (p *FirefoxBase) PromptInfo() map[string]any {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if !p.promptOpen || p.lastPromptOpened == nil {
		return nil
	}
	return cloneMap(p.lastPromptOpened)
}

func (p *FirefoxBase) LastPromptOpened() map[string]any {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return cloneMap(p.lastPromptOpened)
}

func (p *FirefoxBase) LastPromptClosed() map[string]any {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return cloneMap(p.lastPromptClosed)
}

func (p *FirefoxBase) WaitPrompt(timeout time.Duration) (map[string]any, error) {
	if err := p.ensureInitialized(); err != nil {
		return nil, err
	}
	if timeout <= 0 {
		timeout = p.baseTimeout()
	}
	info, matched, err := support.WaitUntil(func() (map[string]any, bool, error) {
		prompt := p.PromptInfo()
		return prompt, prompt != nil, nil
	}, timeout, waitPollInterval)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, support.NewWaitTimeoutError(fmt.Sprintf("等待 prompt 超时 (%s)", timeout), nil)
	}
	return info, nil
}

func (p *FirefoxBase) HandlePrompt(accept bool, text *string, timeout time.Duration) error {
	if _, err := p.WaitPrompt(timeout); err != nil {
		return err
	}
	_, err := bidi.HandleUserPrompt(p.browserDriver(), p.ContextID(), accept, text, p.baseTimeout())
	return err
}

func (p *FirefoxBase) WaitReadyState(target string, timeout time.Duration) error {
	if err := p.ensureInitialized(); err != nil {
		return err
	}
	if timeout <= 0 {
		timeout = p.pageLoadTimeout()
	}
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" {
		target = "complete"
	}
	_, matched, err := support.WaitUntil(func() (string, bool, error) {
		state, err := p.ReadyState()
		if err != nil {
			return "", false, err
		}
		return state, readyStateReached(state, target), nil
	}, timeout, waitPollInterval)
	if err != nil {
		return err
	}
	if !matched {
		return support.NewWaitTimeoutError(fmt.Sprintf("等待 readyState=%s 超时 (%s)", target, timeout), nil)
	}
	return nil
}

func (p *FirefoxBase) WaitLoadComplete(timeout time.Duration) error {
	return p.WaitReadyState("complete", timeout)
}

func (p *FirefoxBase) WaitURLContains(fragment string, timeout time.Duration) error {
	if err := p.ensureInitialized(); err != nil {
		return err
	}
	if timeout <= 0 {
		timeout = p.pageLoadTimeout()
	}
	_, matched, err := support.WaitUntil(func() (string, bool, error) {
		url, err := p.URL()
		if err != nil {
			return "", false, err
		}
		return url, strings.Contains(url, fragment), nil
	}, timeout, waitPollInterval)
	if err != nil {
		return err
	}
	if !matched {
		return support.NewWaitTimeoutError(fmt.Sprintf("等待 URL 包含 %q 超时 (%s)", fragment, timeout), nil)
	}
	return nil
}

func (p *FirefoxBase) WaitTitleContains(fragment string, timeout time.Duration) error {
	if err := p.ensureInitialized(); err != nil {
		return err
	}
	if timeout <= 0 {
		timeout = p.pageLoadTimeout()
	}
	_, matched, err := support.WaitUntil(func() (string, bool, error) {
		title, err := p.Title()
		if err != nil {
			return "", false, err
		}
		return title, strings.Contains(title, fragment), nil
	}, timeout, waitPollInterval)
	if err != nil {
		return err
	}
	if !matched {
		return support.NewWaitTimeoutError(fmt.Sprintf("等待标题包含 %q 超时 (%s)", fragment, timeout), nil)
	}
	return nil
}

// GetFrame 获取当前上下文下的单个直接子 frame。
//
// locatorOrIndexOrContext 支持：
//   - nil：返回第一个 child context
//   - 整数：按 0-based 子 frame 序号获取
//   - string：先按 context id 精确匹配，未命中时回退为 locator
//   - 其他 locator 兼容类型：按 iframe 元素定位并用 src/url 匹配
func (p *FirefoxBase) GetFrame(locatorOrIndexOrContext any) (*FirefoxFrame, error) {
	children, err := p.frameChildren()
	if err != nil {
		return nil, err
	}
	if len(children) == 0 {
		return nil, nil
	}

	if locatorOrIndexOrContext == nil {
		return p.newChildFrame(stringify(children[0]["context"]))
	}

	if index, ok := frameIndexFromAny(locatorOrIndexOrContext); ok {
		if index < 0 || index >= len(children) {
			return nil, nil
		}
		return p.newChildFrame(stringify(children[index]["context"]))
	}

	if contextID, ok := locatorOrIndexOrContext.(string); ok && contextID != "" {
		for _, child := range children {
			if stringify(child["context"]) == contextID {
				return p.newChildFrame(contextID)
			}
		}
	}

	ele, err := p.Ele(locatorOrIndexOrContext, 1, p.elementFindTimeout())
	if err != nil {
		return nil, err
	}
	if ele == nil {
		return nil, nil
	}

	eleSrc, err := p.frameElementSource(ele)
	if err != nil {
		return nil, err
	}
	for _, child := range children {
		childContextID := stringify(child["context"])
		childURL := stringify(child["url"])
		if eleSrc != "" && strings.Contains(childURL, eleSrc) {
			return p.newChildFrame(childContextID)
		}
	}

	return p.newChildFrame(stringify(children[0]["context"]))
}

// GetFrames 获取当前上下文的全部直接子 frame。
func (p *FirefoxBase) GetFrames() ([]*FirefoxFrame, error) {
	children, err := p.frameChildren()
	if err != nil {
		return nil, err
	}

	frames := make([]*FirefoxFrame, 0, len(children))
	for _, child := range children {
		frame, err := p.newChildFrame(stringify(child["context"]))
		if err != nil {
			return nil, err
		}
		if frame != nil {
			frames = append(frames, frame)
		}
	}
	return frames, nil
}

func (p *FirefoxBase) ensureInitialized() error {
	if p == nil {
		return support.NewPageDisconnectedError("FirefoxBase 未初始化", nil)
	}
	if p.ContextID() == "" {
		return support.NewContextLostError("context id 不能为空", nil)
	}
	driver := p.browserDriver()
	if driver == nil || !driver.IsRunning() {
		return support.NewPageDisconnectedError("Firefox driver 未连接", nil)
	}
	return nil
}

func (p *FirefoxBase) ensurePromptTracking() error {
	driver := p.browserDriver()
	if driver == nil {
		return support.NewPageDisconnectedError("Firefox driver 未连接", nil)
	}
	contextID := p.ContextID()
	subscription, err := bidi.Subscribe(driver, []string{
		"browsingContext.userPromptOpened",
		"browsingContext.userPromptClosed",
	}, []string{contextID}, p.baseTimeout())
	if err != nil {
		return err
	}
	if err := driver.SetCallback("browsingContext.userPromptOpened", func(params map[string]any) {
		p.mu.Lock()
		p.lastPromptOpened = cloneMap(params)
		p.promptOpen = true
		p.mu.Unlock()
	}, contextID, true); err != nil {
		return err
	}
	if err := driver.SetCallback("browsingContext.userPromptClosed", func(params map[string]any) {
		p.mu.Lock()
		p.lastPromptClosed = cloneMap(params)
		p.promptOpen = false
		p.mu.Unlock()
	}, contextID, true); err != nil {
		return err
	}
	p.mu.Lock()
	p.promptSubscription = subscription.Subscription
	p.mu.Unlock()
	return nil
}

func (p *FirefoxBase) clearPromptTracking(driver *base.BrowserBiDiDriver, contextID string) {
	if driver == nil || contextID == "" {
		return
	}
	p.mu.Lock()
	subscription := p.promptSubscription
	p.promptSubscription = ""
	p.mu.Unlock()
	driver.RemoveCallback("browsingContext.userPromptOpened", contextID, true)
	driver.RemoveCallback("browsingContext.userPromptClosed", contextID, true)
	if subscription != "" {
		_ = bidi.Unsubscribe(driver, nil, nil, []string{subscription}, p.baseTimeout())
	}
}

func (p *FirefoxBase) ensureXPathPickerPreload() error {
	options := p.options()
	if options == nil || !options.XPathPickerEnabled() {
		return nil
	}
	key := p.xpathPickerKey()
	if key == "" {
		return nil
	}
	if _, exists := xpathPickerPreloads.Load(key); exists {
		return nil
	}
	result, err := bidi.AddPreloadScript(p.browserDriver(), xpathPickerScript, nil, nil, "", p.baseTimeout())
	if err != nil {
		return err
	}
	xpathPickerPreloads.Store(key, result.Script)
	return nil
}

func (p *FirefoxBase) reinjectXPathPickerIfNeeded() error {
	options := p.options()
	if options == nil || !options.XPathPickerEnabled() {
		return nil
	}
	if _, err := p.RunJSExpr("(" + xpathPickerScript + ")()"); err != nil {
		return err
	}
	_, err := p.RunJS(xpathPickerBridgeScript, xpathPickerScript)
	return err
}

func (p *FirefoxBase) locateByTextFallback(text string, startNodes []map[string]any) ([]map[string]any, error) {
	args := []any{text}
	if len(startNodes) > 0 {
		args = append(args, cloneMap(startNodes[0]))
	}
	result, err := bidi.CallFunction(
		p.browserDriver(),
		textFindFunction,
		map[string]any{"context": p.ContextID()},
		args,
		nil,
		nil,
		"",
		"root",
		map[string]any{"maxDomDepth": 0, "includeShadowTree": "open"},
		false,
		p.elementFindTimeout(),
	)
	if err != nil {
		return nil, err
	}
	switch typed := support.ParseBiDiValue(result.Result.Raw).(type) {
	case []any:
		return anySliceToMapSlice(typed), nil
	case []map[string]any:
		return cloneMapSlice(typed), nil
	default:
		return []map[string]any{}, nil
	}
}

func (p *FirefoxBase) options() *config.FirefoxOptions {
	if p == nil {
		return config.NewFirefoxOptions()
	}
	p.mu.RLock()
	browser := p.browser
	p.mu.RUnlock()
	if browser == nil {
		return config.NewFirefoxOptions()
	}
	options := browser.Options()
	if options == nil {
		return config.NewFirefoxOptions()
	}
	return options
}

func (p *FirefoxBase) profilePath() string {
	return p.options().ProfilePath()
}

// DefaultUserContext 返回当前页面配置中的默认 user context。
func (p *FirefoxBase) DefaultUserContext() string {
	return p.options().UserContext()
}

// ApplyUserAgentOverride 通过 preload script 回退方式应用 UA 覆盖。
func (p *FirefoxBase) ApplyUserAgentOverride(userAgent string) error {
	return p.setUserAgent(userAgent)
}

func (p *FirefoxBase) browserDriver() *base.BrowserBiDiDriver {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.browserDriverLocked()
}

func (p *FirefoxBase) browserDriverLocked() *base.BrowserBiDiDriver {
	if p.browser == nil {
		return nil
	}
	return p.browser.Driver()
}

func (p *FirefoxBase) baseTimeout() time.Duration {
	seconds := p.options().Timeouts().Base
	if seconds <= 0 {
		seconds = support.Settings.BiDiTimeout
	}
	if seconds <= 0 {
		seconds = float64(support.DefaultBiDiTimeoutSeconds)
	}
	return secondsDuration(seconds)
}

// BaseTimeout 返回基础超时，供元素对象复用。
func (p *FirefoxBase) BaseTimeout() time.Duration {
	return p.baseTimeout()
}

func (p *FirefoxBase) pageLoadTimeout() time.Duration {
	seconds := p.options().Timeouts().PageLoad
	if seconds <= 0 {
		seconds = support.Settings.PageLoadTimeout
	}
	if seconds <= 0 {
		seconds = float64(support.DefaultPageLoadTimeoutSeconds)
	}
	return secondsDuration(seconds)
}

func (p *FirefoxBase) scriptTimeout() time.Duration {
	seconds := p.options().Timeouts().Script
	if seconds <= 0 {
		seconds = support.Settings.ScriptTimeout
	}
	if seconds <= 0 {
		seconds = float64(support.DefaultScriptTimeoutSeconds)
	}
	return secondsDuration(seconds)
}

// ScriptTimeout 返回脚本超时，供元素对象复用。
func (p *FirefoxBase) ScriptTimeout() time.Duration {
	return p.scriptTimeout()
}

func (p *FirefoxBase) elementFindTimeout() time.Duration {
	seconds := support.Settings.ElementFindTimeout
	if seconds <= 0 {
		seconds = float64(support.DefaultElementFindTimeoutSeconds)
	}
	return secondsDuration(seconds)
}

// ElementFindTimeout 返回元素查找超时，供元素对象复用。
func (p *FirefoxBase) ElementFindTimeout() time.Duration {
	return p.elementFindTimeout()
}

func (p *FirefoxBase) xpathPickerKey() string {
	if p == nil {
		return ""
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.browser == nil {
		return ""
	}
	return p.browser.Address() + "|" + p.browser.SessionID()
}

func (p *FirefoxBase) setLoadingState(state string, isLoading bool) {
	p.mu.Lock()
	p.readyState = state
	p.isLoading = isLoading
	p.mu.Unlock()
}

func shouldUseExpr(script string, asExpr *bool, hasArgs bool) bool {
	if asExpr != nil {
		return *asExpr
	}
	if hasArgs {
		return false
	}
	switch {
	case strings.HasPrefix(script, "return "):
		return false
	case strings.HasPrefix(script, "function"),
		strings.HasPrefix(script, "async function"):
		return false
	case strings.HasPrefix(script, "("):
		return true
	case strings.Contains(script, ";"),
		strings.Contains(script, "\n"):
		return false
	default:
		return true
	}
}

func normalizeFunction(script string) string {
	script = strings.TrimSpace(script)
	switch {
	case strings.HasPrefix(script, "function"),
		strings.HasPrefix(script, "async function"),
		strings.HasPrefix(script, "("),
		strings.Contains(script, "=>"):
		return script
	case strings.HasPrefix(script, "return "),
		strings.Contains(script, ";"),
		strings.Contains(script, "\n"):
		return "function(){" + script + "}"
	default:
		return "function(){ return (" + script + "); }"
	}
}

func loadModeWait(mode config.FirefoxLoadMode) string {
	switch mode {
	case config.LoadModeEager:
		return "interactive"
	case config.LoadModeNone:
		return "none"
	default:
		return "complete"
	}
}

func readyStateReached(state string, target string) bool {
	switch target {
	case "interactive":
		return state == "interactive" || state == "complete"
	case "complete":
		return state == "complete"
	default:
		return state == target
	}
}

func secondsDuration(seconds float64) time.Duration {
	if seconds <= 0 {
		return time.Second
	}
	return time.Duration(seconds * float64(time.Second))
}

func parsePDFOptions(options map[string]any) (*bool, map[string]any, string, map[string]any, []string, *float64, *bool) {
	if options == nil {
		return nil, nil, "", nil, nil, nil, nil
	}
	var backgroundPtr *bool
	if background, ok := options["background"].(bool); ok {
		backgroundPtr = &background
	}
	var scalePtr *float64
	if scale, ok := toFloat64(options["scale"]); ok {
		scalePtr = &scale
	}
	var shrinkPtr *bool
	if shrink, ok := options["shrinkToFit"].(bool); ok {
		shrinkPtr = &shrink
	}
	return backgroundPtr, cloneMapFromAny(options["margin"]), stringify(options["orientation"]), cloneMapFromAny(options["page"]), anyToStrings(options["pageRanges"]), scalePtr, shrinkPtr
}

func cookieView(cookie map[string]any, allInfo bool) map[string]any {
	if allInfo {
		return cloneMap(cookie)
	}
	return map[string]any{
		"name":   cookie["name"],
		"value":  cookie["value"],
		"domain": cookie["domain"],
		"path":   cookie["path"],
	}
}

func cookieInfoToMap(cookie units.CookieInfo) map[string]any {
	if len(cookie.Raw) > 0 {
		return cloneMap(cookie.Raw)
	}

	result := map[string]any{
		"name":  cookie.Name,
		"value": cookie.Value,
	}
	if cookie.Domain != "" {
		result["domain"] = cookie.Domain
	}
	if cookie.Path != "" {
		result["path"] = cookie.Path
	}
	if cookie.HTTPOnly {
		result["httpOnly"] = cookie.HTTPOnly
	}
	if cookie.Secure {
		result["secure"] = cookie.Secure
	}
	if cookie.SameSite != "" {
		result["sameSite"] = cookie.SameSite
	}
	if cookie.Expiry != nil {
		result["expiry"] = cookie.Expiry
	}
	return result
}

func cloneMapFromAny(value any) map[string]any {
	if typed, ok := value.(map[string]any); ok {
		return cloneMap(typed)
	}
	return nil
}

func anyToMapSlice(value any) []map[string]any {
	switch typed := value.(type) {
	case []map[string]any:
		return cloneMapSlice(typed)
	case []any:
		return anySliceToMapSlice(typed)
	default:
		return []map[string]any{}
	}
}

func anySliceToMapSlice(values []any) []map[string]any {
	result := make([]map[string]any, 0, len(values))
	for _, value := range values {
		if mapped, ok := value.(map[string]any); ok {
			result = append(result, cloneMap(mapped))
		}
	}
	return result
}

func cloneMapSlice(values []map[string]any) []map[string]any {
	if len(values) == 0 {
		return []map[string]any{}
	}
	cloned := make([]map[string]any, len(values))
	for index, value := range values {
		cloned[index] = cloneMap(value)
	}
	return cloned
}

func cloneMap(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}
	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

func buildElementLocatorInfo(locator any, startNodes []map[string]any) map[string]any {
	info := map[string]any{
		"locator": locator,
	}
	if len(startNodes) > 0 {
		info["startNodes"] = cloneMapSlice(startNodes)
	}
	return info
}

func anyToStrings(value any) []string {
	switch typed := value.(type) {
	case []string:
		return cloneStrings(typed)
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			result = append(result, stringify(item))
		}
		return result
	default:
		return nil
	}
}

func frameIndexFromAny(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int8:
		return int(typed), true
	case int16:
		return int(typed), true
	case int32:
		return int(typed), true
	case int64:
		return int(typed), true
	case uint:
		return int(typed), true
	case uint8:
		return int(typed), true
	case uint16:
		return int(typed), true
	case uint32:
		return int(typed), true
	case uint64:
		if typed > uint64(^uint(0)>>1) {
			return 0, false
		}
		return int(typed), true
	default:
		return 0, false
	}
}

func toFloat64(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	default:
		return 0, false
	}
}

func stringify(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case nil:
		return ""
	default:
		return fmt.Sprint(typed)
	}
}

func decodeBase64(value any) ([]byte, error) {
	text := stringify(value)
	if text == "" {
		return []byte{}, nil
	}
	return base64.StdEncoding.DecodeString(text)
}

func writeBytes(path string, data []byte) error {
	absPath, err := filepath.Abs(path)
	if err == nil {
		path = absPath
	}
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0o644)
}

func (p *FirefoxBase) frameChildren() ([]map[string]any, error) {
	if err := p.ensureInitialized(); err != nil {
		return nil, err
	}

	result, err := bidi.GetTree(p.browserDriver(), nil, p.ContextID(), p.baseTimeout())
	if err != nil {
		return nil, err
	}

	contexts := anyToMapSlice(result["contexts"])
	if len(contexts) == 0 {
		return []map[string]any{}, nil
	}
	return anyToMapSlice(contexts[0]["children"]), nil
}

func (p *FirefoxBase) newChildFrame(contextID string) (*FirefoxFrame, error) {
	if contextID == "" {
		return nil, nil
	}
	return NewFirefoxFrame(p.Browser(), contextID, p, p.pageOwner())
}

func (p *FirefoxBase) frameElementSource(node map[string]any) (string, error) {
	value, err := p.RunJS("(el) => el.getAttribute('src') || ''", node)
	if err != nil {
		return "", err
	}
	return stringify(value), nil
}
