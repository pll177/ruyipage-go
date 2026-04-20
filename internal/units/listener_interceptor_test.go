package units

import (
	"reflect"
	"testing"
	"time"

	"github.com/pll177/ruyipage-go/internal/base"
)

type testNetworkOwner struct {
	contextID string
	browser   *base.BrowserBiDiDriver
	driver    *base.ContextDriver
	timeout   time.Duration
}

func newTestNetworkOwner(address string, contextID string) *testNetworkOwner {
	browser := base.NewBrowserBiDiDriver(address)
	return &testNetworkOwner{
		contextID: contextID,
		browser:   browser,
		driver:    base.NewContextDriver(browser, contextID),
		timeout:   2 * time.Second,
	}
}

func (o *testNetworkOwner) ContextID() string {
	return o.contextID
}

func (o *testNetworkOwner) BrowserDriver() *base.BrowserBiDiDriver {
	return o.browser
}

func (o *testNetworkOwner) Driver() *base.ContextDriver {
	return o.driver
}

func (o *testNetworkOwner) BaseTimeout() time.Duration {
	return o.timeout
}

func TestListenerStopRemovesOnlyCurrentContextCallbacks(t *testing.T) {
	owner1 := newTestNetworkOwner("listener-stop-1", "tab-1")
	owner2 := newTestNetworkOwner("listener-stop-1", "tab-2")

	listener1 := NewListener(owner1)
	listener2 := NewListener(owner2)

	if err := owner1.Driver().SetCallback(listenerResponseCompleted, listener1.onResponseCompleted, false); err != nil {
		t.Fatalf("set callback for listener1 response: %v", err)
	}
	if err := owner1.Driver().SetCallback(listenerFetchError, listener1.onFetchError, false); err != nil {
		t.Fatalf("set callback for listener1 fetch error: %v", err)
	}
	if err := owner2.Driver().SetCallback(listenerResponseCompleted, listener2.onResponseCompleted, false); err != nil {
		t.Fatalf("set callback for listener2 response: %v", err)
	}
	if err := owner2.Driver().SetCallback(listenerFetchError, listener2.onFetchError, false); err != nil {
		t.Fatalf("set callback for listener2 fetch error: %v", err)
	}

	listener1.listening = true
	listener2.listening = true

	listener1.Stop()

	assertDriverHasCallbackRoute(t, owner1.BrowserDriver(), listenerResponseCompleted, "tab-1", false)
	assertDriverHasCallbackRoute(t, owner1.BrowserDriver(), listenerFetchError, "tab-1", false)
	assertDriverHasCallbackRoute(t, owner1.BrowserDriver(), listenerResponseCompleted, "tab-2", true)
	assertDriverHasCallbackRoute(t, owner1.BrowserDriver(), listenerFetchError, "tab-2", true)
}

func TestInterceptorStopRemovesOnlyCurrentContextCallbacks(t *testing.T) {
	owner1 := newTestNetworkOwner("interceptor-stop-1", "tab-1")
	owner2 := newTestNetworkOwner("interceptor-stop-1", "tab-2")

	interceptor1 := NewInterceptor(owner1)
	interceptor2 := NewInterceptor(owner2)

	if err := owner1.Driver().SetCallback(listenerBeforeRequestSent, interceptor1.onIntercept, false); err != nil {
		t.Fatalf("set callback for interceptor1 request: %v", err)
	}
	if err := owner1.Driver().SetCallback("network.responseStarted", interceptor1.onResponseIntercept, false); err != nil {
		t.Fatalf("set callback for interceptor1 response: %v", err)
	}
	if err := owner1.Driver().SetCallback("network.authRequired", interceptor1.onAuth, false); err != nil {
		t.Fatalf("set callback for interceptor1 auth: %v", err)
	}
	if err := owner2.Driver().SetCallback(listenerBeforeRequestSent, interceptor2.onIntercept, false); err != nil {
		t.Fatalf("set callback for interceptor2 request: %v", err)
	}
	if err := owner2.Driver().SetCallback("network.responseStarted", interceptor2.onResponseIntercept, false); err != nil {
		t.Fatalf("set callback for interceptor2 response: %v", err)
	}
	if err := owner2.Driver().SetCallback("network.authRequired", interceptor2.onAuth, false); err != nil {
		t.Fatalf("set callback for interceptor2 auth: %v", err)
	}

	interceptor1.active = true
	interceptor2.active = true

	interceptor1.Stop()

	assertDriverHasCallbackRoute(t, owner1.BrowserDriver(), listenerBeforeRequestSent, "tab-1", false)
	assertDriverHasCallbackRoute(t, owner1.BrowserDriver(), "network.responseStarted", "tab-1", false)
	assertDriverHasCallbackRoute(t, owner1.BrowserDriver(), "network.authRequired", "tab-1", false)
	assertDriverHasCallbackRoute(t, owner1.BrowserDriver(), listenerBeforeRequestSent, "tab-2", true)
	assertDriverHasCallbackRoute(t, owner1.BrowserDriver(), "network.responseStarted", "tab-2", true)
	assertDriverHasCallbackRoute(t, owner1.BrowserDriver(), "network.authRequired", "tab-2", true)
}

func assertDriverHasCallbackRoute(t *testing.T, driver *base.BrowserBiDiDriver, event string, context string, want bool) {
	t.Helper()

	got := driverHasCallbackRoute(driver, event, context)
	if got != want {
		t.Fatalf("callback route (%s, %s) present=%v, want %v", event, context, got, want)
	}
}

func driverHasCallbackRoute(driver *base.BrowserBiDiDriver, event string, context string) bool {
	if driver == nil {
		return false
	}

	value := reflect.ValueOf(driver).Elem().FieldByName("callbacks")
	iter := value.MapRange()
	for iter.Next() {
		key := iter.Key()
		if key.FieldByName("event").String() == event && key.FieldByName("context").String() == context {
			return true
		}
	}
	return false
}
