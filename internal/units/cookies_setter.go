package units

import "ruyipage-go/internal/support"

type cookiesSetterOwner interface {
	SetCookies(cookies any) error
	DeleteCookies(filter map[string]any) error
}

// CookiesSetter 提供高层 Cookie 写入/删除入口。
type CookiesSetter struct {
	owner cookiesSetterOwner
}

// NewCookiesSetter 创建 Cookie setter。
func NewCookiesSetter(owner cookiesSetterOwner) *CookiesSetter {
	return &CookiesSetter{owner: owner}
}

// Set 设置 Cookie。
func (s *CookiesSetter) Set(cookies any) error {
	if s == nil || s.owner == nil {
		return nil
	}
	return s.owner.SetCookies(cookies)
}

// Remove 删除指定名称的 Cookie；可选限定 domain。
func (s *CookiesSetter) Remove(name string, domain string) error {
	if s == nil || s.owner == nil {
		return nil
	}
	if name == "" {
		return support.NewRuyiPageError("cookie 名称不能为空", nil)
	}

	filter := map[string]any{"name": name}
	if domain != "" {
		filter["domain"] = domain
	}
	return s.owner.DeleteCookies(filter)
}

// Clear 清空当前作用域内全部 Cookie。
func (s *CookiesSetter) Clear() error {
	if s == nil || s.owner == nil {
		return nil
	}
	return s.owner.DeleteCookies(nil)
}
