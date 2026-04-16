package units

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pll177/ruyipage-go/internal/bidi"
)

// DownloadEvent 表示单条下载事件快照。
type DownloadEvent struct {
	Method            string
	Params            map[string]any
	Context           string
	Navigation        string
	Timestamp         any
	URL               string
	SuggestedFilename string
	Status            string
}

// NewDownloadEvent 从原始事件参数构建下载事件。
func NewDownloadEvent(method string, params map[string]any) DownloadEvent {
	raw := cloneNetworkMapDeep(params)
	return DownloadEvent{
		Method:            method,
		Params:            raw,
		Context:           stringifyNetworkValue(raw["context"]),
		Navigation:        stringifyNetworkValue(raw["navigation"]),
		Timestamp:         cloneNetworkValueDeep(raw["timestamp"]),
		URL:               stringifyNetworkValue(raw["url"]),
		SuggestedFilename: firstNonEmptyString(raw["suggestedFilename"], raw["filename"], raw["downloadFileName"]),
		Status:            stringifyNetworkValue(raw["status"]),
	}
}

// DownloadsManager 提供页面级下载行为与事件管理能力。
type DownloadsManager struct {
	owner networkOwner

	mu             sync.RWMutex
	listening      bool
	subscriptionID string
	queue          *packetQueue[*DownloadEvent]
	events         []DownloadEvent
}

// NewDownloadsManager 创建下载管理器。
func NewDownloadsManager(owner networkOwner) *DownloadsManager {
	return &DownloadsManager{
		owner: owner,
		queue: newPacketQueue[*DownloadEvent](128),
	}
}

// Listening 返回当前是否正在监听下载事件。
func (m *DownloadsManager) Listening() bool {
	if m == nil {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.listening
}

// Events 返回当前缓存的下载事件副本。
func (m *DownloadsManager) Events() []DownloadEvent {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return cloneDownloadEvents(m.events)
}

// SetBehavior 设置下载行为。
func (m *DownloadsManager) SetBehavior(behavior string, path string, contexts any, userContexts any) error {
	if m == nil || m.owner == nil {
		return nil
	}

	normalizedContexts, err := normalizeUnitStringList(contexts, "contexts", false)
	if err != nil {
		return err
	}
	normalizedUserContexts, err := normalizeUnitStringList(userContexts, "userContexts", false)
	if err != nil {
		return err
	}
	if len(normalizedContexts) > 0 && len(normalizedUserContexts) > 0 {
		return fmt.Errorf("contexts 与 userContexts 不能同时设置")
	}
	if len(normalizedContexts) == 0 && len(normalizedUserContexts) == 0 {
		normalizedContexts = []string{m.owner.ContextID()}
	}

	_, err = bidi.SetDownloadBehavior(
		m.owner.BrowserDriver(),
		behavior,
		path,
		normalizedContexts,
		normalizedUserContexts,
		resolveUnitTimeout(m.owner),
	)
	return err
}

// SetPath 将当前页面下载策略设置为 allow，并指定下载目录。
func (m *DownloadsManager) SetPath(path string) error {
	if path == "" {
		path = "."
	}
	absPath, err := filepath.Abs(path)
	if err == nil {
		path = absPath
	}
	return m.SetBehavior("allow", path, nil, nil)
}

// Start 开始监听当前页面下载事件。
func (m *DownloadsManager) Start() error {
	if m == nil || m.owner == nil {
		return nil
	}

	m.Stop()
	m.Clear()

	result, err := bidi.Subscribe(
		m.owner.BrowserDriver(),
		[]string{"browsingContext.downloadWillBegin", "browsingContext.downloadEnd"},
		[]string{m.owner.ContextID()},
		resolveUnitTimeout(m.owner),
	)
	if err != nil {
		return err
	}

	callbackDriver := m.owner.Driver()
	if err := callbackDriver.SetCallback("browsingContext.downloadWillBegin", m.onDownloadWillBegin, false); err != nil {
		_ = bidi.Unsubscribe(m.owner.BrowserDriver(), nil, nil, []string{result.Subscription}, resolveUnitTimeout(m.owner))
		return err
	}
	if err := callbackDriver.SetCallback("browsingContext.downloadEnd", m.onDownloadEnd, false); err != nil {
		callbackDriver.RemoveCallback("browsingContext.downloadWillBegin", false)
		_ = bidi.Unsubscribe(m.owner.BrowserDriver(), nil, nil, []string{result.Subscription}, resolveUnitTimeout(m.owner))
		return err
	}

	m.mu.Lock()
	m.listening = true
	m.subscriptionID = result.Subscription
	m.queue = newPacketQueue[*DownloadEvent](128)
	m.events = nil
	m.mu.Unlock()
	return nil
}

// Stop 停止监听下载事件。
func (m *DownloadsManager) Stop() {
	if m == nil || m.owner == nil {
		return
	}

	m.mu.Lock()
	subscriptionID := m.subscriptionID
	wasListening := m.listening
	m.listening = false
	m.subscriptionID = ""
	m.mu.Unlock()

	if !wasListening {
		return
	}

	callbackDriver := m.owner.Driver()
	callbackDriver.RemoveCallback("browsingContext.downloadWillBegin", false)
	callbackDriver.RemoveCallback("browsingContext.downloadEnd", false)
	if subscriptionID != "" {
		_ = bidi.Unsubscribe(
			m.owner.BrowserDriver(),
			nil,
			nil,
			[]string{subscriptionID},
			resolveUnitTimeout(m.owner),
		)
	}
}

// Clear 清空已缓存下载事件与等待队列。
func (m *DownloadsManager) Clear() {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.events = nil
	queue := m.queue
	m.mu.Unlock()
	if queue != nil {
		queue.Clear()
	}
}

// Wait 等待一个匹配条件的下载事件。
func (m *DownloadsManager) Wait(method string, timeout time.Duration, filename string, status string) *DownloadEvent {
	if m == nil {
		return nil
	}
	if timeout <= 0 {
		timeout = resolveUnitTimeout(m.owner)
	}

	deadline := time.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil
		}
		item, ok := m.queue.Pull(remaining)
		if !ok || item == nil {
			return nil
		}
		if matchDownloadEvent(*item, method, filename, status) {
			event := cloneDownloadEvent(*item)
			return &event
		}
	}
}

// WaitChain 等待一次下载的开始与结束事件链。
func (m *DownloadsManager) WaitChain(filename string, timeout time.Duration) (*DownloadEvent, *DownloadEvent) {
	if m == nil {
		return nil, nil
	}
	if timeout <= 0 {
		timeout = resolveUnitTimeout(m.owner)
	}

	deadline := time.Now().Add(timeout)
	beginTimeout := time.Until(deadline)
	if beginTimeout <= 0 {
		return nil, nil
	}
	begin := m.Wait("browsingContext.downloadWillBegin", beginTimeout, filename, "")
	if begin == nil {
		return nil, nil
	}
	endTimeout := time.Until(deadline)
	if endTimeout <= 0 {
		return begin, nil
	}
	end := m.Wait("browsingContext.downloadEnd", endTimeout, filename, "")
	return begin, end
}

// FileExists 检查下载文件是否已稳定落盘。
func (m *DownloadsManager) FileExists(path string, minSize int64) bool {
	if minSize <= 0 {
		minSize = 1
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && info.Size() >= minSize
}

// WaitFile 等待下载文件落盘。
func (m *DownloadsManager) WaitFile(path string, timeout time.Duration, minSize int64) bool {
	if timeout <= 0 {
		timeout = resolveUnitTimeout(m.owner)
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if m.FileExists(path, minSize) {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return m.FileExists(path, minSize)
}

func (m *DownloadsManager) onDownloadWillBegin(params map[string]any) {
	m.push("browsingContext.downloadWillBegin", params)
}

func (m *DownloadsManager) onDownloadEnd(params map[string]any) {
	m.push("browsingContext.downloadEnd", params)
}

func (m *DownloadsManager) push(method string, params map[string]any) {
	if m == nil {
		return
	}
	if contextID := stringifyNetworkValue(params["context"]); contextID != "" && contextID != m.owner.ContextID() {
		return
	}

	event := NewDownloadEvent(method, params)

	m.mu.Lock()
	if !m.listening {
		m.mu.Unlock()
		return
	}
	m.events = append(m.events, cloneDownloadEvent(event))
	queue := m.queue
	m.mu.Unlock()

	copied := cloneDownloadEvent(event)
	if queue != nil {
		queue.Push(&copied)
	}
}

func cloneDownloadEvents(values []DownloadEvent) []DownloadEvent {
	if len(values) == 0 {
		return nil
	}
	result := make([]DownloadEvent, len(values))
	for index, value := range values {
		result[index] = cloneDownloadEvent(value)
	}
	return result
}

func cloneDownloadEvent(event DownloadEvent) DownloadEvent {
	return DownloadEvent{
		Method:            event.Method,
		Params:            cloneNetworkMapDeep(event.Params),
		Context:           event.Context,
		Navigation:        event.Navigation,
		Timestamp:         cloneNetworkValueDeep(event.Timestamp),
		URL:               event.URL,
		SuggestedFilename: event.SuggestedFilename,
		Status:            event.Status,
	}
}

func matchDownloadEvent(event DownloadEvent, method string, filename string, status string) bool {
	if method != "" && event.Method != method {
		return false
	}
	if filename != "" && event.SuggestedFilename != filename {
		return false
	}
	if status != "" && event.Status != status {
		return false
	}
	return true
}

func firstNonEmptyString(values ...any) string {
	for _, value := range values {
		if text := stringifyNetworkValue(value); text != "" {
			return text
		}
	}
	return ""
}
