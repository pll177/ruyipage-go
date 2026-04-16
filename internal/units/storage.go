package units

import (
	"fmt"
	"strings"
)

type storageOwner interface {
	RunJS(script string, args ...any) (any, error)
}

// StorageManager 提供 localStorage / sessionStorage 的高层读写能力。
type StorageManager struct {
	owner       storageOwner
	storageType string
}

// NewStorageManager 创建存储管理器。
func NewStorageManager(owner storageOwner, storageType string) *StorageManager {
	return &StorageManager{
		owner:       owner,
		storageType: normalizeStorageType(storageType),
	}
}

// Set 设置单个存储项。
func (m *StorageManager) Set(key string, value any) error {
	if m == nil || m.owner == nil {
		return nil
	}
	_, err := m.owner.RunJS(fmt.Sprintf("(key, value) => %s.setItem(key, value)", m.storageType), key, fmt.Sprint(value))
	return err
}

// Get 读取单个存储项；不存在时返回空字符串。
func (m *StorageManager) Get(key string) (string, error) {
	if m == nil || m.owner == nil {
		return "", nil
	}
	value, err := m.owner.RunJS(fmt.Sprintf("(key) => %s.getItem(key)", m.storageType), key)
	if err != nil || value == nil {
		return "", err
	}
	return fmt.Sprint(value), nil
}

// Remove 删除单个存储项。
func (m *StorageManager) Remove(key string) error {
	if m == nil || m.owner == nil {
		return nil
	}
	_, err := m.owner.RunJS(fmt.Sprintf("(key) => %s.removeItem(key)", m.storageType), key)
	return err
}

// Clear 清空全部存储项。
func (m *StorageManager) Clear() error {
	if m == nil || m.owner == nil {
		return nil
	}
	_, err := m.owner.RunJS(fmt.Sprintf("%s.clear()", m.storageType))
	return err
}

// Keys 返回全部 key。
func (m *StorageManager) Keys() ([]string, error) {
	if m == nil || m.owner == nil {
		return []string{}, nil
	}
	value, err := m.owner.RunJS(fmt.Sprintf("Object.keys(%s)", m.storageType))
	if err != nil {
		return nil, err
	}
	return anyToStringSlice(value), nil
}

// Items 返回全部键值对快照。
func (m *StorageManager) Items() (map[string]string, error) {
	if m == nil || m.owner == nil {
		return map[string]string{}, nil
	}
	value, err := m.owner.RunJS(fmt.Sprintf(`(() => {
		const items = {};
		for (let i = 0; i < %s.length; i++) {
			const key = %s.key(i);
			items[key] = %s.getItem(key);
		}
		return items;
	})()`, m.storageType, m.storageType, m.storageType))
	if err != nil {
		return nil, err
	}
	return anyToStringMap(value), nil
}

// Len 返回当前存储项数量。
func (m *StorageManager) Len() (int, error) {
	if m == nil || m.owner == nil {
		return 0, nil
	}
	value, err := m.owner.RunJS(fmt.Sprintf("%s.length", m.storageType))
	if err != nil {
		return 0, err
	}
	return anyToInt(value), nil
}

// Contains 检查指定 key 是否存在。
func (m *StorageManager) Contains(key string) (bool, error) {
	if m == nil || m.owner == nil {
		return false, nil
	}
	value, err := m.owner.RunJS(fmt.Sprintf("(key) => %s.getItem(key) !== null", m.storageType), key)
	if err != nil {
		return false, err
	}
	return anyToBool(value), nil
}

func normalizeStorageType(storageType string) string {
	if strings.EqualFold(storageType, "sessionStorage") {
		return "sessionStorage"
	}
	return "localStorage"
}

func anyToStringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			result = append(result, fmt.Sprint(item))
		}
		return result
	default:
		return []string{}
	}
}

func anyToStringMap(value any) map[string]string {
	switch typed := value.(type) {
	case map[string]string:
		result := make(map[string]string, len(typed))
		for key, item := range typed {
			result[key] = item
		}
		return result
	case map[string]any:
		result := make(map[string]string, len(typed))
		for key, item := range typed {
			if item == nil {
				result[key] = ""
				continue
			}
			result[key] = fmt.Sprint(item)
		}
		return result
	default:
		return map[string]string{}
	}
}

func anyToInt(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int8:
		return int(typed)
	case int16:
		return int(typed)
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case uint:
		return int(typed)
	case uint8:
		return int(typed)
	case uint16:
		return int(typed)
	case uint32:
		return int(typed)
	case uint64:
		return int(typed)
	case float32:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
	}
}

func anyToBool(value any) bool {
	typed, _ := value.(bool)
	return typed
}
