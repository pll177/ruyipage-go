package pages

import (
	"time"

	"github.com/pll177/ruyipage-go/internal/browser"
	"github.com/pll177/ruyipage-go/internal/support"
)

// FirefoxTab 是最小可用标签页对象，供 FirefoxPage 直接返回。
type FirefoxTab struct {
	*FirefoxBase

	page    *FirefoxPage
	firefox *browser.Firefox
}

// Base 返回共享页面基类。
func (t *FirefoxTab) Base() *FirefoxBase {
	if t == nil {
		return nil
	}
	return t.FirefoxBase
}

// Page 返回所属顶层页面。
func (t *FirefoxTab) Page() *FirefoxPage {
	if t == nil {
		return nil
	}
	return t.page
}

// Browser 返回所属浏览器实例。
func (t *FirefoxTab) Browser() *browser.Firefox {
	if t == nil {
		return nil
	}
	return t.firefox
}

// Activate 激活当前标签页。
func (t *FirefoxTab) Activate() error {
	if t == nil || t.firefox == nil || t.FirefoxBase == nil {
		return nil
	}
	if err := t.firefox.ActivateTab(t.ContextID()); err != nil {
		return err
	}
	if t.page != nil {
		return t.page.SetContextID(t.ContextID())
	}
	return nil
}

// Close 关闭当前标签页；others=true 时保留当前、关闭其他标签页。
func (t *FirefoxTab) Close(others bool) error {
	if t == nil || t.firefox == nil {
		return nil
	}

	currentPageContext := ""
	if t.page != nil {
		currentPageContext = t.page.ContextID()
	}

	err := t.firefox.CloseTabs([]string{t.ContextID()}, others)
	if t.page != nil {
		tabIDs := t.firefox.TabIDs()
		t.page.purgeClosedTabs(tabIDs)
		switch {
		case others:
			if containsFirefoxTabContext(tabIDs, t.ContextID()) {
				if switchErr := t.page.SetContextID(t.ContextID()); err == nil {
					err = switchErr
				}
			}
		case currentPageContext == t.ContextID() && len(tabIDs) > 0:
			if switchErr := t.page.SetContextID(tabIDs[len(tabIDs)-1]); err == nil {
				err = switchErr
			}
		}
	}
	return err
}

// Save 保存当前标签页为 HTML 或 PDF。
func (t *FirefoxTab) Save(path string, name string, asPDF bool) (string, error) {
	if t == nil || t.FirefoxBase == nil {
		return "", support.NewPageDisconnectedError("FirefoxTab 未初始化", nil)
	}
	return saveFirefoxPageArtifact(t.FirefoxBase, path, name, asPDF)
}

// WaitReadyState 是对嵌入基类的显式转发，便于后续补全 tab 专属 API 时保持签名稳定。
func (t *FirefoxTab) WaitReadyState(target string, timeout time.Duration) error {
	if t == nil || t.FirefoxBase == nil {
		return nil
	}
	return t.FirefoxBase.WaitReadyState(target, timeout)
}

func containsFirefoxTabContext(contextIDs []string, target string) bool {
	for _, contextID := range contextIDs {
		if contextID == target {
			return true
		}
	}
	return false
}
