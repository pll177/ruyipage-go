package support

import (
	"fmt"
	"regexp"
	"strings"
)

var cssSelectorPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^\[`),
	regexp.MustCompile(`^\w+\[`),
	regexp.MustCompile(`^\w+\s*>`),
	regexp.MustCompile(`^\w+\s*\+`),
	regexp.MustCompile(`^\w+\s*~`),
	regexp.MustCompile(`^\*`),
	regexp.MustCompile(`^\w+:`),
}

// LocatorTuple 表示 Go 侧对 Python `(type, value)` 形式的等价承接。
type LocatorTuple struct {
	Type  string
	Value any
}

// NewLocatorTuple 创建二元定位器。
func NewLocatorTuple(locatorType string, locatorValue any) LocatorTuple {
	return LocatorTuple{Type: locatorType, Value: locatorValue}
}

// ParseLocator 将定位器输入解析为 BiDi locator 字典。
func ParseLocator(locator any) (map[string]any, error) {
	switch typed := locator.(type) {
	case string:
		return parseLocatorString(typed)
	case LocatorTuple:
		return parseLocatorTuple(typed.Type, typed.Value)
	case *LocatorTuple:
		if typed == nil {
			return nil, NewLocatorError("定位器不能为空", nil)
		}
		return parseLocatorTuple(typed.Type, typed.Value)
	case [2]string:
		return parseLocatorTuple(typed[0], typed[1])
	case [2]any:
		return parseLocatorPair(typed[0], typed[1], locator)
	case []string:
		if len(typed) != 2 {
			return nil, NewLocatorError(fmt.Sprintf("二元定位器必须是 (type, value) 格式: %#v", locator), nil)
		}
		return parseLocatorTuple(typed[0], typed[1])
	case []any:
		if len(typed) != 2 {
			return nil, NewLocatorError(fmt.Sprintf("二元定位器必须是 (type, value) 格式: %#v", locator), nil)
		}
		return parseLocatorPair(typed[0], typed[1], locator)
	default:
		return nil, NewLocatorError(fmt.Sprintf("定位器必须是字符串或二元定位器: %T", locator), nil)
	}
}

// LooksLikeCSSSelector 判断字符串是否看起来像 CSS 选择器。
func LooksLikeCSSSelector(value string) bool {
	for _, pattern := range cssSelectorPatterns {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}

// CSSValueEscape 转义 CSS 属性值中的引号。
func CSSValueEscape(value string) string {
	value = strings.ReplaceAll(value, `'`, `\'`)
	return strings.ReplaceAll(value, `"`, `\"`)
}

func parseLocatorPair(rawType any, rawValue any, original any) (map[string]any, error) {
	locatorType, ok := rawType.(string)
	if !ok {
		return nil, NewLocatorError(fmt.Sprintf("二元定位器的 type 必须是字符串: %#v", original), nil)
	}
	return parseLocatorTuple(locatorType, rawValue)
}

func parseLocatorTuple(locatorType string, locatorValue any) (map[string]any, error) {
	normalized := normalizeLocatorType(locatorType)
	switch normalized {
	case "css", "cssselector":
		return map[string]any{"type": "css", "value": locatorValue}, nil
	case "xpath":
		return map[string]any{"type": "xpath", "value": locatorValue}, nil
	case "text", "innertext":
		return map[string]any{"type": "innerText", "value": locatorValue}, nil
	case "accessibility":
		if attributes, ok := locatorValue.(map[string]any); ok {
			return map[string]any{"type": "accessibility", "value": cloneAnyMapDeep(attributes)}, nil
		}
		if attributes, ok := locatorValue.(map[string]string); ok {
			converted := make(map[string]any, len(attributes))
			for key, value := range attributes {
				converted[key] = value
			}
			return map[string]any{"type": "accessibility", "value": converted}, nil
		}
		return map[string]any{"type": "accessibility", "value": map[string]any{"name": locatorValue}}, nil
	default:
		return nil, NewLocatorError(fmt.Sprintf("不支持的定位器类型: %s", locatorType), nil)
	}
}

func parseLocatorString(locator string) (map[string]any, error) {
	locator = strings.TrimSpace(locator)
	if locator == "" {
		return nil, NewLocatorError("定位器不能为空", nil)
	}

	if strings.HasPrefix(locator, "css:") || strings.HasPrefix(locator, "c:") {
		prefix := "css:"
		if strings.HasPrefix(locator, "c:") {
			prefix = "c:"
		}
		return map[string]any{"type": "css", "value": strings.TrimSpace(locator[len(prefix):])}, nil
	}

	if strings.HasPrefix(locator, "xpath:") || strings.HasPrefix(locator, "x:") {
		prefix := "xpath:"
		if strings.HasPrefix(locator, "x:") {
			prefix = "x:"
		}
		return map[string]any{"type": "xpath", "value": strings.TrimSpace(locator[len(prefix):])}, nil
	}

	if strings.HasPrefix(locator, "/") || strings.HasPrefix(locator, "./") || strings.HasPrefix(locator, "(") {
		return map[string]any{"type": "xpath", "value": locator}, nil
	}

	if strings.HasPrefix(locator, "text=") {
		return map[string]any{
			"type":      "innerText",
			"value":     locator[5:],
			"matchType": "full",
		}, nil
	}

	if strings.HasPrefix(locator, "text:") {
		return map[string]any{"type": "innerText", "value": locator[5:]}, nil
	}

	if strings.HasPrefix(locator, "#") {
		return map[string]any{"type": "css", "value": locator}, nil
	}

	if strings.HasPrefix(locator, ".") && !strings.HasPrefix(locator, "./") {
		return map[string]any{"type": "css", "value": locator}, nil
	}

	if strings.HasPrefix(locator, "tag:") {
		return parseTagLocator(locator[4:])
	}

	if strings.HasPrefix(locator, "@@") {
		return parseMultiAttrLocator(locator, "")
	}

	if strings.HasPrefix(locator, "@") {
		return parseSingleAttrLocator(locator[1:], "")
	}

	if LooksLikeCSSSelector(locator) {
		return map[string]any{"type": "css", "value": locator}, nil
	}

	return map[string]any{"type": "innerText", "value": locator}, nil
}

func parseTagLocator(value string) (map[string]any, error) {
	doubleAt := strings.Index(value, "@@")
	singleAt := strings.Index(value, "@")

	switch {
	case doubleAt >= 0 && (singleAt < 0 || doubleAt <= singleAt):
		tag := strings.TrimSpace(value[:doubleAt])
		return parseMultiAttrLocator(value[doubleAt:], tag)
	case singleAt >= 0:
		tag := strings.TrimSpace(value[:singleAt])
		return parseSingleAttrLocator(value[singleAt+1:], tag)
	default:
		return map[string]any{"type": "css", "value": strings.TrimSpace(value)}, nil
	}
}

func parseSingleAttrLocator(value string, tag string) (map[string]any, error) {
	if strings.Contains(value, "=") {
		parts := strings.SplitN(value, "=", 2)
		attr := strings.TrimSpace(parts[0])
		attrValue := strings.TrimSpace(parts[1])

		if attr == "text()" {
			if tag != "" {
				return map[string]any{
					"type":  "xpath",
					"value": fmt.Sprintf("//%s[contains(text(), %s)]", tag, xpathStringLiteral(attrValue)),
				}, nil
			}
			return map[string]any{"type": "innerText", "value": attrValue}, nil
		}

		return map[string]any{
			"type":  "css",
			"value": fmt.Sprintf("%s[%s='%s']", tag, attr, CSSValueEscape(attrValue)),
		}, nil
	}

	attr := strings.TrimSpace(value)
	return map[string]any{
		"type":  "css",
		"value": fmt.Sprintf("%s[%s]", tag, attr),
	}, nil
}

func parseMultiAttrLocator(locator string, tag string) (map[string]any, error) {
	rawParts := strings.Split(locator, "@@")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		if part != "" {
			parts = append(parts, part)
		}
	}

	cssParts := make([]string, 0, len(parts))
	xpathParts := make([]string, 0, len(parts))
	hasText := false

	for _, part := range parts {
		if strings.Contains(part, "=") {
			attrParts := strings.SplitN(part, "=", 2)
			attr := strings.TrimSpace(attrParts[0])
			attrValue := strings.TrimSpace(attrParts[1])

			if attr == "text()" {
				hasText = true
				xpathParts = append(xpathParts, fmt.Sprintf("contains(text(), %s)", xpathStringLiteral(attrValue)))
				continue
			}

			cssParts = append(cssParts, fmt.Sprintf("[%s='%s']", attr, CSSValueEscape(attrValue)))
			xpathParts = append(xpathParts, fmt.Sprintf("@%s=%s", attr, xpathStringLiteral(attrValue)))
			continue
		}

		attr := strings.TrimSpace(part)
		cssParts = append(cssParts, fmt.Sprintf("[%s]", attr))
		xpathParts = append(xpathParts, "@"+attr)
	}

	if hasText {
		tagName := tag
		if tagName == "" {
			tagName = "*"
		}
		return map[string]any{
			"type":  "xpath",
			"value": fmt.Sprintf("//%s[%s]", tagName, strings.Join(xpathParts, " and ")),
		}, nil
	}

	return map[string]any{
		"type":  "css",
		"value": tag + strings.Join(cssParts, ""),
	}, nil
}

func normalizeLocatorType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "_", "")
	return value
}

func xpathStringLiteral(value string) string {
	switch {
	case !strings.Contains(value, `"`):
		return `"` + value + `"`
	case !strings.Contains(value, `'`):
		return `'` + value + `'`
	default:
		parts := strings.Split(value, `"`)
		quoted := make([]string, 0, len(parts)*2-1)
		for index, part := range parts {
			if part != "" {
				quoted = append(quoted, `"`+part+`"`)
			}
			if index < len(parts)-1 {
				quoted = append(quoted, `'"'`)
			}
		}
		return "concat(" + strings.Join(quoted, ", ") + ")"
	}
}
