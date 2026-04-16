package units

import (
	"strings"
	"sync"
	"time"

	"github.com/pll177/ruyipage-go/internal/bidi"
)

// LogEntry 表示单条 console / javascript 日志。
type LogEntry struct {
	Level      string
	Text       string
	Timestamp  any
	Source     map[string]any
	LogType    string
	Method     string
	Args       []any
	StackTrace any
}

// NewLogEntry 从原始事件参数构建日志条目。
func NewLogEntry(params map[string]any) LogEntry {
	return NewLogEntryFromData(bidi.ParseLogEntryData(params))
}

// NewLogEntryFromData 从底层日志数据构建公开日志条目。
func NewLogEntryFromData(data bidi.LogEntryData) LogEntry {
	return LogEntry{
		Level:      data.Level,
		Text:       data.Text,
		Timestamp:  cloneNetworkValueDeep(data.Timestamp),
		Source:     cloneNetworkMapDeep(data.Source),
		LogType:    data.LogType,
		Method:     data.Method,
		Args:       cloneLogArgs(data.Args),
		StackTrace: cloneNetworkValueDeep(data.StackTrace),
	}
}

// ConsoleListener 提供页面级控制台日志监听能力。
type ConsoleListener struct {
	owner networkOwner

	mu             sync.RWMutex
	listening      bool
	subscriptionID string
	levelFilter    string
	queue          *packetQueue[*LogEntry]
	entries        []LogEntry
	realtime       func(LogEntry)
}

// NewConsoleListener 创建控制台监听器。
func NewConsoleListener(owner networkOwner) *ConsoleListener {
	return &ConsoleListener{
		owner: owner,
		queue: newPacketQueue[*LogEntry](128),
	}
}

// Listening 返回当前是否正在监听。
func (l *ConsoleListener) Listening() bool {
	if l == nil {
		return false
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.listening
}

// Entries 返回已捕获日志副本。
func (l *ConsoleListener) Entries() []LogEntry {
	if l == nil {
		return nil
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	return cloneLogEntries(l.entries)
}

// Start 开始监听 log.entryAdded。
func (l *ConsoleListener) Start(level string) error {
	if l == nil {
		return nil
	}
	if l.owner == nil {
		return nil
	}

	l.Stop()
	l.Clear()

	result, err := bidi.Subscribe(
		l.owner.BrowserDriver(),
		[]string{"log.entryAdded"},
		[]string{l.owner.ContextID()},
		resolveUnitTimeout(l.owner),
	)
	if err != nil {
		return err
	}

	if err := l.owner.Driver().SetGlobalCallback("log.entryAdded", l.onEntry, false); err != nil {
		_ = bidi.Unsubscribe(
			l.owner.BrowserDriver(),
			nil,
			nil,
			[]string{result.Subscription},
			resolveUnitTimeout(l.owner),
		)
		return err
	}

	l.mu.Lock()
	l.listening = true
	l.subscriptionID = result.Subscription
	l.levelFilter = normalizeLogLevel(level)
	l.queue = newPacketQueue[*LogEntry](128)
	l.entries = nil
	l.mu.Unlock()
	return nil
}

// Stop 停止监听并清理订阅。
func (l *ConsoleListener) Stop() {
	if l == nil || l.owner == nil {
		return
	}

	l.mu.Lock()
	subscriptionID := l.subscriptionID
	wasListening := l.listening
	l.listening = false
	l.subscriptionID = ""
	l.levelFilter = ""
	l.mu.Unlock()

	if !wasListening {
		return
	}

	l.owner.Driver().RemoveGlobalCallback("log.entryAdded", false)
	if subscriptionID != "" {
		_ = bidi.Unsubscribe(
			l.owner.BrowserDriver(),
			nil,
			nil,
			[]string{subscriptionID},
			resolveUnitTimeout(l.owner),
		)
	}
}

// Wait 等待匹配条件的日志条目。
func (l *ConsoleListener) Wait(level string, text string, timeout time.Duration) *LogEntry {
	if l == nil {
		return nil
	}
	if timeout <= 0 {
		timeout = resolveUnitTimeout(l.owner)
	}

	deadline := time.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil
		}
		item, ok := l.queue.Pull(remaining)
		if !ok || item == nil {
			return nil
		}
		if matchLogEntry(*item, level, text) {
			entry := cloneLogEntry(*item)
			return &entry
		}
	}
}

// Get 返回按条件过滤后的日志副本。
func (l *ConsoleListener) Get(level string, text string) []LogEntry {
	if l == nil {
		return nil
	}
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]LogEntry, 0, len(l.entries))
	for _, entry := range l.entries {
		if matchLogEntry(entry, level, text) {
			result = append(result, cloneLogEntry(entry))
		}
	}
	return result
}

// Clear 清空已缓存日志与等待队列。
func (l *ConsoleListener) Clear() {
	if l == nil {
		return
	}
	l.mu.Lock()
	l.entries = nil
	queue := l.queue
	l.mu.Unlock()
	if queue != nil {
		queue.Clear()
	}
}

// OnEntry 注册实时日志回调；传 nil 表示移除。
func (l *ConsoleListener) OnEntry(callback func(LogEntry)) *ConsoleListener {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	l.realtime = callback
	l.mu.Unlock()
	return l
}

func (l *ConsoleListener) onEntry(params map[string]any) {
	if l == nil {
		return
	}

	contextID := sourceContext(params)
	if contextID != "" && contextID != l.owner.ContextID() {
		return
	}

	entry := NewLogEntry(params)

	l.mu.Lock()
	if !l.listening {
		l.mu.Unlock()
		return
	}
	if level := l.levelFilter; level != "" && normalizeLogLevel(entry.Level) != level {
		l.mu.Unlock()
		return
	}
	l.entries = append(l.entries, cloneLogEntry(entry))
	queue := l.queue
	callback := l.realtime
	l.mu.Unlock()

	copied := cloneLogEntry(entry)
	if queue != nil {
		queue.Push(&copied)
	}
	if callback != nil {
		safeRunLogCallback(callback, cloneLogEntry(entry))
	}
}

func cloneLogEntries(values []LogEntry) []LogEntry {
	if len(values) == 0 {
		return nil
	}
	result := make([]LogEntry, len(values))
	for index, value := range values {
		result[index] = cloneLogEntry(value)
	}
	return result
}

func cloneLogEntry(entry LogEntry) LogEntry {
	return LogEntry{
		Level:      entry.Level,
		Text:       entry.Text,
		Timestamp:  cloneNetworkValueDeep(entry.Timestamp),
		Source:     cloneNetworkMapDeep(entry.Source),
		LogType:    entry.LogType,
		Method:     entry.Method,
		Args:       cloneLogArgs(entry.Args),
		StackTrace: cloneNetworkValueDeep(entry.StackTrace),
	}
}

func cloneLogArgs(values []any) []any {
	if len(values) == 0 {
		return nil
	}
	cloned, _ := cloneNetworkValueDeep(values).([]any)
	return cloned
}

func matchLogEntry(entry LogEntry, level string, text string) bool {
	if normalized := normalizeLogLevel(level); normalized != "" && normalizeLogLevel(entry.Level) != normalized {
		return false
	}
	if text != "" && !strings.Contains(entry.Text, text) {
		return false
	}
	return true
}

func normalizeLogLevel(level string) string {
	return strings.ToLower(strings.TrimSpace(level))
}

func safeRunLogCallback(callback func(LogEntry), entry LogEntry) {
	defer func() {
		_ = recover()
	}()
	callback(entry)
}
