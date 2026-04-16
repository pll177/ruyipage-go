package bidi

import (
	"time"

	"ruyipage-go/internal/support"
)

const defaultPermissionOrigin = "https://example.com"

type permissionsCommandDriver interface {
	Run(method string, params map[string]any, timeout time.Duration) (map[string]any, error)
}

// SetPermission 调用 permissions.setPermission 设置浏览器权限。
func SetPermission(
	driver permissionsCommandDriver,
	descriptor map[string]any,
	state string,
	origin string,
	contexts any,
	timeout time.Duration,
) (map[string]any, error) {
	if driver == nil {
		return nil, support.NewPageDisconnectedError("permissions driver 未初始化", nil)
	}

	params := map[string]any{
		"descriptor": cloneAnyMapDeep(descriptor),
		"state":      state,
		"origin":     resolvePermissionOrigin(origin),
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return nil, err
	} else if include {
		params["contexts"] = normalizedContexts
	}

	result, err := driver.Run("permissions.setPermission", params, timeout)
	if err != nil {
		if isUnsupportedBiDiCommandError(err) {
			return nil, nil
		}
		return nil, err
	}
	return cloneAnyMapDeep(result), nil
}

func resolvePermissionOrigin(origin string) string {
	if origin == "" {
		return defaultPermissionOrigin
	}
	return origin
}
