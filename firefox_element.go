package ruyipage

import (
	"sync"
	"time"

	internalelements "ruyipage-go/internal/elements"
)

// FirefoxElement 是动态元素公开包装对象。
type FirefoxElement struct {
	inner *internalelements.FirefoxElement
	page  *FirefoxPage

	managersMu sync.Mutex
	clicker    *Clicker
	scroll     *ElementScroller
	rect       *ElementRect
	setter     *ElementSetter
	states     *ElementStates
	waiter     *ElementWaiter
	selects    *SelectElement
}

func newFirefoxElementFromInner(inner *internalelements.FirefoxElement, page *FirefoxPage) *FirefoxElement {
	if inner == nil {
		return nil
	}
	return &FirefoxElement{inner: inner, page: page}
}

func wrapFirefoxElements(inner []*internalelements.FirefoxElement, page *FirefoxPage) []*FirefoxElement {
	if len(inner) == 0 {
		return []*FirefoxElement{}
	}
	result := make([]*FirefoxElement, 0, len(inner))
	for _, item := range inner {
		if wrapped := newFirefoxElementFromInner(item, page); wrapped != nil {
			result = append(result, wrapped)
		}
	}
	return result
}

// Ele 查找单个动态元素。
func (p *FirefoxBase) Ele(locator any, index int, timeout time.Duration) (*FirefoxElement, error) {
	if p == nil || p.inner == nil {
		return nil, nil
	}
	element, err := p.inner.FindElement(locator, index, timeout, nil)
	if err != nil || element == nil {
		return nil, err
	}
	return newFirefoxElementFromInner(element, p.page), nil
}

// Eles 查找全部动态元素。
func (p *FirefoxBase) Eles(locator any, timeout time.Duration) ([]*FirefoxElement, error) {
	if p == nil || p.inner == nil {
		return []*FirefoxElement{}, nil
	}
	elements, err := p.inner.FindElements(locator, timeout, nil)
	if err != nil {
		return nil, err
	}
	return wrapFirefoxElements(elements, p.page), nil
}

// Page 返回所属顶层页面。
func (e *FirefoxElement) Page() *FirefoxPage {
	if e == nil {
		return nil
	}
	return e.page
}

// SharedID 返回元素 shared id。
func (e *FirefoxElement) SharedID() string {
	if e == nil || e.inner == nil {
		return ""
	}
	return e.inner.SharedID()
}

// Handle 返回元素 handle。
func (e *FirefoxElement) Handle() string {
	if e == nil || e.inner == nil {
		return ""
	}
	return e.inner.Handle()
}

func (e *FirefoxElement) String() string {
	if e == nil || e.inner == nil {
		return "<FirefoxElement >"
	}
	return e.inner.String()
}

func (e *FirefoxElement) Tag() (string, error) {
	return forwardElementString(e, (*internalelements.FirefoxElement).Tag)
}
func (e *FirefoxElement) Text() (string, error) {
	return forwardElementString(e, (*internalelements.FirefoxElement).Text)
}
func (e *FirefoxElement) InnerHTML() (string, error) {
	return forwardElementString(e, (*internalelements.FirefoxElement).InnerHTML)
}
func (e *FirefoxElement) HTML() (string, error) {
	return forwardElementString(e, (*internalelements.FirefoxElement).HTML)
}
func (e *FirefoxElement) OuterHTML() (string, error) {
	return forwardElementString(e, (*internalelements.FirefoxElement).OuterHTML)
}
func (e *FirefoxElement) Value() (string, error) {
	return forwardElementString(e, (*internalelements.FirefoxElement).Value)
}
func (e *FirefoxElement) Link() (string, error) {
	return forwardElementString(e, (*internalelements.FirefoxElement).Link)
}
func (e *FirefoxElement) Src() (string, error) {
	return forwardElementString(e, (*internalelements.FirefoxElement).Src)
}
func (e *FirefoxElement) Attr(name string) (string, error) {
	return forwardElementStringArg(e, name, (*internalelements.FirefoxElement).Attr)
}
func (e *FirefoxElement) Property(name string) (any, error) {
	return forwardElementAnyArg(e, name, (*internalelements.FirefoxElement).Property)
}
func (e *FirefoxElement) Style(name string, pseudo string) (string, error) {
	if e == nil || e.inner == nil {
		return "", nil
	}
	return e.inner.Style(name, pseudo)
}
func (e *FirefoxElement) Attrs() (map[string]string, error) {
	if e == nil || e.inner == nil {
		return map[string]string{}, nil
	}
	return e.inner.Attrs()
}
func (e *FirefoxElement) Pseudo() (map[string]string, error) {
	if e == nil || e.inner == nil {
		return map[string]string{}, nil
	}
	return e.inner.Pseudo()
}
func (e *FirefoxElement) IsDisplayed() (bool, error) {
	return forwardElementBool(e, (*internalelements.FirefoxElement).IsDisplayed)
}
func (e *FirefoxElement) IsEnabled() (bool, error) {
	return forwardElementBool(e, (*internalelements.FirefoxElement).IsEnabled)
}
func (e *FirefoxElement) IsChecked() (bool, error) {
	return forwardElementBool(e, (*internalelements.FirefoxElement).IsChecked)
}
func (e *FirefoxElement) Size() (map[string]int, error) {
	if e == nil || e.inner == nil {
		return map[string]int{}, nil
	}
	return e.inner.Size()
}
func (e *FirefoxElement) Location() (map[string]int, error) {
	if e == nil || e.inner == nil {
		return map[string]int{}, nil
	}
	return e.inner.Location()
}

func (e *FirefoxElement) ViewportMidpoint() (map[string]int, error) {
	if e == nil || e.inner == nil {
		return map[string]int{}, nil
	}
	return e.inner.ViewportMidpoint()
}

func (e *FirefoxElement) ShadowRoot() (*FirefoxElement, error) {
	if e == nil || e.inner == nil {
		return nil, nil
	}
	root, err := e.inner.ShadowRoot()
	if err != nil || root == nil {
		return nil, err
	}
	return newFirefoxElementFromInner(root, e.page), nil
}
func (e *FirefoxElement) ClosedShadowRoot() (*FirefoxElement, error) {
	if e == nil || e.inner == nil {
		return nil, nil
	}
	root, err := e.inner.ClosedShadowRoot()
	if err != nil || root == nil {
		return nil, err
	}
	return newFirefoxElementFromInner(root, e.page), nil
}
func (e *FirefoxElement) WithShadow(mode string) (*FirefoxElement, error) {
	if e == nil || e.inner == nil {
		return nil, nil
	}
	root, err := e.inner.WithShadow(mode)
	if err != nil || root == nil {
		return nil, err
	}
	return newFirefoxElementFromInner(root, e.page), nil
}

func (e *FirefoxElement) ClickSelf(byJS bool, timeout time.Duration) error {
	if e == nil || e.inner == nil {
		return nil
	}
	return e.inner.ClickSelf(byJS, timeout)
}
func (e *FirefoxElement) RightClick() error {
	if e == nil || e.inner == nil {
		return nil
	}
	return e.inner.RightClick()
}
func (e *FirefoxElement) DoubleClick() error {
	if e == nil || e.inner == nil {
		return nil
	}
	return e.inner.DoubleClick()
}
func (e *FirefoxElement) Input(text any, clear bool, byJS bool) error {
	if e == nil || e.inner == nil {
		return nil
	}
	return e.inner.Input(text, clear, byJS)
}
func (e *FirefoxElement) UploadFiles(files ...string) error {
	if e == nil || e.inner == nil {
		return nil
	}
	return e.inner.UploadFiles(files...)
}
func (e *FirefoxElement) Clear() error {
	if e == nil || e.inner == nil {
		return nil
	}
	return e.inner.Clear()
}
func (e *FirefoxElement) Hover() error {
	if e == nil || e.inner == nil {
		return nil
	}
	return e.inner.Hover()
}
func (e *FirefoxElement) DragTo(target any, duration time.Duration) error {
	if e == nil || e.inner == nil {
		return nil
	}
	if other, ok := target.(*FirefoxElement); ok {
		if other == nil {
			return nil
		}
		return e.inner.DragTo(other.inner, duration)
	}
	return e.inner.DragTo(target, duration)
}
func (e *FirefoxElement) Screenshot(path string) ([]byte, error) {
	if e == nil || e.inner == nil {
		return []byte{}, nil
	}
	return e.inner.Screenshot(path)
}
func (e *FirefoxElement) ScrollToSee(center bool) error {
	if e == nil || e.inner == nil {
		return nil
	}
	return e.inner.ScrollToSee(center)
}
func (e *FirefoxElement) Focus() error {
	if e == nil || e.inner == nil {
		return nil
	}
	return e.inner.Focus()
}
func (e *FirefoxElement) RunJS(script string, args ...any) (any, error) {
	if e == nil || e.inner == nil {
		return nil, nil
	}
	return e.inner.RunJS(script, args...)
}

func (e *FirefoxElement) Parent(locator any, index int) (*FirefoxElement, error) {
	if e == nil || e.inner == nil {
		return nil, nil
	}
	parent, err := e.inner.Parent(locator, index)
	if err != nil || parent == nil {
		return nil, err
	}
	return newFirefoxElementFromInner(parent, e.page), nil
}
func (e *FirefoxElement) Child(locator any, index int, timeout time.Duration) (*FirefoxElement, error) {
	if e == nil || e.inner == nil {
		return nil, nil
	}
	child, err := e.inner.Child(locator, index, timeout)
	if err != nil || child == nil {
		return nil, err
	}
	return newFirefoxElementFromInner(child, e.page), nil
}
func (e *FirefoxElement) Children(locator any, timeout time.Duration) ([]*FirefoxElement, error) {
	if e == nil || e.inner == nil {
		return []*FirefoxElement{}, nil
	}
	children, err := e.inner.Children(locator, timeout)
	if err != nil {
		return nil, err
	}
	return wrapFirefoxElements(children, e.page), nil
}
func (e *FirefoxElement) Next(locator any, index int) (*FirefoxElement, error) {
	if e == nil || e.inner == nil {
		return nil, nil
	}
	next, err := e.inner.Next(locator, index)
	if err != nil || next == nil {
		return nil, err
	}
	return newFirefoxElementFromInner(next, e.page), nil
}
func (e *FirefoxElement) Prev(locator any, index int) (*FirefoxElement, error) {
	if e == nil || e.inner == nil {
		return nil, nil
	}
	prev, err := e.inner.Prev(locator, index)
	if err != nil || prev == nil {
		return nil, err
	}
	return newFirefoxElementFromInner(prev, e.page), nil
}
func (e *FirefoxElement) Ele(locator any, index int, timeout time.Duration) (*FirefoxElement, error) {
	if e == nil || e.inner == nil {
		return nil, nil
	}
	child, err := e.inner.Ele(locator, index, timeout)
	if err != nil || child == nil {
		return nil, err
	}
	return newFirefoxElementFromInner(child, e.page), nil
}
func (e *FirefoxElement) Eles(locator any, timeout time.Duration) ([]*FirefoxElement, error) {
	if e == nil || e.inner == nil {
		return []*FirefoxElement{}, nil
	}
	children, err := e.inner.Eles(locator, timeout)
	if err != nil {
		return nil, err
	}
	return wrapFirefoxElements(children, e.page), nil
}

func (e *FirefoxElement) SEle(locator any) (StaticNode, error) {
	if e == nil || e.inner == nil {
		return &NoneElement{}, nil
	}
	node, err := e.inner.SEle(locator)
	if err != nil {
		return nil, err
	}
	return wrapStaticNode(node), nil
}

func (e *FirefoxElement) SEles(locator any) ([]*StaticElement, error) {
	if e == nil || e.inner == nil {
		return []*StaticElement{}, nil
	}
	nodes, err := e.inner.SEles(locator)
	if err != nil {
		return nil, err
	}
	return wrapStaticElements(nodes), nil
}

func forwardElementString(element *FirefoxElement, fn func(*internalelements.FirefoxElement) (string, error)) (string, error) {
	if element == nil || element.inner == nil {
		return "", nil
	}
	return fn(element.inner)
}

func forwardElementStringArg(element *FirefoxElement, argument string, fn func(*internalelements.FirefoxElement, string) (string, error)) (string, error) {
	if element == nil || element.inner == nil {
		return "", nil
	}
	return fn(element.inner, argument)
}

func forwardElementAnyArg(element *FirefoxElement, argument string, fn func(*internalelements.FirefoxElement, string) (any, error)) (any, error) {
	if element == nil || element.inner == nil {
		return nil, nil
	}
	return fn(element.inner, argument)
}

func forwardElementBool(element *FirefoxElement, fn func(*internalelements.FirefoxElement) (bool, error)) (bool, error) {
	if element == nil || element.inner == nil {
		return false, nil
	}
	return fn(element.inner)
}
