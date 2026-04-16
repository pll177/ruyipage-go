package base

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/pll177/ruyipage-go/internal/support"
)

const (
	defaultTransportConnectTimeout = time.Duration(support.DefaultBrowserConnectTimeoutSeconds) * time.Second
	defaultTransportRequestTimeout = 30 * time.Second
	defaultTransportEventBuffer    = 64
)

// TransportEvent 表示 Transport 收到的异步 BiDi 事件。
type TransportEvent struct {
	Type   string
	Method string
	Params map[string]any
	Raw    json.RawMessage
}

// EventHandler 是异步事件处理回调。
type EventHandler func(event TransportEvent)

// DisconnectHandler 是异常断链回调。
type DisconnectHandler func(err error)

type websocketConn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}

type websocketDialFunc func(url string, timeout time.Duration) (websocketConn, error)

type transportRequestMessage struct {
	ID     int64          `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
}

type transportEnvelope struct {
	ID         *int64         `json:"id,omitempty"`
	Type       string         `json:"type,omitempty"`
	Method     string         `json:"method,omitempty"`
	Params     map[string]any `json:"params,omitempty"`
	Result     map[string]any `json:"result,omitempty"`
	Error      string         `json:"error,omitempty"`
	Message    string         `json:"message,omitempty"`
	Stacktrace string         `json:"stacktrace,omitempty"`
}

type transportResult struct {
	response transportEnvelope
	err      error
}

// BiDiTransport 提供 BiDi WebSocket 的连接、收发、关联与关闭能力。
type BiDiTransport struct {
	wsURL string

	dial websocketDialFunc

	connMu sync.RWMutex
	conn   websocketConn

	handlerMu        sync.RWMutex
	eventHandler     EventHandler
	disconnectHandle DisconnectHandler

	sendMu    sync.Mutex
	pendingMu sync.Mutex
	pending   map[int64]chan transportResult

	nextID atomic.Int64

	closeErrMu   sync.RWMutex
	closeErr     error
	shutdownOnce sync.Once
	done         chan struct{}
	eventQueue   chan TransportEvent

	recvWG  sync.WaitGroup
	eventWG sync.WaitGroup
}

// NewBiDiTransport 创建一个新的 BiDi WebSocket 传输层实例。
func NewBiDiTransport(wsURL string) *BiDiTransport {
	return &BiDiTransport{
		wsURL:      wsURL,
		dial:       dialWebsocket,
		pending:    make(map[int64]chan transportResult),
		done:       make(chan struct{}),
		eventQueue: make(chan TransportEvent, defaultTransportEventBuffer),
	}
}

// SetEventHandler 设置异步事件回调。
func (t *BiDiTransport) SetEventHandler(handler EventHandler) {
	t.handlerMu.Lock()
	defer t.handlerMu.Unlock()
	t.eventHandler = handler
}

// SetDisconnectHandler 设置异常断链回调。
func (t *BiDiTransport) SetDisconnectHandler(handler DisconnectHandler) {
	t.handlerMu.Lock()
	defer t.handlerMu.Unlock()
	t.disconnectHandle = handler
}

// Connect 建立 WebSocket 连接并启动后台收发协程。
func (t *BiDiTransport) Connect(timeout time.Duration) error {
	if timeout <= 0 {
		timeout = defaultTransportConnectTimeout
	}
	if t.isClosed() {
		return support.NewPageDisconnectedError("BiDiTransport 已关闭，无法重新连接", t.closeError())
	}

	t.connMu.Lock()
	defer t.connMu.Unlock()

	if t.conn != nil {
		return nil
	}

	conn, err := t.dial(t.wsURL, timeout)
	if err != nil {
		return support.NewBrowserConnectError(fmt.Sprintf("BiDi WebSocket 连接失败 %s", t.wsURL), err)
	}

	t.conn = conn

	t.recvWG.Add(1)
	go t.recvLoop(conn)

	t.eventWG.Add(1)
	go t.eventLoop()

	return nil
}

// IsConnected 返回当前连接是否仍可用。
func (t *BiDiTransport) IsConnected() bool {
	if t.isClosed() {
		return false
	}

	t.connMu.RLock()
	defer t.connMu.RUnlock()

	return t.conn != nil
}

// Run 发送 BiDi 命令并阻塞等待对应响应。
func (t *BiDiTransport) Run(method string, params map[string]any, timeout time.Duration) (map[string]any, error) {
	if timeout <= 0 {
		timeout = defaultTransportRequestTimeout
	}
	if params == nil {
		params = map[string]any{}
	}

	responseCh := make(chan transportResult, 1)
	commandID := t.nextID.Add(1)
	if err := t.registerPending(commandID, responseCh); err != nil {
		return nil, err
	}
	defer t.unregisterPending(commandID)

	payload, err := json.Marshal(transportRequestMessage{
		ID:     commandID,
		Method: method,
		Params: params,
	})
	if err != nil {
		return nil, support.NewRuyiPageError("BiDi 请求序列化失败", err)
	}

	if err := t.writePayload(payload); err != nil {
		return nil, err
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case result := <-responseCh:
		if result.err != nil {
			return nil, result.err
		}
		if result.response.Type == "error" {
			return nil, support.NewBiDiError(
				result.response.Error,
				result.response.Message,
				result.response.Stacktrace,
				nil,
			)
		}
		if result.response.Result == nil {
			return map[string]any{}, nil
		}
		return result.response.Result, nil
	case <-timer.C:
		t.unregisterPending(commandID)
		return nil, support.NewBiDiError(
			"timeout",
			fmt.Sprintf("命令超时: %s (%s)", method, timeout),
			"",
			nil,
		)
	}
}

// Close 主动关闭当前 Transport，并唤醒所有等待中的请求。
func (t *BiDiTransport) Close() error {
	t.shutdown(support.NewPageDisconnectedError("BiDiTransport 已关闭", nil), false)
	t.recvWG.Wait()
	t.eventWG.Wait()
	return nil
}

func (t *BiDiTransport) recvLoop(conn websocketConn) {
	defer t.recvWG.Done()

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			t.shutdown(support.NewPageDisconnectedError("BiDi WebSocket 连接已断开", err), true)
			return
		}

		var envelope transportEnvelope
		if err := json.Unmarshal(payload, &envelope); err != nil {
			t.shutdown(support.NewPageDisconnectedError("BiDi WebSocket 收到非法 JSON 消息", err), true)
			return
		}

		if envelope.ID != nil {
			t.handleResponse(*envelope.ID, envelope)
			continue
		}

		if envelope.Method == "" {
			continue
		}

		t.enqueueEvent(TransportEvent{
			Type:   envelope.Type,
			Method: envelope.Method,
			Params: envelope.Params,
			Raw:    append(json.RawMessage(nil), payload...),
		})
	}
}

func (t *BiDiTransport) eventLoop() {
	defer t.eventWG.Done()

	for {
		select {
		case event := <-t.eventQueue:
			t.dispatchEvent(event)
		case <-t.done:
			for {
				select {
				case event := <-t.eventQueue:
					t.dispatchEvent(event)
				default:
					return
				}
			}
		}
	}
}

func (t *BiDiTransport) dispatchEvent(event TransportEvent) {
	t.handlerMu.RLock()
	handler := t.eventHandler
	t.handlerMu.RUnlock()
	if handler == nil {
		return
	}

	defer func() {
		_ = recover()
	}()

	handler(event)
}

func (t *BiDiTransport) handleResponse(commandID int64, response transportEnvelope) {
	t.pendingMu.Lock()
	responseCh, ok := t.pending[commandID]
	if ok {
		delete(t.pending, commandID)
	}
	t.pendingMu.Unlock()
	if !ok {
		return
	}

	select {
	case responseCh <- transportResult{response: response}:
	default:
	}
}

func (t *BiDiTransport) enqueueEvent(event TransportEvent) {
	select {
	case <-t.done:
		return
	case t.eventQueue <- event:
	}
}

func (t *BiDiTransport) registerPending(commandID int64, responseCh chan transportResult) error {
	if !t.IsConnected() {
		return support.NewPageDisconnectedError("WebSocket 连接未建立", t.closeError())
	}

	t.pendingMu.Lock()
	defer t.pendingMu.Unlock()

	if t.isClosed() {
		return support.NewPageDisconnectedError("WebSocket 连接已断开", t.closeError())
	}
	t.pending[commandID] = responseCh
	return nil
}

func (t *BiDiTransport) unregisterPending(commandID int64) {
	t.pendingMu.Lock()
	defer t.pendingMu.Unlock()
	delete(t.pending, commandID)
}

func (t *BiDiTransport) writePayload(payload []byte) error {
	t.sendMu.Lock()
	defer t.sendMu.Unlock()

	t.connMu.RLock()
	conn := t.conn
	t.connMu.RUnlock()
	if conn == nil || t.isClosed() {
		return support.NewPageDisconnectedError("BiDiTransport 未连接，无法发送消息", t.closeError())
	}

	if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
		wrapped := support.NewPageDisconnectedError("BiDi 消息发送失败", err)
		t.shutdown(wrapped, true)
		return wrapped
	}
	return nil
}

func (t *BiDiTransport) shutdown(reason error, notifyDisconnect bool) {
	t.shutdownOnce.Do(func() {
		t.setCloseError(reason)

		t.connMu.Lock()
		conn := t.conn
		t.conn = nil
		t.connMu.Unlock()

		close(t.done)
		t.failPending(reason)

		if conn != nil {
			_ = conn.Close()
		}

		if notifyDisconnect {
			t.handlerMu.RLock()
			handler := t.disconnectHandle
			t.handlerMu.RUnlock()
			if handler != nil {
				go t.safeCallDisconnect(handler, reason)
			}
		}
	})
}

func (t *BiDiTransport) failPending(reason error) {
	t.pendingMu.Lock()
	pending := t.pending
	t.pending = make(map[int64]chan transportResult)
	t.pendingMu.Unlock()

	for _, responseCh := range pending {
		select {
		case responseCh <- transportResult{err: reason}:
		default:
		}
	}
}

func (t *BiDiTransport) safeCallDisconnect(handler DisconnectHandler, reason error) {
	defer func() {
		_ = recover()
	}()
	handler(reason)
}

func (t *BiDiTransport) isClosed() bool {
	select {
	case <-t.done:
		return true
	default:
		return false
	}
}

func (t *BiDiTransport) closeError() error {
	t.closeErrMu.RLock()
	defer t.closeErrMu.RUnlock()
	return t.closeErr
}

func (t *BiDiTransport) setCloseError(err error) {
	t.closeErrMu.Lock()
	defer t.closeErrMu.Unlock()
	t.closeErr = err
}

func dialWebsocket(url string, timeout time.Duration) (websocketConn, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout: timeout,
	}
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
