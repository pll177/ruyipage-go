package units

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"

	"ruyipage-go/internal/adapter"
	"ruyipage-go/internal/support"
)

// ProfileResolver 按需解析当前 profile 路径。
type ProfileResolver func() string

// PageRefresher 表示可触发页面刷新的一方。
type PageRefresher interface {
	Refresh() error
}

// ConfigManager 提供 about:config 的完整文件/运行时控制能力。
type ConfigManager struct {
	resolveProfile ProfileResolver
	refresher      PageRefresher
	marionetteHost string
	marionettePort int
}

// NewConfigManager 创建绑定固定 profile 路径的配置管理器。
func NewConfigManager(profilePath string) *ConfigManager {
	return NewConfigManagerWithResolver(func() string {
		return profilePath
	})
}

// NewConfigManagerWithResolver 创建按需解析 profile 路径的配置管理器。
func NewConfigManagerWithResolver(resolver ProfileResolver) *ConfigManager {
	if resolver == nil {
		resolver = func() string { return "" }
	}
	return &ConfigManager{
		resolveProfile: resolver,
		marionetteHost: adapter.DefaultMarionetteHost,
		marionettePort: adapter.DefaultMarionettePort,
	}
}

// SetRefresher 设置默认刷新器。
func (m *ConfigManager) SetRefresher(refresher PageRefresher) {
	if m == nil {
		return
	}
	m.refresher = refresher
}

// SetMarionetteAddress 设置 Marionette 地址。
func (m *ConfigManager) SetMarionetteAddress(host string, port int) {
	if m == nil {
		return
	}
	if host != "" {
		m.marionetteHost = host
	}
	if port > 0 {
		m.marionettePort = port
	}
}

// ProfilePath 返回当前解析到的 profile 路径。
func (m *ConfigManager) ProfilePath() string {
	if m == nil || m.resolveProfile == nil {
		return ""
	}
	return m.resolveProfile()
}

// Get 读取 pref，优先级为 Marionette → user.js → prefs.js。
func (m *ConfigManager) Get(key string) (any, error) {
	if m == nil || key == "" {
		return nil, nil
	}

	branch := adapter.NewPrefBranchWithMarionette(m.ProfilePath(), m.marionetteHost, m.marionettePort)
	if value, ok, err := branch.Get(key); err != nil {
		return nil, err
	} else if ok {
		return value, nil
	}

	value, ok, err := m.prefsJS().Read(key)
	if err != nil || !ok {
		return nil, err
	}
	return value, nil
}

// GetActual 从 prefs.js 读取 Firefox 实际运行值。
func (m *ConfigManager) GetActual(key string) (any, error) {
	if m == nil || key == "" {
		return nil, nil
	}

	value, ok, err := m.prefsJS().Read(key)
	if err != nil || !ok {
		return nil, err
	}
	return value, nil
}

// GetAll 读取 user.js 中匹配前缀的全部 pref。
func (m *ConfigManager) GetAll(prefix string) (map[string]any, error) {
	if m == nil {
		return map[string]any{}, nil
	}
	return m.userJS().ReadPrefix(prefix)
}

// Set 写入 user.js。
func (m *ConfigManager) Set(key string, value any) error {
	if m == nil {
		return nil
	}
	if m.ProfilePath() == "" {
		return errors.New("未设置 profile 路径")
	}
	return m.userJS().Write(key, value)
}

// SetMany 批量写入 user.js。
func (m *ConfigManager) SetMany(prefs map[string]any) error {
	if m == nil || len(prefs) == 0 {
		return nil
	}
	if m.ProfilePath() == "" {
		return errors.New("未设置 profile 路径")
	}
	return m.userJS().WriteMany(prefs)
}

// Reset 从 user.js 删除 pref。
func (m *ConfigManager) Reset(key string) error {
	if m == nil || m.ProfilePath() == "" {
		return nil
	}
	return m.userJS().Remove(key)
}

// Lock 通过 policies.json 锁定 pref。
func (m *ConfigManager) Lock(key string, value any) error {
	if m == nil {
		return nil
	}
	if m.ProfilePath() == "" {
		return errors.New("未设置 profile 路径")
	}
	return m.policies().SetLockedPref(key, value)
}

// Unlock 从 policies.json 中解锁 pref。
func (m *ConfigManager) Unlock(key string) error {
	if m == nil || m.ProfilePath() == "" {
		return nil
	}
	return m.policies().UnlockPref(key)
}

// ApplyNow 写入 user.js 并尝试触发刷新。
func (m *ConfigManager) ApplyNow(key string, value any, refresher ...PageRefresher) error {
	if err := m.Set(key, value); err != nil {
		return err
	}
	m.tryRefresh(refresher...)
	return nil
}

// ApplyManyNow 批量写入 user.js 并尝试触发刷新。
func (m *ConfigManager) ApplyManyNow(prefs map[string]any, refresher ...PageRefresher) error {
	if err := m.SetMany(prefs); err != nil {
		return err
	}
	m.tryRefresh(refresher...)
	return nil
}

// Isolate 创建一个独立 profile，并复制当前 user.js。
func (m *ConfigManager) Isolate(baseDir string) (string, error) {
	newProfile, err := os.MkdirTemp(baseDir, "ruyipage_")
	if err != nil {
		return "", err
	}

	profilePath := m.ProfilePath()
	if profilePath == "" {
		return newProfile, nil
	}

	source := filepath.Join(profilePath, "user.js")
	if _, err := os.Stat(source); err != nil {
		if os.IsNotExist(err) {
			return newProfile, nil
		}
		return "", err
	}

	content, err := os.ReadFile(source)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(newProfile, "user.js"), content, 0o644); err != nil {
		return "", err
	}
	return newProfile, nil
}

// Diff 比较 user.js 与 prefs.js 的差异。
func (m *ConfigManager) Diff() (map[string]map[string]any, error) {
	userPrefs, err := m.userJS().ReadAll()
	if err != nil {
		return nil, err
	}
	actualPrefs, err := m.prefsJS().ReadAll()
	if err != nil {
		return nil, err
	}

	result := map[string]map[string]any{}
	keys := map[string]struct{}{}
	for key := range userPrefs {
		keys[key] = struct{}{}
	}
	for key := range actualPrefs {
		keys[key] = struct{}{}
	}

	for key := range keys {
		userValue, userOK := userPrefs[key]
		actualValue, actualOK := actualPrefs[key]
		if userOK && actualOK && reflect.DeepEqual(userValue, actualValue) {
			continue
		}
		result[key] = map[string]any{
			"user":   userValue,
			"actual": actualValue,
		}
	}
	return result, nil
}

func (m *ConfigManager) tryRefresh(refresher ...PageRefresher) {
	target := m.refresher
	if len(refresher) > 0 && refresher[0] != nil {
		target = refresher[0]
	}
	if target == nil {
		return
	}
	_ = target.Refresh()
}

func (m *ConfigManager) userJS() *support.JSPrefsFile {
	if m == nil || m.ProfilePath() == "" {
		return support.NewJSPrefsFile("")
	}
	return support.NewJSPrefsFile(filepath.Join(m.ProfilePath(), "user.js"))
}

func (m *ConfigManager) prefsJS() *support.JSPrefsFile {
	if m == nil || m.ProfilePath() == "" {
		return support.NewJSPrefsFile("")
	}
	return support.NewJSPrefsFile(filepath.Join(m.ProfilePath(), "prefs.js"))
}

func (m *ConfigManager) policies() *support.PoliciesFile {
	if m == nil || m.ProfilePath() == "" {
		return support.NewPoliciesFile("")
	}
	return support.NewPoliciesFileFromProfile(m.ProfilePath())
}
