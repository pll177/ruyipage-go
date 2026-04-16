package support

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var userPrefPattern = regexp.MustCompile(`user_pref\s*\(\s*["'](.+?)["']\s*,\s*(.+?)\s*\)`)

// ParsePrefValue 将 user.js / prefs.js 中的值字面量解析为 Go 值。
func ParsePrefValue(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "true" {
		return true
	}
	if trimmed == "false" {
		return false
	}
	if len(trimmed) >= 2 {
		if (trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"') ||
			(trimmed[0] == '\'' && trimmed[len(trimmed)-1] == '\'') {
			return trimmed[1 : len(trimmed)-1]
		}
	}

	if parsed, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return parsed
	}
	if parsed, err := strconv.ParseFloat(trimmed, 64); err == nil {
		return parsed
	}
	return trimmed
}

// FormatPrefValue 将 Go 值格式化为 user.js / prefs.js 可写入的字面量。
func FormatPrefValue(value any) string {
	switch typed := value.(type) {
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(typed)
	case int8:
		return strconv.FormatInt(int64(typed), 10)
	case int16:
		return strconv.FormatInt(int64(typed), 10)
	case int32:
		return strconv.FormatInt(int64(typed), 10)
	case int64:
		return strconv.FormatInt(typed, 10)
	case uint:
		return strconv.FormatUint(uint64(typed), 10)
	case uint8:
		return strconv.FormatUint(uint64(typed), 10)
	case uint16:
		return strconv.FormatUint(uint64(typed), 10)
	case uint32:
		return strconv.FormatUint(uint64(typed), 10)
	case uint64:
		return strconv.FormatUint(typed, 10)
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case string:
		return QuotePrefString(typed)
	default:
		return QuotePrefString(fmt.Sprintf("%v", typed))
	}
}

// QuotePrefString 对字符串做 user.js 兼容转义。
func QuotePrefString(value string) string {
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}

// BuildUserPrefLine 构建单行 user_pref 声明。
func BuildUserPrefLine(key string, value any) string {
	return fmt.Sprintf(`user_pref("%s", %s);`, key, FormatPrefValue(value))
}

// ParseUserPrefsContent 解析整段 user.js / prefs.js 文本。
func ParseUserPrefsContent(content string) map[string]any {
	result := map[string]any{}
	for _, match := range userPrefPattern.FindAllStringSubmatch(content, -1) {
		if len(match) != 3 {
			continue
		}
		result[match[1]] = ParsePrefValue(match[2])
	}
	return result
}

// JSPrefsFile 封装 user.js / prefs.js 的读写。
type JSPrefsFile struct {
	path string
}

// NewJSPrefsFile 创建一个 JS prefs 文件封装。
func NewJSPrefsFile(path string) *JSPrefsFile {
	return &JSPrefsFile{path: path}
}

// Path 返回文件路径。
func (f *JSPrefsFile) Path() string {
	if f == nil {
		return ""
	}
	return f.path
}

// ReadAll 读取全部 pref。
func (f *JSPrefsFile) ReadAll() (map[string]any, error) {
	if f == nil || f.path == "" {
		return map[string]any{}, nil
	}

	content, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	return ParseUserPrefsContent(string(content)), nil
}

// Read 读取单个 pref。
func (f *JSPrefsFile) Read(key string) (any, bool, error) {
	if f == nil || f.path == "" {
		return nil, false, nil
	}

	content, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	pattern := regexp.MustCompile(`user_pref\s*\(\s*["']` + regexp.QuoteMeta(key) + `["']\s*,\s*(.+?)\s*\)`)
	match := pattern.FindStringSubmatch(string(content))
	if len(match) != 2 {
		return nil, false, nil
	}
	return ParsePrefValue(match[1]), true, nil
}

// ReadPrefix 读取指定前缀的全部 pref。
func (f *JSPrefsFile) ReadPrefix(prefix string) (map[string]any, error) {
	all, err := f.ReadAll()
	if err != nil {
		return nil, err
	}

	result := map[string]any{}
	for key, value := range all {
		if strings.HasPrefix(key, prefix) {
			result[key] = value
		}
	}
	return result, nil
}

// Write 写入单个 pref，已存在则覆盖。
func (f *JSPrefsFile) Write(key string, value any) error {
	if f == nil || f.path == "" {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
		return err
	}

	content := ""
	if raw, err := os.ReadFile(f.path); err == nil {
		content = string(raw)
	} else if !os.IsNotExist(err) {
		return err
	}

	line := BuildUserPrefLine(key, value)
	pattern := regexp.MustCompile(`user_pref\s*\(\s*["']` + regexp.QuoteMeta(key) + `["'].*?\);`)
	if pattern.MatchString(content) {
		content = pattern.ReplaceAllString(content, line)
	} else if strings.TrimSpace(content) == "" {
		content = line + "\n"
	} else {
		content = strings.TrimRight(content, "\r\n") + "\n" + line + "\n"
	}

	return os.WriteFile(f.path, []byte(content), 0o644)
}

// WriteMany 逐项写入多个 pref。
func (f *JSPrefsFile) WriteMany(prefs map[string]any) error {
	if len(prefs) == 0 {
		return nil
	}

	keys := make([]string, 0, len(prefs))
	for key := range prefs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if err := f.Write(key, prefs[key]); err != nil {
			return err
		}
	}
	return nil
}

// RewriteAll 用给定 pref 集合整体覆盖文件内容。
func (f *JSPrefsFile) RewriteAll(prefs map[string]any) error {
	if f == nil || f.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
		return err
	}

	keys := make([]string, 0, len(prefs))
	for key := range prefs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, BuildUserPrefLine(key, prefs[key]))
	}

	content := ""
	if len(lines) > 0 {
		content = strings.Join(lines, "\n") + "\n"
	}
	return os.WriteFile(f.path, []byte(content), 0o644)
}

// Remove 删除单个 pref。
func (f *JSPrefsFile) Remove(key string) error {
	if f == nil || f.path == "" {
		return nil
	}

	content, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	pattern := regexp.MustCompile(`(?m)^\s*user_pref\s*\(\s*["']` + regexp.QuoteMeta(key) + `["'].*?\);\s*\r?\n?`)
	updated := pattern.ReplaceAll(content, nil)
	return os.WriteFile(f.path, updated, 0o644)
}

// PoliciesFile 封装 distribution/policies.json 的读写。
type PoliciesFile struct {
	path string
}

// NewPoliciesFile 创建一个 policies.json 封装。
func NewPoliciesFile(path string) *PoliciesFile {
	return &PoliciesFile{path: path}
}

// NewPoliciesFileFromProfile 通过 profile 路径定位 policies.json。
func NewPoliciesFileFromProfile(profilePath string) *PoliciesFile {
	return &PoliciesFile{
		path: filepath.Join(filepath.Dir(profilePath), "distribution", "policies.json"),
	}
}

// Path 返回 policies.json 路径。
func (f *PoliciesFile) Path() string {
	if f == nil {
		return ""
	}
	return f.path
}

// Read 读取 policies.json。
func (f *PoliciesFile) Read() (map[string]any, error) {
	if f == nil || f.path == "" {
		return map[string]any{}, nil
	}

	content, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, err
	}

	result := map[string]any{}
	if len(bytes.TrimSpace(content)) == 0 {
		return result, nil
	}
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// WriteMerge 将新策略深度合并进现有 policies.json。
func (f *PoliciesFile) WriteMerge(policies map[string]any) error {
	if f == nil || f.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
		return err
	}

	existing, err := f.Read()
	if err != nil {
		return err
	}
	DeepMergeMaps(existing, CloneAnyMap(policies))
	return writeIndentedJSONFile(f.path, existing)
}

// SetLockedPref 通过 Preferences 策略锁定 pref。
func (f *PoliciesFile) SetLockedPref(key string, value any) error {
	return f.WriteMerge(map[string]any{
		"policies": map[string]any{
			"Preferences": map[string]any{
				key: map[string]any{
					"Value":  value,
					"Status": "locked",
				},
			},
		},
	})
}

// UnlockPref 从 Preferences 策略中移除 pref。
func (f *PoliciesFile) UnlockPref(key string) error {
	if f == nil || f.path == "" {
		return nil
	}

	data, err := f.Read()
	if err != nil {
		return err
	}

	policies, ok := data["policies"].(map[string]any)
	if !ok || policies == nil {
		return writeIndentedJSONFile(f.path, data)
	}
	preferences, ok := policies["Preferences"].(map[string]any)
	if ok && preferences != nil {
		delete(preferences, key)
	}
	return writeIndentedJSONFile(f.path, data)
}

// DeepMergeMaps 将 override 深度合并到 base。
func DeepMergeMaps(base map[string]any, override map[string]any) {
	if base == nil || override == nil {
		return
	}

	for key, value := range override {
		overrideMap, overrideIsMap := value.(map[string]any)
		baseMap, baseIsMap := base[key].(map[string]any)
		if overrideIsMap && baseIsMap {
			DeepMergeMaps(baseMap, overrideMap)
			continue
		}
		base[key] = cloneAnyValue(value)
	}
}

// CloneAnyMap 递归复制 map[string]any。
func CloneAnyMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}

	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = cloneAnyValue(value)
	}
	return result
}

func cloneAnyValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return CloneAnyMap(typed)
	case []any:
		result := make([]any, len(typed))
		for index, item := range typed {
			result[index] = cloneAnyValue(item)
		}
		return result
	default:
		return typed
	}
}

func writeIndentedJSONFile(path string, data map[string]any) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
