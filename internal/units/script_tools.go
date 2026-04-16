package units

import "github.com/pll177/ruyipage-go/internal/bidi"

// RealmInfo 表示单个 realm 信息。
type RealmInfo struct {
	Raw     map[string]any
	Realm   string
	Type    string
	Context string
	Origin  string
}

// ScriptRemoteValue 表示脚本执行返回的远端值。
type ScriptRemoteValue struct {
	Raw      map[string]any
	Type     string
	Handle   string
	SharedID string
	Value    any
}

// ScriptResult 表示脚本执行结果。
type ScriptResult struct {
	Raw              map[string]any
	Type             string
	Result           ScriptRemoteValue
	ExceptionDetails any
}

// Success 返回脚本是否执行成功。
func (r ScriptResult) Success() bool {
	return r.Type == "success"
}

// PreloadScript 表示预加载脚本标识。
type PreloadScript struct {
	ID string
}

// NewRealmInfo 从原始 BiDi realm 结果创建公开类型。
func NewRealmInfo(data map[string]any) RealmInfo {
	raw := cloneAnyMapDeep(data)
	return RealmInfo{
		Raw:     raw,
		Realm:   readString(raw, "realm"),
		Type:    readString(raw, "type"),
		Context: readString(raw, "context"),
		Origin:  readString(raw, "origin"),
	}
}

// NewRealmInfoFromData 从内部 bidi 结构创建公开类型。
func NewRealmInfoFromData(data bidi.RealmInfoData) RealmInfo {
	return RealmInfo{
		Raw:     cloneAnyMapDeep(data.Raw),
		Realm:   data.Realm,
		Type:    data.Type,
		Context: data.Context,
		Origin:  data.Origin,
	}
}

// NewRealmInfosFromData 批量转换内部 realm 数据。
func NewRealmInfosFromData(data []bidi.RealmInfoData) []RealmInfo {
	if data == nil {
		return nil
	}

	result := make([]RealmInfo, len(data))
	for index, item := range data {
		result[index] = NewRealmInfoFromData(item)
	}
	return result
}

// NewScriptRemoteValue 从原始 BiDi RemoteValue 结果创建公开类型。
func NewScriptRemoteValue(data map[string]any) ScriptRemoteValue {
	raw := cloneAnyMapDeep(data)
	return ScriptRemoteValue{
		Raw:      raw,
		Type:     readString(raw, "type"),
		Handle:   readString(raw, "handle"),
		SharedID: readString(raw, "sharedId"),
		Value:    cloneAnyValueDeep(raw["value"]),
	}
}

// NewScriptRemoteValueFromData 从内部 bidi 结构创建公开类型。
func NewScriptRemoteValueFromData(data bidi.ScriptRemoteValueData) ScriptRemoteValue {
	return ScriptRemoteValue{
		Raw:      cloneAnyMapDeep(data.Raw),
		Type:     data.Type,
		Handle:   data.Handle,
		SharedID: data.SharedID,
		Value:    cloneAnyValueDeep(data.Value),
	}
}

// NewScriptResult 从原始脚本结果创建公开类型。
func NewScriptResult(data map[string]any) ScriptResult {
	raw := cloneAnyMapDeep(data)
	remoteValue, _ := raw["result"].(map[string]any)
	return ScriptResult{
		Raw:              raw,
		Type:             readString(raw, "type"),
		Result:           NewScriptRemoteValue(remoteValue),
		ExceptionDetails: cloneAnyValueDeep(raw["exceptionDetails"]),
	}
}

// NewScriptResultFromData 从内部 bidi 结构创建公开类型。
func NewScriptResultFromData(data bidi.ScriptResultData) ScriptResult {
	return ScriptResult{
		Raw:              cloneAnyMapDeep(data.Raw),
		Type:             data.Type,
		Result:           NewScriptRemoteValueFromData(data.Result),
		ExceptionDetails: cloneAnyValueDeep(data.ExceptionDetails),
	}
}

// NewPreloadScript 创建公开的预加载脚本结果。
func NewPreloadScript(scriptID string) PreloadScript {
	return PreloadScript{ID: scriptID}
}

// NewPreloadScriptFromData 从内部 bidi 结构创建公开类型。
func NewPreloadScriptFromData(data bidi.PreloadScriptData) PreloadScript {
	return PreloadScript{ID: data.Script}
}

func cloneAnyMapDeep(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}

	dst := make(map[string]any, len(src))
	for key, value := range src {
		dst[key] = cloneAnyValueDeep(value)
	}
	return dst
}

func cloneAnyValueDeep(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneAnyMapDeep(typed)
	case []any:
		dst := make([]any, len(typed))
		for index, item := range typed {
			dst[index] = cloneAnyValueDeep(item)
		}
		return dst
	case []string:
		return append([]string(nil), typed...)
	case []map[string]any:
		dst := make([]map[string]any, len(typed))
		for index, item := range typed {
			dst[index] = cloneAnyMapDeep(item)
		}
		return dst
	default:
		return value
	}
}

func readString(values map[string]any, key string) string {
	if values == nil {
		return ""
	}

	value, _ := values[key].(string)
	return value
}
