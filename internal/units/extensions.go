package units

import (
	"sync"
	"time"

	"ruyipage-go/internal/base"
	"ruyipage-go/internal/bidi"
)

type extensionManagerOwner interface {
	BrowserDriver() *base.BrowserBiDiDriver
	BaseTimeout() time.Duration
}

// ExtensionManager 提供 WebExtension 安装与卸载入口。
type ExtensionManager struct {
	owner extensionManagerOwner

	mu        sync.RWMutex
	installed map[string]string
}

// NewExtensionManager 创建扩展管理器。
func NewExtensionManager(owner extensionManagerOwner) *ExtensionManager {
	return &ExtensionManager{
		owner:     owner,
		installed: make(map[string]string),
	}
}

// Install 安装目录或压缩包形式的扩展；当前浏览器不支持时返回空 ID。
func (m *ExtensionManager) Install(path string) (string, error) {
	if m == nil || m.owner == nil {
		return "", nil
	}
	result, err := bidi.Install(m.owner.BrowserDriver(), path, m.owner.BaseTimeout())
	if err != nil {
		return "", err
	}
	extensionID := stringifyNetworkValue(result["extension"])
	if extensionID == "" {
		return "", nil
	}

	m.mu.Lock()
	m.installed[extensionID] = path
	m.mu.Unlock()
	return extensionID, nil
}

// InstallDir 安装解压目录扩展。
func (m *ExtensionManager) InstallDir(path string) (string, error) {
	return m.Install(path)
}

// InstallArchive 安装 xpi / 压缩包扩展。
func (m *ExtensionManager) InstallArchive(path string) (string, error) {
	return m.Install(path)
}

// Uninstall 卸载指定扩展。
func (m *ExtensionManager) Uninstall(extensionID string) error {
	if m == nil || m.owner == nil {
		return nil
	}
	_, err := bidi.Uninstall(m.owner.BrowserDriver(), extensionID, m.owner.BaseTimeout())
	if err != nil {
		return err
	}

	m.mu.Lock()
	delete(m.installed, extensionID)
	m.mu.Unlock()
	return nil
}

// UninstallAll 卸载当前管理器记录的全部扩展。
func (m *ExtensionManager) UninstallAll() error {
	if m == nil {
		return nil
	}

	m.mu.RLock()
	ids := make([]string, 0, len(m.installed))
	for extensionID := range m.installed {
		ids = append(ids, extensionID)
	}
	m.mu.RUnlock()

	for _, extensionID := range ids {
		if err := m.Uninstall(extensionID); err != nil {
			return err
		}
	}
	return nil
}

// InstalledExtensions 返回已安装扩展的快照副本。
func (m *ExtensionManager) InstalledExtensions() map[string]string {
	if m == nil {
		return map[string]string{}
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]string, len(m.installed))
	for extensionID, path := range m.installed {
		result[extensionID] = path
	}
	return result
}
