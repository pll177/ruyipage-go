package support

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"sort"
	"strconv"
)

const maxSafeBiDINumber = uint64(9007199254740991)

// BiDiSharedReferenceProvider 表示可直接提供 shared reference 的对象。
type BiDiSharedReferenceProvider interface {
	BiDiSharedReference() map[string]any
}

// BiDiSharedIDProvider 表示可提供 sharedId 的对象。
type BiDiSharedIDProvider interface {
	BiDiSharedID() string
}

// BiDiHandleProvider 表示可提供 handle 的对象。
type BiDiHandleProvider interface {
	BiDiHandle() string
}

// SerializeBiDiValue 将 Go 值转换为 BiDi LocalValue 结构。
func SerializeBiDiValue(value any) map[string]any {
	if value == nil {
		return map[string]any{"type": "null"}
	}

	if provider, ok := value.(BiDiSharedReferenceProvider); ok {
		return normalizeSharedReference(provider.BiDiSharedReference())
	}

	if provider, ok := value.(BiDiSharedIDProvider); ok {
		handle := ""
		if handleProvider, ok := value.(BiDiHandleProvider); ok {
			handle = handleProvider.BiDiHandle()
		}
		return MakeSharedRef(provider.BiDiSharedID(), handle)
	}

	switch typed := value.(type) {
	case bool:
		return map[string]any{"type": "boolean", "value": typed}
	case string:
		return map[string]any{"type": "string", "value": typed}
	case int:
		return serializeSignedInteger(int64(typed), typed)
	case int8:
		return serializeSignedInteger(int64(typed), typed)
	case int16:
		return serializeSignedInteger(int64(typed), typed)
	case int32:
		return serializeSignedInteger(int64(typed), typed)
	case int64:
		return serializeSignedInteger(typed, typed)
	case uint:
		return serializeUnsignedInteger(uint64(typed), typed)
	case uint8:
		return serializeUnsignedInteger(uint64(typed), typed)
	case uint16:
		return serializeUnsignedInteger(uint64(typed), typed)
	case uint32:
		return serializeUnsignedInteger(uint64(typed), typed)
	case uint64:
		return serializeUnsignedInteger(typed, typed)
	case float32:
		return serializeFloat(float64(typed))
	case float64:
		return serializeFloat(typed)
	case []any:
		return serializeSequence(typed)
	case []string:
		items := make([]any, len(typed))
		for index, item := range typed {
			items[index] = item
		}
		return serializeSequence(items)
	case []map[string]any:
		items := make([]any, len(typed))
		for index, item := range typed {
			items[index] = item
		}
		return serializeSequence(items)
	case map[string]any:
		if hasSharedID(typed) {
			return normalizeSharedReference(typed)
		}
		return serializeObjectMap(typed)
	case map[string]string:
		if sharedID, ok := typed["sharedId"]; ok {
			return MakeSharedRef(sharedID, typed["handle"])
		}
		items := make(map[string]any, len(typed))
		for key, item := range typed {
			items[key] = item
		}
		return serializeObjectMap(items)
	case map[string]struct{}:
		keys := make([]any, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		return serializeSet(keys)
	}

	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Slice, reflect.Array:
		return serializeReflectSequence(reflected)
	case reflect.Map:
		return serializeReflectMap(reflected)
	default:
		return map[string]any{"type": "string", "value": fmt.Sprint(value)}
	}
}

// ParseBiDiValue 将 BiDi RemoteValue 转换为 Go 值。
func ParseBiDiValue(node any) any {
	values, ok := node.(map[string]any)
	if !ok {
		return node
	}

	typeValue, _ := values["type"].(string)
	switch typeValue {
	case "null", "undefined":
		return nil
	case "string":
		if value, ok := values["value"].(string); ok {
			return value
		}
		return ""
	case "number":
		return parseBiDiNumber(values["value"])
	case "boolean":
		if value, ok := values["value"].(bool); ok {
			return value
		}
		return false
	case "bigint":
		return parseBiDiBigInt(values["value"])
	case "array":
		return parseBiDiArray(values["value"])
	case "object":
		return parseBiDiObject(values["value"])
	case "map":
		return parseBiDiMap(values["value"])
	case "set":
		return parseBiDiSet(values["value"])
	case "date":
		if value, ok := values["value"].(string); ok {
			return value
		}
		return ""
	case "regexp":
		return cloneAnyValueDeep(values["value"])
	case "node", "window", "error":
		return cloneAnyMapDeep(values)
	default:
		if value, ok := values["value"]; ok {
			return cloneAnyValueDeep(value)
		}
		return cloneAnyMapDeep(values)
	}
}

// MakeSharedRef 创建 BiDi SharedReference。
func MakeSharedRef(sharedID string, handle string) map[string]any {
	ref := map[string]any{
		"type":     "sharedReference",
		"sharedId": sharedID,
	}
	if handle != "" {
		ref["handle"] = handle
	}
	return ref
}

func serializeSignedInteger(value int64, original any) map[string]any {
	if value < -int64(maxSafeBiDINumber) || value > int64(maxSafeBiDINumber) {
		return map[string]any{"type": "bigint", "value": fmt.Sprintf("%d", value)}
	}
	return map[string]any{"type": "number", "value": original}
}

func serializeUnsignedInteger(value uint64, original any) map[string]any {
	if value > maxSafeBiDINumber {
		return map[string]any{"type": "bigint", "value": fmt.Sprintf("%d", value)}
	}
	return map[string]any{"type": "number", "value": original}
}

func serializeFloat(value float64) map[string]any {
	switch {
	case math.IsNaN(value):
		return map[string]any{"type": "number", "value": "NaN"}
	case math.IsInf(value, 1):
		return map[string]any{"type": "number", "value": "Infinity"}
	case math.IsInf(value, -1):
		return map[string]any{"type": "number", "value": "-Infinity"}
	case value == 0 && math.Signbit(value):
		return map[string]any{"type": "number", "value": "-0"}
	default:
		return map[string]any{"type": "number", "value": value}
	}
}

func serializeSequence(items []any) map[string]any {
	serialized := make([]any, len(items))
	for index, item := range items {
		serialized[index] = SerializeBiDiValue(item)
	}
	return map[string]any{"type": "array", "value": serialized}
}

func serializeReflectSequence(value reflect.Value) map[string]any {
	items := make([]any, value.Len())
	for index := 0; index < value.Len(); index++ {
		items[index] = value.Index(index).Interface()
	}
	return serializeSequence(items)
}

func serializeObjectMap(items map[string]any) map[string]any {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	pairs := make([]any, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, []any{key, SerializeBiDiValue(items[key])})
	}
	return map[string]any{"type": "object", "value": pairs}
}

func serializeReflectMap(value reflect.Value) map[string]any {
	if value.Type().Elem().Kind() == reflect.Struct && value.Type().Elem().NumField() == 0 {
		items := make([]any, 0, value.Len())
		for _, key := range value.MapKeys() {
			items = append(items, key.Interface())
		}
		return serializeSet(items)
	}

	if sharedRef, ok := serializeReflectSharedReference(value); ok {
		return sharedRef
	}

	keys := value.MapKeys()
	sort.Slice(keys, func(i int, j int) bool {
		return fmt.Sprint(keys[i].Interface()) < fmt.Sprint(keys[j].Interface())
	})

	pairs := make([]any, 0, len(keys))
	for _, key := range keys {
		pairKey := any(key.Interface())
		if key.Kind() == reflect.String {
			pairKey = key.String()
		} else {
			pairKey = SerializeBiDiValue(key.Interface())
		}
		pairs = append(pairs, []any{pairKey, SerializeBiDiValue(value.MapIndex(key).Interface())})
	}
	return map[string]any{"type": "object", "value": pairs}
}

func serializeSet(items []any) map[string]any {
	sort.Slice(items, func(i int, j int) bool {
		return fmt.Sprint(items[i]) < fmt.Sprint(items[j])
	})

	serialized := make([]any, len(items))
	for index, item := range items {
		serialized[index] = SerializeBiDiValue(item)
	}
	return map[string]any{"type": "set", "value": serialized}
}

func serializeReflectSharedReference(value reflect.Value) (map[string]any, bool) {
	if value.Type().Key().Kind() != reflect.String {
		return nil, false
	}

	sharedValue := value.MapIndex(reflect.ValueOf("sharedId"))
	if !sharedValue.IsValid() {
		return nil, false
	}

	handleValue := value.MapIndex(reflect.ValueOf("handle"))
	return MakeSharedRef(fmt.Sprint(sharedValue.Interface()), reflectValueString(handleValue)), true
}

func parseBiDiNumber(value any) any {
	text, ok := value.(string)
	if ok {
		switch text {
		case "NaN":
			return math.NaN()
		case "Infinity":
			return math.Inf(1)
		case "-Infinity":
			return math.Inf(-1)
		case "-0":
			return math.Copysign(0, -1)
		}
	}
	return value
}

func parseBiDiBigInt(value any) any {
	text, ok := value.(string)
	if !ok {
		return int64(0)
	}

	if parsed, err := strconv.ParseInt(text, 10, 64); err == nil {
		return parsed
	}

	result := new(big.Int)
	if _, ok := result.SetString(text, 10); ok {
		return result
	}
	return text
}

func parseBiDiArray(value any) []any {
	rawItems, ok := value.([]any)
	if !ok {
		return []any{}
	}

	items := make([]any, len(rawItems))
	for index, item := range rawItems {
		items[index] = ParseBiDiValue(item)
	}
	return items
}

func parseBiDiObject(value any) any {
	pairs := parseBiDiPairs(value)
	stringResult := make(map[string]any, len(pairs))
	genericResult := make(map[any]any, len(pairs))
	allStringKeys := true

	for _, pair := range pairs {
		key := pair.key
		if keyText, ok := key.(string); ok {
			stringResult[keyText] = pair.value
			genericResult[keyText] = pair.value
			continue
		}
		allStringKeys = false
		genericResult[makeComparableKey(key)] = pair.value
	}

	if allStringKeys {
		return stringResult
	}
	return genericResult
}

func parseBiDiMap(value any) map[any]any {
	pairs := parseBiDiPairs(value)
	result := make(map[any]any, len(pairs))
	for _, pair := range pairs {
		result[makeComparableKey(pair.key)] = pair.value
	}
	return result
}

func parseBiDiSet(value any) any {
	rawItems, ok := value.([]any)
	if !ok {
		return map[any]struct{}{}
	}

	result := make(map[any]struct{}, len(rawItems))
	for _, item := range rawItems {
		result[makeComparableKey(ParseBiDiValue(item))] = struct{}{}
	}
	return result
}

type bidiPair struct {
	key   any
	value any
}

func parseBiDiPairs(value any) []bidiPair {
	rawPairs, ok := value.([]any)
	if !ok {
		return []bidiPair{}
	}

	result := make([]bidiPair, 0, len(rawPairs))
	for _, rawPair := range rawPairs {
		pairItems, ok := rawPair.([]any)
		if !ok || len(pairItems) != 2 {
			continue
		}

		key := pairItems[0]
		if _, ok := key.(string); !ok {
			key = ParseBiDiValue(key)
		}
		result = append(result, bidiPair{
			key:   key,
			value: ParseBiDiValue(pairItems[1]),
		})
	}
	return result
}

func makeComparableKey(value any) any {
	if value == nil {
		return nil
	}

	valueType := reflect.TypeOf(value)
	if valueType != nil && valueType.Comparable() {
		return value
	}
	return fmt.Sprintf("%#v", value)
}

func hasSharedID(value map[string]any) bool {
	_, ok := value["sharedId"]
	return ok
}

func normalizeSharedReference(value map[string]any) map[string]any {
	if value == nil {
		return MakeSharedRef("", "")
	}

	sharedID := ""
	if rawSharedID, ok := value["sharedId"]; ok {
		sharedID = fmt.Sprint(rawSharedID)
	}

	handle := ""
	if rawHandle, ok := value["handle"]; ok {
		handle = fmt.Sprint(rawHandle)
	}
	return MakeSharedRef(sharedID, handle)
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

func reflectValueString(value reflect.Value) string {
	if !value.IsValid() {
		return ""
	}
	if text, ok := value.Interface().(string); ok {
		return text
	}
	return fmt.Sprint(value.Interface())
}
