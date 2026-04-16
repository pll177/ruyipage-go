package bidi

// LogEntryData 表示 log.entryAdded 事件的单条日志数据。
//
// 该模块仅承接事件数据结构，不包含额外命令封装。
type LogEntryData struct {
	Raw        map[string]any
	Level      string
	Text       string
	Timestamp  any
	Source     map[string]any
	LogType    string
	Method     string
	Args       []any
	StackTrace any
}

// ParseLogEntryData 从 log.entryAdded 事件参数解析日志对象。
func ParseLogEntryData(params map[string]any) LogEntryData {
	raw := cloneAnyMapDeep(params)

	var source map[string]any
	if sourceValue, ok := raw["source"].(map[string]any); ok {
		source = cloneAnyMapDeep(sourceValue)
	}

	var args []any
	if argsValue, ok := cloneAnyValueDeep(raw["args"]).([]any); ok {
		args = argsValue
	}

	return LogEntryData{
		Raw:        raw,
		Level:      readString(raw, "level"),
		Text:       readString(raw, "text"),
		Timestamp:  cloneAnyValueDeep(raw["timestamp"]),
		Source:     source,
		LogType:    readString(raw, "type"),
		Method:     readString(raw, "method"),
		Args:       args,
		StackTrace: cloneAnyValueDeep(raw["stackTrace"]),
	}
}
