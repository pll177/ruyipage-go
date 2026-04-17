package exampleutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	ruyipage "github.com/pll177/ruyipage-go"
)

const (
	EnvFirefoxPath     = "RUYIPAGE_EXAMPLE_FIREFOX_PATH"
	EnvAttachCommand   = "RUYIPAGE_EXAMPLE_ATTACH_COMMAND"
	EnvProxyHost       = "RUYIPAGE_EXAMPLE_PROXY_HOST"
	EnvProxyPort       = "RUYIPAGE_EXAMPLE_PROXY_PORT"
	EnvProxyUsername   = "RUYIPAGE_EXAMPLE_PROXY_USERNAME"
	EnvProxyPassword   = "RUYIPAGE_EXAMPLE_PROXY_PASSWORD"
	DefaultFirefoxPath = `C:\Program Files\Mozilla Firefox\firefox.exe`
	defaultProxyHost   = "proxy.example.com"
	defaultProxyPort   = 8080
)

// CheckRow 表示结果表中的一行。
type CheckRow struct {
	Item   string
	Status string
	Note   string
}

// FixedVisibleOptions 返回显式绑定固定 Firefox 路径的可见浏览器配置。
func FixedVisibleOptions() *ruyipage.FirefoxOptions {
	return VisibleOptions().WithBrowserPath(FirefoxPath())
}

// FirefoxPath 返回示例当前使用的 Firefox 路径。
func FirefoxPath() string {
	return ResolveEnvPath(EnvFirefoxPath, DefaultFirefoxPath)
}

// AttachCommand 返回接管示例建议手工执行的启动命令。
func AttachCommand() string {
	command := strings.TrimSpace(os.Getenv(EnvAttachCommand))
	if command != "" {
		return command
	}
	return fmt.Sprintf(`"%s" -remote-debugging-port 9222`, FirefoxPath())
}

// AddCheck 向结果表追加一行。
func AddCheck(rows *[]CheckRow, item string, status string, note string) {
	if rows == nil {
		return
	}
	*rows = append(*rows, CheckRow{
		Item:   item,
		Status: status,
		Note:   note,
	})
}

// PrintChecks 以 Markdown 表格输出检查结果。
func PrintChecks(rows []CheckRow) {
	fmt.Println("\n| 项目 | 状态 | 说明 |")
	fmt.Println("| --- | --- | --- |")
	for _, row := range rows {
		fmt.Printf("| %s | %s | %s |\n", row.Item, row.Status, row.Note)
	}
}

// FixedProxyParts 返回示例代理配置；host/port 可回退到安全占位值，账号密码仅从环境变量读取。
func FixedProxyParts() (host string, port int, username string, password string, err error) {
	host = ResolveEnvPath(EnvProxyHost, defaultProxyHost)
	port = ResolveEnvInt(EnvProxyPort, defaultProxyPort)
	if port < 1 || port > 65535 {
		return "", 0, "", "", fmt.Errorf("代理端口无效: %d", port)
	}
	username = strings.TrimSpace(os.Getenv(EnvProxyUsername))
	password = strings.TrimSpace(os.Getenv(EnvProxyPassword))
	return host, port, username, password, nil
}

// FixedProxyURL 返回不含账号密码的 HTTP 代理地址。
func FixedProxyURL() (string, error) {
	host, port, _, _, err := FixedProxyParts()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://%s:%d", host, port), nil
}

// WriteFixedProxyFPFile 生成包含固定代理认证信息的临时 fpfile。
func WriteFixedProxyFPFile(dir string) (string, error) {
	_, _, username, password, err := FixedProxyParts()
	if err != nil {
		return "", err
	}
	if username == "" || password == "" {
		return "", fmt.Errorf(
			"请通过 %s 和 %s 提供代理认证信息",
			EnvProxyUsername,
			EnvProxyPassword,
		)
	}
	if strings.TrimSpace(dir) == "" {
		dir = os.TempDir()
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "proxy_auth_profile.txt")
	content := fmt.Sprintf("httpauth.username:%s\nhttpauth.password:%s\n", username, password)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// ResolveEnvPath 返回环境变量指定的路径，否则使用 fallback。
func ResolveEnvPath(envKey string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(envKey))
	if value != "" {
		return value
	}
	return fallback
}

// ResolveEnvInt 返回环境变量指定的整数，否则使用 fallback。
func ResolveEnvInt(envKey string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(envKey))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

// PrintManualKeepOpen 输出“保持浏览器打开供人工观察”的统一说明。
func PrintManualKeepOpen(duration time.Duration, reason string) {
	if duration > 0 {
		fmt.Printf("\n浏览器将保持打开 %.0f 秒，供人工观察。", duration.Seconds())
	} else {
		fmt.Print("\n浏览器将保持打开，供人工观察。")
	}
	if strings.TrimSpace(reason) != "" {
		fmt.Printf(" 说明: %s", reason)
	}
	fmt.Println()
}

// DecodeNetworkText 尝试把 collector/network 数据解码成 UTF-8 文本。
func DecodeNetworkText(data *ruyipage.NetworkData) string {
	if data == nil || !data.HasData() {
		return ""
	}

	for _, value := range []any{data.Bytes, data.Base64, data.Raw["body"], data.Raw["data"], data.Raw["value"]} {
		if text, ok := decodeNetworkTextValue(value); ok {
			return text
		}
	}
	if data.Raw != nil {
		return fmt.Sprint(data.Raw)
	}
	return ""
}

func decodeNetworkTextValue(value any) (string, bool) {
	switch typed := value.(type) {
	case nil:
		return "", false
	case string:
		return typed, true
	case []byte:
		return string(typed), true
	case map[string]any:
		typeName := strings.TrimSpace(fmt.Sprint(typed["type"]))
		rawValue := typed["value"]
		switch typeName {
		case "string":
			return fmt.Sprint(rawValue), true
		case "base64":
			decoded, err := decodeBase64String(fmt.Sprint(rawValue))
			if err != nil {
				return fmt.Sprint(rawValue), true
			}
			return decoded, true
		default:
			if rawValue == nil {
				return "", false
			}
			return fmt.Sprint(rawValue), true
		}
	default:
		return fmt.Sprint(typed), true
	}
}

func decodeBase64String(value string) (string, error) {
	data, err := ruyiBase64Decode(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
