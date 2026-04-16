package support

import (
	"fmt"
	"sort"
	"strings"
)

// CookiesToDict 将 Cookie 列表转换为 name -> value 字典。
func CookiesToDict(cookies []map[string]any) map[string]string {
	result := make(map[string]string, len(cookies))
	for _, cookie := range cookies {
		name := strings.TrimSpace(fmt.Sprint(cookie["name"]))
		if name == "" {
			continue
		}

		value := cookie["value"]
		if valueMap, ok := value.(map[string]any); ok {
			value = ParseBiDiValue(valueMap)
		}
		if value == nil {
			result[name] = ""
			continue
		}
		result[name] = fmt.Sprint(value)
	}
	return result
}

// DictToCookies 将 name -> value 字典转换为 Cookie 列表。
func DictToCookies(cookieDict map[string]any, domain string) []map[string]any {
	if len(cookieDict) == 0 {
		return []map[string]any{}
	}

	keys := make([]string, 0, len(cookieDict))
	for key := range cookieDict {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]map[string]any, 0, len(keys))
	for _, key := range keys {
		result = append(result, map[string]any{
			"name":   key,
			"value":  fmt.Sprint(cookieDict[key]),
			"domain": domain,
		})
	}
	return result
}

// CookieStrToList 将 document.cookie 风格字符串转换为 Cookie 列表。
func CookieStrToList(cookieStr string) []map[string]any {
	if cookieStr == "" {
		return []map[string]any{}
	}

	result := make([]map[string]any, 0)
	for _, pair := range strings.Split(cookieStr, ";") {
		pair = strings.TrimSpace(pair)
		if !strings.Contains(pair, "=") {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		result = append(result, map[string]any{
			"name":  strings.TrimSpace(parts[0]),
			"value": strings.TrimSpace(parts[1]),
		})
	}
	return result
}
