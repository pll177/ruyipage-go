package pages

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"ruyipage-go/internal/bidi"
	"ruyipage-go/internal/support"
)

// HandleCloudflareChallenge 自动尝试处理常见 Cloudflare Turnstile 验证。
func (p *FirefoxBase) HandleCloudflareChallenge(timeout time.Duration, checkInterval time.Duration) bool {
	if p == nil {
		return false
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if checkInterval <= 0 {
		checkInterval = 2 * time.Second
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if !p.IsConnected() {
			return false
		}

		contextID, found := p.findCloudflareContext()
		if !found {
			time.Sleep(checkInterval)
			continue
		}

		clickX, clickY, ok := p.resolveCloudflareClickPoint(contextID)
		if !ok {
			time.Sleep(checkInterval)
			continue
		}

		_, err := bidi.PerformActions(
			p.browserDriver(),
			contextID,
			[]map[string]any{
				{
					"type":       "pointer",
					"id":         "mouse_cf",
					"parameters": map[string]any{"pointerType": "mouse"},
					"actions": []map[string]any{
						{"type": "pointerMove", "x": clickX, "y": clickY, "duration": 0},
						{"type": "pause", "duration": 50 + random.Intn(101)},
						{"type": "pointerDown", "button": 0},
						{"type": "pause", "duration": 80 + random.Intn(81)},
						{"type": "pointerUp", "button": 0},
					},
				},
			},
			p.baseTimeout(),
		)
		if err != nil {
			time.Sleep(checkInterval)
			continue
		}

		time.Sleep(3 * time.Second)
		if p.cloudflareAppearsPassed() {
			return true
		}
	}

	return false
}

func (p *FirefoxBase) findCloudflareContext() (string, bool) {
	tree, err := bidi.GetTree(p.browserDriver(), nil, "", p.baseTimeout())
	if err != nil {
		return "", false
	}
	return findCloudflareContextInTree(tree["contexts"])
}

func findCloudflareContextInTree(value any) (string, bool) {
	items, ok := value.([]any)
	if !ok {
		if typed, typedOK := value.([]map[string]any); typedOK {
			items = make([]any, 0, len(typed))
			for _, item := range typed {
				items = append(items, item)
			}
		}
	}
	for _, item := range items {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		urlText := strings.ToLower(strings.TrimSpace(anyString(row["url"])))
		if strings.Contains(urlText, "challenges.cloudflare.com") ||
			strings.Contains(urlText, "turnstile") ||
			strings.Contains(urlText, "cf-chl") {
			return anyString(row["context"]), true
		}
		if contextID, found := findCloudflareContextInTree(row["children"]); found {
			return contextID, true
		}
	}
	return "", false
}

func (p *FirefoxBase) resolveCloudflareClickPoint(contextID string) (int, int, bool) {
	result, err := bidi.Evaluate(
		p.browserDriver(),
		`(() => {
			const checkbox =
				document.querySelector('input[type="checkbox"]') ||
				document.querySelector('[role="checkbox"]') ||
				document.querySelector('label');
			if (checkbox) {
				const rect = checkbox.getBoundingClientRect();
				return {
					found: true,
					x: Math.round(rect.left + rect.width / 2),
					y: Math.round(rect.top + rect.height / 2)
				};
			}

			const rect = document.documentElement.getBoundingClientRect();
			return {
				found: false,
				w: Math.round(rect.width),
				h: Math.round(rect.height)
			};
		})()`,
		map[string]any{"context": contextID},
		boolPointer(false),
		"",
		"",
		nil,
		false,
		p.baseTimeout(),
	)
	if err != nil || !result.Success() {
		return 0, 0, false
	}

	decoded, _ := decodeScriptObjectResult(result)
	if truthy(decoded["found"]) {
		return intValue(decoded["x"]), intValue(decoded["y"]), true
	}
	width := intValue(decoded["w"])
	height := intValue(decoded["h"])
	if width <= 0 || height <= 0 {
		return 0, 0, false
	}
	return 35, height / 2, true
}

func (p *FirefoxBase) cloudflareAppearsPassed() bool {
	bodyValue, err := p.RunJS(`return document.body ? document.body.innerText : ""`)
	if err != nil {
		return false
	}
	bodyText := strings.ToLower(strings.TrimSpace(anyString(bodyValue)))
	if len(bodyText) > 200 && !strings.Contains(bodyText[:minInt(len(bodyText), 500)], "verify") {
		return true
	}

	if contextID, found := p.findCloudflareContext(); found && contextID != "" {
		return false
	}
	return true
}

func decodeScriptObjectResult(result bidi.ScriptResultData) (map[string]any, bool) {
	node := map[string]any{
		"type":  result.Result.Type,
		"value": result.Result.Value,
	}
	decoded, ok := support.ParseBiDiValue(node).(map[string]any)
	return decoded, ok
}

func boolPointer(value bool) *bool {
	return &value
}

func anyString(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func truthy(value any) bool {
	boolean, _ := value.(bool)
	return boolean
}

func intValue(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case float32:
		return int(typed)
	default:
		return 0
	}
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
