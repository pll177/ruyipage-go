package ruyipage

import (
	"time"

	internalelements "ruyipage-go/internal/elements"
)

// NoneElement 是空元素兜底公开包装对象。
type NoneElement struct {
	inner *internalelements.NoneElement
}

func newNoneElementFromInner(inner *internalelements.NoneElement) *NoneElement {
	return &NoneElement{inner: inner}
}

func (n *NoneElement) String() string {
	return "<NoneElement>"
}

func (n *NoneElement) IsNone() bool {
	return true
}

func (n *NoneElement) Valid() bool {
	return false
}

func (n *NoneElement) Tag() string {
	return ""
}

func (n *NoneElement) Text() string {
	return ""
}

func (n *NoneElement) HTML() string {
	return ""
}

func (n *NoneElement) OuterHTML() string {
	return ""
}

func (n *NoneElement) InnerHTML() string {
	return ""
}

func (n *NoneElement) Attrs() map[string]string {
	return map[string]string{}
}

func (n *NoneElement) Attr(name string) string {
	return ""
}

func (n *NoneElement) Link() string {
	return ""
}

func (n *NoneElement) Src() string {
	return ""
}

func (n *NoneElement) Value() string {
	return ""
}

func (n *NoneElement) Property(name string) any {
	return nil
}

func (n *NoneElement) Style(name string, pseudo string) string {
	return ""
}

func (n *NoneElement) ClickSelf(byJS bool, timeout time.Duration) error {
	return nil
}

func (n *NoneElement) Input(text any, clear bool, byJS bool) error {
	return nil
}

func (n *NoneElement) Clear() error {
	return nil
}

func (n *NoneElement) Hover() error {
	return nil
}

func (n *NoneElement) DragTo(target any, duration time.Duration) error {
	return nil
}

func (n *NoneElement) Focus() error {
	return nil
}

func (n *NoneElement) Screenshot(path string) ([]byte, error) {
	return []byte{}, nil
}

func (n *NoneElement) RunJS(script string, args ...any) (any, error) {
	return nil, nil
}

func (n *NoneElement) Parent(locator any, index int) StaticNode {
	return &NoneElement{}
}

func (n *NoneElement) Child(locator any, index int) StaticNode {
	return &NoneElement{}
}

func (n *NoneElement) Children(locator any) []*StaticElement {
	return []*StaticElement{}
}

func (n *NoneElement) Next(locator any, index int) StaticNode {
	return &NoneElement{}
}

func (n *NoneElement) Prev(locator any, index int) StaticNode {
	return &NoneElement{}
}

func (n *NoneElement) Ele(locator any) StaticNode {
	return &NoneElement{}
}

func (n *NoneElement) Eles(locator any) []*StaticElement {
	return []*StaticElement{}
}
