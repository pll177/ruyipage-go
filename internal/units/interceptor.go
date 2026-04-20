package units

import (
	"sync"
	"time"

	"github.com/pll177/ruyipage-go/internal/base"
	"github.com/pll177/ruyipage-go/internal/bidi"
	"github.com/pll177/ruyipage-go/internal/support"
)

// InterceptedRequest 表示一次被拦截的请求/响应。
type InterceptedRequest struct {
	mu sync.Mutex

	Raw             map[string]any
	Request         map[string]any
	Response        map[string]any
	RequestID       string
	URL             string
	Method          string
	Headers         map[string]string
	ResponseHeaders map[string]string
	Phase           string
	Status          int

	driver            *base.BrowserBiDiDriver
	collector         *DataCollector
	responseCollector *DataCollector
	timeout           time.Duration

	handled            bool
	bodyLoaded         bool
	body               string
	responseBodyLoaded bool
	responseBody       string
}

// NewInterceptedRequest 基于 BiDi 事件参数构建高层请求对象。
func NewInterceptedRequest(
	params map[string]any,
	driver *base.BrowserBiDiDriver,
	collector *DataCollector,
	responseCollector *DataCollector,
	timeout time.Duration,
) *InterceptedRequest {
	request, _ := params["request"].(map[string]any)
	response, _ := params["response"].(map[string]any)

	phase := ""
	if intercepts, ok := params["intercepts"].([]any); ok && len(intercepts) > 0 {
		phase = stringifyNetworkValue(intercepts[0])
	}

	return &InterceptedRequest{
		Raw:               cloneNetworkMapDeep(params),
		Request:           cloneNetworkMapDeep(request),
		Response:          cloneNetworkMapDeep(response),
		RequestID:         stringifyNetworkValue(request["request"]),
		URL:               stringifyNetworkValue(request["url"]),
		Method:            stringifyNetworkValue(request["method"]),
		Headers:           bidiHeadersToMap(request["headers"], false),
		ResponseHeaders:   bidiHeadersToMap(response["headers"], false),
		Phase:             phase,
		Status:            intNetworkValue(response["status"]),
		driver:            driver,
		collector:         collector,
		responseCollector: responseCollector,
		timeout:           timeout,
	}
}

// Body 返回请求体文本；没有可读 body 时返回空字符串。
func (r *InterceptedRequest) Body() string {
	if r == nil {
		return ""
	}
	r.mu.Lock()
	if r.bodyLoaded {
		defer r.mu.Unlock()
		return r.body
	}
	r.mu.Unlock()

	body := r.loadBody()

	r.mu.Lock()
	r.body = body
	r.bodyLoaded = true
	r.mu.Unlock()
	return body
}

// ResponseBody 返回响应体文本；未启用 collectResponse 或暂无可读 body 时返回空字符串。
func (r *InterceptedRequest) ResponseBody() string {
	if r == nil {
		return ""
	}
	r.mu.Lock()
	if r.responseBodyLoaded {
		defer r.mu.Unlock()
		return r.responseBody
	}
	r.mu.Unlock()

	body := r.loadResponseBody()

	r.mu.Lock()
	r.responseBody = body
	r.responseBodyLoaded = true
	r.mu.Unlock()
	return body
}

// IsResponsePhase 返回当前拦截是否处于 responseStarted 阶段。
func (r *InterceptedRequest) IsResponsePhase() bool {
	if r == nil {
		return false
	}
	return len(r.Response) > 0
}

// Handled 返回当前请求是否已经被处理。
func (r *InterceptedRequest) Handled() bool {
	if r == nil {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.handled
}

// ContinueRequest 放行请求并可选改写 URL、Method、Headers、Body。
func (r *InterceptedRequest) ContinueRequest(url string, method string, headers any, body any) error {
	if !r.markHandled() {
		return nil
	}
	_, err := bidi.ContinueRequest(
		r.driver,
		r.RequestID,
		normalizeBytesValue(body, false),
		nil,
		normalizeBiDiHeaders(headers),
		method,
		url,
		r.resolveTimeout(),
	)
	if err != nil {
		return support.NewNetworkInterceptError("ContinueRequest 失败", err)
	}
	return nil
}

// Fail 直接中止当前请求。
func (r *InterceptedRequest) Fail() error {
	if !r.markHandled() {
		return nil
	}
	_, err := bidi.FailRequest(r.driver, r.RequestID, r.resolveTimeout())
	if err != nil {
		return support.NewNetworkInterceptError("FailRequest 失败", err)
	}
	return nil
}

// ContinueWithAuth 处理认证挑战。
func (r *InterceptedRequest) ContinueWithAuth(action string, username string, password string) error {
	if !r.markHandled() {
		return nil
	}
	var credentials any
	if action == "provideCredentials" {
		credentials = map[string]any{
			"type":     "password",
			"username": username,
			"password": password,
		}
	}
	_, err := bidi.ContinueWithAuth(r.driver, r.RequestID, action, credentials, r.resolveTimeout())
	if err != nil {
		return support.NewNetworkInterceptError("ContinueWithAuth 失败", err)
	}
	return nil
}

// ContinueResponse 放行被拦截的响应并可选改写响应头与状态。
func (r *InterceptedRequest) ContinueResponse(headers any, reasonPhrase string, statusCode *int) error {
	if !r.markHandled() {
		return nil
	}
	_, err := bidi.ContinueResponse(
		r.driver,
		r.RequestID,
		nil,
		nil,
		normalizeBiDiHeaders(headers),
		reasonPhrase,
		statusCode,
		r.resolveTimeout(),
	)
	if err != nil {
		return support.NewNetworkInterceptError("ContinueResponse 失败", err)
	}
	return nil
}

// Mock 直接为当前请求提供模拟响应。
func (r *InterceptedRequest) Mock(body any, statusCode int, headers any, reasonPhrase string) error {
	if !r.markHandled() {
		return nil
	}
	if statusCode == 0 {
		statusCode = 200
	}
	if reasonPhrase == "" {
		reasonPhrase = "OK"
	}
	if headers == nil {
		headers = []map[string]any{
			{
				"name":  "content-type",
				"value": map[string]any{"type": "string", "value": "text/plain"},
			},
		}
	}
	_, err := bidi.ProvideResponse(
		r.driver,
		r.RequestID,
		normalizeBytesValue(body, true),
		nil,
		normalizeBiDiHeaders(headers),
		reasonPhrase,
		&statusCode,
		r.resolveTimeout(),
	)
	if err != nil {
		return support.NewNetworkInterceptError("ProvideResponse 失败", err)
	}
	return nil
}

func (r *InterceptedRequest) markHandled() bool {
	if r == nil {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handled {
		return false
	}
	r.handled = true
	return true
}

func (r *InterceptedRequest) resolveTimeout() time.Duration {
	if r == nil || r.timeout <= 0 {
		return networkDefaultTimeout()
	}
	return r.timeout
}

func (r *InterceptedRequest) loadBody() string {
	if r == nil {
		return ""
	}
	if body, ok := decodeNetworkBodyValue(r.Request["body"]); ok {
		return body
	}
	if body, ok := decodeNetworkBodyValue(r.Raw["body"]); ok {
		return body
	}
	return loadCollectedBody(r.collector, r.RequestID, "request", 1, 0)
}

func (r *InterceptedRequest) loadResponseBody() string {
	if r == nil {
		return ""
	}
	if body, ok := decodeNetworkBodyValue(r.Response["body"]); ok {
		return body
	}
	return loadCollectedBody(r.responseCollector, r.RequestID, "response", 10, 300*time.Millisecond)
}

func loadCollectedBody(collector *DataCollector, requestID string, dataType string, attempts int, delay time.Duration) string {
	if collector == nil || requestID == "" {
		return ""
	}
	if attempts <= 0 {
		attempts = 1
	}
	for attempt := 0; attempt < attempts; attempt++ {
		data, err := collector.Get(requestID, dataType)
		if err == nil && data != nil {
			if body, ok := decodeNetworkBodyValue(data.Bytes); ok {
				return body
			}
			if body, ok := decodeNetworkBodyValue(data.Base64); ok {
				return body
			}
			if body, ok := decodeNetworkBodyValue(data.Raw["body"]); ok {
				return body
			}
			if body, ok := decodeNetworkBodyValue(data.Raw["data"]); ok {
				return body
			}
			if body, ok := decodeNetworkBodyValue(data.Raw["value"]); ok {
				return body
			}
			if body, ok := decodeNetworkBodyValue(data.Raw); ok {
				return body
			}
		}
		if attempt+1 < attempts && delay > 0 {
			time.Sleep(delay)
		}
	}
	return ""
}

// Interceptor 提供高层请求拦截与改写能力。
type Interceptor struct {
	owner networkOwner

	mu                sync.RWMutex
	active            bool
	interceptID       string
	subscriptionID    string
	requestCollector  *DataCollector
	responseCollector *DataCollector
	handler           func(*InterceptedRequest)
	queue             *packetQueue[*InterceptedRequest]
	phases            []string
}

// NewInterceptor 创建网络拦截器。
func NewInterceptor(owner networkOwner) *Interceptor {
	return &Interceptor{
		owner: owner,
		queue: newPacketQueue[*InterceptedRequest](128),
	}
}

// Active 返回当前是否处于拦截状态。
func (i *Interceptor) Active() bool {
	if i == nil {
		return false
	}
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.active
}

// Start 启动网络拦截。
func (i *Interceptor) Start(
	handler func(*InterceptedRequest),
	urlPatterns []map[string]any,
	phases []string,
	collectResponse ...bool,
) (*Interceptor, error) {
	if i == nil {
		return nil, nil
	}
	i.Stop()

	if len(phases) == 0 {
		phases = []string{"beforeRequestSent"}
	}
	timeout := i.resolveTimeout()
	driver := i.owner.BrowserDriver()
	callbackDriver := i.owner.Driver()
	contexts := []string{i.owner.ContextID()}
	collectResponseEnabled := len(collectResponse) > 0 && collectResponse[0]

	manager := NewNetworkManager(i.owner)
	var requestCollector *DataCollector
	if hasString(phases, "beforeRequestSent") {
		if collector, err := manager.AddDataCollector([]string{"beforeRequestSent"}, []string{"request"}, 0); err == nil {
			requestCollector = collector
		}
	}

	var responseCollector *DataCollector
	if collectResponseEnabled {
		if collector, err := manager.AddDataCollector([]string{listenerResponseCompleted}, []string{"response"}, 0); err == nil {
			responseCollector = collector
		}
	}

	result, err := bidi.AddIntercept(driver, phases, urlPatterns, contexts, timeout)
	if err != nil {
		if requestCollector != nil {
			_ = requestCollector.Remove()
		}
		if responseCollector != nil {
			_ = responseCollector.Remove()
		}
		return nil, err
	}

	events := make([]string, 0, len(phases))
	if hasString(phases, "beforeRequestSent") {
		events = append(events, listenerBeforeRequestSent)
	}
	if hasString(phases, "responseStarted") {
		events = append(events, "network.responseStarted")
	}
	if hasString(phases, "authRequired") {
		events = append(events, "network.authRequired")
	}

	subscriptionID := ""
	if len(events) > 0 {
		subscription, err := bidi.Subscribe(driver, events, contexts, timeout)
		if err != nil {
			_, _ = bidi.RemoveIntercept(driver, stringifyNetworkValue(result["intercept"]), timeout)
			if requestCollector != nil {
				_ = requestCollector.Remove()
			}
			if responseCollector != nil {
				_ = responseCollector.Remove()
			}
			return nil, err
		}
		subscriptionID = subscription.Subscription
	}

	if hasString(phases, "beforeRequestSent") {
		if err := callbackDriver.SetCallback(listenerBeforeRequestSent, i.onIntercept, false); err != nil {
			i.cleanupStart(stringifyNetworkValue(result["intercept"]), subscriptionID, requestCollector, responseCollector)
			return nil, err
		}
	}
	if hasString(phases, "responseStarted") {
		if err := callbackDriver.SetCallback("network.responseStarted", i.onResponseIntercept, false); err != nil {
			callbackDriver.RemoveCallback(listenerBeforeRequestSent, false)
			i.cleanupStart(stringifyNetworkValue(result["intercept"]), subscriptionID, requestCollector, responseCollector)
			return nil, err
		}
	}
	if hasString(phases, "authRequired") {
		if err := callbackDriver.SetCallback("network.authRequired", i.onAuth, false); err != nil {
			callbackDriver.RemoveCallback(listenerBeforeRequestSent, false)
			callbackDriver.RemoveCallback("network.responseStarted", false)
			i.cleanupStart(stringifyNetworkValue(result["intercept"]), subscriptionID, requestCollector, responseCollector)
			return nil, err
		}
	}

	i.mu.Lock()
	i.active = true
	i.interceptID = stringifyNetworkValue(result["intercept"])
	i.subscriptionID = subscriptionID
	i.requestCollector = requestCollector
	i.responseCollector = responseCollector
	i.handler = handler
	i.queue = newPacketQueue[*InterceptedRequest](128)
	i.phases = append([]string(nil), phases...)
	i.mu.Unlock()
	return i, nil
}

// StartRequests 仅拦截 beforeRequestSent。
func (i *Interceptor) StartRequests(
	handler func(*InterceptedRequest),
	urlPatterns []map[string]any,
	collectResponse ...bool,
) (*Interceptor, error) {
	return i.Start(handler, urlPatterns, []string{"beforeRequestSent"}, collectResponse...)
}

// StartResponses 仅拦截 responseStarted。
func (i *Interceptor) StartResponses(
	handler func(*InterceptedRequest),
	urlPatterns []map[string]any,
	collectResponse ...bool,
) (*Interceptor, error) {
	enabled := true
	if len(collectResponse) > 0 {
		enabled = collectResponse[0]
	}
	return i.Start(handler, urlPatterns, []string{"responseStarted"}, enabled)
}

// Stop 停止当前拦截。
func (i *Interceptor) Stop() {
	if i == nil || i.owner == nil {
		return
	}

	i.mu.Lock()
	interceptID := i.interceptID
	subscriptionID := i.subscriptionID
	requestCollector := i.requestCollector
	responseCollector := i.responseCollector
	wasActive := i.active
	i.active = false
	i.interceptID = ""
	i.subscriptionID = ""
	i.requestCollector = nil
	i.responseCollector = nil
	i.handler = nil
	i.phases = nil
	i.mu.Unlock()

	if !wasActive {
		return
	}

	callbackDriver := i.owner.Driver()
	callbackDriver.RemoveCallback(listenerBeforeRequestSent, false)
	callbackDriver.RemoveCallback("network.responseStarted", false)
	callbackDriver.RemoveCallback("network.authRequired", false)

	timeout := i.resolveTimeout()
	if interceptID != "" {
		_, _ = bidi.RemoveIntercept(i.owner.BrowserDriver(), interceptID, timeout)
	}
	if subscriptionID != "" {
		_ = bidi.Unsubscribe(i.owner.BrowserDriver(), nil, nil, []string{subscriptionID}, timeout)
	}
	if requestCollector != nil {
		_ = requestCollector.Remove()
	}
	if responseCollector != nil {
		_ = responseCollector.Remove()
	}
}

// Wait 等待一个被拦截的请求；超时返回 nil。
func (i *Interceptor) Wait(timeout time.Duration) *InterceptedRequest {
	if i == nil {
		return nil
	}
	value, ok := i.queue.Pull(timeout)
	if !ok || value == nil {
		return nil
	}
	return value
}

func (i *Interceptor) cleanupStart(
	interceptID string,
	subscriptionID string,
	requestCollector *DataCollector,
	responseCollector *DataCollector,
) {
	timeout := i.resolveTimeout()
	if interceptID != "" {
		_, _ = bidi.RemoveIntercept(i.owner.BrowserDriver(), interceptID, timeout)
	}
	if subscriptionID != "" {
		_ = bidi.Unsubscribe(i.owner.BrowserDriver(), nil, nil, []string{subscriptionID}, timeout)
	}
	if requestCollector != nil {
		_ = requestCollector.Remove()
	}
	if responseCollector != nil {
		_ = responseCollector.Remove()
	}
}

func (i *Interceptor) resolveTimeout() time.Duration {
	if i == nil || i.owner == nil {
		return networkDefaultTimeout()
	}
	timeout := i.owner.BaseTimeout()
	if timeout <= 0 {
		return networkDefaultTimeout()
	}
	return timeout
}

func (i *Interceptor) currentHandler() func(*InterceptedRequest) {
	if i == nil {
		return nil
	}
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.handler
}

func (i *Interceptor) currentCollector() *DataCollector {
	if i == nil {
		return nil
	}
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.requestCollector
}

func (i *Interceptor) currentResponseCollector() *DataCollector {
	if i == nil {
		return nil
	}
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.responseCollector
}

func (i *Interceptor) isActive() bool {
	if i == nil {
		return false
	}
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.active
}

func (i *Interceptor) onIntercept(params map[string]any) {
	if !i.isActive() {
		return
	}
	req := NewInterceptedRequest(
		params,
		i.owner.BrowserDriver(),
		i.currentCollector(),
		i.currentResponseCollector(),
		i.resolveTimeout(),
	)
	if handler := i.currentHandler(); handler != nil {
		safeRunInterceptHandler(handler, req)
		if !req.Handled() {
			_ = req.ContinueRequest("", "", nil, nil)
		}
		return
	}
	i.queue.Push(req)
}

func (i *Interceptor) onResponseIntercept(params map[string]any) {
	if !i.isActive() {
		return
	}
	req := NewInterceptedRequest(
		params,
		i.owner.BrowserDriver(),
		i.currentCollector(),
		i.currentResponseCollector(),
		i.resolveTimeout(),
	)
	if handler := i.currentHandler(); handler != nil {
		safeRunInterceptHandler(handler, req)
		if !req.Handled() {
			_ = req.ContinueResponse(nil, "", nil)
		}
		return
	}
	i.queue.Push(req)
}

func (i *Interceptor) onAuth(params map[string]any) {
	if !i.isActive() {
		return
	}
	req := NewInterceptedRequest(
		params,
		i.owner.BrowserDriver(),
		i.currentCollector(),
		i.currentResponseCollector(),
		i.resolveTimeout(),
	)
	if handler := i.currentHandler(); handler != nil {
		safeRunInterceptHandler(handler, req)
		if !req.Handled() {
			_ = req.ContinueWithAuth("default", "", "")
		}
		return
	}
	i.queue.Push(req)
}

func safeRunInterceptHandler(handler func(*InterceptedRequest), req *InterceptedRequest) {
	defer func() {
		_ = recover()
	}()
	handler(req)
}

func hasString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
