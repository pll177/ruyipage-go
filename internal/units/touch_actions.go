package units

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/pll177/ruyipage-go/internal/base"
	"github.com/pll177/ruyipage-go/internal/bidi"
)

type touchActionsOwner interface {
	ContextID() string
	BrowserDriver() *base.BrowserBiDiDriver
	BaseTimeout() time.Duration
	ViewportSize() map[string]int
}

// TouchActions 表示基于 pointerType=touch 的触摸动作链。
type TouchActions struct {
	owner touchActionsOwner

	fingers map[int][]map[string]any
	x       float64
	y       float64
	random  *rand.Rand
}

// NewTouchActions 创建触摸动作链管理器。
func NewTouchActions(owner touchActionsOwner) *TouchActions {
	return &TouchActions{
		owner:   owner,
		fingers: make(map[int][]map[string]any),
		random:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// MoveTo 移动手指到目标位置。
func (t *TouchActions) MoveTo(target any, offsetX int, offsetY int, duration time.Duration, fid int) *TouchActions {
	if t == nil {
		return nil
	}
	x, y := resolveActionPosition(target, t.x, t.y)
	x += float64(offsetX)
	y += float64(offsetY)
	t.x = x
	t.y = y
	t.finger(fid, true)
	t.fingers[fid] = append(t.fingers[fid], map[string]any{
		"type":     "pointerMove",
		"x":        int(x),
		"y":        int(y),
		"duration": actionDurationMS(duration, 50),
	})
	return t
}

// TouchDown 按下触摸点。
func (t *TouchActions) TouchDown(target any, fid int) *TouchActions {
	if t == nil {
		return nil
	}
	if target != nil {
		t.MoveTo(target, 0, 0, 50*time.Millisecond, fid)
	}
	t.finger(fid, true)
	t.fingers[fid] = append(t.fingers[fid], map[string]any{
		"type":   "pointerDown",
		"button": 0,
	})
	return t
}

// TouchUp 抬起触摸点。
func (t *TouchActions) TouchUp(target any, fid int) *TouchActions {
	if t == nil {
		return nil
	}
	if target != nil {
		t.MoveTo(target, 0, 0, 50*time.Millisecond, fid)
	}
	t.finger(fid, true)
	t.fingers[fid] = append(t.fingers[fid], map[string]any{
		"type":   "pointerUp",
		"button": 0,
	})
	return t
}

// Pause 在指定手指动作序列中暂停。
func (t *TouchActions) Pause(duration time.Duration, fid int) *TouchActions {
	if t == nil {
		return nil
	}
	t.finger(fid, true)
	t.fingers[fid] = append(t.fingers[fid], map[string]any{
		"type":     "pause",
		"duration": actionDurationMS(duration, 100),
	})
	return t
}

// Tap 执行轻触。
func (t *TouchActions) Tap(target any, times int) *TouchActions {
	if t == nil {
		return nil
	}
	if target != nil {
		t.MoveTo(target, 0, 0, 50*time.Millisecond, 0)
	}
	if times <= 0 {
		times = 1
	}
	for index := 0; index < times; index++ {
		t.TouchDown(nil, 0).
			Pause(time.Duration(t.random.Intn(61)+60)*time.Millisecond, 0).
			TouchUp(nil, 0)
		if times > 1 && index < times-1 {
			t.Pause(time.Duration(t.random.Intn(71)+80)*time.Millisecond, 0)
		}
	}
	return t
}

// DoubleTap 执行双击。
func (t *TouchActions) DoubleTap(target any) *TouchActions {
	if t == nil {
		return nil
	}
	if target != nil {
		t.MoveTo(target, 0, 0, 50*time.Millisecond, 0)
	}
	return t.TouchDown(nil, 0).
		Pause(60*time.Millisecond, 0).
		TouchUp(nil, 0).
		Pause(100*time.Millisecond, 0).
		TouchDown(nil, 0).
		Pause(60*time.Millisecond, 0).
		TouchUp(nil, 0)
}

// LongPress 执行长按。
func (t *TouchActions) LongPress(target any, duration time.Duration) *TouchActions {
	if t == nil {
		return nil
	}
	if duration <= 0 {
		duration = 800 * time.Millisecond
	}
	if target != nil {
		t.MoveTo(target, 0, 0, 50*time.Millisecond, 0)
	}
	return t.TouchDown(nil, 0).Pause(duration, 0).TouchUp(nil, 0)
}

// Swipe 从起点平滑滑动到终点。
func (t *TouchActions) Swipe(x1 int, y1 int, x2 int, y2 int, duration time.Duration, steps int) *TouchActions {
	if t == nil {
		return nil
	}
	if duration <= 0 {
		duration = 400 * time.Millisecond
	}
	if steps <= 0 {
		steps = int(duration / (16 * time.Millisecond))
		if steps < 10 {
			steps = 10
		}
	}
	stepMS := actionDurationMS(duration, 400) / steps
	if stepMS <= 0 {
		stepMS = 1
	}

	t.finger(0, true)
	t.fingers[0] = append(t.fingers[0],
		map[string]any{"type": "pointerMove", "x": x1, "y": y1, "duration": 0},
		map[string]any{"type": "pointerDown", "button": 0},
	)
	for index := 1; index <= steps; index++ {
		ratio := float64(index) / float64(steps)
		t.fingers[0] = append(t.fingers[0], map[string]any{
			"type":     "pointerMove",
			"x":        int(float64(x1) + float64(x2-x1)*ratio),
			"y":        int(float64(y1) + float64(y2-y1)*ratio),
			"duration": stepMS,
		})
	}
	t.fingers[0] = append(t.fingers[0], map[string]any{"type": "pointerUp", "button": 0})
	t.x = float64(x2)
	t.y = float64(y2)
	return t
}

// SwipeUp 从屏幕下半区域向上滑动。
func (t *TouchActions) SwipeUp(distance int, x *int, duration time.Duration) *TouchActions {
	width, height := t.viewportSize()
	centerX := width / 2
	if x != nil {
		centerX = *x
	}
	startY := int(float64(height) * 0.65)
	return t.Swipe(centerX, startY, centerX, startY-distance, duration, 0)
}

// SwipeDown 从屏幕上半区域向下滑动。
func (t *TouchActions) SwipeDown(distance int, x *int, duration time.Duration) *TouchActions {
	width, height := t.viewportSize()
	centerX := width / 2
	if x != nil {
		centerX = *x
	}
	startY := int(float64(height) * 0.35)
	return t.Swipe(centerX, startY, centerX, startY+distance, duration, 0)
}

// SwipeLeft 从屏幕右侧向左滑动。
func (t *TouchActions) SwipeLeft(distance int, y *int, duration time.Duration) *TouchActions {
	width, height := t.viewportSize()
	centerY := height / 2
	if y != nil {
		centerY = *y
	}
	startX := int(float64(width) * 0.65)
	return t.Swipe(startX, centerY, startX-distance, centerY, duration, 0)
}

// SwipeRight 从屏幕左侧向右滑动。
func (t *TouchActions) SwipeRight(distance int, y *int, duration time.Duration) *TouchActions {
	width, height := t.viewportSize()
	centerY := height / 2
	if y != nil {
		centerY = *y
	}
	startX := int(float64(width) * 0.35)
	return t.Swipe(startX, centerY, startX+distance, centerY, duration, 0)
}

// PinchIn 双指捏合缩小。
func (t *TouchActions) PinchIn(cx *int, cy *int, startGap int, endGap int, duration time.Duration) *TouchActions {
	return t.twoFingerZoom(cx, cy, startGap, endGap, duration)
}

// PinchOut 双指张开放大。
func (t *TouchActions) PinchOut(cx *int, cy *int, startGap int, endGap int, duration time.Duration) *TouchActions {
	return t.twoFingerZoom(cx, cy, startGap, endGap, duration)
}

// Rotate 双指绕中心旋转。
func (t *TouchActions) Rotate(cx *int, cy *int, radius int, startAngle float64, endAngle float64, duration time.Duration) *TouchActions {
	if t == nil {
		return nil
	}
	if radius <= 0 {
		radius = 100
	}
	if duration <= 0 {
		duration = 500 * time.Millisecond
	}
	centerX, centerY := t.resolveCenter(cx, cy)
	steps := int(duration / (16 * time.Millisecond))
	if steps < 10 {
		steps = 10
	}
	stepMS := actionDurationMS(duration, 500) / steps
	if stepMS <= 0 {
		stepMS = 1
	}

	for _, fid := range []int{0, 1} {
		t.finger(fid, true)
		angleOffset := 0.0
		if fid == 1 {
			angleOffset = 180
		}
		radians := (startAngle + angleOffset) * math.Pi / 180
		t.fingers[fid] = append(t.fingers[fid],
			map[string]any{
				"type":     "pointerMove",
				"x":        int(float64(centerX) + float64(radius)*math.Cos(radians)),
				"y":        int(float64(centerY) + float64(radius)*math.Sin(radians)),
				"duration": 0,
			},
			map[string]any{"type": "pointerDown", "button": 0},
		)
	}
	for index := 1; index <= steps; index++ {
		angle := startAngle + (endAngle-startAngle)*float64(index)/float64(steps)
		for _, fid := range []int{0, 1} {
			angleOffset := 0.0
			if fid == 1 {
				angleOffset = 180
			}
			radians := (angle + angleOffset) * math.Pi / 180
			t.fingers[fid] = append(t.fingers[fid], map[string]any{
				"type":     "pointerMove",
				"x":        int(float64(centerX) + float64(radius)*math.Cos(radians)),
				"y":        int(float64(centerY) + float64(radius)*math.Sin(radians)),
				"duration": stepMS,
			})
		}
	}
	for _, fid := range []int{0, 1} {
		t.fingers[fid] = append(t.fingers[fid], map[string]any{"type": "pointerUp", "button": 0})
	}
	return t
}

// Flick 执行快速轻弹。
func (t *TouchActions) Flick(target any, vx int, vy int, duration time.Duration) *TouchActions {
	if t == nil {
		return nil
	}
	if duration <= 0 {
		duration = 150 * time.Millisecond
	}
	if target != nil {
		t.MoveTo(target, 0, 0, 50*time.Millisecond, 0)
	}
	dx := int(float64(vx) * duration.Seconds())
	dy := int(float64(vy) * duration.Seconds())
	return t.Swipe(int(t.x), int(t.y), int(t.x)+dx, int(t.y)+dy, duration, 0)
}

// Perform 执行所有触摸动作。
func (t *TouchActions) Perform() error {
	if t == nil || t.owner == nil {
		return nil
	}
	if len(t.fingers) == 0 {
		return nil
	}

	maxLen := 0
	ids := make([]int, 0, len(t.fingers))
	for fid, actions := range t.fingers {
		ids = append(ids, fid)
		if len(actions) > maxLen {
			maxLen = len(actions)
		}
	}
	sort.Ints(ids)
	for _, fid := range ids {
		for len(t.fingers[fid]) < maxLen {
			t.fingers[fid] = append(t.fingers[fid], map[string]any{"type": "pause", "duration": 0})
		}
	}

	actions := make([]map[string]any, 0, len(ids))
	for _, fid := range ids {
		actions = append(actions, map[string]any{
			"type":       "pointer",
			"id":         "touch" + stringifyActionValue(fid),
			"parameters": map[string]any{"pointerType": "touch"},
			"actions":    cloneActionRows(t.fingers[fid]),
		})
	}
	_, err := bidi.PerformActions(t.owner.BrowserDriver(), t.owner.ContextID(), actions, t.owner.BaseTimeout())
	t.fingers = make(map[int][]map[string]any)
	t.x = 0
	t.y = 0
	return err
}

// ReleaseAll 释放全部触摸点。
func (t *TouchActions) ReleaseAll() error {
	if t == nil || t.owner == nil {
		return nil
	}
	_, err := bidi.ReleaseActions(t.owner.BrowserDriver(), t.owner.ContextID(), t.owner.BaseTimeout())
	t.fingers = make(map[int][]map[string]any)
	return err
}

func (t *TouchActions) finger(fid int, ensure bool) []map[string]any {
	if t.fingers == nil {
		t.fingers = make(map[int][]map[string]any)
	}
	if ensure {
		if _, ok := t.fingers[fid]; !ok {
			t.fingers[fid] = []map[string]any{}
		}
	}
	return t.fingers[fid]
}

func (t *TouchActions) viewportSize() (int, int) {
	if t == nil || t.owner == nil {
		return 0, 0
	}
	size := t.owner.ViewportSize()
	return size["width"], size["height"]
}

func (t *TouchActions) resolveCenter(cx *int, cy *int) (int, int) {
	width, height := t.viewportSize()
	centerX := width / 2
	centerY := height / 2
	if cx != nil {
		centerX = *cx
	}
	if cy != nil {
		centerY = *cy
	}
	return centerX, centerY
}

func (t *TouchActions) twoFingerZoom(cx *int, cy *int, startGap int, endGap int, duration time.Duration) *TouchActions {
	if t == nil {
		return nil
	}
	if startGap <= 0 {
		startGap = 200
	}
	if endGap <= 0 {
		endGap = 60
	}
	if duration <= 0 {
		duration = 400 * time.Millisecond
	}
	centerX, centerY := t.resolveCenter(cx, cy)
	steps := int(duration / (16 * time.Millisecond))
	if steps < 10 {
		steps = 10
	}
	stepMS := actionDurationMS(duration, 400) / steps
	if stepMS <= 0 {
		stepMS = 1
	}

	for _, fid := range []int{0, 1} {
		t.finger(fid, true)
		sign := -1
		if fid == 1 {
			sign = 1
		}
		startX := centerX + sign*(startGap/2)
		t.fingers[fid] = append(t.fingers[fid],
			map[string]any{"type": "pointerMove", "x": startX, "y": centerY, "duration": 0},
			map[string]any{"type": "pointerDown", "button": 0},
		)
	}
	for index := 1; index <= steps; index++ {
		gap := float64(startGap) + float64(endGap-startGap)*float64(index)/float64(steps)
		half := int(gap / 2)
		for _, fid := range []int{0, 1} {
			sign := -1
			if fid == 1 {
				sign = 1
			}
			t.fingers[fid] = append(t.fingers[fid], map[string]any{
				"type":     "pointerMove",
				"x":        centerX + sign*half,
				"y":        centerY,
				"duration": stepMS,
			})
		}
	}
	for _, fid := range []int{0, 1} {
		t.fingers[fid] = append(t.fingers[fid], map[string]any{"type": "pointerUp", "button": 0})
	}
	return t
}
