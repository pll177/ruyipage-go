package bidi

import (
	"fmt"
	"time"

	"ruyipage-go/internal/support"
)

// SessionStatusResult 表示 session.status 的结果。
type SessionStatusResult struct {
	Ready   bool
	Message string
}

// SessionNewResult 表示 session.new 的结果。
type SessionNewResult struct {
	SessionID    string
	Capabilities map[string]any
}

// SessionSubscribeResult 表示 session.subscribe 的结果。
type SessionSubscribeResult struct {
	Subscription string
}

type sessionCommandDriver interface {
	Run(method string, params map[string]any, timeout time.Duration) (map[string]any, error)
	SessionID() string
	SetSessionID(sessionID string)
}

// Status 查询远端当前 session 状态。
func Status(driver sessionCommandDriver, timeout time.Duration) (SessionStatusResult, error) {
	if driver == nil {
		return SessionStatusResult{}, support.NewPageDisconnectedError("session driver 未初始化", nil)
	}

	result, err := driver.Run("session.status", nil, timeout)
	if err != nil {
		return SessionStatusResult{}, err
	}

	return SessionStatusResult{
		Ready:   readBool(result, "ready"),
		Message: readString(result, "message"),
	}, nil
}

// New 创建新的 BiDi session，并在成功后同步 driver 上保存的 session id。
func New(
	driver sessionCommandDriver,
	capabilities map[string]any,
	userPromptHandler map[string]string,
	timeout time.Duration,
) (SessionNewResult, error) {
	if driver == nil {
		return SessionNewResult{}, support.NewPageDisconnectedError("session driver 未初始化", nil)
	}

	params := map[string]any{
		"capabilities": buildSessionCapabilities(capabilities, userPromptHandler),
	}

	result, err := driver.Run("session.new", params, timeout)
	if err != nil {
		return SessionNewResult{}, err
	}

	sessionResult := SessionNewResult{
		SessionID:    readString(result, "sessionId"),
		Capabilities: cloneAnyMapValue(result["capabilities"]),
	}
	driver.SetSessionID(sessionResult.SessionID)
	return sessionResult, nil
}

// End 结束当前 session；成功后会清空 driver 上保存的 session id。
func End(driver sessionCommandDriver, timeout time.Duration) error {
	if driver == nil {
		return support.NewPageDisconnectedError("session driver 未初始化", nil)
	}

	_, err := driver.Run("session.end", nil, timeout)
	if err != nil {
		return err
	}

	driver.SetSessionID("")
	return nil
}

// Subscribe 订阅指定事件。
func Subscribe(driver sessionCommandDriver, events any, contexts any, timeout time.Duration) (SessionSubscribeResult, error) {
	if driver == nil {
		return SessionSubscribeResult{}, support.NewPageDisconnectedError("session driver 未初始化", nil)
	}

	normalizedEvents, err := normalizeRequiredStringList(events, "events")
	if err != nil {
		return SessionSubscribeResult{}, err
	}

	params := map[string]any{
		"events": normalizedEvents,
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return SessionSubscribeResult{}, err
	} else if include {
		params["contexts"] = normalizedContexts
	}

	result, err := driver.Run("session.subscribe", params, timeout)
	if err != nil {
		return SessionSubscribeResult{}, err
	}

	return SessionSubscribeResult{
		Subscription: readString(result, "subscription"),
	}, nil
}

// Unsubscribe 取消指定事件订阅。
func Unsubscribe(driver sessionCommandDriver, events any, contexts any, subscriptions any, timeout time.Duration) error {
	if driver == nil {
		return support.NewPageDisconnectedError("session driver 未初始化", nil)
	}

	params := map[string]any{}
	if normalizedSubscriptions, include, err := normalizeOptionalStringList(subscriptions, "subscriptions"); err != nil {
		return err
	} else if include {
		params["subscriptions"] = normalizedSubscriptions
		_, err = driver.Run("session.unsubscribe", params, timeout)
		return err
	}

	if normalizedEvents, include, err := normalizeOptionalStringList(events, "events"); err != nil {
		return err
	} else if include {
		params["events"] = normalizedEvents
	}

	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return err
	} else if include {
		params["contexts"] = normalizedContexts
	}

	_, err := driver.Run("session.unsubscribe", params, timeout)
	return err
}

func buildSessionCapabilities(capabilities map[string]any, userPromptHandler map[string]string) map[string]any {
	caps := cloneAnyMap(capabilities)
	if caps == nil {
		caps = map[string]any{}
	}

	if len(userPromptHandler) == 0 {
		return caps
	}

	alwaysMatch := cloneAnyMapValue(caps["alwaysMatch"])
	if alwaysMatch == nil {
		alwaysMatch = map[string]any{}
	}
	alwaysMatch["unhandledPromptBehavior"] = cloneStringMap(userPromptHandler)
	caps["alwaysMatch"] = alwaysMatch
	return caps
}

func normalizeRequiredStringList(value any, field string) ([]string, error) {
	switch typed := value.(type) {
	case string:
		return []string{typed}, nil
	case []string:
		return cloneStringSlice(typed), nil
	default:
		return nil, fmt.Errorf("%s 参数必须为 string 或 []string", field)
	}
}

func normalizeOptionalStringList(value any, field string) ([]string, bool, error) {
	if value == nil {
		return nil, false, nil
	}

	switch typed := value.(type) {
	case string:
		if typed == "" {
			return nil, false, nil
		}
		return []string{typed}, true, nil
	case []string:
		if len(typed) == 0 {
			return nil, false, nil
		}
		return cloneStringSlice(typed), true, nil
	default:
		return nil, false, fmt.Errorf("%s 参数必须为 string 或 []string", field)
	}
}

func cloneAnyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}

	dst := make(map[string]any, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func cloneAnyMapValue(value any) map[string]any {
	src, ok := value.(map[string]any)
	if !ok || src == nil {
		return nil
	}
	return cloneAnyMap(src)
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}

	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func cloneStringSlice(src []string) []string {
	if src == nil {
		return nil
	}

	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

func readBool(values map[string]any, key string) bool {
	if values == nil {
		return false
	}

	value, _ := values[key].(bool)
	return value
}

func readString(values map[string]any, key string) string {
	if values == nil {
		return ""
	}

	value, _ := values[key].(string)
	return value
}
