package ruyipage

import (
	"fmt"
	"sync"
	"time"

	"ruyipage-go/internal/config"
	internalpages "ruyipage-go/internal/pages"
	"ruyipage-go/internal/support"
)

var (
	firefoxPageRegistryMu sync.Mutex
	firefoxPageRegistry   = make(map[string]*FirefoxPage)
)

// FirefoxPage 是顶层页面对象，复用同地址浏览器与默认 tab 语义。
type FirefoxPage struct {
	*FirefoxBase

	inner      *internalpages.FirefoxPage
	browser    *Firefox
	addressKey string

	mu   sync.Mutex
	tabs map[string]*FirefoxTab
}

// NewFirefoxPage 按默认地址、FirefoxOptions 或指定地址创建顶层页面对象。
func NewFirefoxPage(addrOrOpts any) (*FirefoxPage, error) {
	raw, addressKey, err := resolveFirefoxPageInput(addrOrOpts)
	if err != nil {
		return nil, err
	}

	firefoxPageRegistryMu.Lock()
	defer firefoxPageRegistryMu.Unlock()

	if existing := firefoxPageRegistry[addressKey]; existing != nil {
		return existing, nil
	}

	inner, err := internalpages.NewFirefoxPage(raw)
	if err != nil {
		return nil, err
	}

	page := newFirefoxPageFromInner(inner, addressKey)
	firefoxPageRegistry[addressKey] = page
	return page, nil
}

func newFirefoxPageFromInner(inner *internalpages.FirefoxPage, addressKey string) *FirefoxPage {
	if inner == nil {
		return nil
	}

	base := newFirefoxBaseFromInner(inner.Base())
	page := &FirefoxPage{
		FirefoxBase: base,
		inner:       inner,
		browser:     newFirefoxFromInner(inner.Browser()),
		addressKey:  addressKey,
		tabs:        make(map[string]*FirefoxTab),
	}
	if base != nil {
		base.page = page
	}
	return page
}

// Browser 返回 Firefox 浏览器实例。
func (p *FirefoxPage) Browser() *Firefox {
	if p == nil {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.browser != nil {
		return p.browser
	}
	if p.inner == nil {
		return nil
	}
	p.browser = newFirefoxFromInner(p.inner.Browser())
	return p.browser
}

// TabsCount 返回当前标签页数量。
func (p *FirefoxPage) TabsCount() int {
	if p == nil || p.inner == nil {
		return 0
	}
	return p.inner.TabsCount()
}

// TabIDs 返回当前标签页 ID 列表副本。
func (p *FirefoxPage) TabIDs() []string {
	if p == nil || p.inner == nil {
		return []string{}
	}
	tabIDs := p.inner.TabIDs()
	p.purgeClosedTabs(tabIDs)
	return tabIDs
}

// LatestTab 返回最新标签页。
func (p *FirefoxPage) LatestTab() *FirefoxTab {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.wrapTab(p.inner.LatestTab())
}

// NewTab 新建标签页。
func (p *FirefoxPage) NewTab(url string, background bool) (*FirefoxTab, error) {
	if p == nil || p.inner == nil {
		return nil, nil
	}
	tab, err := p.inner.NewTab(url, background)
	if err != nil && tab == nil {
		return nil, err
	}
	return p.wrapTab(tab), err
}

// GetTab 获取单个标签页。
func (p *FirefoxPage) GetTab(idOrNum any, title string, url string) (*FirefoxTab, error) {
	if p == nil || p.inner == nil {
		return nil, nil
	}
	tab, err := p.inner.GetTab(idOrNum, title, url)
	if err != nil || tab == nil {
		return nil, err
	}
	return p.wrapTab(tab), nil
}

// GetTabs 获取全部匹配标签页。
func (p *FirefoxPage) GetTabs(title string, url string) ([]*FirefoxTab, error) {
	if p == nil || p.inner == nil {
		return []*FirefoxTab{}, nil
	}
	tabs, err := p.inner.GetTabs(title, url)
	if err != nil {
		return nil, err
	}
	result := make([]*FirefoxTab, 0, len(tabs))
	for _, tab := range tabs {
		result = append(result, p.wrapTab(tab))
	}
	return result, nil
}

// Close 关闭当前标签页。
func (p *FirefoxPage) Close() error {
	if p == nil || p.inner == nil {
		return nil
	}
	err := p.inner.Close()
	p.purgeClosedTabs(p.inner.TabIDs())
	return err
}

// CloseOtherTabs 关闭其他标签页；未指定时默认保留当前标签页。
func (p *FirefoxPage) CloseOtherTabs(tabOrIDs any) error {
	if p == nil || p.inner == nil {
		return nil
	}

	contextIDs, err := internalpages.NormalizeFirefoxPageKeepIDs(tabOrIDs)
	if err != nil {
		return err
	}
	err = p.inner.CloseOtherTabs(contextIDs)
	p.purgeClosedTabs(p.inner.TabIDs())
	return err
}

// Quit 关闭浏览器并从 page 单例缓存移除。
func (p *FirefoxPage) Quit(timeout time.Duration, force bool) error {
	if p == nil || p.inner == nil {
		return nil
	}

	err := p.inner.Quit(timeout, force)

	firefoxPageRegistryMu.Lock()
	delete(firefoxPageRegistry, p.addressKey)
	firefoxPageRegistryMu.Unlock()

	p.mu.Lock()
	p.tabs = make(map[string]*FirefoxTab)
	p.mu.Unlock()
	return err
}

// Save 保存当前页面。
func (p *FirefoxPage) Save(path string, name string, asPDF bool) (string, error) {
	if p == nil || p.inner == nil {
		return "", nil
	}
	return p.inner.Save(path, name, asPDF)
}

func (p *FirefoxPage) String() string {
	if p == nil || p.inner == nil {
		return "<FirefoxPage >"
	}
	return p.inner.String()
}

func (p *FirefoxPage) wrapTab(inner *internalpages.FirefoxTab) *FirefoxTab {
	if inner == nil {
		return nil
	}
	browser := p.Browser()

	if !support.Settings.SingletonTabObj {
		tab := newFirefoxTabFromInner(inner, p)
		tab.browser = browser
		return tab
	}

	contextID := inner.ContextID()
	p.mu.Lock()
	defer p.mu.Unlock()

	if existing := p.tabs[contextID]; existing != nil {
		existing.page = p
		existing.browser = browser
		return existing
	}

	tab := newFirefoxTabFromInner(inner, p)
	tab.browser = browser
	p.tabs[contextID] = tab
	return tab
}

func (p *FirefoxPage) purgeClosedTabs(tabIDs []string) {
	if p == nil || !support.Settings.SingletonTabObj {
		return
	}

	alive := make(map[string]struct{}, len(tabIDs))
	for _, contextID := range tabIDs {
		alive[contextID] = struct{}{}
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	for contextID := range p.tabs {
		if _, ok := alive[contextID]; ok {
			continue
		}
		delete(p.tabs, contextID)
	}
}

func resolveFirefoxPageInput(addrOrOpts any) (*config.FirefoxOptions, string, error) {
	var raw *config.FirefoxOptions

	switch typed := addrOrOpts.(type) {
	case nil:
		raw = config.NewFirefoxOptions()
	case string:
		raw = config.NewFirefoxOptions().WithAddress(typed)
	case *FirefoxOptions:
		if typed == nil {
			raw = config.NewFirefoxOptions()
		} else {
			raw = typed.raw().Clone()
		}
	case FirefoxOptions:
		raw = typed.raw().Clone()
	default:
		return nil, "", support.NewRuyiPageError(
			fmt.Sprintf("FirefoxPage 构造参数必须是 nil、string、FirefoxOptions 或 *FirefoxOptions，当前为 %T", addrOrOpts),
			nil,
		)
	}

	if err := raw.Validate(); err != nil {
		return nil, "", err
	}
	return raw, raw.Address(), nil
}
