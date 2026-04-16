package ruyipage

import internalunits "github.com/pll177/ruyipage-go/internal/units"

type (
	// DataPacket 表示一次网络抓包结果。
	DataPacket = internalunits.DataPacket
	// Listener 表示高层网络监听器。
	Listener = internalunits.Listener
	// InterceptedRequest 表示被拦截的网络请求。
	InterceptedRequest = internalunits.InterceptedRequest
	// Interceptor 表示高层网络拦截器。
	Interceptor = internalunits.Interceptor
	// NetworkData 表示 network.getData 的高层结果。
	NetworkData = internalunits.NetworkData
	// DataCollector 表示网络数据收集器句柄。
	DataCollector = internalunits.DataCollector
	// NetworkManager 表示 network 模块高层管理器。
	NetworkManager = internalunits.NetworkManager
)

// Listen 返回页面级网络监听器。
func (p *FirefoxBase) Listen() *Listener {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.Listen()
}

// Intercept 返回页面级网络拦截器。
func (p *FirefoxBase) Intercept() *Interceptor {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.Intercept()
}

// Network 返回页面级 network 高层管理器。
func (p *FirefoxBase) Network() *NetworkManager {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.Network()
}
