package bidi

import (
	"time"

	"github.com/pll177/ruyipage-go/internal/support"
)

type storageCommandDriver interface {
	Run(method string, params map[string]any, timeout time.Duration) (map[string]any, error)
}

// GetCookies 调用 storage.getCookies 获取 Cookie 列表。
func GetCookies(
	driver storageCommandDriver,
	filter map[string]any,
	partition map[string]any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{}
	if len(filter) > 0 {
		params["filter"] = cloneAnyMapDeep(filter)
	}
	if normalized := normalizeStoragePartition(partition); len(normalized) > 0 {
		params["partition"] = normalized
	}

	return runStorageCommand(driver, "storage.getCookies", params, timeout)
}

// SetCookie 调用 storage.setCookie 设置单个 Cookie。
func SetCookie(
	driver storageCommandDriver,
	cookie map[string]any,
	partition map[string]any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{
		"cookie": cloneAnyMapDeep(cookie),
	}
	if normalized := normalizeStoragePartition(partition); len(normalized) > 0 {
		params["partition"] = normalized
	}

	return runStorageCommand(driver, "storage.setCookie", params, timeout)
}

// DeleteCookies 调用 storage.deleteCookies 删除 Cookie。
func DeleteCookies(
	driver storageCommandDriver,
	filter map[string]any,
	partition map[string]any,
	timeout time.Duration,
) (map[string]any, error) {
	params := map[string]any{}
	if len(filter) > 0 {
		params["filter"] = cloneAnyMapDeep(filter)
	}
	if normalized := normalizeStoragePartition(partition); len(normalized) > 0 {
		params["partition"] = normalized
	}

	return runStorageCommand(driver, "storage.deleteCookies", params, timeout)
}

func runStorageCommand(
	driver storageCommandDriver,
	method string,
	params map[string]any,
	timeout time.Duration,
) (map[string]any, error) {
	if driver == nil {
		return nil, support.NewPageDisconnectedError("storage driver 未初始化", nil)
	}

	result, err := driver.Run(method, params, timeout)
	if err != nil {
		return nil, err
	}
	return cloneAnyMapDeep(result), nil
}

func normalizeStoragePartition(partition map[string]any) map[string]any {
	if len(partition) == 0 {
		return nil
	}

	normalized := cloneAnyMapDeep(partition)
	if _, ok := normalized["type"]; ok {
		return normalized
	}

	if _, ok := normalized["context"]; ok {
		normalized["type"] = "context"
		return normalized
	}

	if _, ok := normalized["userContext"]; ok {
		normalized["type"] = "storageKey"
		return normalized
	}
	if _, ok := normalized["sourceOrigin"]; ok {
		normalized["type"] = "storageKey"
		return normalized
	}

	return normalized
}
