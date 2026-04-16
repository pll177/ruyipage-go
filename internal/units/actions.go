package units

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/pll177/ruyipage-go/internal/base"
	"github.com/pll177/ruyipage-go/internal/bidi"
)

type actionsOwner interface {
	ContextID() string
	BrowserDriver() *base.BrowserBiDiDriver
	BaseTimeout() time.Duration
}

type viewportPointTarget interface {
	ViewportMidpoint() (map[string]int, error)
}

type scrollableActionTarget interface {
	ScrollToSee(center bool) error
}

// Actions 表示页面级鼠标、键盘、滚轮动作链。
type Actions struct {
	owner actionsOwner

	pointerActions []map[string]any
	keyActions     []map[string]any
	wheelActions   []map[string]any

	currentX float64
	currentY float64
	random   *rand.Rand
}

// NewActions 创建动作链管理器。
func NewActions(owner actionsOwner) *Actions {
	return &Actions{
		owner:  owner,
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// MoveTo 移动鼠标到目标位置。
func (a *Actions) MoveTo(target any, offsetX int, offsetY int, duration time.Duration, origin any) *Actions {
	if a == nil {
		return nil
	}
	x, y := resolveActionPosition(target, a.currentX, a.currentY)
	x += float64(offsetX)
	y += float64(offsetY)
	action := map[string]any{
		"type":     "pointerMove",
		"x":        int(x),
		"y":        int(y),
		"duration": actionDurationMS(duration, 100),
	}
	if normalized, include := normalizeActionOrigin(origin); include {
		action["origin"] = normalized
	}
	a.pointerActions = append(a.pointerActions, action)
	a.currentX = x
	a.currentY = y
	return a
}

// Move 基于当前位置移动鼠标。
func (a *Actions) Move(offsetX int, offsetY int, duration time.Duration) *Actions {
	if a == nil {
		return nil
	}
	a.currentX += float64(offsetX)
	a.currentY += float64(offsetY)
	a.pointerActions = append(a.pointerActions, map[string]any{
		"type":     "pointerMove",
		"x":        int(a.currentX),
		"y":        int(a.currentY),
		"duration": actionDurationMS(duration, 100),
	})
	return a
}

// Click 执行左键点击。
func (a *Actions) Click(on any, times int) *Actions {
	if a == nil {
		return nil
	}
	if on != nil {
		a.MoveTo(on, 0, 0, 100*time.Millisecond, nil)
	}
	if times <= 0 {
		times = 1
	}
	for index := 0; index < times; index++ {
		a.pointerActions = append(a.pointerActions,
			map[string]any{"type": "pointerDown", "button": 0},
			map[string]any{"type": "pause", "duration": 50},
			map[string]any{"type": "pointerUp", "button": 0},
		)
	}
	return a
}

// DoubleClick 执行双击。
func (a *Actions) DoubleClick(on any) *Actions {
	return a.Click(on, 2)
}

// MiddleClick 执行中键点击。
func (a *Actions) MiddleClick(on any) *Actions {
	return a.clickByButton(on, 1)
}

// RightClick 执行右键点击。
func (a *Actions) RightClick(on any) *Actions {
	return a.clickByButton(on, 2)
}

// DBClick 保留旧命名双击别名。
func (a *Actions) DBClick(on any) *Actions {
	return a.DoubleClick(on)
}

// RClick 保留旧命名右键别名。
func (a *Actions) RClick(on any) *Actions {
	return a.RightClick(on)
}

// Hold 按住鼠标按钮。
func (a *Actions) Hold(on any, button int) *Actions {
	if a == nil {
		return nil
	}
	if on != nil {
		a.MoveTo(on, 0, 0, 100*time.Millisecond, nil)
	}
	a.pointerActions = append(a.pointerActions, map[string]any{
		"type":   "pointerDown",
		"button": button,
	})
	return a
}

// Release 释放鼠标按钮。
func (a *Actions) Release(on any, button int) *Actions {
	if a == nil {
		return nil
	}
	if on != nil {
		a.MoveTo(on, 0, 0, 100*time.Millisecond, nil)
	}
	a.pointerActions = append(a.pointerActions, map[string]any{
		"type":   "pointerUp",
		"button": button,
	})
	return a
}

// DragTo 从 source 拖拽到 target。
func (a *Actions) DragTo(source any, target any, duration time.Duration, steps int) *Actions {
	if a == nil {
		return nil
	}
	if steps <= 0 {
		steps = 20
	}
	if duration <= 0 {
		duration = 500 * time.Millisecond
	}
	startX, startY := resolveActionPosition(source, a.currentX, a.currentY)
	endX, endY := resolveActionPosition(target, startX, startY)
	stepMS := actionDurationMS(duration, 500) / steps
	if stepMS <= 0 {
		stepMS = 1
	}

	a.pointerActions = append(a.pointerActions,
		map[string]any{
			"type":     "pointerMove",
			"origin":   "viewport",
			"x":        int(startX),
			"y":        int(startY),
			"duration": 0,
		},
		map[string]any{"type": "pointerDown", "button": 0},
		map[string]any{"type": "pause", "duration": 120},
	)

	for index := 1; index <= steps; index++ {
		ratio := float64(index) / float64(steps)
		a.pointerActions = append(a.pointerActions, map[string]any{
			"type":     "pointerMove",
			"origin":   "viewport",
			"x":        int(startX + (endX-startX)*ratio),
			"y":        int(startY + (endY-startY)*ratio),
			"duration": stepMS,
		})
	}

	a.pointerActions = append(a.pointerActions,
		map[string]any{"type": "pause", "duration": 120},
		map[string]any{"type": "pointerUp", "button": 0},
	)
	a.currentX = endX
	a.currentY = endY
	return a
}

// Drag 是 DragTo 的别名。
func (a *Actions) Drag(source any, target any, duration time.Duration, steps int) *Actions {
	return a.DragTo(source, target, duration, steps)
}

// KeyDown 按下按键但不释放。
func (a *Actions) KeyDown(key string) *Actions {
	if a == nil {
		return nil
	}
	a.keyActions = append(a.keyActions, map[string]any{"type": "keyDown", "value": key})
	return a
}

// KeyUp 释放按键。
func (a *Actions) KeyUp(key string) *Actions {
	if a == nil {
		return nil
	}
	a.keyActions = append(a.keyActions, map[string]any{"type": "keyUp", "value": key})
	return a
}

// Combo 执行组合键。
func (a *Actions) Combo(keys ...string) *Actions {
	if a == nil {
		return nil
	}
	for _, key := range keys {
		a.keyActions = append(a.keyActions, map[string]any{"type": "keyDown", "value": key})
	}
	for index := len(keys) - 1; index >= 0; index-- {
		a.keyActions = append(a.keyActions, map[string]any{"type": "keyUp", "value": keys[index]})
	}
	return a
}

// Type 逐字符输入文本。
func (a *Actions) Type(text any, interval time.Duration) *Actions {
	if a == nil {
		return nil
	}
	intervalMS := actionDurationMS(interval, 0)
	for _, ch := range []rune(stringifyActionValue(text)) {
		a.keyActions = append(a.keyActions, map[string]any{"type": "keyDown", "value": string(ch)})
		if intervalMS > 0 {
			a.keyActions = append(a.keyActions, map[string]any{"type": "pause", "duration": intervalMS})
		}
		a.keyActions = append(a.keyActions, map[string]any{"type": "keyUp", "value": string(ch)})
	}
	return a
}

// Press 按下并释放单个按键。
func (a *Actions) Press(key string) *Actions {
	if a == nil {
		return nil
	}
	a.keyActions = append(a.keyActions,
		map[string]any{"type": "keyDown", "value": key},
		map[string]any{"type": "keyUp", "value": key},
	)
	return a
}

// Scroll 执行滚轮滚动。
func (a *Actions) Scroll(deltaX int, deltaY int, on any, origin any) *Actions {
	if a == nil {
		return nil
	}
	x := a.currentX
	y := a.currentY
	if on != nil {
		x, y = resolveActionPosition(on, x, y)
	}
	action := map[string]any{
		"type":   "scroll",
		"x":      int(x),
		"y":      int(y),
		"deltaX": deltaX,
		"deltaY": deltaY,
	}
	if normalized, include := normalizeActionOrigin(origin); include {
		action["origin"] = normalized
	}
	a.wheelActions = append(a.wheelActions, action)
	return a
}

// Wait 在动作链中插入暂停。
func (a *Actions) Wait(duration time.Duration) *Actions {
	if a == nil {
		return nil
	}
	ms := actionDurationMS(duration, 0)
	a.pointerActions = append(a.pointerActions, map[string]any{"type": "pause", "duration": ms})
	a.keyActions = append(a.keyActions, map[string]any{"type": "pause", "duration": ms})
	return a
}

// Perform 执行累积的动作。
func (a *Actions) Perform() error {
	if a == nil || a.owner == nil {
		return nil
	}

	actions := make([]map[string]any, 0, 3)
	if len(a.pointerActions) > 0 {
		actions = append(actions, map[string]any{
			"type":       "pointer",
			"id":         "mouse0",
			"parameters": map[string]any{"pointerType": "mouse"},
			"actions":    cloneActionRows(a.pointerActions),
		})
	}
	if len(a.keyActions) > 0 {
		actions = append(actions, map[string]any{
			"type":    "key",
			"id":      "keyboard0",
			"actions": cloneActionRows(a.keyActions),
		})
	}
	if len(a.wheelActions) > 0 {
		actions = append(actions, map[string]any{
			"type":    "wheel",
			"id":      "wheel0",
			"actions": cloneActionRows(a.wheelActions),
		})
	}
	if len(actions) == 0 {
		return nil
	}

	_, err := bidi.PerformActions(a.owner.BrowserDriver(), a.owner.ContextID(), actions, a.owner.BaseTimeout())
	a.pointerActions = nil
	a.keyActions = nil
	a.wheelActions = nil
	return err
}

// ReleaseAll 释放当前上下文中所有按住的输入源状态。
func (a *Actions) ReleaseAll() error {
	if a == nil || a.owner == nil {
		return nil
	}
	_, err := bidi.ReleaseActions(a.owner.BrowserDriver(), a.owner.ContextID(), a.owner.BaseTimeout())
	return err
}

// HumanMove 执行拟人化移动。
func (a *Actions) HumanMove(target any, style string) *Actions {
	if a == nil {
		return nil
	}
	if scrollable, ok := target.(scrollableActionTarget); ok {
		_ = scrollable.ScrollToSee(true)
	}

	endX, endY := resolveActionPosition(target, a.currentX, a.currentY)
	path := bidi.BuildHumanMousePath(
		bidi.MousePoint{X: a.currentX, Y: a.currentY},
		bidi.MousePoint{X: endX, Y: endY},
	)

	for _, point := range path {
		a.pointerActions = append(a.pointerActions, map[string]any{
			"type":     "pointerMove",
			"x":        int(point.X),
			"y":        int(point.Y),
			"duration": a.random.Intn(13) + 8,
		})
	}
	for index := 0; index < a.random.Intn(3)+2; index++ {
		a.pointerActions = append(a.pointerActions, map[string]any{
			"type":     "pointerMove",
			"x":        int(endX) + a.random.Intn(5) - 2,
			"y":        int(endY) + a.random.Intn(3) - 1,
			"duration": a.random.Intn(31) + 20,
		})
	}
	a.pointerActions = append(a.pointerActions, map[string]any{
		"type":     "pointerMove",
		"x":        int(endX),
		"y":        int(endY),
		"duration": a.random.Intn(16) + 15,
	})
	a.currentX = endX
	a.currentY = endY
	_ = style
	return a
}

// HumanClick 执行拟人化点击。
func (a *Actions) HumanClick(on any, button string) *Actions {
	if a == nil {
		return nil
	}
	if on != nil {
		a.HumanMove(on, "")
	}
	a.Wait(time.Duration(a.random.Intn(101)+50) * time.Millisecond)
	buttonID := 0
	switch button {
	case "middle":
		buttonID = 1
	case "right":
		buttonID = 2
	}
	a.pointerActions = append(a.pointerActions,
		map[string]any{"type": "pointerDown", "button": buttonID},
		map[string]any{"type": "pause", "duration": a.random.Intn(51) + 40},
		map[string]any{"type": "pointerUp", "button": buttonID},
	)
	return a
}

// HumanType 执行拟人化输入。
func (a *Actions) HumanType(text any, minDelay time.Duration, maxDelay time.Duration) *Actions {
	if a == nil {
		return nil
	}
	if minDelay <= 0 {
		minDelay = 45 * time.Millisecond
	}
	if maxDelay <= 0 {
		maxDelay = 240 * time.Millisecond
	}
	if maxDelay < minDelay {
		maxDelay = minDelay
	}
	minMS := actionDurationMS(minDelay, 45)
	maxMS := actionDurationMS(maxDelay, 240)
	for _, ch := range []rune(stringifyActionValue(text)) {
		a.keyActions = append(a.keyActions, map[string]any{"type": "keyDown", "value": string(ch)})
		delay := minMS
		if maxMS > minMS {
			delay += a.random.Intn(maxMS - minMS + 1)
		}
		if delay > 0 {
			a.keyActions = append(a.keyActions, map[string]any{"type": "pause", "duration": delay})
		}
		a.keyActions = append(a.keyActions, map[string]any{"type": "keyUp", "value": string(ch)})
	}
	return a
}

func (a *Actions) clickByButton(on any, button int) *Actions {
	if a == nil {
		return nil
	}
	if on != nil {
		a.MoveTo(on, 0, 0, 100*time.Millisecond, nil)
	}
	a.pointerActions = append(a.pointerActions,
		map[string]any{"type": "pointerDown", "button": button},
		map[string]any{"type": "pause", "duration": 50},
		map[string]any{"type": "pointerUp", "button": button},
	)
	return a
}

func resolveActionPosition(target any, currentX float64, currentY float64) (float64, float64) {
	switch typed := target.(type) {
	case nil:
		return currentX, currentY
	case map[string]int:
		return float64(typed["x"]), float64(typed["y"])
	case map[string]any:
		return float64(intFromActionAny(typed["x"])), float64(intFromActionAny(typed["y"]))
	case []int:
		if len(typed) >= 2 {
			return float64(typed[0]), float64(typed[1])
		}
	case []any:
		if len(typed) >= 2 {
			return float64(intFromActionAny(typed[0])), float64(intFromActionAny(typed[1]))
		}
	case [2]int:
		return float64(typed[0]), float64(typed[1])
	case viewportPointTarget:
		if point, err := typed.ViewportMidpoint(); err == nil {
			return float64(point["x"]), float64(point["y"])
		}
	}
	return currentX, currentY
}

func actionDurationMS(duration time.Duration, defaultMS int) int {
	if duration <= 0 {
		return defaultMS
	}
	return int(duration / time.Millisecond)
}

func normalizeActionOrigin(origin any) (any, bool) {
	switch typed := origin.(type) {
	case nil:
		return nil, false
	case string:
		if typed == "" || typed == "viewport" {
			return nil, false
		}
		return typed, true
	default:
		return typed, true
	}
}

func cloneActionRows(rows []map[string]any) []map[string]any {
	if len(rows) == 0 {
		return nil
	}
	cloned := make([]map[string]any, len(rows))
	for index, row := range rows {
		cloned[index] = cloneActionRow(row)
	}
	return cloned
}

func cloneActionRow(row map[string]any) map[string]any {
	if row == nil {
		return nil
	}
	cloned := make(map[string]any, len(row))
	for key, value := range row {
		if nested, ok := value.(map[string]any); ok {
			cloned[key] = cloneActionRow(nested)
			continue
		}
		cloned[key] = value
	}
	return cloned
}

func stringifyActionValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case nil:
		return ""
	default:
		return fmt.Sprint(typed)
	}
}

func intFromActionAny(value any) int {
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
