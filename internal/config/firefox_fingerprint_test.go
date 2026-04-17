package config

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestFirefoxOptionsWithAutoFPFileWritesExpectedFingerprint(t *testing.T) {
	installAutoFPIPStub(t, func(proxy string) (autoFPIPInfoResponse, error) {
		if proxy != "http://proxy.example.com:8888" {
			t.Fatalf("unexpected proxy: %s", proxy)
		}
		return jpAutoFPIPInfo(), nil
	})

	opts := NewFirefoxOptions().
		WithProxy("http://proxy.example.com:8888").
		WithWindowSize(551, 500)
	if _, err := opts.WithAutoFPFile(); err != nil {
		t.Fatalf("WithAutoFPFile returned error: %v", err)
	}

	fpfilePath := opts.FPFile()
	t.Cleanup(func() {
		_ = os.Remove(fpfilePath)
	})

	fields := readFingerprintFields(t, fpfilePath)
	assertFingerprintField(t, fields, "webdriver", "0")
	assertFingerprintField(t, fields, "local_webrtc_ipv4", "1.2.3.4")
	assertFingerprintField(t, fields, "public_webrtc_ipv4", "1.2.3.4")
	assertFingerprintField(t, fields, "timezone", "Asia/Tokyo")
	assertFingerprintField(t, fields, "language", "ja-JP,ja")
	assertFingerprintField(t, fields, "speech.voices.default.name", "Microsoft Haruka Desktop - Japanese")
	assertFingerprintField(t, fields, "font_system", "windows")
	assertFingerprintField(t, fields, "useragent", defaultWindowsUserAgent)
	assertFingerprintField(t, fields, "width", "551")
	assertFingerprintField(t, fields, "height", "500")
	assertFingerprintField(t, fields, "webgl.version", defaultWebGLVersion)
	if _, ok := fields["canvas"]; !ok {
		t.Fatalf("expected canvas field to be present")
	}
	if _, ok := fields["httpauth.username"]; ok {
		t.Fatalf("did not expect proxy auth fields for proxy without credentials")
	}

	command, err := opts.BuildCommand()
	if err != nil {
		t.Fatalf("BuildCommand returned error: %v", err)
	}
	assertCommandContains(t, command, "--width=551")
	assertCommandContains(t, command, "--height=500")
}

func TestFirefoxOptionsWithAutoFPFileUsesDefaultWindowSizeWhenUnset(t *testing.T) {
	installAutoFPIPStub(t, func(proxy string) (autoFPIPInfoResponse, error) {
		if proxy != "http://proxy.example.com:8888" {
			t.Fatalf("unexpected proxy: %s", proxy)
		}
		return jpAutoFPIPInfo(), nil
	})

	opts := NewFirefoxOptions().WithProxy("http://proxy.example.com:8888")
	if _, err := opts.WithAutoFPFile(); err != nil {
		t.Fatalf("WithAutoFPFile returned error: %v", err)
	}

	fpfilePath := opts.FPFile()
	t.Cleanup(func() {
		_ = os.Remove(fpfilePath)
	})

	fields := readFingerprintFields(t, fpfilePath)
	assertFingerprintField(t, fields, "width", fmt.Sprintf("%d", defaultQuickStartWidth))
	assertFingerprintField(t, fields, "height", fmt.Sprintf("%d", defaultQuickStartHeight))

	command, err := opts.BuildCommand()
	if err != nil {
		t.Fatalf("BuildCommand returned error: %v", err)
	}
	assertCommandContains(t, command, fmt.Sprintf("--width=%d", defaultQuickStartWidth))
	assertCommandContains(t, command, fmt.Sprintf("--height=%d", defaultQuickStartHeight))
}

func TestFirefoxOptionsWithAutoFPFileIncludesProxyAuthForHongKong(t *testing.T) {
	installAutoFPIPStub(t, func(proxy string) (autoFPIPInfoResponse, error) {
		if proxy != "http://user:pass@proxy.example.com:8080" {
			t.Fatalf("unexpected proxy: %s", proxy)
		}
		return hkAutoFPIPInfo(), nil
	})

	opts := NewFirefoxOptions().
		WithProxy("http://user:pass@proxy.example.com:8080").
		WithWindowSize(800, 600)
	if _, err := opts.WithAutoFPFile(); err != nil {
		t.Fatalf("WithAutoFPFile returned error: %v", err)
	}

	fpfilePath := opts.FPFile()
	t.Cleanup(func() {
		_ = os.Remove(fpfilePath)
	})

	fields := readFingerprintFields(t, fpfilePath)
	assertFingerprintField(t, fields, "local_webrtc_ipv4", "151.243.25.92")
	assertFingerprintField(t, fields, "public_webrtc_ipv4", "151.243.25.92")
	assertFingerprintField(t, fields, "language", "zh-HK,zh,en-US,en")
	assertFingerprintField(t, fields, "timezone", "Asia/Hong_Kong")
	assertFingerprintField(t, fields, "speech.voices.default.name", "Microsoft Tracy - Chinese (Traditional, Hong Kong SAR)")
	assertFingerprintField(t, fields, "httpauth.username", "user")
	assertFingerprintField(t, fields, "httpauth.password", "pass")
	if _, ok := fields["speech.voices.remote"]; ok {
		t.Fatalf("did not expect remote speech voices to be written")
	}
}

func TestFirefoxOptionsWithAutoFPFileRejectsInvalidIP234Payload(t *testing.T) {
	testCases := []struct {
		name        string
		response    autoFPIPInfoResponse
		wantErrPart string
	}{
		{
			name: "missing country",
			response: autoFPIPInfoResponse{
				IP:          "1.2.3.4",
				City:        "new york",
				CountryCode: "US",
				Timezone:    "America/New_York",
				Region:      "New York",
			},
			wantErrPart: "country 字段",
		},
		{
			name: "missing timezone",
			response: autoFPIPInfoResponse{
				IP:          "1.2.3.4",
				City:        "new york",
				Country:     "United States",
				CountryCode: "US",
				Region:      "New York",
			},
			wantErrPart: "timezone 字段",
		},
		{
			name: "invalid ip",
			response: autoFPIPInfoResponse{
				IP:          "not-an-ip",
				City:        "new york",
				Country:     "United States",
				CountryCode: "US",
				Timezone:    "America/New_York",
				Region:      "New York",
			},
			wantErrPart: "无效 IP",
		},
		{
			name: "unsupported country code",
			response: autoFPIPInfoResponse{
				IP:          "1.2.3.4",
				City:        "unknown",
				Country:     "Unknown",
				CountryCode: "ZZ",
				Timezone:    "Asia/Tokyo",
				Region:      "Unknown",
			},
			wantErrPart: "未支持的 country_code",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			installAutoFPIPStub(t, func(proxy string) (autoFPIPInfoResponse, error) {
				return testCase.response, nil
			})

			opts := NewFirefoxOptions().WithProxy("http://proxy.example.com:8888")
			_, err := opts.WithAutoFPFile()
			if err == nil {
				t.Fatalf("expected WithAutoFPFile to fail")
			}
			if !strings.Contains(err.Error(), testCase.wantErrPart) {
				t.Fatalf("error = %q, want substring %q", err.Error(), testCase.wantErrPart)
			}
		})
	}
}

func TestFirefoxOptionsWithAutoFPFileAllowsMissingProxy(t *testing.T) {
	installAutoFPIPStub(t, func(proxy string) (autoFPIPInfoResponse, error) {
		if proxy != "" {
			t.Fatalf("expected empty proxy, got %q", proxy)
		}
		return jpAutoFPIPInfo(), nil
	})

	opts := NewFirefoxOptions()
	if _, err := opts.WithAutoFPFile(); err != nil {
		t.Fatalf("WithAutoFPFile returned error without proxy: %v", err)
	}
}

func TestFirefoxOptionsRejectsConfigChangesAfterAutoFPFile(t *testing.T) {
	testCases := []struct {
		name   string
		mutate func(t *testing.T, opts *FirefoxOptions)
	}{
		{"WithBrowserPath", func(t *testing.T, opts *FirefoxOptions) { opts.WithBrowserPath(`D:\Firefox\firefox.exe`) }},
		{"WithAddress", func(t *testing.T, opts *FirefoxOptions) { opts.WithAddress("127.0.0.1:9333") }},
		{"WithPort", func(t *testing.T, opts *FirefoxOptions) { opts.WithPort(9333) }},
		{"WithProfile", func(t *testing.T, opts *FirefoxOptions) { opts.WithProfile(t.TempDir()) }},
		{"WithUserDir", func(t *testing.T, opts *FirefoxOptions) { opts.WithUserDir(t.TempDir()) }},
		{"WithArgument", func(t *testing.T, opts *FirefoxOptions) { opts.WithArgument("--new-arg", "1") }},
		{"WithoutArgument", func(t *testing.T, opts *FirefoxOptions) { opts.WithoutArgument("--width") }},
		{"WithPreference", func(t *testing.T, opts *FirefoxOptions) { opts.WithPreference("browser.tabs.warnOnClose", true) }},
		{"WithUserPromptHandler", func(t *testing.T, opts *FirefoxOptions) {
			opts.WithUserPromptHandler(map[string]string{"default": "dismiss"})
		}},
		{"Headless", func(t *testing.T, opts *FirefoxOptions) { opts.Headless(true) }},
		{"WithProxy", func(t *testing.T, opts *FirefoxOptions) { opts.WithProxy("http://proxy-b.example.com:9999") }},
		{"WithDownloadPath", func(t *testing.T, opts *FirefoxOptions) { opts.WithDownloadPath(t.TempDir()) }},
		{"WithLoadMode", func(t *testing.T, opts *FirefoxOptions) { opts.WithLoadMode(LoadModeEager) }},
		{"WithTimeouts", func(t *testing.T, opts *FirefoxOptions) { opts.WithTimeouts(1, 2, 3) }},
		{"WithBaseTimeout", func(t *testing.T, opts *FirefoxOptions) { opts.WithBaseTimeout(1) }},
		{"WithPageLoadTimeout", func(t *testing.T, opts *FirefoxOptions) { opts.WithPageLoadTimeout(2) }},
		{"WithScriptTimeout", func(t *testing.T, opts *FirefoxOptions) { opts.WithScriptTimeout(3) }},
		{"ExistingOnly", func(t *testing.T, opts *FirefoxOptions) { opts.ExistingOnly(true) }},
		{"AutoPortEnabled", func(t *testing.T, opts *FirefoxOptions) { opts.AutoPortEnabled(true) }},
		{"WithAutoPortStart", func(t *testing.T, opts *FirefoxOptions) { opts.WithAutoPortStart(9300) }},
		{"WithRetry", func(t *testing.T, opts *FirefoxOptions) { opts.WithRetry(1, 0.5) }},
		{"WithRetryTimes", func(t *testing.T, opts *FirefoxOptions) { opts.WithRetryTimes(1) }},
		{"WithRetryInterval", func(t *testing.T, opts *FirefoxOptions) { opts.WithRetryInterval(0.5) }},
		{"WithUserContext", func(t *testing.T, opts *FirefoxOptions) { opts.WithUserContext("personal") }},
		{"WithFPFile", func(t *testing.T, opts *FirefoxOptions) { opts.WithFPFile(`C:\tmp\manual.txt`) }},
		{"PrivateMode", func(t *testing.T, opts *FirefoxOptions) { opts.PrivateMode(true) }},
		{"XPathPickerEnabled", func(t *testing.T, opts *FirefoxOptions) { opts.XPathPickerEnabled(true) }},
		{"ActionVisualEnabled", func(t *testing.T, opts *FirefoxOptions) { opts.ActionVisualEnabled(true) }},
		{"CloseBrowserOnExitEnabled", func(t *testing.T, opts *FirefoxOptions) { opts.CloseBrowserOnExitEnabled(false) }},
		{"WithWindowSize", func(t *testing.T, opts *FirefoxOptions) { opts.WithWindowSize(2000, 1200) }},
		{"QuickStart", func(t *testing.T, opts *FirefoxOptions) {
			opts.QuickStart(FirefoxQuickStartOptions{
				Port:         9333,
				WindowWidth:  1440,
				WindowHeight: 900,
			})
		}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			opts := newFrozenAutoFPOptions(t)
			testCase.mutate(t, opts)
			assertAutoFPMutationBlocked(t, opts, testCase.name)
		})
	}
}

func TestFirefoxOptionsRejectsSecondAutoFPFileCall(t *testing.T) {
	opts := newFrozenAutoFPOptions(t)

	_, err := opts.WithAutoFPFile()
	if err == nil {
		t.Fatalf("expected second WithAutoFPFile to fail")
	}
	if !strings.Contains(err.Error(), "WithAutoFPFile()") {
		t.Fatalf("error = %q, want WithAutoFPFile mutation guidance", err.Error())
	}
}

func TestApplyProxyPreferencesStripsUserInfo(t *testing.T) {
	prefs := map[string]any{}
	applyProxyPreferences(prefs, "http://user:pass@proxy.example.com:8080")

	if got := prefs["network.proxy.http"]; got != "proxy.example.com" {
		t.Fatalf("network.proxy.http = %v, want proxy.example.com", got)
	}
	if got := prefs["network.proxy.http_port"]; got != 8080 {
		t.Fatalf("network.proxy.http_port = %v, want 8080", got)
	}
	if got := prefs["network.proxy.ssl"]; got != "proxy.example.com" {
		t.Fatalf("network.proxy.ssl = %v, want proxy.example.com", got)
	}
	if got := prefs["network.proxy.ssl_port"]; got != 8080 {
		t.Fatalf("network.proxy.ssl_port = %v, want 8080", got)
	}
}

func installAutoFPIPStub(t *testing.T, stub func(proxy string) (autoFPIPInfoResponse, error)) {
	t.Helper()
	oldFetch := fetchAutoFPIPInfo
	fetchAutoFPIPInfo = stub
	t.Cleanup(func() {
		fetchAutoFPIPInfo = oldFetch
	})
}

func newFrozenAutoFPOptions(t *testing.T) *FirefoxOptions {
	t.Helper()
	installAutoFPIPStub(t, func(proxy string) (autoFPIPInfoResponse, error) {
		if proxy != "http://proxy.example.com:8888" {
			t.Fatalf("unexpected proxy: %s", proxy)
		}
		return jpAutoFPIPInfo(), nil
	})

	opts := NewFirefoxOptions().
		WithProxy("http://proxy.example.com:8888").
		WithWindowSize(551, 500)
	if _, err := opts.WithAutoFPFile(); err != nil {
		t.Fatalf("WithAutoFPFile returned error: %v", err)
	}
	return opts
}

func assertAutoFPMutationBlocked(t *testing.T, opts *FirefoxOptions, methodName string) {
	t.Helper()
	want := fmt.Sprintf("%s()", methodName)

	err := opts.Validate()
	if err == nil {
		t.Fatalf("expected Validate to fail after %s", methodName)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("Validate error = %q, want %q", err.Error(), want)
	}

	_, err = opts.BuildCommand()
	if err == nil {
		t.Fatalf("expected BuildCommand to fail after %s", methodName)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("BuildCommand error = %q, want %q", err.Error(), want)
	}
}

func jpAutoFPIPInfo() autoFPIPInfoResponse {
	return autoFPIPInfoResponse{
		IP:          "1.2.3.4",
		City:        "tokyo",
		Country:     "Japan",
		CountryCode: "JP",
		Timezone:    "Asia/Tokyo",
		Region:      "Tokyo",
	}
}

func hkAutoFPIPInfo() autoFPIPInfoResponse {
	return autoFPIPInfoResponse{
		IP:          "151.243.25.92",
		City:        "hong kong",
		Country:     "Hong Kong",
		CountryCode: "HK",
		Timezone:    "Asia/Hong_Kong",
		Region:      "Hong Kong",
	}
}

func readFingerprintFields(t *testing.T, path string) map[string]string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fingerprint file %s: %v", path, err)
	}

	fields := map[string]string{}
	for _, rawLine := range strings.Split(string(content), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			t.Fatalf("invalid fingerprint line: %s", line)
		}
		fields[parts[0]] = parts[1]
	}
	return fields
}

func assertFingerprintField(t *testing.T, fields map[string]string, key string, want string) {
	t.Helper()

	got, ok := fields[key]
	if !ok {
		t.Fatalf("missing fingerprint field %s", key)
	}
	if got != want {
		t.Fatalf("%s = %q, want %q", key, got, want)
	}
}

func assertCommandContains(t *testing.T, command []string, want string) {
	t.Helper()
	assertCommandContainsCount(t, command, want, 1)
}

func assertCommandContainsCount(t *testing.T, command []string, want string, count int) {
	t.Helper()

	got := 0
	for _, item := range command {
		if item == want {
			got++
		}
	}
	if got != count {
		t.Fatalf("command contains %q %d times, want %d; command=%v", want, got, count, command)
	}
}
