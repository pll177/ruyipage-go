package base

import "fmt"

const (
	defaultBasePageType    = "BasePage"
	defaultBaseElementType = "BaseElement"
	maxPreviewTextLength   = 30
)

// BasePage 提供页面对象共享的字符串表示能力。
type BasePage struct {
	typeName  string
	urlGetter func() string
}

// NewBasePage 创建页面基类。
func NewBasePage(typeName string, urlGetter func() string) BasePage {
	page := BasePage{}
	page.SetTypeName(typeName)
	page.SetURLGetter(urlGetter)
	return page
}

// SetTypeName 设置字符串表示时使用的类型名。
func (b *BasePage) SetTypeName(typeName string) {
	if b == nil {
		return
	}
	if typeName == "" {
		typeName = defaultBasePageType
	}
	b.typeName = typeName
}

// SetURLGetter 设置 URL 读取函数。
func (b *BasePage) SetURLGetter(urlGetter func() string) {
	if b == nil {
		return
	}
	b.urlGetter = urlGetter
}

// TypeName 返回页面类型名。
func (b *BasePage) TypeName() string {
	if b == nil || b.typeName == "" {
		return defaultBasePageType
	}
	return b.typeName
}

// String 返回与 Python 版对齐的基础字符串表示。
func (b *BasePage) String() string {
	if b == nil {
		return fmt.Sprintf("<%s >", defaultBasePageType)
	}

	url := ""
	if b.urlGetter != nil {
		url = b.urlGetter()
	}
	return fmt.Sprintf("<%s %s>", b.TypeName(), url)
}

// BaseElement 提供元素对象共享的字符串表示能力。
type BaseElement struct {
	typeName   string
	tagGetter  func() string
	textGetter func() string
}

// NewBaseElement 创建元素基类。
func NewBaseElement(typeName string, tagGetter func() string, textGetter func() string) BaseElement {
	element := BaseElement{}
	element.SetTypeName(typeName)
	element.SetTagGetter(tagGetter)
	element.SetTextGetter(textGetter)
	return element
}

// SetTypeName 设置字符串表示时使用的类型名。
func (b *BaseElement) SetTypeName(typeName string) {
	if b == nil {
		return
	}
	if typeName == "" {
		typeName = defaultBaseElementType
	}
	b.typeName = typeName
}

// SetTagGetter 设置标签读取函数。
func (b *BaseElement) SetTagGetter(tagGetter func() string) {
	if b == nil {
		return
	}
	b.tagGetter = tagGetter
}

// SetTextGetter 设置文本读取函数。
func (b *BaseElement) SetTextGetter(textGetter func() string) {
	if b == nil {
		return
	}
	b.textGetter = textGetter
}

// TypeName 返回元素类型名。
func (b *BaseElement) TypeName() string {
	if b == nil || b.typeName == "" {
		return defaultBaseElementType
	}
	return b.typeName
}

// String 返回与 Python 版对齐的基础字符串表示。
func (b *BaseElement) String() string {
	if b == nil {
		return fmt.Sprintf("<%s ? \"\">", defaultBaseElementType)
	}

	tag := "?"
	if b.tagGetter != nil {
		if value := b.tagGetter(); value != "" {
			tag = value
		}
	}

	text := ""
	if b.textGetter != nil {
		text = b.textGetter()
	}
	if len(text) > maxPreviewTextLength {
		text = text[:maxPreviewTextLength] + "..."
	}

	return fmt.Sprintf("<%s %s %q>", b.TypeName(), tag, text)
}
