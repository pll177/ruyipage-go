package bidi

import (
	"os"
	"time"

	"ruyipage-go/internal/support"
)

type webExtensionCommandDriver interface {
	Run(method string, params map[string]any, timeout time.Duration) (map[string]any, error)
}

// Install 调用 webExtension.install 安装扩展。
func Install(
	driver webExtensionCommandDriver,
	path string,
	timeout time.Duration,
) (map[string]any, error) {
	if driver == nil {
		return nil, support.NewPageDisconnectedError("webExtension driver 未初始化", nil)
	}

	params := map[string]any{
		"extensionData": map[string]any{
			"type": resolveExtensionPathType(path),
			"path": path,
		},
	}

	result, err := driver.Run("webExtension.install", params, timeout)
	if err != nil {
		if isUnsupportedBiDiCommandError(err) {
			return nil, nil
		}
		return nil, err
	}
	return cloneAnyMapDeep(result), nil
}

// Uninstall 调用 webExtension.uninstall 卸载扩展。
func Uninstall(
	driver webExtensionCommandDriver,
	extensionID string,
	timeout time.Duration,
) (map[string]any, error) {
	if driver == nil {
		return nil, support.NewPageDisconnectedError("webExtension driver 未初始化", nil)
	}

	result, err := driver.Run(
		"webExtension.uninstall",
		map[string]any{"extension": extensionID},
		timeout,
	)
	if err != nil {
		if isUnsupportedBiDiCommandError(err) {
			return nil, nil
		}
		return nil, err
	}
	return cloneAnyMapDeep(result), nil
}

func resolveExtensionPathType(path string) string {
	info, err := os.Stat(path)
	if err == nil && !info.IsDir() {
		return "archivePath"
	}
	return "path"
}
