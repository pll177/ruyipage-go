package units

import (
	"fmt"

	"ruyipage-go/internal/support"
)

// CookieInfo 表示单个 Cookie 的公开信息。
type CookieInfo struct {
	Raw      map[string]any
	Name     string
	Value    string
	Domain   string
	Path     string
	HTTPOnly bool
	Secure   bool
	SameSite string
	Expiry   any
}

// NewCookieInfo 从原始 Cookie 数据创建 CookieInfo。
func NewCookieInfo(data map[string]any) CookieInfo {
	raw := cloneAnyMapDeep(data)
	return CookieInfo{
		Raw:      raw,
		Name:     cookieString(raw["name"]),
		Value:    cookieValueString(raw["value"]),
		Domain:   cookieString(raw["domain"]),
		Path:     cookieString(raw["path"]),
		HTTPOnly: cookieBool(raw["httpOnly"]),
		Secure:   cookieBool(raw["secure"]),
		SameSite: cookieString(raw["sameSite"]),
		Expiry:   cloneAnyValueDeep(raw["expiry"]),
	}
}

// NewCookieInfos 批量转换 Cookie 列表。
func NewCookieInfos(data []map[string]any) []CookieInfo {
	if len(data) == 0 {
		return []CookieInfo{}
	}

	result := make([]CookieInfo, 0, len(data))
	for _, item := range data {
		result = append(result, NewCookieInfo(item))
	}
	return result
}

// NewCookieInfosFromAny 从任意 Cookie 切片值批量转换。
func NewCookieInfosFromAny(value any) []CookieInfo {
	switch typed := value.(type) {
	case []map[string]any:
		return NewCookieInfos(typed)
	case []any:
		items := make([]map[string]any, 0, len(typed))
		for _, entry := range typed {
			if mapped, ok := entry.(map[string]any); ok {
				items = append(items, mapped)
			}
		}
		return NewCookieInfos(items)
	default:
		return []CookieInfo{}
	}
}

func cookieValueString(value any) string {
	if mapped, ok := value.(map[string]any); ok {
		value = support.ParseBiDiValue(mapped)
	}
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func cookieString(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func cookieBool(value any) bool {
	typed, _ := value.(bool)
	return typed
}
