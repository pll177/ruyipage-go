package elements

import (
	"encoding/base64"
	stderrors "errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pll177/ruyipage-go/internal/base"
	"github.com/pll177/ruyipage-go/internal/bidi"
	"github.com/pll177/ruyipage-go/internal/support"
	"github.com/pll177/ruyipage-go/internal/units"
)

const (
	defaultElementClickTimeout = 1500 * time.Millisecond
	elementWaitPollInterval    = 100 * time.Millisecond
)

// Owner 表示 FirefoxElement 依赖的页面 / tab / frame 最小能力集合。
type Owner interface {
	ContextID() string
	BrowserDriver() *base.BrowserBiDiDriver
	BaseTimeout() time.Duration
	ScriptTimeout() time.Duration
	ElementFindTimeout() time.Duration
	FindElement(locator any, index int, timeout time.Duration, startNodes []map[string]any) (*FirefoxElement, error)
	FindElements(locator any, timeout time.Duration, startNodes []map[string]any) ([]*FirefoxElement, error)
}

// FirefoxElement 表示动态 DOM 节点。
type FirefoxElement struct {
	base.BaseElement

	mu sync.RWMutex

	owner       Owner
	sharedID    string
	handle      string
	nodeInfo    map[string]any
	locatorInfo any

	clicker *units.Clicker
	scroll  *units.ElementScroller
	rect    *units.ElementRect
	setter  *units.ElementSetter
	states  *units.ElementStates
	waiter  *units.ElementWaiter
	selects *units.SelectElement
}

// NewFirefoxElement 创建动态元素对象。
func NewFirefoxElement(owner Owner, sharedID string, handle string, nodeInfo map[string]any, locatorInfo any) *FirefoxElement {
	if owner == nil || sharedID == "" {
		return nil
	}

	element := &FirefoxElement{
		owner:       owner,
		sharedID:    sharedID,
		handle:      handle,
		nodeInfo:    cloneAnyMap(nodeInfo),
		locatorInfo: cloneLocatorInfo(locatorInfo),
	}
	element.BaseElement = base.NewBaseElement(
		"FirefoxElement",
		func() string {
			tag, err := element.Tag()
			if err != nil {
				return ""
			}
			return tag
		},
		func() string {
			text, err := element.Text()
			if err != nil {
				return ""
			}
			return text
		},
	)
	return element
}

// FromNode 根据 BiDi RemoteValue 节点结果创建元素对象。
func FromNode(owner Owner, nodeData map[string]any, locatorInfo any) *FirefoxElement {
	if len(nodeData) == 0 {
		return nil
	}

	nodeType := strings.TrimSpace(strings.ToLower(readString(nodeData, "type")))
	switch nodeType {
	case "", "node":
		if sharedID := readString(nodeData, "sharedId"); sharedID != "" {
			return NewFirefoxElement(owner, sharedID, readString(nodeData, "handle"), readNestedMap(nodeData, "value"), locatorInfo)
		}
		value := readNestedMap(nodeData, "value")
		if sharedID := readString(value, "sharedId"); sharedID != "" {
			return NewFirefoxElement(owner, sharedID, readString(value, "handle"), value, locatorInfo)
		}
	}
	if sharedID := readString(nodeData, "sharedId"); sharedID != "" {
		return NewFirefoxElement(owner, sharedID, readString(nodeData, "handle"), readNestedMap(nodeData, "value"), locatorInfo)
	}
	return nil
}

// BiDiSharedID 返回 shared id。
func (e *FirefoxElement) BiDiSharedID() string {
	if e == nil {
		return ""
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.sharedID
}

// BiDiHandle 返回远端 handle。
func (e *FirefoxElement) BiDiHandle() string {
	if e == nil {
		return ""
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.handle
}

// BiDiSharedReference 返回 shared reference 结构。
func (e *FirefoxElement) BiDiSharedReference() map[string]any {
	if e == nil {
		return support.MakeSharedRef("", "")
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	return support.MakeSharedRef(e.sharedID, e.handle)
}

// Owner 返回元素所属 owner。
func (e *FirefoxElement) Owner() Owner {
	if e == nil {
		return nil
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.owner
}

// SharedID 返回当前元素的 shared id。
func (e *FirefoxElement) SharedID() string {
	return e.BiDiSharedID()
}

// Handle 返回当前元素的远端 handle。
func (e *FirefoxElement) Handle() string {
	return e.BiDiHandle()
}

// NodeInfo 返回节点信息快照。
func (e *FirefoxElement) NodeInfo() map[string]any {
	if e == nil {
		return nil
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	return cloneAnyMap(e.nodeInfo)
}

// Tag 返回标签名。
func (e *FirefoxElement) Tag() (string, error) {
	if e == nil {
		return "", nil
	}
	e.mu.RLock()
	localName := readString(e.nodeInfo, "localName")
	e.mu.RUnlock()
	if localName != "" {
		return strings.ToLower(localName), nil
	}
	value, err := e.callOnSelfParsed("(el) => el.tagName ? el.tagName.toLowerCase() : ''")
	if err != nil {
		return "", err
	}
	return stringify(value), nil
}

// Text 返回元素文本。
func (e *FirefoxElement) Text() (string, error) {
	value, err := e.callOnSelfParsed("(el) => el.textContent || ''")
	if err != nil {
		return "", err
	}
	return stringify(value), nil
}

// InnerHTML 返回 innerHTML。
func (e *FirefoxElement) InnerHTML() (string, error) {
	value, err := e.callOnSelfParsed("(el) => el.innerHTML || ''")
	if err != nil {
		return "", err
	}
	return stringify(value), nil
}

// HTML 返回 outerHTML。
func (e *FirefoxElement) HTML() (string, error) {
	value, err := e.callOnSelfParsed("(el) => el.outerHTML || ''")
	if err != nil {
		return "", err
	}
	return stringify(value), nil
}

// OuterHTML 返回 outerHTML 别名。
func (e *FirefoxElement) OuterHTML() (string, error) {
	return e.HTML()
}

// Value 返回元素 value 属性。
func (e *FirefoxElement) Value() (string, error) {
	value, err := e.callOnSelfParsed("(el) => el.value")
	if err != nil {
		return "", err
	}
	return stringify(value), nil
}

// Attrs 返回全部属性字典。
func (e *FirefoxElement) Attrs() (map[string]string, error) {
	if e == nil {
		return map[string]string{}, nil
	}

	e.mu.RLock()
	cached := cloneAnyMap(readNestedMap(e.nodeInfo, "attributes"))
	e.mu.RUnlock()
	if len(cached) > 0 {
		return anyMapToStringMap(cached), nil
	}

	value, err := e.callOnSelfParsed(`(el) => {
		const attrs = {};
		for (let index = 0; index < el.attributes.length; index += 1) {
			const attr = el.attributes[index];
			attrs[attr.name] = attr.value;
		}
		return attrs;
	}`)
	if err != nil {
		return nil, err
	}
	return anyMapToStringMap(asAnyMap(value)), nil
}

// Link 返回 href 的绝对地址。
func (e *FirefoxElement) Link() (string, error) {
	href, err := e.Attr("href")
	if err != nil || href == "" {
		return href, err
	}
	value, err := e.callOnSelfParsed("(el) => el.href || ''")
	if err != nil {
		return href, err
	}
	return stringify(value), nil
}

// Src 返回 src 的绝对地址。
func (e *FirefoxElement) Src() (string, error) {
	src, err := e.Attr("src")
	if err != nil || src == "" {
		return src, err
	}
	value, err := e.callOnSelfParsed("(el) => el.src || ''")
	if err != nil {
		return src, err
	}
	return stringify(value), nil
}

// Attr 返回属性值。
func (e *FirefoxElement) Attr(name string) (string, error) {
	value, err := e.callOnSelfParsed("(el, name) => el.getAttribute(name)", name)
	if err != nil {
		return "", err
	}
	return stringify(value), nil
}

// Property 返回 JS 属性值。
func (e *FirefoxElement) Property(name string) (any, error) {
	return e.callOnSelfParsed("(el, name) => el[name]", name)
}

// Style 返回计算样式值。
func (e *FirefoxElement) Style(name string, pseudo string) (string, error) {
	value, err := e.callOnSelfParsed(
		"(el, name, pseudo) => window.getComputedStyle(el, pseudo || null).getPropertyValue(name)",
		name,
		pseudo,
	)
	if err != nil {
		return "", err
	}
	return stringify(value), nil
}

// Pseudo 返回 before / after 伪元素 content。
func (e *FirefoxElement) Pseudo() (map[string]string, error) {
	before, err := e.Style("content", "::before")
	if err != nil {
		return nil, err
	}
	after, err := e.Style("content", "::after")
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"before": before,
		"after":  after,
	}, nil
}

// IsDisplayed 判断元素是否可见。
func (e *FirefoxElement) IsDisplayed() (bool, error) {
	value, err := e.callOnSelfParsed(`(el) => {
		const style = window.getComputedStyle(el);
		return style.display !== 'none'
			&& style.visibility !== 'hidden'
			&& style.opacity !== '0'
			&& el.offsetWidth > 0
			&& el.offsetHeight > 0;
	}`)
	if err != nil {
		return false, err
	}
	return toBool(value), nil
}

// IsEnabled 判断元素是否可用。
func (e *FirefoxElement) IsEnabled() (bool, error) {
	value, err := e.callOnSelfParsed("(el) => !el.disabled")
	if err != nil {
		return false, err
	}
	return toBool(value), nil
}

// IsChecked 判断元素是否已选中。
func (e *FirefoxElement) IsChecked() (bool, error) {
	value, err := e.callOnSelfParsed("(el) => !!el.checked")
	if err != nil {
		return false, err
	}
	return toBool(value), nil
}

// IsSelected 判断元素是否处于 selected 状态。
func (e *FirefoxElement) IsSelected() (bool, error) {
	value, err := e.callOnSelfParsed("(el) => !!el.selected")
	if err != nil {
		return false, err
	}
	return toBool(value), nil
}

// IsInViewport 判断元素是否在视口内。
func (e *FirefoxElement) IsInViewport() (bool, error) {
	value, err := e.callOnSelfParsed(`(el) => {
		const rect = el.getBoundingClientRect();
		return rect.bottom >= 0
			&& rect.right >= 0
			&& rect.top <= (window.innerHeight || document.documentElement.clientHeight)
			&& rect.left <= (window.innerWidth || document.documentElement.clientWidth);
	}`)
	if err != nil {
		return false, err
	}
	return toBool(value), nil
}

// HasRect 判断元素是否有非零矩形区域。
func (e *FirefoxElement) HasRect() (bool, error) {
	size, err := e.Size()
	if err != nil {
		return false, err
	}
	return size["width"] > 0 && size["height"] > 0, nil
}

// Size 返回元素宽高。
func (e *FirefoxElement) Size() (map[string]int, error) {
	value, err := e.callOnSelfParsed(`(el) => {
		const rect = el.getBoundingClientRect();
		return {width: Math.round(rect.width), height: Math.round(rect.height)};
	}`)
	if err != nil {
		return nil, err
	}
	return normalizeXYWH(value, "width", "height"), nil
}

// Location 返回元素文档坐标。
func (e *FirefoxElement) Location() (map[string]int, error) {
	value, err := e.callOnSelfParsed(`(el) => {
		const rect = el.getBoundingClientRect();
		return {
			x: Math.round(rect.left + window.scrollX),
			y: Math.round(rect.top + window.scrollY)
		};
	}`)
	if err != nil {
		return nil, err
	}
	return normalizeXYWH(value, "x", "y"), nil
}

// ViewportLocation 返回元素视口坐标。
func (e *FirefoxElement) ViewportLocation() (map[string]int, error) {
	value, err := e.callOnSelfParsed(`(el) => {
		const rect = el.getBoundingClientRect();
		return {x: Math.round(rect.left), y: Math.round(rect.top)};
	}`)
	if err != nil {
		return nil, err
	}
	return normalizeXYWH(value, "x", "y"), nil
}

// ViewportMidpoint 返回元素视口中心点。
func (e *FirefoxElement) ViewportMidpoint() (map[string]int, error) {
	return e.centerPoint(false)
}

// ShadowRoot 返回 open shadow root。
func (e *FirefoxElement) ShadowRoot() (*FirefoxElement, error) {
	raw, err := e.callOnSelfRaw("(el) => el.shadowRoot")
	if err != nil {
		return nil, err
	}
	return FromNode(e.Owner(), raw, nil), nil
}

// ClosedShadowRoot 通过调试桥尝试读取 closed shadow root。
func (e *FirefoxElement) ClosedShadowRoot() (*FirefoxElement, error) {
	raw, err := e.callOnSelfRaw(`(el) => {
		if (typeof window.__ruyiGetClosedShadowRoot !== 'function') return null;
		return window.__ruyiGetClosedShadowRoot(el);
	}`)
	if err != nil {
		return nil, err
	}
	return FromNode(e.Owner(), raw, nil), nil
}

// WithShadow 根据 mode 返回对应的 shadow root。
func (e *FirefoxElement) WithShadow(mode string) (*FirefoxElement, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "open":
		return e.ShadowRoot()
	case "closed":
		return e.ClosedShadowRoot()
	default:
		return nil, support.NewRuyiPageError("mode 只能是 open 或 closed", nil)
	}
}

// ClickSelf 点击当前元素。
func (e *FirefoxElement) ClickSelf(byJS bool, timeout time.Duration) error {
	if byJS {
		_, err := e.callOnSelfRaw("(el) => el.click()")
		return err
	}

	if timeout <= 0 {
		timeout = defaultElementClickTimeout
	}
	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		if err := e.ScrollToSee(true); err != nil {
			lastErr = err
		} else {
			point, err := e.centerPoint(false)
			if err != nil {
				lastErr = err
			} else if point["x"] == 0 && point["y"] == 0 {
				lastErr = support.NewCanNotClickError("无法获取元素可点击坐标", nil)
			} else {
				if err := e.performPointerActions([]map[string]any{
					{"type": "pointerMove", "x": point["x"], "y": point["y"], "duration": 50},
					{"type": "pointerDown", "button": 0},
					{"type": "pause", "duration": 50},
					{"type": "pointerUp", "button": 0},
				}); err == nil {
					return nil
				} else {
					lastErr = err
				}
			}
		}

		if time.Now().After(deadline) {
			break
		}
		time.Sleep(elementWaitPollInterval)
	}
	if lastErr == nil {
		lastErr = support.NewCanNotClickError("元素不可点击", nil)
	}
	var noRectErr *support.NoRectError
	var canNotClickErr *support.CanNotClickError
	if stderrors.As(lastErr, &noRectErr) || stderrors.As(lastErr, &canNotClickErr) {
		return lastErr
	}
	return support.NewCanNotClickError("元素不可点击", lastErr)
}

// ClickOffset 在元素左上角偏移点点击。
func (e *FirefoxElement) ClickOffset(offsetX int, offsetY int) error {
	location, err := e.ViewportLocation()
	if err != nil {
		return err
	}
	return e.performPointerActions([]map[string]any{
		{
			"type":     "pointerMove",
			"x":        location["x"] + offsetX,
			"y":        location["y"] + offsetY,
			"duration": 50,
		},
		{"type": "pointerDown", "button": 0},
		{"type": "pause", "duration": 50},
		{"type": "pointerUp", "button": 0},
	})
}

// RightClick 执行右键点击。
func (e *FirefoxElement) RightClick() error {
	if err := e.ScrollToSee(true); err != nil {
		return err
	}
	point, err := e.centerPoint(false)
	if err != nil {
		return err
	}
	return e.performPointerActions([]map[string]any{
		{"type": "pointerMove", "x": point["x"], "y": point["y"], "duration": 50},
		{"type": "pointerDown", "button": 2},
		{"type": "pause", "duration": 50},
		{"type": "pointerUp", "button": 2},
	})
}

// MiddleClick 执行中键点击。
func (e *FirefoxElement) MiddleClick() error {
	if err := e.ScrollToSee(true); err != nil {
		return err
	}
	point, err := e.centerPoint(false)
	if err != nil {
		return err
	}
	return e.performPointerActions([]map[string]any{
		{"type": "pointerMove", "x": point["x"], "y": point["y"], "duration": 50},
		{"type": "pointerDown", "button": 1},
		{"type": "pause", "duration": 50},
		{"type": "pointerUp", "button": 1},
	})
}

// DoubleClick 执行双击。
func (e *FirefoxElement) DoubleClick() error {
	if err := e.ScrollToSee(true); err != nil {
		return err
	}
	point, err := e.centerPoint(false)
	if err != nil {
		return err
	}
	return e.performPointerActions([]map[string]any{
		{"type": "pointerMove", "x": point["x"], "y": point["y"], "duration": 50},
		{"type": "pointerDown", "button": 0},
		{"type": "pause", "duration": 50},
		{"type": "pointerUp", "button": 0},
		{"type": "pause", "duration": 50},
		{"type": "pointerDown", "button": 0},
		{"type": "pause", "duration": 50},
		{"type": "pointerUp", "button": 0},
	})
}

// ClickForNewTab 点击当前元素并等待新顶层 context 打开。
func (e *FirefoxElement) ClickForNewTab(timeout time.Duration) (string, error) {
	if timeout <= 0 {
		timeout = 6 * time.Second
	}

	oldTabs, err := e.topLevelContextIDs()
	if err != nil {
		return "", err
	}
	if err := e.ClickSelf(false, 0); err != nil {
		var noRectErr *support.NoRectError
		var canNotClickErr *support.CanNotClickError
		if stderrors.As(err, &noRectErr) || stderrors.As(err, &canNotClickErr) {
			if jsErr := e.ClickSelf(true, 0); jsErr != nil {
				return "", jsErr
			}
		} else {
			return "", err
		}
	}

	value, matched, err := support.WaitUntil(func() (string, bool, error) {
		currentTabs, err := e.topLevelContextIDs()
		if err != nil {
			return "", false, err
		}
		for contextID := range currentTabs {
			if _, exists := oldTabs[contextID]; !exists {
				return contextID, true, nil
			}
		}
		return "", false, nil
	}, timeout, 300*time.Millisecond)
	if err != nil {
		return "", err
	}
	if !matched {
		if support.Settings.RaiseWhenWaitFailed {
			return "", support.NewWaitTimeoutError("等待新标签页打开超时", nil)
		}
		return "", nil
	}
	return value, nil
}

// Input 向元素输入内容；文件输入会自动切换到 setFiles。
func (e *FirefoxElement) Input(text any, clear bool, byJS bool) error {
	tag, err := e.Tag()
	if err != nil {
		return err
	}
	inputType, err := e.Attr("type")
	if err != nil {
		return err
	}
	if tag == "input" && strings.EqualFold(inputType, "file") {
		return e.UploadFiles(anyToStrings(text)...)
	}

	if byJS {
		if clear {
			if err := e.SetValue(""); err != nil {
				return err
			}
		}
		return e.SetValue(stringify(text))
	}

	if err := e.Focus(); err != nil {
		return err
	}
	if clear {
		if err := e.Clear(); err != nil {
			return err
		}
	}

	keyActions, err := bidi.BuildKeyAction(stringify(text))
	if err != nil {
		return err
	}
	_, err = bidi.PerformActions(e.browserDriver(), e.contextID(), keyActions, e.baseTimeout())
	return err
}

// UploadFiles 上传文件到文件输入框。
func (e *FirefoxElement) UploadFiles(files ...string) error {
	result, err := bidi.SetFiles(e.browserDriver(), e.contextID(), e.BiDiSharedReference(), cloneStringSlice(files), e.baseTimeout())
	if err != nil {
		return err
	}
	_ = result
	return nil
}

// Clear 清空输入内容。
func (e *FirefoxElement) Clear() error {
	if err := e.Focus(); err != nil {
		return err
	}
	actions, err := bidi.BuildKeyAction([]bidi.KeyChord{{Modifier: support.Keys.CONTROL, Key: "a"}})
	if err != nil {
		return err
	}
	if _, err = bidi.PerformActions(e.browserDriver(), e.contextID(), actions, e.baseTimeout()); err != nil {
		return err
	}
	deleteActions, err := bidi.BuildKeyAction([]string{support.Keys.DELETE})
	if err != nil {
		return err
	}
	_, err = bidi.PerformActions(e.browserDriver(), e.contextID(), deleteActions, e.baseTimeout())
	return err
}

// Hover 让鼠标悬停到元素中心点。
func (e *FirefoxElement) Hover() error {
	if err := e.ScrollToSee(true); err != nil {
		return err
	}
	point, err := e.centerPoint(false)
	if err != nil {
		return err
	}
	return e.performPointerActions([]map[string]any{
		{"type": "pointerMove", "x": point["x"], "y": point["y"], "duration": 100},
	})
}

// DragTo 将当前元素拖拽到目标元素或坐标。
func (e *FirefoxElement) DragTo(target any, duration time.Duration) error {
	if duration <= 0 {
		duration = 500 * time.Millisecond
	}

	var end map[string]int
	switch typed := target.(type) {
	case *FirefoxElement:
		if typed == nil {
			return support.NewRuyiPageError("拖拽目标不能为空", nil)
		}
		if err := typed.ScrollToSee(true); err != nil {
			return err
		}
		point, err := typed.centerPoint(false)
		if err != nil {
			return err
		}
		end = point
	case map[string]int:
		end = map[string]int{"x": typed["x"], "y": typed["y"]}
	case map[string]any:
		end = map[string]int{"x": intFromAny(typed["x"]), "y": intFromAny(typed["y"])}
	case []int:
		if len(typed) < 2 {
			return support.NewRuyiPageError("拖拽坐标至少需要两个值", nil)
		}
		end = map[string]int{"x": typed[0], "y": typed[1]}
	case []any:
		if len(typed) < 2 {
			return support.NewRuyiPageError("拖拽坐标至少需要两个值", nil)
		}
		end = map[string]int{"x": intFromAny(typed[0]), "y": intFromAny(typed[1])}
	default:
		return support.NewRuyiPageError(fmt.Sprintf("不支持的拖拽目标类型: %T", target), nil)
	}

	if err := e.ScrollToSee(true); err != nil {
		return err
	}
	start, err := e.centerPoint(false)
	if err != nil {
		return err
	}

	totalMS := int(duration / time.Millisecond)
	if totalMS < 50 {
		totalMS = 50
	}
	steps := totalMS / 30
	if steps < 8 {
		steps = 8
	}
	if steps > 30 {
		steps = 30
	}
	stepDuration := totalMS / steps

	actions := []map[string]any{
		{"type": "pointerMove", "origin": "viewport", "x": start["x"], "y": start["y"], "duration": 0},
		{"type": "pointerDown", "button": 0},
		{"type": "pause", "duration": 80},
	}
	for index := 1; index <= steps; index++ {
		actions = append(actions, map[string]any{
			"type":     "pointerMove",
			"origin":   "viewport",
			"x":        start["x"] + (end["x"]-start["x"])*index/steps,
			"y":        start["y"] + (end["y"]-start["y"])*index/steps,
			"duration": stepDuration,
		})
	}
	actions = append(actions,
		map[string]any{"type": "pause", "duration": 80},
		map[string]any{"type": "pointerUp", "button": 0},
	)
	return e.performPointerActions(actions)
}

// Screenshot 对元素执行截图。
func (e *FirefoxElement) Screenshot(path string) ([]byte, error) {
	result, err := bidi.CaptureScreenshot(
		e.browserDriver(),
		e.contextID(),
		"viewport",
		nil,
		map[string]any{
			"type":    "element",
			"element": e.BiDiSharedReference(),
		},
		e.baseTimeout(),
	)
	if err != nil {
		return nil, err
	}

	data, err := base64.StdEncoding.DecodeString(readString(result, "data"))
	if err != nil {
		return nil, err
	}
	if path != "" {
		if err := writeBytes(path, data); err != nil {
			return nil, err
		}
	}
	return data, nil
}

// Focus 聚焦当前元素。
func (e *FirefoxElement) Focus() error {
	_, err := e.callOnSelfRaw("(el) => el.focus()")
	return err
}

// ScrollToSee 将元素滚动到视口内。
func (e *FirefoxElement) ScrollToSee(center bool) error {
	block := "nearest"
	if center {
		block = "center"
	}
	_, err := e.callOnSelfRaw(fmt.Sprintf(`(el) => el.scrollIntoView({block: %q, inline: "nearest"})`, block))
	return err
}

// ScrollToTop 将元素内部滚动到顶部。
func (e *FirefoxElement) ScrollToTop() error {
	return e.scrollElementUntil(func() (bool, error) {
		value, err := e.callOnSelfParsed("(el) => (el.scrollTop || 0) <= 0")
		return toBool(value), err
	}, 0, -600, 20, 100*time.Millisecond)
}

// ScrollToBottom 将元素内部滚动到底部。
func (e *FirefoxElement) ScrollToBottom() error {
	return e.scrollElementUntil(func() (bool, error) {
		value, err := e.callOnSelfParsed("(el) => el.scrollTop + el.clientHeight >= el.scrollHeight - 2")
		return toBool(value), err
	}, 0, 600, 20, 100*time.Millisecond)
}

// ScrollBy 在元素内部执行滚动。
func (e *FirefoxElement) ScrollBy(deltaX int, deltaY int) error {
	point, err := e.centerPoint(false)
	if err != nil {
		return err
	}
	actions := bidi.BuildWheelAction(point["x"], point["y"], &bidi.WheelActionOptions{
		DeltaX: deltaX,
		DeltaY: deltaY,
		Origin: "viewport",
	})
	_, err = bidi.PerformActions(e.browserDriver(), e.contextID(), actions, e.baseTimeout())
	return err
}

// SetAttr 设置属性值。
func (e *FirefoxElement) SetAttr(name string, value string) error {
	_, err := e.callOnSelfRaw("(el, name, value) => el.setAttribute(name, value)", name, value)
	return err
}

// RemoveAttr 删除属性。
func (e *FirefoxElement) RemoveAttr(name string) error {
	_, err := e.callOnSelfRaw("(el, name) => el.removeAttribute(name)", name)
	return err
}

// SetProp 设置 JS 属性。
func (e *FirefoxElement) SetProp(name string, value any) error {
	_, err := e.callOnSelfRaw("(el, name, value) => { el[name] = value; }", name, value)
	return err
}

// SetStyle 设置内联样式。
func (e *FirefoxElement) SetStyle(name string, value string) error {
	_, err := e.callOnSelfRaw("(el, name, value) => { el.style.setProperty(name, value); }", name, value)
	return err
}

// SetInnerHTML 设置 innerHTML。
func (e *FirefoxElement) SetInnerHTML(html string) error {
	_, err := e.callOnSelfRaw("(el, html) => { el.innerHTML = html; }", html)
	return err
}

// SetValue 使用 JS 设置 value 并触发 input/change。
func (e *FirefoxElement) SetValue(value string) error {
	_, err := e.callOnSelfRaw(`(el, value) => {
		el.value = value;
		el.dispatchEvent(new Event("input", {bubbles: true}));
		el.dispatchEvent(new Event("change", {bubbles: true}));
	}`, value)
	return err
}

// SelectByText 通过文本选择 <select> 选项。
func (e *FirefoxElement) SelectByText(text string, timeout time.Duration, mode string) (bool, error) {
	resolvedMode, err := resolveSelectMode(mode)
	if err != nil {
		return false, err
	}
	targetIndex, found, err := e.selectTargetIndex(func(option map[string]any) bool {
		optionText := strings.TrimSpace(stringify(option["text"]))
		return optionText == text || strings.Contains(optionText, text)
	})
	if err != nil || !found {
		return false, err
	}
	if ok, err := e.selectByIndexNative(targetIndex, timeout); err != nil {
		return false, err
	} else if ok {
		return true, nil
	}
	if resolvedMode != "compat" {
		return false, nil
	}
	return e.selectByTextJS(text)
}

// SelectByValue 通过 value 选择 <select> 选项。
func (e *FirefoxElement) SelectByValue(value string, mode string) (bool, error) {
	resolvedMode, err := resolveSelectMode(mode)
	if err != nil {
		return false, err
	}
	targetIndex, found, err := e.selectTargetIndex(func(option map[string]any) bool {
		return stringify(option["value"]) == value
	})
	if err != nil || !found {
		return false, err
	}
	if ok, err := e.selectByIndexNative(targetIndex, 0); err != nil {
		return false, err
	} else if ok {
		return true, nil
	}
	if resolvedMode != "compat" {
		return false, nil
	}
	return e.selectByValueJS(value)
}

// SelectByIndex 通过索引选择 <select> 选项。
func (e *FirefoxElement) SelectByIndex(index int, mode string) (bool, error) {
	resolvedMode, err := resolveSelectMode(mode)
	if err != nil {
		return false, err
	}
	if ok, err := e.selectByIndexNative(index, 0); err != nil {
		return false, err
	} else if ok {
		return true, nil
	}
	if resolvedMode != "compat" {
		return false, nil
	}
	return e.selectByIndexJS(index)
}

// SelectCancelByIndex 取消指定索引选项的选中状态。
func (e *FirefoxElement) SelectCancelByIndex(index int) (bool, error) {
	value, err := e.callOnSelfParsed(`(el, targetIndex) => {
		if (!el || el.tagName.toLowerCase() !== 'select') return false;
		const option = el.options[targetIndex];
		if (!option) return false;
		option.selected = false;
		el.dispatchEvent(new Event("change", {bubbles: true}));
		return true;
	}`, index)
	if err != nil {
		return false, err
	}
	return toBool(value), nil
}

// SelectCancelByText 取消指定文本选项的选中状态。
func (e *FirefoxElement) SelectCancelByText(text string) (bool, error) {
	value, err := e.callOnSelfParsed(`(el, targetText) => {
		if (!el || el.tagName.toLowerCase() !== 'select') return false;
		for (const option of Array.from(el.options)) {
			if (option.text === targetText || option.textContent.trim() === targetText) {
				option.selected = false;
				el.dispatchEvent(new Event("change", {bubbles: true}));
				return true;
			}
		}
		return false;
	}`, text)
	if err != nil {
		return false, err
	}
	return toBool(value), nil
}

// SelectAllOptions 选中全部选项。
func (e *FirefoxElement) SelectAllOptions() error {
	_, err := e.callOnSelfRaw(`(el) => {
		if (!el || el.tagName.toLowerCase() !== 'select') return;
		for (const option of Array.from(el.options)) option.selected = true;
		el.dispatchEvent(new Event("change", {bubbles: true}));
	}`)
	return err
}

// DeselectAllOptions 取消全部选项。
func (e *FirefoxElement) DeselectAllOptions() error {
	_, err := e.callOnSelfRaw(`(el) => {
		if (!el || el.tagName.toLowerCase() !== 'select') return;
		for (const option of Array.from(el.options)) option.selected = false;
		el.dispatchEvent(new Event("change", {bubbles: true}));
	}`)
	return err
}

// SelectOptions 返回全部 option 快照。
func (e *FirefoxElement) SelectOptions() ([]map[string]any, error) {
	value, err := e.callOnSelfParsed(`(el) => {
		if (!el || el.tagName.toLowerCase() !== 'select') return [];
		return Array.from(el.options).map((option) => ({
			index: option.index,
			text: option.text,
			value: option.value,
			selected: option.selected,
			disabled: option.disabled,
		}));
	}`)
	if err != nil {
		return nil, err
	}
	return anySliceToMapSlice(value), nil
}

// SelectSelectedOption 返回当前选中项。
func (e *FirefoxElement) SelectSelectedOption() (map[string]any, error) {
	value, err := e.callOnSelfParsed(`(el) => {
		if (!el || el.tagName.toLowerCase() !== 'select' || el.selectedIndex < 0) return null;
		const option = el.options[el.selectedIndex];
		return option ? {
			text: option.text,
			value: option.value,
			index: option.index,
			selected: option.selected,
			disabled: option.disabled,
		} : null;
	}`)
	if err != nil {
		return nil, err
	}
	return asAnyMap(value), nil
}

// SelectIsMulti 判断 <select> 是否为 multiple。
func (e *FirefoxElement) SelectIsMulti() (bool, error) {
	value, err := e.callOnSelfParsed("(el) => !!(el && el.tagName.toLowerCase() === 'select' && el.multiple)")
	if err != nil {
		return false, err
	}
	return toBool(value), nil
}

// WaitDisplayed 等待元素显示。
func (e *FirefoxElement) WaitDisplayed(timeout time.Duration) (bool, error) {
	return e.waitState(timeout, func() (bool, error) { return e.IsDisplayed() }, "等待元素显示超时")
}

// WaitHidden 等待元素隐藏。
func (e *FirefoxElement) WaitHidden(timeout time.Duration) (bool, error) {
	return e.waitState(timeout, func() (bool, error) {
		displayed, err := e.IsDisplayed()
		return !displayed, err
	}, "等待元素隐藏超时")
}

// WaitEnabled 等待元素可用。
func (e *FirefoxElement) WaitEnabled(timeout time.Duration) (bool, error) {
	return e.waitState(timeout, func() (bool, error) { return e.IsEnabled() }, "等待元素可用超时")
}

// WaitDisabled 等待元素禁用。
func (e *FirefoxElement) WaitDisabled(timeout time.Duration) (bool, error) {
	return e.waitState(timeout, func() (bool, error) {
		enabled, err := e.IsEnabled()
		return !enabled, err
	}, "等待元素禁用超时")
}

// RunJS 在元素上执行 JS，this 绑定为当前元素。
func (e *FirefoxElement) RunJS(script string, args ...any) (any, error) {
	if err := e.ensureReady(); err != nil {
		return nil, err
	}
	result, err := bidi.CallFunction(
		e.browserDriver(),
		script,
		map[string]any{"context": e.contextID()},
		args,
		e.BiDiSharedReference(),
		nil,
		"",
		"root",
		map[string]any{"maxDomDepth": 0, "includeShadowTree": "open"},
		false,
		e.scriptTimeout(),
	)
	if err != nil {
		if isNodeLostError(err) {
			if refreshErr := e.refreshID(); refreshErr != nil {
				return nil, refreshErr
			}
			result, err = bidi.CallFunction(
				e.browserDriver(),
				script,
				map[string]any{"context": e.contextID()},
				args,
				e.BiDiSharedReference(),
				nil,
				"",
				"root",
				map[string]any{"maxDomDepth": 0, "includeShadowTree": "open"},
				false,
				e.scriptTimeout(),
			)
		}
		if err != nil {
			if wrapped := wrapContextLoss(err); wrapped != nil {
				return nil, wrapped
			}
			return nil, err
		}
	}
	return support.ParseBiDiValue(result.Result.Raw), nil
}

// Parent 获取父元素。
func (e *FirefoxElement) Parent(locator any, index int) (*FirefoxElement, error) {
	if locator != nil {
		selector, err := selectorText(locator)
		if err != nil {
			return nil, err
		}
		raw, err := e.callOnSelfRaw(`(el, sel, idx) => {
			let current = el.parentElement;
			let count = 0;
			while (current) {
				try {
					if (current.matches && current.matches(sel)) {
						count += 1;
						if (count >= idx) return current;
					}
				} catch (error) {}
				current = current.parentElement;
			}
			return null;
		}`, selector, normalizePositiveIndex(index))
		if err != nil {
			return nil, err
		}
		return FromNode(e.Owner(), raw, nil), nil
	}

	raw, err := e.callOnSelfRaw(`(el, idx) => {
		let current = el;
		for (let index = 0; index < idx; index += 1) {
			current = current ? current.parentElement : null;
		}
		return current;
	}`, normalizePositiveIndex(index))
	if err != nil {
		return nil, err
	}
	return FromNode(e.Owner(), raw, nil), nil
}

// Child 获取单个子元素。
func (e *FirefoxElement) Child(locator any, index int, timeout time.Duration) (*FirefoxElement, error) {
	if locator != nil {
		return e.Ele(locator, index, timeout)
	}
	raw, err := e.callOnSelfRaw("(el, idx) => el.children[idx - 1] || null", normalizePositiveIndex(index))
	if err != nil {
		return nil, err
	}
	return FromNode(e.Owner(), raw, nil), nil
}

// Children 获取全部子元素。
func (e *FirefoxElement) Children(locator any, timeout time.Duration) ([]*FirefoxElement, error) {
	if locator != nil {
		return e.Eles(locator, timeout)
	}
	raw, err := e.callOnSelfRaw("(el) => Array.from(el.children)")
	if err != nil {
		return nil, err
	}
	return remoteArrayToElements(e.Owner(), raw, nil), nil
}

// Next 获取后续兄弟元素。
func (e *FirefoxElement) Next(locator any, index int) (*FirefoxElement, error) {
	if locator != nil {
		selector, err := selectorText(locator)
		if err != nil {
			return nil, err
		}
		raw, err := e.callOnSelfRaw(`(el, sel, idx) => {
			let current = el.nextElementSibling;
			let count = 0;
			while (current) {
				try {
					if (!sel || (current.matches && current.matches(sel))) {
						count += 1;
						if (count >= idx) return current;
					}
				} catch (error) {}
				current = current.nextElementSibling;
			}
			return null;
		}`, selector, normalizePositiveIndex(index))
		if err != nil {
			return nil, err
		}
		return FromNode(e.Owner(), raw, nil), nil
	}

	raw, err := e.callOnSelfRaw(`(el, idx) => {
		let current = el;
		for (let index = 0; index < idx; index += 1) {
			current = current ? current.nextElementSibling : null;
		}
		return current;
	}`, normalizePositiveIndex(index))
	if err != nil {
		return nil, err
	}
	return FromNode(e.Owner(), raw, nil), nil
}

// Prev 获取前置兄弟元素。
func (e *FirefoxElement) Prev(locator any, index int) (*FirefoxElement, error) {
	if locator != nil {
		selector, err := selectorText(locator)
		if err != nil {
			return nil, err
		}
		raw, err := e.callOnSelfRaw(`(el, sel, idx) => {
			let current = el.previousElementSibling;
			let count = 0;
			while (current) {
				try {
					if (!sel || (current.matches && current.matches(sel))) {
						count += 1;
						if (count >= idx) return current;
					}
				} catch (error) {}
				current = current.previousElementSibling;
			}
			return null;
		}`, selector, normalizePositiveIndex(index))
		if err != nil {
			return nil, err
		}
		return FromNode(e.Owner(), raw, nil), nil
	}

	raw, err := e.callOnSelfRaw(`(el, idx) => {
		let current = el;
		for (let index = 0; index < idx; index += 1) {
			current = current ? current.previousElementSibling : null;
		}
		return current;
	}`, normalizePositiveIndex(index))
	if err != nil {
		return nil, err
	}
	return FromNode(e.Owner(), raw, nil), nil
}

// Ele 在当前元素内查找单个子元素。
func (e *FirefoxElement) Ele(locator any, index int, timeout time.Duration) (*FirefoxElement, error) {
	return e.findRelativeElement(locator, index, timeout)
}

// Eles 在当前元素内查找全部子元素。
func (e *FirefoxElement) Eles(locator any, timeout time.Duration) ([]*FirefoxElement, error) {
	if e == nil || e.Owner() == nil {
		return []*FirefoxElement{}, nil
	}
	return e.Owner().FindElements(locator, timeout, []map[string]any{e.BiDiSharedReference()})
}

// SEle 在当前动态元素的 innerHTML 上执行静态解析查找。
func (e *FirefoxElement) SEle(locator any) (StaticNode, error) {
	if e == nil {
		return NewNoneElement("s_ele", map[string]any{"locator": locator}), nil
	}
	htmlText, err := e.InnerHTML()
	if err != nil {
		return nil, err
	}
	return MakeStaticElement(htmlText, locator), nil
}

// SEles 在当前动态元素的 innerHTML 上执行静态解析查找。
func (e *FirefoxElement) SEles(locator any) ([]*StaticElement, error) {
	if e == nil {
		return []*StaticElement{}, nil
	}
	htmlText, err := e.InnerHTML()
	if err != nil {
		return nil, err
	}
	return MakeStaticElements(htmlText, locator), nil
}

// Click 返回点击管理器。
func (e *FirefoxElement) Click() *units.Clicker {
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.clicker == nil {
		e.clicker = units.NewClicker(e)
	}
	return e.clicker
}

// Scroll 返回滚动管理器。
func (e *FirefoxElement) Scroll() *units.ElementScroller {
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.scroll == nil {
		e.scroll = units.NewElementScroller(e)
	}
	return e.scroll
}

// Rect 返回矩形管理器。
func (e *FirefoxElement) Rect() *units.ElementRect {
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.rect == nil {
		e.rect = units.NewElementRect(e)
	}
	return e.rect
}

// Set 返回 setter 管理器。
func (e *FirefoxElement) Set() *units.ElementSetter {
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.setter == nil {
		e.setter = units.NewElementSetter(e)
	}
	return e.setter
}

// States 返回状态管理器。
func (e *FirefoxElement) States() *units.ElementStates {
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.states == nil {
		e.states = units.NewElementStates(e)
	}
	return e.states
}

// Wait 返回等待管理器。
func (e *FirefoxElement) Wait() *units.ElementWaiter {
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.waiter == nil {
		e.waiter = units.NewElementWaiter(e)
	}
	return e.waiter
}

// Select 返回 <select> 管理器。
func (e *FirefoxElement) Select() *units.SelectElement {
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.selects == nil {
		e.selects = units.NewSelectElement(e)
	}
	return e.selects
}

func (e *FirefoxElement) String() string {
	if e == nil {
		return "<FirefoxElement >"
	}
	tag := "?"
	if value, err := e.Tag(); err == nil && value != "" {
		tag = value
	}
	attrs, _ := e.Attrs()
	parts := []string{tag}
	if attrs["id"] != "" {
		parts = append(parts, "#"+attrs["id"])
	}
	if attrs["class"] != "" {
		classNames := strings.Fields(attrs["class"])
		if len(classNames) > 0 {
			parts = append(parts, "."+classNames[0])
		}
	}
	return fmt.Sprintf("<FirefoxElement %s>", strings.Join(parts, ""))
}

func (e *FirefoxElement) ensureReady() error {
	if e == nil {
		return support.NewPageDisconnectedError("FirefoxElement 未初始化", nil)
	}
	if e.Owner() == nil {
		return support.NewPageDisconnectedError("FirefoxElement owner 未初始化", nil)
	}
	if e.contextID() == "" {
		return support.NewContextLostError("context id 不能为空", nil)
	}
	if driver := e.browserDriver(); driver == nil || !driver.IsRunning() {
		return support.NewPageDisconnectedError("Firefox driver 未连接", nil)
	}
	if e.SharedID() == "" {
		return support.NewElementLostError("元素 shared id 为空", nil)
	}
	return nil
}

func (e *FirefoxElement) findRelativeElement(locator any, index int, timeout time.Duration) (*FirefoxElement, error) {
	if e == nil || e.Owner() == nil {
		return nil, nil
	}
	return e.Owner().FindElement(locator, index, timeout, []map[string]any{e.BiDiSharedReference()})
}

func (e *FirefoxElement) waitState(timeout time.Duration, check func() (bool, error), message string) (bool, error) {
	timeout = resolveElementTimeout(timeout)
	deadline := time.Now().Add(timeout)
	for {
		matched, err := check()
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
		if time.Now().After(deadline) {
			if support.Settings.RaiseWhenWaitFailed {
				return false, support.NewWaitTimeoutError(fmt.Sprintf("%s (%s)", message, timeout), nil)
			}
			return false, nil
		}
		time.Sleep(elementWaitPollInterval)
	}
}

func (e *FirefoxElement) scrollElementUntil(check func() (bool, error), deltaX int, deltaY int, maxSteps int, pause time.Duration) error {
	for index := 0; index < maxSteps; index++ {
		matched, err := check()
		if err != nil {
			return err
		}
		if matched {
			return nil
		}
		if err := e.ScrollBy(deltaX, deltaY); err != nil {
			return err
		}
		time.Sleep(pause)
	}
	return nil
}

func (e *FirefoxElement) topLevelContextIDs() (map[string]struct{}, error) {
	result, err := bidi.GetTree(e.browserDriver(), nil, "", e.baseTimeout())
	if err != nil {
		return nil, err
	}
	contexts := anyToMapSlice(result["contexts"])
	ids := make(map[string]struct{}, len(contexts))
	for _, context := range contexts {
		contextID := stringify(context["context"])
		if contextID != "" {
			ids[contextID] = struct{}{}
		}
	}
	return ids, nil
}

func resolveSelectMode(mode string) (string, error) {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = "compat"
	}
	switch mode {
	case "native_only", "native_first", "compat":
		return mode, nil
	default:
		return "", support.NewRuyiPageError("mode 必须是 native_only、native_first 或 compat", nil)
	}
}

func (e *FirefoxElement) selectTargetIndex(matcher func(option map[string]any) bool) (int, bool, error) {
	options, err := e.SelectOptions()
	if err != nil {
		return 0, false, err
	}
	for _, option := range options {
		if matcher(option) {
			return intFromAny(option["index"]), true, nil
		}
	}
	return 0, false, nil
}

func (e *FirefoxElement) selectState() (map[string]any, error) {
	value, err := e.callOnSelfParsed(`(el) => {
		if (!el || el.tagName.toLowerCase() !== 'select') {
			throw new Error('当前元素不是 <select>');
		}
		const rect = el.getBoundingClientRect();
		return {
			selectedIndex: el.selectedIndex,
			value: el.value,
			multiple: !!el.multiple,
			size: Number(el.size || 0),
			disabled: !!el.disabled,
			focused: document.activeElement === el,
			rect: {x: rect.x, y: rect.y, width: rect.width, height: rect.height},
			options: Array.from(el.options).map((option) => ({
				text: option.text,
				value: option.value,
				selected: option.selected,
				index: option.index,
				disabled: option.disabled,
			})),
		};
	}`)
	if err != nil {
		return nil, err
	}
	return asAnyMap(value), nil
}

func (e *FirefoxElement) selectFocusNative() (bool, error) {
	if err := e.ScrollToSee(true); err != nil {
		return false, err
	}
	if err := e.selectNativeClick(); err != nil {
		return false, err
	}
	time.Sleep(60 * time.Millisecond)
	state, err := e.selectState()
	if err == nil && toBool(state["focused"]) {
		return true, nil
	}
	if err := e.selectNativeClick(); err != nil {
		return false, err
	}
	time.Sleep(60 * time.Millisecond)
	state, err = e.selectState()
	if err != nil {
		return false, err
	}
	return toBool(state["focused"]), nil
}

func (e *FirefoxElement) selectNativeClick() error {
	_, _ = bidi.Activate(e.browserDriver(), e.contextID(), e.baseTimeout())
	actions := []map[string]any{
		{
			"type":       "pointer",
			"id":         "mouse0",
			"parameters": map[string]any{"pointerType": "mouse"},
			"actions": []map[string]any{
				{
					"type":     "pointerMove",
					"x":        0,
					"y":        0,
					"duration": 0,
					"origin": map[string]any{
						"type":    "element",
						"element": e.BiDiSharedReference(),
					},
				},
				{"type": "pointerDown", "button": 0},
				{"type": "pause", "duration": 50},
				{"type": "pointerUp", "button": 0},
			},
		},
	}
	_, err := bidi.PerformActions(e.browserDriver(), e.contextID(), actions, e.baseTimeout())
	return err
}

func (e *FirefoxElement) selectKeyPress(key string) error {
	actions, err := bidi.BuildKeyAction(key)
	if err != nil {
		return err
	}
	_, err = bidi.PerformActions(e.browserDriver(), e.contextID(), actions, e.baseTimeout())
	if err == nil {
		time.Sleep(20 * time.Millisecond)
	}
	return err
}

func (e *FirefoxElement) selectByIndexNative(index int, timeout time.Duration) (bool, error) {
	state, err := e.selectState()
	if err != nil {
		return false, err
	}
	if toBool(state["disabled"]) || toBool(state["multiple"]) {
		return false, nil
	}

	options := anyToMapSlice(state["options"])
	if index < 0 || index >= len(options) {
		return false, nil
	}
	if toBool(options[index]["disabled"]) {
		return false, nil
	}

	if intFromAny(state["selectedIndex"]) == index {
		return true, nil
	}
	if ok, err := e.selectFocusNative(); err != nil {
		return false, err
	} else if !ok {
		return false, nil
	}

	currentIndex := intFromAny(state["selectedIndex"])
	maxSteps := len(options) + 3
	if maxSteps < 1 {
		maxSteps = 1
	}
	deadline := time.Now().Add(resolveElementTimeout(timeout))

	for step := 0; step < maxSteps; step++ {
		if currentIndex == index {
			break
		}
		key := support.Keys.DOWN
		if currentIndex > index {
			key = support.Keys.UP
		}
		if err := e.selectKeyPress(key); err != nil {
			return false, err
		}
		state, err = e.selectState()
		if err != nil {
			return false, err
		}
		newIndex := intFromAny(state["selectedIndex"])
		if newIndex == currentIndex {
			if err := e.selectKeyPress(support.Keys.HOME); err != nil {
				return false, err
			}
			state, err = e.selectState()
			if err != nil {
				return false, err
			}
			newIndex = intFromAny(state["selectedIndex"])
		}
		currentIndex = newIndex
		if time.Now().After(deadline) {
			break
		}
	}
	if currentIndex != index {
		return false, nil
	}
	if err := e.selectKeyPress(support.Keys.ENTER); err != nil {
		return false, err
	}
	state, err = e.selectState()
	if err != nil {
		return false, err
	}
	return intFromAny(state["selectedIndex"]) == index, nil
}

func (e *FirefoxElement) selectByTextJS(text string) (bool, error) {
	value, err := e.callOnSelfParsed(`(el, targetText) => {
		if (!el || el.tagName.toLowerCase() !== 'select') return false;
		for (const option of Array.from(el.options)) {
			if (option.text === targetText || option.textContent.trim() === targetText) {
				option.selected = true;
				el.value = option.value;
				el.dispatchEvent(new Event("change", {bubbles: true}));
				return true;
			}
		}
		for (const option of Array.from(el.options)) {
			if ((option.text || '').includes(targetText) || (option.textContent || '').includes(targetText)) {
				option.selected = true;
				el.value = option.value;
				el.dispatchEvent(new Event("change", {bubbles: true}));
				return true;
			}
		}
		return false;
	}`, text)
	if err != nil {
		return false, err
	}
	return toBool(value), nil
}

func (e *FirefoxElement) selectByValueJS(value string) (bool, error) {
	valueResult, err := e.callOnSelfParsed(`(el, targetValue) => {
		if (!el || el.tagName.toLowerCase() !== 'select') return false;
		for (const option of Array.from(el.options)) {
			if (option.value === targetValue) {
				option.selected = true;
				el.value = option.value;
				el.dispatchEvent(new Event("change", {bubbles: true}));
				return true;
			}
		}
		return false;
	}`, value)
	if err != nil {
		return false, err
	}
	return toBool(valueResult), nil
}

func (e *FirefoxElement) selectByIndexJS(index int) (bool, error) {
	value, err := e.callOnSelfParsed(`(el, targetIndex) => {
		if (!el || el.tagName.toLowerCase() !== 'select') return false;
		if (targetIndex < 0 || targetIndex >= el.options.length) return false;
		el.selectedIndex = targetIndex;
		el.dispatchEvent(new Event("change", {bubbles: true}));
		return true;
	}`, index)
	if err != nil {
		return false, err
	}
	return toBool(value), nil
}

func (e *FirefoxElement) centerPoint(scrollFirst bool) (map[string]int, error) {
	script := `(el) => {
		const rect = el.getBoundingClientRect();
		if (rect.width === 0 && rect.height === 0) return null;
		return {
			x: Math.round(rect.left + rect.width / 2),
			y: Math.round(rect.top + rect.height / 2)
		};
	}`
	if scrollFirst {
		script = `(el) => {
			el.scrollIntoView({block: "center", inline: "nearest"});
			const rect = el.getBoundingClientRect();
			if (rect.width === 0 && rect.height === 0) return null;
			return {
				x: Math.round(rect.left + rect.width / 2),
				y: Math.round(rect.top + rect.height / 2)
			};
		}`
	}
	value, err := e.callOnSelfParsed(script)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, support.NewNoRectError("元素没有可用矩形区域", nil)
	}
	result := normalizeXYWH(value, "x", "y")
	if result["x"] == 0 && result["y"] == 0 {
		size, sizeErr := e.Size()
		if sizeErr == nil && size["width"] == 0 && size["height"] == 0 {
			return nil, support.NewNoRectError("元素没有可用矩形区域", nil)
		}
	}
	return result, nil
}

func (e *FirefoxElement) callOnSelfParsed(functionDeclaration string, args ...any) (any, error) {
	raw, err := e.callOnSelfRaw(functionDeclaration, args...)
	if err != nil {
		return nil, err
	}
	return support.ParseBiDiValue(raw), nil
}

func (e *FirefoxElement) callOnSelfRaw(functionDeclaration string, args ...any) (map[string]any, error) {
	if err := e.ensureReady(); err != nil {
		return nil, err
	}

	callArgs := make([]any, 0, len(args)+1)
	callArgs = append(callArgs, e.BiDiSharedReference())
	callArgs = append(callArgs, args...)

	result, err := bidi.CallFunction(
		e.browserDriver(),
		functionDeclaration,
		map[string]any{"context": e.contextID()},
		callArgs,
		nil,
		nil,
		"",
		"root",
		map[string]any{"maxDomDepth": 0, "includeShadowTree": "open"},
		false,
		e.scriptTimeout(),
	)
	if err != nil {
		if isNodeLostError(err) {
			if refreshErr := e.refreshID(); refreshErr != nil {
				return nil, refreshErr
			}
			callArgs[0] = e.BiDiSharedReference()
			result, err = bidi.CallFunction(
				e.browserDriver(),
				functionDeclaration,
				map[string]any{"context": e.contextID()},
				callArgs,
				nil,
				nil,
				"",
				"root",
				map[string]any{"maxDomDepth": 0, "includeShadowTree": "open"},
				false,
				e.scriptTimeout(),
			)
		}
		if err != nil {
			if wrapped := wrapContextLoss(err); wrapped != nil {
				return nil, wrapped
			}
			return nil, err
		}
	}
	return cloneAnyMap(result.Result.Raw), nil
}

func (e *FirefoxElement) refreshID() error {
	if e == nil || e.Owner() == nil {
		return support.NewElementLostError("元素引用已失效", nil)
	}

	locator, startNodes, ok := decodeLocatorInfo(e.locatorInfo)
	if !ok {
		var rebuilt bool
		locator, rebuilt = rebuildLocatorFromNodeInfo(e.NodeInfo())
		if !rebuilt {
			return support.NewElementLostError(fmt.Sprintf("元素引用已失效: %s", e.SharedID()), nil)
		}
	}

	reloaded, err := e.Owner().FindElement(locator, 1, e.Owner().ElementFindTimeout(), startNodes)
	if err != nil {
		return err
	}
	if reloaded == nil {
		return support.NewElementLostError(fmt.Sprintf("元素引用已失效: %s", e.SharedID()), nil)
	}

	e.mu.Lock()
	e.sharedID = reloaded.sharedID
	e.handle = reloaded.handle
	e.nodeInfo = cloneAnyMap(reloaded.nodeInfo)
	e.mu.Unlock()
	return nil
}

func (e *FirefoxElement) performPointerActions(pointerActions []map[string]any) error {
	actions := []map[string]any{
		{
			"type":       "pointer",
			"id":         "mouse0",
			"parameters": map[string]any{"pointerType": "mouse"},
			"actions":    cloneAnyMapSlice(pointerActions),
		},
	}
	_, err := bidi.PerformActions(e.browserDriver(), e.contextID(), actions, e.baseTimeout())
	return err
}

func (e *FirefoxElement) browserDriver() *base.BrowserBiDiDriver {
	if e == nil || e.Owner() == nil {
		return nil
	}
	return e.Owner().BrowserDriver()
}

func (e *FirefoxElement) contextID() string {
	if e == nil || e.Owner() == nil {
		return ""
	}
	return e.Owner().ContextID()
}

func (e *FirefoxElement) baseTimeout() time.Duration {
	if e == nil || e.Owner() == nil {
		return time.Second
	}
	return e.Owner().BaseTimeout()
}

func (e *FirefoxElement) scriptTimeout() time.Duration {
	if e == nil || e.Owner() == nil {
		return time.Second
	}
	return e.Owner().ScriptTimeout()
}

func remoteArrayToElements(owner Owner, raw map[string]any, locatorInfo any) []*FirefoxElement {
	if readString(raw, "type") != "array" {
		return []*FirefoxElement{}
	}
	values, _ := raw["value"].([]any)
	result := make([]*FirefoxElement, 0, len(values))
	for _, value := range values {
		if node, ok := value.(map[string]any); ok {
			if element := FromNode(owner, node, locatorInfo); element != nil {
				result = append(result, element)
			}
		}
	}
	return result
}

func resolveElementTimeout(timeout time.Duration) time.Duration {
	if timeout > 0 {
		return timeout
	}
	seconds := support.Settings.ElementFindTimeout
	if seconds <= 0 {
		seconds = support.DefaultElementFindTimeoutSeconds
	}
	return time.Duration(seconds * float64(time.Second))
}

func isNodeLostError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "no such node") || strings.Contains(text, "stale")
}

func wrapContextLoss(err error) error {
	if err == nil {
		return nil
	}
	var bidiErr *support.BiDiError
	if !stderrors.As(err, &bidiErr) {
		return nil
	}
	lowered := strings.ToLower(bidiErr.Error())
	if strings.Contains(lowered, "no such frame") || strings.Contains(lowered, "no such browsing context") || strings.Contains(lowered, "context lost") {
		return support.NewContextLostError("元素所属 context 已失效", err)
	}
	return nil
}

func decodeLocatorInfo(info any) (any, []map[string]any, bool) {
	mapped, ok := info.(map[string]any)
	if !ok || len(mapped) == 0 {
		return nil, nil, false
	}
	locator, exists := mapped["locator"]
	if !exists {
		return nil, nil, false
	}
	startNodes := anyToMapSlice(mapped["startNodes"])
	return locator, startNodes, true
}

func rebuildLocatorFromNodeInfo(nodeInfo map[string]any) (any, bool) {
	tag := strings.ToLower(readString(nodeInfo, "localName"))
	attrs := readNestedMap(nodeInfo, "attributes")
	if tag == "" {
		return nil, false
	}
	if id := readString(attrs, "id"); id != "" {
		return "#" + id, true
	}
	if className := readString(attrs, "class"); className != "" {
		parts := strings.Fields(className)
		if len(parts) > 0 {
			return tag + "." + strings.Join(parts, "."), true
		}
	}
	return nil, false
}

func selectorText(locator any) (string, error) {
	switch typed := locator.(type) {
	case string:
		if typed == "" {
			return "*", nil
		}
		return typed, nil
	default:
		parsed, err := support.ParseLocator(locator)
		if err != nil {
			return "", err
		}
		if readString(parsed, "type") != "css" {
			return "", support.NewLocatorError("当前相对父/兄弟定位仅支持 CSS 选择器", nil)
		}
		return stringify(parsed["value"]), nil
	}
}

func normalizePositiveIndex(index int) int {
	if index <= 0 {
		return 1
	}
	return index
}

func writeBytes(path string, data []byte) error {
	absPath, err := filepath.Abs(path)
	if err == nil {
		path = absPath
	}
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0o644)
}

func readString(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	return stringify(values[key])
}

func readNestedMap(values map[string]any, key string) map[string]any {
	if values == nil {
		return nil
	}
	nested, _ := values[key].(map[string]any)
	return cloneAnyMap(nested)
}

func stringify(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case nil:
		return ""
	default:
		return fmt.Sprint(typed)
	}
}

func toBool(value any) bool {
	typed, _ := value.(bool)
	return typed
}

func intFromAny(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int8:
		return int(typed)
	case int16:
		return int(typed)
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case uint:
		return int(typed)
	case uint8:
		return int(typed)
	case uint16:
		return int(typed)
	case uint32:
		return int(typed)
	case uint64:
		return int(typed)
	case float64:
		return int(typed)
	case float32:
		return int(typed)
	default:
		return 0
	}
}

func normalizeXYWH(value any, xKey string, yKey string) map[string]int {
	mapped := asAnyMap(value)
	return map[string]int{
		xKey: intFromAny(mapped[xKey]),
		yKey: intFromAny(mapped[yKey]),
	}
}

func asAnyMap(value any) map[string]any {
	mapped, _ := value.(map[string]any)
	return cloneAnyMap(mapped)
}

func anyMapToStringMap(values map[string]any) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = stringify(value)
	}
	return result
}

func anySliceToMapSlice(value any) []map[string]any {
	switch typed := value.(type) {
	case []map[string]any:
		return cloneAnyMapSlice(typed)
	case []any:
		result := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if mapped, ok := item.(map[string]any); ok {
				result = append(result, cloneAnyMap(mapped))
			}
		}
		return result
	default:
		return []map[string]any{}
	}
}

func anyToMapSlice(value any) []map[string]any {
	switch typed := value.(type) {
	case []map[string]any:
		return cloneAnyMapSlice(typed)
	case []any:
		return anySliceToMapSlice(typed)
	default:
		return nil
	}
}

func anyToStrings(value any) []string {
	switch typed := value.(type) {
	case string:
		return []string{typed}
	case []string:
		return cloneStringSlice(typed)
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			result = append(result, stringify(item))
		}
		return result
	default:
		return []string{stringify(value)}
	}
}

func cloneLocatorInfo(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneAnyMap(typed)
	case []map[string]any:
		return cloneAnyMapSlice(typed)
	case []any:
		cloned := make([]any, len(typed))
		copy(cloned, typed)
		return cloned
	default:
		return value
	}
}

func cloneAnyMap(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}
	cloned := make(map[string]any, len(values))
	for key, value := range values {
		switch typed := value.(type) {
		case map[string]any:
			cloned[key] = cloneAnyMap(typed)
		case []map[string]any:
			cloned[key] = cloneAnyMapSlice(typed)
		case []any:
			items := make([]any, len(typed))
			copy(items, typed)
			cloned[key] = items
		default:
			cloned[key] = value
		}
	}
	return cloned
}

func cloneAnyMapSlice(values []map[string]any) []map[string]any {
	if values == nil {
		return nil
	}
	cloned := make([]map[string]any, len(values))
	for index, value := range values {
		cloned[index] = cloneAnyMap(value)
	}
	return cloned
}

func cloneStringSlice(values []string) []string {
	if values == nil {
		return nil
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}
