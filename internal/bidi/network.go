package bidi

import (
	"time"

	"ruyipage-go/internal/support"
)

const defaultNetworkMaxEncodedDataSize = 10485760

type networkCommandDriver interface {
	Run(method string, params map[string]any, timeout time.Duration) (map[string]any, error)
}

// AddIntercept 调用 network.addIntercept 注册网络拦截。
func AddIntercept(
	driver networkCommandDriver,
	phases any,
	urlPatterns []map[string]any,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	normalizedPhases, err := normalizeRequiredStringList(phases, "phases")
	if err != nil {
		return nil, err
	}

	params := map[string]any{
		"phases": normalizedPhases,
	}
	if len(urlPatterns) > 0 {
		params["urlPatterns"] = cloneAnyMapSliceDeep(urlPatterns)
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}

	return runNetworkCommand(driver, "network.addIntercept", params, timeout)
}

// RemoveIntercept 调用 network.removeIntercept 移除拦截。
func RemoveIntercept(
	driver networkCommandDriver,
	interceptID string,
	timeout time.Duration,
) (map[string]any, error) {
	return runNetworkCommand(
		driver,
		"network.removeIntercept",
		map[string]any{"intercept": interceptID},
		timeout,
	)
}

// ContinueRequest 调用 network.continueRequest 放行并可修改请求。
func ContinueRequest(
	driver networkCommandDriver,
	requestID string,
	body any,
	cookies any,
	headers any,
	method string,
	url string,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"request": requestID,
	}
	if body != nil {
		params["body"] = cloneAnyValueDeep(body)
	}
	if cookies != nil {
		params["cookies"] = cloneAnyValueDeep(cookies)
	}
	if headers != nil {
		params["headers"] = cloneAnyValueDeep(headers)
	}
	if method != "" {
		params["method"] = method
	}
	if url != "" {
		params["url"] = url
	}

	return runNetworkCommand(driver, "network.continueRequest", params, timeout)
}

// ContinueResponse 调用 network.continueResponse 放行并可修改响应。
func ContinueResponse(
	driver networkCommandDriver,
	requestID string,
	cookies any,
	credentials any,
	headers any,
	reasonPhrase string,
	statusCode *int,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"request": requestID,
	}
	if cookies != nil {
		params["cookies"] = cloneAnyValueDeep(cookies)
	}
	if credentials != nil {
		params["credentials"] = cloneAnyValueDeep(credentials)
	}
	if headers != nil {
		params["headers"] = cloneAnyValueDeep(headers)
	}
	if reasonPhrase != "" {
		params["reasonPhrase"] = reasonPhrase
	}
	if statusCode != nil {
		params["statusCode"] = *statusCode
	}

	return runNetworkCommand(driver, "network.continueResponse", params, timeout)
}

// ContinueWithAuth 调用 network.continueWithAuth 处理认证挑战。
func ContinueWithAuth(
	driver networkCommandDriver,
	requestID string,
	action string,
	credentials any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"request": requestID,
		"action":  resolveNetworkStringDefault(action, "default"),
	}
	if credentials != nil {
		params["credentials"] = cloneAnyValueDeep(credentials)
	}

	return runNetworkCommand(driver, "network.continueWithAuth", params, timeout)
}

// FailRequest 调用 network.failRequest 中止请求。
func FailRequest(
	driver networkCommandDriver,
	requestID string,
	timeout time.Duration,
) (map[string]any, error) {
	return runNetworkCommand(
		driver,
		"network.failRequest",
		map[string]any{"request": requestID},
		timeout,
	)
}

// ProvideResponse 调用 network.provideResponse 为请求提供模拟响应。
func ProvideResponse(
	driver networkCommandDriver,
	requestID string,
	body any,
	cookies any,
	headers any,
	reasonPhrase string,
	statusCode *int,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"request": requestID,
	}
	if body != nil {
		params["body"] = cloneAnyValueDeep(body)
	}
	if cookies != nil {
		params["cookies"] = cloneAnyValueDeep(cookies)
	}
	if headers != nil {
		params["headers"] = cloneAnyValueDeep(headers)
	}
	if reasonPhrase != "" {
		params["reasonPhrase"] = reasonPhrase
	}
	if statusCode != nil {
		params["statusCode"] = *statusCode
	}

	return runNetworkCommand(driver, "network.provideResponse", params, timeout)
}

// SetCacheBehavior 调用 network.setCacheBehavior 设置缓存行为。
func SetCacheBehavior(
	driver networkCommandDriver,
	behavior string,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"cacheBehavior": resolveNetworkStringDefault(behavior, "default"),
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}

	return runNetworkCommand(driver, "network.setCacheBehavior", params, timeout)
}

// SetExtraHeaders 调用 network.setExtraHeaders 设置额外请求头。
func SetExtraHeaders(
	driver networkCommandDriver,
	headers any,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"headers": cloneAnyValueDeep(headers),
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}

	return runNetworkCommand(driver, "network.setExtraHeaders", params, timeout)
}

// AddDataCollector 调用 network.addDataCollector 注册数据收集器。
func AddDataCollector(
	driver networkCommandDriver,
	events any,
	contexts any,
	maxEncodedDataSize int,
	dataTypes any,
	timeout time.Duration,
) (map[string]any, error) {
	normalizedEvents, err := normalizeRequiredStringList(events, "events")
	if err != nil {
		return nil, err
	}
	normalizedDataTypes, err := resolveNetworkDataTypes(dataTypes)
	if err != nil {
		return nil, err
	}

	params := map[string]any{
		"events":             normalizedEvents,
		"maxEncodedDataSize": resolveNetworkMaxEncodedDataSize(maxEncodedDataSize),
		"dataTypes":          normalizedDataTypes,
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}

	return runNetworkCommand(driver, "network.addDataCollector", params, timeout)
}

// RemoveDataCollector 调用 network.removeDataCollector 移除收集器。
func RemoveDataCollector(
	driver networkCommandDriver,
	collectorID string,
	timeout time.Duration,
) (map[string]any, error) {
	return runNetworkCommand(
		driver,
		"network.removeDataCollector",
		map[string]any{"collector": collectorID},
		timeout,
	)
}

// GetData 调用 network.getData 获取收集器持有的数据。
func GetData(
	driver networkCommandDriver,
	collectorID string,
	requestID string,
	dataType string,
	timeout time.Duration,
) (map[string]any, error) {
	return runNetworkCommand(
		driver,
		"network.getData",
		map[string]any{
			"collector": collectorID,
			"request":   requestID,
			"dataType":  resolveNetworkStringDefault(dataType, "response"),
		},
		timeout,
	)
}

// DisownData 调用 network.disownData 释放收集器持有的数据。
func DisownData(
	driver networkCommandDriver,
	collectorID string,
	requestID string,
	dataType string,
	timeout time.Duration,
) (map[string]any, error) {
	return runNetworkCommand(
		driver,
		"network.disownData",
		map[string]any{
			"collector": collectorID,
			"request":   requestID,
			"dataType":  resolveNetworkStringDefault(dataType, "response"),
		},
		timeout,
	)
}

func runNetworkCommand(
	driver networkCommandDriver,
	method string,
	params map[string]any,
	timeout time.Duration,
) (map[string]any, error) {
	if driver == nil {
		return nil, support.NewPageDisconnectedError("network driver 未初始化", nil)
	}

	result, err := driver.Run(method, params, timeout)
	if err != nil {
		return nil, err
	}
	return cloneAnyMapDeep(result), nil
}

func resolveNetworkStringDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func resolveNetworkMaxEncodedDataSize(value int) int {
	if value == 0 {
		return defaultNetworkMaxEncodedDataSize
	}
	return value
}

func resolveNetworkDataTypes(value any) ([]string, error) {
	normalized, include, err := normalizeOptionalStringList(value, "dataTypes")
	if err != nil {
		return nil, err
	}
	if !include {
		return []string{"request", "response"}, nil
	}
	return normalized, nil
}
