package config

import (
	cryptorand "crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	autoFPIPInfoURL          = "https://ip234.in/ip.json"
	defaultWindowsFontSystem = "windows"
	defaultWindowsUserAgent  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:147.0) Gecko/20100101 Firefox/151.0"
	defaultWebGLVersion      = "WebGL 1.0 (OpenGL ES 2.0 Chromium)"
	defaultWebGLGLSLVersion  = "WebGL GLSL ES 1.0 (OpenGL ES GLSL ES 1.0 Chromium)"
	defaultWebGLMaxTexture   = 16384
	defaultWebGLMaxCubeMap   = 16384
	defaultWebGLImageUnits   = 32
	defaultWebGLVertexAttr   = 16
	defaultWebGLPointSize    = 1024
	defaultWebGLViewportDim  = 16384
	defaultCanvasSeedMax     = 1_000_000_000
)

var (
	fetchAutoFPIPInfo = fetchIP234FingerprintProfile
)

type autoFPIPInfoResponse struct {
	IP          string `json:"ip"`
	City        string `json:"city"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Timezone    string `json:"timezone"`
	Region      string `json:"region"`
}

type autoFPVoiceProfile struct {
	Language    string
	VoiceNames  []string
	VoiceLangs  []string
	DefaultName string
	DefaultLang string
}

type autoFPWebGLProfile struct {
	Vendor   string
	Renderer string
}

type autoFPProxyAuth struct {
	Username       string
	Password       string
	HasCredentials bool
}

var autoFPVoiceProfiles = map[string]autoFPVoiceProfile{
	"AU": {
		Language:    "en-AU,en",
		VoiceNames:  []string{"Microsoft Catherine - English (Australia)"},
		VoiceLangs:  []string{"en-AU"},
		DefaultName: "Microsoft Catherine - English (Australia)",
		DefaultLang: "en-AU",
	},
	"BR": {
		Language:    "pt-BR,pt",
		VoiceNames:  []string{"Microsoft Maria Desktop - Portuguese(Brazil)"},
		VoiceLangs:  []string{"pt-BR"},
		DefaultName: "Microsoft Maria Desktop - Portuguese(Brazil)",
		DefaultLang: "pt-BR",
	},
	"CA": {
		Language:    "en-CA,en",
		VoiceNames:  []string{"Microsoft Linda - English (Canada)"},
		VoiceLangs:  []string{"en-CA"},
		DefaultName: "Microsoft Linda - English (Canada)",
		DefaultLang: "en-CA",
	},
	"CN": {
		Language:    "zh-CN,zh",
		VoiceNames:  []string{"Microsoft Huihui Desktop - Chinese (Simplified)"},
		VoiceLangs:  []string{"zh-CN"},
		DefaultName: "Microsoft Huihui Desktop - Chinese (Simplified)",
		DefaultLang: "zh-CN",
	},
	"DE": {
		Language:    "de-DE,de",
		VoiceNames:  []string{"Microsoft Hedda Desktop - German"},
		VoiceLangs:  []string{"de-DE"},
		DefaultName: "Microsoft Hedda Desktop - German",
		DefaultLang: "de-DE",
	},
	"ES": {
		Language:    "es-ES,es",
		VoiceNames:  []string{"Microsoft Helena Desktop - Spanish"},
		VoiceLangs:  []string{"es-ES"},
		DefaultName: "Microsoft Helena Desktop - Spanish",
		DefaultLang: "es-ES",
	},
	"FR": {
		Language:    "fr-FR,fr",
		VoiceNames:  []string{"Microsoft Hortense Desktop - French"},
		VoiceLangs:  []string{"fr-FR"},
		DefaultName: "Microsoft Hortense Desktop - French",
		DefaultLang: "fr-FR",
	},
	"GB": {
		Language:    "en-GB,en",
		VoiceNames:  []string{"Microsoft Hazel Desktop - English (Great Britain)"},
		VoiceLangs:  []string{"en-GB"},
		DefaultName: "Microsoft Hazel Desktop - English (Great Britain)",
		DefaultLang: "en-GB",
	},
	"HK": {
		Language:    "zh-HK,zh,en-US,en",
		VoiceNames:  []string{"Microsoft Tracy - Chinese (Traditional, Hong Kong SAR)"},
		VoiceLangs:  []string{"zh-HK"},
		DefaultName: "Microsoft Tracy - Chinese (Traditional, Hong Kong SAR)",
		DefaultLang: "zh-HK",
	},
	"IE": {
		Language:    "en-IE,en",
		VoiceNames:  []string{"Microsoft Sean - English (Ireland)"},
		VoiceLangs:  []string{"en-IE"},
		DefaultName: "Microsoft Sean - English (Ireland)",
		DefaultLang: "en-IE",
	},
	"IN": {
		Language:    "en-IN,en",
		VoiceNames:  []string{"Microsoft Heera - English (India)"},
		VoiceLangs:  []string{"en-IN"},
		DefaultName: "Microsoft Heera - English (India)",
		DefaultLang: "en-IN",
	},
	"IT": {
		Language:    "it-IT,it",
		VoiceNames:  []string{"Microsoft Elsa Desktop - Italian"},
		VoiceLangs:  []string{"it-IT"},
		DefaultName: "Microsoft Elsa Desktop - Italian",
		DefaultLang: "it-IT",
	},
	"JP": {
		Language:    "ja-JP,ja",
		VoiceNames:  []string{"Microsoft Haruka Desktop - Japanese", "Microsoft Ichiro Desktop - Japanese", "Microsoft Ayumi - Japanese (Japan)"},
		VoiceLangs:  []string{"ja-JP", "ja-JP", "ja-JP"},
		DefaultName: "Microsoft Haruka Desktop - Japanese",
		DefaultLang: "ja-JP",
	},
	"KR": {
		Language:    "ko-KR,ko",
		VoiceNames:  []string{"Microsoft Heami - Korean (Korea)"},
		VoiceLangs:  []string{"ko-KR"},
		DefaultName: "Microsoft Heami - Korean (Korea)",
		DefaultLang: "ko-KR",
	},
	"MX": {
		Language:    "es-MX,es",
		VoiceNames:  []string{"Microsoft Sabina - Spanish (Mexico)"},
		VoiceLangs:  []string{"es-MX"},
		DefaultName: "Microsoft Sabina - Spanish (Mexico)",
		DefaultLang: "es-MX",
	},
	"NL": {
		Language:    "nl-NL,nl",
		VoiceNames:  []string{"Microsoft Frank - Dutch (Netherlands)"},
		VoiceLangs:  []string{"nl-NL"},
		DefaultName: "Microsoft Frank - Dutch (Netherlands)",
		DefaultLang: "nl-NL",
	},
	"NZ": {
		Language:    "en-NZ,en",
		VoiceNames:  []string{"Microsoft Molly - English (New Zealand)"},
		VoiceLangs:  []string{"en-NZ"},
		DefaultName: "Microsoft Molly - English (New Zealand)",
		DefaultLang: "en-NZ",
	},
	"PT": {
		Language:    "pt-PT,pt",
		VoiceNames:  []string{"Microsoft Helia - Portuguese (Portugal)"},
		VoiceLangs:  []string{"pt-PT"},
		DefaultName: "Microsoft Helia - Portuguese (Portugal)",
		DefaultLang: "pt-PT",
	},
	"RU": {
		Language:    "ru-RU,ru",
		VoiceNames:  []string{"Microsoft Irina Desktop - Russian"},
		VoiceLangs:  []string{"ru-RU"},
		DefaultName: "Microsoft Irina Desktop - Russian",
		DefaultLang: "ru-RU",
	},
	"SG": {
		Language:    "en-SG,en,zh-CN,zh",
		VoiceNames:  []string{"Microsoft Zira Desktop - English (United States)"},
		VoiceLangs:  []string{"en-SG"},
		DefaultName: "Microsoft Zira Desktop - English (United States)",
		DefaultLang: "en-SG",
	},
	"TR": {
		Language:    "tr-TR,tr",
		VoiceNames:  []string{"Microsoft Tolga Desktop - Turkish"},
		VoiceLangs:  []string{"tr-TR"},
		DefaultName: "Microsoft Tolga Desktop - Turkish",
		DefaultLang: "tr-TR",
	},
	"TW": {
		Language:    "zh-TW,zh",
		VoiceNames:  []string{"Microsoft Hanhan Desktop - Chinese (Traditional)"},
		VoiceLangs:  []string{"zh-TW"},
		DefaultName: "Microsoft Hanhan Desktop - Chinese (Traditional)",
		DefaultLang: "zh-TW",
	},
	"US": {
		Language:    "en-US,en",
		VoiceNames:  []string{"Microsoft David Desktop - English (United States)", "Microsoft Zira Desktop - English (United States)"},
		VoiceLangs:  []string{"en-US", "en-US"},
		DefaultName: "Microsoft David Desktop - English (United States)",
		DefaultLang: "en-US",
	},
}

var autoFPHardwareConcurrencyProfiles = []int{4, 6, 8, 12, 16}

func (o *FirefoxOptions) withAutoFPFile() (*FirefoxOptions, error) {
	o.ensureDefaults()
	if o.rejectAutoFPMutation("WithAutoFPFile") {
		return o, o.autoFPMutationError()
	}
	width, height := o.resolveAutoFPWindowSize()

	ipInfo, err := fetchAutoFPIPInfo(o.proxy)
	if err != nil {
		return o, err
	}
	proxyAuth, err := extractProxyAuth(o.proxy)
	if err != nil {
		return o, err
	}

	canvasSeed, err := randomPositiveInt(defaultCanvasSeedMax)
	if err != nil {
		return o, err
	}

	webglProfile := pickAutoFPWebGLProfile()
	hardwareConcurrency := pickAutoFPHardwareConcurrency()
	fingerprintText, err := buildAutoFPFingerprintText(ipInfo, proxyAuth, webglProfile, width, height, hardwareConcurrency, canvasSeed)
	if err != nil {
		return o, err
	}

	fpfile, err := os.CreateTemp("", "ruyipage-fp-*.txt")
	if err != nil {
		return o, err
	}

	path := fpfile.Name()
	if _, err := fpfile.WriteString(fingerprintText); err != nil {
		_ = fpfile.Close()
		_ = os.Remove(path)
		return o, err
	}
	if err := fpfile.Close(); err != nil {
		_ = os.Remove(path)
		return o, err
	}

	o.SetManagedFPFile(path)
	o.freezeAutoFP(width, height)
	return o, nil
}

func buildAutoFPFingerprintText(
	ipInfo autoFPIPInfoResponse,
	proxyAuth autoFPProxyAuth,
	webglProfile autoFPWebGLProfile,
	width int,
	height int,
	hardwareConcurrency int,
	canvasSeed int,
) (string, error) {
	normalizedIPInfo, ip, voiceProfile, err := normalizeAutoFPIPInfo(ipInfo)
	if err != nil {
		return "", err
	}

	lines := []string{"webdriver:0"}
	if ipv4 := ip.To4(); ipv4 != nil {
		ipv4Text := ipv4.String()
		lines = append(lines,
			"local_webrtc_ipv4:"+sanitizeFPValue(ipv4Text),
			"public_webrtc_ipv4:"+sanitizeFPValue(ipv4Text),
		)
	} else {
		ipv6Text := ip.String()
		lines = append(lines,
			"local_webrtc_ipv6:"+sanitizeFPValue(ipv6Text),
			"public_webrtc_ipv6:"+sanitizeFPValue(ipv6Text),
		)
	}

	lines = append(lines,
		"timezone:"+sanitizeFPValue(normalizedIPInfo.Timezone),
		"language:"+sanitizeFPValue(voiceProfile.Language),
		"speech.voices.local:"+sanitizeFPValue(strings.Join(voiceProfile.VoiceNames, "|")),
		"speech.voices.local.langs:"+sanitizeFPValue(strings.Join(voiceProfile.VoiceLangs, "|")),
		"speech.voices.default.name:"+sanitizeFPValue(voiceProfile.DefaultName),
		"speech.voices.default.lang:"+sanitizeFPValue(voiceProfile.DefaultLang),
		"font_system:"+defaultWindowsFontSystem,
		"useragent:"+sanitizeFPValue(defaultWindowsUserAgent),
		fmt.Sprintf("hardwareConcurrency:%d", hardwareConcurrency),
		"webgl.vendor:"+sanitizeFPValue(webglProfile.Vendor),
		"webgl.renderer:"+sanitizeFPValue(webglProfile.Renderer),
		"webgl.version:"+defaultWebGLVersion,
		"webgl.glsl_version:"+defaultWebGLGLSLVersion,
		"webgl.unmasked_vendor:"+sanitizeFPValue(webglProfile.Vendor),
		"webgl.unmasked_renderer:"+sanitizeFPValue(webglProfile.Renderer),
		fmt.Sprintf("webgl.max_texture_size:%d", defaultWebGLMaxTexture),
		fmt.Sprintf("webgl.max_cube_map_texture_size:%d", defaultWebGLMaxCubeMap),
		fmt.Sprintf("webgl.max_texture_image_units:%d", defaultWebGLImageUnits),
		fmt.Sprintf("webgl.max_vertex_attribs:%d", defaultWebGLVertexAttr),
		fmt.Sprintf("webgl.aliased_point_size_max:%d", defaultWebGLPointSize),
		fmt.Sprintf("webgl.max_viewport_dim:%d", defaultWebGLViewportDim),
		fmt.Sprintf("width:%d", width),
		fmt.Sprintf("height:%d", height),
		fmt.Sprintf("canvas:%d", canvasSeed),
	)

	if proxyAuth.HasCredentials {
		lines = append(lines, "httpauth.username:"+sanitizeFPValue(proxyAuth.Username))
		lines = append(lines, "httpauth.password:"+sanitizeFPValue(proxyAuth.Password))
	}

	return strings.Join(lines, "\n") + "\n", nil
}

func fetchIP234FingerprintProfile(proxy string) (autoFPIPInfoResponse, error) {
	transport, err := buildProxyAwareHTTPTransport(proxy)
	if err != nil {
		return autoFPIPInfoResponse{}, err
	}

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	req, err := http.NewRequest(http.MethodGet, autoFPIPInfoURL, nil)
	if err != nil {
		return autoFPIPInfoResponse{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", defaultWindowsUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return autoFPIPInfoResponse{}, fmt.Errorf("请求 ip234 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return autoFPIPInfoResponse{}, fmt.Errorf("ip234 返回异常状态 %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result autoFPIPInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return autoFPIPInfoResponse{}, fmt.Errorf("解析 ip234 响应失败: %w", err)
	}
	if _, _, _, err := normalizeAutoFPIPInfo(result); err != nil {
		return autoFPIPInfoResponse{}, err
	}
	return result, nil
}

func buildProxyAwareHTTPTransport(proxy string) (*http.Transport, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if strings.TrimSpace(proxy) == "" {
		return transport, nil
	}

	proxyURL, err := parseProxyURL(proxy)
	if err != nil {
		return nil, err
	}
	transport.Proxy = http.ProxyURL(proxyURL)
	return transport, nil
}

func parseProxyURL(proxy string) (*url.URL, error) {
	trimmed := strings.TrimSpace(proxy)
	if trimmed == "" {
		return nil, nil
	}
	if !strings.Contains(trimmed, "://") {
		trimmed = "http://" + trimmed
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("代理地址无效: %w", err)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("代理地址缺少 host: %s", proxy)
	}
	return parsed, nil
}

func extractProxyAuth(proxy string) (autoFPProxyAuth, error) {
	proxyURL, err := parseProxyURL(proxy)
	if err != nil {
		return autoFPProxyAuth{}, err
	}
	if proxyURL == nil || proxyURL.User == nil {
		return autoFPProxyAuth{}, nil
	}

	username := proxyURL.User.Username()
	password, hasPassword := proxyURL.User.Password()
	if username == "" && !hasPassword {
		return autoFPProxyAuth{}, nil
	}

	return autoFPProxyAuth{
		Username:       username,
		Password:       password,
		HasCredentials: true,
	}, nil
}

func resolveAutoFPVoiceProfile(countryCode string) (autoFPVoiceProfile, error) {
	normalizedCountryCode := strings.ToUpper(strings.TrimSpace(countryCode))
	if profile, ok := autoFPVoiceProfiles[normalizedCountryCode]; ok {
		return profile, nil
	}
	return autoFPVoiceProfile{}, fmt.Errorf("ip234 返回了未支持的 country_code: %q", countryCode)
}

func normalizeAutoFPIPInfo(ipInfo autoFPIPInfoResponse) (autoFPIPInfoResponse, net.IP, autoFPVoiceProfile, error) {
	normalized := autoFPIPInfoResponse{
		IP:          strings.TrimSpace(ipInfo.IP),
		City:        strings.TrimSpace(ipInfo.City),
		Country:     strings.TrimSpace(ipInfo.Country),
		CountryCode: strings.ToUpper(strings.TrimSpace(ipInfo.CountryCode)),
		Timezone:    strings.TrimSpace(ipInfo.Timezone),
		Region:      strings.TrimSpace(ipInfo.Region),
	}

	if normalized.IP == "" {
		return autoFPIPInfoResponse{}, nil, autoFPVoiceProfile{}, fmt.Errorf("ip234 响应缺少 ip 字段")
	}
	ip := net.ParseIP(normalized.IP)
	if ip == nil {
		return autoFPIPInfoResponse{}, nil, autoFPVoiceProfile{}, fmt.Errorf("ip234 返回了无效 IP: %q", ipInfo.IP)
	}
	if normalized.City == "" {
		return autoFPIPInfoResponse{}, nil, autoFPVoiceProfile{}, fmt.Errorf("ip234 响应缺少 city 字段")
	}
	if normalized.Country == "" {
		return autoFPIPInfoResponse{}, nil, autoFPVoiceProfile{}, fmt.Errorf("ip234 响应缺少 country 字段")
	}
	if !isValidAutoFPCountryCode(normalized.CountryCode) {
		return autoFPIPInfoResponse{}, nil, autoFPVoiceProfile{}, fmt.Errorf("ip234 返回了无效 country_code: %q", ipInfo.CountryCode)
	}
	if normalized.Timezone == "" {
		return autoFPIPInfoResponse{}, nil, autoFPVoiceProfile{}, fmt.Errorf("ip234 响应缺少 timezone 字段")
	}
	if _, err := time.LoadLocation(normalized.Timezone); err != nil {
		return autoFPIPInfoResponse{}, nil, autoFPVoiceProfile{}, fmt.Errorf("ip234 返回了无效 timezone: %q", ipInfo.Timezone)
	}
	if normalized.Region == "" {
		return autoFPIPInfoResponse{}, nil, autoFPVoiceProfile{}, fmt.Errorf("ip234 响应缺少 region 字段")
	}

	voiceProfile, err := resolveAutoFPVoiceProfile(normalized.CountryCode)
	if err != nil {
		return autoFPIPInfoResponse{}, nil, autoFPVoiceProfile{}, err
	}

	return normalized, ip, voiceProfile, nil
}

func isValidAutoFPCountryCode(countryCode string) bool {
	if len(countryCode) != 2 {
		return false
	}
	for _, char := range countryCode {
		if char < 'A' || char > 'Z' {
			return false
		}
	}
	return true
}

func pickAutoFPWebGLProfile() autoFPWebGLProfile {
	if len(autoFPWindowsWebGLProfiles) == 0 {
		return autoFPWebGLProfile{
			Vendor:   "Google Inc. (AMD)",
			Renderer: "ANGLE (AMD, AMD Radeon RX 6800 XT Direct3D11 vs_5_0 ps_5_0, D3D11)",
		}
	}

	index, err := randomPositiveInt(len(autoFPWindowsWebGLProfiles))
	if err != nil {
		return autoFPWindowsWebGLProfiles[0]
	}
	return autoFPWindowsWebGLProfiles[index-1]
}

func pickAutoFPHardwareConcurrency() int {
	actual := runtime.NumCPU()
	if actual <= 0 {
		return 8
	}

	best := autoFPHardwareConcurrencyProfiles[0]
	bestDistance := absInt(best - actual)
	for _, candidate := range autoFPHardwareConcurrencyProfiles[1:] {
		distance := absInt(candidate - actual)
		if distance < bestDistance {
			best = candidate
			bestDistance = distance
		}
	}
	return best
}

func randomPositiveInt(max int) (int, error) {
	if max <= 0 {
		return 0, fmt.Errorf("随机上限必须大于 0，当前为 %d", max)
	}

	value, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(value.Int64()) + 1, nil
}

func sanitizeFPValue(value string) string {
	replacer := strings.NewReplacer("\r", " ", "\n", " ")
	return strings.TrimSpace(replacer.Replace(value))
}

func sameCleanPath(a string, b string) bool {
	if strings.TrimSpace(a) == "" || strings.TrimSpace(b) == "" {
		return false
	}

	left, err := filepath.Abs(a)
	if err != nil {
		left = a
	}
	right, err := filepath.Abs(b)
	if err != nil {
		right = b
	}
	return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
