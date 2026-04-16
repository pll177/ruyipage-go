package support

// ByValues 表示公开可复用的定位器类型常量集。
type ByValues struct {
	CSS           string
	XPATH         string
	TEXT          string
	INNER_TEXT    string
	ACCESSIBILITY string
	ID            string
	CLASS_NAME    string
	TAG_NAME      string
	NAME          string
	LINK_TEXT     string
}

// By 提供与 Python 版一致的点号访问风格。
var By = ByValues{
	CSS:           "css",
	XPATH:         "xpath",
	TEXT:          "text",
	INNER_TEXT:    "innerText",
	ACCESSIBILITY: "accessibility",
	ID:            "id",
	CLASS_NAME:    "class name",
	TAG_NAME:      "tag name",
	NAME:          "name",
	LINK_TEXT:     "link text",
}
