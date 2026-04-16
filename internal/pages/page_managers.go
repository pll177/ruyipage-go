package pages

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"ruyipage-go/internal/bidi"
	"ruyipage-go/internal/elements"
	"ruyipage-go/internal/support"
)

const (
	pageManagerPollInterval   = 300 * time.Millisecond
	pageScrollStepCount       = 20
	pageScrollStepPause       = 100 * time.Millisecond
	pageScrollDefaultDistance = 800
)

type PageScroller struct {
	page *FirefoxBase
}

func (s *PageScroller) ToTop() error {
	if s == nil || s.page == nil {
		return nil
	}
	return s.page.scrollUntil(func() bool {
		return s.page.scrollPosition()["y"] <= 0
	}, 0, -pageScrollDefaultDistance, pageScrollStepCount, pageScrollStepPause)
}

func (s *PageScroller) ToBottom() error {
	if s == nil || s.page == nil {
		return nil
	}
	return s.page.scrollUntil(func() bool {
		value, err := s.page.RunJSExpr("window.innerHeight + window.scrollY >= document.documentElement.scrollHeight - 2")
		return err == nil && toPageBool(value)
	}, 0, pageScrollDefaultDistance, pageScrollStepCount, pageScrollStepPause)
}

func (s *PageScroller) ToHalf() error {
	if s == nil || s.page == nil {
		return nil
	}
	pageSize := s.page.pageSize()
	return s.ToLocation(0, pageSize["height"]/2)
}

func (s *PageScroller) ToRightmost() error {
	if s == nil || s.page == nil {
		return nil
	}
	return s.page.scrollUntil(func() bool {
		value, err := s.page.RunJSExpr("window.innerWidth + window.scrollX >= document.documentElement.scrollWidth - 2")
		return err == nil && toPageBool(value)
	}, pageScrollDefaultDistance, 0, pageScrollStepCount, pageScrollStepPause)
}

func (s *PageScroller) ToLeftmost() error {
	if s == nil || s.page == nil {
		return nil
	}
	return s.page.scrollUntil(func() bool {
		return s.page.scrollPosition()["x"] <= 0
	}, -pageScrollDefaultDistance, 0, pageScrollStepCount, pageScrollStepPause)
}

func (s *PageScroller) Down(pixel int) error {
	if s == nil || s.page == nil {
		return nil
	}
	return s.page.performWheelScroll(0, pixel, nil)
}

func (s *PageScroller) Up(pixel int) error {
	if s == nil || s.page == nil {
		return nil
	}
	return s.page.performWheelScroll(0, -pixel, nil)
}

func (s *PageScroller) Right(pixel int) error {
	if s == nil || s.page == nil {
		return nil
	}
	return s.page.performWheelScroll(pixel, 0, nil)
}

func (s *PageScroller) Left(pixel int) error {
	if s == nil || s.page == nil {
		return nil
	}
	return s.page.performWheelScroll(-pixel, 0, nil)
}

func (s *PageScroller) ToSee(target any, center bool) error {
	if s == nil || s.page == nil {
		return nil
	}

	ele, err := s.page.resolveScrollTarget(target)
	if err != nil || ele == nil {
		return err
	}

	if !center {
		if inViewport, inErr := ele.IsInViewport(); inErr == nil && inViewport {
			return nil
		}
	}

	viewportHeight := s.page.viewportSize()["height"]
	if viewportHeight <= 0 {
		viewportHeight = 1
	}

	err = s.page.scrollUntil(func() bool {
		inViewport, inErr := ele.IsInViewport()
		return inErr == nil && inViewport
	}, 0, s.page.scrollDirectionForElement(ele, viewportHeight), pageScrollStepCount, pageScrollStepPause)
	if err != nil {
		return err
	}

	if !center {
		return nil
	}

	viewportMidpoint := s.page.viewportMidpoint()
	eleMidpoint, midErr := ele.ViewportMidpoint()
	if midErr != nil {
		return midErr
	}
	deltaY := eleMidpoint["y"] - viewportMidpoint["y"]
	if deltaY == 0 {
		return nil
	}
	if err := s.page.performWheelScroll(0, deltaY, nil); err != nil {
		return err
	}
	time.Sleep(pageScrollStepPause)
	return nil
}

func (s *PageScroller) ToLocation(x int, y int) error {
	if s == nil || s.page == nil {
		return nil
	}
	position := s.page.scrollPosition()
	return s.page.performWheelScroll(x-position["x"], y-position["y"], nil)
}

type TabRect struct {
	page *FirefoxBase
}

func (r *TabRect) WindowSize() map[string]int {
	if r == nil || r.page == nil {
		return zeroPagePoint("width", "height")
	}
	return r.page.windowSize()
}

func (r *TabRect) ViewportSize() map[string]int {
	if r == nil || r.page == nil {
		return zeroPagePoint("width", "height")
	}
	return r.page.viewportSize()
}

func (r *TabRect) PageSize() map[string]int {
	if r == nil || r.page == nil {
		return zeroPagePoint("width", "height")
	}
	return r.page.pageSize()
}

func (r *TabRect) ScrollPosition() map[string]int {
	if r == nil || r.page == nil {
		return zeroPagePoint("x", "y")
	}
	return r.page.scrollPosition()
}

func (r *TabRect) WindowLocation() map[string]int {
	if r == nil || r.page == nil {
		return zeroPagePoint("x", "y")
	}
	return r.page.windowLocation()
}

func (r *TabRect) ViewportMidpoint() map[string]int {
	if r == nil || r.page == nil {
		return zeroPagePoint("x", "y")
	}
	return r.page.viewportMidpoint()
}

type PageStates struct {
	page *FirefoxBase
}

func (s *PageStates) IsLoaded() bool {
	return strings.EqualFold(s.ReadyState(), "complete")
}

func (s *PageStates) IsAlive() bool {
	if s == nil || s.page == nil {
		return false
	}
	_, err := s.page.RunJSExpr("1")
	return err == nil
}

func (s *PageStates) IsLoading() bool {
	return strings.EqualFold(s.ReadyState(), "loading")
}

func (s *PageStates) ReadyState() string {
	if s == nil || s.page == nil {
		return ""
	}
	state, err := s.page.ReadyState()
	if err != nil {
		return ""
	}
	return state
}

func (s *PageStates) HasAlert() bool {
	if s == nil || s.page == nil {
		return false
	}
	result, err := bidi.GetTree(s.page.browserDriver(), nil, s.page.ContextID(), s.page.baseTimeout())
	if err != nil {
		return false
	}
	contexts := anyToMapSlice(result["contexts"])
	if len(contexts) == 0 {
		return false
	}
	return contexts[0]["userPrompt"] != nil
}

type PageSetter struct {
	page *FirefoxBase
}

func (s *PageSetter) Cookies(cookies any) error {
	if s == nil || s.page == nil {
		return nil
	}
	return s.page.SetCookies(cookies)
}

func (s *PageSetter) UserAgent(ua string) error {
	if s == nil || s.page == nil {
		return nil
	}
	return s.page.Emulation().SetUserAgent(ua, "")
}

func (s *PageSetter) Viewport(width int, height int, devicePixelRatio *float64) error {
	if s == nil || s.page == nil {
		return nil
	}
	return s.page.setViewport(width, height, devicePixelRatio)
}

func (s *PageSetter) Headers(headers any) error {
	if s == nil || s.page == nil {
		return nil
	}
	return s.page.setExtraHeaders(headers)
}

func (s *PageSetter) DownloadPath(path string) error {
	if s == nil || s.page == nil {
		return nil
	}
	return s.page.setDownloadPath(path)
}

func (s *PageSetter) BypassCSP(bypass bool) error {
	if s == nil || s.page == nil {
		return nil
	}
	supported, err := s.page.Emulation().SetBypassCSP(bypass)
	if err != nil {
		return err
	}
	if supported {
		return nil
	}
	return s.page.setBypassCSP(bypass)
}

func (s *PageSetter) ScrollBar(hide bool) error {
	if s == nil || s.page == nil {
		return nil
	}
	return s.page.setScrollBar(hide)
}

type PageWaiter struct {
	page *FirefoxBase
}

func (w *PageWaiter) Sleep(duration time.Duration) {
	if duration <= 0 {
		return
	}
	time.Sleep(duration)
}

func (w *PageWaiter) EleDisplayed(locator any, timeout time.Duration) (*elements.FirefoxElement, error) {
	if w == nil || w.page == nil {
		return nil, nil
	}
	value, _, err := waitPageCondition(w.page, timeout, fmt.Sprintf("等待元素可见: %#v", locator), func() (*elements.FirefoxElement, bool, error) {
		ele, err := w.page.firstFoundElement(locator)
		if err != nil || ele == nil {
			return nil, false, err
		}
		displayed, err := ele.IsDisplayed()
		return ele, err == nil && displayed, err
	})
	return value, err
}

func (w *PageWaiter) EleHidden(locator any, timeout time.Duration) (bool, error) {
	if w == nil || w.page == nil {
		return false, nil
	}
	value, _, err := waitPageCondition(w.page, timeout, fmt.Sprintf("等待元素隐藏: %#v", locator), func() (bool, bool, error) {
		ele, err := w.page.firstFoundElement(locator)
		if err != nil {
			return false, false, err
		}
		if ele == nil {
			return true, true, nil
		}
		displayed, err := ele.IsDisplayed()
		return !displayed, err == nil && !displayed, err
	})
	return value, err
}

func (w *PageWaiter) EleDeleted(locator any, timeout time.Duration) (bool, error) {
	if w == nil || w.page == nil {
		return false, nil
	}
	value, _, err := waitPageCondition(w.page, timeout, fmt.Sprintf("等待元素删除: %#v", locator), func() (bool, bool, error) {
		nodes, err := w.page.findNodes(locator, 100*time.Millisecond, nil)
		if err != nil {
			return false, false, err
		}
		return len(nodes) == 0, len(nodes) == 0, nil
	})
	return value, err
}

func (w *PageWaiter) Ele(locator any, timeout time.Duration) (*elements.FirefoxElement, error) {
	if w == nil || w.page == nil {
		return nil, nil
	}
	value, _, err := waitPageCondition(w.page, timeout, fmt.Sprintf("等待元素出现: %#v", locator), func() (*elements.FirefoxElement, bool, error) {
		ele, err := w.page.firstFoundElement(locator)
		return ele, ele != nil, err
	})
	return value, err
}

func (w *PageWaiter) TitleIs(title string, timeout time.Duration) (bool, error) {
	if w == nil || w.page == nil {
		return false, nil
	}
	value, _, err := waitPageCondition(w.page, timeout, fmt.Sprintf("等待标题为: %s", title), func() (bool, bool, error) {
		current, err := w.page.Title()
		if err != nil {
			return false, false, err
		}
		matched := current == title
		return matched, matched, nil
	})
	return value, err
}

func (w *PageWaiter) TitleContains(fragment string, timeout time.Duration) (bool, error) {
	if w == nil || w.page == nil {
		return false, nil
	}
	value, _, err := waitPageCondition(w.page, timeout, fmt.Sprintf("等待标题包含: %s", fragment), func() (bool, bool, error) {
		title, err := w.page.Title()
		if err != nil {
			return false, false, err
		}
		matched := strings.Contains(title, fragment)
		return matched, matched, nil
	})
	return value, err
}

func (w *PageWaiter) URLContains(fragment string, timeout time.Duration) (bool, error) {
	if w == nil || w.page == nil {
		return false, nil
	}
	value, _, err := waitPageCondition(w.page, timeout, fmt.Sprintf("等待 URL 包含: %s", fragment), func() (bool, bool, error) {
		current, err := w.page.URL()
		if err != nil {
			return false, false, err
		}
		matched := strings.Contains(current, fragment)
		return matched, matched, nil
	})
	return value, err
}

func (w *PageWaiter) URLChange(currentURL string, timeout time.Duration) (bool, error) {
	if w == nil || w.page == nil {
		return false, nil
	}
	if currentURL == "" {
		url, err := w.page.URL()
		if err != nil {
			return false, err
		}
		currentURL = url
	}
	value, _, err := waitPageCondition(w.page, timeout, "等待 URL 变化", func() (bool, bool, error) {
		current, err := w.page.URL()
		if err != nil {
			return false, false, err
		}
		matched := current != currentURL
		return matched, matched, nil
	})
	return value, err
}

func (w *PageWaiter) DocLoaded(timeout time.Duration) (bool, error) {
	if w == nil || w.page == nil {
		return false, nil
	}
	value, _, err := waitPageCondition(w.page, timeout, "等待页面加载", func() (bool, bool, error) {
		state, err := w.page.ReadyState()
		if err != nil {
			return false, false, err
		}
		matched := state == "complete"
		return matched, matched, nil
	})
	return value, err
}

func (w *PageWaiter) LoadStart(timeout time.Duration) (bool, error) {
	if w == nil || w.page == nil {
		return false, nil
	}
	value, _, err := waitPageCondition(w.page, timeout, "等待加载开始", func() (bool, bool, error) {
		state, err := w.page.ReadyState()
		if err != nil {
			return false, false, err
		}
		matched := state == "loading"
		return matched, matched, nil
	})
	return value, err
}

func (w *PageWaiter) JSResult(script string, timeout time.Duration) (any, error) {
	if w == nil || w.page == nil {
		return nil, nil
	}
	value, _, err := waitPageCondition(w.page, timeout, fmt.Sprintf("等待 JS 结果: %s", truncatePageText(script, 30)), func() (any, bool, error) {
		result, err := w.page.RunJS(script)
		if err != nil {
			return nil, false, err
		}
		return result, isTruthyPageValue(result), nil
	})
	return value, err
}

func (w *PageWaiter) ReadyState(target string, timeout time.Duration) error {
	if w == nil || w.page == nil {
		return nil
	}
	return w.page.WaitReadyState(target, timeout)
}

func (w *PageWaiter) LoadComplete(timeout time.Duration) error {
	if w == nil || w.page == nil {
		return nil
	}
	return w.page.WaitLoadComplete(timeout)
}

func (w *PageWaiter) URLContainsLegacy(fragment string, timeout time.Duration) error {
	if w == nil || w.page == nil {
		return nil
	}
	return w.page.WaitURLContains(fragment, timeout)
}

func (w *PageWaiter) TitleContainsLegacy(fragment string, timeout time.Duration) error {
	if w == nil || w.page == nil {
		return nil
	}
	return w.page.WaitTitleContains(fragment, timeout)
}

func waitPageCondition[T any](page *FirefoxBase, timeout time.Duration, message string, condition func() (T, bool, error)) (T, bool, error) {
	timeout = page.resolveManagerWaitTimeout(timeout)
	value, matched, err := support.WaitUntil(condition, timeout, pageManagerPollInterval)
	if err != nil {
		return value, false, err
	}
	if !matched && support.Settings.RaiseWhenWaitFailed {
		var zero T
		return zero, false, support.NewWaitTimeoutError(message, nil)
	}
	return value, matched, nil
}

func (p *FirefoxBase) resolveManagerWaitTimeout(timeout time.Duration) time.Duration {
	if timeout > 0 {
		return timeout
	}
	return p.elementFindTimeout()
}

func (p *FirefoxBase) firstFoundElement(locator any) (*elements.FirefoxElement, error) {
	nodes, err := p.findNodes(locator, 100*time.Millisecond, nil)
	if err != nil || len(nodes) == 0 {
		return nil, err
	}
	return elements.FromNode(p, nodes[0], buildElementLocatorInfo(locator, nil)), nil
}

func (p *FirefoxBase) resolveScrollTarget(target any) (*elements.FirefoxElement, error) {
	switch typed := target.(type) {
	case nil:
		return nil, nil
	case *elements.FirefoxElement:
		return typed, nil
	default:
		return p.FindElement(target, 1, p.elementFindTimeout(), nil)
	}
}

func (p *FirefoxBase) scrollDirectionForElement(ele *elements.FirefoxElement, viewportHeight int) int {
	if ele == nil {
		return pageScrollDefaultDistance
	}
	midpoint, err := ele.ViewportMidpoint()
	if err != nil {
		return pageScrollDefaultDistance
	}
	if midpoint["y"] > viewportHeight {
		return 500
	}
	return -500
}

func (p *FirefoxBase) scrollUntil(check func() bool, stepX int, stepY int, maxSteps int, pause time.Duration) error {
	if check == nil {
		return nil
	}
	for index := 0; index < maxSteps; index++ {
		if check() {
			return nil
		}
		if err := p.performWheelScroll(stepX, stepY, nil); err != nil {
			return err
		}
		time.Sleep(pause)
	}
	_ = check()
	return nil
}

func (p *FirefoxBase) performWheelScroll(deltaX int, deltaY int, origin map[string]int) error {
	if err := p.ensureInitialized(); err != nil {
		return err
	}
	if origin == nil {
		origin = p.viewportMidpoint()
	}
	actions := []map[string]any{
		{
			"type": "wheel",
			"id":   "wheel0",
			"actions": []map[string]any{
				{
					"type":   "scroll",
					"x":      origin["x"],
					"y":      origin["y"],
					"deltaX": deltaX,
					"deltaY": deltaY,
				},
			},
		},
	}
	_, err := bidi.PerformActions(p.browserDriver(), p.ContextID(), actions, p.baseTimeout())
	return err
}

func (p *FirefoxBase) windowSize() map[string]int {
	return p.pageIntMap("({width: window.outerWidth || 0, height: window.outerHeight || 0})", "width", "height")
}

func (p *FirefoxBase) viewportSize() map[string]int {
	return p.pageIntMap("({width: window.innerWidth || 0, height: window.innerHeight || 0})", "width", "height")
}

func (p *FirefoxBase) pageSize() map[string]int {
	return p.pageIntMap(`({
		width: Math.max(document.documentElement ? document.documentElement.scrollWidth : 0, document.body ? document.body.scrollWidth : 0),
		height: Math.max(document.documentElement ? document.documentElement.scrollHeight : 0, document.body ? document.body.scrollHeight : 0)
	})`, "width", "height")
}

func (p *FirefoxBase) scrollPosition() map[string]int {
	return p.pageIntMap("({x: window.scrollX || 0, y: window.scrollY || 0})", "x", "y")
}

func (p *FirefoxBase) windowLocation() map[string]int {
	return p.pageIntMap("({x: window.screenX || 0, y: window.screenY || 0})", "x", "y")
}

func (p *FirefoxBase) viewportMidpoint() map[string]int {
	viewport := p.viewportSize()
	return map[string]int{
		"x": viewport["width"] / 2,
		"y": viewport["height"] / 2,
	}
}

func (p *FirefoxBase) pageIntMap(expression string, xKey string, yKey string) map[string]int {
	if p == nil {
		return zeroPagePoint(xKey, yKey)
	}
	value, err := p.RunJSExpr(expression)
	if err != nil {
		return zeroPagePoint(xKey, yKey)
	}
	mapped, _ := value.(map[string]any)
	return map[string]int{
		xKey: intFromPageValue(mapped[xKey]),
		yKey: intFromPageValue(mapped[yKey]),
	}
}

func (p *FirefoxBase) setViewport(width int, height int, devicePixelRatio *float64) error {
	if err := p.ensureInitialized(); err != nil {
		return err
	}
	returned, err := bidi.SetViewport(p.browserDriver(), p.ContextID(), &width, &height, devicePixelRatio, p.baseTimeout())
	_ = returned
	return err
}

func (p *FirefoxBase) setUserAgent(ua string) error {
	if err := p.ensureInitialized(); err != nil {
		return err
	}

	p.mu.Lock()
	oldID := p.uaPreloadScriptID
	p.uaPreloadScriptID = ""
	p.mu.Unlock()

	if oldID != "" {
		_ = bidi.RemovePreloadScript(p.browserDriver(), oldID, p.baseTimeout())
	}

	injectJS := fmt.Sprintf(`() => {
		Object.defineProperty(navigator, "userAgent", {
			configurable: true,
			get: () => %s
		});
	}`, strconv.Quote(ua))

	result, err := bidi.AddPreloadScript(p.browserDriver(), injectJS, nil, []string{p.ContextID()}, "", p.baseTimeout())
	if err != nil {
		return err
	}

	p.mu.Lock()
	p.uaPreloadScriptID = result.Script
	p.mu.Unlock()

	asExpr := false
	_, _ = p.runJSRaw(injectJS, &asExpr, "", p.baseTimeout())
	return nil
}

func (p *FirefoxBase) setBypassCSP(enabled bool) error {
	if err := p.ensureInitialized(); err != nil {
		return err
	}

	p.mu.Lock()
	oldID := p.cspPreloadScriptID
	p.cspPreloadScriptID = ""
	p.mu.Unlock()

	if oldID != "" {
		_ = bidi.RemovePreloadScript(p.browserDriver(), oldID, p.baseTimeout())
	}
	if !enabled {
		return nil
	}

	injectJS := `() => {
		const removeMeta = () => {
			document.querySelectorAll('meta[http-equiv="Content-Security-Policy"]').forEach((el) => el.remove());
		};
		removeMeta();
		const root = document.documentElement || document;
		if (!root) return;
		const observer = new MutationObserver(() => removeMeta());
		observer.observe(root, {childList: true, subtree: true});
	}`

	result, err := bidi.AddPreloadScript(p.browserDriver(), injectJS, nil, []string{p.ContextID()}, "", p.baseTimeout())
	if err != nil {
		return err
	}

	p.mu.Lock()
	p.cspPreloadScriptID = result.Script
	p.mu.Unlock()

	asExpr := false
	_, _ = p.runJSRaw(injectJS, &asExpr, "", p.baseTimeout())
	return nil
}

func (p *FirefoxBase) setExtraHeaders(headers any) error {
	if err := p.ensureInitialized(); err != nil {
		return err
	}
	normalized, err := normalizePageHeaders(headers)
	if err != nil {
		return err
	}
	_, err = bidi.SetExtraHeaders(p.browserDriver(), normalized, []string{p.ContextID()}, p.baseTimeout())
	return err
}

func (p *FirefoxBase) setDownloadPath(path string) error {
	if err := p.ensureInitialized(); err != nil {
		return err
	}
	return p.Downloads().SetPath(path)
}

func (p *FirefoxBase) setScrollBar(hide bool) error {
	if err := p.ensureInitialized(); err != nil {
		return err
	}
	value := ""
	if hide {
		value = "hidden"
	}
	_, err := p.RunJS(`document.documentElement.style.overflow = arguments[0]`, value)
	return err
}

func normalizePageHeaders(headers any) ([]map[string]any, error) {
	switch typed := headers.(type) {
	case nil:
		return []map[string]any{}, nil
	case []map[string]any:
		return cloneMapSlice(typed), nil
	case map[string]string:
		result := make([]map[string]any, 0, len(typed))
		for name, value := range typed {
			result = append(result, map[string]any{
				"name": name,
				"value": map[string]any{
					"type":  "string",
					"value": value,
				},
			})
		}
		return result, nil
	case map[string]any:
		result := make([]map[string]any, 0, len(typed))
		for name, value := range typed {
			result = append(result, map[string]any{
				"name": name,
				"value": map[string]any{
					"type":  "string",
					"value": stringify(value),
				},
			})
		}
		return result, nil
	default:
		return nil, support.NewRuyiPageError(fmt.Sprintf("headers 参数必须是 map 或 []map，当前为 %T", headers), nil)
	}
}

func intFromPageValue(value any) int {
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
	case float32:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
	}
}

func zeroPagePoint(keys ...string) map[string]int {
	result := make(map[string]int, len(keys))
	for _, key := range keys {
		result[key] = 0
	}
	return result
}

func toPageBool(value any) bool {
	typed, _ := value.(bool)
	return typed
}

func isTruthyPageValue(value any) bool {
	if value == nil {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return typed != ""
	case int:
		return typed != 0
	case int8:
		return typed != 0
	case int16:
		return typed != 0
	case int32:
		return typed != 0
	case int64:
		return typed != 0
	case uint:
		return typed != 0
	case uint8:
		return typed != 0
	case uint16:
		return typed != 0
	case uint32:
		return typed != 0
	case uint64:
		return typed != 0
	case float32:
		return typed != 0
	case float64:
		return typed != 0
	case []any:
		return len(typed) > 0
	case map[string]any:
		return len(typed) > 0
	}

	refValue := reflect.ValueOf(value)
	switch refValue.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return refValue.Len() > 0
	case reflect.Bool:
		return refValue.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return refValue.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return refValue.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return refValue.Float() != 0
	case reflect.Interface, reflect.Pointer:
		return !refValue.IsNil()
	default:
		return true
	}
}

func truncatePageText(text string, maxLength int) string {
	if len(text) <= maxLength || maxLength <= 0 {
		return text
	}
	return text[:maxLength]
}
