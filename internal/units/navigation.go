package units

import (
	"strings"
	"sync"
	"time"

	"github.com/pll177/ruyipage-go/internal/base"
	"github.com/pll177/ruyipage-go/internal/bidi"
)

type NavigationCommandDriver interface {
	Run(method string, params map[string]any, timeout time.Duration) (map[string]any, error)
	SessionID() string
	SetSessionID(sessionID string)
}

type NavigationCallbackDriver interface {
	SetCallback(event string, callback base.EventCallback, immediate bool) error
	RemoveCallback(event string, immediate bool)
}

// NavigationEvent 表示一次导航相关事件快照。
type NavigationEvent struct {
	Method     string
	Params     map[string]any
	Context    string
	Navigation string
	Timestamp  any
	URL        string
}

// NavigationTracker 表示页面级导航事件跟踪器。
type NavigationTracker struct {
	browserDriverFn func() NavigationCommandDriver
	contextDriverFn func() NavigationCallbackDriver
	contextIDFn     func() string
	timeoutFn       func() time.Duration

	mu             sync.RWMutex
	listening      bool
	entries        []NavigationEvent
	queue          chan NavigationEvent
	subscriptionID string
	events         []string
}

// DefaultNavigationEvents 是默认订阅的导航事件列表。
var DefaultNavigationEvents = []string{
	"browsingContext.navigationStarted",
	"browsingContext.fragmentNavigated",
	"browsingContext.historyUpdated",
	"browsingContext.domContentLoaded",
	"browsingContext.load",
	"browsingContext.navigationCommitted",
	"browsingContext.navigationFailed",
}

// NewNavigationTracker 创建导航事件跟踪器。
func NewNavigationTracker(
	browserDriverFn func() NavigationCommandDriver,
	contextDriverFn func() NavigationCallbackDriver,
	contextIDFn func() string,
	timeoutFn func() time.Duration,
) *NavigationTracker {
	return &NavigationTracker{
		browserDriverFn: browserDriverFn,
		contextDriverFn: contextDriverFn,
		contextIDFn:     contextIDFn,
		timeoutFn:       timeoutFn,
		queue:           make(chan NavigationEvent, 128),
	}
}

// Listening 返回当前是否处于监听状态。
func (n *NavigationTracker) Listening() bool {
	if n == nil {
		return false
	}
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.listening
}

// Entries 返回已捕获事件副本。
func (n *NavigationTracker) Entries() []NavigationEvent {
	if n == nil {
		return nil
	}
	n.mu.RLock()
	defer n.mu.RUnlock()
	entries := make([]NavigationEvent, len(n.entries))
	copy(entries, n.entries)
	return entries
}

// Start 开始跟踪导航事件。
func (n *NavigationTracker) Start(events []string) error {
	if n == nil {
		return nil
	}
	driver := n.browserDriver()
	callbackDriver := n.contextDriver()
	contextID := n.contextID()
	if driver == nil || callbackDriver == nil || contextID == "" {
		return nil
	}
	n.Stop()
	n.Clear()
	if len(events) == 0 {
		events = append([]string{}, DefaultNavigationEvents...)
	} else {
		events = append([]string{}, events...)
	}

	result, err := bidi.Subscribe(driver, events, []string{contextID}, n.resolveTimeout())
	if err != nil {
		return err
	}

	registered := make([]string, 0, len(events))
	for _, event := range events {
		if err := callbackDriver.SetCallback(event, n.makeHandler(event), false); err != nil {
			for _, registeredEvent := range registered {
				callbackDriver.RemoveCallback(registeredEvent, false)
			}
			_ = bidi.Unsubscribe(driver, nil, nil, []string{result.Subscription}, n.resolveTimeout())
			return err
		}
		registered = append(registered, event)
	}

	n.mu.Lock()
	n.events = events
	n.subscriptionID = result.Subscription
	n.listening = true
	n.mu.Unlock()
	return nil
}

// Stop 停止导航事件跟踪。
func (n *NavigationTracker) Stop() {
	if n == nil {
		return
	}
	n.mu.Lock()
	events := append([]string{}, n.events...)
	subscriptionID := n.subscriptionID
	n.events = nil
	n.subscriptionID = ""
	n.listening = false
	n.mu.Unlock()

	callbackDriver := n.contextDriver()
	if callbackDriver == nil {
		return
	}
	for _, event := range events {
		callbackDriver.RemoveCallback(event, false)
	}
	if subscriptionID != "" {
		if driver := n.browserDriver(); driver != nil {
			_ = bidi.Unsubscribe(driver, nil, nil, []string{subscriptionID}, n.resolveTimeout())
		}
	}
}

// Clear 清空历史记录与等待队列。
func (n *NavigationTracker) Clear() {
	if n == nil {
		return
	}
	n.mu.Lock()
	n.entries = nil
	queue := n.queue
	n.mu.Unlock()

	for {
		select {
		case <-queue:
		default:
			return
		}
	}
}

// Wait 等待匹配条件的导航事件，超时返回 nil。
func (n *NavigationTracker) Wait(event string, timeout time.Duration, urlContains string) *NavigationEvent {
	if n == nil {
		return nil
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	for {
		select {
		case item := <-n.queue:
			if navigationEventMatch(item, event, urlContains) {
				matched := item
				return &matched
			}
		case <-deadline.C:
			return nil
		}
	}
}

// WaitForFragment 等待 fragmentNavigated 命中指定片段。
func (n *NavigationTracker) WaitForFragment(fragment string, timeout time.Duration) *NavigationEvent {
	if fragment != "" && fragment[0] != '#' {
		fragment = "#" + fragment
	}
	return n.Wait("browsingContext.fragmentNavigated", timeout, fragment)
}

// WaitForLoad 等待 load 事件。
func (n *NavigationTracker) WaitForLoad(timeout time.Duration) *NavigationEvent {
	return n.Wait("browsingContext.load", timeout, "")
}

func (n *NavigationTracker) resolveTimeout() time.Duration {
	if n == nil || n.timeoutFn == nil {
		return 5 * time.Second
	}
	timeout := n.timeoutFn()
	if timeout <= 0 {
		return 5 * time.Second
	}
	return timeout
}

func (n *NavigationTracker) makeHandler(event string) base.EventCallback {
	return func(params map[string]any) {
		item := NavigationEvent{
			Method:     event,
			Params:     cloneActionRow(params),
			Context:    stringifyActionValue(params["context"]),
			Navigation: stringifyActionValue(params["navigation"]),
			Timestamp:  params["timestamp"],
			URL:        stringifyActionValue(params["url"]),
		}

		n.mu.Lock()
		n.entries = append(n.entries, item)
		queue := n.queue
		n.mu.Unlock()

		select {
		case queue <- item:
		default:
			<-queue
			queue <- item
		}
	}
}

func (n *NavigationTracker) browserDriver() NavigationCommandDriver {
	if n == nil || n.browserDriverFn == nil {
		return nil
	}
	return n.browserDriverFn()
}

func (n *NavigationTracker) contextDriver() NavigationCallbackDriver {
	if n == nil || n.contextDriverFn == nil {
		return nil
	}
	return n.contextDriverFn()
}

func (n *NavigationTracker) contextID() string {
	if n == nil || n.contextIDFn == nil {
		return ""
	}
	return n.contextIDFn()
}

func navigationEventMatch(item NavigationEvent, event string, urlContains string) bool {
	if event != "" && item.Method != event {
		return false
	}
	if urlContains != "" && !strings.Contains(item.URL, urlContains) {
		return false
	}
	return true
}
