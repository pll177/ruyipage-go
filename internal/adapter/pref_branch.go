package adapter

import (
	"errors"
	"path/filepath"

	"github.com/pll177/ruyipage-go/internal/support"
)

// PrefBranch 是 nsIPrefBranch 的文件/Marionette 适配层。
type PrefBranch struct {
	profilePath string
	host        string
	port        int

	marionette *MarionetteClient
}

// NewPrefBranch 创建一个 pref 分支适配器。
func NewPrefBranch(profilePath string) *PrefBranch {
	return NewPrefBranchWithMarionette(profilePath, DefaultMarionetteHost, DefaultMarionettePort)
}

// NewPrefBranchWithMarionette 创建带指定 Marionette 地址的 pref 分支适配器。
func NewPrefBranchWithMarionette(profilePath string, host string, port int) *PrefBranch {
	return &PrefBranch{
		profilePath: profilePath,
		host:        host,
		port:        port,
	}
}

// Get 读取 pref，优先尝试 Marionette，再回退到 user.js。
func (p *PrefBranch) Get(key string) (any, bool, error) {
	if p == nil || key == "" {
		return nil, false, nil
	}

	if client := p.marionetteClient(); client != nil {
		if value, ok, err := client.GetPref(key); err == nil && ok {
			return value, true, nil
		}
	}

	return p.userJS().Read(key)
}

// GetAll 从 user.js 读取指定前缀的全部 pref。
func (p *PrefBranch) GetAll(prefix string) (map[string]any, error) {
	if p == nil {
		return map[string]any{}, nil
	}
	return p.userJS().ReadPrefix(prefix)
}

// Set 写入 pref 到 user.js。
func (p *PrefBranch) Set(key string, value any) error {
	if p == nil {
		return nil
	}
	if p.profilePath == "" {
		return errors.New("未设置 profile 路径，无法写入 pref")
	}
	return p.userJS().Write(key, value)
}

// Reset 从 user.js 删除 pref。
func (p *PrefBranch) Reset(key string) error {
	if p == nil || p.profilePath == "" {
		return nil
	}
	return p.userJS().Remove(key)
}

func (p *PrefBranch) userJS() *support.JSPrefsFile {
	if p == nil || p.profilePath == "" {
		return support.NewJSPrefsFile("")
	}
	return support.NewJSPrefsFile(filepath.Join(p.profilePath, "user.js"))
}

func (p *PrefBranch) marionetteClient() *MarionetteClient {
	if p == nil {
		return nil
	}
	if p.marionette == nil {
		p.marionette = NewMarionetteClient(p.host, p.port)
	}
	return p.marionette
}
