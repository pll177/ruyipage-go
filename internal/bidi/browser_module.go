package bidi

import (
	"fmt"
	"time"

	"github.com/pll177/ruyipage-go/internal/support"
)

type browserCommandDriver interface {
	Run(method string, params map[string]any, timeout time.Duration) (map[string]any, error)
}

// CloseBrowser 调用 browser.close 关闭浏览器。
func CloseBrowser(driver browserCommandDriver, timeout time.Duration) (map[string]any, error) {
	return runBrowserCommand(driver, "browser.close", nil, timeout)
}

// CreateUserContext 调用 browser.createUserContext 创建新的 user context。
func CreateUserContext(driver browserCommandDriver, timeout time.Duration) (map[string]any, error) {
	return runBrowserCommand(driver, "browser.createUserContext", nil, timeout)
}

// GetUserContexts 调用 browser.getUserContexts 获取全部 user context。
func GetUserContexts(driver browserCommandDriver, timeout time.Duration) (map[string]any, error) {
	return runBrowserCommand(driver, "browser.getUserContexts", nil, timeout)
}

// RemoveUserContext 调用 browser.removeUserContext 删除指定 user context。
func RemoveUserContext(
	driver browserCommandDriver,
	userContext string,
	timeout time.Duration,
) (map[string]any, error) {
	return runBrowserCommand(
		driver,
		"browser.removeUserContext",
		map[string]any{"userContext": userContext},
		timeout,
	)
}

// GetClientWindows 调用 browser.getClientWindows 获取浏览器窗口信息。
func GetClientWindows(driver browserCommandDriver, timeout time.Duration) (map[string]any, error) {
	return runBrowserCommand(driver, "browser.getClientWindows", nil, timeout)
}

// SetClientWindowState 调用 browser.setClientWindowState 设置窗口状态与几何信息。
func SetClientWindowState(
	driver browserCommandDriver,
	clientWindow string,
	state string,
	width *int,
	height *int,
	x *int,
	y *int,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"clientWindow": clientWindow,
	}
	if state != "" {
		params["state"] = state
	}
	if width != nil {
		params["width"] = *width
	}
	if height != nil {
		params["height"] = *height
	}
	if x != nil {
		params["x"] = *x
	}
	if y != nil {
		params["y"] = *y
	}

	return runBrowserCommand(driver, "browser.setClientWindowState", params, timeout)
}

// SetDownloadBehavior 调用 browser.setDownloadBehavior 设置下载行为。
func SetDownloadBehavior(
	driver browserCommandDriver,
	behavior string,
	downloadPath string,
	contexts any,
	userContexts any,
	timeout time.Duration,
) (map[string]any, error) {
	normalizedContexts, includeContexts, err := normalizeOptionalStringList(contexts, "contexts")
	if err != nil {
		return nil, err
	}
	normalizedUserContexts, includeUserContexts, err := normalizeOptionalStringList(userContexts, "userContexts")
	if err != nil {
		return nil, err
	}
	if includeContexts && includeUserContexts {
		return nil, fmt.Errorf("contexts 与 userContexts 不能同时设置")
	}

	resolvedBehavior := resolveBrowserDownloadBehavior(behavior)
	downloadBehavior := map[string]any{
		"type":     resolveBrowserDownloadBehaviorType(resolvedBehavior),
		"behavior": resolvedBehavior,
	}
	if downloadPath != "" {
		downloadBehavior["downloadPath"] = downloadPath
	}

	params := map[string]any{
		"downloadBehavior": downloadBehavior,
	}
	if includeContexts {
		params["contexts"] = normalizedContexts
	}
	if includeUserContexts {
		params["userContexts"] = normalizedUserContexts
	}

	return runBrowserCommand(driver, "browser.setDownloadBehavior", params, timeout)
}

func runBrowserCommand(
	driver browserCommandDriver,
	method string,
	params map[string]any,
	timeout time.Duration,
) (map[string]any, error) {
	if driver == nil {
		return nil, support.NewPageDisconnectedError("browser driver 未初始化", nil)
	}

	result, err := driver.Run(method, params, timeout)
	if err != nil {
		return nil, err
	}
	return cloneAnyMapDeep(result), nil
}

func resolveBrowserDownloadBehavior(behavior string) string {
	if behavior == "" {
		return "allow"
	}
	return behavior
}

func resolveBrowserDownloadBehaviorType(behavior string) string {
	if behavior == "allow" || behavior == "allowAndOpen" {
		return "allowed"
	}
	return "denied"
}
