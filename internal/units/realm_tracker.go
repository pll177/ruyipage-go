package units

import (
	"sort"
	"sync"

	"ruyipage-go/internal/bidi"
)

// RealmTracker 提供 realm 生命周期跟踪能力。
type RealmTracker struct {
	owner networkOwner

	mu             sync.RWMutex
	listening      bool
	subscriptionID string
	realms         map[string]RealmInfo
	created        func(RealmInfo)
	destroyed      func(string)
}

// NewRealmTracker 创建 realm 生命周期跟踪器。
func NewRealmTracker(owner networkOwner) *RealmTracker {
	return &RealmTracker{
		owner:  owner,
		realms: make(map[string]RealmInfo),
	}
}

// Listening 返回当前是否正在跟踪 realm 事件。
func (t *RealmTracker) Listening() bool {
	if t == nil {
		return false
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.listening
}

// Start 启动 realm 生命周期跟踪。
func (t *RealmTracker) Start() error {
	if t == nil || t.owner == nil {
		return nil
	}

	t.Stop()

	result, err := bidi.Subscribe(
		t.owner.BrowserDriver(),
		[]string{"script.realmCreated", "script.realmDestroyed"},
		[]string{t.owner.ContextID()},
		resolveUnitTimeout(t.owner),
	)
	if err != nil {
		return err
	}

	driver := t.owner.Driver()
	if err := driver.SetGlobalCallback("script.realmCreated", t.onCreated, false); err != nil {
		_ = bidi.Unsubscribe(t.owner.BrowserDriver(), nil, nil, []string{result.Subscription}, resolveUnitTimeout(t.owner))
		return err
	}
	if err := driver.SetGlobalCallback("script.realmDestroyed", t.onDestroyed, false); err != nil {
		driver.RemoveGlobalCallback("script.realmCreated", false)
		_ = bidi.Unsubscribe(t.owner.BrowserDriver(), nil, nil, []string{result.Subscription}, resolveUnitTimeout(t.owner))
		return err
	}

	realms := make(map[string]RealmInfo)
	if values, listErr := bidi.GetRealms(t.owner.BrowserDriver(), t.owner.ContextID(), "", resolveUnitTimeout(t.owner)); listErr == nil {
		for _, value := range values {
			info := NewRealmInfoFromData(value)
			realms[info.Realm] = info
		}
	}

	t.mu.Lock()
	t.listening = true
	t.subscriptionID = result.Subscription
	t.realms = realms
	t.mu.Unlock()
	return nil
}

// Stop 停止 realm 跟踪。
func (t *RealmTracker) Stop() {
	if t == nil || t.owner == nil {
		return
	}

	t.mu.Lock()
	subscriptionID := t.subscriptionID
	wasListening := t.listening
	t.listening = false
	t.subscriptionID = ""
	t.mu.Unlock()

	if !wasListening {
		return
	}

	driver := t.owner.Driver()
	driver.RemoveGlobalCallback("script.realmCreated", false)
	driver.RemoveGlobalCallback("script.realmDestroyed", false)
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

// List 返回当前已知 realm 列表。
func (t *RealmTracker) List() []RealmInfo {
	if t == nil {
		return nil
	}
	t.mu.RLock()
	defer t.mu.RUnlock()

	ids := make([]string, 0, len(t.realms))
	for id := range t.realms {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	result := make([]RealmInfo, 0, len(ids))
	for _, id := range ids {
		result = append(result, cloneRealmInfo(t.realms[id]))
	}
	return result
}

// OnCreated 注册 realm 创建回调；传 nil 表示移除。
func (t *RealmTracker) OnCreated(callback func(RealmInfo)) *RealmTracker {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	t.created = callback
	t.mu.Unlock()
	return t
}

// OnDestroyed 注册 realm 销毁回调；传 nil 表示移除。
func (t *RealmTracker) OnDestroyed(callback func(string)) *RealmTracker {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	t.destroyed = callback
	t.mu.Unlock()
	return t
}

func (t *RealmTracker) onCreated(params map[string]any) {
	if t == nil {
		return
	}
	contextID := stringifyNetworkValue(params["context"])
	if contextID != "" && contextID != t.owner.ContextID() {
		return
	}

	info := NewRealmInfo(params)

	t.mu.Lock()
	if !t.listening {
		t.mu.Unlock()
		return
	}
	if t.realms == nil {
		t.realms = make(map[string]RealmInfo)
	}
	t.realms[info.Realm] = cloneRealmInfo(info)
	callback := t.created
	t.mu.Unlock()

	if callback != nil {
		safeRunRealmCreatedCallback(callback, cloneRealmInfo(info))
	}
}

func (t *RealmTracker) onDestroyed(params map[string]any) {
	if t == nil {
		return
	}
	realmID := stringifyNetworkValue(params["realm"])
	if realmID == "" {
		return
	}

	t.mu.Lock()
	if !t.listening {
		t.mu.Unlock()
		return
	}
	delete(t.realms, realmID)
	callback := t.destroyed
	t.mu.Unlock()

	if callback != nil {
		safeRunRealmDestroyedCallback(callback, realmID)
	}
}

func cloneRealmInfo(value RealmInfo) RealmInfo {
	return RealmInfo{
		Raw:     cloneNetworkMapDeep(value.Raw),
		Realm:   value.Realm,
		Type:    value.Type,
		Context: value.Context,
		Origin:  value.Origin,
	}
}

func safeRunRealmCreatedCallback(callback func(RealmInfo), info RealmInfo) {
	defer func() {
		_ = recover()
	}()
	callback(info)
}

func safeRunRealmDestroyedCallback(callback func(string), realmID string) {
	defer func() {
		_ = recover()
	}()
	callback(realmID)
}
