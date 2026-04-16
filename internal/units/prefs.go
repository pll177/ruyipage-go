package units

import (
	"errors"
	"path/filepath"

	"ruyipage-go/internal/support"
)

// PrefsManager 提供轻量的 user.js 偏好读写能力。
type PrefsManager struct {
	resolveProfile ProfileResolver
}

// NewPrefsManager 创建绑定固定 profile 路径的 PrefsManager。
func NewPrefsManager(profilePath string) *PrefsManager {
	return NewPrefsManagerWithResolver(func() string {
		return profilePath
	})
}

// NewPrefsManagerWithResolver 创建按需解析 profile 路径的 PrefsManager。
func NewPrefsManagerWithResolver(resolver ProfileResolver) *PrefsManager {
	if resolver == nil {
		resolver = func() string { return "" }
	}
	return &PrefsManager{resolveProfile: resolver}
}

// ProfilePath 返回当前解析到的 profile 路径。
func (m *PrefsManager) ProfilePath() string {
	if m == nil || m.resolveProfile == nil {
		return ""
	}
	return m.resolveProfile()
}

// Get 从 user.js 读取单个 pref。
func (m *PrefsManager) Get(key string) (any, error) {
	if m == nil || key == "" {
		return nil, nil
	}
	value, ok, err := m.userJS().Read(key)
	if err != nil || !ok {
		return nil, err
	}
	return value, nil
}

// Set 写入 user.js。
func (m *PrefsManager) Set(key string, value any) error {
	return m.SetPersistent(key, value)
}

// SetPersistent 持久化写入 user.js。
func (m *PrefsManager) SetPersistent(key string, value any) error {
	if m == nil {
		return nil
	}
	if m.ProfilePath() == "" {
		return errors.New("无法获取 profile 路径")
	}
	return m.userJS().Write(key, value)
}

// Reset 从 user.js 删除 pref。
func (m *PrefsManager) Reset(key string) error {
	if m == nil || m.ProfilePath() == "" {
		return nil
	}
	return m.userJS().Remove(key)
}

// GetAll 读取指定前缀的 user.js pref。
func (m *PrefsManager) GetAll(prefix string) (map[string]any, error) {
	if m == nil {
		return map[string]any{}, nil
	}
	return m.userJS().ReadPrefix(prefix)
}

// SaveToProfile 为兼容占位；user.js 本身已是持久化文件。
func (m *PrefsManager) SaveToProfile() error {
	return nil
}

func (m *PrefsManager) userJS() *support.JSPrefsFile {
	if m == nil || m.ProfilePath() == "" {
		return support.NewJSPrefsFile("")
	}
	return support.NewJSPrefsFile(filepath.Join(m.ProfilePath(), "user.js"))
}
