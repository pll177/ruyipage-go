package support

import (
	"regexp"
	"strings"
)

var (
	validURLPattern = regexp.MustCompile(`(?i)^https?://(?:(?:[A-Z0-9](?:[A-Z0-9-]{0,61}[A-Z0-9])?\.)+[A-Z]{2,6}\.?|localhost|\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(?::\d+)?(?:/?|[/?]\S+)$`)
	httpURLScheme   = regexp.MustCompile(`(?i)^https?://`)
)

// IsValidURL 检查字符串是否为合法的 HTTP(S) URL。
func IsValidURL(url string) bool {
	return validURLPattern.MatchString(url)
}

// EnsureURL 确保 URL 带有 HTTP(S) 协议前缀。
func EnsureURL(url string) string {
	if url == "" {
		return url
	}
	if httpURLScheme.MatchString(url) {
		return url
	}
	if strings.HasPrefix(url, "//") {
		return "https:" + url
	}
	return "https://" + url
}
