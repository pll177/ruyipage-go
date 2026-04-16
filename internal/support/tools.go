package support

import (
	stderrors "errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const maxPortNumber = 65535

var (
	// ErrNilWaitCondition 表示等待条件函数为空。
	ErrNilWaitCondition = stderrors.New("等待条件函数不能为空")
	// ErrInvalidWaitTimeout 表示等待超时时间非法。
	ErrInvalidWaitTimeout = stderrors.New("等待超时时间必须大于等于 0")
	// ErrInvalidWaitInterval 表示等待轮询间隔非法。
	ErrInvalidWaitInterval = stderrors.New("等待轮询间隔必须大于 0")
	// ErrInvalidPortRange 表示端口范围非法。
	ErrInvalidPortRange = stderrors.New("端口范围非法")
	// ErrNoFreePort 表示端口范围内没有可用端口。
	ErrNoFreePort = stderrors.New("端口范围内没有可用端口")
)

// WaitUntil 轮询 condition，直到命中、返回错误或超时。
//
// 返回值遵循“三态返回”：
//   - 命中时：返回 value, true, nil
//   - 条件报错：返回零值, false, err
//   - 超时时：返回零值, false, nil
func WaitUntil[T any](condition func() (T, bool, error), timeout, interval time.Duration) (T, bool, error) {
	var zero T

	if condition == nil {
		return zero, false, ErrNilWaitCondition
	}
	if timeout < 0 {
		return zero, false, fmt.Errorf("%w: %s", ErrInvalidWaitTimeout, timeout)
	}

	deadline := time.Now().Add(timeout)

	for {
		value, matched, err := condition()
		if err != nil {
			return zero, false, err
		}
		if matched {
			return value, true, nil
		}

		remaining := time.Until(deadline)
		if remaining <= 0 {
			return zero, false, nil
		}
		if interval <= 0 {
			return zero, false, fmt.Errorf("%w: %s", ErrInvalidWaitInterval, interval)
		}

		waitFor := minDuration(interval, remaining)
		timer := time.NewTimer(waitFor)
		<-timer.C
	}
}

// IsPortOpen 检查目标主机端口是否可连接。
func IsPortOpen(host string, port int, timeout time.Duration) bool {
	if host == "" || port < 1 || port > maxPortNumber {
		return false
	}
	if timeout <= 0 {
		timeout = time.Second
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// FindFreePort 在 [start, end) 范围内查找本机可用端口。
func FindFreePort(start, end int) (int, error) {
	if err := validatePortRange(start, end); err != nil {
		return 0, err
	}

	for port := start; port < end; port++ {
		listener, err := net.Listen("tcp", net.JoinHostPort(DefaultHost, strconv.Itoa(port)))
		if err != nil {
			continue
		}
		_ = listener.Close()
		return port, nil
	}

	return 0, fmt.Errorf("%w: [%d, %d)", ErrNoFreePort, start, end)
}

// CleanText 压缩连续空白并清理首尾空白。
func CleanText(text string) string {
	if text == "" {
		return ""
	}
	return strings.Join(strings.Fields(text), " ")
}

// MakeValidFilename 生成去除 Windows 非法字符后的文件名。
func MakeValidFilename(name string, maxLength int) string {
	if name == "" || maxLength <= 0 {
		return ""
	}

	sanitized := strings.Map(func(r rune) rune {
		switch r {
		case '\\', '/', ':', '*', '?', '"', '<', '>', '|':
			return -1
		default:
			return r
		}
	}, name)

	runes := []rune(sanitized)
	if len(runes) > maxLength {
		return string(runes[:maxLength])
	}
	return sanitized
}

func validatePortRange(start, end int) error {
	switch {
	case start < 1:
		return fmt.Errorf("%w: [%d, %d)", ErrInvalidPortRange, start, end)
	case end < 1:
		return fmt.Errorf("%w: [%d, %d)", ErrInvalidPortRange, start, end)
	case start >= end:
		return fmt.Errorf("%w: [%d, %d)", ErrInvalidPortRange, start, end)
	case start > maxPortNumber:
		return fmt.Errorf("%w: [%d, %d)", ErrInvalidPortRange, start, end)
	case end > maxPortNumber+1:
		return fmt.Errorf("%w: [%d, %d)", ErrInvalidPortRange, start, end)
	default:
		return nil
	}
}

func minDuration(left, right time.Duration) time.Duration {
	if left < right {
		return left
	}
	return right
}
