package pages

// FirefoxFrame 表示 iframe/frame 的独立 browsing context。
type FirefoxFrame struct {
	*FirefoxBase

	parent *FirefoxBase
	page   *FirefoxPage
}

// NewFirefoxFrame 创建绑定到父页面/父 frame 的 frame 对象。
func NewFirefoxFrame(browser FirefoxBrowser, contextID string, parent *FirefoxBase, page *FirefoxPage) (*FirefoxFrame, error) {
	basePage, err := NewFirefoxBase(browser, contextID)
	if err != nil {
		return nil, err
	}
	basePage.BasePage.SetTypeName("FirefoxFrame")
	basePage.setPageOwner(page)

	return &FirefoxFrame{
		FirefoxBase: basePage,
		parent:      parent,
		page:        page,
	}, nil
}

// Base 返回共享页面基类。
func (f *FirefoxFrame) Base() *FirefoxBase {
	if f == nil {
		return nil
	}
	return f.FirefoxBase
}

// Parent 返回父页面或父 frame。
func (f *FirefoxFrame) Parent() *FirefoxBase {
	if f == nil {
		return nil
	}
	return f.parent
}

// Page 返回所属顶层页面。
func (f *FirefoxFrame) Page() *FirefoxPage {
	if f == nil {
		return nil
	}
	return f.page
}

// IsCrossOrigin 判断当前 frame 是否与父页面/父 frame 跨域。
func (f *FirefoxFrame) IsCrossOrigin() bool {
	if f == nil || f.FirefoxBase == nil || f.parent == nil {
		return true
	}

	parentOrigin, err := f.parent.RunJSExpr("location.origin")
	if err != nil {
		return true
	}
	myOrigin, err := f.RunJSExpr("location.origin")
	if err != nil {
		return true
	}
	return stringify(parentOrigin) != stringify(myOrigin)
}
