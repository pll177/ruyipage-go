package ruyipage

import "time"

// PageScroller 是页面级滚动管理器。
type PageScroller struct {
	owner *FirefoxBase
}

// TabRect 是页面/标签页/Frame 位置尺寸管理器。
type TabRect struct {
	owner *FirefoxBase
}

// PageSetter 是页面级 setter 管理器。
type PageSetter struct {
	owner *FirefoxBase
}

// PageStates 是页面级状态管理器。
type PageStates struct {
	owner *FirefoxBase
}

// PageWaiter 是页面级等待管理器。
type PageWaiter struct {
	owner *FirefoxBase
}

// Clicker 是元素点击管理器。
type Clicker struct {
	owner *FirefoxElement
}

// ElementScroller 是元素滚动管理器。
type ElementScroller struct {
	owner *FirefoxElement
}

// ElementRect 是元素矩形管理器。
type ElementRect struct {
	owner *FirefoxElement
}

// ElementSetter 是元素 setter 管理器。
type ElementSetter struct {
	owner *FirefoxElement
}

// ElementStates 是元素状态管理器。
type ElementStates struct {
	owner *FirefoxElement
}

// ElementWaiter 是元素等待管理器。
type ElementWaiter struct {
	owner *FirefoxElement
}

// SelectElement 是 <select> 管理器。
type SelectElement struct {
	owner *FirefoxElement
}

func (p *FirefoxBase) Scroll() *PageScroller {
	if p == nil || p.inner == nil {
		return nil
	}
	p.managersMu.Lock()
	defer p.managersMu.Unlock()
	if p.scroll == nil {
		p.scroll = &PageScroller{owner: p}
	}
	return p.scroll
}

func (p *FirefoxBase) Rect() *TabRect {
	if p == nil || p.inner == nil {
		return nil
	}
	p.managersMu.Lock()
	defer p.managersMu.Unlock()
	if p.rect == nil {
		p.rect = &TabRect{owner: p}
	}
	return p.rect
}

func (p *FirefoxBase) Set() *PageSetter {
	if p == nil || p.inner == nil {
		return nil
	}
	p.managersMu.Lock()
	defer p.managersMu.Unlock()
	if p.setter == nil {
		p.setter = &PageSetter{owner: p}
	}
	return p.setter
}

func (p *FirefoxBase) States() *PageStates {
	if p == nil || p.inner == nil {
		return nil
	}
	p.managersMu.Lock()
	defer p.managersMu.Unlock()
	if p.states == nil {
		p.states = &PageStates{owner: p}
	}
	return p.states
}

func (p *FirefoxBase) Wait() *PageWaiter {
	if p == nil || p.inner == nil {
		return nil
	}
	p.managersMu.Lock()
	defer p.managersMu.Unlock()
	if p.waiter == nil {
		p.waiter = &PageWaiter{owner: p}
	}
	return p.waiter
}

func (e *FirefoxElement) Click() *Clicker {
	if e == nil || e.inner == nil {
		return nil
	}
	e.managersMu.Lock()
	defer e.managersMu.Unlock()
	if e.clicker == nil {
		e.clicker = &Clicker{owner: e}
	}
	return e.clicker
}

func (e *FirefoxElement) Scroll() *ElementScroller {
	if e == nil || e.inner == nil {
		return nil
	}
	e.managersMu.Lock()
	defer e.managersMu.Unlock()
	if e.scroll == nil {
		e.scroll = &ElementScroller{owner: e}
	}
	return e.scroll
}

func (e *FirefoxElement) Rect() *ElementRect {
	if e == nil || e.inner == nil {
		return nil
	}
	e.managersMu.Lock()
	defer e.managersMu.Unlock()
	if e.rect == nil {
		e.rect = &ElementRect{owner: e}
	}
	return e.rect
}

func (e *FirefoxElement) Set() *ElementSetter {
	if e == nil || e.inner == nil {
		return nil
	}
	e.managersMu.Lock()
	defer e.managersMu.Unlock()
	if e.setter == nil {
		e.setter = &ElementSetter{owner: e}
	}
	return e.setter
}

func (e *FirefoxElement) States() *ElementStates {
	if e == nil || e.inner == nil {
		return nil
	}
	e.managersMu.Lock()
	defer e.managersMu.Unlock()
	if e.states == nil {
		e.states = &ElementStates{owner: e}
	}
	return e.states
}

func (e *FirefoxElement) Wait() *ElementWaiter {
	if e == nil || e.inner == nil {
		return nil
	}
	e.managersMu.Lock()
	defer e.managersMu.Unlock()
	if e.waiter == nil {
		e.waiter = &ElementWaiter{owner: e}
	}
	return e.waiter
}

func (e *FirefoxElement) Select() *SelectElement {
	if e == nil || e.inner == nil {
		return nil
	}
	e.managersMu.Lock()
	defer e.managersMu.Unlock()
	if e.selects == nil {
		e.selects = &SelectElement{owner: e}
	}
	return e.selects
}

func (s *PageScroller) ToTop() error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Scroll().ToTop()
}

func (s *PageScroller) ToBottom() error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Scroll().ToBottom()
}

func (s *PageScroller) ToHalf() error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Scroll().ToHalf()
}

func (s *PageScroller) ToRightmost() error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Scroll().ToRightmost()
}

func (s *PageScroller) ToLeftmost() error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Scroll().ToLeftmost()
}

func (s *PageScroller) Down(pixel int) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Scroll().Down(pixel)
}

func (s *PageScroller) Up(pixel int) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Scroll().Up(pixel)
}

func (s *PageScroller) Right(pixel int) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Scroll().Right(pixel)
}

func (s *PageScroller) Left(pixel int) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Scroll().Left(pixel)
}

func (s *PageScroller) ToSee(target any, center bool) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	if element, ok := target.(*FirefoxElement); ok {
		target = element.inner
	}
	return s.owner.inner.Scroll().ToSee(target, center)
}

func (s *PageScroller) ToLocation(x int, y int) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Scroll().ToLocation(x, y)
}

func (r *TabRect) WindowSize() map[string]int {
	if r == nil || r.owner == nil || r.owner.inner == nil {
		return zeroManagerPoint("width", "height")
	}
	return r.owner.inner.Rect().WindowSize()
}

func (r *TabRect) ViewportSize() map[string]int {
	if r == nil || r.owner == nil || r.owner.inner == nil {
		return zeroManagerPoint("width", "height")
	}
	return r.owner.inner.Rect().ViewportSize()
}

func (r *TabRect) PageSize() map[string]int {
	if r == nil || r.owner == nil || r.owner.inner == nil {
		return zeroManagerPoint("width", "height")
	}
	return r.owner.inner.Rect().PageSize()
}

func (r *TabRect) ScrollPosition() map[string]int {
	if r == nil || r.owner == nil || r.owner.inner == nil {
		return zeroManagerPoint("x", "y")
	}
	return r.owner.inner.Rect().ScrollPosition()
}

func (r *TabRect) WindowLocation() map[string]int {
	if r == nil || r.owner == nil || r.owner.inner == nil {
		return zeroManagerPoint("x", "y")
	}
	return r.owner.inner.Rect().WindowLocation()
}

func (r *TabRect) ViewportMidpoint() map[string]int {
	if r == nil || r.owner == nil || r.owner.inner == nil {
		return zeroManagerPoint("x", "y")
	}
	return r.owner.inner.Rect().ViewportMidpoint()
}

func (s *PageSetter) Cookies(cookies any) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Set().Cookies(cookies)
}

func (s *PageSetter) UserAgent(ua string) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Set().UserAgent(ua)
}

func (s *PageSetter) Viewport(width int, height int, devicePixelRatio *float64) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Set().Viewport(width, height, devicePixelRatio)
}

func (s *PageSetter) Headers(headers any) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Set().Headers(headers)
}

func (s *PageSetter) DownloadPath(path string) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Set().DownloadPath(path)
}

func (s *PageSetter) BypassCSP(bypass bool) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Set().BypassCSP(bypass)
}

func (s *PageSetter) ScrollBar(hide bool) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Set().ScrollBar(hide)
}

func (s *PageStates) IsLoaded() bool {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return false
	}
	return s.owner.inner.States().IsLoaded()
}

func (s *PageStates) IsAlive() bool {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return false
	}
	return s.owner.inner.States().IsAlive()
}

func (s *PageStates) IsLoading() bool {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return false
	}
	return s.owner.inner.States().IsLoading()
}

func (s *PageStates) ReadyState() string {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return ""
	}
	return s.owner.inner.States().ReadyState()
}

func (s *PageStates) HasAlert() bool {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return false
	}
	return s.owner.inner.States().HasAlert()
}

func (w *PageWaiter) Sleep(duration time.Duration) {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return
	}
	w.owner.inner.Wait().Sleep(duration)
}

func (w *PageWaiter) EleDisplayed(locator any, timeout time.Duration) (*FirefoxElement, error) {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return nil, nil
	}
	element, err := w.owner.inner.Wait().EleDisplayed(locator, timeout)
	if err != nil || element == nil {
		return nil, err
	}
	return newFirefoxElementFromInner(element, w.owner.page), nil
}

func (w *PageWaiter) EleHidden(locator any, timeout time.Duration) (bool, error) {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return false, nil
	}
	return w.owner.inner.Wait().EleHidden(locator, timeout)
}

func (w *PageWaiter) EleDeleted(locator any, timeout time.Duration) (bool, error) {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return false, nil
	}
	return w.owner.inner.Wait().EleDeleted(locator, timeout)
}

func (w *PageWaiter) Ele(locator any, timeout time.Duration) (*FirefoxElement, error) {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return nil, nil
	}
	element, err := w.owner.inner.Wait().Ele(locator, timeout)
	if err != nil || element == nil {
		return nil, err
	}
	return newFirefoxElementFromInner(element, w.owner.page), nil
}

func (w *PageWaiter) TitleIs(title string, timeout time.Duration) (bool, error) {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return false, nil
	}
	return w.owner.inner.Wait().TitleIs(title, timeout)
}

func (w *PageWaiter) TitleContains(fragment string, timeout time.Duration) (bool, error) {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return false, nil
	}
	return w.owner.inner.Wait().TitleContains(fragment, timeout)
}

func (w *PageWaiter) URLContains(fragment string, timeout time.Duration) (bool, error) {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return false, nil
	}
	return w.owner.inner.Wait().URLContains(fragment, timeout)
}

func (w *PageWaiter) URLChange(currentURL string, timeout time.Duration) (bool, error) {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return false, nil
	}
	return w.owner.inner.Wait().URLChange(currentURL, timeout)
}

func (w *PageWaiter) DocLoaded(timeout time.Duration) (bool, error) {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return false, nil
	}
	return w.owner.inner.Wait().DocLoaded(timeout)
}

func (w *PageWaiter) LoadStart(timeout time.Duration) (bool, error) {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return false, nil
	}
	return w.owner.inner.Wait().LoadStart(timeout)
}

func (w *PageWaiter) JSResult(script string, timeout time.Duration) (any, error) {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return nil, nil
	}
	return w.owner.inner.Wait().JSResult(script, timeout)
}

func (w *PageWaiter) ReadyState(target string, timeout time.Duration) error {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return nil
	}
	return w.owner.inner.Wait().ReadyState(target, timeout)
}

func (w *PageWaiter) LoadComplete(timeout time.Duration) error {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return nil
	}
	return w.owner.inner.Wait().LoadComplete(timeout)
}

func (c *Clicker) Left(times int) error {
	if c == nil || c.owner == nil || c.owner.inner == nil {
		return nil
	}
	return c.owner.inner.Click().Left(times)
}

func (c *Clicker) Right() error {
	if c == nil || c.owner == nil || c.owner.inner == nil {
		return nil
	}
	return c.owner.inner.Click().Right()
}

func (c *Clicker) Middle() error {
	if c == nil || c.owner == nil || c.owner.inner == nil {
		return nil
	}
	return c.owner.inner.Click().Middle()
}

func (c *Clicker) ByJS() error {
	if c == nil || c.owner == nil || c.owner.inner == nil {
		return nil
	}
	return c.owner.inner.Click().ByJS()
}

func (c *Clicker) At(offsetX int, offsetY int) error {
	if c == nil || c.owner == nil || c.owner.inner == nil {
		return nil
	}
	return c.owner.inner.Click().At(offsetX, offsetY)
}

func (c *Clicker) ForNewTab(timeout time.Duration) (*FirefoxTab, error) {
	if c == nil || c.owner == nil || c.owner.inner == nil {
		return nil, nil
	}
	contextID, err := c.owner.inner.Click().ForNewTab(timeout)
	if err != nil || contextID == "" || c.owner.page == nil {
		return nil, err
	}
	return c.owner.page.GetTab(contextID, "", "")
}

func (s *ElementScroller) ToTop() error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Scroll().ToTop()
}

func (s *ElementScroller) ToBottom() error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Scroll().ToBottom()
}

func (s *ElementScroller) Down(pixel int) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Scroll().Down(pixel)
}

func (s *ElementScroller) Up(pixel int) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Scroll().Up(pixel)
}

func (s *ElementScroller) Right(pixel int) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Scroll().Right(pixel)
}

func (s *ElementScroller) Left(pixel int) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Scroll().Left(pixel)
}

func (s *ElementScroller) ToSee(center bool) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Scroll().ToSee(center)
}

func (r *ElementRect) Size() map[string]int {
	if r == nil || r.owner == nil || r.owner.inner == nil {
		return zeroManagerPoint("width", "height")
	}
	value, err := r.owner.inner.Rect().Size()
	if err != nil {
		return zeroManagerPoint("width", "height")
	}
	return value
}

func (r *ElementRect) Location() map[string]int {
	if r == nil || r.owner == nil || r.owner.inner == nil {
		return zeroManagerPoint("x", "y")
	}
	value, err := r.owner.inner.Rect().Location()
	if err != nil {
		return zeroManagerPoint("x", "y")
	}
	return value
}

func (r *ElementRect) Midpoint() map[string]int {
	if r == nil || r.owner == nil || r.owner.inner == nil {
		return zeroManagerPoint("x", "y")
	}
	value, err := r.owner.inner.Rect().Midpoint()
	if err != nil {
		return zeroManagerPoint("x", "y")
	}
	return value
}

func (r *ElementRect) ClickPoint() map[string]int {
	if r == nil || r.owner == nil || r.owner.inner == nil {
		return zeroManagerPoint("x", "y")
	}
	value, err := r.owner.inner.Rect().ClickPoint()
	if err != nil {
		return zeroManagerPoint("x", "y")
	}
	return value
}

func (r *ElementRect) ViewportLocation() map[string]int {
	if r == nil || r.owner == nil || r.owner.inner == nil {
		return zeroManagerPoint("x", "y")
	}
	value, err := r.owner.inner.Rect().ViewportLocation()
	if err != nil {
		return zeroManagerPoint("x", "y")
	}
	return value
}

func (r *ElementRect) ViewportMidpoint() map[string]int {
	if r == nil || r.owner == nil || r.owner.inner == nil {
		return zeroManagerPoint("x", "y")
	}
	value, err := r.owner.inner.Rect().ViewportMidpoint()
	if err != nil {
		return zeroManagerPoint("x", "y")
	}
	return value
}

func (r *ElementRect) Corners() []map[string]int {
	location := r.ViewportLocation()
	size := r.Size()
	return []map[string]int{
		{"x": location["x"], "y": location["y"]},
		{"x": location["x"] + size["width"], "y": location["y"]},
		{"x": location["x"] + size["width"], "y": location["y"] + size["height"]},
		{"x": location["x"], "y": location["y"] + size["height"]},
	}
}

func (s *ElementSetter) Attr(name string, value string) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Set().Attr(name, value)
}

func (s *ElementSetter) RemoveAttr(name string) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Set().RemoveAttr(name)
}

func (s *ElementSetter) Prop(name string, value any) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Set().Prop(name, value)
}

func (s *ElementSetter) Style(name string, value string) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Set().Style(name, value)
}

func (s *ElementSetter) InnerHTML(html string) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Set().InnerHTML(html)
}

func (s *ElementSetter) Value(value string) error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Set().Value(value)
}

func (s *ElementStates) IsDisplayed() bool {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return false
	}
	value, err := s.owner.inner.States().IsDisplayed()
	return err == nil && value
}

func (s *ElementStates) IsEnabled() bool {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return false
	}
	value, err := s.owner.inner.States().IsEnabled()
	return err == nil && value
}

func (s *ElementStates) IsChecked() bool {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return false
	}
	value, err := s.owner.inner.States().IsChecked()
	return err == nil && value
}

func (s *ElementStates) IsSelected() bool {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return false
	}
	value, err := s.owner.inner.States().IsSelected()
	return err == nil && value
}

func (s *ElementStates) IsInViewport() bool {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return false
	}
	value, err := s.owner.inner.States().IsInViewport()
	return err == nil && value
}

func (s *ElementStates) HasRect() bool {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return false
	}
	value, err := s.owner.inner.States().HasRect()
	return err == nil && value
}

func (w *ElementWaiter) Sleep(duration time.Duration) {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return
	}
	w.owner.inner.Wait().Sleep(duration)
}

func (w *ElementWaiter) Displayed(timeout time.Duration) (*FirefoxElement, error) {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return nil, nil
	}
	matched, err := w.owner.inner.Wait().Displayed(timeout)
	if err != nil || !matched {
		return nil, err
	}
	return w.owner, nil
}

func (w *ElementWaiter) Hidden(timeout time.Duration) (bool, error) {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return false, nil
	}
	return w.owner.inner.Wait().Hidden(timeout)
}

func (w *ElementWaiter) Enabled(timeout time.Duration) (*FirefoxElement, error) {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return nil, nil
	}
	matched, err := w.owner.inner.Wait().Enabled(timeout)
	if err != nil || !matched {
		return nil, err
	}
	return w.owner, nil
}

func (w *ElementWaiter) Disabled(timeout time.Duration) (bool, error) {
	if w == nil || w.owner == nil || w.owner.inner == nil {
		return false, nil
	}
	return w.owner.inner.Wait().Disabled(timeout)
}

func (s *SelectElement) ByText(text string, timeout time.Duration, mode string) (bool, error) {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return false, nil
	}
	return s.owner.inner.Select().ByText(text, timeout, mode)
}

func (s *SelectElement) ByValue(value string, mode string) (bool, error) {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return false, nil
	}
	return s.owner.inner.Select().ByValue(value, mode)
}

func (s *SelectElement) ByIndex(index int, mode string) (bool, error) {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return false, nil
	}
	return s.owner.inner.Select().ByIndex(index, mode)
}

func (s *SelectElement) CancelByIndex(index int) (bool, error) {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return false, nil
	}
	return s.owner.inner.Select().CancelByIndex(index)
}

func (s *SelectElement) CancelByText(text string) (bool, error) {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return false, nil
	}
	return s.owner.inner.Select().CancelByText(text)
}

func (s *SelectElement) SelectAll() error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Select().SelectAll()
}

func (s *SelectElement) DeselectAll() error {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return nil
	}
	return s.owner.inner.Select().DeselectAll()
}

func (s *SelectElement) Options() []map[string]any {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return []map[string]any{}
	}
	value, err := s.owner.inner.Select().Options()
	if err != nil {
		return []map[string]any{}
	}
	return value
}

func (s *SelectElement) SelectedOption() map[string]any {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return map[string]any{}
	}
	value, err := s.owner.inner.Select().SelectedOption()
	if err != nil || value == nil {
		return map[string]any{}
	}
	return value
}

func (s *SelectElement) IsMulti() bool {
	if s == nil || s.owner == nil || s.owner.inner == nil {
		return false
	}
	value, err := s.owner.inner.Select().IsMulti()
	return err == nil && value
}

func zeroManagerPoint(keys ...string) map[string]int {
	result := make(map[string]int, len(keys))
	for _, key := range keys {
		result[key] = 0
	}
	return result
}
