package ruyipage

import (
	"sync"
	"time"

	internalpages "ruyipage-go/internal/pages"
	internalunits "ruyipage-go/internal/units"
)

// FirefoxBase 是页面、tab、frame 共享的公开页面基类包装层。
type FirefoxBase struct {
	inner *internalpages.FirefoxBase
	page  *FirefoxPage

	managersMu sync.Mutex
	scroll     *PageScroller
	rect       *TabRect
	setter     *PageSetter
	states     *PageStates
	waiter     *PageWaiter
	actions    *internalunits.Actions
	touch      *internalunits.TouchActions
	window     *internalunits.WindowManager
	navigation *internalunits.NavigationTracker
}

func newFirefoxBaseFromInner(inner *internalpages.FirefoxBase) *FirefoxBase {
	if inner == nil {
		return nil
	}
	return &FirefoxBase{inner: inner}
}

func (p *FirefoxBase) String() string {
	if p == nil || p.inner == nil {
		return "<FirefoxBase >"
	}
	return p.inner.String()
}

func (p *FirefoxBase) ContextID() string {
	if p == nil || p.inner == nil {
		return ""
	}
	return p.inner.ContextID()
}

func (p *FirefoxBase) IsConnected() bool {
	if p == nil || p.inner == nil {
		return false
	}
	return p.inner.IsConnected()
}

func (p *FirefoxBase) URL() (string, error) {
	if p == nil || p.inner == nil {
		return "", nil
	}
	return p.inner.URL()
}

func (p *FirefoxBase) Title() (string, error) {
	if p == nil || p.inner == nil {
		return "", nil
	}
	return p.inner.Title()
}

func (p *FirefoxBase) HTML() (string, error) {
	if p == nil || p.inner == nil {
		return "", nil
	}
	return p.inner.HTML()
}

func (p *FirefoxBase) SEle(locator any) (StaticNode, error) {
	if p == nil || p.inner == nil {
		return &NoneElement{}, nil
	}
	node, err := p.inner.SEle(locator)
	if err != nil {
		return nil, err
	}
	return wrapStaticNode(node), nil
}

func (p *FirefoxBase) SEles(locator any) ([]*StaticElement, error) {
	if p == nil || p.inner == nil {
		return []*StaticElement{}, nil
	}
	nodes, err := p.inner.SEles(locator)
	if err != nil {
		return nil, err
	}
	return wrapStaticElements(nodes), nil
}

func (p *FirefoxBase) ReadyState() (string, error) {
	if p == nil || p.inner == nil {
		return "", nil
	}
	return p.inner.ReadyState()
}

func (p *FirefoxBase) Navigate(url string, wait string) error {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.Navigate(url, wait)
}

func (p *FirefoxBase) Get(url string) error {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.Get(url)
}

func (p *FirefoxBase) Activate() error {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.Activate()
}

func (p *FirefoxBase) Reload(ignoreCache bool, wait string) error {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.Reload(ignoreCache, wait)
}

func (p *FirefoxBase) Refresh() error {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.Refresh()
}

func (p *FirefoxBase) Back() error {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.Back()
}

func (p *FirefoxBase) Forward() error {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.Forward()
}

func (p *FirefoxBase) RunJS(script string, args ...any) (any, error) {
	if p == nil || p.inner == nil {
		return nil, nil
	}
	return p.inner.RunJS(script, args...)
}

func (p *FirefoxBase) RunJSInSandbox(script string, sandbox string, args ...any) (any, error) {
	if p == nil || p.inner == nil {
		return nil, nil
	}
	return p.inner.RunJSInSandbox(script, sandbox, args...)
}

func (p *FirefoxBase) RunJSExpr(expression string, args ...any) (any, error) {
	if p == nil || p.inner == nil {
		return nil, nil
	}
	return p.inner.RunJSExpr(expression, args...)
}

func (p *FirefoxBase) RunJSExprInSandbox(expression string, sandbox string, args ...any) (any, error) {
	if p == nil || p.inner == nil {
		return nil, nil
	}
	return p.inner.RunJSExprInSandbox(expression, sandbox, args...)
}

func (p *FirefoxBase) RunJSRaw(script string, args ...any) (ScriptResult, error) {
	if p == nil || p.inner == nil {
		return ScriptResult{}, nil
	}
	return p.inner.RunJSRaw(script, args...)
}

func (p *FirefoxBase) EvalHandle(expression string, awaitPromise bool) (ScriptResult, error) {
	if p == nil || p.inner == nil {
		return ScriptResult{}, nil
	}
	return p.inner.EvalHandle(expression, awaitPromise)
}

func (p *FirefoxBase) GetRealms(typeName string) ([]RealmInfo, error) {
	if p == nil || p.inner == nil {
		return nil, nil
	}
	return p.inner.GetRealms(typeName)
}

func (p *FirefoxBase) AddPreloadScript(script string) (PreloadScript, error) {
	if p == nil || p.inner == nil {
		return PreloadScript{}, nil
	}
	return p.inner.AddPreloadScript(script)
}

func (p *FirefoxBase) RemovePreloadScript(scriptID string) error {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.RemovePreloadScript(scriptID)
}

func (p *FirefoxBase) DisownHandles(handles []string) error {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.DisownHandles(handles)
}

func (p *FirefoxBase) Cookies(allInfo bool) ([]CookieInfo, error) {
	if p == nil || p.inner == nil {
		return []CookieInfo{}, nil
	}
	return p.inner.Cookies(allInfo)
}

func (p *FirefoxBase) SetCookie(cookie map[string]any) error {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.SetCookie(cookie)
}

func (p *FirefoxBase) SetCookies(cookies any) error {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.SetCookies(cookies)
}

func (p *FirefoxBase) DeleteCookies(filter map[string]any) error {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.DeleteCookies(filter)
}

func (p *FirefoxBase) CookiesSetter() *CookiesSetter {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.CookiesSetter()
}

func (p *FirefoxBase) LocalStorage() *StorageManager {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.LocalStorage()
}

func (p *FirefoxBase) SessionStorage() *StorageManager {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.SessionStorage()
}

func (p *FirefoxBase) Screenshot(path string, fullPage bool) ([]byte, error) {
	if p == nil || p.inner == nil {
		return []byte{}, nil
	}
	return p.inner.Screenshot(path, fullPage)
}

func (p *FirefoxBase) PDF(path string, options map[string]any) ([]byte, error) {
	if p == nil || p.inner == nil {
		return []byte{}, nil
	}
	return p.inner.PDF(path, options)
}

func (p *FirefoxBase) SavePage(path string) error {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.SavePage(path)
}

func (p *FirefoxBase) PromptOpen() bool {
	if p == nil || p.inner == nil {
		return false
	}
	return p.inner.PromptOpen()
}

func (p *FirefoxBase) PromptInfo() map[string]any {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.PromptInfo()
}

func (p *FirefoxBase) LastPromptOpened() map[string]any {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.LastPromptOpened()
}

func (p *FirefoxBase) LastPromptClosed() map[string]any {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.LastPromptClosed()
}

func (p *FirefoxBase) WaitPrompt(timeout time.Duration) (map[string]any, error) {
	if p == nil || p.inner == nil {
		return nil, nil
	}
	return p.inner.WaitPrompt(timeout)
}

func (p *FirefoxBase) HandlePrompt(accept bool, text *string, timeout time.Duration) error {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.HandlePrompt(accept, text, timeout)
}

func (p *FirefoxBase) WaitReadyState(target string, timeout time.Duration) error {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.WaitReadyState(target, timeout)
}

func (p *FirefoxBase) WaitLoadComplete(timeout time.Duration) error {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.WaitLoadComplete(timeout)
}

func (p *FirefoxBase) WaitURLContains(fragment string, timeout time.Duration) error {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.WaitURLContains(fragment, timeout)
}

func (p *FirefoxBase) WaitTitleContains(fragment string, timeout time.Duration) error {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.WaitTitleContains(fragment, timeout)
}

// GetFrame 获取当前上下文下的单个直接子 frame。
//
// locatorOrIndexOrContext 支持：
//   - nil：返回第一个 child frame
//   - 整数：按 0-based 子 frame 序号获取
//   - string：先按 context id 匹配，未命中时回退为 locator
//   - 其他 locator 兼容类型：按 iframe 元素定位
func (p *FirefoxBase) GetFrame(locatorOrIndexOrContext any) (*FirefoxFrame, error) {
	if p == nil || p.inner == nil {
		return nil, nil
	}
	frame, err := p.inner.GetFrame(locatorOrIndexOrContext)
	if err != nil || frame == nil {
		return nil, err
	}
	return newFirefoxFrameFromInner(frame, p.page, p), nil
}

// GetFrames 获取当前上下文的全部直接子 frame。
func (p *FirefoxBase) GetFrames() ([]*FirefoxFrame, error) {
	if p == nil || p.inner == nil {
		return []*FirefoxFrame{}, nil
	}
	frames, err := p.inner.GetFrames()
	if err != nil {
		return nil, err
	}
	result := make([]*FirefoxFrame, 0, len(frames))
	for _, frame := range frames {
		result = append(result, newFirefoxFrameFromInner(frame, p.page, p))
	}
	return result, nil
}
