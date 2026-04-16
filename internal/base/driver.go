package base

import (
	stderrors "errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pll177/ruyipage-go/internal/support"
)

var (
	browserDriverRegistryMu sync.Mutex
	browserDriverRegistry   = make(map[string]*BrowserBiDiDriver)
)

// BrowserBiDiDriver 提供浏览器级 BiDi 连接、命令与事件入口。
type BrowserBiDiDriver struct {
	address string

	mu         sync.RWMutex
	transport  *BiDiTransport
	dispatcher *CommandDispatcher
	emitter    *EventEmitter

	callbacks         map[eventRoute]*EventSubscription
	immediateHandlers map[eventRoute]EventCallback
	sessionID         string
	running           bool
	closing           bool

	alertFlag atomic.Bool
}

// NewBrowserBiDiDriver 按 address 返回复用的浏览器级 driver。
func NewBrowserBiDiDriver(address string) *BrowserBiDiDriver {
	browserDriverRegistryMu.Lock()
	defer browserDriverRegistryMu.Unlock()

	if existing := browserDriverRegistry[address]; existing != nil {
		return existing
	}

	driver := &BrowserBiDiDriver{
		address:           address,
		emitter:           NewEventEmitter(),
		callbacks:         make(map[eventRoute]*EventSubscription),
		immediateHandlers: make(map[eventRoute]EventCallback),
	}
	browserDriverRegistry[address] = driver
	return driver
}

// Address 返回当前 driver 绑定的地址。
func (d *BrowserBiDiDriver) Address() string {
	if d == nil {
		return ""
	}
	return d.address
}

// Start 建立 WebSocket 连接；已运行时幂等返回。
func (d *BrowserBiDiDriver) Start(wsURL string, connectTimeout time.Duration) error {
	if d == nil {
		return support.NewPageDisconnectedError("BrowserBiDiDriver 未初始化", nil)
	}

	wsURL = d.resolveWSURL(wsURL)

	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return nil
	}

	d.ensureEventStateLocked()
	transport, dispatcher := d.newTransportLocked(wsURL)
	d.transport = transport
	d.dispatcher = dispatcher
	d.closing = false
	d.mu.Unlock()

	if err := transport.Connect(connectTimeout); err != nil {
		d.mu.Lock()
		if d.transport == transport {
			d.transport = nil
			d.dispatcher = nil
		}
		d.mu.Unlock()
		return err
	}

	d.mu.Lock()
	if d.transport == transport {
		d.running = true
		d.closing = false
	}
	d.mu.Unlock()
	return nil
}

// Stop 主动关闭当前连接，并从单例注册表移除。
func (d *BrowserBiDiDriver) Stop() error {
	if d == nil {
		return nil
	}

	d.mu.Lock()
	transport := d.transport
	emitter := d.emitter

	d.transport = nil
	d.dispatcher = nil
	d.emitter = nil
	d.callbacks = make(map[eventRoute]*EventSubscription)
	d.immediateHandlers = make(map[eventRoute]EventCallback)
	d.sessionID = ""
	d.running = false
	d.closing = true
	d.alertFlag.Store(false)
	d.mu.Unlock()

	if transport != nil {
		_ = transport.Close()
	}
	if emitter != nil {
		_ = emitter.Close()
	}

	browserDriverRegistryMu.Lock()
	if browserDriverRegistry[d.address] == d {
		delete(browserDriverRegistry, d.address)
	}
	browserDriverRegistryMu.Unlock()

	return nil
}

// Close 是 Stop 的别名。
func (d *BrowserBiDiDriver) Close() error {
	return d.Stop()
}

// Reconnect 关闭旧连接并重建 transport；事件订阅会被清空。
func (d *BrowserBiDiDriver) Reconnect(wsURL string, connectTimeout time.Duration) error {
	if d == nil {
		return support.NewPageDisconnectedError("BrowserBiDiDriver 未初始化", nil)
	}

	wsURL = d.resolveWSURL(wsURL)

	d.mu.Lock()
	oldTransport := d.transport
	oldEmitter := d.emitter

	d.emitter = NewEventEmitter()
	d.callbacks = make(map[eventRoute]*EventSubscription)
	d.immediateHandlers = make(map[eventRoute]EventCallback)
	d.sessionID = ""
	d.running = false
	d.closing = true
	d.alertFlag.Store(false)

	transport, dispatcher := d.newTransportLocked(wsURL)
	d.transport = transport
	d.dispatcher = dispatcher
	d.mu.Unlock()

	if oldTransport != nil {
		_ = oldTransport.Close()
	}
	if oldEmitter != nil {
		_ = oldEmitter.Close()
	}

	if err := transport.Connect(connectTimeout); err != nil {
		d.mu.Lock()
		if d.transport == transport {
			d.transport = nil
			d.dispatcher = nil
		}
		d.mu.Unlock()
		return err
	}

	d.mu.Lock()
	if d.transport == transport {
		d.running = true
		d.closing = false
	}
	d.mu.Unlock()
	return nil
}

// MarkClosing 标记当前连接即将被主动关闭。
func (d *BrowserBiDiDriver) MarkClosing() {
	if d == nil {
		return
	}

	d.mu.Lock()
	d.closing = true
	d.mu.Unlock()
}

// IsRunning 返回当前连接是否可用。
func (d *BrowserBiDiDriver) IsRunning() bool {
	if d == nil {
		return false
	}

	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.running
}

// SessionID 返回当前保存的 session id。
func (d *BrowserBiDiDriver) SessionID() string {
	if d == nil {
		return ""
	}

	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.sessionID
}

// SetSessionID 更新当前 session id。
func (d *BrowserBiDiDriver) SetSessionID(sessionID string) {
	if d == nil {
		return
	}

	d.mu.Lock()
	d.sessionID = sessionID
	d.mu.Unlock()
}

// AlertFlag 返回当前浏览器级用户提示框状态。
func (d *BrowserBiDiDriver) AlertFlag() bool {
	if d == nil {
		return false
	}
	return d.alertFlag.Load()
}

// Run 通过 dispatcher 执行同步 BiDi 命令。
func (d *BrowserBiDiDriver) Run(method string, params map[string]any, timeout time.Duration) (map[string]any, error) {
	if d == nil {
		return nil, support.NewPageDisconnectedError("BrowserBiDiDriver 未初始化", nil)
	}

	d.mu.RLock()
	dispatcher := d.dispatcher
	transport := d.transport
	running := d.running
	d.mu.RUnlock()

	if !running || dispatcher == nil {
		return nil, support.NewPageDisconnectedError("WebSocket 连接未建立", nil)
	}

	result, err := dispatcher.Dispatch(method, params, timeout)
	if err == nil {
		return result, nil
	}

	if isInvalidSessionError(err) {
		d.SetSessionID("")
	}

	var disconnectErr *support.PageDisconnectedError
	if stderrors.As(err, &disconnectErr) {
		d.handleDisconnectState(transport)
	}

	return nil, err
}

// SetCallback 注册或覆盖指定事件回调；callback 为 nil 时等价于移除。
func (d *BrowserBiDiDriver) SetCallback(event string, callback EventCallback, context string, immediate bool) error {
	if d == nil {
		return support.NewPageDisconnectedError("BrowserBiDiDriver 未初始化", nil)
	}

	route := eventRoute{event: event, context: context}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.ensureEventStateLocked()

	if immediate {
		if callback == nil {
			delete(d.immediateHandlers, route)
			return nil
		}
		d.immediateHandlers[route] = callback
		return nil
	}

	if existing := d.callbacks[route]; existing != nil {
		_ = d.emitter.Off(existing)
		delete(d.callbacks, route)
	}

	if callback == nil {
		return nil
	}

	subscription, err := d.emitter.On(event, context, callback)
	if err != nil {
		return err
	}
	d.callbacks[route] = subscription
	return nil
}

// RemoveCallback 移除指定事件回调。
func (d *BrowserBiDiDriver) RemoveCallback(event string, context string, immediate bool) {
	if d == nil {
		return
	}

	route := eventRoute{event: event, context: context}

	d.mu.Lock()
	defer d.mu.Unlock()

	if immediate {
		delete(d.immediateHandlers, route)
		return
	}

	subscription := d.callbacks[route]
	delete(d.callbacks, route)
	if d.emitter != nil && subscription != nil {
		_ = d.emitter.Off(subscription)
	}
}

func (d *BrowserBiDiDriver) resolveWSURL(wsURL string) string {
	if wsURL != "" {
		return wsURL
	}
	return fmt.Sprintf("ws://%s/session", d.address)
}

func (d *BrowserBiDiDriver) ensureEventStateLocked() {
	if d.emitter == nil {
		d.emitter = NewEventEmitter()
	}
	if d.callbacks == nil {
		d.callbacks = make(map[eventRoute]*EventSubscription)
	}
	if d.immediateHandlers == nil {
		d.immediateHandlers = make(map[eventRoute]EventCallback)
	}
}

func (d *BrowserBiDiDriver) newTransportLocked(wsURL string) (*BiDiTransport, *CommandDispatcher) {
	transport := NewBiDiTransport(wsURL)
	transport.SetEventHandler(func(event TransportEvent) {
		d.handleTransportEvent(transport, event)
	})
	transport.SetDisconnectHandler(func(err error) {
		d.handleDisconnect(transport, err)
	})
	return transport, NewCommandDispatcher(transport)
}

func (d *BrowserBiDiDriver) handleTransportEvent(source *BiDiTransport, event TransportEvent) {
	if d == nil {
		return
	}

	context := extractEventContext(event.Params)

	switch event.Method {
	case "browsingContext.userPromptOpened":
		d.alertFlag.Store(true)
	case "browsingContext.userPromptClosed":
		d.alertFlag.Store(false)
	}

	handlers, emitter := d.snapshotHandlersForEvent(source, event.Method, context)
	for _, handler := range handlers {
		go d.safeCallImmediateHandler(handler, event.Params)
	}

	if emitter != nil {
		_ = emitter.Emit(event.Method, context, event.Params)
	}
}

func (d *BrowserBiDiDriver) snapshotHandlersForEvent(source *BiDiTransport, event string, context string) ([]EventCallback, *EventEmitter) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.transport != source {
		return nil, nil
	}

	var handlers []EventCallback
	for route, handler := range d.immediateHandlers {
		if route.event != event {
			continue
		}
		if route.context != "" && route.context != context {
			continue
		}
		handlers = append(handlers, handler)
	}

	return handlers, d.emitter
}

func (d *BrowserBiDiDriver) safeCallImmediateHandler(handler EventCallback, params map[string]any) {
	defer func() {
		_ = recover()
	}()
	handler(params)
}

func (d *BrowserBiDiDriver) handleDisconnect(source *BiDiTransport, err error) {
	if d == nil {
		return
	}

	d.handleDisconnectState(source)
}

func (d *BrowserBiDiDriver) handleDisconnectState(source *BiDiTransport) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.transport != source {
		return
	}

	d.transport = nil
	d.dispatcher = nil
	d.sessionID = ""
	d.running = false
	d.alertFlag.Store(false)
}

func extractEventContext(params map[string]any) string {
	if params == nil {
		return ""
	}

	context, _ := params["context"].(string)
	return context
}

func isInvalidSessionError(err error) bool {
	var bidiErr *support.BiDiError
	if !stderrors.As(err, &bidiErr) {
		return false
	}

	code := strings.ToLower(bidiErr.Code)
	message := strings.ToLower(bidiErr.BiDiMessage)
	return strings.Contains(code, "invalid session id") || strings.Contains(message, "invalid session id")
}

// ContextDriver 为特定 browsing context 绑定上下文参数。
type ContextDriver struct {
	browserDriver *BrowserBiDiDriver
	contextID     string
}

// NewContextDriver 创建绑定 context id 的轻量 driver。
func NewContextDriver(browserDriver *BrowserBiDiDriver, contextID string) *ContextDriver {
	return &ContextDriver{
		browserDriver: browserDriver,
		contextID:     contextID,
	}
}

// ContextID 返回绑定的 context id。
func (d *ContextDriver) ContextID() string {
	if d == nil {
		return ""
	}
	return d.contextID
}

// IsRunning 返回底层浏览器级 driver 是否仍可用。
func (d *ContextDriver) IsRunning() bool {
	if d == nil || d.browserDriver == nil {
		return false
	}
	return d.browserDriver.IsRunning()
}

// AlertFlag 返回浏览器级用户提示框状态。
func (d *ContextDriver) AlertFlag() bool {
	if d == nil || d.browserDriver == nil {
		return false
	}
	return d.browserDriver.AlertFlag()
}

// Run 自动注入 context 相关参数后转发到底层浏览器级 driver。
func (d *ContextDriver) Run(method string, params map[string]any, timeout time.Duration) (map[string]any, error) {
	if d == nil || d.browserDriver == nil {
		return nil, support.NewPageDisconnectedError("ContextDriver 未初始化", nil)
	}

	resolvedParams := cloneParams(params)

	switch {
	case strings.HasPrefix(method, "browsingContext."),
		strings.HasPrefix(method, "input."),
		strings.HasPrefix(method, "emulation."):
		if _, exists := resolvedParams["context"]; !exists {
			resolvedParams["context"] = d.contextID
		}

	case method == "script.evaluate" || method == "script.callFunction":
		target, ok := cloneNestedMap(resolvedParams, "target")
		if !ok {
			target = map[string]any{}
			resolvedParams["target"] = target
		}
		if _, exists := target["context"]; !exists {
			target["context"] = d.contextID
		}

	case method == "storage.getCookies" || method == "storage.setCookie" || method == "storage.deleteCookies":
		partition, ok := cloneNestedMap(resolvedParams, "partition")
		if !ok {
			partition = map[string]any{}
			resolvedParams["partition"] = partition
		}
		if _, exists := partition["context"]; !exists {
			partition["type"] = "context"
			partition["context"] = d.contextID
		}
	}

	return d.browserDriver.Run(method, resolvedParams, timeout)
}

// SetCallback 注册限定于当前 context 的事件回调。
func (d *ContextDriver) SetCallback(event string, callback EventCallback, immediate bool) error {
	if d == nil || d.browserDriver == nil {
		return support.NewPageDisconnectedError("ContextDriver 未初始化", nil)
	}
	return d.browserDriver.SetCallback(event, callback, d.contextID, immediate)
}

// SetGlobalCallback 注册不限 context 的全局事件回调。
func (d *ContextDriver) SetGlobalCallback(event string, callback EventCallback, immediate bool) error {
	if d == nil || d.browserDriver == nil {
		return support.NewPageDisconnectedError("ContextDriver 未初始化", nil)
	}
	return d.browserDriver.SetCallback(event, callback, "", immediate)
}

// RemoveCallback 移除当前 context 的事件回调。
func (d *ContextDriver) RemoveCallback(event string, immediate bool) {
	if d == nil || d.browserDriver == nil {
		return
	}
	d.browserDriver.RemoveCallback(event, d.contextID, immediate)
}

// RemoveGlobalCallback 移除全局事件回调。
func (d *ContextDriver) RemoveGlobalCallback(event string, immediate bool) {
	if d == nil || d.browserDriver == nil {
		return
	}
	d.browserDriver.RemoveCallback(event, "", immediate)
}

func cloneParams(params map[string]any) map[string]any {
	if params == nil {
		return map[string]any{}
	}

	cloned := make(map[string]any, len(params))
	for key, value := range params {
		cloned[key] = value
	}
	return cloned
}

func cloneNestedMap(params map[string]any, key string) (map[string]any, bool) {
	if params == nil {
		return nil, false
	}

	value, exists := params[key]
	if !exists {
		return nil, false
	}

	nested, ok := value.(map[string]any)
	if !ok {
		return nil, false
	}

	cloned := make(map[string]any, len(nested))
	for nestedKey, nestedValue := range nested {
		cloned[nestedKey] = nestedValue
	}
	params[key] = cloned
	return cloned, true
}
