package units

import (
	stderrors "errors"
	"time"

	"github.com/pll177/ruyipage-go/internal/support"
)

// ClickerElement 表示 Clicker 依赖的最小元素能力。
type ClickerElement interface {
	ClickSelf(byJS bool, timeout time.Duration) error
	RightClick() error
	MiddleClick() error
	DoubleClick() error
	ClickOffset(offsetX int, offsetY int) error
	ClickForNewTab(timeout time.Duration) (string, error)
}

// Clicker 提供元素点击相关的懒加载入口壳层。
type Clicker struct {
	element ClickerElement
}

// NewClicker 创建点击管理器。
func NewClicker(element ClickerElement) *Clicker {
	return &Clicker{element: element}
}

// Left 执行左键点击；times=2 时走双击语义。
func (c *Clicker) Left(times int) error {
	if c == nil || c.element == nil {
		return nil
	}
	switch {
	case times <= 1:
		err := c.element.ClickSelf(false, 0)
		if err == nil {
			return nil
		}
		var noRectErr *support.NoRectError
		var canNotClickErr *support.CanNotClickError
		if stderrors.As(err, &noRectErr) || stderrors.As(err, &canNotClickErr) {
			return c.element.ClickSelf(true, 0)
		}
		return err
	case times == 2:
		return c.element.DoubleClick()
	default:
		for index := 0; index < times; index++ {
			if err := c.element.ClickSelf(false, 0); err != nil {
				return err
			}
		}
		return nil
	}
}

// Right 执行右键点击。
func (c *Clicker) Right() error {
	if c == nil || c.element == nil {
		return nil
	}
	return c.element.RightClick()
}

// Middle 执行中键点击。
func (c *Clicker) Middle() error {
	if c == nil || c.element == nil {
		return nil
	}
	return c.element.MiddleClick()
}

// ByJS 使用 JavaScript 方式点击。
func (c *Clicker) ByJS() error {
	if c == nil || c.element == nil {
		return nil
	}
	return c.element.ClickSelf(true, 0)
}

// At 在元素左上角偏移点执行点击。
func (c *Clicker) At(offsetX int, offsetY int) error {
	if c == nil || c.element == nil {
		return nil
	}
	return c.element.ClickOffset(offsetX, offsetY)
}

// ForNewTab 点击并等待新标签页 context id。
func (c *Clicker) ForNewTab(timeout time.Duration) (string, error) {
	if c == nil || c.element == nil {
		return "", nil
	}
	return c.element.ClickForNewTab(timeout)
}

// ElementScrollerTarget 表示元素滚动管理器依赖的最小能力。
type ElementScrollerTarget interface {
	ScrollToTop() error
	ScrollToBottom() error
	ScrollBy(deltaX int, deltaY int) error
	ScrollToSee(center bool) error
}

// ElementScroller 提供元素滚动相关的懒加载入口壳层。
type ElementScroller struct {
	element ElementScrollerTarget
}

// NewElementScroller 创建元素滚动管理器。
func NewElementScroller(element ElementScrollerTarget) *ElementScroller {
	return &ElementScroller{element: element}
}

// ToTop 滚动到元素内部顶部。
func (s *ElementScroller) ToTop() error {
	if s == nil || s.element == nil {
		return nil
	}
	return s.element.ScrollToTop()
}

// ToBottom 滚动到元素内部底部。
func (s *ElementScroller) ToBottom() error {
	if s == nil || s.element == nil {
		return nil
	}
	return s.element.ScrollToBottom()
}

// Down 向下滚动指定像素。
func (s *ElementScroller) Down(pixel int) error {
	if s == nil || s.element == nil {
		return nil
	}
	return s.element.ScrollBy(0, pixel)
}

// Up 向上滚动指定像素。
func (s *ElementScroller) Up(pixel int) error {
	if s == nil || s.element == nil {
		return nil
	}
	return s.element.ScrollBy(0, -pixel)
}

// Right 向右滚动指定像素。
func (s *ElementScroller) Right(pixel int) error {
	if s == nil || s.element == nil {
		return nil
	}
	return s.element.ScrollBy(pixel, 0)
}

// Left 向左滚动指定像素。
func (s *ElementScroller) Left(pixel int) error {
	if s == nil || s.element == nil {
		return nil
	}
	return s.element.ScrollBy(-pixel, 0)
}

// ToSee 将元素本身滚动到视口内。
func (s *ElementScroller) ToSee(center bool) error {
	if s == nil || s.element == nil {
		return nil
	}
	return s.element.ScrollToSee(center)
}

// ElementRectTarget 表示元素位置尺寸管理器依赖的最小能力。
type ElementRectTarget interface {
	Size() (map[string]int, error)
	Location() (map[string]int, error)
	ViewportLocation() (map[string]int, error)
	ViewportMidpoint() (map[string]int, error)
}

// ElementRect 提供元素矩形信息相关的懒加载入口壳层。
type ElementRect struct {
	element ElementRectTarget
}

// NewElementRect 创建元素矩形管理器。
func NewElementRect(element ElementRectTarget) *ElementRect {
	return &ElementRect{element: element}
}

// Size 返回元素宽高。
func (r *ElementRect) Size() (map[string]int, error) {
	if r == nil || r.element == nil {
		return map[string]int{}, nil
	}
	return r.element.Size()
}

// Location 返回元素文档坐标。
func (r *ElementRect) Location() (map[string]int, error) {
	if r == nil || r.element == nil {
		return map[string]int{}, nil
	}
	return r.element.Location()
}

// Midpoint 返回元素文档坐标中心点。
func (r *ElementRect) Midpoint() (map[string]int, error) {
	if r == nil || r.element == nil {
		return map[string]int{}, nil
	}
	return r.element.ViewportMidpoint()
}

// ClickPoint 返回点击推荐点；当前与 Midpoint 对齐。
func (r *ElementRect) ClickPoint() (map[string]int, error) {
	return r.Midpoint()
}

// ViewportLocation 返回元素视口坐标左上角。
func (r *ElementRect) ViewportLocation() (map[string]int, error) {
	if r == nil || r.element == nil {
		return map[string]int{}, nil
	}
	return r.element.ViewportLocation()
}

// ViewportMidpoint 返回元素视口坐标中心点。
func (r *ElementRect) ViewportMidpoint() (map[string]int, error) {
	if r == nil || r.element == nil {
		return map[string]int{}, nil
	}
	return r.element.ViewportMidpoint()
}

// Corners 返回四个角坐标。
func (r *ElementRect) Corners() (map[string]map[string]int, error) {
	if r == nil || r.element == nil {
		return map[string]map[string]int{}, nil
	}
	location, err := r.element.ViewportLocation()
	if err != nil {
		return nil, err
	}
	size, err := r.element.Size()
	if err != nil {
		return nil, err
	}
	left := location["x"]
	top := location["y"]
	right := left + size["width"]
	bottom := top + size["height"]
	return map[string]map[string]int{
		"top_left":     {"x": left, "y": top},
		"top_right":    {"x": right, "y": top},
		"bottom_left":  {"x": left, "y": bottom},
		"bottom_right": {"x": right, "y": bottom},
	}, nil
}

// ElementSetterTarget 表示元素 setter 依赖的最小能力。
type ElementSetterTarget interface {
	SetAttr(name string, value string) error
	RemoveAttr(name string) error
	SetProp(name string, value any) error
	SetStyle(name string, value string) error
	SetInnerHTML(html string) error
	SetValue(value string) error
}

// ElementSetter 提供元素属性修改相关的懒加载入口壳层。
type ElementSetter struct {
	element ElementSetterTarget
}

// NewElementSetter 创建元素 setter。
func NewElementSetter(element ElementSetterTarget) *ElementSetter {
	return &ElementSetter{element: element}
}

func (s *ElementSetter) Attr(name string, value string) error {
	if s == nil || s.element == nil {
		return nil
	}
	return s.element.SetAttr(name, value)
}

func (s *ElementSetter) RemoveAttr(name string) error {
	if s == nil || s.element == nil {
		return nil
	}
	return s.element.RemoveAttr(name)
}

func (s *ElementSetter) Prop(name string, value any) error {
	if s == nil || s.element == nil {
		return nil
	}
	return s.element.SetProp(name, value)
}

func (s *ElementSetter) Style(name string, value string) error {
	if s == nil || s.element == nil {
		return nil
	}
	return s.element.SetStyle(name, value)
}

func (s *ElementSetter) InnerHTML(html string) error {
	if s == nil || s.element == nil {
		return nil
	}
	return s.element.SetInnerHTML(html)
}

func (s *ElementSetter) Value(value string) error {
	if s == nil || s.element == nil {
		return nil
	}
	return s.element.SetValue(value)
}

// ElementStatesTarget 表示元素状态管理器依赖的最小能力。
type ElementStatesTarget interface {
	IsDisplayed() (bool, error)
	IsEnabled() (bool, error)
	IsChecked() (bool, error)
	IsSelected() (bool, error)
	IsInViewport() (bool, error)
	HasRect() (bool, error)
}

// ElementStates 提供元素状态查询相关的懒加载入口壳层。
type ElementStates struct {
	element ElementStatesTarget
}

// NewElementStates 创建元素状态管理器。
func NewElementStates(element ElementStatesTarget) *ElementStates {
	return &ElementStates{element: element}
}

func (s *ElementStates) IsDisplayed() (bool, error) {
	if s == nil || s.element == nil {
		return false, nil
	}
	return s.element.IsDisplayed()
}

func (s *ElementStates) IsEnabled() (bool, error) {
	if s == nil || s.element == nil {
		return false, nil
	}
	return s.element.IsEnabled()
}

func (s *ElementStates) IsChecked() (bool, error) {
	if s == nil || s.element == nil {
		return false, nil
	}
	return s.element.IsChecked()
}

func (s *ElementStates) IsSelected() (bool, error) {
	if s == nil || s.element == nil {
		return false, nil
	}
	return s.element.IsSelected()
}

func (s *ElementStates) IsInViewport() (bool, error) {
	if s == nil || s.element == nil {
		return false, nil
	}
	return s.element.IsInViewport()
}

func (s *ElementStates) HasRect() (bool, error) {
	if s == nil || s.element == nil {
		return false, nil
	}
	return s.element.HasRect()
}

// ElementWaiterTarget 表示元素等待管理器依赖的最小能力。
type ElementWaiterTarget interface {
	WaitDisplayed(timeout time.Duration) (bool, error)
	WaitHidden(timeout time.Duration) (bool, error)
	WaitEnabled(timeout time.Duration) (bool, error)
	WaitDisabled(timeout time.Duration) (bool, error)
}

// ElementWaiter 提供元素等待相关的懒加载入口壳层。
type ElementWaiter struct {
	element ElementWaiterTarget
}

// NewElementWaiter 创建元素等待管理器。
func NewElementWaiter(element ElementWaiterTarget) *ElementWaiter {
	return &ElementWaiter{element: element}
}

func (w *ElementWaiter) Sleep(duration time.Duration) {
	if duration <= 0 {
		return
	}
	time.Sleep(duration)
}

func (w *ElementWaiter) Displayed(timeout time.Duration) (bool, error) {
	if w == nil || w.element == nil {
		return false, nil
	}
	return w.element.WaitDisplayed(timeout)
}

func (w *ElementWaiter) Hidden(timeout time.Duration) (bool, error) {
	if w == nil || w.element == nil {
		return false, nil
	}
	return w.element.WaitHidden(timeout)
}

func (w *ElementWaiter) Enabled(timeout time.Duration) (bool, error) {
	if w == nil || w.element == nil {
		return false, nil
	}
	return w.element.WaitEnabled(timeout)
}

func (w *ElementWaiter) Disabled(timeout time.Duration) (bool, error) {
	if w == nil || w.element == nil {
		return false, nil
	}
	return w.element.WaitDisabled(timeout)
}

// SelectElementTarget 表示 select 管理器依赖的最小能力。
type SelectElementTarget interface {
	SelectByText(text string, timeout time.Duration, mode string) (bool, error)
	SelectByValue(value string, mode string) (bool, error)
	SelectByIndex(index int, mode string) (bool, error)
	SelectCancelByIndex(index int) (bool, error)
	SelectCancelByText(text string) (bool, error)
	SelectAllOptions() error
	DeselectAllOptions() error
	SelectOptions() ([]map[string]any, error)
	SelectSelectedOption() (map[string]any, error)
	SelectIsMulti() (bool, error)
}

// SelectElement 提供 select 元素相关的懒加载入口壳层。
type SelectElement struct {
	element SelectElementTarget
}

// NewSelectElement 创建 select 管理器。
func NewSelectElement(element SelectElementTarget) *SelectElement {
	return &SelectElement{element: element}
}

func (s *SelectElement) ByText(text string, timeout time.Duration, mode string) (bool, error) {
	if s == nil || s.element == nil {
		return false, nil
	}
	return s.element.SelectByText(text, timeout, mode)
}

func (s *SelectElement) ByValue(value string, mode string) (bool, error) {
	if s == nil || s.element == nil {
		return false, nil
	}
	return s.element.SelectByValue(value, mode)
}

func (s *SelectElement) ByIndex(index int, mode string) (bool, error) {
	if s == nil || s.element == nil {
		return false, nil
	}
	return s.element.SelectByIndex(index, mode)
}

func (s *SelectElement) CancelByIndex(index int) (bool, error) {
	if s == nil || s.element == nil {
		return false, nil
	}
	return s.element.SelectCancelByIndex(index)
}

func (s *SelectElement) CancelByText(text string) (bool, error) {
	if s == nil || s.element == nil {
		return false, nil
	}
	return s.element.SelectCancelByText(text)
}

func (s *SelectElement) SelectAll() error {
	if s == nil || s.element == nil {
		return nil
	}
	return s.element.SelectAllOptions()
}

func (s *SelectElement) DeselectAll() error {
	if s == nil || s.element == nil {
		return nil
	}
	return s.element.DeselectAllOptions()
}

func (s *SelectElement) Options() ([]map[string]any, error) {
	if s == nil || s.element == nil {
		return []map[string]any{}, nil
	}
	return s.element.SelectOptions()
}

func (s *SelectElement) SelectedOption() (map[string]any, error) {
	if s == nil || s.element == nil {
		return map[string]any{}, nil
	}
	return s.element.SelectSelectedOption()
}

func (s *SelectElement) IsMulti() (bool, error) {
	if s == nil || s.element == nil {
		return false, nil
	}
	return s.element.SelectIsMulti()
}
