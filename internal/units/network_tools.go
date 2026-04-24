package units

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pll177/ruyipage-go/internal/base"
	"github.com/pll177/ruyipage-go/internal/bidi"
	"github.com/pll177/ruyipage-go/internal/support"
)

type networkOwner interface {
	ContextID() string
	BrowserDriver() *base.BrowserBiDiDriver
	Driver() *base.ContextDriver
	BaseTimeout() time.Duration
}

// NetworkData 表示单次 network.getData 结果。
type NetworkData struct {
	Raw    map[string]any
	Bytes  any
	Base64 any
}

// NewNetworkData 从底层返回构建高层结果对象。
func NewNetworkData(data map[string]any) *NetworkData {
	raw := cloneNetworkMapDeep(data)
	return &NetworkData{
		Raw:    raw,
		Bytes:  cloneNetworkValueDeep(raw["bytes"]),
		Base64: cloneNetworkValueDeep(raw["base64"]),
	}
}

// HasData 返回当前结果是否包含有效数据。
func (d *NetworkData) HasData() bool {
	if d == nil {
		return false
	}
	return d.Bytes != nil || d.Base64 != nil
}

// DataCollector 表示 network.addDataCollector 返回的收集器句柄。
type DataCollector struct {
	manager *NetworkManager
	ID      string
}

// Get 读取指定请求的数据。
func (c *DataCollector) Get(requestID string, dataType string) (*NetworkData, error) {
	return c.getWithTimeout(requestID, dataType, 0)
}

func (c *DataCollector) getWithTimeout(requestID string, dataType string, timeout time.Duration) (*NetworkData, error) {
	if c == nil || c.manager == nil {
		return NewNetworkData(nil), nil
	}
	return c.manager.getDataWithTimeout(c.ID, requestID, dataType, timeout)
}

// Disown 释放指定请求的数据。
func (c *DataCollector) Disown(requestID string, dataType string) error {
	if c == nil || c.manager == nil {
		return nil
	}
	return c.manager.DisownData(c.ID, requestID, dataType)
}

// Remove 移除当前 collector。
func (c *DataCollector) Remove() error {
	if c == nil || c.manager == nil {
		return nil
	}
	return c.manager.RemoveDataCollector(c.ID)
}

// NetworkManager 提供额外请求头、缓存行为与 collector 管理入口。
type NetworkManager struct {
	owner networkOwner
}

// NewNetworkManager 创建高层 network 管理器。
func NewNetworkManager(owner networkOwner) *NetworkManager {
	return &NetworkManager{owner: owner}
}

// SetExtraHeaders 设置当前 context 后续请求的额外请求头。
func (m *NetworkManager) SetExtraHeaders(headers any) error {
	if m == nil || m.owner == nil {
		return nil
	}
	_, err := bidi.SetExtraHeaders(
		m.owner.BrowserDriver(),
		normalizeBiDiHeaders(headers),
		[]string{m.owner.ContextID()},
		m.resolveTimeout(),
	)
	return err
}

// ClearExtraHeaders 清空当前 context 的额外请求头。
func (m *NetworkManager) ClearExtraHeaders() error {
	if m == nil || m.owner == nil {
		return nil
	}
	_, err := bidi.SetExtraHeaders(
		m.owner.BrowserDriver(),
		[]map[string]any{},
		[]string{m.owner.ContextID()},
		m.resolveTimeout(),
	)
	return err
}

// SetCacheBehavior 设置当前 context 的缓存策略。
func (m *NetworkManager) SetCacheBehavior(behavior string) error {
	if m == nil || m.owner == nil {
		return nil
	}
	_, err := bidi.SetCacheBehavior(
		m.owner.BrowserDriver(),
		behavior,
		[]string{m.owner.ContextID()},
		m.resolveTimeout(),
	)
	return err
}

// AddDataCollector 注册数据收集器。
func (m *NetworkManager) AddDataCollector(events any, dataTypes any, maxEncodedDataSize int) (*DataCollector, error) {
	if m == nil || m.owner == nil {
		return &DataCollector{}, nil
	}
	result, err := bidi.AddDataCollector(
		m.owner.BrowserDriver(),
		events,
		[]string{m.owner.ContextID()},
		maxEncodedDataSize,
		dataTypes,
		m.resolveTimeout(),
	)
	if err != nil {
		return nil, err
	}
	return &DataCollector{
		manager: m,
		ID:      stringifyNetworkValue(result["collector"]),
	}, nil
}

// RemoveDataCollector 移除指定 collector。
func (m *NetworkManager) RemoveDataCollector(collectorID string) error {
	if m == nil || m.owner == nil {
		return nil
	}
	_, err := bidi.RemoveDataCollector(m.owner.BrowserDriver(), collectorID, m.resolveTimeout())
	return err
}

// GetData 读取指定 collector 中的数据。
func (m *NetworkManager) GetData(collectorID string, requestID string, dataType string) (*NetworkData, error) {
	return m.getDataWithTimeout(collectorID, requestID, dataType, 0)
}

func (m *NetworkManager) getDataWithTimeout(collectorID string, requestID string, dataType string, timeout time.Duration) (*NetworkData, error) {
	if m == nil || m.owner == nil {
		return NewNetworkData(nil), nil
	}
	if timeout <= 0 {
		timeout = m.resolveTimeout()
	}
	result, err := bidi.GetData(m.owner.BrowserDriver(), collectorID, requestID, dataType, timeout)
	if err != nil {
		return nil, err
	}
	return NewNetworkData(result), nil
}

// DisownData 释放指定 collector 中的数据。
func (m *NetworkManager) DisownData(collectorID string, requestID string, dataType string) error {
	if m == nil || m.owner == nil {
		return nil
	}
	_, err := bidi.DisownData(m.owner.BrowserDriver(), collectorID, requestID, dataType, m.resolveTimeout())
	return err
}

func (m *NetworkManager) resolveTimeout() time.Duration {
	if m == nil || m.owner == nil {
		return networkDefaultTimeout()
	}
	timeout := m.owner.BaseTimeout()
	if timeout <= 0 {
		return networkDefaultTimeout()
	}
	return timeout
}

func networkDefaultTimeout() time.Duration {
	seconds := support.Settings.BiDiTimeout
	if seconds <= 0 {
		seconds = float64(support.DefaultBiDiTimeoutSeconds)
	}
	return time.Duration(seconds * float64(time.Second))
}

func normalizeBiDiHeaders(headers any) any {
	switch typed := headers.(type) {
	case nil:
		return nil
	case map[string]string:
		return headerRowsFromStringMap(typed)
	case map[string]any:
		return headerRowsFromAnyMap(typed)
	default:
		return cloneNetworkValueDeep(headers)
	}
}

func headerRowsFromStringMap(headers map[string]string) []map[string]any {
	if len(headers) == 0 {
		return []map[string]any{}
	}
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	rows := make([]map[string]any, 0, len(keys))
	for _, key := range keys {
		rows = append(rows, map[string]any{
			"name":  key,
			"value": map[string]any{"type": "string", "value": headers[key]},
		})
	}
	return rows
}

func headerRowsFromAnyMap(headers map[string]any) []map[string]any {
	if len(headers) == 0 {
		return []map[string]any{}
	}
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	rows := make([]map[string]any, 0, len(keys))
	for _, key := range keys {
		rows = append(rows, map[string]any{
			"name":  key,
			"value": map[string]any{"type": "string", "value": fmt.Sprint(headers[key])},
		})
	}
	return rows
}

func bidiHeadersToMap(headers any, lowerName bool) map[string]string {
	rows := make(map[string]string)
	items, ok := headers.([]any)
	if !ok {
		if typed, typedOK := headers.([]map[string]any); typedOK {
			items = make([]any, 0, len(typed))
			for _, item := range typed {
				items = append(items, item)
			}
		}
	}
	for _, item := range items {
		header, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name := stringifyNetworkValue(header["name"])
		if lowerName {
			name = strings.ToLower(name)
		}
		value := header["value"]
		if valueMap, ok := value.(map[string]any); ok {
			value = valueMap["value"]
		}
		rows[name] = stringifyNetworkValue(value)
	}
	return rows
}

func normalizeBytesValue(value any, preferBase64 bool) any {
	switch typed := value.(type) {
	case nil:
		return nil
	case map[string]any:
		return cloneNetworkMapDeep(typed)
	case []byte:
		return map[string]any{
			"type":  "base64",
			"value": base64.StdEncoding.EncodeToString(typed),
		}
	case string:
		if preferBase64 {
			return map[string]any{
				"type":  "base64",
				"value": base64.StdEncoding.EncodeToString([]byte(typed)),
			}
		}
		return map[string]any{
			"type":  "string",
			"value": typed,
		}
	default:
		return map[string]any{
			"type":  "string",
			"value": fmt.Sprint(typed),
		}
	}
}

func decodeNetworkBodyValue(value any) (string, bool) {
	switch typed := value.(type) {
	case nil:
		return "", false
	case string:
		return typed, true
	case []byte:
		return string(typed), true
	case map[string]any:
		valueType := stringifyNetworkValue(typed["type"])
		rawValue := typed["value"]
		switch valueType {
		case "string":
			return stringifyNetworkValue(rawValue), true
		case "base64":
			text := stringifyNetworkValue(rawValue)
			decoded, err := base64.StdEncoding.DecodeString(text)
			if err != nil {
				return text, true
			}
			return string(decoded), true
		default:
			if rawValue == nil {
				return "", false
			}
			return stringifyNetworkValue(rawValue), true
		}
	default:
		return fmt.Sprint(typed), true
	}
}

func cloneNetworkMapDeep(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}
	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = cloneNetworkValueDeep(value)
	}
	return cloned
}

func cloneNetworkSliceDeep(values []any) []any {
	if values == nil {
		return nil
	}
	cloned := make([]any, len(values))
	for index, value := range values {
		cloned[index] = cloneNetworkValueDeep(value)
	}
	return cloned
}

func cloneNetworkStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneNetworkValueDeep(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneNetworkMapDeep(typed)
	case []map[string]any:
		rows := make([]map[string]any, len(typed))
		for index, row := range typed {
			rows[index] = cloneNetworkMapDeep(row)
		}
		return rows
	case []string:
		rows := make([]string, len(typed))
		copy(rows, typed)
		return rows
	case []any:
		return cloneNetworkSliceDeep(typed)
	default:
		return typed
	}
}

func stringifyNetworkValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case nil:
		return ""
	default:
		return fmt.Sprint(typed)
	}
}

func intNetworkValue(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int8:
		return int(typed)
	case int16:
		return int(typed)
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case uint:
		return int(typed)
	case uint8:
		return int(typed)
	case uint16:
		return int(typed)
	case uint32:
		return int(typed)
	case uint64:
		return int(typed)
	case float32:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
	}
}

func boolNetworkValue(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return typed == "true"
	default:
		return false
	}
}

type packetQueue[T any] struct {
	mu    sync.Mutex
	items chan T
}

func newPacketQueue[T any](size int) *packetQueue[T] {
	if size <= 0 {
		size = 128
	}
	return &packetQueue[T]{
		items: make(chan T, size),
	}
}

func (q *packetQueue[T]) Push(value T) {
	if q == nil {
		return
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	select {
	case q.items <- value:
	default:
		<-q.items
		q.items <- value
	}
}

func (q *packetQueue[T]) Pull(timeout time.Duration) (T, bool) {
	var zero T
	if q == nil {
		return zero, false
	}
	if timeout <= 0 {
		timeout = networkDefaultTimeout()
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case value := <-q.items:
		return value, true
	case <-timer.C:
		return zero, false
	}
}

func (q *packetQueue[T]) Clear() {
	if q == nil {
		return
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	for {
		select {
		case <-q.items:
		default:
			return
		}
	}
}
