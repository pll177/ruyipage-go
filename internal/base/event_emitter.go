package base

import (
	"errors"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
)

var (
	// ErrEventEmitterClosed 表示事件总线已经关闭，不能再注册或广播。
	ErrEventEmitterClosed = errors.New("event emitter 已关闭")
	// ErrNilEventCallback 表示事件处理器不能为空。
	ErrNilEventCallback = errors.New("event callback 不能为空")
)

const (
	subscriptionStateActive uint32 = iota
	subscriptionStateRemoved
)

// EventCallback 是事件回调。
type EventCallback func(params map[string]any)

type eventRoute struct {
	event   string
	context string
}

// EventSubscription 表示一次事件订阅。
type EventSubscription struct {
	emitter  *EventEmitter
	route    eventRoute
	id       uint64
	handler  EventCallback
	identity uintptr
	once     bool
	state    atomic.Uint32
}

// Active 返回当前订阅是否仍可接收事件。
func (s *EventSubscription) Active() bool {
	return s != nil && s.state.Load() == subscriptionStateActive
}

// Cancel 注销当前订阅。
func (s *EventSubscription) Cancel() bool {
	if s == nil || s.emitter == nil {
		return false
	}
	return s.emitter.Off(s)
}

func (s *EventSubscription) markRemoved() bool {
	return s != nil && s.state.CompareAndSwap(subscriptionStateActive, subscriptionStateRemoved)
}

func (s *EventSubscription) invoke(params map[string]any) {
	if s == nil {
		return
	}

	if s.once {
		if !s.markRemoved() {
			return
		}
		s.emitter.detachSubscription(s)
		s.safeCall(params)
		return
	}

	if s.state.Load() != subscriptionStateActive {
		return
	}
	s.safeCall(params)
}

func (s *EventSubscription) safeCall(params map[string]any) {
	defer func() {
		_ = recover()
	}()
	s.handler(params)
}

// EventEmitter 提供并发安全的事件订阅与广播。
type EventEmitter struct {
	mu     sync.RWMutex
	closed bool
	nextID atomic.Uint64
	routes map[eventRoute]map[uint64]*EventSubscription
	dedup  map[eventRoute]map[uintptr]*EventSubscription
}

// NewEventEmitter 创建事件总线。
func NewEventEmitter() *EventEmitter {
	return &EventEmitter{
		routes: make(map[eventRoute]map[uint64]*EventSubscription),
		dedup:  make(map[eventRoute]map[uintptr]*EventSubscription),
	}
}

// On 注册持续事件监听。
func (e *EventEmitter) On(event string, context string, handler EventCallback) (*EventSubscription, error) {
	return e.add(event, context, handler, false)
}

// Once 注册一次性事件监听。
func (e *EventEmitter) Once(event string, context string, handler EventCallback) (*EventSubscription, error) {
	return e.add(event, context, handler, true)
}

// Off 注销订阅；重复注销安全返回 false。
func (e *EventEmitter) Off(subscription *EventSubscription) bool {
	if subscription == nil || subscription.emitter != e {
		return false
	}
	if !subscription.markRemoved() {
		return false
	}

	e.detachSubscription(subscription)
	return true
}

// Emit 向匹配的订阅广播事件。
func (e *EventEmitter) Emit(event string, context string, params map[string]any) error {
	if params == nil {
		params = map[string]any{}
	}

	subscriptions, err := e.snapshot(event, context)
	if err != nil {
		return err
	}

	for _, subscription := range subscriptions {
		subscription.invoke(params)
	}
	return nil
}

// Close 关闭事件总线并释放全部订阅。
func (e *EventEmitter) Close() error {
	if e == nil {
		return nil
	}

	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return nil
	}
	e.closed = true

	routes := e.routes
	e.routes = make(map[eventRoute]map[uint64]*EventSubscription)
	e.dedup = make(map[eventRoute]map[uintptr]*EventSubscription)
	e.mu.Unlock()

	for _, routeSubscriptions := range routes {
		for _, subscription := range routeSubscriptions {
			subscription.markRemoved()
		}
	}
	return nil
}

func (e *EventEmitter) add(event string, context string, handler EventCallback, once bool) (*EventSubscription, error) {
	if handler == nil {
		return nil, ErrNilEventCallback
	}

	route := eventRoute{event: event, context: context}
	identity := handlerIdentity(handler)

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return nil, ErrEventEmitterClosed
	}

	if existing := e.getExistingSubscription(route, identity); existing != nil {
		return existing, nil
	}

	subscription := &EventSubscription{
		emitter:  e,
		route:    route,
		id:       e.nextID.Add(1),
		handler:  handler,
		identity: identity,
		once:     once,
	}
	subscription.state.Store(subscriptionStateActive)

	if e.routes[route] == nil {
		e.routes[route] = make(map[uint64]*EventSubscription)
	}
	if e.dedup[route] == nil {
		e.dedup[route] = make(map[uintptr]*EventSubscription)
	}

	e.routes[route][subscription.id] = subscription
	e.dedup[route][identity] = subscription
	return subscription, nil
}

func (e *EventEmitter) getExistingSubscription(route eventRoute, identity uintptr) *EventSubscription {
	dedupForRoute := e.dedup[route]
	if dedupForRoute == nil {
		return nil
	}

	subscription := dedupForRoute[identity]
	if subscription == nil || !subscription.Active() {
		return nil
	}
	return subscription
}

func (e *EventEmitter) snapshot(event string, context string) ([]*EventSubscription, error) {
	if e == nil {
		return nil, ErrEventEmitterClosed
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.closed {
		return nil, ErrEventEmitterClosed
	}

	exactRoute := eventRoute{event: event, context: context}
	subscriptions := make([]*EventSubscription, 0, len(e.routes[exactRoute]))
	subscriptions = appendSubscriptions(subscriptions, e.routes[exactRoute])

	if context != "" {
		globalRoute := eventRoute{event: event, context: ""}
		subscriptions = appendSubscriptions(subscriptions, e.routes[globalRoute])
	}

	sort.Slice(subscriptions, func(i, j int) bool {
		return subscriptions[i].id < subscriptions[j].id
	})

	return subscriptions, nil
}

func (e *EventEmitter) detachSubscription(subscription *EventSubscription) {
	if e == nil || subscription == nil {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	routeSubscriptions := e.routes[subscription.route]
	if routeSubscriptions != nil {
		if existing := routeSubscriptions[subscription.id]; existing == subscription {
			delete(routeSubscriptions, subscription.id)
		}
		if len(routeSubscriptions) == 0 {
			delete(e.routes, subscription.route)
		}
	}

	dedupForRoute := e.dedup[subscription.route]
	if dedupForRoute != nil {
		if existing := dedupForRoute[subscription.identity]; existing == subscription {
			delete(dedupForRoute, subscription.identity)
		}
		if len(dedupForRoute) == 0 {
			delete(e.dedup, subscription.route)
		}
	}
}

func appendSubscriptions(target []*EventSubscription, routeSubscriptions map[uint64]*EventSubscription) []*EventSubscription {
	for _, subscription := range routeSubscriptions {
		target = append(target, subscription)
	}
	return target
}

func handlerIdentity(handler EventCallback) uintptr {
	return reflect.ValueOf(handler).Pointer()
}
