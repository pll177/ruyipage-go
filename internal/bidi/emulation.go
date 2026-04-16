package bidi

import (
	"fmt"
	"strings"
	"time"

	"ruyipage-go/internal/support"
)

type emulationCommandDriver interface {
	Run(method string, params map[string]any, timeout time.Duration) (map[string]any, error)
}

// SetUserAgentOverride 调用 emulation.setUserAgentOverride 覆盖当前 UA。
func SetUserAgentOverride(
	driver emulationCommandDriver,
	userAgent string,
	platform string,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"userAgent": userAgent,
	}
	if platform != "" {
		params["platform"] = platform
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}
	return runEmulationCommand(driver, "emulation.setUserAgentOverride", params, timeout)
}

// SetGeolocationOverride 调用 emulation.setGeolocationOverride 覆盖地理位置。
func SetGeolocationOverride(
	driver emulationCommandDriver,
	latitude *float64,
	longitude *float64,
	accuracy *float64,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{}
	if latitude != nil && longitude != nil {
		coordinates := map[string]any{
			"latitude":  *latitude,
			"longitude": *longitude,
		}
		if accuracy != nil {
			coordinates["accuracy"] = *accuracy
		}
		params["coordinates"] = coordinates
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}
	return runEmulationCommand(driver, "emulation.setGeolocationOverride", params, timeout)
}

// SetTimezoneOverride 调用 emulation.setTimezoneOverride 覆盖时区。
func SetTimezoneOverride(
	driver emulationCommandDriver,
	timezoneID string,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	if timezoneID == "" {
		return map[string]any{}, nil
	}

	params := map[string]any{
		"timezone": timezoneID,
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}
	return runEmulationCommand(driver, "emulation.setTimezoneOverride", params, timeout)
}

// SetLocaleOverride 调用 emulation.setLocaleOverride 覆盖 locale。
func SetLocaleOverride(
	driver emulationCommandDriver,
	locales any,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	locale, err := normalizeEmulationLocale(locales)
	if err != nil {
		return nil, err
	}
	if locale == "" {
		return map[string]any{}, nil
	}

	params := map[string]any{
		"locale": locale,
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}
	return runEmulationCommand(driver, "emulation.setLocaleOverride", params, timeout)
}

// SetScreenOrientationOverride 调用 emulation.setScreenOrientationOverride 覆盖屏幕方向。
func SetScreenOrientationOverride(
	driver emulationCommandDriver,
	orientationType string,
	angle int,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"screenOrientation": map[string]any{
			"type":    orientationType,
			"angle":   angle,
			"natural": resolveScreenOrientationNatural(orientationType),
		},
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}
	return runEmulationCommand(driver, "emulation.setScreenOrientationOverride", params, timeout)
}

// SetScreenSettingsOverride 调用 emulation.setScreenSettingsOverride 覆盖屏幕尺寸与 DPR。
func SetScreenSettingsOverride(
	driver emulationCommandDriver,
	width *int,
	height *int,
	devicePixelRatio *float64,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{}
	if width != nil || height != nil {
		screenArea := map[string]any{}
		if width != nil {
			screenArea["width"] = *width
		}
		if height != nil {
			screenArea["height"] = *height
		}
		params["screenArea"] = screenArea
	}
	if devicePixelRatio != nil {
		params["devicePixelRatio"] = *devicePixelRatio
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}
	return runEmulationCommand(driver, "emulation.setScreenSettingsOverride", params, timeout)
}

// SetNetworkConditions 调用 emulation.setNetworkConditions 切换离线/在线。
func SetNetworkConditions(
	driver emulationCommandDriver,
	offline bool,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"networkConditions": map[string]any{
			"type": resolveNetworkConditionsType(offline),
		},
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}
	return runEmulationCommand(driver, "emulation.setNetworkConditions", params, timeout)
}

// SetTouchOverride 调用 emulation.setTouchOverride 设置最大触点数；nil 表示清除覆盖。
func SetTouchOverride(
	driver emulationCommandDriver,
	maxTouchPoints *int,
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

	params := map[string]any{
		"maxTouchPoints": nil,
	}
	if maxTouchPoints != nil {
		params["maxTouchPoints"] = *maxTouchPoints
	}
	if includeContexts {
		params["contexts"] = normalizedContexts
	}
	if includeUserContexts {
		params["userContexts"] = normalizedUserContexts
	}
	return runEmulationCommand(driver, "emulation.setTouchOverride", params, timeout)
}

// SetMediaFeaturesOverride 调用 emulation.setMediaFeaturesOverride 覆盖媒体特性。
func SetMediaFeaturesOverride(
	driver emulationCommandDriver,
	features []map[string]any,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"features": cloneAnyMapSliceDeep(features),
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}
	return runEmulationCommand(driver, "emulation.setMediaFeaturesOverride", params, timeout)
}

// SetDocumentCookieDisabled 调用 emulation.setDocumentCookieDisabled 切换 document.cookie 能力。
func SetDocumentCookieDisabled(
	driver emulationCommandDriver,
	disabled bool,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"disabled": disabled,
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}
	return runEmulationCommand(driver, "emulation.setDocumentCookieDisabled", params, timeout)
}

// SetBypassCSPOverride 调用 emulation.setBypassCSP 切换 CSP 绕过。
func SetBypassCSPOverride(
	driver emulationCommandDriver,
	enabled bool,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"enabled": enabled,
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}
	return runEmulationCommand(driver, "emulation.setBypassCSP", params, timeout)
}

// SetFocusEmulation 调用 emulation.setFocusEmulation 切换焦点模拟。
func SetFocusEmulation(
	driver emulationCommandDriver,
	enabled bool,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"enabled": enabled,
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}
	return runEmulationCommand(driver, "emulation.setFocusEmulation", params, timeout)
}

// SetHardwareConcurrency 调用 emulation.setHardwareConcurrency 覆盖 CPU 核心数。
func SetHardwareConcurrency(
	driver emulationCommandDriver,
	concurrency int,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"hardwareConcurrency": concurrency,
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}
	return runEmulationCommand(driver, "emulation.setHardwareConcurrency", params, timeout)
}

// SetScriptingEnabled 调用 emulation.setScriptingEnabled 切换脚本开关。
func SetScriptingEnabled(
	driver emulationCommandDriver,
	enabled bool,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"enabled": enabled,
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}
	return runEmulationCommand(driver, "emulation.setScriptingEnabled", params, timeout)
}

// SetScrollbarTypeOverride 调用 emulation.setScrollbarTypeOverride 覆盖滚动条类型。
func SetScrollbarTypeOverride(
	driver emulationCommandDriver,
	scrollbarType string,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"type": scrollbarType,
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}
	return runEmulationCommand(driver, "emulation.setScrollbarTypeOverride", params, timeout)
}

// SetForcedColorsModeThemeOverride 调用 emulation.setForcedColorsModeThemeOverride 覆盖强制颜色模式。
func SetForcedColorsModeThemeOverride(
	driver emulationCommandDriver,
	mode string,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"mode": mode,
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}
	return runEmulationCommand(driver, "emulation.setForcedColorsModeThemeOverride", params, timeout)
}

func runEmulationCommand(
	driver emulationCommandDriver,
	method string,
	params map[string]any,
	timeout time.Duration,
) (map[string]any, error) {
	if driver == nil {
		return nil, support.NewPageDisconnectedError("emulation driver 未初始化", nil)
	}

	result, err := driver.Run(method, params, timeout)
	if err != nil {
		if isUnsupportedBiDiCommandError(err) {
			return nil, nil
		}
		return nil, err
	}
	return cloneAnyMapDeep(result), nil
}

func normalizeEmulationLocale(locales any) (string, error) {
	switch typed := locales.(type) {
	case nil:
		return "", nil
	case string:
		return typed, nil
	case []string:
		if len(typed) == 0 {
			return "", nil
		}
		return typed[0], nil
	default:
		return "", fmt.Errorf("locales 参数必须为 string 或 []string")
	}
}

func resolveScreenOrientationNatural(orientationType string) string {
	if strings.Contains(strings.ToLower(orientationType), "portrait") {
		return "portrait"
	}
	return "landscape"
}

func resolveNetworkConditionsType(offline bool) string {
	if offline {
		return "offline"
	}
	return "online"
}
