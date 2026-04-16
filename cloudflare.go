package ruyipage

import "time"

// HandleCloudflareChallenge 自动尝试处理常见 Cloudflare Turnstile 验证。
func (p *FirefoxBase) HandleCloudflareChallenge(timeout time.Duration, checkInterval time.Duration) bool {
	if p == nil || p.inner == nil {
		return false
	}
	return p.inner.HandleCloudflareChallenge(timeout, checkInterval)
}
