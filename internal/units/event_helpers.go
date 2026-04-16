package units

import (
	"fmt"
	"strings"
	"time"
)

func normalizeUnitStringList(value any, field string, required bool) ([]string, error) {
	if value == nil {
		if required {
			return nil, fmt.Errorf("%s 参数不能为空", field)
		}
		return nil, nil
	}

	var values []string
	switch typed := value.(type) {
	case string:
		text := strings.TrimSpace(typed)
		if text != "" {
			values = []string{text}
		}
	case []string:
		values = make([]string, 0, len(typed))
		for _, item := range typed {
			text := strings.TrimSpace(item)
			if text != "" {
				values = append(values, text)
			}
		}
	case []any:
		values = make([]string, 0, len(typed))
		for _, item := range typed {
			text := strings.TrimSpace(stringifyNetworkValue(item))
			if text != "" {
				values = append(values, text)
			}
		}
	default:
		return nil, fmt.Errorf("%s 参数必须为 string、[]string 或 []any", field)
	}

	values = dedupeUnitStrings(values)
	if required && len(values) == 0 {
		return nil, fmt.Errorf("%s 参数不能为空", field)
	}
	return values, nil
}

func dedupeUnitStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func resolveUnitTimeout(owner networkOwner) time.Duration {
	if owner == nil {
		return networkDefaultTimeout()
	}
	timeout := owner.BaseTimeout()
	if timeout <= 0 {
		return networkDefaultTimeout()
	}
	return timeout
}

func sourceContext(params map[string]any) string {
	source, _ := params["source"].(map[string]any)
	return stringifyNetworkValue(source["context"])
}

func cloneMapFromAny(value any) map[string]any {
	mapped, _ := value.(map[string]any)
	return cloneNetworkMapDeep(mapped)
}
