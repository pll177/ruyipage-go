package testserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	defaultHost = "127.0.0.1"
	defaultPort = 8888
)

// TestServer 是 examples 共享的轻量本地 HTTP 测试服务器。
type TestServer struct {
	Host string
	Port int

	mu       sync.Mutex
	server   *http.Server
	listener net.Listener
}

// New 创建示例测试服务器；host 为空时默认 127.0.0.1，port 为 0 时走 Python 默认 8888。
func New(host string, port int) *TestServer {
	if strings.TrimSpace(host) == "" {
		host = defaultHost
	}
	if port == 0 {
		port = defaultPort
	}
	return &TestServer{Host: host, Port: port}
}

// Start 启动本地测试服务器；重复调用会直接复用已启动实例。
func (s *TestServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		return nil
	}

	host := s.Host
	if strings.TrimSpace(host) == "" {
		host = defaultHost
	}
	port := s.Port
	if port == 0 {
		port = defaultPort
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	s.listener = listener
	if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
		s.Port = tcpAddr.Port
	}
	s.Host = host

	server := &http.Server{
		Handler:  http.HandlerFunc(s.serveHTTP),
		ErrorLog: log.New(io.Discard, "", 0),
	}
	s.server = server

	go func() {
		err := server.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			_ = listener.Close()
		}
	}()
	return nil
}

// Stop 停止本地测试服务器。
func (s *TestServer) Stop() error {
	s.mu.Lock()
	server := s.server
	listener := s.listener
	s.server = nil
	s.listener = nil
	s.mu.Unlock()

	if server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if listener != nil {
		_ = listener.Close()
	}
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// GetURL 返回指定路径的完整访问地址。
func (s *TestServer) GetURL(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	host := s.Host
	if strings.TrimSpace(host) == "" {
		host = defaultHost
	}
	port := s.Port
	if port == 0 {
		port = defaultPort
	}
	return fmt.Sprintf("http://%s:%d%s", host, port, path)
}

func (s *TestServer) serveHTTP(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodOptions:
		s.handleOptions(writer)
		return
	case http.MethodGet:
		s.handleGet(writer, request)
		return
	case http.MethodPost:
		s.handlePost(writer, request)
		return
	default:
		writeText(writer, http.StatusNotFound, "not found", nil)
	}
}

func (s *TestServer) handleGet(writer http.ResponseWriter, request *http.Request) {
	switch request.URL.Path {
	case "/set-cookie":
		writeText(writer, http.StatusOK, "cookie set ok", [][2]string{
			{"Set-Cookie", "server_cookie=server_value; Path=/"},
			{"Set-Cookie", "session_id=abc123; Path=/; HttpOnly"},
		})
	case "/get-cookie":
		writeText(writer, http.StatusOK, request.Header.Get("Cookie"), nil)
	case "/api/data":
		writeJSON(writer, http.StatusOK, map[string]any{
			"status": "ok",
			"data":   map[string]any{"message": "来自测试服务器的数据"},
		}, corsHeaders())
	case "/api/headers":
		writeJSON(writer, http.StatusOK, headerMap(request.Header), corsHeaders())
	case "/api/collector":
		writeJSON(writer, http.StatusOK, map[string]any{
			"status":  "ok",
			"source":  "collector",
			"message": "stable response body for data collector",
			"headers": headerMap(request.Header),
		}, corsHeaders())
	case "/api/error":
		writeJSON(writer, http.StatusInternalServerError, map[string]any{
			"status":  "error",
			"message": "server error",
		}, corsHeaders())
	case "/api/mock-source":
		writeJSON(writer, http.StatusOK, map[string]any{
			"status": "ok",
			"source": "real-server",
		}, corsHeaders())
	case "/api/slow":
		writeJSON(writer, http.StatusOK, map[string]any{
			"status": "ok",
			"source": "slow-server",
		}, corsHeaders())
	case "/api/auth":
		expected := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass"))
		if request.Header.Get("Authorization") != expected {
			writeEmpty(writer, http.StatusUnauthorized, [][2]string{
				{"WWW-Authenticate", `Basic realm="RuyiTest"`},
			})
			return
		}
		writeJSON(writer, http.StatusOK, map[string]any{
			"status": "ok",
			"auth":   true,
			"user":   "user",
		}, corsHeaders())
	case "/download/text":
		writeText(writer, http.StatusOK, "hello download", [][2]string{
			{"Content-Disposition", `attachment; filename="test.txt"`},
			{"Cache-Control", "no-store"},
		})
	case "/download/json":
		writeJSON(writer, http.StatusOK, map[string]any{"ok": true}, [][2]string{
			{"Content-Disposition", `attachment; filename="test.json"`},
			{"Cache-Control", "no-store"},
		})
	case "/nav/basic":
		writeHTML(writer, http.StatusOK, "<!DOCTYPE html>\n<html><head><meta charset='utf-8'><title>Nav Basic</title></head>\n<body><h1>Nav Basic</h1></body></html>")
	case "/nav/fragment":
		writeHTML(writer, http.StatusOK, "<!DOCTYPE html>\n<html><head><meta charset='utf-8'><title>Nav Fragment</title></head>\n<body>\n<h1 id='a'>A</h1>\n<div style='height:1200px'></div>\n<h1 id='b'>B</h1>\n</body></html>")
	case "/nav/history":
		writeHTML(writer, http.StatusOK, "<!DOCTYPE html>\n<html><head><meta charset='utf-8'><title>Nav History</title></head>\n<body><h1>Nav History</h1></body></html>")
	default:
		writeText(writer, http.StatusNotFound, "not found", nil)
	}
}

func (s *TestServer) handlePost(writer http.ResponseWriter, request *http.Request) {
	switch request.URL.Path {
	case "/api/echo":
		body, _ := io.ReadAll(request.Body)
		writeJSON(writer, http.StatusOK, map[string]any{
			"status":       "ok",
			"method":       http.MethodPost,
			"body":         string(body),
			"content_type": request.Header.Get("Content-Type"),
		}, corsHeaders())
	default:
		writeText(writer, http.StatusNotFound, "not found", corsHeaders())
	}
}

func (s *TestServer) handleOptions(writer http.ResponseWriter) {
	writeEmpty(writer, http.StatusNoContent, corsHeaders())
}

func corsHeaders() [][2]string {
	return [][2]string{
		{"Access-Control-Allow-Origin", "*"},
		{"Access-Control-Allow-Methods", "GET, POST, OPTIONS"},
		{"Access-Control-Allow-Headers", "Content-Type, X-Ruyi-Demo, User-Agent"},
	}
}

func headerMap(headers http.Header) map[string]string {
	result := make(map[string]string, len(headers))
	for key, values := range headers {
		result[key] = strings.Join(values, ", ")
	}
	return result
}

func writeText(writer http.ResponseWriter, status int, body string, headers [][2]string) {
	data := []byte(body)
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	for _, header := range headers {
		writer.Header().Add(header[0], header[1])
	}
	writer.Header().Set("Content-Length", strconvItoa(len(data)))
	writer.WriteHeader(status)
	_, _ = writer.Write(data)
}

func writeHTML(writer http.ResponseWriter, status int, body string) {
	data := []byte(body)
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.Header().Set("Content-Length", strconvItoa(len(data)))
	writer.WriteHeader(status)
	_, _ = writer.Write(data)
}

func writeJSON(writer http.ResponseWriter, status int, payload any, headers [][2]string) {
	data, _ := json.Marshal(payload)
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	for _, header := range headers {
		writer.Header().Add(header[0], header[1])
	}
	writer.Header().Set("Content-Length", strconvItoa(len(data)))
	writer.WriteHeader(status)
	_, _ = writer.Write(data)
}

func writeEmpty(writer http.ResponseWriter, status int, headers [][2]string) {
	for _, header := range headers {
		writer.Header().Add(header[0], header[1])
	}
	writer.Header().Set("Content-Length", "0")
	writer.WriteHeader(status)
}

func strconvItoa(value int) string {
	return fmt.Sprintf("%d", value)
}
