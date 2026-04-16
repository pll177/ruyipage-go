package bidi

import (
	"fmt"
	"time"

	"ruyipage-go/internal/support"
)

// RealmInfoData 表示 script.getRealms 返回的单个 realm。
type RealmInfoData struct {
	Raw     map[string]any
	Realm   string
	Type    string
	Context string
	Origin  string
}

// ScriptRemoteValueData 表示脚本执行结果中的远端值。
type ScriptRemoteValueData struct {
	Raw      map[string]any
	Type     string
	Handle   string
	SharedID string
	Value    any
}

// ScriptResultData 表示 script.evaluate / script.callFunction 的返回结果。
type ScriptResultData struct {
	Raw              map[string]any
	Type             string
	Result           ScriptRemoteValueData
	ExceptionDetails any
}

// Success 返回本次脚本执行是否成功。
func (r ScriptResultData) Success() bool {
	return r.Type == "success"
}

// PreloadScriptData 表示 script.addPreloadScript 的返回结果。
type PreloadScriptData struct {
	Raw    map[string]any
	Script string
}

type scriptCommandDriver interface {
	Run(method string, params map[string]any, timeout time.Duration) (map[string]any, error)
}

// Evaluate 调用 script.evaluate 执行表达式。
func Evaluate(
	driver scriptCommandDriver,
	expression string,
	target map[string]any,
	awaitPromise *bool,
	sandbox string,
	resultOwnership string,
	serializationOptions map[string]any,
	userActivation bool,
	timeout time.Duration,
) (ScriptResultData, error) {
	params := map[string]any{
		"expression":      expression,
		"target":          resolveScriptTarget(target),
		"awaitPromise":    resolveScriptBoolDefault(awaitPromise, true),
		"resultOwnership": resolveScriptStringDefault(resultOwnership, "root"),
	}
	if len(serializationOptions) > 0 {
		params["serializationOptions"] = cloneAnyMapDeep(serializationOptions)
	}
	if sandbox != "" {
		params["sandbox"] = sandbox
	}
	if userActivation {
		params["userActivation"] = true
	}

	return runScriptResultCommand(driver, "script.evaluate", params, timeout)
}

// CallFunction 调用 script.callFunction 执行函数声明。
func CallFunction(
	driver scriptCommandDriver,
	functionDeclaration string,
	target map[string]any,
	arguments []any,
	this any,
	awaitPromise *bool,
	sandbox string,
	resultOwnership string,
	serializationOptions map[string]any,
	userActivation bool,
	timeout time.Duration,
) (ScriptResultData, error) {
	params := map[string]any{
		"functionDeclaration": functionDeclaration,
		"target":              resolveScriptTarget(target),
		"awaitPromise":        resolveScriptBoolDefault(awaitPromise, true),
		"resultOwnership":     resolveScriptStringDefault(resultOwnership, "root"),
	}
	if len(arguments) > 0 {
		serialized := make([]any, len(arguments))
		for index, argument := range arguments {
			serialized[index] = serializeScriptCallArgument(argument)
		}
		params["arguments"] = serialized
	}
	if this != nil {
		params["this"] = serializeScriptCallArgument(this)
	}
	if len(serializationOptions) > 0 {
		params["serializationOptions"] = cloneAnyMapDeep(serializationOptions)
	}
	if sandbox != "" {
		params["sandbox"] = sandbox
	}
	if userActivation {
		params["userActivation"] = true
	}

	return runScriptResultCommand(driver, "script.callFunction", params, timeout)
}

// AddPreloadScript 调用 script.addPreloadScript 注册预加载脚本。
func AddPreloadScript(
	driver scriptCommandDriver,
	functionDeclaration string,
	arguments []any,
	contexts any,
	sandbox string,
	timeout time.Duration,
) (PreloadScriptData, error) {
	params := map[string]any{
		"functionDeclaration": functionDeclaration,
	}
	if len(arguments) > 0 {
		serialized := make([]any, len(arguments))
		for index, argument := range arguments {
			serialized[index] = serializePreloadArgument(argument)
		}
		params["arguments"] = serialized
	}
	if normalizedContexts, include, err := normalizeOptionalStringList(contexts, "contexts"); err != nil {
		return PreloadScriptData{}, err
	} else if include {
		params["contexts"] = normalizedContexts
	}
	if sandbox != "" {
		params["sandbox"] = sandbox
	}

	if driver == nil {
		return PreloadScriptData{}, support.NewPageDisconnectedError("script driver 未初始化", nil)
	}

	result, err := driver.Run("script.addPreloadScript", params, timeout)
	if err != nil {
		return PreloadScriptData{}, err
	}
	return parsePreloadScriptData(result), nil
}

// RemovePreloadScript 调用 script.removePreloadScript 清理预加载脚本。
func RemovePreloadScript(driver scriptCommandDriver, scriptID string, timeout time.Duration) error {
	if driver == nil {
		return support.NewPageDisconnectedError("script driver 未初始化", nil)
	}

	_, err := driver.Run("script.removePreloadScript", map[string]any{"script": scriptID}, timeout)
	return err
}

// GetRealms 调用 script.getRealms 获取当前 realm 列表。
func GetRealms(driver scriptCommandDriver, context string, type_ string, timeout time.Duration) ([]RealmInfoData, error) {
	if driver == nil {
		return nil, support.NewPageDisconnectedError("script driver 未初始化", nil)
	}

	params := map[string]any{}
	if context != "" {
		params["context"] = context
	}
	if type_ != "" {
		params["type"] = type_
	}

	result, err := driver.Run("script.getRealms", params, timeout)
	if err != nil {
		return nil, err
	}

	return parseRealmInfoList(result["realms"]), nil
}

// Disown 调用 script.disown 释放远端对象句柄。
func Disown(driver scriptCommandDriver, handles []string, target map[string]any, timeout time.Duration) error {
	if driver == nil {
		return support.NewPageDisconnectedError("script driver 未初始化", nil)
	}

	params := map[string]any{
		"handles": cloneStringSlice(handles),
		"target":  resolveScriptTarget(target),
	}
	_, err := driver.Run("script.disown", params, timeout)
	return err
}

func runScriptResultCommand(
	driver scriptCommandDriver,
	method string,
	params map[string]any,
	timeout time.Duration,
) (ScriptResultData, error) {
	if driver == nil {
		return ScriptResultData{}, support.NewPageDisconnectedError("script driver 未初始化", nil)
	}

	result, err := driver.Run(method, params, timeout)
	if err != nil {
		return ScriptResultData{}, err
	}

	parsed := parseScriptResultData(result)
	if parsed.Type == "exception" {
		return parsed, support.NewJavaScriptError(resolveScriptExceptionMessage(parsed), parsed.ExceptionDetails, nil)
	}
	return parsed, nil
}

func resolveScriptTarget(target map[string]any) map[string]any {
	if len(target) == 0 {
		return map[string]any{}
	}
	return cloneAnyMapDeep(target)
}

func resolveScriptBoolDefault(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func resolveScriptStringDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func serializeScriptCallArgument(value any) any {
	if typed, ok := value.(map[string]any); ok && hasScriptProtocolMarkers(typed) {
		return cloneAnyMapDeep(typed)
	}
	return support.SerializeBiDiValue(value)
}

func serializePreloadArgument(value any) any {
	if typed, ok := value.(map[string]any); ok {
		return cloneAnyMapDeep(typed)
	}
	return support.SerializeBiDiValue(value)
}

func hasScriptProtocolMarkers(value map[string]any) bool {
	if value == nil {
		return false
	}
	_, hasSharedID := value["sharedId"]
	_, hasType := value["type"]
	return hasSharedID || hasType
}

func parseRealmInfoList(value any) []RealmInfoData {
	switch typed := value.(type) {
	case []map[string]any:
		result := make([]RealmInfoData, len(typed))
		for index, item := range typed {
			result[index] = parseRealmInfoData(item)
		}
		return result
	case []any:
		result := make([]RealmInfoData, 0, len(typed))
		for _, item := range typed {
			if realm, ok := item.(map[string]any); ok {
				result = append(result, parseRealmInfoData(realm))
			}
		}
		return result
	default:
		return []RealmInfoData{}
	}
}

func parseRealmInfoData(data map[string]any) RealmInfoData {
	raw := cloneAnyMapDeep(data)
	return RealmInfoData{
		Raw:     raw,
		Realm:   readString(raw, "realm"),
		Type:    readString(raw, "type"),
		Context: readString(raw, "context"),
		Origin:  readString(raw, "origin"),
	}
}

func parseScriptRemoteValueData(data map[string]any) ScriptRemoteValueData {
	raw := cloneAnyMapDeep(data)
	return ScriptRemoteValueData{
		Raw:      raw,
		Type:     readString(raw, "type"),
		Handle:   readString(raw, "handle"),
		SharedID: readString(raw, "sharedId"),
		Value:    cloneAnyValueDeep(raw["value"]),
	}
}

func parseScriptResultData(data map[string]any) ScriptResultData {
	raw := cloneAnyMapDeep(data)
	resultValue, _ := raw["result"].(map[string]any)
	return ScriptResultData{
		Raw:              raw,
		Type:             readString(raw, "type"),
		Result:           parseScriptRemoteValueData(resultValue),
		ExceptionDetails: cloneAnyValueDeep(raw["exceptionDetails"]),
	}
}

func parsePreloadScriptData(data map[string]any) PreloadScriptData {
	raw := cloneAnyMapDeep(data)
	return PreloadScriptData{
		Raw:    raw,
		Script: readString(raw, "script"),
	}
}

func resolveScriptExceptionMessage(result ScriptResultData) string {
	if details, ok := result.ExceptionDetails.(map[string]any); ok {
		if text := readString(details, "text"); text != "" {
			return text
		}
	}
	if len(result.Raw) > 0 {
		return fmt.Sprintf("%v", result.Raw)
	}
	return "JavaScript 执行失败"
}
