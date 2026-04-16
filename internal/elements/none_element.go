package elements

import "time"

// NoneElement 表示静态解析或空对象兜底返回值。
type NoneElement struct {
	method string
	args   map[string]any
}

// NewNoneElement 创建空元素对象。
func NewNoneElement(method string, args map[string]any) *NoneElement {
	return &NoneElement{
		method: method,
		args:   cloneAnyMap(args),
	}
}

// String 返回空元素字符串表示。
func (n *NoneElement) String() string {
	return "<NoneElement>"
}

// IsNone 返回当前对象是否为空元素。
func (n *NoneElement) IsNone() bool {
	return true
}

// Valid 返回当前对象是否有效。
func (n *NoneElement) Valid() bool {
	return false
}

// Tag 返回空字符串。
func (n *NoneElement) Tag() string {
	return ""
}

// Text 返回空字符串。
func (n *NoneElement) Text() string {
	return ""
}

// HTML 返回空字符串。
func (n *NoneElement) HTML() string {
	return ""
}

// OuterHTML 返回空字符串。
func (n *NoneElement) OuterHTML() string {
	return ""
}

// InnerHTML 返回空字符串。
func (n *NoneElement) InnerHTML() string {
	return ""
}

// Attrs 返回空 map。
func (n *NoneElement) Attrs() map[string]string {
	return map[string]string{}
}

// Attr 返回空字符串。
func (n *NoneElement) Attr(name string) string {
	return ""
}

// Link 返回空字符串。
func (n *NoneElement) Link() string {
	return ""
}

// Src 返回空字符串。
func (n *NoneElement) Src() string {
	return ""
}

// Value 返回空字符串。
func (n *NoneElement) Value() string {
	return ""
}

// Property 返回 nil。
func (n *NoneElement) Property(name string) any {
	return nil
}

// Style 返回空字符串。
func (n *NoneElement) Style(name string, pseudo string) string {
	return ""
}

// ClickSelf 返回 nil，保持空对象链路安全。
func (n *NoneElement) ClickSelf(byJS bool, timeout time.Duration) error {
	return nil
}

// Input 返回 nil，保持空对象链路安全。
func (n *NoneElement) Input(text any, clear bool, byJS bool) error {
	return nil
}

// Clear 返回 nil，保持空对象链路安全。
func (n *NoneElement) Clear() error {
	return nil
}

// Hover 返回 nil，保持空对象链路安全。
func (n *NoneElement) Hover() error {
	return nil
}

// DragTo 返回 nil，保持空对象链路安全。
func (n *NoneElement) DragTo(target any, duration time.Duration) error {
	return nil
}

// Focus 返回 nil，保持空对象链路安全。
func (n *NoneElement) Focus() error {
	return nil
}

// Screenshot 返回空字节与 nil error。
func (n *NoneElement) Screenshot(path string) ([]byte, error) {
	return []byte{}, nil
}

// RunJS 返回 nil, nil。
func (n *NoneElement) RunJS(script string, args ...any) (any, error) {
	return nil, nil
}

// Parent 返回新的 NoneElement。
func (n *NoneElement) Parent(locator any, index int) StaticNode {
	return NewNoneElement("parent", map[string]any{"locator": locator, "index": index})
}

// Child 返回新的 NoneElement。
func (n *NoneElement) Child(locator any, index int) StaticNode {
	return NewNoneElement("child", map[string]any{"locator": locator, "index": index})
}

// Children 返回空切片。
func (n *NoneElement) Children(locator any) []*StaticElement {
	return []*StaticElement{}
}

// Next 返回新的 NoneElement。
func (n *NoneElement) Next(locator any, index int) StaticNode {
	return NewNoneElement("next", map[string]any{"locator": locator, "index": index})
}

// Prev 返回新的 NoneElement。
func (n *NoneElement) Prev(locator any, index int) StaticNode {
	return NewNoneElement("prev", map[string]any{"locator": locator, "index": index})
}

// Ele 返回新的 NoneElement。
func (n *NoneElement) Ele(locator any) StaticNode {
	return NewNoneElement("ele", map[string]any{"locator": locator})
}

// Eles 返回空切片。
func (n *NoneElement) Eles(locator any) []*StaticElement {
	return []*StaticElement{}
}
