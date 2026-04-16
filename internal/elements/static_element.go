package elements

import (
	"strings"

	"github.com/andybalholm/cascadia"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"

	"github.com/pll177/ruyipage-go/internal/base"
	"github.com/pll177/ruyipage-go/internal/support"
)

// StaticNode 表示静态 HTML 解析后的节点统一读取接口。
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

// StaticElement 表示从 HTML 字符串解析得到的静态节点。
type StaticElement struct {
	base.BaseElement

	document *html.Node
	node     *html.Node
}

// MakeStaticElement 从 HTML 字符串构造静态节点；未命中时返回 NoneElement。
func MakeStaticElement(htmlText string, locator any) StaticNode {
	document, root, err := parseStaticHTML(htmlText)
	if err != nil || root == nil {
		return NewNoneElement("s_ele", map[string]any{"locator": locator})
	}
	if locator == nil {
		return NewStaticElement(document, root)
	}

	nodes, err := queryStaticNodes(root, locator, true)
	if err != nil || len(nodes) == 0 {
		return NewNoneElement("s_ele", map[string]any{"locator": locator})
	}
	return NewStaticElement(document, nodes[0])
}

// MakeStaticElements 从 HTML 字符串构造多个静态节点。
func MakeStaticElements(htmlText string, locator any) []*StaticElement {
	document, root, err := parseStaticHTML(htmlText)
	if err != nil || root == nil {
		return []*StaticElement{}
	}
	if locator == nil {
		return []*StaticElement{NewStaticElement(document, root)}
	}

	nodes, err := queryStaticNodes(root, locator, true)
	if err != nil {
		return []*StaticElement{}
	}
	return wrapStaticElements(document, nodes)
}

// NewStaticElement 根据 document 与 node 创建静态节点对象。
func NewStaticElement(document *html.Node, node *html.Node) *StaticElement {
	if node == nil {
		return nil
	}
	if document == nil {
		document = topHTMLNode(node)
	}

	element := &StaticElement{
		document: document,
		node:     node,
	}
	element.BaseElement = base.NewBaseElement(
		"StaticElement",
		func() string { return element.Tag() },
		func() string { return element.Text() },
	)
	return element
}

// IsNone 返回当前对象是否为空节点。
func (e *StaticElement) IsNone() bool {
	return false
}

// Valid 返回当前对象是否有效。
func (e *StaticElement) Valid() bool {
	return e != nil && e.node != nil
}

// Tag 返回标签名；非元素节点返回空字符串。
func (e *StaticElement) Tag() string {
	if e == nil || e.node == nil || e.node.Type != html.ElementNode {
		return ""
	}
	return strings.ToLower(e.node.Data)
}

// Text 返回节点及其后代文本。
func (e *StaticElement) Text() string {
	if e == nil || e.node == nil {
		return ""
	}
	switch e.node.Type {
	case html.TextNode:
		return e.node.Data
	case html.CommentNode:
		return ""
	default:
		return htmlquery.InnerText(e.node)
	}
}

// HTML 返回 outerHTML。
func (e *StaticElement) HTML() string {
	if e == nil || e.node == nil {
		return ""
	}
	return htmlquery.OutputHTML(e.node, true)
}

// OuterHTML 返回 outerHTML 别名。
func (e *StaticElement) OuterHTML() string {
	return e.HTML()
}

// InnerHTML 返回内部 HTML。
func (e *StaticElement) InnerHTML() string {
	if e == nil || e.node == nil {
		return ""
	}
	switch e.node.Type {
	case html.ElementNode, html.DocumentNode:
		return htmlquery.OutputHTML(e.node, false)
	case html.TextNode:
		return e.node.Data
	default:
		return ""
	}
}

// Attrs 返回属性副本。
func (e *StaticElement) Attrs() map[string]string {
	if e == nil || e.node == nil || e.node.Type != html.ElementNode {
		return map[string]string{}
	}
	result := make(map[string]string, len(e.node.Attr))
	for _, attr := range e.node.Attr {
		result[attr.Key] = attr.Val
	}
	return result
}

// Attr 返回单个属性值。
func (e *StaticElement) Attr(name string) string {
	if e == nil || e.node == nil || e.node.Type != html.ElementNode {
		return ""
	}
	return htmlquery.SelectAttr(e.node, name)
}

// Link 返回 href 属性值。
func (e *StaticElement) Link() string {
	return e.Attr("href")
}

// Src 返回 src 属性值。
func (e *StaticElement) Src() string {
	return e.Attr("src")
}

// Value 返回 value 属性值。
func (e *StaticElement) Value() string {
	return e.Attr("value")
}

// Parent 返回匹配的父节点。
func (e *StaticElement) Parent(locator any, index int) StaticNode {
	if !e.Valid() {
		return NewNoneElement("parent", map[string]any{"locator": locator, "index": index})
	}
	index = normalizeStaticIndex(index)
	current := e.node.Parent
	for current != nil {
		if current.Type == html.ElementNode && nodeMatchesLocator(current, locator) {
			index--
			if index == 0 {
				return NewStaticElement(e.document, current)
			}
		}
		current = current.Parent
	}
	return NewNoneElement("parent", map[string]any{"locator": locator, "index": index})
}

// Child 返回匹配的直接子节点。
func (e *StaticElement) Child(locator any, index int) StaticNode {
	children := e.Children(locator)
	index = normalizeStaticIndex(index)
	if index > len(children) {
		return NewNoneElement("child", map[string]any{"locator": locator, "index": index})
	}
	return children[index-1]
}

// Children 返回匹配的全部直接子节点。
func (e *StaticElement) Children(locator any) []*StaticElement {
	if !e.Valid() {
		return []*StaticElement{}
	}

	children := make([]*StaticElement, 0)
	for child := e.node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode {
			continue
		}
		if !nodeMatchesLocator(child, locator) {
			continue
		}
		children = append(children, NewStaticElement(e.document, child))
	}
	return children
}

// Next 返回匹配的后续兄弟节点。
func (e *StaticElement) Next(locator any, index int) StaticNode {
	if !e.Valid() {
		return NewNoneElement("next", map[string]any{"locator": locator, "index": index})
	}
	index = normalizeStaticIndex(index)
	for current := nextElementSibling(e.node); current != nil; current = nextElementSibling(current) {
		if !nodeMatchesLocator(current, locator) {
			continue
		}
		index--
		if index == 0 {
			return NewStaticElement(e.document, current)
		}
	}
	return NewNoneElement("next", map[string]any{"locator": locator, "index": index})
}

// Prev 返回匹配的前序兄弟节点。
func (e *StaticElement) Prev(locator any, index int) StaticNode {
	if !e.Valid() {
		return NewNoneElement("prev", map[string]any{"locator": locator, "index": index})
	}
	index = normalizeStaticIndex(index)
	for current := prevElementSibling(e.node); current != nil; current = prevElementSibling(current) {
		if !nodeMatchesLocator(current, locator) {
			continue
		}
		index--
		if index == 0 {
			return NewStaticElement(e.document, current)
		}
	}
	return NewNoneElement("prev", map[string]any{"locator": locator, "index": index})
}

// Ele 在当前节点后代内查找首个匹配节点。
func (e *StaticElement) Ele(locator any) StaticNode {
	if !e.Valid() {
		return NewNoneElement("ele", map[string]any{"locator": locator})
	}
	if locator == nil {
		return e
	}
	nodes, err := queryStaticNodes(e.node, locator, false)
	if err != nil || len(nodes) == 0 {
		return NewNoneElement("ele", map[string]any{"locator": locator})
	}
	return NewStaticElement(e.document, nodes[0])
}

// Eles 在当前节点后代内查找全部匹配节点。
func (e *StaticElement) Eles(locator any) []*StaticElement {
	if !e.Valid() {
		return []*StaticElement{}
	}
	if locator == nil {
		return []*StaticElement{e}
	}
	nodes, err := queryStaticNodes(e.node, locator, false)
	if err != nil {
		return []*StaticElement{}
	}
	return wrapStaticElements(e.document, nodes)
}

func wrapStaticElements(document *html.Node, nodes []*html.Node) []*StaticElement {
	if len(nodes) == 0 {
		return []*StaticElement{}
	}
	result := make([]*StaticElement, 0, len(nodes))
	for _, node := range nodes {
		if element := NewStaticElement(document, node); element != nil {
			result = append(result, element)
		}
	}
	return result
}

func parseStaticHTML(htmlText string) (*html.Node, *html.Node, error) {
	document, err := htmlquery.Parse(strings.NewReader(htmlText))
	if err != nil {
		return nil, nil, err
	}
	root := htmlquery.FindOne(document, "/html")
	if root == nil {
		root = topHTMLNode(document)
	}
	return document, root, nil
}

func queryStaticNodes(root *html.Node, locator any, includeSelf bool) ([]*html.Node, error) {
	if root == nil {
		return []*html.Node{}, nil
	}
	parsed, err := support.ParseLocator(locator)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(stringify(parsed["type"])) {
	case "css":
		selector, err := cascadia.ParseGroup(stringify(parsed["value"]))
		if err != nil {
			return nil, err
		}
		nodes := cascadia.QueryAll(root, selector)
		if includeSelf && selector.Match(root) {
			nodes = append([]*html.Node{root}, nodes...)
		}
		return uniqueStaticNodes(nodes, includeSelf, root), nil

	case "xpath":
		nodes, err := htmlquery.QueryAll(root, stringify(parsed["value"]))
		if err != nil {
			return nil, err
		}
		return uniqueStaticNodes(nodes, includeSelf, root), nil

	case "innertext":
		fullMatch := strings.EqualFold(stringify(parsed["matchType"]), "full")
		return queryStaticText(root, stringify(parsed["value"]), includeSelf, fullMatch), nil

	default:
		return nil, support.NewLocatorError("静态解析暂不支持该定位器类型", nil)
	}
}

func queryStaticText(root *html.Node, text string, includeSelf bool, fullMatch bool) []*html.Node {
	if root == nil {
		return []*html.Node{}
	}

	result := make([]*html.Node, 0)
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node != root || includeSelf {
			if node.Type == html.ElementNode && textMatches(directText(node), text, fullMatch) {
				result = append(result, node)
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return uniqueStaticNodes(result, true, nil)
}

func nodeMatchesLocator(node *html.Node, locator any) bool {
	if node == nil || locator == nil {
		return node != nil
	}
	parsed, err := support.ParseLocator(locator)
	if err != nil {
		return false
	}

	switch strings.ToLower(stringify(parsed["type"])) {
	case "css":
		selector, err := cascadia.ParseGroup(stringify(parsed["value"]))
		if err != nil {
			return false
		}
		return selector.Match(node)

	case "xpath":
		if matchesXPath(node, stringify(parsed["value"])) {
			return true
		}
		root := topHTMLNode(node)
		if root != nil && root != node {
			return nodeIncluded(queryXPath(root, stringify(parsed["value"])), node)
		}
		return false

	case "innertext":
		fullMatch := strings.EqualFold(stringify(parsed["matchType"]), "full")
		return node.Type == html.ElementNode && textMatches(directText(node), stringify(parsed["value"]), fullMatch)

	default:
		return false
	}
}

func matchesXPath(node *html.Node, expr string) bool {
	if node == nil || expr == "" {
		return false
	}
	return nodeIncluded(queryXPath(node, expr), node)
}

func queryXPath(root *html.Node, expr string) []*html.Node {
	if root == nil || expr == "" {
		return nil
	}
	nodes, err := htmlquery.QueryAll(root, expr)
	if err != nil {
		return nil
	}
	return nodes
}

func nodeIncluded(nodes []*html.Node, target *html.Node) bool {
	for _, node := range nodes {
		if node == target {
			return true
		}
	}
	return false
}

func uniqueStaticNodes(nodes []*html.Node, includeSelf bool, self *html.Node) []*html.Node {
	if len(nodes) == 0 {
		return []*html.Node{}
	}
	result := make([]*html.Node, 0, len(nodes))
	seen := make(map[*html.Node]struct{}, len(nodes))
	for _, node := range nodes {
		if node == nil {
			continue
		}
		if !includeSelf && self != nil && node == self {
			continue
		}
		if _, exists := seen[node]; exists {
			continue
		}
		seen[node] = struct{}{}
		result = append(result, node)
	}
	return result
}

func textMatches(current string, expected string, fullMatch bool) bool {
	if fullMatch {
		return current == expected
	}
	return strings.Contains(current, expected)
}

func directText(node *html.Node) string {
	if node == nil {
		return ""
	}
	if node.Type == html.TextNode {
		return node.Data
	}
	var builder strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode {
			builder.WriteString(child.Data)
		}
	}
	return builder.String()
}

func topHTMLNode(node *html.Node) *html.Node {
	if node == nil {
		return nil
	}
	current := node
	for current.Parent != nil {
		current = current.Parent
	}
	if current.Type == html.DocumentNode {
		if root := htmlquery.FindOne(current, "/html"); root != nil {
			return root
		}
	}
	if current.Type == html.DocumentNode {
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode {
				return child
			}
		}
	}
	return current
}

func nextElementSibling(node *html.Node) *html.Node {
	if node == nil {
		return nil
	}
	for sibling := node.NextSibling; sibling != nil; sibling = sibling.NextSibling {
		if sibling.Type == html.ElementNode {
			return sibling
		}
	}
	return nil
}

func prevElementSibling(node *html.Node) *html.Node {
	if node == nil {
		return nil
	}
	for sibling := node.PrevSibling; sibling != nil; sibling = sibling.PrevSibling {
		if sibling.Type == html.ElementNode {
			return sibling
		}
	}
	return nil
}

func normalizeStaticIndex(index int) int {
	if index <= 0 {
		return 1
	}
	return index
}
