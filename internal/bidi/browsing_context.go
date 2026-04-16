package bidi

import (
	"time"

	"ruyipage-go/internal/support"
)

type browsingContextCommandDriver interface {
	Run(method string, params map[string]any, timeout time.Duration) (map[string]any, error)
}

// Navigate 调用 browsingContext.navigate 导航到目标 URL。
func Navigate(
	driver browsingContextCommandDriver,
	context string,
	url string,
	wait string,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"context": context,
		"url":     url,
		"wait":    resolveBrowsingContextDefault(wait, "complete"),
	}
	return runBrowsingContextCommand(driver, "browsingContext.navigate", params, timeout)
}

// GetTree 调用 browsingContext.getTree 获取上下文树。
func GetTree(
	driver browsingContextCommandDriver,
	maxDepth *int,
	root string,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{}
	if maxDepth != nil {
		params["maxDepth"] = *maxDepth
	}
	if root != "" {
		params["root"] = root
	}
	return runBrowsingContextCommand(driver, "browsingContext.getTree", params, timeout)
}

// Create 调用 browsingContext.create 创建新的上下文。
func Create(
	driver browsingContextCommandDriver,
	type_ string,
	referenceContext string,
	background bool,
	userContext string,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"type": resolveBrowsingContextDefault(type_, "tab"),
	}
	if referenceContext != "" {
		params["referenceContext"] = referenceContext
	}
	if background {
		params["background"] = true
	}
	if userContext != "" {
		params["userContext"] = userContext
	}
	return runBrowsingContextCommand(driver, "browsingContext.create", params, timeout)
}

// Close 调用 browsingContext.close 关闭上下文。
func Close(
	driver browsingContextCommandDriver,
	context string,
	promptUnload bool,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"context": context,
	}
	if promptUnload {
		params["promptUnload"] = true
	}
	return runBrowsingContextCommand(driver, "browsingContext.close", params, timeout)
}

// Activate 调用 browsingContext.activate 激活上下文。
func Activate(
	driver browsingContextCommandDriver,
	context string,
	timeout time.Duration,
) (map[string]any, error) {
	return runBrowsingContextCommand(
		driver,
		"browsingContext.activate",
		map[string]any{"context": context},
		timeout,
	)
}

// CaptureScreenshot 调用 browsingContext.captureScreenshot 截图。
func CaptureScreenshot(
	driver browsingContextCommandDriver,
	context string,
	origin string,
	format map[string]any,
	clip map[string]any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"context": context,
		"origin":  resolveBrowsingContextDefault(origin, "viewport"),
	}
	if len(format) > 0 {
		params["format"] = cloneAnyMapDeep(format)
	}
	if len(clip) > 0 {
		params["clip"] = cloneAnyMapDeep(clip)
	}
	return runBrowsingContextCommand(driver, "browsingContext.captureScreenshot", params, timeout)
}

// Print 调用 browsingContext.print 输出 PDF。
func Print(
	driver browsingContextCommandDriver,
	context string,
	background *bool,
	margin map[string]any,
	orientation string,
	page map[string]any,
	pageRanges []string,
	scale *float64,
	shrinkToFit *bool,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"context": context,
	}
	if background != nil {
		params["background"] = *background
	}
	if len(margin) > 0 {
		params["margin"] = cloneAnyMapDeep(margin)
	}
	if orientation != "" {
		params["orientation"] = orientation
	}
	if len(page) > 0 {
		params["page"] = cloneAnyMapDeep(page)
	}
	if len(pageRanges) > 0 {
		params["pageRanges"] = cloneStringSlice(pageRanges)
	}
	if scale != nil {
		params["scale"] = *scale
	}
	if shrinkToFit != nil {
		params["shrinkToFit"] = *shrinkToFit
	}
	return runBrowsingContextCommand(driver, "browsingContext.print", params, timeout)
}

// Reload 调用 browsingContext.reload 重载页面。
func Reload(
	driver browsingContextCommandDriver,
	context string,
	ignoreCache bool,
	wait string,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"context": context,
		"wait":    resolveBrowsingContextDefault(wait, "complete"),
	}
	if ignoreCache {
		params["ignoreCache"] = true
	}
	return runBrowsingContextCommand(driver, "browsingContext.reload", params, timeout)
}

// TraverseHistory 调用 browsingContext.traverseHistory 执行历史导航。
func TraverseHistory(
	driver browsingContextCommandDriver,
	context string,
	delta int,
	timeout time.Duration,
) (map[string]any, error) {
	return runBrowsingContextCommand(
		driver,
		"browsingContext.traverseHistory",
		map[string]any{
			"context": context,
			"delta":   delta,
		},
		timeout,
	)
}

// HandleUserPrompt 调用 browsingContext.handleUserPrompt 处理用户提示框。
func HandleUserPrompt(
	driver browsingContextCommandDriver,
	context string,
	accept bool,
	userText *string,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"context": context,
		"accept":  accept,
	}
	if userText != nil {
		params["userText"] = *userText
	}
	return runBrowsingContextCommand(driver, "browsingContext.handleUserPrompt", params, timeout)
}

// LocateNodes 调用 browsingContext.locateNodes 查找节点。
func LocateNodes(
	driver browsingContextCommandDriver,
	context string,
	locator map[string]any,
	maxNodeCount *int,
	serializationOptions map[string]any,
	startNodes []map[string]any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"context": context,
		"locator": cloneAnyMapDeep(locator),
	}
	if maxNodeCount != nil {
		params["maxNodeCount"] = *maxNodeCount
	}
	if len(serializationOptions) > 0 {
		params["serializationOptions"] = cloneAnyMapDeep(serializationOptions)
	}
	if len(startNodes) > 0 {
		params["startNodes"] = cloneAnyMapSliceDeep(startNodes)
	}
	return runBrowsingContextCommand(driver, "browsingContext.locateNodes", params, timeout)
}

// SetViewport 调用 browsingContext.setViewport 设置视口。
func SetViewport(
	driver browsingContextCommandDriver,
	context string,
	width *int,
	height *int,
	devicePixelRatio *float64,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"context": context,
	}
	if width != nil && height != nil {
		params["viewport"] = map[string]any{
			"width":  *width,
			"height": *height,
		}
	}
	if devicePixelRatio != nil {
		params["devicePixelRatio"] = *devicePixelRatio
	}
	return runBrowsingContextCommand(driver, "browsingContext.setViewport", params, timeout)
}

// SetBypassCSP 调用 browsingContext.setBypassCSP 设置 CSP 绕过开关。
func SetBypassCSP(
	driver browsingContextCommandDriver,
	context string,
	enabled bool,
	timeout time.Duration,
) (map[string]any, error) {
	return runBrowsingContextCommand(
		driver,
		"browsingContext.setBypassCSP",
		map[string]any{
			"context": context,
			"enabled": enabled,
		},
		timeout,
	)
}

func runBrowsingContextCommand(
	driver browsingContextCommandDriver,
	method string,
	params map[string]any,
	timeout time.Duration,
) (map[string]any, error) {
	if driver == nil {
		return nil, support.NewPageDisconnectedError("browsingContext driver 未初始化", nil)
	}

	result, err := driver.Run(method, params, timeout)
	if err != nil {
		return nil, err
	}
	return cloneAnyMapDeep(result), nil
}

func resolveBrowsingContextDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func cloneAnyMapDeep(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}

	dst := make(map[string]any, len(src))
	for key, value := range src {
		dst[key] = cloneAnyValueDeep(value)
	}
	return dst
}

func cloneAnyValueDeep(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneAnyMapDeep(typed)
	case map[string]string:
		dst := make(map[string]string, len(typed))
		for key, item := range typed {
			dst[key] = item
		}
		return dst
	case []any:
		dst := make([]any, len(typed))
		for index, item := range typed {
			dst[index] = cloneAnyValueDeep(item)
		}
		return dst
	case []string:
		return cloneStringSlice(typed)
	case []map[string]any:
		return cloneAnyMapSliceDeep(typed)
	default:
		return value
	}
}

func cloneAnyMapSliceDeep(src []map[string]any) []map[string]any {
	if src == nil {
		return nil
	}

	dst := make([]map[string]any, len(src))
	for index, item := range src {
		dst[index] = cloneAnyMapDeep(item)
	}
	return dst
}
