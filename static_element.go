package ruyipage

import internalelements "github.com/pll177/ruyipage-go/internal/elements"

// StaticNode 表示静态 HTML 解析后的统一节点接口。
type StaticNode interface {
	String() string
	IsNone() bool
	Valid() bool
	Tag() string
	Text() string
	HTML() string
	OuterHTML() string
	InnerHTML() string
	Attrs() map[string]string
	Attr(name string) string
	Link() string
	Src() string
	Value() string
	Parent(locator any, index int) StaticNode
	Child(locator any, index int) StaticNode
	Children(locator any) []*StaticElement
	Next(locator any, index int) StaticNode
	Prev(locator any, index int) StaticNode
	Ele(locator any) StaticNode
	Eles(locator any) []*StaticElement
}

// StaticElement 是静态 HTML 节点公开包装对象。
type StaticElement struct {
	inner *internalelements.StaticElement
}

func newStaticElementFromInner(inner *internalelements.StaticElement) *StaticElement {
	if inner == nil {
		return nil
	}
	return &StaticElement{inner: inner}
}

func wrapStaticNode(inner internalelements.StaticNode) StaticNode {
	switch typed := inner.(type) {
	case *internalelements.StaticElement:
		return newStaticElementFromInner(typed)
	case *internalelements.NoneElement:
		return newNoneElementFromInner(typed)
	case nil:
		return &NoneElement{}
	default:
		return &NoneElement{}
	}
}

func wrapStaticElements(inner []*internalelements.StaticElement) []*StaticElement {
	if len(inner) == 0 {
		return []*StaticElement{}
	}
	result := make([]*StaticElement, 0, len(inner))
	for _, item := range inner {
		if wrapped := newStaticElementFromInner(item); wrapped != nil {
			result = append(result, wrapped)
		}
	}
	return result
}

func (e *StaticElement) String() string {
	if e == nil || e.inner == nil {
		return `<StaticElement ? "">`
	}
	return e.inner.String()
}

func (e *StaticElement) IsNone() bool {
	return false
}

func (e *StaticElement) Valid() bool {
	return e != nil && e.inner != nil
}

func (e *StaticElement) Tag() string {
	if e == nil || e.inner == nil {
		return ""
	}
	return e.inner.Tag()
}

func (e *StaticElement) Text() string {
	if e == nil || e.inner == nil {
		return ""
	}
	return e.inner.Text()
}

func (e *StaticElement) HTML() string {
	if e == nil || e.inner == nil {
		return ""
	}
	return e.inner.HTML()
}

func (e *StaticElement) OuterHTML() string {
	if e == nil || e.inner == nil {
		return ""
	}
	return e.inner.OuterHTML()
}

func (e *StaticElement) InnerHTML() string {
	if e == nil || e.inner == nil {
		return ""
	}
	return e.inner.InnerHTML()
}

func (e *StaticElement) Attrs() map[string]string {
	if e == nil || e.inner == nil {
		return map[string]string{}
	}
	return e.inner.Attrs()
}

func (e *StaticElement) Attr(name string) string {
	if e == nil || e.inner == nil {
		return ""
	}
	return e.inner.Attr(name)
}

func (e *StaticElement) Link() string {
	if e == nil || e.inner == nil {
		return ""
	}
	return e.inner.Link()
}

func (e *StaticElement) Src() string {
	if e == nil || e.inner == nil {
		return ""
	}
	return e.inner.Src()
}

func (e *StaticElement) Value() string {
	if e == nil || e.inner == nil {
		return ""
	}
	return e.inner.Value()
}

func (e *StaticElement) Parent(locator any, index int) StaticNode {
	if e == nil || e.inner == nil {
		return &NoneElement{}
	}
	return wrapStaticNode(e.inner.Parent(locator, index))
}

func (e *StaticElement) Child(locator any, index int) StaticNode {
	if e == nil || e.inner == nil {
		return &NoneElement{}
	}
	return wrapStaticNode(e.inner.Child(locator, index))
}

func (e *StaticElement) Children(locator any) []*StaticElement {
	if e == nil || e.inner == nil {
		return []*StaticElement{}
	}
	return wrapStaticElements(e.inner.Children(locator))
}

func (e *StaticElement) Next(locator any, index int) StaticNode {
	if e == nil || e.inner == nil {
		return &NoneElement{}
	}
	return wrapStaticNode(e.inner.Next(locator, index))
}

func (e *StaticElement) Prev(locator any, index int) StaticNode {
	if e == nil || e.inner == nil {
		return &NoneElement{}
	}
	return wrapStaticNode(e.inner.Prev(locator, index))
}

func (e *StaticElement) Ele(locator any) StaticNode {
	if e == nil || e.inner == nil {
		return &NoneElement{}
	}
	return wrapStaticNode(e.inner.Ele(locator))
}

func (e *StaticElement) Eles(locator any) []*StaticElement {
	if e == nil || e.inner == nil {
		return []*StaticElement{}
	}
	return wrapStaticElements(e.inner.Eles(locator))
}
