package ruyipage

import internalunits "ruyipage-go/internal/units"

type (
	// CookieInfo 表示单个 Cookie 的公开信息。
	CookieInfo = internalunits.CookieInfo
	// CookiesSetter 表示高层 Cookie 写入/删除管理器。
	CookiesSetter = internalunits.CookiesSetter
	// StorageManager 表示 local/session storage 管理器。
	StorageManager = internalunits.StorageManager
)
