package pages

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"ruyipage-go/internal/browser"
	"ruyipage-go/internal/config"
	"ruyipage-go/internal/support"
)

// FirefoxPage 是顶层页面入口，承接浏览器对象和默认 tab 语义。
type FirefoxPage struct {
	*FirefoxBase

	mu      sync.RWMutex
	firefox *browser.Firefox
	tabs    map[string]*FirefoxTab
}

// NewFirefoxPage 创建顶层页面对象；若浏览器尚无 tab，会自动创建一个默认 tab。
func NewFirefoxPage(options *config.FirefoxOptions) (*FirefoxPage, error) {
	firefox, err := browser.NewFirefox(options)
	if err != nil {
		return nil, err
	}

	return NewFirefoxPageFromBrowser(firefox, "")
}

// NewFirefoxPageFromBrowser 基于已有浏览器对象构造顶层页面对象。
func NewFirefoxPageFromBrowser(firefox *browser.Firefox, contextID string) (*FirefoxPage, error) {
	if firefox == nil {
		return nil, support.NewPageDisconnectedError("FirefoxPage 未初始化", nil)
	}

	tabIDs := firefox.TabIDs()
	if contextID == "" && len(tabIDs) > 0 {
		contextID = tabIDs[0]
	}
	if contextID == "" {
		var err error
		contextID, err = firefox.NewTab("", false)
		if err != nil {
			return nil, err
		}
	}

	basePage, err := NewFirefoxBase(firefox, contextID)
	if err != nil {
		return nil, err
	}
	basePage.BasePage.SetTypeName("FirefoxPage")
	page := &FirefoxPage{
		FirefoxBase: basePage,
		firefox:     firefox,
		tabs:        make(map[string]*FirefoxTab),
	}
	basePage.setPageOwner(page)
	return page, nil
}

// Base 返回共享页面基类。
func (p *FirefoxPage) Base() *FirefoxBase {
	if p == nil {
		return nil
	}
	return p.FirefoxBase
}

// Browser 返回底层 Firefox 生命周期对象。
func (p *FirefoxPage) Browser() *browser.Firefox {
	if p == nil {
		return nil
	}

	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.firefox
}

// TabsCount 返回当前 tab 数量。
func (p *FirefoxPage) TabsCount() int {
	firefox := p.Browser()
	if firefox == nil {
		return 0
	}
	return firefox.TabsCount()
}

// TabIDs 返回当前 tab id 列表副本。
func (p *FirefoxPage) TabIDs() []string {
	firefox := p.Browser()
	if firefox == nil {
		return []string{}
	}

	tabIDs := firefox.TabIDs()
	p.purgeClosedTabs(tabIDs)
	return tabIDs
}

// LatestTab 返回最新 tab。
func (p *FirefoxPage) LatestTab() *FirefoxTab {
	firefox := p.Browser()
	if firefox == nil {
		return nil
	}

	contextID := firefox.LatestTabID()
	if contextID == "" {
		return nil
	}
	tab, err := p.getOrCreateTab(contextID)
	if err != nil {
		return nil
	}
	return tab
}

// NewTab 创建新 tab，并在传入 URL 时立即导航。
func (p *FirefoxPage) NewTab(url string, background bool) (*FirefoxTab, error) {
	firefox := p.Browser()
	if firefox == nil {
		return nil, support.NewPageDisconnectedError("FirefoxPage 未初始化", nil)
	}

	contextID, err := firefox.NewTab(url, background)
	if contextID == "" {
		return nil, err
	}

	tab, tabErr := p.getOrCreateTab(contextID)
	if err != nil {
		return tab, err
	}
	return tab, tabErr
}

// GetTab 获取单个 tab；支持 context id、1-based 序号、负数倒序和标题/URL 模糊匹配。
func (p *FirefoxPage) GetTab(idOrNum any, title string, url string) (*FirefoxTab, error) {
	tabIDs := p.TabIDs()
	if len(tabIDs) == 0 {
		return nil, nil
	}

	if index, ok := firefoxTabIndexFromAny(idOrNum); ok {
		if index > 0 {
			index--
		}
		if index >= 0 && index < len(tabIDs) {
			return p.getOrCreateTab(tabIDs[index])
		}
		if index < 0 && -len(tabIDs) <= index {
			return p.getOrCreateTab(tabIDs[len(tabIDs)+index])
		}
		return nil, nil
	}

	if contextID, ok := idOrNum.(string); ok {
		for _, existing := range tabIDs {
			if existing == contextID {
				return p.getOrCreateTab(contextID)
			}
		}
		return nil, nil
	}

	if title != "" || url != "" {
		for _, contextID := range tabIDs {
			tab, err := p.getOrCreateTab(contextID)
			if err != nil {
				return nil, err
			}
			if title != "" {
				currentTitle, err := tab.Title()
				if err != nil {
					return nil, err
				}
				if strings.Contains(currentTitle, title) {
					return tab, nil
				}
			}
			if url != "" {
				currentURL, err := tab.URL()
				if err != nil {
					return nil, err
				}
				if strings.Contains(currentURL, url) {
					return tab, nil
				}
			}
		}
		return nil, nil
	}

	return p.getOrCreateTab(tabIDs[0])
}

// GetTabs 获取匹配条件的所有 tab；title/url 同时传入时按 AND 过滤。
func (p *FirefoxPage) GetTabs(title string, url string) ([]*FirefoxTab, error) {
	tabIDs := p.TabIDs()
	result := make([]*FirefoxTab, 0, len(tabIDs))
	for _, contextID := range tabIDs {
		tab, err := p.getOrCreateTab(contextID)
		if err != nil {
			return nil, err
		}
		if title != "" {
			currentTitle, err := tab.Title()
			if err != nil {
				return nil, err
			}
			if !strings.Contains(currentTitle, title) {
				continue
			}
		}
		if url != "" {
			currentURL, err := tab.URL()
			if err != nil {
				return nil, err
			}
			if !strings.Contains(currentURL, url) {
				continue
			}
		}
		result = append(result, tab)
	}
	return result, nil
}

// Close 关闭当前 tab，并在仍有剩余 tab 时切换到最后一个。
func (p *FirefoxPage) Close() error {
	if p == nil || p.FirefoxBase == nil {
		return nil
	}

	firefox := p.Browser()
	if firefox == nil {
		return support.NewPageDisconnectedError("FirefoxPage 未初始化", nil)
	}

	currentContext := p.ContextID()
	err := firefox.CloseTabs([]string{currentContext}, false)
	tabIDs := firefox.TabIDs()
	p.purgeClosedTabs(tabIDs)
	if len(tabIDs) > 0 {
		switchErr := p.SetContextID(tabIDs[len(tabIDs)-1])
		if err != nil {
			return err
		}
		return switchErr
	}
	return err
}

// CloseOtherTabs 关闭除保留 tab 外的其他 tab；未指定时默认保留当前 tab。
func (p *FirefoxPage) CloseOtherTabs(keepContextIDs []string) error {
	firefox := p.Browser()
	if firefox == nil {
		return support.NewPageDisconnectedError("FirefoxPage 未初始化", nil)
	}

	targets := cloneFirefoxPageStrings(keepContextIDs)
	if len(targets) == 0 {
		if contextID := p.ContextID(); contextID != "" {
			targets = append(targets, contextID)
		}
	}

	err := firefox.CloseTabs(targets, true)
	p.purgeClosedTabs(firefox.TabIDs())
	return err
}

// Quit 关闭浏览器并清空 tab 缓存。
func (p *FirefoxPage) Quit(timeout time.Duration, force bool) error {
	firefox := p.Browser()
	if firefox == nil {
		return nil
	}

	err := firefox.Quit(timeout, force)
	p.mu.Lock()
	p.tabs = make(map[string]*FirefoxTab)
	p.mu.Unlock()
	return err
}

// Save 保存当前页面为 HTML 或 PDF。
func (p *FirefoxPage) Save(path string, name string, asPDF bool) (string, error) {
	if p == nil || p.FirefoxBase == nil {
		return "", support.NewPageDisconnectedError("FirefoxPage 未初始化", nil)
	}
	return saveFirefoxPageArtifact(p.FirefoxBase, path, name, asPDF)
}

func (p *FirefoxPage) getOrCreateTab(contextID string) (*FirefoxTab, error) {
	if contextID == "" {
		return nil, nil
	}
	firefox := p.Browser()

	if support.Settings.SingletonTabObj {
		p.mu.RLock()
		existing := p.tabs[contextID]
		p.mu.RUnlock()
		if existing != nil {
			existing.page = p
			existing.firefox = firefox
			return existing, nil
		}
	}

	basePage, err := NewFirefoxBase(firefox, contextID)
	if err != nil {
		return nil, err
	}
	basePage.BasePage.SetTypeName("FirefoxTab")
	basePage.setPageOwner(p)

	tab := &FirefoxTab{
		FirefoxBase: basePage,
		page:        p,
		firefox:     firefox,
	}

	if support.Settings.SingletonTabObj {
		p.mu.Lock()
		if existing := p.tabs[contextID]; existing != nil {
			existing.page = p
			existing.firefox = firefox
			p.mu.Unlock()
			return existing, nil
		}
		p.tabs[contextID] = tab
		p.mu.Unlock()
	}
	return tab, nil
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

type firefoxPageArtifactSaver interface {
	Title() (string, error)
	SavePage(path string) error
	PDF(path string, options map[string]any) ([]byte, error)
}

func saveFirefoxPageArtifact(target firefoxPageArtifactSaver, path string, name string, asPDF bool) (string, error) {
	if target == nil {
		return "", support.NewPageDisconnectedError("页面对象未初始化", nil)
	}

	if path == "" {
		path = "."
	}
	if name == "" {
		title, err := target.Title()
		if err != nil {
			return "", err
		}
		if title == "" {
			title = "page"
		}
		name = support.MakeValidFilename(title, 50)
		if name == "" {
			name = "page"
		}
	}

	extension := ".html"
	if asPDF {
		extension = ".pdf"
	}

	filePath := filepath.Join(path, name+extension)
	if asPDF {
		_, err := target.PDF(filePath, nil)
		return filePath, err
	}
	return filePath, target.SavePage(filePath)
}

func firefoxTabIndexFromAny(value any) (int, bool) {
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

func cloneFirefoxPageStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

// NormalizeFirefoxPageKeepIDs 将单个 tab / id 或切片归一化为 context id 列表。
func NormalizeFirefoxPageKeepIDs(value any) ([]string, error) {
	switch typed := value.(type) {
	case nil:
		return nil, nil
	case string:
		if typed == "" {
			return []string{}, nil
		}
		return []string{typed}, nil
	case []string:
		return cloneFirefoxPageStrings(typed), nil
	case interface{ ContextID() string }:
		contextID := typed.ContextID()
		if contextID == "" {
			return []string{}, nil
		}
		return []string{contextID}, nil
	}

	refValue := reflect.ValueOf(value)
	if !refValue.IsValid() {
		return nil, nil
	}
	if refValue.Kind() != reflect.Slice && refValue.Kind() != reflect.Array {
		return nil, support.NewRuyiPageError(
			fmt.Sprintf("保留标签页参数必须是 context id、tab 对象或它们的切片，当前为 %T", value),
			nil,
		)
	}

	result := make([]string, 0, refValue.Len())
	for index := 0; index < refValue.Len(); index++ {
		part, err := NormalizeFirefoxPageKeepIDs(refValue.Index(index).Interface())
		if err != nil {
			return nil, err
		}
		result = append(result, part...)
	}
	return result, nil
}
