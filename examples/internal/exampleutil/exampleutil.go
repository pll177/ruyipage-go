package exampleutil

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	ruyipage "github.com/pll177/ruyipage-go"
)

const (
	// EnvOutputRoot 允许测试把示例输出重定向到临时目录。
	EnvOutputRoot = "RUYIPAGE_EXAMPLE_OUTPUT_ROOT"
	// EnvServerPort 允许测试覆盖本地示例 HTTP 服务端口。
	EnvServerPort = "RUYIPAGE_EXAMPLE_SERVER_PORT"
)

// RunMain 统一处理示例主入口的错误输出与退出码。
func RunMain(run func() error) {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "\n✗ 测试失败: %v\n", err)
		os.Exit(1)
	}
}

// RepoRoot 返回 ruyipage-go 模块根目录。
func RepoRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("无法定位示例辅助文件路径")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..")), nil
}

// VisibleOptions 返回与 Python 示例对齐的可见浏览器配置。
func VisibleOptions() *ruyipage.FirefoxOptions {
	return ruyipage.NewFirefoxOptions().Headless(false)
}

// TestPagePath 返回 examples 基础测试页的绝对路径。
func TestPagePath(name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		name = "test_page.html"
	}
	root, err := RepoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "testdata", "examples", "test_pages", name), nil
}

// TestPageURL 返回 examples 基础测试页的 file:/// URL。
func TestPageURL(name string) (string, error) {
	path, err := TestPagePath(name)
	if err != nil {
		return "", err
	}
	return FileURLFromPath(path)
}

// FileURLFromPath 将本地文件路径转换为 file:/// URL。
func FileURLFromPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return (&url.URL{
		Scheme: "file",
		Path:   "/" + filepath.ToSlash(absPath),
	}).String(), nil
}

// OutputDir 返回示例输出目录；测试可通过环境变量重定向到临时目录。
func OutputDir(exampleName string) (string, error) {
	base := strings.TrimSpace(os.Getenv(EnvOutputRoot))
	if base == "" {
		root, err := RepoRoot()
		if err != nil {
			return "", err
		}
		base = filepath.Join(root, "examples", exampleName, "output")
	} else {
		base = filepath.Join(base, exampleName)
	}
	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", err
	}
	return base, nil
}

// ServerPort 返回示例 HTTP 服务端口；默认保留 Python 示例的 8888。
func ServerPort(defaultPort int) int {
	raw := strings.TrimSpace(os.Getenv(EnvServerPort))
	if raw == "" {
		return defaultPort
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 || value > 65535 {
		return defaultPort
	}
	return value
}
