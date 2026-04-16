package adapter

import (
	"sort"
	"strings"
	"sync"
	"time"

	"ruyipage-go/internal/base"
	"ruyipage-go/internal/bidi"
	"ruyipage-go/internal/support"
)

const defaultContextEventTimeout = 10 * time.Second

var subscribeContextEvents = func(driver contextEventDriver, events any, contexts any, timeout time.Duration) (bidi.SessionSubscribeResult, error) {
	return bidi.Subscribe(driver, events, contexts, timeout)
}

// ContextInfo 表示一个 browsing context 的注册表快照。
type ContextInfo struct {
	ID       string
	URL      string
	Parent   string
	Children []string
}

// ContextRegistry 维护已知 context 的树形结构。
type ContextRegistry struct {
	mu       sync.RWMutex
	contexts map[string]ContextInfo
}

// NewContextRegistry 创建一个新的 context 注册表。
func NewContextRegistry() *ContextRegistry {
	return &ContextRegistry{
		contexts: map[string]ContextInfo{},
	}
}

// Register 写入或更新一个 context。
func (r *ContextRegistry) Register(contextID string, url string, parent string, children []string) {
	if r == nil || contextID == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, existed := r.contexts[contextID]
	info := ContextInfo{
		ID:       contextID,
		URL:      url,
		Parent:   parent,
		Children: cloneStringSlice(children),
	}

	if existed && len(info.Children) == 0 {
		info.Children = cloneStringSlice(existing.Children)
	}
	r.contexts[contextID] = info

	if existed && existing.Parent != "" && existing.Parent != parent {
		r.removeChildLocked(existing.Parent, contextID)
	}
	if parent != "" {
		r.appendChildLocked(parent, contextID)
	}
}

// Unregister 删除一个 context，并从父节点 children 中移除。
func (r *ContextRegistry) Unregister(contextID string) {
	if r == nil || contextID == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	info, exists := r.contexts[contextID]
	if !exists {
		return
	}
	delete(r.contexts, contextID)

	if info.Parent != "" {
		r.removeChildLocked(info.Parent, contextID)
	}
	for parentID, parent := range r.contexts {
		filtered := filterOut(parent.Children, contextID)
		if len(filtered) != len(parent.Children) {
			parent.Children = filtered
			r.contexts[parentID] = parent
		}
	}
}

// UpdateURL 更新指定 context 的 URL。
func (r *ContextRegistry) UpdateURL(contextID string, url string) {
	if r == nil || contextID == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	info, exists := r.contexts[contextID]
	if !exists {
		return
	}
	info.URL = url
	r.contexts[contextID] = info
}

// Get 读取指定 context 信息。
func (r *ContextRegistry) Get(contextID string) (ContextInfo, bool) {
	if r == nil {
		return ContextInfo{}, false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	info, exists := r.contexts[contextID]
	if !exists {
		return ContextInfo{}, false
	}
	info.Children = cloneStringSlice(info.Children)
	return info, true
}

// Children 返回直接子 context 列表。
func (r *ContextRegistry) Children(contextID string) []string {
	info, exists := r.Get(contextID)
	if !exists {
		return []string{}
	}
	return info.Children
}

// FindByURL 按 URL 子串查找 context id。
func (r *ContextRegistry) FindByURL(pattern string) []string {
	if r == nil {
		return []string{}
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]string, 0)
	for contextID, info := range r.contexts {
		if pattern == "" || containsSubstring(info.URL, pattern) {
			result = append(result, contextID)
		}
	}
	sort.Strings(result)
	return result
}

// AllIDs 返回全部 context id。
func (r *ContextRegistry) AllIDs() []string {
	if r == nil {
		return []string{}
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]string, 0, len(r.contexts))
	for contextID := range r.contexts {
		result = append(result, contextID)
	}
	sort.Strings(result)
	return result
}

// SyncFromTree 使用 browsingContext.getTree 的结果整体刷新注册表。
func (r *ContextRegistry) SyncFromTree(treeContexts []map[string]any) {
	if r == nil {
		return
	}

	updated := map[string]ContextInfo{}
	for _, item := range treeContexts {
		walkContextTree(updated, item, "")
	}

	r.mu.Lock()
	r.contexts = updated
	r.mu.Unlock()
}

func (r *ContextRegistry) appendChildLocked(parentID string, childID string) {
	parent, exists := r.contexts[parentID]
	if !exists {
		parent = ContextInfo{ID: parentID, Children: []string{}}
	}
	for _, existing := range parent.Children {
		if existing == childID {
			r.contexts[parentID] = parent
			return
		}
	}
	parent.Children = append(parent.Children, childID)
	r.contexts[parentID] = parent
}

func (r *ContextRegistry) removeChildLocked(parentID string, childID string) {
	parent, exists := r.contexts[parentID]
	if !exists {
		return
	}
	parent.Children = filterOut(parent.Children, childID)
	r.contexts[parentID] = parent
}

func walkContextTree(updated map[string]ContextInfo, raw map[string]any, parent string) {
	contextID := readContextString(raw["context"])
	if contextID == "" {
		return
	}

	resolvedParent := parent
	if resolvedParent == "" {
		resolvedParent = readContextString(raw["parent"])
	}

	children := readContextChildren(raw["children"])
	info := ContextInfo{
		ID:       contextID,
		URL:      readContextString(raw["url"]),
		Parent:   resolvedParent,
		Children: cloneStringSlice(children),
	}
	updated[contextID] = info

	if resolvedParent != "" {
		parentInfo := updated[resolvedParent]
		parentInfo.ID = resolvedParent
		parentInfo.Children = appendUniqueChild(parentInfo.Children, contextID)
		updated[resolvedParent] = parentInfo
	}

	rawChildren, _ := raw["children"].([]map[string]any)
	if len(rawChildren) == 0 {
		for _, child := range readMapSlice(raw["children"]) {
			walkContextTree(updated, child, contextID)
		}
		return
	}
	for _, child := range rawChildren {
		walkContextTree(updated, child, contextID)
	}
}

type contextEventDriver interface {
	Run(method string, params map[string]any, timeout time.Duration) (map[string]any, error)
	SessionID() string
	SetSessionID(sessionID string)
	SetCallback(event string, callback base.EventCallback, context string, immediate bool) error
	RemoveCallback(event string, context string, immediate bool)
}

// ContextEventAdapter 订阅 BiDi context 生命周期事件并维护注册表。
type ContextEventAdapter struct {
	driver     contextEventDriver
	registry   *ContextRegistry
	timeout    time.Duration
	subscribed bool
}

// NewContextEventAdapter 创建一个新的 context 事件适配器。
func NewContextEventAdapter(driver contextEventDriver, registry *ContextRegistry) *ContextEventAdapter {
	return &ContextEventAdapter{
		driver:   driver,
		registry: registry,
		timeout:  defaultContextEventTimeout,
	}
}

// Start 开始订阅 context 生命周期事件。
func (a *ContextEventAdapter) Start() error {
	if a == nil || a.driver == nil || a.registry == nil {
		return support.NewPageDisconnectedError("ContextEventAdapter 未初始化", nil)
	}
	if a.subscribed {
		return nil
	}

	events := []string{
		"browsingContext.contextCreated",
		"browsingContext.contextDestroyed",
		"browsingContext.navigationStarted",
		"browsingContext.load",
	}
	if _, err := subscribeContextEvents(a.driver, events, nil, a.timeout); err != nil {
		return err
	}

	if err := a.driver.SetCallback("browsingContext.contextCreated", a.onCreated, "", false); err != nil {
		return err
	}
	if err := a.driver.SetCallback("browsingContext.contextDestroyed", a.onDestroyed, "", false); err != nil {
		a.driver.RemoveCallback("browsingContext.contextCreated", "", false)
		return err
	}
	if err := a.driver.SetCallback("browsingContext.navigationStarted", a.onNavigation, "", false); err != nil {
		a.driver.RemoveCallback("browsingContext.contextCreated", "", false)
		a.driver.RemoveCallback("browsingContext.contextDestroyed", "", false)
		return err
	}
	if err := a.driver.SetCallback("browsingContext.load", a.onNavigation, "", false); err != nil {
		a.driver.RemoveCallback("browsingContext.contextCreated", "", false)
		a.driver.RemoveCallback("browsingContext.contextDestroyed", "", false)
		a.driver.RemoveCallback("browsingContext.navigationStarted", "", false)
		return err
	}

	a.subscribed = true
	return nil
}

// Stop 取消事件回调注册。
func (a *ContextEventAdapter) Stop() {
	if a == nil || a.driver == nil || !a.subscribed {
		return
	}

	a.driver.RemoveCallback("browsingContext.contextCreated", "", false)
	a.driver.RemoveCallback("browsingContext.contextDestroyed", "", false)
	a.driver.RemoveCallback("browsingContext.navigationStarted", "", false)
	a.driver.RemoveCallback("browsingContext.load", "", false)
	a.subscribed = false
}

func (a *ContextEventAdapter) onCreated(params map[string]any) {
	contextID := readContextString(params["context"])
	if contextID == "" {
		return
	}

	a.registry.Register(
		contextID,
		readContextString(params["url"]),
		readContextString(params["parent"]),
		readContextChildren(params["children"]),
	)
}

func (a *ContextEventAdapter) onDestroyed(params map[string]any) {
	a.registry.Unregister(readContextString(params["context"]))
}

func (a *ContextEventAdapter) onNavigation(params map[string]any) {
	contextID := readContextString(params["context"])
	if contextID == "" {
		return
	}
	a.registry.UpdateURL(contextID, readContextString(params["url"]))
}

func readContextChildren(value any) []string {
	switch typed := value.(type) {
	case []string:
		return cloneStringSlice(typed)
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			switch child := item.(type) {
			case string:
				if child != "" {
					result = append(result, child)
				}
			case map[string]any:
				if contextID := readContextString(child["context"]); contextID != "" {
					result = append(result, contextID)
				}
			}
		}
		return result
	case []map[string]any:
		result := make([]string, 0, len(typed))
		for _, child := range typed {
			if contextID := readContextString(child["context"]); contextID != "" {
				result = append(result, contextID)
			}
		}
		return result
	default:
		return []string{}
	}
}

func readMapSlice(value any) []map[string]any {
	switch typed := value.(type) {
	case []map[string]any:
		result := make([]map[string]any, len(typed))
		copy(result, typed)
		return result
	case []any:
		result := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if mapped, ok := item.(map[string]any); ok {
				result = append(result, mapped)
			}
		}
		return result
	default:
		return []map[string]any{}
	}
}

func readContextString(value any) string {
	text, _ := value.(string)
	return text
}

func cloneStringSlice(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

func filterOut(values []string, target string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == target {
			continue
		}
		result = append(result, value)
	}
	return result
}

func appendUniqueChild(values []string, target string) []string {
	for _, value := range values {
		if value == target {
			return values
		}
	}
	return append(values, target)
}

func containsSubstring(text string, pattern string) bool {
	return strings.Contains(text, pattern)
}
