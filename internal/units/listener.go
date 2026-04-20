package units

import (
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/pll177/ruyipage-go/internal/bidi"
)

const (
	listenerResponseCompleted = "network.responseCompleted"
	listenerFetchError        = "network.fetchError"
	listenerBeforeRequestSent = "network.beforeRequestSent"
)

// DataPacket 表示一次网络监听结果。
type DataPacket struct {
	Request   map[string]any
	Response  map[string]any
	EventType string
	URL       string
	Method    string
	Status    int
	Headers   map[string]string
	Body      any
	Timestamp any
}

// IsFailed 返回当前数据包是否来自 fetchError。
func (p DataPacket) IsFailed() bool {
	return p.EventType == "fetchError"
}

// Listener 提供高层网络抓包能力。
type Listener struct {
	owner networkOwner

	mu             sync.RWMutex
	listening      bool
	allTargets     bool
	targets        []string
	regexps        []*regexp.Regexp
	methodFilter   string
	subscriptionID string
	queue          *packetQueue[*DataPacket]
	packets        []DataPacket
}

// NewListener 创建网络监听器。
func NewListener(owner networkOwner) *Listener {
	return &Listener{
		owner: owner,
		queue: newPacketQueue[*DataPacket](256),
	}
}

// Listening 返回当前是否处于监听状态。
func (l *Listener) Listening() bool {
	if l == nil {
		return false
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.listening
}

// Steps 返回已捕获的数据包快照。
func (l *Listener) Steps() []DataPacket {
	if l == nil {
		return []DataPacket{}
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make([]DataPacket, len(l.packets))
	for index, packet := range l.packets {
		result[index] = cloneDataPacket(packet)
	}
	return result
}

// Start 开始监听网络响应与失败事件。
func (l *Listener) Start(targets any, isRegex bool, method string) error {
	if l == nil {
		return nil
	}
	l.Stop()

	compiled, patterns, allTargets := normalizeListenerTargets(targets, isRegex)
	driver := l.owner.BrowserDriver()
	result, err := bidi.Subscribe(driver, []string{
		listenerBeforeRequestSent,
		listenerResponseCompleted,
		listenerFetchError,
	}, []string{l.owner.ContextID()}, l.resolveTimeout())
	if err != nil {
		return err
	}

	callbackDriver := l.owner.Driver()
	if err := callbackDriver.SetCallback(listenerResponseCompleted, l.onResponseCompleted, false); err != nil {
		_ = bidi.Unsubscribe(driver, nil, nil, []string{result.Subscription}, l.resolveTimeout())
		return err
	}
	if err := callbackDriver.SetCallback(listenerFetchError, l.onFetchError, false); err != nil {
		callbackDriver.RemoveCallback(listenerResponseCompleted, false)
		_ = bidi.Unsubscribe(driver, nil, nil, []string{result.Subscription}, l.resolveTimeout())
		return err
	}

	l.mu.Lock()
	l.listening = true
	l.allTargets = allTargets
	l.targets = patterns
	l.regexps = compiled
	l.methodFilter = strings.ToUpper(strings.TrimSpace(method))
	l.subscriptionID = result.Subscription
	l.queue = newPacketQueue[*DataPacket](256)
	l.packets = nil
	l.mu.Unlock()
	return nil
}

// Stop 停止监听并清理订阅与回调。
func (l *Listener) Stop() {
	if l == nil || l.owner == nil {
		return
	}

	l.mu.Lock()
	subscriptionID := l.subscriptionID
	wasListening := l.listening
	l.listening = false
	l.subscriptionID = ""
	l.mu.Unlock()

	if !wasListening {
		return
	}

	l.owner.Driver().RemoveCallback(listenerResponseCompleted, false)
	l.owner.Driver().RemoveCallback(listenerFetchError, false)
	if subscriptionID != "" {
		_ = bidi.Unsubscribe(l.owner.BrowserDriver(), nil, nil, []string{subscriptionID}, l.resolveTimeout())
	}
}

// Wait 等待一个数据包；超时返回 nil。
func (l *Listener) Wait(timeout time.Duration) *DataPacket {
	if l == nil {
		return nil
	}
	value, ok := l.queue.Pull(timeout)
	if !ok || value == nil {
		return nil
	}
	packet := cloneDataPacket(*value)
	return &packet
}

// WaitCount 等待多个数据包；超时返回已捕获结果。
func (l *Listener) WaitCount(timeout time.Duration, count int) []*DataPacket {
	if l == nil {
		return nil
	}
	if count <= 0 {
		count = 1
	}
	if timeout <= 0 {
		timeout = l.resolveTimeout()
	}
	deadline := time.Now().Add(timeout)
	result := make([]*DataPacket, 0, count)
	for len(result) < count {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		packet := l.Wait(remaining)
		if packet == nil {
			break
		}
		result = append(result, packet)
	}
	return result
}

// Clear 清空当前缓存的抓包结果与等待队列。
func (l *Listener) Clear() {
	if l == nil {
		return
	}
	l.mu.Lock()
	l.packets = nil
	queue := l.queue
	l.mu.Unlock()
	queue.Clear()
}

func (l *Listener) resolveTimeout() time.Duration {
	if l == nil || l.owner == nil {
		return networkDefaultTimeout()
	}
	timeout := l.owner.BaseTimeout()
	if timeout <= 0 {
		return networkDefaultTimeout()
	}
	return timeout
}

func (l *Listener) onResponseCompleted(params map[string]any) {
	if !l.shouldCapture(params) {
		return
	}

	request, _ := params["request"].(map[string]any)
	response, _ := params["response"].(map[string]any)
	packet := DataPacket{
		Request:   cloneNetworkMapDeep(request),
		Response:  cloneNetworkMapDeep(response),
		EventType: "responseCompleted",
		URL:       stringifyNetworkValue(request["url"]),
		Method:    stringifyNetworkValue(request["method"]),
		Status:    intNetworkValue(response["status"]),
		Headers:   bidiHeadersToMap(response["headers"], true),
		Body:      cloneNetworkValueDeep(params["body"]),
		Timestamp: cloneNetworkValueDeep(params["timestamp"]),
	}
	l.push(packet)
}

func (l *Listener) onFetchError(params map[string]any) {
	if !l.shouldCapture(params) {
		return
	}

	request, _ := params["request"].(map[string]any)
	packet := DataPacket{
		Request:   cloneNetworkMapDeep(request),
		EventType: "fetchError",
		URL:       stringifyNetworkValue(request["url"]),
		Method:    stringifyNetworkValue(request["method"]),
		Body:      cloneNetworkValueDeep(params["body"]),
		Timestamp: cloneNetworkValueDeep(params["timestamp"]),
		Headers:   map[string]string{},
	}
	l.push(packet)
}

func (l *Listener) shouldCapture(params map[string]any) bool {
	if l == nil {
		return false
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	if !l.listening {
		return false
	}

	request, _ := params["request"].(map[string]any)
	url := stringifyNetworkValue(request["url"])
	method := strings.ToUpper(stringifyNetworkValue(request["method"]))
	if l.methodFilter != "" && method != l.methodFilter {
		return false
	}
	if l.allTargets {
		return true
	}
	for _, pattern := range l.targets {
		if strings.Contains(url, pattern) {
			return true
		}
	}
	for _, pattern := range l.regexps {
		if pattern.MatchString(url) {
			return true
		}
	}
	return false
}

func (l *Listener) push(packet DataPacket) {
	if l == nil {
		return
	}
	l.mu.Lock()
	l.packets = append(l.packets, cloneDataPacket(packet))
	queue := l.queue
	l.mu.Unlock()

	copied := cloneDataPacket(packet)
	queue.Push(&copied)
}

func normalizeListenerTargets(targets any, isRegex bool) ([]*regexp.Regexp, []string, bool) {
	switch typed := targets.(type) {
	case nil:
		return nil, nil, true
	case bool:
		if typed {
			return nil, nil, true
		}
		return nil, nil, true
	case string:
		if typed == "" {
			return nil, nil, true
		}
		if isRegex {
			return compilePatterns([]string{typed}), nil, false
		}
		return nil, []string{typed}, false
	case []string:
		if len(typed) == 0 {
			return nil, nil, true
		}
		if isRegex {
			return compilePatterns(typed), nil, false
		}
		values := make([]string, len(typed))
		copy(values, typed)
		return nil, values, false
	case []any:
		if len(typed) == 0 {
			return nil, nil, true
		}
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			text := stringifyNetworkValue(item)
			if text != "" {
				values = append(values, text)
			}
		}
		if len(values) == 0 {
			return nil, nil, true
		}
		if isRegex {
			return compilePatterns(values), nil, false
		}
		return nil, values, false
	default:
		return nil, nil, true
	}
}

func compilePatterns(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err == nil {
			compiled = append(compiled, re)
		}
	}
	return compiled
}

func cloneDataPacket(packet DataPacket) DataPacket {
	return DataPacket{
		Request:   cloneNetworkMapDeep(packet.Request),
		Response:  cloneNetworkMapDeep(packet.Response),
		EventType: packet.EventType,
		URL:       packet.URL,
		Method:    packet.Method,
		Status:    packet.Status,
		Headers:   cloneNetworkStringMap(packet.Headers),
		Body:      cloneNetworkValueDeep(packet.Body),
		Timestamp: cloneNetworkValueDeep(packet.Timestamp),
	}
}
