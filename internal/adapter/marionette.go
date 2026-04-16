package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

const (
	// DefaultMarionettePort 是 Firefox 默认的 Marionette 端口。
	DefaultMarionettePort = 2828
	// DefaultMarionetteHost 是 Firefox 默认的 Marionette 主机。
	DefaultMarionetteHost      = "127.0.0.1"
	defaultMarionetteIOTimeout = 5 * time.Second
)

// MarionetteClient 是最小化的 Marionette TCP 客户端。
type MarionetteClient struct {
	host      string
	port      int
	ioTimeout time.Duration
}

// NewMarionetteClient 创建一个新的 Marionette 客户端。
func NewMarionetteClient(host string, port int) *MarionetteClient {
	if host == "" {
		host = DefaultMarionetteHost
	}
	if port <= 0 {
		port = DefaultMarionettePort
	}

	return &MarionetteClient{
		host:      host,
		port:      port,
		ioTimeout: defaultMarionetteIOTimeout,
	}
}

// IsAvailable 检测 Marionette 端口是否可用。
func (c *MarionetteClient) IsAvailable() bool {
	if c == nil {
		return false
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(c.host, strconv.Itoa(c.port)), time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// GetPref 读取单个运行时 pref。
func (c *MarionetteClient) GetPref(key string) (any, bool, error) {
	if c == nil || key == "" {
		return nil, false, nil
	}

	conn, err := c.connect()
	if err != nil {
		return nil, false, err
	}
	defer conn.Close()

	response, err := c.run(conn, []any{
		0,
		1,
		"getPrefs",
		map[string]any{"prefs": []string{key}},
	})
	if err != nil {
		return nil, false, err
	}

	values, ok := response.([]any)
	if !ok || len(values) != 4 {
		return nil, false, nil
	}
	if values[2] != nil {
		return nil, false, nil
	}

	result, ok := values[3].(map[string]any)
	if !ok {
		return nil, false, nil
	}
	prefs, ok := result["prefs"].(map[string]any)
	if !ok {
		return nil, false, nil
	}

	value, exists := prefs[key]
	return value, exists, nil
}

func (c *MarionetteClient) connect() (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(c.host, strconv.Itoa(c.port)), c.ioTimeout)
	if err != nil {
		return nil, err
	}

	if err := conn.SetDeadline(time.Now().Add(c.ioTimeout)); err != nil {
		_ = conn.Close()
		return nil, err
	}

	// 连接建立后会先收到一帧握手信息，读掉即可。
	if _, err := recvMarionetteFrame(conn); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

func (c *MarionetteClient) run(conn net.Conn, message any) (any, error) {
	if err := sendMarionetteFrame(conn, message); err != nil {
		return nil, err
	}
	return recvMarionetteFrame(conn)
}

func sendMarionetteFrame(writer io.Writer, message any) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	frame := fmt.Sprintf("%d:%s", len(payload), payload)
	_, err = io.WriteString(writer, frame)
	return err
}

func recvMarionetteFrame(reader io.Reader) (any, error) {
	lengthBuffer := bytes.Buffer{}
	singleByte := make([]byte, 1)

	for {
		if _, err := io.ReadFull(reader, singleByte); err != nil {
			return nil, err
		}
		if singleByte[0] == ':' {
			break
		}
		lengthBuffer.WriteByte(singleByte[0])
	}

	expectedLength, err := strconv.Atoi(lengthBuffer.String())
	if err != nil {
		return nil, err
	}
	if expectedLength < 0 {
		return nil, fmt.Errorf("marionette 消息长度非法: %d", expectedLength)
	}

	payload := make([]byte, expectedLength)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}

	var message any
	if err := json.Unmarshal(payload, &message); err != nil {
		return nil, err
	}
	return message, nil
}
