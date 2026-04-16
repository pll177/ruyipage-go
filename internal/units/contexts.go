package units

import (
	"time"

	"github.com/pll177/ruyipage-go/internal/base"
	"github.com/pll177/ruyipage-go/internal/bidi"
)

type contextManagerOwner interface {
	ContextID() string
	BrowserDriver() *base.BrowserBiDiDriver
	BaseTimeout() time.Duration
}

// ContextInfo 表示单个 browsingContext 快照。
type ContextInfo struct {
	Raw            map[string]any
	Context        string
	URL            string
	UserContext    string
	Parent         string
	OriginalOpener any
	ClientWindow   string
	Children       []ContextInfo
}

// ContextTree 表示 browsingContext.getTree 返回的高层结果。
type ContextTree struct {
	Raw      map[string]any
	Contexts []ContextInfo
}

// ContextManager 提供 browsingContext 与 user context 的高层操作入口。
type ContextManager struct {
	owner contextManagerOwner
}

// NewContextManager 创建上下文管理器。
func NewContextManager(owner contextManagerOwner) *ContextManager {
	return &ContextManager{owner: owner}
}

// NewContextInfoFromData 从原始协议字段构建 ContextInfo。
func NewContextInfoFromData(data map[string]any) ContextInfo {
	raw := cloneNetworkMapDeep(data)
	children := anyContextChildren(raw["children"])
	return ContextInfo{
		Raw:            raw,
		Context:        stringifyNetworkValue(raw["context"]),
		URL:            stringifyNetworkValue(raw["url"]),
		UserContext:    stringifyNetworkValue(raw["userContext"]),
		Parent:         stringifyNetworkValue(raw["parent"]),
		OriginalOpener: cloneNetworkValueDeep(raw["originalOpener"]),
		ClientWindow:   stringifyNetworkValue(raw["clientWindow"]),
		Children:       children,
	}
}

// NewContextTreeFromData 从原始协议字段构建 ContextTree。
func NewContextTreeFromData(data map[string]any) ContextTree {
	raw := cloneNetworkMapDeep(data)
	return ContextTree{
		Raw:      raw,
		Contexts: anyContextInfos(raw["contexts"]),
	}
}

// GetTree 返回当前浏览器的 browsing context 树。
func (m *ContextManager) GetTree(maxDepth *int, root string) (ContextTree, error) {
	if m == nil || m.owner == nil {
		return ContextTree{}, nil
	}
	result, err := bidi.GetTree(m.owner.BrowserDriver(), maxDepth, root, m.owner.BaseTimeout())
	if err != nil {
		return ContextTree{}, err
	}
	return NewContextTreeFromData(result), nil
}

// CreateTab 创建新的 tab；userContext 为空时使用默认上下文。
func (m *ContextManager) CreateTab(background bool, userContext string, referenceContext string) (string, error) {
	if m == nil || m.owner == nil {
		return "", nil
	}
	result, err := bidi.Create(
		m.owner.BrowserDriver(),
		"tab",
		referenceContext,
		background,
		userContext,
		m.owner.BaseTimeout(),
	)
	if err != nil {
		return "", err
	}
	return stringifyNetworkValue(result["context"]), nil
}

// CreateWindow 创建新的 window；userContext 为空时使用默认上下文。
func (m *ContextManager) CreateWindow(background bool, userContext string) (string, error) {
	if m == nil || m.owner == nil {
		return "", nil
	}
	result, err := bidi.Create(
		m.owner.BrowserDriver(),
		"window",
		"",
		background,
		userContext,
		m.owner.BaseTimeout(),
	)
	if err != nil {
		return "", err
	}
	return stringifyNetworkValue(result["context"]), nil
}

// Close 关闭指定或当前 context。
func (m *ContextManager) Close(context string, promptUnload bool) error {
	if m == nil || m.owner == nil {
		return nil
	}
	target := context
	if target == "" {
		target = m.owner.ContextID()
	}
	_, err := bidi.Close(m.owner.BrowserDriver(), target, promptUnload, m.owner.BaseTimeout())
	return err
}

// Reload 重载指定或当前 context。
func (m *ContextManager) Reload(ignoreCache bool, wait string, context string) (map[string]any, error) {
	if m == nil || m.owner == nil {
		return map[string]any{}, nil
	}
	target := context
	if target == "" {
		target = m.owner.ContextID()
	}
	return bidi.Reload(m.owner.BrowserDriver(), target, ignoreCache, wait, m.owner.BaseTimeout())
}

// SetViewport 设置指定或当前 context 的 viewport。
func (m *ContextManager) SetViewport(width int, height int, devicePixelRatio *float64, context string) error {
	if m == nil || m.owner == nil {
		return nil
	}
	target := context
	if target == "" {
		target = m.owner.ContextID()
	}
	_, err := bidi.SetViewport(m.owner.BrowserDriver(), target, &width, &height, devicePixelRatio, m.owner.BaseTimeout())
	return err
}

// SetBypassCSP 调用 browsingContext.setBypassCSP 设置当前 context 的 CSP 绕过。
func (m *ContextManager) SetBypassCSP(enabled bool, context string) error {
	if m == nil || m.owner == nil {
		return nil
	}
	target := context
	if target == "" {
		target = m.owner.ContextID()
	}
	_, err := bidi.SetBypassCSP(m.owner.BrowserDriver(), target, enabled, m.owner.BaseTimeout())
	return err
}

// CreateUserContext 创建新的 browser user context。
func (m *ContextManager) CreateUserContext() (string, error) {
	if m == nil || m.owner == nil {
		return "", nil
	}
	result, err := bidi.CreateUserContext(m.owner.BrowserDriver(), m.owner.BaseTimeout())
	if err != nil {
		return "", err
	}
	return stringifyNetworkValue(result["userContext"]), nil
}

// GetUserContexts 返回当前浏览器全部 user context 信息。
func (m *ContextManager) GetUserContexts() ([]map[string]any, error) {
	if m == nil || m.owner == nil {
		return []map[string]any{}, nil
	}
	result, err := bidi.GetUserContexts(m.owner.BrowserDriver(), m.owner.BaseTimeout())
	if err != nil {
		return nil, err
	}
	return anyContextRows(result["userContexts"]), nil
}

// RemoveUserContext 删除指定 browser user context。
func (m *ContextManager) RemoveUserContext(userContext string) error {
	if m == nil || m.owner == nil {
		return nil
	}
	_, err := bidi.RemoveUserContext(m.owner.BrowserDriver(), userContext, m.owner.BaseTimeout())
	return err
}

// GetClientWindows 返回当前浏览器的 client window 信息。
func (m *ContextManager) GetClientWindows() ([]map[string]any, error) {
	if m == nil || m.owner == nil {
		return []map[string]any{}, nil
	}
	result, err := bidi.GetClientWindows(m.owner.BrowserDriver(), m.owner.BaseTimeout())
	if err != nil {
		return nil, err
	}
	return anyContextRows(result["clientWindows"]), nil
}

// SetWindowState 设置指定 client window 的状态和几何信息。
func (m *ContextManager) SetWindowState(
	clientWindow string,
	state string,
	width *int,
	height *int,
	x *int,
	y *int,
) error {
	if m == nil || m.owner == nil {
		return nil
	}
	_, err := bidi.SetClientWindowState(
		m.owner.BrowserDriver(),
		clientWindow,
		state,
		width,
		height,
		x,
		y,
		m.owner.BaseTimeout(),
	)
	return err
}

func anyContextChildren(value any) []ContextInfo {
	rows := anyContextRows(value)
	if len(rows) == 0 {
		return nil
	}
	result := make([]ContextInfo, 0, len(rows))
	for _, row := range rows {
		result = append(result, NewContextInfoFromData(row))
	}
	return result
}

func anyContextInfos(value any) []ContextInfo {
	rows := anyContextRows(value)
	if len(rows) == 0 {
		return nil
	}
	result := make([]ContextInfo, 0, len(rows))
	for _, row := range rows {
		result = append(result, NewContextInfoFromData(row))
	}
	return result
}

func anyContextRows(value any) []map[string]any {
	switch typed := value.(type) {
	case nil:
		return []map[string]any{}
	case []map[string]any:
		result := make([]map[string]any, 0, len(typed))
		for _, row := range typed {
			result = append(result, cloneNetworkMapDeep(row))
		}
		return result
	case []any:
		result := make([]map[string]any, 0, len(typed))
		for _, row := range typed {
			mapped, ok := row.(map[string]any)
			if !ok {
				continue
			}
			result = append(result, cloneNetworkMapDeep(mapped))
		}
		return result
	default:
		return []map[string]any{}
	}
}
