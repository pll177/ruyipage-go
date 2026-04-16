package ruyipage

import internalpages "ruyipage-go/internal/pages"

// FirefoxFrame 是 iframe/frame 的公开包装对象。
type FirefoxFrame struct {
	*FirefoxBase

	inner  *internalpages.FirefoxFrame
	page   *FirefoxPage
	parent *FirefoxBase
}

func newFirefoxFrameFromInner(inner *internalpages.FirefoxFrame, page *FirefoxPage, parent *FirefoxBase) *FirefoxFrame {
	if inner == nil {
		return nil
	}

	base := newFirefoxBaseFromInner(inner.Base())
	frame := &FirefoxFrame{
		FirefoxBase: base,
		inner:       inner,
		page:        page,
		parent:      parent,
	}
	if base != nil {
		base.page = page
	}
	return frame
}

// Parent 返回父页面或父 frame 的基础包装层。
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

// Browser 返回所属浏览器对象。
func (f *FirefoxFrame) Browser() *Firefox {
	if f == nil {
		return nil
	}
	if f.page != nil {
		return f.page.Browser()
	}
	return nil
}

// IsCrossOrigin 判断当前 frame 是否与父上下文跨域。
func (f *FirefoxFrame) IsCrossOrigin() bool {
	if f == nil || f.inner == nil {
		return true
	}
	return f.inner.IsCrossOrigin()
}

func (f *FirefoxFrame) String() string {
	if f == nil || f.inner == nil {
		return "<FirefoxFrame >"
	}
	return f.inner.String()
}
