package ruyipage

import internalpages "ruyipage-go/internal/pages"

// FirefoxTab 是页面返回的最小可用标签页对象。
type FirefoxTab struct {
	*FirefoxBase

	inner   *internalpages.FirefoxTab
	page    *FirefoxPage
	browser *Firefox
}

func newFirefoxTabFromInner(inner *internalpages.FirefoxTab, page *FirefoxPage) *FirefoxTab {
	if inner == nil {
		return nil
	}

	base := newFirefoxBaseFromInner(inner.Base())
	tab := &FirefoxTab{
		FirefoxBase: base,
		inner:       inner,
		page:        page,
		browser:     newFirefoxFromInner(inner.Browser()),
	}
	if base != nil {
		base.page = page
	}
	return tab
}

// Page 返回所属页面对象。
func (t *FirefoxTab) Page() *FirefoxPage {
	if t == nil {
		return nil
	}
	return t.page
}

// Browser 返回所属浏览器对象。
func (t *FirefoxTab) Browser() *Firefox {
	if t == nil {
		return nil
	}
	if t.browser != nil {
		return t.browser
	}
	if t.page != nil {
		return t.page.Browser()
	}
	return nil
}

// Activate 激活当前标签页并返回自身。
func (t *FirefoxTab) Activate() (*FirefoxTab, error) {
	if t == nil || t.inner == nil {
		return t, nil
	}
	return t, t.inner.Activate()
}

// Close 关闭当前标签页；others=true 时保留当前并关闭其他标签页。
func (t *FirefoxTab) Close(others bool) error {
	if t == nil || t.inner == nil {
		return nil
	}
	err := t.inner.Close(others)
	if t.page != nil {
		t.page.purgeClosedTabs(t.page.TabIDs())
	}
	return err
}

// Save 保存当前标签页。
func (t *FirefoxTab) Save(path string, name string, asPDF bool) (string, error) {
	if t == nil || t.inner == nil {
		return "", nil
	}
	return t.inner.Save(path, name, asPDF)
}

func (t *FirefoxTab) String() string {
	if t == nil || t.inner == nil {
		return "<FirefoxTab >"
	}
	return t.inner.String()
}
