package units

import (
	"sync"
	"time"

	"ruyipage-go/internal/base"
	"ruyipage-go/internal/bidi"
)

// BidiEvent 表示单个 BiDi 事件的高层快照。
type BidiEvent struct {
	Method         string
	Params         map[string]any
	Context        string
	Navigation     string
	Timestamp      any
	URL            string
	Request        map[string]any
	Response       map[string]any
	IsBlocked      bool
	ErrorText      string
	AuthChallenge  map[string]any
	Realm          string
	Source         map[string]any
	Channel        string
	Data           any
	Multiple       bool
	UserPromptType string
	Accepted       bool
	Message        string
}

// NewBidiEvent 从原始事件参数构建高层事件对象。
func NewBidiEvent(method string, params map[string]any) BidiEvent {
	raw := cloneNetworkMapDeep(params)
	return BidiEvent{
		Method:         method,
		Params:         raw,
		Context:        stringifyNetworkValue(raw["context"]),
		Navigation:     stringifyNetworkValue(raw["navigation"]),
		Timestamp:      cloneNetworkValueDeep(raw["timestamp"]),
		URL:            stringifyNetworkValue(raw["url"]),
		Request:        cloneMapFromAny(raw["request"]),
		Response:       cloneMapFromAny(raw["response"]),
		IsBlocked:      readUnitBool(raw["isBlocked"]),
		ErrorText:      stringifyNetworkValue(raw["errorText"]),
		AuthChallenge:  cloneMapFromAny(raw["authChallenge"]),
		Realm:          stringifyNetworkValue(raw["realm"]),
		Source:         cloneMapFromAny(raw["source"]),
		Channel:        stringifyNetworkValue(raw["channel"]),
		Data:           cloneNetworkValueDeep(raw["data"]),
		Multiple:       readUnitBool(raw["multiple"]),
		UserPromptType: stringifyNetworkValue(raw["type"]),
		Accepted:       readUnitBool(raw["accepted"]),
		Message:        stringifyNetworkValue(raw["message"]),
	}
}

type eventCallbackDriver interface {
	SetGlobalCallback(event string, callback base.EventCallback, immediate bool) error
	RemoveGlobalCallback(event string, immediate bool)
}

// EventTracker 提供通用 BiDi 事件监听与等待能力。
type EventTracker struct {
	owner networkOwner

	mu             sync.RWMutex
	listening      bool
	events         []string
	entries        []BidiEvent
	queue          *packetQueue[*BidiEvent]
	subscriptionID string
}

// NewEventTracker 创建通用事件跟踪器。
func NewEventTracker(owner networkOwner) *EventTracker {
	return &EventTracker{
		owner: owner,
		queue: newPacketQueue[*BidiEvent](128),
	}
}

// Listening 返回当前是否正在监听。
func (t *EventTracker) Listening() bool {
	if t == nil {
		return false
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.listening
}

// Entries 返回已捕获事件副本。
func (t *EventTracker) Entries() []BidiEvent {
	if t == nil {
		return nil
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return cloneBidiEvents(t.entries)
}

// Start 开始监听指定事件列表。
func (t *EventTracker) Start(events any, contexts any) error {
	if t == nil || t.owner == nil {
		return nil
	}

	normalizedEvents, err := normalizeUnitStringList(events, "events", true)
	if err != nil {
		return err
	}
	normalizedContexts, err := normalizeUnitStringList(contexts, "contexts", false)
	if err != nil {
		return err
	}
	if len(normalizedContexts) == 0 {
		normalizedContexts = []string{t.owner.ContextID()}
	}

	t.Stop()
	t.Clear()

	result, err := bidi.Subscribe(t.owner.BrowserDriver(), normalizedEvents, normalizedContexts, resolveUnitTimeout(t.owner))
	if err != nil {
		return err
	}

	callbackDriver := t.callbackDriver()
	registered := make([]string, 0, len(normalizedEvents))
	for _, event := range normalizedEvents {
		if err := callbackDriver.SetGlobalCallback(event, t.makeHandler(event), false); err != nil {
			for _, registeredEvent := range registered {
				callbackDriver.RemoveGlobalCallback(registeredEvent, false)
			}
			_ = bidi.Unsubscribe(t.owner.BrowserDriver(), nil, nil, []string{result.Subscription}, resolveUnitTimeout(t.owner))
			return err
		}
		registered = append(registered, event)
	}

	t.mu.Lock()
	t.listening = true
	t.events = append([]string(nil), normalizedEvents...)
	t.subscriptionID = result.Subscription
	t.queue = newPacketQueue[*BidiEvent](128)
	t.entries = nil
	t.mu.Unlock()
	return nil
}

// Stop 停止当前事件监听。
func (t *EventTracker) Stop() {
	if t == nil || t.owner == nil {
		return
	}

	t.mu.Lock()
	events := append([]string(nil), t.events...)
	subscriptionID := t.subscriptionID
	wasListening := t.listening
	t.listening = false
	t.events = nil
	t.subscriptionID = ""
	t.mu.Unlock()

	if !wasListening {
		return
	}

	callbackDriver := t.callbackDriver()
	for _, event := range events {
		callbackDriver.RemoveGlobalCallback(event, false)
	}
	if subscriptionID != "" {
		_ = bidi.Unsubscribe(
			t.owner.BrowserDriver(),
			nil,
			nil,
			[]string{subscriptionID},
			resolveUnitTimeout(t.owner),
		)
	}
}

// Clear 清空事件缓存与等待队列。
func (t *EventTracker) Clear() {
	if t == nil {
		return
	}
	t.mu.Lock()
	t.entries = nil
	queue := t.queue
	t.mu.Unlock()
	if queue != nil {
		queue.Clear()
	}
}

// Wait 等待匹配事件名的事件；event 为空时接受任意事件。
func (t *EventTracker) Wait(event string, timeout time.Duration) *BidiEvent {
	if t == nil {
		return nil
	}
	if timeout <= 0 {
		timeout = resolveUnitTimeout(t.owner)
	}

	deadline := time.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil
		}
		item, ok := t.queue.Pull(remaining)
		if !ok || item == nil {
			return nil
		}
		if event == "" || item.Method == event {
			value := cloneBidiEvent(*item)
			return &value
		}
	}
}

func (t *EventTracker) callbackDriver() eventCallbackDriver {
	return t.owner.Driver()
}

func (t *EventTracker) makeHandler(event string) base.EventCallback {
	return func(params map[string]any) {
		item := NewBidiEvent(event, params)

		t.mu.Lock()
		if !t.listening {
			t.mu.Unlock()
			return
		}
		t.entries = append(t.entries, cloneBidiEvent(item))
		queue := t.queue
		t.mu.Unlock()

		copied := cloneBidiEvent(item)
		if queue != nil {
			queue.Push(&copied)
		}
	}
}

func cloneBidiEvents(values []BidiEvent) []BidiEvent {
	if len(values) == 0 {
		return nil
	}
	result := make([]BidiEvent, len(values))
	for index, value := range values {
		result[index] = cloneBidiEvent(value)
	}
	return result
}

func cloneBidiEvent(event BidiEvent) BidiEvent {
	return BidiEvent{
		Method:         event.Method,
		Params:         cloneNetworkMapDeep(event.Params),
		Context:        event.Context,
		Navigation:     event.Navigation,
		Timestamp:      cloneNetworkValueDeep(event.Timestamp),
		URL:            event.URL,
		Request:        cloneNetworkMapDeep(event.Request),
		Response:       cloneNetworkMapDeep(event.Response),
		IsBlocked:      event.IsBlocked,
		ErrorText:      event.ErrorText,
		AuthChallenge:  cloneNetworkMapDeep(event.AuthChallenge),
		Realm:          event.Realm,
		Source:         cloneNetworkMapDeep(event.Source),
		Channel:        event.Channel,
		Data:           cloneNetworkValueDeep(event.Data),
		Multiple:       event.Multiple,
		UserPromptType: event.UserPromptType,
		Accepted:       event.Accepted,
		Message:        event.Message,
	}
}

func readUnitBool(value any) bool {
	typed, _ := value.(bool)
	return typed
}
