package bidi

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/pll177/ruyipage-go/internal/support"
)

// MousePoint 表示鼠标轨迹中的二维点。
type MousePoint struct {
	X float64
	Y float64
}

// KeyChord 表示一个修饰键组合。
type KeyChord struct {
	Modifier string
	Key      string
}

// PenActionOptions 表示 pen pointer 动作的可选参数。
type PenActionOptions struct {
	Pressure           float64
	TiltX              int
	TiltY              int
	Twist              int
	TangentialPressure float64
	Button             int
	Duration           int
	AltitudeAngle      *float64
	AzimuthAngle       *float64
	Width              *int
	Height             *int
}

// NewPenActionOptions 返回与 Python 版对齐的默认 pen 动作参数。
func NewPenActionOptions() PenActionOptions {
	return PenActionOptions{
		Pressure: 0.5,
		Duration: 50,
	}
}

// WheelActionOptions 表示 wheel scroll 动作的可选参数。
type WheelActionOptions struct {
	DeltaX    int
	DeltaY    int
	DeltaZ    int
	DeltaMode int
	Duration  int
	Origin    any
}

// NewWheelActionOptions 返回与 Python 版对齐的默认 wheel 动作参数。
func NewWheelActionOptions() WheelActionOptions {
	return WheelActionOptions{
		DeltaY: 120,
		Origin: "viewport",
	}
}

type inputCommandDriver interface {
	Run(method string, params map[string]any, timeout time.Duration) (map[string]any, error)
}

type inputRandomSource interface {
	Float64() float64
	Intn(n int) int
	NormFloat64() float64
}

type humanMouseBuildOptions struct {
	random         inputRandomSource
	styleOverride  string
	jitterOverride *bool
}

// PerformActions 执行 input.performActions。
func PerformActions(
	driver inputCommandDriver,
	context string,
	actions []map[string]any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"context": context,
		"actions": cloneAnyMapSliceDeep(actions),
	}
	return runInputCommand(driver, "input.performActions", params, timeout)
}

// ReleaseActions 执行 input.releaseActions。
func ReleaseActions(driver inputCommandDriver, context string, timeout time.Duration) (map[string]any, error) {
	return runInputCommand(driver, "input.releaseActions", map[string]any{"context": context}, timeout)
}

// SetFiles 执行 input.setFiles。
func SetFiles(
	driver inputCommandDriver,
	context string,
	element map[string]any,
	files []string,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"context": context,
		"element": cloneAnyMapDeep(element),
		"files":   cloneStringSliceExact(files),
	}
	return runInputCommand(driver, "input.setFiles", params, timeout)
}

// BuildHumanMousePath 生成拟人鼠标轨迹点列表。
func BuildHumanMousePath(start MousePoint, end MousePoint) []MousePoint {
	return buildHumanMousePathWithOptions(start, end, humanMouseBuildOptions{})
}

// BuildHumanClickActions 构建完整的拟人点击 actions。
func BuildHumanClickActions(target MousePoint, start *MousePoint) []map[string]any {
	return buildHumanClickActionsWithOptions(target, start, humanMouseBuildOptions{})
}

// BuildPenAction 构建 pen pointer 动作。
//
// 传入 nil options 时使用 Python 版默认值；
// 若需在默认值上修改，请从 NewPenActionOptions() 返回值开始调整。
func BuildPenAction(x int, y int, options *PenActionOptions) []map[string]any {
	resolved := NewPenActionOptions()
	if options != nil {
		resolved = *options
	}

	moveAction := map[string]any{
		"type":               "pointerMove",
		"x":                  x,
		"y":                  y,
		"duration":           resolved.Duration,
		"pressure":           resolved.Pressure,
		"tiltX":              resolved.TiltX,
		"tiltY":              resolved.TiltY,
		"twist":              resolved.Twist,
		"tangentialPressure": resolved.TangentialPressure,
	}
	downAction := map[string]any{
		"type":     "pointerDown",
		"button":   resolved.Button,
		"pressure": resolved.Pressure,
		"tiltX":    resolved.TiltX,
		"tiltY":    resolved.TiltY,
	}

	if resolved.AltitudeAngle != nil {
		moveAction["altitudeAngle"] = *resolved.AltitudeAngle
		downAction["altitudeAngle"] = *resolved.AltitudeAngle
	}
	if resolved.AzimuthAngle != nil {
		moveAction["azimuthAngle"] = *resolved.AzimuthAngle
		downAction["azimuthAngle"] = *resolved.AzimuthAngle
	}
	if resolved.Width != nil {
		moveAction["width"] = *resolved.Width
	}
	if resolved.Height != nil {
		moveAction["height"] = *resolved.Height
	}

	return []map[string]any{
		{
			"type":       "pointer",
			"id":         "pen0",
			"parameters": map[string]any{"pointerType": "pen"},
			"actions": []map[string]any{
				moveAction,
				downAction,
				{"type": "pointerUp", "button": resolved.Button},
			},
		},
	}
}

// BuildKeyAction 构建键盘动作序列。
//
// 支持：
//   - string：逐字符输入
//   - []string：逐项按键
//   - []KeyChord：组合键序列
//   - []any：元素可为 string / KeyChord / [2]string / []string(len=2)
func BuildKeyAction(keys any) ([]map[string]any, error) {
	actions, err := buildKeyActions(keys)
	if err != nil {
		return nil, err
	}

	return []map[string]any{
		{
			"type":    "key",
			"id":      "kbd0",
			"actions": actions,
		},
	}, nil
}

// BuildWheelAction 构建 wheel scroll 动作。
//
// 传入 nil options 时使用 Python 版默认值；
// 若需在默认值上修改，请从 NewWheelActionOptions() 返回值开始调整。
func BuildWheelAction(x int, y int, options *WheelActionOptions) []map[string]any {
	resolved := NewWheelActionOptions()
	if options != nil {
		resolved = *options
		if resolved.Origin == nil {
			resolved.Origin = "viewport"
		}
	}

	action := map[string]any{
		"type":   "scroll",
		"x":      x,
		"y":      y,
		"deltaX": resolved.DeltaX,
		"deltaY": resolved.DeltaY,
	}
	if resolved.DeltaZ != 0 {
		action["deltaZ"] = resolved.DeltaZ
	}
	if resolved.DeltaMode != 0 {
		action["deltaMode"] = resolved.DeltaMode
	}
	if resolved.Duration != 0 {
		action["duration"] = resolved.Duration
	}
	if origin, include := normalizeWheelOrigin(resolved.Origin); include {
		action["origin"] = origin
	}

	return []map[string]any{
		{
			"type": "wheel",
			"id":   "wheel0",
			"actions": []map[string]any{
				action,
			},
		},
	}
}

func runInputCommand(
	driver inputCommandDriver,
	method string,
	params map[string]any,
	timeout time.Duration,
) (map[string]any, error) {
	if driver == nil {
		return nil, support.NewPageDisconnectedError("input driver 未初始化", nil)
	}

	result, err := driver.Run(method, params, timeout)
	if err != nil {
		return nil, err
	}
	return cloneAnyMapDeep(result), nil
}

func buildHumanMousePathWithOptions(start MousePoint, end MousePoint, options humanMouseBuildOptions) []MousePoint {
	random := resolveInputRandomSource(options.random)

	dx := end.X - start.X
	dy := end.Y - start.Y
	dist := math.Hypot(dx, dy)
	if dist == 0 {
		dist = 1.0
	}

	steps := int(maxFloat(12, minFloat(52, math.Round(dist/randFloatRange(random, 10, 22)))))
	oversample := randIntRange(random, 3, 4)
	curvature := randFloatRange(random, 0.55, 0.82)
	style := chooseHumanMouseStyle(random, options.styleOverride)

	var raw []MousePoint
	switch style {
	case "line_then_arc":
		ratio := randFloatRange(random, 0.45, 0.75)
		mid := lerpPoint(start, end, ratio)
		raw = concatMousePaths(
			linePath(start, mid, maxInt(2, int(float64(steps)*ratio)), oversample, nil),
			arcPath(mid, end, maxInt(2, steps-int(float64(steps)*ratio)), curvature, oversample, MousePoint{}, false, random),
		)
	case "line":
		raw = linePath(start, end, steps, oversample, nil)
	case "arc":
		raw = arcPath(start, end, steps, curvature, oversample, MousePoint{}, false, random)
	default:
		overshoot := overshootPoint(start, end, random)
		control := ctrlArc(overshoot, end, 0.9, random)
		raw = concatMousePaths(
			linePath(start, overshoot, maxInt(2, int(float64(steps)*0.55)), oversample, nil),
			arcPath(overshoot, end, maxInt(2, steps-int(float64(steps)*0.55)), curvature, oversample, control, true, random),
		)
	}

	maxNorm := minFloat(7.5, maxFloat(2.5, dist*randFloatRange(random, 0.006, 0.011)))
	maxTan := minFloat(4.0, maxFloat(1.2, dist*randFloatRange(random, 0.003, 0.008)))
	if shouldApplyJitter(random, options.jitterOverride) {
		return applyJitter(raw, maxNorm, maxTan, 6, 6, random)
	}
	return raw
}

func buildHumanClickActionsWithOptions(target MousePoint, start *MousePoint, options humanMouseBuildOptions) []map[string]any {
	random := resolveInputRandomSource(options.random)

	resolvedStart := MousePoint{
		X: float64(randIntRange(random, 100, 900)),
		Y: float64(randIntRange(random, 100, 600)),
	}
	if start != nil {
		resolvedStart = *start
	}

	pathOptions := options
	pathOptions.random = random
	path := buildHumanMousePathWithOptions(resolvedStart, target, pathOptions)

	actions := []map[string]any{
		{
			"type":     "pointerMove",
			"x":        int(resolvedStart.X),
			"y":        int(resolvedStart.Y),
			"duration": 0,
		},
	}

	prevX := resolvedStart.X
	prevY := resolvedStart.Y
	for _, point := range path {
		bx := int(point.X)
		by := int(point.Y)
		dist := math.Hypot(float64(bx)-prevX, float64(by)-prevY)
		actions = append(actions, map[string]any{
			"type":     "pointerMove",
			"x":        bx,
			"y":        by,
			"duration": maxInt(8, int(dist*randFloatRange(random, 1.5, 3.0))),
		})
		prevX = float64(bx)
		prevY = float64(by)
	}

	targetX := int(target.X)
	targetY := int(target.Y)
	for index := 0; index < randIntRange(random, 2, 4); index++ {
		actions = append(actions, map[string]any{
			"type":     "pointerMove",
			"x":        targetX + randIntRange(random, -2, 2),
			"y":        targetY + randIntRange(random, -1, 1),
			"duration": randIntRange(random, 20, 50),
		})
	}

	actions = append(actions,
		map[string]any{
			"type":     "pointerMove",
			"x":        targetX,
			"y":        targetY,
			"duration": randIntRange(random, 15, 30),
		},
		map[string]any{
			"type":     "pause",
			"duration": randIntRange(random, 80, 300),
		},
		map[string]any{
			"type":   "pointerDown",
			"button": 0,
		},
		map[string]any{
			"type":     "pause",
			"duration": randIntRange(random, 80, 180),
		},
		map[string]any{
			"type":   "pointerUp",
			"button": 0,
		},
		map[string]any{
			"type":     "pointerMove",
			"x":        targetX + randIntRange(random, 5, 20),
			"y":        targetY + randIntRange(random, -5, 5),
			"duration": randIntRange(random, 80, 150),
		},
	)

	return []map[string]any{
		{
			"type":       "pointer",
			"id":         "mouse0",
			"parameters": map[string]any{"pointerType": "mouse"},
			"actions":    actions,
		},
	}
}

func buildKeyActions(keys any) ([]map[string]any, error) {
	var actions []map[string]any

	switch typed := keys.(type) {
	case string:
		actions = make([]map[string]any, 0, len([]rune(typed))*2)
		for _, ch := range typed {
			actions = appendSimpleKeyActions(actions, string(ch))
		}
		return actions, nil
	case []string:
		actions = make([]map[string]any, 0, len(typed)*2)
		for _, item := range typed {
			actions = appendSimpleKeyActions(actions, item)
		}
		return actions, nil
	case []KeyChord:
		actions = make([]map[string]any, 0, len(typed)*4)
		for _, item := range typed {
			actions = appendChordKeyActions(actions, item.Modifier, item.Key)
		}
		return actions, nil
	case []any:
		actions = make([]map[string]any, 0, len(typed)*4)
		for index, item := range typed {
			var err error
			actions, err = appendDynamicKeyActions(actions, item)
			if err != nil {
				return nil, fmt.Errorf("keys[%d] %w", index, err)
			}
		}
		return actions, nil
	default:
		return nil, fmt.Errorf("keys 参数必须为 string、[]string、[]any 或 []KeyChord")
	}
}

func appendDynamicKeyActions(actions []map[string]any, item any) ([]map[string]any, error) {
	switch typed := item.(type) {
	case string:
		return appendSimpleKeyActions(actions, typed), nil
	case KeyChord:
		return appendChordKeyActions(actions, typed.Modifier, typed.Key), nil
	case *KeyChord:
		if typed == nil {
			return nil, fmt.Errorf("参数必须为 string、KeyChord 或长度为 2 的 []string/[2]string")
		}
		return appendChordKeyActions(actions, typed.Modifier, typed.Key), nil
	case [2]string:
		return appendChordKeyActions(actions, typed[0], typed[1]), nil
	case []string:
		if len(typed) != 2 {
			return nil, fmt.Errorf("参数必须为 string、KeyChord 或长度为 2 的 []string/[2]string")
		}
		return appendChordKeyActions(actions, typed[0], typed[1]), nil
	default:
		return nil, fmt.Errorf("参数必须为 string、KeyChord 或长度为 2 的 []string/[2]string")
	}
}

func appendSimpleKeyActions(actions []map[string]any, value string) []map[string]any {
	return append(actions,
		map[string]any{"type": "keyDown", "value": value},
		map[string]any{"type": "keyUp", "value": value},
	)
}

func appendChordKeyActions(actions []map[string]any, modifier string, key string) []map[string]any {
	return append(actions,
		map[string]any{"type": "keyDown", "value": modifier},
		map[string]any{"type": "keyDown", "value": key},
		map[string]any{"type": "keyUp", "value": key},
		map[string]any{"type": "keyUp", "value": modifier},
	)
}

func normalizeWheelOrigin(origin any) (any, bool) {
	if origin == nil {
		return nil, false
	}

	if text, ok := origin.(string); ok {
		if text == "" || text == "viewport" {
			return nil, false
		}
		return text, true
	}

	return cloneAnyValueDeep(origin), true
}

func resolveInputRandomSource(random inputRandomSource) inputRandomSource {
	if random != nil {
		return random
	}
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func chooseHumanMouseStyle(random inputRandomSource, override string) string {
	if override != "" {
		return override
	}

	styles := []string{"line_then_arc", "line", "arc", "line_overshoot_arc_back"}
	weights := []float64{0.40, 0.22, 0.28, 0.10}
	roll := random.Float64()
	accumulated := 0.0
	for index, style := range styles {
		accumulated += weights[index]
		if roll <= accumulated {
			return style
		}
	}
	return styles[len(styles)-1]
}

func shouldApplyJitter(random inputRandomSource, override *bool) bool {
	if override != nil {
		return *override
	}
	return random.Float64() < 0.75
}

func easeOutCubic(t float64) float64 {
	return 1 - math.Pow(1-t, 3)
}

func easeInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return 1 - math.Pow(-2*t+2, 2)/2
}

func lerp(a float64, b float64, t float64) float64 {
	return a + (b-a)*t
}

func lerpPoint(start MousePoint, end MousePoint, t float64) MousePoint {
	return MousePoint{
		X: lerp(start.X, end.X, t),
		Y: lerp(start.Y, end.Y, t),
	}
}

func bezierQuadratic(start MousePoint, control MousePoint, end MousePoint, t float64) MousePoint {
	u := 1 - t
	return MousePoint{
		X: u*u*start.X + 2*u*t*control.X + t*t*end.X,
		Y: u*u*start.Y + 2*u*t*control.Y + t*t*end.Y,
	}
}

func ctrlArc(start MousePoint, end MousePoint, curvature float64, random inputRandomSource) MousePoint {
	dx := end.X - start.X
	dy := end.Y - start.Y
	dist := math.Hypot(dx, dy)
	if dist == 0 {
		dist = 1.0
	}

	mx := (start.X + end.X) * 0.5
	my := (start.Y + end.Y) * 0.5
	nx := -dy / dist
	ny := dx / dist
	side := 1.0
	if random.Intn(2) == 0 {
		side = -1.0
	}

	offset := maxFloat(60.0, minFloat(dist*curvature+randFloatRange(random, -0.12, 0.12)*dist*curvature, 520.0))
	return MousePoint{
		X: mx + nx*offset*side,
		Y: my + ny*offset*side,
	}
}

func arcPath(
	start MousePoint,
	end MousePoint,
	steps int,
	curvature float64,
	oversample int,
	control MousePoint,
	hasControl bool,
	random inputRandomSource,
) []MousePoint {
	if !hasControl {
		control = ctrlArc(start, end, curvature, random)
	}

	total := maxInt(steps*oversample, steps)
	path := make([]MousePoint, total)
	for index := 1; index <= total; index++ {
		path[index-1] = bezierQuadratic(start, control, end, easeOutCubic(float64(index)/float64(total)))
	}
	return path
}

func linePath(start MousePoint, end MousePoint, steps int, oversample int, ease func(float64) float64) []MousePoint {
	if ease == nil {
		ease = easeInOutQuad
	}

	total := maxInt(steps*oversample, steps)
	path := make([]MousePoint, total)
	for index := 1; index <= total; index++ {
		path[index-1] = lerpPoint(start, end, ease(float64(index)/float64(total)))
	}
	return path
}

func smoothSeries(length int, sigma float64, smoothK int, random inputRandomSource) []float64 {
	value := 0.0
	raw := make([]float64, 0, length)
	for index := 0; index < length; index++ {
		value += random.NormFloat64() * sigma
		raw = append(raw, value)
	}
	if smoothK <= 1 {
		return raw
	}

	window := maxInt(1, smoothK)
	accumulated := 0.0
	for index := 0; index < window; index++ {
		accumulated += raw[index]
	}

	smoothed := []float64{accumulated / float64(window)}
	for index := window; index < length; index++ {
		accumulated += raw[index] - raw[index-window]
		smoothed = append(smoothed, accumulated/float64(window))
	}

	prefixLength := length - len(smoothed)
	output := make([]float64, 0, length)
	for index := 0; index < prefixLength; index++ {
		output = append(output, smoothed[0])
	}
	output = append(output, smoothed...)
	return output
}

func applyJitter(
	path []MousePoint,
	maxNorm float64,
	maxTan float64,
	keepEnd int,
	keepStart int,
	random inputRandomSource,
) []MousePoint {
	count := len(path)
	if count < 3 {
		return path
	}

	type tangentBasis struct {
		tx float64
		ty float64
		nx float64
		ny float64
	}

	tangents := make([]tangentBasis, 0, count)
	for index := 0; index < count; index++ {
		var dx float64
		var dy float64
		switch {
		case index == 0:
			dx = path[1].X - path[0].X
			dy = path[1].Y - path[0].Y
		case index == count-1:
			dx = path[index].X - path[index-1].X
			dy = path[index].Y - path[index-1].Y
		default:
			dx = path[index+1].X - path[index-1].X
			dy = path[index+1].Y - path[index-1].Y
		}

		length := math.Hypot(dx, dy)
		if length == 0 {
			length = 1.0
		}
		tx := dx / length
		ty := dy / length
		tangents = append(tangents, tangentBasis{
			tx: tx,
			ty: ty,
			nx: -ty,
			ny: tx,
		})
	}

	tanNoise := smoothSeries(count, 0.55, maxInt(5, count/30), random)
	normNoise := smoothSeries(count, 0.9, maxInt(6, count/28), random)
	output := make([]MousePoint, 0, count)

	for index, point := range path {
		t := float64(index) / float64(count-1)
		edge := (0.5 - math.Abs(t-0.5)) / 0.5

		weight := 1.0
		if keepStart > 0 && index < keepStart {
			weight = float64(index) / float64(keepStart)
		} else if keepEnd > 0 && index > count-keepEnd-1 {
			weight = float64(count-1-index) / float64(keepEnd)
		}

		weight = maxFloat(0.0, minFloat(1.0, 0.35+0.65*edge)) * weight
		basis := tangents[index]
		output = append(output, MousePoint{
			X: point.X + basis.tx*tanNoise[index]*maxTan*weight + basis.nx*normNoise[index]*maxNorm*weight,
			Y: point.Y + basis.ty*tanNoise[index]*maxTan*weight + basis.ny*normNoise[index]*maxNorm*weight,
		})
	}

	return output
}

func overshootPoint(start MousePoint, end MousePoint, random inputRandomSource) MousePoint {
	dx := end.X - start.X
	dy := end.Y - start.Y
	dist := math.Hypot(dx, dy)
	if dist == 0 {
		dist = 1.0
	}

	ux := dx / dist
	uy := dy / dist
	padding := maxFloat(24.0, minFloat(dist*randFloatRange(random, 0.10, 0.25), 180.0))
	return MousePoint{
		X: end.X + ux*padding,
		Y: end.Y + uy*padding,
	}
}

func concatMousePaths(segments ...[]MousePoint) []MousePoint {
	var output []MousePoint
	for _, segment := range segments {
		if len(segment) == 0 {
			continue
		}
		if len(output) > 0 && output[len(output)-1] == segment[0] {
			output = append(output, segment[1:]...)
			continue
		}
		output = append(output, segment...)
	}
	return output
}

func cloneStringSliceExact(src []string) []string {
	if src == nil {
		return nil
	}

	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

func randFloatRange(random inputRandomSource, min float64, max float64) float64 {
	return min + (max-min)*random.Float64()
}

func randIntRange(random inputRandomSource, min int, max int) int {
	if max <= min {
		return min
	}
	return min + random.Intn(max-min+1)
}

func maxFloat(left float64, right float64) float64 {
	if left > right {
		return left
	}
	return right
}

func minFloat(left float64, right float64) float64 {
	if left < right {
		return left
	}
	return right
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
